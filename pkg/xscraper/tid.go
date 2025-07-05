package xscraper

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"
)

type TransactionIDPair struct {
	AnimationKey              string `json:"animationKey"`
	Verification              string `json:"verification"`
	VerificationBase64Decoded []byte `json:"-"`
}

var (
	transactionIDPairs               = []TransactionIDPair{}
	transactionIDPairsMu             sync.RWMutex
	transactionIDPairsUpdateInterval = 30 * time.Minute // 每30分钟更新一次
)

// fetchTransactionIDPairs 从远程获取并解析transaction ID pairs
func fetchTransactionIDPairs() error {
	resp, err := http.Get("https://raw.githubusercontent.com/fa0311/x-client-transaction-id-pair-dict/refs/heads/main/pair.json")
	if err != nil {
		return fmt.Errorf("failed to fetch transaction ID pairs: %w", err)
	}
	defer resp.Body.Close() // nolint:errcheck

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to fetch transaction ID pairs: status code %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	var newPairs []TransactionIDPair
	if err := json.Unmarshal(body, &newPairs); err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	if len(newPairs) == 0 {
		return fmt.Errorf("no transaction ID pairs found in the response")
	}

	// 解码base64数据
	for i := range newPairs {
		newPairs[i].VerificationBase64Decoded, err = base64.StdEncoding.DecodeString(newPairs[i].Verification)
		if err != nil {
			return fmt.Errorf("failed to decode base64 verification %s: %w", newPairs[i].AnimationKey, err)
		}
	}

	// 更新全局变量（线程安全）
	transactionIDPairsMu.Lock()
	transactionIDPairs = newPairs
	transactionIDPairsMu.Unlock()

	return nil
}

// updateTransactionIDPairs 定时更新transaction ID pairs
func updateTransactionIDPairs() {
	ticker := time.NewTicker(transactionIDPairsUpdateInterval)
	defer ticker.Stop()

	for range ticker.C {
		if err := fetchTransactionIDPairs(); err != nil {
			log.Printf("Failed to update transaction ID pairs: %v", err)
		}
	}
}

func init() {
	// 初始化时获取数据
	if err := fetchTransactionIDPairs(); err != nil {
		panic(err)
	}

	// 启动定时更新goroutine
	go updateTransactionIDPairs()
}

func xClientTransactionID(method, path string) string {
	transactionIDPairsMu.RLock()
	if len(transactionIDPairs) == 0 {
		transactionIDPairsMu.RUnlock()
		panic("no transaction ID pairs available")
	}
	pair := transactionIDPairs[rand.Intn(len(transactionIDPairs))]
	transactionIDPairsMu.RUnlock()

	return generateTransactionId(method, path, pair.VerificationBase64Decoded, pair.AnimationKey)
}

// generateTransactionId generates a unique transaction ID.
func generateTransactionId(method, path string, key []byte, animationKey string) string {
	const DEFAULT_KEYWORD = "obfiowerehiring"
	const ADDITIONAL_RANDOM_NUMBER byte = 3

	timeNow := time.Now().Unix() - 1682924400
	timeNowBytes := []byte{
		byte(timeNow & 0xff),
		byte((timeNow >> 8) & 0xff),
		byte((timeNow >> 16) & 0xff),
		byte((timeNow >> 24) & 0xff),
	}
	data := fmt.Sprintf("%s!%s!%d%s%s", method, path, timeNow, DEFAULT_KEYWORD, animationKey)
	h := sha256.New()
	h.Write([]byte(data))
	hashBytes := h.Sum(nil)

	randomNum := byte(rand.Intn(256))
	randomXor := func(b []byte) []byte {
		for i := range b {
			b[i] ^= randomNum
		}
		return b
	}

	var bytesArr []byte
	bytesArr = append(bytesArr, randomNum)
	bytesArr = append(bytesArr, randomXor(key)...)
	bytesArr = append(bytesArr, randomXor(timeNowBytes)...)
	bytesArr = append(bytesArr, randomXor(hashBytes[:16])...)
	bytesArr = append(bytesArr, ADDITIONAL_RANDOM_NUMBER^randomNum)

	base64Encoded := strings.TrimRightFunc(base64.StdEncoding.EncodeToString(bytesArr), func(c rune) bool { return c == '=' })
	return base64Encoded
}
