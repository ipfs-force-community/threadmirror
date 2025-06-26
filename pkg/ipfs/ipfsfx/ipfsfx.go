package ipfsfx

import (
	"github.com/ipfs-force-community/threadmirror/pkg/ipfs"
	"go.uber.org/fx"
)

type Config struct {
	// Backend specifies which IPFS backend to use: "kubo" or "pdp"
	Backend string
	// NodeURL is the URL/multiaddr of the IPFS node
	NodeURL string
}

// Module provides the fx module for ipfs
var Module = fx.Module("ipfs",
	fx.Provide(NewStorage),
)

// NewStorage creates a new IPFS Storage instance based on configuration
func NewStorage(config *Config) (ipfs.Storage, error) {
	switch config.Backend {
	case "kubo":
		return ipfs.NewKuboStorage(config.NodeURL)
	default:
		// Default to Kubo storage
		return ipfs.NewKuboStorage(config.NodeURL)
	}
}
