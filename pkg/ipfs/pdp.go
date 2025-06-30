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
	"net/http"
	"os"
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
}

func NewPDP(serviceURL, serviceName, privateKey string) (*PDP, error) {
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

	return &PDP{serviceURL: serviceURL, serviceName: serviceName, privateKey: ecdsaPrivKey}, nil
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
	checkData := map[string]interface{}{
		"name": "sha2-256-trunc254-padded",
		"hash": hashHex,
		"size": pieceSize,
	}

	// Prepare the request data
	reqData := map[string]interface{}{
		"check": checkData,
	}

	reqBody, err := json.Marshal(reqData)
	if err != nil {
		return cid.Undef, fmt.Errorf("failed to marshal request data: %v", err)
	}
	if err := uploadOnePiece(http.DefaultClient, p.serviceURL, reqBody, jwtToken, content, pieceSize); err != nil {
		return cid.Undef, fmt.Errorf("failed to upload piece: %v", err)
	}

	return pieceCIDComputed, nil
}

func (p *PDP) Get(ctx context.Context, cid cid.Cid) (io.ReadCloser, error) {
	// Create the download URL
	downloadURL := fmt.Sprintf("%s/piece/%s", p.serviceURL, cid.String())

	// Create the GET request
	req, err := http.NewRequest("GET", downloadURL, nil)
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

func uploadOnePiece(client *http.Client, serviceURL string, reqBody []byte, jwtToken string, r io.ReadSeeker, pieceSize int64) error {
	req, err := http.NewRequest("POST", serviceURL+"/pdp/piece", bytes.NewReader(reqBody))
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
