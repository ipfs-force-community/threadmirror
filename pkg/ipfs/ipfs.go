package ipfs

import (
	"context"
	"io"

	"github.com/ipfs/go-cid"
)

// Storage defines the interface for IPFS storage backends
type Storage interface {
	// Add adds content to IPFS and returns the CID
	Add(ctx context.Context, content io.ReadSeeker) (cid.Cid, error)
	// Get retrieves content from IPFS by CID
	Get(ctx context.Context, cid cid.Cid) (io.ReadCloser, error)
}
