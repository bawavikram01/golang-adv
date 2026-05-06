// =============================================================================
// LESSON 10: PRODUCTION-GRADE GO — Building Systems That Don't Break at 3AM
// =============================================================================
//
// This lesson covers the patterns that keep Go services running in production:
//   - Graceful shutdown with signal handling
//   - Structured logging
//   - Health checks
//   - Configuration management
//   - Error wrapping and sentinel errors
//   - Middleware chains
//   - Dependency injection without frameworks
// =============================================================================

package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// =============================================================================
// PART 1: Structured Error Handling
// =============================================================================

// Sentinel errors — package-level, immutable error values
var (
	ErrNotFound     = errors.New("not found")
	ErrUnauthorized = errors.New("unauthorized")
	ErrConflict     = errors.New("conflict")
	ErrInternal     = errors.New("internal error")
)

// Domain error with context
type AppError struct {
	Op      string // operation that failed
	Kind    error  // category (sentinel error)
	Err     error  // underlying error
	Message string // human-readable message
}

func (e *AppError) Error() string {
	return fmt.Sprintf("%s: %s: %v", e.Op, e.Message, e.Err)
}

func (e *AppError) Unwrap() error {
	return e.Kind // allows errors.Is(err, ErrNotFound) to work
}

// Helper constructors
func NotFoundError(op, msg string, err error) *AppError {
	return &AppError{Op: op, Kind: ErrNotFound, Err: err, Message: msg}
}

// Error handling in practice:
func getUserByID(id int64) (*UserDTO, error) {
	if id <= 0 {
		return nil, NotFoundError("getUserByID", fmt.Sprintf("user %d", id), nil)
	}
	return &UserDTO{ID: id, Name: "Vikram"}, nil
}

type UserDTO struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// Caller checks error category, not string matching:
func handleGetUser(id int64) {
	user, err := getUserByID(id)
	if err != nil {
		// Use errors.Is for sentinel errors (works through wrapping)
		if errors.Is(err, ErrNotFound) {
			fmt.Printf("  404: %v\n", err)
			return
		}
		fmt.Printf("  500: %v\n", err)
		return
	}
	fmt.Printf("  200: %+v\n", user)
}

// =============================================================================
// PART 2: Structured Logging with slog (Go 1.21+)
// =============================================================================

func setupLogger() *slog.Logger {
	// JSON handler for production (machine-parseable)
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
		// Add source file info
		AddSource: false,
	})

	logger := slog.New(handler)
	slog.SetDefault(logger) // set as default logger
	return logger
}

// Middleware that adds request context to logger
type contextKeyType string

const loggerKey contextKeyType = "logger"

func LoggerFromContext(ctx context.Context) *slog.Logger {
	if l, ok := ctx.Value(loggerKey).(*slog.Logger); ok {
		return l
	}
	return slog.Default()
}

func WithLogger(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}

// =============================================================================
// PART 3: Graceful Shutdown
// =============================================================================

type Application struct {
	httpServer *http.Server
	logger     *slog.Logger
	shutdownFns []func(context.Context) error
	mu         sync.Mutex
}

func NewApplication(logger *slog.Logger) *Application {
	return &Application{logger: logger}
}

func (app *Application) OnShutdown(fn func(context.Context) error) {
	app.mu.Lock()
	defer app.mu.Unlock()
	app.shutdownFns = append(app.shutdownFns, fn)
}

func (app *Application) Start(addr string) error {
	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status": "healthy",
			"time":   time.Now().UTC().Format(time.RFC3339),
		})
	})

	// Readiness check (for k8s)
	mux.HandleFunc("GET /ready", func(w http.ResponseWriter, r *http.Request) {
		// Check dependencies (DB, cache, etc.)
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "ready")
	})

	// Example API endpoint
	mux.HandleFunc("GET /api/users/{id}", app.handleGetUser)

	// Apply middleware
	handler := app.recoveryMiddleware(
		app.loggingMiddleware(
			app.requestIDMiddleware(mux),
		),
	)

	app.httpServer = &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Register HTTP server shutdown
	app.OnShutdown(func(ctx context.Context) error {
		return app.httpServer.Shutdown(ctx)
	})

	app.logger.Info("server starting", "addr", addr)
	return app.httpServer.ListenAndServe()
}

func (app *Application) Shutdown(ctx context.Context) error {
	app.logger.Info("shutting down gracefully...")

	app.mu.Lock()
	fns := make([]func(context.Context) error, len(app.shutdownFns))
	copy(fns, app.shutdownFns)
	app.mu.Unlock()

	var errs []error
	for _, fn := range fns {
		if err := fn(ctx); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

// =============================================================================
// PART 4: HTTP Middleware Stack
// =============================================================================

func (app *Application) requestIDMiddleware(next http.Handler) http.Handler {
	var counter uint64
	var mu sync.Mutex
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		counter++
		id := fmt.Sprintf("req-%d", counter)
		mu.Unlock()

		// Add request ID to context logger
		logger := app.logger.With("request_id", id)
		ctx := WithLogger(r.Context(), logger)

		w.Header().Set("X-Request-ID", id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

func (app *Application) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rw, r)

		logger := LoggerFromContext(r.Context())
		logger.Info("request completed",
			"method", r.Method,
			"path", r.URL.Path,
			"status", rw.status,
			"duration_ms", time.Since(start).Milliseconds(),
		)
	})
}

func (app *Application) recoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				logger := LoggerFromContext(r.Context())
				logger.Error("panic recovered",
					"panic", rec,
					"path", r.URL.Path,
				)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func (app *Application) handleGetUser(w http.ResponseWriter, r *http.Request) {
	logger := LoggerFromContext(r.Context())
	id := r.PathValue("id")
	logger.Info("fetching user", "user_id", id)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"id":   id,
		"name": "Vikram",
	})
}

// =============================================================================
// PART 5: Dependency Injection Without Frameworks
// =============================================================================
// Go favors explicit dependency injection via constructors.

type UserService struct {
	repo   UserRepository
	cache  CacheService
	logger *slog.Logger
}

type UserRepository interface {
	FindByID(ctx context.Context, id int64) (*UserDTO, error)
}

type CacheService interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
}

func NewUserService(repo UserRepository, cache CacheService, logger *slog.Logger) *UserService {
	return &UserService{repo: repo, cache: cache, logger: logger}
}

// =============================================================================
// PART 6: Configuration Pattern
// =============================================================================

type Config struct {
	Server   ServerCfg
	Database DatabaseCfg
}

type ServerCfg struct {
	Host            string        `json:"host"`
	Port            int           `json:"port"`
	ReadTimeout     time.Duration `json:"read_timeout"`
	WriteTimeout    time.Duration `json:"write_timeout"`
	ShutdownTimeout time.Duration `json:"shutdown_timeout"`
}

type DatabaseCfg struct {
	DSN             string        `json:"dsn"`
	MaxOpenConns    int           `json:"max_open_conns"`
	MaxIdleConns    int           `json:"max_idle_conns"`
	ConnMaxLifetime time.Duration `json:"conn_max_lifetime"`
}

func DefaultConfig() Config {
	return Config{
		Server: ServerCfg{
			Host:            "0.0.0.0",
			Port:            8080,
			ReadTimeout:     10 * time.Second,
			WriteTimeout:    10 * time.Second,
			ShutdownTimeout: 30 * time.Second,
		},
		Database: DatabaseCfg{
			MaxOpenConns:    25,
			MaxIdleConns:    5,
			ConnMaxLifetime: 5 * time.Minute,
		},
	}
}

func main() {
	// Error handling
	fmt.Println("=== Structured Error Handling ===")
	handleGetUser(1)
	handleGetUser(-1)

	// Structured logging
	fmt.Println("\n=== Structured Logging (slog) ===")
	logger := setupLogger()
	logger.Info("application starting",
		"version", "1.0.0",
		"env", "production",
	)
	logger.With("component", "db").Info("connected",
		"host", "localhost",
		"pool_size", 25,
	)

	// Graceful shutdown demo
	fmt.Println("\n=== Graceful Shutdown Pattern ===")
	fmt.Println("Production-ready server startup pattern:")
	fmt.Println(`
    app := NewApplication(logger)
    
    // Start server in goroutine
    go func() {
        if err := app.Start(":8080"); err != http.ErrServerClosed {
            logger.Error("server error", "error", err)
        }
    }()
    
    // Wait for interrupt signal
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit
    
    // Graceful shutdown with timeout
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    if err := app.Shutdown(ctx); err != nil {
        logger.Error("shutdown error", "error", err)
    }
`)

	// Show config
	cfg := DefaultConfig()
	cfgJSON, _ := json.MarshalIndent(cfg, "", "  ")
	fmt.Printf("\n=== Default Config ===\n%s\n", cfgJSON)

	fmt.Println("\n=== PRODUCTION CHECKLIST ===")
	fmt.Println("1. Structured errors with errors.Is/As (not string matching)")
	fmt.Println("2. Structured logging (slog) with request context")
	fmt.Println("3. Graceful shutdown with signal handling")
	fmt.Println("4. Health/readiness endpoints for k8s")
	fmt.Println("5. Panic recovery middleware")
	fmt.Println("6. Request timeouts (ReadTimeout, WriteTimeout)")
	fmt.Println("7. Explicit dependency injection via constructors")
	fmt.Println("8. Configuration with sensible defaults")
	fmt.Println("9. Middleware chain (logging, auth, recovery, rate limiting)")
	fmt.Println("10. Connection pool tuning (DB, Redis)")

	// Prevent unused import warnings
	_ = os.Stdout
	_ = signal.Notify
	_ = syscall.SIGTERM
	_ = context.Background
}
