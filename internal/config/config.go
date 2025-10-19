package config

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/pelletier/go-toml/v2"
)

// Config represents the complete proxy configuration loaded from TOML.
type Config struct {
	Proxy    ProxyConfig    `toml:"proxy"`
	Upstream UpstreamConfig `toml:"upstream"`
	Clash    ClashConfig    `toml:"clash"`
}

// ProxyConfig contains listener and timeout settings for the proxy server.
type ProxyConfig struct {
	ListenAddr        string   `toml:"listen_addr"`
	ReadHeaderTimeout Duration `toml:"read_header_timeout"`
	RequestTimeout    Duration `toml:"request_timeout"`
	IdleTimeout       Duration `toml:"idle_timeout"`
}

// UpstreamConfig defines the target registry or HTTP service to forward requests to.
type UpstreamConfig struct {
	URL string `toml:"url"`
}

// ClashConfig holds outbound proxy configuration that the Go proxy should dial through.
type ClashConfig struct {
	Mode string `toml:"mode"`
	Host string `toml:"host"`
	Port int    `toml:"port"`
}

// Load reads configuration from path.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	return &cfg, nil
}

// Validate ensures the configuration has the required values.
func (c *Config) Validate() error {
	if c == nil {
		return errors.New("config is nil")
	}
	if c.Proxy.ListenAddr == "" {
		return errors.New("proxy.listen_addr is required")
	}
	if c.Upstream.URL == "" {
		return errors.New("upstream.url is required")
	}
	if c.Clash.Mode == "" {
		c.Clash.Mode = "http"
	}
	if c.Clash.Host == "" || c.Clash.Port == 0 {
		return errors.New("clash.host and clash.port are required")
	}
	if c.Proxy.ReadHeaderTimeout.Duration == 0 {
		c.Proxy.ReadHeaderTimeout.Duration = 5 * time.Second
	}
	if c.Proxy.RequestTimeout.Duration == 0 {
		c.Proxy.RequestTimeout.Duration = 60 * time.Second
	}
	if c.Proxy.IdleTimeout.Duration == 0 {
		c.Proxy.IdleTimeout.Duration = 90 * time.Second
	}
	return nil
}

// Duration wraps time.Duration to allow decoding human-readable strings from TOML.
type Duration struct {
	time.Duration
}

// UnmarshalText parses values like "5s" or "1m30s" into a Duration.
func (d *Duration) UnmarshalText(text []byte) error {
	// Treat empty values as zero without error.
	if len(text) == 0 {
		d.Duration = 0
		return nil
	}
	parsed, err := time.ParseDuration(string(text))
	if err != nil {
		return fmt.Errorf("parse duration: %w", err)
	}
	d.Duration = parsed
	return nil
}

// MarshalText allows Duration to be encoded back into TOML if needed.
func (d Duration) MarshalText() ([]byte, error) {
	return []byte(d.Duration.String()), nil
}
