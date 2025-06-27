package ipfsfx

import (
	"fmt"

	"github.com/ipfs-force-community/threadmirror/pkg/ipfs"
	"go.uber.org/fx"
)

// BackendConfig defines the interface for backend-specific configurations
type BackendConfig interface {
	GetBackend() string
	Validate() error
}

// KuboConfig represents configuration for Kubo IPFS backend
type KuboConfig struct {
	NodeURL string `json:"node_url" yaml:"node_url"`
}

func (k *KuboConfig) GetBackend() string {
	return "kubo"
}

func (k *KuboConfig) Validate() error {
	if k.NodeURL == "" {
		return fmt.Errorf("node_url is required for kubo backend")
	}
	return nil
}

// PDPConfig represents configuration for PDP IPFS backend
type PDPConfig struct {
	ServiceURL  string `json:"service_url" yaml:"service_url"`
	ServiceName string `json:"service_name" yaml:"service_name"`
	PrivateKey  string `json:"private_key" yaml:"private_key"`
}

func (p *PDPConfig) GetBackend() string {
	return "pdp"
}

func (p *PDPConfig) Validate() error {
	if p.ServiceURL == "" {
		return fmt.Errorf("service_url is required for pdp backend")
	}
	if p.ServiceName == "" {
		return fmt.Errorf("service_name is required for pdp backend")
	}
	if p.PrivateKey == "" {
		return fmt.Errorf("private_key is required for pdp backend")
	}
	return nil
}

// Config represents the main configuration that wraps backend-specific configs
type Config struct {
	Backend string      `json:"backend" yaml:"backend"`
	Kubo    *KuboConfig `json:"kubo,omitempty" yaml:"kubo,omitempty"`
	PDP     *PDPConfig  `json:"pdp,omitempty" yaml:"pdp,omitempty"`
}

// GetBackendConfig returns the appropriate backend configuration
func (c *Config) GetBackendConfig() (BackendConfig, error) {
	switch c.Backend {
	case "kubo":
		if c.Kubo == nil {
			return nil, fmt.Errorf("kubo configuration is required when backend is 'kubo'")
		}
		return c.Kubo, c.Kubo.Validate()
	case "pdp":
		if c.PDP == nil {
			return nil, fmt.Errorf("pdp configuration is required when backend is 'pdp'")
		}
		return c.PDP, c.PDP.Validate()
	default:
		return nil, fmt.Errorf("unsupported backend: %s, supported backends: kubo, pdp", c.Backend)
	}
}

// Module provides the fx module for ipfs
var Module = fx.Module("ipfs",
	fx.Provide(NewStorage),
)

// NewStorage creates a new IPFS Storage instance based on configuration
func NewStorage(config *Config) (ipfs.Storage, error) {
	backendConfig, err := config.GetBackendConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get backend config: %w", err)
	}

	switch backendConfig.GetBackend() {
	case "kubo":
		kuboConfig := backendConfig.(*KuboConfig)
		return ipfs.NewKuboStorage(kuboConfig.NodeURL)
	case "pdp":
		pdpConfig := backendConfig.(*PDPConfig)
		return ipfs.NewPDP(pdpConfig.ServiceURL, pdpConfig.ServiceName, pdpConfig.PrivateKey)
	default:
		return nil, fmt.Errorf("unsupported backend: %s", backendConfig.GetBackend())
	}
}
