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

	// "google.golang.org/grpc/credentials/insecure" // Will be used when registering services

	"github.com/hb-chen/opskills/internal/agent"
	httpapi "github.com/hb-chen/opskills/internal/api/http"
	"github.com/hb-chen/opskills/internal/config"
	"github.com/hb-chen/opskills/pkg/grpc/gateway"
	"github.com/hb-chen/opskills/pkg/logger"
)

//go:embed web
var WebFS embed.FS

// Serve starts both HTTP and gRPC servers
func Serve(ctx context.Context, cfg *config.Config, pipeline *agent.Pipeline) error {
	wg := &sync.WaitGroup{}

	// Start gRPC server
	if cfg.Server.GRPC.Addr != "" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := runGRPC(ctx, cfg.Server.GRPC.Addr); err != nil {
				logger.Errorf("gRPC server error: %v", err)
			}
		}()
	}

	// Start HTTP server (with grpc-gateway, Web UI and API)
	if cfg.Server.HTTP.Addr != "" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := runHTTP(ctx, cfg.Server.HTTP.Addr, cfg.Server.GRPC.Addr, pipeline); err != nil {
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
func runGRPC(ctx context.Context, addr string) error {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	s := grpc.NewServer()

	// Register services here
	// For now, this is a placeholder
	// TODO: Register OpsService

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
func runHTTP(ctx context.Context, httpAddr, grpcAddr string, pipeline *agent.Pipeline) error {
	// Create gateway with error handler
	gw := gateway.New(
		runtime.WithErrorHandler(httpErrorHandler),
	)

	// Connect to gRPC server
	// opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	// Register gateway handlers
	// For now, this is a placeholder
	// TODO: Register OpsService gateway handlers
	// err := ops.RegisterOpsServiceHandlerFromEndpoint(ctx, gw.Mux(), grpcAddr, opts)
	// if err != nil {
	// 	return fmt.Errorf("failed to register gateway: %w", err)
	// }

	// Create HTTP handlers for Web UI and SSE API
	logWrapper := &loggerWrapper{}
	handlers := httpapi.NewHandlers(pipeline, logWrapper)

	// Helper function to convert standard http.HandlerFunc to gateway HandlerFunc
	toGatewayHandler := func(fn func(http.ResponseWriter, *http.Request)) runtime.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
			fn(w, r)
		}
	}

	// Register non-gRPC API routes using gateway Group
	apiGroup := gw.Group("/api")
	apiGroup.POST("/run", toGatewayHandler(handlers.HandleRun))
	apiGroup.GET("/run", toGatewayHandler(handlers.HandleRun))

	// Register /api/v1/tasks
	v1Group := gw.Group("/api/v1")
	v1Group.POST("/tasks", toGatewayHandler(handlers.SubmitTask))
	v1Group.GET("/tasks", toGatewayHandler(handlers.GetTaskStatus))

	// Register /health
	if err := gw.Mux().HandlePath("GET", "/health", toGatewayHandler(handlers.HealthCheck)); err != nil {
		return fmt.Errorf("failed to register health endpoint: %w", err)
	}

	// Serve static web files from embedded filesystem
	var baseHandler http.Handler = gw

	// Create filesystem from embedded web directory
	webFileSystem, err := fs.Sub(WebFS, "web")
	if err == nil {
		fileServer := http.FileServer(http.FS(webFileSystem))
		// Create a wrapper handler that checks for web files first, then falls back to gateway
		baseHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			path := r.URL.Path
			// Check if it's an API path or static file path
			if strings.HasPrefix(path, "/api") || strings.HasPrefix(path, "/health") {
				// API paths go to gateway
				gw.ServeHTTP(w, r)
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

	logger.Infof("HTTP server listening on %s", httpAddr)

	go func() {
		<-ctx.Done()
		logger.Info("Stopping HTTP server...")
		_ = srv.Shutdown(ctx)
	}()

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("HTTP server failed: %w", err)
	}

	return nil
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
		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

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
