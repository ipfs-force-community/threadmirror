package ipfs

import (
	"context"
	"fmt"
	"io"

	"github.com/ipfs/boxo/files"
	"github.com/ipfs/boxo/path"
	"github.com/ipfs/go-cid"
	"github.com/ipfs/kubo/client/rpc"
	iface "github.com/ipfs/kubo/core/coreiface"
	"github.com/ipfs/kubo/core/coreiface/options"
	"github.com/multiformats/go-multiaddr"
)

// KuboStorage implements Storage interface using IPFS Kubo RPC client
type KuboStorage struct {
	client iface.CoreAPI
}

// NewKuboStorage creates a new KuboStorage instance connected to a Kubo node
func NewKuboStorage(nodeURL string) (*KuboStorage, error) {
	// Parse the multiaddr
	addr, err := multiaddr.NewMultiaddr(nodeURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse node URL: %w", err)
	}

	// Create RPC client
	client, err := rpc.NewApi(addr)
	if err != nil {
		return nil, fmt.Errorf("failed to create IPFS client: %w", err)
	}

	return &KuboStorage{client: client}, nil
}

// Add adds content to IPFS and returns the CID
func (k *KuboStorage) Add(ctx context.Context, content io.ReadSeeker) (cid.Cid, error) {
	// Convert io.Reader to files.Node
	file := files.NewReaderFile(content)

	path, err := k.client.Unixfs().Add(ctx, file, options.Unixfs.Pin(true))
	if err != nil {
		return cid.Undef, fmt.Errorf("failed to add content to IPFS: %w", err)
	}
	return path.RootCid(), nil
}

// Get retrieves content from IPFS by CID
func (k *KuboStorage) Get(ctx context.Context, cid cid.Cid) (io.ReadCloser, error) {
	p := path.FromCid(cid)

	node, err := k.client.Unixfs().Get(ctx, p)
	if err != nil {
		return nil, fmt.Errorf("failed to get content from IPFS: %w", err)
	}

	// Cast to files.File which implements io.ReadCloser
	file, ok := node.(files.File)
	if !ok {
		return nil, fmt.Errorf("node is not a file")
	}

	return file, nil
}
