package config

import (
	"time"

	"github.com/urfave/cli/v2"
)

// ServerConfig holds server configuration
type ServerConfig struct {
	Addr         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	Debug        bool
}

// SupabaseConfig holds Supabase-related configuration
type SupabaseConfig struct {
	ProjectReference string
	ApiAnnoKey       string
	BucketNames      SupabaseBucketNames
	JWTKey           string
}

// SupabaseBucketNames holds bucket names configuration
type SupabaseBucketNames struct {
	PostImages string
}

type DatabaseConfig struct {
	// postgres, sqlite
	Driver string
	DSN    string
}

func LoadServerConfigFromCLI(c *cli.Context) *ServerConfig {
	return &ServerConfig{
		Addr:         c.String("server-addr"),
		ReadTimeout:  c.Duration("server-read-timeout"),
		WriteTimeout: c.Duration("server-write-timeout"),
		Debug:        c.Bool("debug"),
	}
}

func LoadDatabaseConfigFromCLI(c *cli.Context) *DatabaseConfig {
	return &DatabaseConfig{
		Driver: c.String("db-driver"),
		DSN:    c.String("db-dsn"),
	}
}

func LoadSupabaseConfigFromCLI(c *cli.Context) *SupabaseConfig {
	return &SupabaseConfig{
		ProjectReference: c.String("supabase-project-ref"),
		ApiAnnoKey:       c.String("supabase-api-anno-key"),
		JWTKey:           c.String("subpabase-jwt-key"),
		BucketNames: SupabaseBucketNames{
			PostImages: c.String("supabase-bucket-post-images"),
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

// GetSupabaseCLIFlags returns Supabase-related CLI flags
func GetSupabaseCLIFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:    "supabase-project-ref",
			Usage:   "Supabase project reference",
			EnvVars: []string{"SUPABASE_PROJECT_REF"},
		},
		&cli.StringFlag{
			Name:    "supabase-api-anno-key",
			Usage:   "Supabase API anonymous key",
			EnvVars: []string{"SUPABASE_API_ANNO_KEY"},
		},
		&cli.StringFlag{
			Name:    "subpabase-jwt-key",
			Usage:   "Supabase JWT key",
			EnvVars: []string{"SUPABASE_JWT_KEY"},
		},
		&cli.StringFlag{
			Name:    "supabase-bucket-post-images",
			Value:   "post-images",
			Usage:   "Supabase post images bucket name",
			EnvVars: []string{"SUPABASE_BUCKET_POST_IMAGES"},
		},
	}
}
