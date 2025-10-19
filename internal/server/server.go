package server

import (
	"context"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/Shuimo03/image-proxy/internal/config"
	"github.com/Shuimo03/image-proxy/internal/logging"
	"github.com/gin-gonic/gin"
)

// Server wraps the Gin engine and HTTP server lifecycle.
type Server struct {
	cfg         *config.Config
	engine      *gin.Engine
	transport   http.RoundTripper
	upstreamURL *url.URL
	logger      *logging.Logger
}

// New sets up routes and middleware.
func New(cfg *config.Config, transport http.RoundTripper, logger *logging.Logger) (*Server, error) {
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	engine.Use(gin.Recovery())

	u, err := url.Parse(cfg.Upstream.URL)
	if err != nil {
		return nil, err
	}

	s := &Server{
		cfg:         cfg,
		engine:      engine,
		transport:   transport,
		upstreamURL: u,
		logger:      logger,
	}

	engine.GET("/healthz", s.healthz)
	engine.Any("/*proxyPath", s.reverseProxyHandler)

	return s, nil
}

func (s *Server) healthz(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (s *Server) reverseProxyHandler(c *gin.Context) {
	reverseProxy := httputil.NewSingleHostReverseProxy(s.upstreamURL)
	reverseProxy.Transport = s.transport
	reverseProxy.FlushInterval = 100 * time.Millisecond

	originalDirector := reverseProxy.Director
	reverseProxy.Director = func(req *http.Request) {
		originalDirector(req)
		req.Host = s.upstreamURL.Host
	}

	reverseProxy.ErrorHandler = func(rw http.ResponseWriter, req *http.Request, err error) {
		s.logger.Error("proxy error", "err", err)
		http.Error(rw, "proxy error", http.StatusBadGateway)
	}

	reverseProxy.ServeHTTP(c.Writer, c.Request)
}

// Start launches the HTTP server and blocks until context cancelled or server errors.
func (s *Server) Start(ctx context.Context) error {
	httpServer := &http.Server{
		Addr:              s.cfg.Proxy.ListenAddr,
		Handler:           s.engine,
		ReadHeaderTimeout: s.cfg.Proxy.ReadHeaderTimeout,
		IdleTimeout:       s.cfg.Proxy.IdleTimeout,
		WriteTimeout:      s.cfg.Proxy.RequestTimeout,
	}
	errCh := make(chan error, 1)

	go func() {
		s.logger.Info("proxy listening", "addr", s.cfg.Proxy.ListenAddr)
		errCh <- httpServer.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		s.logger.Info("shutdown signal received")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return httpServer.Shutdown(shutdownCtx)
	case err := <-errCh:
		return err
	}
}
