package comm

import (
	"fmt"

	"github.com/skip2/go-qrcode"
)

func GenQrcode(threadURLTemplate, threadID string) ([]byte, error) {
	qr, err := qrcode.New(fmt.Sprintf(threadURLTemplate, threadID), qrcode.Medium)
	if err != nil {
		return nil, err
	}
	qr.DisableBorder = true
	const qrSize = 200
	return qr.PNG(qrSize)
}
