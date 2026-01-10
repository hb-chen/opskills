package server

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"

	"github.com/hb-chen/opskills/internal/agent"
	"github.com/hb-chen/opskills/internal/api"
	"github.com/hb-chen/opskills/internal/config"
	"github.com/hb-chen/opskills/pkg/grpc/gateway"
	"github.com/hb-chen/opskills/pkg/logger"
	ops "github.com/hb-chen/opskills/proto/ops"
)

//go:embed web
var WebFS embed.FS

// Serve starts both HTTP and gRPC servers
func Serve(ctx context.Context, cfg *config.Config, pipeline *agent.Pipeline) error {
	// Create gRPC service
	grpcService := api.NewService(pipeline)

	wg := &sync.WaitGroup{}

	// Start gRPC server
	if cfg.Server.GRPC.Addr != "" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := runGRPC(ctx, cfg.Server.GRPC.Addr, grpcService); err != nil {
				logger.Errorf("gRPC server error: %v", err)
			}
		}()
	}

	// Start HTTP server (with grpc-gateway, Web UI and API)
	if cfg.Server.HTTP.Addr != "" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := runHTTP(ctx, cfg.Server.HTTP.Addr, pipeline, grpcService); err != nil {
				logger.Errorf("HTTP server error: %v", err)
			}
		}()
	}

	// Wait for context cancellation
	<-ctx.Done()
	logger.Info("Shutting down servers...")
	wg.Wait()

	return nil
}

// runGRPC starts the gRPC server
func runGRPC(ctx context.Context, addr string, service *api.Service) error {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	s := grpc.NewServer()

	// Register OpsService
	ops.RegisterOpsServiceServer(s, service)

	logger.Infof("gRPC server listening on %s", addr)

	go func() {
		<-ctx.Done()
		logger.Info("Stopping gRPC server...")
		s.GracefulStop()
	}()

	if err := s.Serve(lis); err != nil {
		return fmt.Errorf("gRPC server failed: %w", err)
	}

	return nil
}

// runHTTP starts the HTTP server with grpc-gateway and additional API routes
// HTTP server uses in-process service registration, not requiring gRPC connection
func runHTTP(ctx context.Context, httpAddr string, pipeline *agent.Pipeline, grpcService *api.Service) error {
	// Create gateway with error handler
	gw := gateway.New(
		runtime.WithErrorHandler(httpErrorHandler),
	)

	// Register gRPC service handlers via gateway (in-process, no gRPC connection needed)
	// This automatically registers all routes defined in proto (e.g., /api/v1/tasks)
	if err := ops.RegisterOpsServiceHandlerServer(ctx, gw.Mux(), grpcService); err != nil {
		return fmt.Errorf("failed to register OpsService gateway handlers: %w", err)
	}

	// Create special route handler for routes that cannot be implemented via gRPC/gateway
	// (e.g., SSE streaming, WebSocket)
	specialHandler := api.NewHandler(pipeline)

	// Create main HTTP mux for routing
	mainMux := http.NewServeMux()

	// Register special routes that bypass gateway (e.g., SSE streaming)
	mainMux.HandleFunc("/api/run", specialHandler.HandleRun)

	// Register /health endpoint (simple endpoint, doesn't need gRPC)
	mainMux.HandleFunc("/health", specialHandler.HealthCheck)

	// Mount gateway to root (gateway handles all proto-defined routes)
	mainMux.Handle("/", gw)

	// Serve static web files from embedded filesystem
	var baseHandler http.Handler = mainMux

	// Create filesystem from embedded web directory
	webFileSystem, err := fs.Sub(WebFS, "web")
	if err == nil {
		fileServer := http.FileServer(http.FS(webFileSystem))
		// Create a wrapper handler that checks for web files first, then falls back to main mux
		baseHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			path := r.URL.Path
			// Check if it's an API path or static file path
			if strings.HasPrefix(path, "/api") || strings.HasPrefix(path, "/health") {
				// API paths go to main mux (which routes to gateway for proto-defined routes or special handlers)
				mainMux.ServeHTTP(w, r)
				return
			}
			// Static file paths
			// For SPA routing, check if file exists, otherwise serve index.html
			if path != "/" && path != "/index.html" {
				// Try to open the file to check if it exists
				file, err := webFileSystem.Open(strings.TrimPrefix(path, "/"))
				if err != nil {
					// File doesn't exist, serve index.html for SPA routing
					r.URL.Path = "/index.html"
				} else {
					file.Close()
				}
			}
			fileServer.ServeHTTP(w, r)
		})
		logger.Infof("Serving web files from embedded filesystem")
	} else {
		logger.Warnf("Failed to load embedded web files: %v", err)
	}

	// Wrap with access log middleware
	finalHandler := accessLogMiddleware(baseHandler)

	// Create HTTP server
	srv := &http.Server{
		Addr:    httpAddr,
		Handler: finalHandler,
	}

	// Start listening to get the actual address
	listener, err := net.Listen("tcp", httpAddr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", httpAddr, err)
	}

	// Get the actual address (in case port was 0, which assigns a random port)
	actualAddr := listener.Addr().String()
	httpURL := formatHTTPURL(actualAddr)

	logger.Infof("HTTP server listening on %s", actualAddr)
	logger.Infof("HTTP access URL: %s", httpURL)

	go func() {
		<-ctx.Done()
		logger.Info("Stopping HTTP server...")
		_ = srv.Shutdown(ctx)
	}()

	// Start serving
	if err := srv.Serve(listener); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("HTTP server failed: %w", err)
	}

	return nil
}

// formatHTTPURL formats the address into a complete HTTP URL
func formatHTTPURL(addr string) string {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		// If parsing fails, assume it's just a port
		logger.Warnf("Failed to parse address %s: %v, assuming it's just a port", addr, err)
		return fmt.Sprintf("http://localhost%s", addr)
	}

	// Handle empty host or all interfaces (IPv4 and IPv6)
	// 0.0.0.0 (IPv4 all interfaces) and any host with only colons (IPv6 unspecified address like ::)
	// all mean listen on all interfaces, so we use localhost for the URL
	if host == "0.0.0.0" || strings.Trim(host, ":") == "" {
		host = "localhost"
	}

	return fmt.Sprintf("http://%s:%s", host, port)
}

// httpErrorHandler handles errors from grpc-gateway
func httpErrorHandler(_ context.Context, _ *runtime.ServeMux, marshaler runtime.Marshaler, w http.ResponseWriter, r *http.Request, err error) {
	logger.Errorf("HTTP error: %v, path: %s", err, r.URL.Path)
	// Use default error handling
	runtime.DefaultHTTPErrorHandler(context.Background(), nil, marshaler, w, r, err)
}

// loggerWrapper wraps the logger package functions to implement the Logger interface
type loggerWrapper struct{}

func (l *loggerWrapper) Debug(v ...interface{})                 { logger.Debug(v...) }
func (l *loggerWrapper) Debugf(format string, v ...interface{}) { logger.Debugf(format, v...) }
func (l *loggerWrapper) Info(v ...interface{})                  { logger.Info(v...) }
func (l *loggerWrapper) Infof(format string, v ...interface{})  { logger.Infof(format, v...) }
func (l *loggerWrapper) Warn(v ...interface{})                  { logger.Warn(v...) }
func (l *loggerWrapper) Warnf(format string, v ...interface{})  { logger.Warnf(format, v...) }
func (l *loggerWrapper) Error(v ...interface{})                 { logger.Error(v...) }
func (l *loggerWrapper) Errorf(format string, v ...interface{}) { logger.Errorf(format, v...) }
func (l *loggerWrapper) Fatal(v ...interface{})                 { logger.Fatal(v...) }
func (l *loggerWrapper) Fatalf(format string, v ...interface{}) { logger.Fatalf(format, v...) }

// accessLogMiddleware creates a middleware that logs HTTP access
func accessLogMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a response writer wrapper to capture status code
		// Preserve Flusher interface for SSE streaming support
		rw := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}
		if flusher, ok := w.(http.Flusher); ok {
			rw.flusher = flusher
		}

		// Process request
		next.ServeHTTP(rw, r)

		// Calculate duration
		duration := time.Since(start)

		// Get client IP
		clientIP := r.RemoteAddr
		if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
			clientIP = forwarded
		} else if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
			clientIP = realIP
		}

		// Get user agent and referer
		userAgent := r.UserAgent()
		if userAgent == "" {
			userAgent = "-"
		}
		referer := r.Referer()
		if referer == "" {
			referer = "-"
		}

		// Log access in Apache Common Log Format style
		responseSize := rw.bytesWritten
		if responseSize == 0 {
			responseSize = -1 // Use -1 to indicate no response body
		}
		logger.Infof("%s - \"%s %s %s\" %d %d \"%s\" \"%s\" %v",
			clientIP,
			r.Method,
			r.URL.Path,
			r.Proto,
			rw.statusCode,
			responseSize,
			referer,
			userAgent,
			duration,
		)
	})
}

// responseWriter wraps http.ResponseWriter to capture status code and response size
type responseWriter struct {
	http.ResponseWriter
	statusCode   int
	bytesWritten int64
	flusher      http.Flusher
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.bytesWritten += int64(n)
	return n, err
}

// Flush implements http.Flusher interface for SSE streaming support
func (rw *responseWriter) Flush() {
	if rw.flusher != nil {
		rw.flusher.Flush()
	}
}
