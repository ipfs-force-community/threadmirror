package config

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/ipfs-force-community/threadmirror/pkg/ipfs/ipfsfx"
	"github.com/ipfs-force-community/threadmirror/pkg/llm/llmfx"
	"github.com/urfave/cli/v2"
)

type CommonConfig struct {
	ThreadURLTemplate string
	Debug             bool
	ThreadMaxRetries  int
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Addr         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration

	AllowedOrigins []string // 允许的跨域Origin
}

type DatabaseConfig struct {
	// postgres, sqlite
	Driver string
	DSN    string
}

// RedisConfig holds Redis configuration for Asynq
type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

// Auth0Config holds Auth0-related configuration
type Auth0Config struct {
	Domain   string
	Audience string
}

// BotCredential represents a set of credentials for one Twitter bot account.
type BotCredential struct {
	Username          string `json:"username"`
	Password          string `json:"password"`
	Email             string `json:"email"`
	APIKey            string `json:"api_key"`
	APIKeySecret      string `json:"api_key_secret"`
	AccessToken       string `json:"access_token"`
	AccessTokenSecret string `json:"access_token_secret"`
}

// CronConfig holds cron-related configuration
type CronConfig struct {
	// Thread status cleanup configuration
	ThreadStatusCleanup struct {
		EnabledIntervalMinutes int
		ScrapingTimeoutMinutes int
		PendingTimeoutMinutes  int
		RetryDelayMinutes      int
		MaxRetries             int
	}

	// Mention check configuration
	MentionCheck struct {
		EnabledIntervalMinutes     int
		RandomizeIntervalMinutes   int
		ExcludeMentionAuthorPrefix string
		MentionUsername            string
	}
}

// BotConfig holds Twitter bot configuration
type BotConfig struct {
	Enable bool
	// Multiple Twitter credentials (supporting multiple bots)
	Credentials []BotCredential

	// Bot behavior settings
	CheckInterval time.Duration

	// Prefix of author screen names to exclude when scanning mentions
	ExcludeMentionAuthorPrefix string

	// Username to monitor for mentions (if empty, uses first credential's username)
	MentionUsername string
}

func LoadCommonConfigFromCLI(c *cli.Context) *CommonConfig {
	return &CommonConfig{
		ThreadURLTemplate: c.String("thread-url-template"),
		Debug:             c.Bool("debug"),
		ThreadMaxRetries:  c.Int("thread-max-retries"),
	}
}

func LoadServerConfigFromCLI(c *cli.Context) *ServerConfig {
	return &ServerConfig{
		Addr:           c.String("server-addr"),
		ReadTimeout:    c.Duration("server-read-timeout"),
		WriteTimeout:   c.Duration("server-write-timeout"),
		AllowedOrigins: c.StringSlice("cors-allowed-origins"),
	}
}

func LoadDatabaseConfigFromCLI(c *cli.Context) *DatabaseConfig {
	return &DatabaseConfig{
		Driver: c.String("db-driver"),
		DSN:    c.String("db-dsn"),
	}
}

func LoadRedisConfigFromCLI(c *cli.Context) *RedisConfig {
	return &RedisConfig{
		Addr:     c.String("redis-addr"),
		Password: c.String("redis-password"),
		DB:       c.Int("redis-db"),
	}
}

func LoadAuth0ConfigFromCLI(c *cli.Context) *Auth0Config {
	return &Auth0Config{
		Domain:   c.String("auth0-domain"),
		Audience: c.String("auth0-audience"),
	}
}

func LoadCronConfigFromCLI(c *cli.Context) *CronConfig {
	return &CronConfig{
		ThreadStatusCleanup: struct {
			EnabledIntervalMinutes int
			ScrapingTimeoutMinutes int
			PendingTimeoutMinutes  int
			RetryDelayMinutes      int
			MaxRetries             int
		}{
			EnabledIntervalMinutes: c.Int("thread-cleanup-interval-minutes"),
			ScrapingTimeoutMinutes: c.Int("thread-scraping-timeout-minutes"),
			PendingTimeoutMinutes:  c.Int("thread-pending-timeout-minutes"),
			RetryDelayMinutes:      c.Int("thread-retry-delay-minutes"),
			MaxRetries:             c.Int("thread-max-retries"),
		},
		MentionCheck: struct {
			EnabledIntervalMinutes     int
			RandomizeIntervalMinutes   int
			ExcludeMentionAuthorPrefix string
			MentionUsername            string
		}{
			EnabledIntervalMinutes:     c.Int("mention-check-interval-minutes"),
			RandomizeIntervalMinutes:   c.Int("mention-check-randomize-interval-minutes"),
			ExcludeMentionAuthorPrefix: c.String("mention-check-exclude-author-prefix"),
			MentionUsername:            c.String("mention-check-username"),
		},
	}
}

func LoadBotConfigFromCLI(c *cli.Context) *BotConfig {
	var creds []BotCredential

	// Prefer JSON-based credentials if provided
	if credJSON := c.String("bot-credentials"); credJSON != "" {
		if err := json.Unmarshal([]byte(credJSON), &creds); err != nil {
			panic(fmt.Errorf("invalid bot-credentials JSON: %w", err))
		}
	}

	// Fallback to single credential flags when JSON not supplied or empty
	if len(creds) == 0 {
		creds = []BotCredential{{
			Username:          c.String("bot-username"),
			Password:          c.String("bot-password"),
			Email:             c.String("bot-email"),
			APIKey:            c.String("bot-api-key"),
			APIKeySecret:      c.String("bot-api-key-secret"),
			AccessToken:       c.String("bot-access-token"),
			AccessTokenSecret: c.String("bot-access-token-secret"),
		}}
	}

	return &BotConfig{
		Enable:                     c.Bool("bot-enable"),
		Credentials:                creds,
		CheckInterval:              c.Duration("bot-check-interval"),
		ExcludeMentionAuthorPrefix: c.String("bot-exclude-mention-author-prefix"),
		MentionUsername:            c.String("bot-mention-username"),
	}
}

func LoadIPFSConfigFromCLI(c *cli.Context) *ipfsfx.Config {
	backend := c.String("ipfs-backend")

	config := &ipfsfx.Config{
		Backend: backend,
	}

	// Configure backend-specific settings
	switch backend {
	case "kubo":
		config.Kubo = &ipfsfx.KuboConfig{
			NodeURL: c.String("ipfs-kubo-node-url"),
		}
	case "pdp":
		config.PDP = &ipfsfx.PDPConfig{
			ServiceURL:  c.String("ipfs-pdp-service-url"),
			ServiceName: c.String("ipfs-pdp-service-name"),
			PrivateKey:  c.String("ipfs-pdp-private-key"),
			ProofSetID:  c.Uint64("ipfs-pdp-proof-set-id"),
		}
	default:
		// Default to kubo if backend is not recognized
		config.Backend = "kubo"
		config.Kubo = &ipfsfx.KuboConfig{
			NodeURL: c.String("ipfs-kubo-node-url"),
		}
	}

	return config
}

func GetCommonCLIFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:    "thread-url-template",
			Usage:   "Thread URL template, e.g. https://threadmirror.xyz/thread/%s",
			EnvVars: []string{"THREAD_URL_TEMPLATE"},
			Value:   "https://threadmirror.xyz/thread/%s",
		},
		&cli.IntFlag{
			Name:    "thread-max-retries",
			Usage:   "Maximum number of retries for thread status updates",
			EnvVars: []string{"THREAD_MAX_RETRIES"},
			Value:   5,
		},
	}
}

// GetServerCLIFlags returns server-related CLI flags
func GetServerCLIFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:    "server-addr",
			Value:   "localhost:8080",
			Usage:   "Server host address",
			EnvVars: []string{"SERVER_ADDR"},
		},
		&cli.DurationFlag{
			Name:    "server-read-timeout",
			Value:   30 * time.Second,
			Usage:   "Server read timeout in seconds",
			EnvVars: []string{"SERVER_READ_TIMEOUT"},
		},
		&cli.DurationFlag{
			Name:    "server-write-timeout",
			Value:   30 * time.Second,
			Usage:   "Server write timeout in seconds",
			EnvVars: []string{"SERVER_WRITE_TIMEOUT"},
		},
		&cli.StringSliceFlag{
			Name:    "cors-allowed-origins",
			Usage:   "CORS allowed origins (comma separated)",
			EnvVars: []string{"CORS_ALLOWED_ORIGINS"},
			Value:   cli.NewStringSlice("https://threadmirror.xyz"),
		},
	}
}

// GetDatabaseCLIFlags returns only database-related CLI flags
func GetDatabaseCLIFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:    "db-driver",
			Value:   "sqlite",
			Usage:   "Database driver (postgres, sqlite)",
			EnvVars: []string{"DB_DRIVER"},
		},
		&cli.StringFlag{
			Name:    "db-dsn",
			Value:   "file::memory:?cache=shared",
			Usage:   "Database connection string",
			EnvVars: []string{"DB_DSN"},
		},
	}
}

// GetRedisCLIFlags returns Redis-related CLI flags
func GetRedisCLIFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:    "redis-addr",
			Value:   "localhost:6379",
			Usage:   "Redis server address",
			EnvVars: []string{"REDIS_ADDR"},
		},
		&cli.StringFlag{
			Name:    "redis-password",
			Value:   "",
			Usage:   "Redis password",
			EnvVars: []string{"REDIS_PASSWORD"},
		},
		&cli.IntFlag{
			Name:    "redis-db",
			Value:   0,
			Usage:   "Redis database number",
			EnvVars: []string{"REDIS_DB"},
		},
	}
}

// GetAuth0CLIFlags returns Auth0-related CLI flags
func GetAuth0CLIFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:    "auth0-domain",
			Usage:   "Auth0 domain",
			EnvVars: []string{"AUTH0_DOMAIN"},
		},
		&cli.StringFlag{
			Name:    "auth0-audience",
			Usage:   "Auth0 API audience",
			EnvVars: []string{"AUTH0_AUDIENCE"},
		},
	}
}

// GetCronCLIFlags returns cron-related CLI flags
func GetCronCLIFlags() []cli.Flag {
	return []cli.Flag{
		&cli.IntFlag{
			Name:    "thread-cleanup-interval-minutes",
			Value:   3,
			Usage:   "Interval in minutes for cleaning up thread statuses",
			EnvVars: []string{"THREAD_CLEANUP_INTERVAL_MINUTES"},
		},
		&cli.IntFlag{
			Name:    "thread-scraping-timeout-minutes",
			Value:   30,
			Usage:   "Timeout in minutes for scraping thread statuses",
			EnvVars: []string{"THREAD_SCRAPING_TIMEOUT_MINUTES"},
		},
		&cli.IntFlag{
			Name:    "thread-pending-timeout-minutes",
			Value:   60,
			Usage:   "Timeout in minutes for pending thread statuses",
			EnvVars: []string{"THREAD_PENDING_TIMEOUT_MINUTES"},
		},
		&cli.IntFlag{
			Name:    "thread-retry-delay-minutes",
			Value:   15,
			Usage:   "Delay in minutes before retrying failed thread status updates",
			EnvVars: []string{"THREAD_RETRY_DELAY_MINUTES"},
		},
		&cli.IntFlag{
			Name:    "thread-max-retries",
			Value:   5,
			Usage:   "Maximum number of retries for failed thread status updates",
			EnvVars: []string{"THREAD_MAX_RETRIES"},
		},
		&cli.IntFlag{
			Name:    "mention-check-interval-minutes",
			Value:   2,
			Usage:   "Base interval in minutes for checking mentions",
			EnvVars: []string{"MENTION_CHECK_INTERVAL_MINUTES"},
		},
		&cli.IntFlag{
			Name:    "mention-check-randomize-interval-minutes",
			Value:   1,
			Usage:   "Randomization range in minutes for mention check interval (+/-)",
			EnvVars: []string{"MENTION_CHECK_RANDOMIZE_INTERVAL_MINUTES"},
		},
		&cli.StringFlag{
			Name:    "mention-check-exclude-author-prefix",
			Value:   "threadmirror",
			Usage:   "Prefix of author screen names to exclude from mention processing",
			EnvVars: []string{"MENTION_CHECK_EXCLUDE_AUTHOR_PREFIX"},
		},
		&cli.StringFlag{
			Name:    "mention-check-username",
			Value:   "threadmirror",
			Usage:   "Username to monitor for mentions (if empty, uses first credential's username)",
			EnvVars: []string{"MENTION_CHECK_USERNAME"},
		},
	}
}

// GetBotCLIFlags returns bot-related CLI flags
func GetBotCLIFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:    "bot-credentials",
			Usage:   "JSON array of bot credentials (overrides individual BOT_* flags)",
			EnvVars: []string{"BOT_CREDENTIALS"},
		},
		&cli.BoolFlag{
			Name:    "bot-enable",
			Usage:   "Enable the bot",
			EnvVars: []string{"BOT_ENABLE"},
			Value:   true,
		},
		&cli.StringFlag{
			Name:    "bot-username",
			Usage:   "Twitter bot username",
			EnvVars: []string{"BOT_USERNAME"},
		},
		&cli.StringFlag{
			Name:    "bot-password",
			Usage:   "Twitter bot password",
			EnvVars: []string{"BOT_PASSWORD"},
		},
		&cli.StringFlag{
			Name:    "bot-email",
			Usage:   "Twitter bot email",
			EnvVars: []string{"BOT_EMAIL"},
		},
		&cli.StringFlag{
			Name:    "bot-api-key",
			Usage:   "Twitter API key",
			EnvVars: []string{"BOT_API_KEY"},
		},
		&cli.StringFlag{
			Name:    "bot-api-key-secret",
			Usage:   "Twitter API key secret",
			EnvVars: []string{"BOT_API_KEY_SECRET"},
		},
		&cli.StringFlag{
			Name:    "bot-access-token",
			Usage:   "Twitter access token",
			EnvVars: []string{"BOT_ACCESS_TOKEN"},
		},
		&cli.StringFlag{
			Name:    "bot-access-token-secret",
			Usage:   "Twitter access token secret",
			EnvVars: []string{"BOT_ACCESS_TOKEN_SECRET"},
		},
		&cli.DurationFlag{
			Name:    "bot-check-interval",
			Value:   5 * time.Minute,
			Usage:   "Interval to check for new mentions",
			EnvVars: []string{"BOT_CHECK_INTERVAL"},
		},
		&cli.StringFlag{
			Name:    "bot-exclude-mention-author-prefix",
			Value:   "threadmirror",
			Usage:   "Prefix of author screen names to exclude from mention processing",
			EnvVars: []string{"BOT_EXCLUDE_MENTION_AUTHOR_PREFIX"},
		},
		&cli.StringFlag{
			Name:    "bot-mention-username",
			Usage:   "Username to monitor for mentions (if empty, uses first credential's username)",
			EnvVars: []string{"BOT_MENTION_USERNAME"},
		},
	}
}

func LoadLLMConfigFromCLI(c *cli.Context) *llmfx.Config {
	return &llmfx.Config{
		OpenAIBaseURL: c.String("openai-base-url"),
		OpenAIAPIKey:  c.String("openai-api-key"),
		OpenAIModel:   c.String("openai-model"),
	}
}

// GetLLMCLIFlags returns LLM-related CLI flags
func GetLLMCLIFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:    "openai-base-url",
			Value:   "https://api.openai.com/v1",
			Usage:   "OpenAI API base URL",
			EnvVars: []string{"OPENAI_BASE_URL"},
		},
		&cli.StringFlag{
			Name:    "openai-api-key",
			Usage:   "OpenAI API key",
			EnvVars: []string{"OPENAI_API_KEY"},
		},
		&cli.StringFlag{
			Name:    "openai-model",
			Value:   "gpt-4o-mini",
			Usage:   "OpenAI model name",
			EnvVars: []string{"OPENAI_MODEL"},
		},
	}
}

// GetIPFSCLIFlags returns IPFS-related CLI flags
func GetIPFSCLIFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:    "ipfs-backend",
			Value:   "kubo",
			Usage:   "IPFS backend (kubo, pdp)",
			EnvVars: []string{"IPFS_BACKEND"},
		},
		// Kubo backend flags
		&cli.StringFlag{
			Name:    "ipfs-kubo-node-url",
			Value:   "/ip4/127.0.0.1/tcp/5001",
			Usage:   "Kubo IPFS node URL/multiaddr",
			EnvVars: []string{"IPFS_KUBO_NODE_URL"},
		},
		// PDP backend flags
		&cli.StringFlag{
			Name:    "ipfs-pdp-service-url",
			Usage:   "PDP service URL",
			EnvVars: []string{"IPFS_PDP_SERVICE_URL"},
		},
		&cli.StringFlag{
			Name:    "ipfs-pdp-service-name",
			Usage:   "PDP service name",
			EnvVars: []string{"IPFS_PDP_SERVICE_NAME"},
		},
		&cli.StringFlag{
			Name:    "ipfs-pdp-private-key",
			Usage:   "PDP private key (PEM format)",
			EnvVars: []string{"IPFS_PDP_PRIVATE_KEY"},
		},
		&cli.Uint64Flag{
			Name:    "ipfs-pdp-proof-set-id",
			Usage:   "PDP proof set ID",
			EnvVars: []string{"IPFS_PDP_PROOF_SET_ID"},
		},
	}
}
