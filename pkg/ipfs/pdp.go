package ipfs

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	commcid "github.com/filecoin-project/go-fil-commcid"
	commp "github.com/filecoin-project/go-fil-commp-hashhash"
	"github.com/golang-jwt/jwt/v5"
	"github.com/ipfs/go-cid"
)

type PDP struct {
	serviceURL  string
	serviceName string
	privateKey  *ecdsa.PrivateKey
	proofSetID  uint64
	logger      *slog.Logger
}

func NewPDP(serviceURL, serviceName, privateKey string, proofSetID uint64, logger *slog.Logger) (*PDP, error) {
	var privateKeyBytes []byte
	if data, err := os.ReadFile(privateKey); err == nil {
		privateKeyBytes = data
	} else {
		privateKeyBytes = bytes.ReplaceAll([]byte(privateKey), []byte("\\n"), []byte("\n"))
	}

	block, _ := pem.Decode(privateKeyBytes)
	if block == nil {
		return nil, fmt.Errorf("failed to parse private key PEM")
	}

	// Parse the private key
	privKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %v", err)
	}
	ecdsaPrivKey, ok := privKey.(*ecdsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("private key is not ECDSA")
	}

	return &PDP{serviceURL: serviceURL, serviceName: serviceName, privateKey: ecdsaPrivKey, proofSetID: proofSetID, logger: logger}, nil
}

func (p *PDP) Add(ctx context.Context, content io.ReadSeeker) (cid.Cid, error) {
	jwtToken, err := createJWTToken(p.serviceName, p.privateKey)
	if err != nil {
		return cid.Undef, fmt.Errorf("failed to create JWT token: %v", err)
	}

	// Compute CommP (PieceCID)
	pieceCIDComputed, pieceSize, _, commpDigest, err := preparePiece(content)
	if err != nil {
		return cid.Undef, fmt.Errorf("failed to prepare piece: %v", err)
	}

	// Prepare the check data
	hashHex := hex.EncodeToString(commpDigest)
	checkData := map[string]any{
		"name": "sha2-256-trunc254-padded",
		"hash": hashHex,
		"size": pieceSize,
	}

	// Prepare the request data
	reqData := map[string]any{
		"check": checkData,
	}

	reqBody, err := json.Marshal(reqData)
	if err != nil {
		return cid.Undef, fmt.Errorf("failed to marshal request data: %v", err)
	}
	if err := uploadOnePiece(ctx, http.DefaultClient, p.serviceURL, reqBody, jwtToken, content, pieceSize); err != nil {
		return cid.Undef, fmt.Errorf("failed to upload piece: %v", err)
	}

	go func() {
		time.Sleep(time.Second * 20)
		if err := p.AddRoots(context.Background(), "", []string{fmt.Sprintf("%s:%s", pieceCIDComputed.String(), pieceCIDComputed.String())}); err != nil {
			p.logger.Error("failed to add roots", "pieceCIDComputed", pieceCIDComputed.String(), "error", err)
		}
	}()

	return pieceCIDComputed, nil
}

func (p *PDP) Get(ctx context.Context, cid cid.Cid) (io.ReadCloser, error) {
	// Create the download URL
	downloadURL := fmt.Sprintf("%s/piece/%s", p.serviceURL, cid.String())

	// Create the GET request
	req, err := http.NewRequestWithContext(ctx, "GET", downloadURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request for CID %s: %v", cid.String(), err)
	}

	// Send the request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to download piece %s: %v", cid.String(), err)
	}

	// Check response status
	if resp.StatusCode != http.StatusOK {
		_ = resp.Body.Close()
		return nil, fmt.Errorf("failed to download piece %s: status code %d", cid.String(), resp.StatusCode)
	}

	return resp.Body, nil
}

func (p *PDP) CreateProofSet(ctx context.Context, recordKeeper, extraDataHexStr string) error {
	// Validate extraData hex string and its decoded length
	if err := validateExtraData(extraDataHexStr); err != nil {
		return err
	}

	jwtToken, err := createJWTToken(p.serviceName, p.privateKey)
	if err != nil {
		return fmt.Errorf("failed to create JWT token: %v", err)
	}

	// Construct the request payload
	requestBody := map[string]string{
		"recordKeeper": recordKeeper,
	}

	requestBodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %v", err)
	}

	// Append /pdp/proof-sets to the service URL
	postURL := p.serviceURL + "/pdp/proof-sets"

	// Create the POST request
	req, err := http.NewRequestWithContext(ctx, "POST", postURL, bytes.NewBuffer(requestBodyBytes))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+jwtToken)
	req.Header.Set("Content-Type", "application/json")

	// Send the request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close() // nolint:errcheck
	_, _ = io.Copy(io.Discard, resp.Body)

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("failed to create proof set, status code %d", resp.StatusCode)
	}

	return nil
}

// rootInputs is Root CID and its subroots. Format: rootCID:subrootCID1+subrootCID2,...
func (p *PDP) AddRoots(ctx context.Context, extraDataHexStr string, rootInputs []string) error {
	// Validate extraData hex string and its decoded length
	if err := validateExtraData(extraDataHexStr); err != nil {
		return err
	}

	jwtToken, err := createJWTToken(p.serviceName, p.privateKey)
	if err != nil {
		return fmt.Errorf("failed to create JWT token: %v", err)
	}

	// Parse the root inputs to construct the request payload
	type SubrootEntry struct {
		SubrootCID string `json:"subrootCid"`
	}

	type AddRootRequest struct {
		RootCID  string         `json:"rootCid"`
		Subroots []SubrootEntry `json:"subroots"`
	}

	var addRootRequests []AddRootRequest

	for _, rootInput := range rootInputs {
		// Expected format: rootCID:subrootCID1,subrootCID2,...
		parts := strings.SplitN(rootInput, ":", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid root input format: %s (%d)", rootInput, len(parts))
		}
		rootCID := parts[0]
		subrootsStr := parts[1]
		subrootCIDStrs := strings.Split(subrootsStr, "+")

		if rootCID == "" || len(subrootCIDStrs) == 0 {
			return fmt.Errorf("rootCID and at least one subrootCID are required")
		}

		var subroots []SubrootEntry
		for _, subrootCID := range subrootCIDStrs {
			subroots = append(subroots, SubrootEntry{SubrootCID: subrootCID})
		}

		addRootRequests = append(addRootRequests, AddRootRequest{
			RootCID:  rootCID,
			Subroots: subroots,
		})
	}

	// Construct the full request payload including extraData
	type AddRootsPayload struct {
		Roots     []AddRootRequest `json:"roots"`
		ExtraData *string          `json:"extraData,omitempty"`
	}

	payload := AddRootsPayload{
		Roots: addRootRequests,
	}
	if extraDataHexStr != "" {
		// Pass the validated 0x-prefixed hex string directly
		payload.ExtraData = &extraDataHexStr
	}

	requestBodyBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %v", err)
	}

	// Construct the POST URL
	postURL := fmt.Sprintf("%s/pdp/proof-sets/%d/roots", p.serviceURL, p.proofSetID)

	// Create the POST request
	req, err := http.NewRequestWithContext(ctx, "POST", postURL, bytes.NewBuffer(requestBodyBytes))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+jwtToken)
	req.Header.Set("Content-Type", "application/json")

	// Send the request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close() // nolint:errcheck

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to add roots, status code %d: %s", resp.StatusCode, string(body))
	} else {
		_, _ = io.Copy(io.Discard, resp.Body)
	}

	return nil
}

func preparePiece(r io.ReadSeeker) (pieceCIDComputed cid.Cid, pieceSize int64, paddedPieceSize uint64, digest []byte, err error) {
	// Create commp calculator
	cp := &commp.Calc{}

	// Copy data into commp calculator
	pieceSize, err = io.Copy(cp, r)
	if err != nil {
		err = fmt.Errorf("failed to read input file: %v", err)
		return
	}

	// Finalize digest
	digest, paddedPieceSize, err = cp.Digest()
	if err != nil {
		err = fmt.Errorf("failed to compute digest: %v", err)
		return
	}

	// Convert digest to CID
	pieceCIDComputed, err = commcid.DataCommitmentV1ToCID(digest)
	if err != nil {
		err = fmt.Errorf("failed to compute piece CID: %v", err)
		return
	}

	// now compute sha256
	if _, err = r.Seek(0, io.SeekStart); err != nil {
		err = fmt.Errorf("failed to seek file: %v", err)
		return
	}

	return
}

func createJWTToken(serviceName string, privateKey *ecdsa.PrivateKey) (string, error) {
	// Create JWT claims
	claims := jwt.MapClaims{
		"service_name": serviceName,
		"exp":          time.Now().Add(time.Hour * 2).Unix(),
	}

	// Create the token
	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)

	// Sign the token
	tokenString, err := token.SignedString(privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %v", err)
	}

	return tokenString, nil
}

func uploadOnePiece(ctx context.Context, client *http.Client, serviceURL string, reqBody []byte, jwtToken string, r io.ReadSeeker, pieceSize int64) error {
	req, err := http.NewRequestWithContext(ctx, "POST", serviceURL+"/pdp/piece", bytes.NewReader(reqBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+jwtToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close() // nolint:errcheck

	switch resp.StatusCode {
	case http.StatusOK:
		// Piece already exists, get the pieceCID from the response
		var respData map[string]string
		err = json.NewDecoder(resp.Body).Decode(&respData)
		if err != nil {
			return fmt.Errorf("failed to parse response: %v", err)
		}

		return nil
	case http.StatusCreated:

		// Get the upload URL from the Location header
		uploadURL := resp.Header.Get("Location")
		if uploadURL == "" {
			return fmt.Errorf("server did not provide upload URL in Location header")
		}

		// Upload the piece data via PUT
		if _, err := r.Seek(0, io.SeekStart); err != nil {
			return fmt.Errorf("failed to seek file: %v", err)
		}
		uploadReq, err := http.NewRequest("PUT", serviceURL+uploadURL, r)
		if err != nil {
			return fmt.Errorf("failed to create upload request: %v", err)
		}
		// Set the Content-Length header
		uploadReq.ContentLength = pieceSize
		// Set the Content-Type header
		uploadReq.Header.Set("Content-Type", "application/octet-stream")

		uploadResp, err := client.Do(uploadReq)
		if err != nil {
			return fmt.Errorf("failed to upload piece data: %v", err)
		}
		defer uploadResp.Body.Close() // nolint:errcheck

		if uploadResp.StatusCode != http.StatusNoContent {
			body, _ := io.ReadAll(uploadResp.Body)
			return fmt.Errorf("upload failed with status code %d: %s", uploadResp.StatusCode, string(body))
		}

		return nil
	default:
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("server returned status code %d: %s", resp.StatusCode, string(body))
	}
}

// validateExtraData checks if the provided hex string is valid and within the size limit.
func validateExtraData(extraDataHexStr string) error {
	if extraDataHexStr == "" {
		return nil // No data to validate
	}
	decoded, err := hex.DecodeString(strings.TrimPrefix(extraDataHexStr, "0x"))
	if err != nil {
		return fmt.Errorf("failed to decode hex in extra-data: %w", err)
	}
	if len(decoded) > 2048 {
		return fmt.Errorf("decoded extra-data exceeds maximum size of 2048 bytes (decoded length: %d)", len(decoded))
	}
	return nil
}
