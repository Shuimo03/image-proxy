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
	ListenAddr        string        `toml:"listen_addr"`
	ReadHeaderTimeout time.Duration `toml:"read_header_timeout"`
	RequestTimeout    time.Duration `toml:"request_timeout"`
	IdleTimeout       time.Duration `toml:"idle_timeout"`
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
	if c.Proxy.ReadHeaderTimeout == 0 {
		c.Proxy.ReadHeaderTimeout = 5 * time.Second
	}
	if c.Proxy.RequestTimeout == 0 {
		c.Proxy.RequestTimeout = 60 * time.Second
	}
	if c.Proxy.IdleTimeout == 0 {
		c.Proxy.IdleTimeout = 90 * time.Second
	}
	return nil
}
