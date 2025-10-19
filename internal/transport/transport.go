package transport

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/Shuimo03/image-proxy/internal/config"
	"golang.org/x/net/proxy"
)

// New creates an HTTP transport that dials upstream through Clash according to the configuration.
func New(cfg *config.Config) (*http.Transport, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is nil")
	}

	baseDialer := &net.Dialer{
		Timeout:   cfg.Proxy.RequestTimeout,
		KeepAlive: 30 * time.Second,
	}

	transport := &http.Transport{
		DialContext:           baseDialer.DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       cfg.Proxy.IdleTimeout,
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: cfg.Proxy.RequestTimeout,
		ExpectContinueTimeout: 1 * time.Second,
	}

	switch cfg.Clash.Mode {
	case "http", "https":
		proxyURL := &url.URL{Scheme: cfg.Clash.Mode, Host: fmt.Sprintf("%s:%d", cfg.Clash.Host, cfg.Clash.Port)}
		transport.Proxy = http.ProxyURL(proxyURL)
	case "socks5":
		proxyURL := &url.URL{Scheme: "socks5", Host: fmt.Sprintf("%s:%d", cfg.Clash.Host, cfg.Clash.Port)}
		dialer, err := proxy.FromURL(proxyURL, proxy.Direct)
		if err != nil {
			return nil, fmt.Errorf("create socks5 dialer: %w", err)
		}

		transport.Proxy = nil
		transport.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
			if cd, ok := dialer.(proxy.ContextDialer); ok {
				return cd.DialContext(ctx, network, addr)
			}
			return dialer.Dial(network, addr)
		}
	default:
		return nil, fmt.Errorf("unsupported clash.mode: %s", cfg.Clash.Mode)
	}

	return transport, nil
}
