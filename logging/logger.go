package logging

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"log/slog"
	"os"
)

// Define all types and variables at the package level for global access.
type ContextHandler struct {
	slog.Handler // interface
}

type AppLogger struct {
	Log *slog.Logger
}

type ContextKey string

const (
	TraceIDKey ContextKey = "TraceID"
)

var (
	logger  AppLogger
	handler ContextHandler
)

// Log returns the global application logger.
func Log() *slog.Logger {
	return logger.Log
}

// Handle adds context information (like TraceID) to the log record before passing it to the underlying handler.
func Handle(ctx context.Context, r slog.Record) error {
	if traceID, ok := ctx.Value(TraceIDKey).(string); ok {
		r.AddAttrs(slog.String(string(TraceIDKey), traceID))
	}
	return handler.Handle(ctx, r)
}

// Setup initializes the global logger with a JSON handler that writes to the provided io.Writer.
func Setup(w io.Writer, options slog.HandlerOptions) {
	baseHandler := slog.NewJSONHandler(w, &options)
	// add in the context handler
	customHandler := &ContextHandler{Handler: baseHandler}
	logger.Log = slog.New(customHandler)
}

// Default initializes the global logger with a Text handler that writes to os.Stdout.
func Default() {
	logger.Log = slog.New(slog.NewTextHandler(os.Stdout, nil))
}

// GenerateID returns a random 16-byte hex string (32 hex chars).
// We use crypto/rand for strong uniqueness properties.
func GenerateID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		// Unlikely path — return zeros if entropy source fails.
		return hex.EncodeToString(b[:])
	}
	return hex.EncodeToString(b[:])
}

// CreateAppDataFolder creates an application data folder in the user's cache directory.
func CreateAppDataFolder(applicationName string) (string, error) {
	dir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	dir = dir + "\\" + applicationName
	err = os.MkdirAll(dir, 0600)
	if err != nil {
		return "", err
	}
	return dir, nil
}

// OpenLogFile opens (or creates) a log file for appending log entries.
func OpenLogFile(fileName string) (*os.File, error) {
	// open the log file for appending log entries
	// create it if it does not exist with permissions rw-r--r--
	// append mode so we do not overwrite existing logs
	fi, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		// log file not ready so default std.err logging here
		slog.Error(fmt.Sprintf("%s\n", "Failed to create logfile for writing"))
		slog.Error(err.Error())
		return &os.File{}, err
	}
	return fi, nil
}

func LoggerOptions() slog.HandlerOptions {
	// TODO: adjust options based on environment
	var options slog.HandlerOptions
	options = slog.HandlerOptions{AddSource: false}
	return options
}
