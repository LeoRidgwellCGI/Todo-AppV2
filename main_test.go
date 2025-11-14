package main

import (
	"context"
	"log/slog"
	"os"
	"strings"
	"testing"
	"time"
)

// TestMain_ContextHandler tests the ContextHandler Handle method.
func TestMain_ContextHandler(t *testing.T) {
	// Create a temp log file
	tmpFile, err := os.CreateTemp("", "test_log_*.log")
	if err != nil {
		t.Fatalf("Failed to create temp log file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// Create handler with context
	opts := slog.HandlerOptions{AddSource: false}
	baseHandler := slog.NewTextHandler(tmpFile, &opts)
	handler := &ContextHandler{Handler: baseHandler}

	// Create context with trace ID
	ctx := context.WithValue(context.Background(), traceIDKey, "test-trace-123")

	// Create a log record
	record := slog.Record{
		Time:    time.Now(),
		Message: "Test message",
		Level:   slog.LevelInfo,
	}

	// Handle the record
	err = handler.Handle(ctx, record)
	if err != nil {
		t.Errorf("Handle failed: %v", err)
	}

	// Verify trace ID was added
	tmpFile.Seek(0, 0)
	content, _ := os.ReadFile(tmpFile.Name())
	logContent := string(content)

	if !strings.Contains(logContent, "test-trace-123") {
		t.Errorf("Expected trace ID in log, got: %s", logContent)
	}
	if !strings.Contains(logContent, "Test message") {
		t.Errorf("Expected log message in log, got: %s", logContent)
	}
}

// TestMain_ContextHandler_NoTraceID tests ContextHandler without trace ID.
func TestMain_ContextHandler_NoTraceID(t *testing.T) {
	// Create a temp log file
	tmpFile, err := os.CreateTemp("", "test_log_*.log")
	if err != nil {
		t.Fatalf("Failed to create temp log file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// Create handler without context
	opts := slog.HandlerOptions{AddSource: false}
	baseHandler := slog.NewTextHandler(tmpFile, &opts)
	handler := &ContextHandler{Handler: baseHandler}

	// Create context without trace ID
	ctx := context.Background()

	// Create a log record
	record := slog.Record{
		Time:    time.Now(),
		Message: "Test message without trace",
		Level:   slog.LevelInfo,
	}

	// Handle the record
	err = handler.Handle(ctx, record)
	if err != nil {
		t.Errorf("Handle failed: %v", err)
	}

	// Verify message is logged even without trace ID
	tmpFile.Seek(0, 0)
	content, _ := os.ReadFile(tmpFile.Name())
	logContent := string(content)

	if !strings.Contains(logContent, "Test message without trace") {
		t.Errorf("Expected log message in log, got: %s", logContent)
	}
}

// TestMain_ContextHandler_WithDifferentLogLevels tests logging at different levels.
func TestMain_ContextHandler_WithDifferentLogLevels(t *testing.T) {
	levels := []slog.Level{
		slog.LevelDebug,
		slog.LevelInfo,
		slog.LevelWarn,
		slog.LevelError,
	}

	for _, level := range levels {
		t.Run(level.String(), func(t *testing.T) {
			tmpFile, err := os.CreateTemp("", "test_log_*.log")
			if err != nil {
				t.Fatalf("Failed to create temp log file: %v", err)
			}
			defer os.Remove(tmpFile.Name())
			defer tmpFile.Close()

			opts := slog.HandlerOptions{AddSource: false, Level: slog.LevelDebug}
			baseHandler := slog.NewTextHandler(tmpFile, &opts)
			handler := &ContextHandler{Handler: baseHandler}

			ctx := context.WithValue(context.Background(), traceIDKey, "test-trace")

			record := slog.Record{
				Time:    time.Now(),
				Message: "Test at level " + level.String(),
				Level:   level,
			}

			err = handler.Handle(ctx, record)
			if err != nil {
				t.Errorf("Handle failed for level %s: %v", level, err)
			}

			tmpFile.Seek(0, 0)
			content, _ := os.ReadFile(tmpFile.Name())
			logContent := string(content)

			if !strings.Contains(logContent, "Test at level") {
				t.Errorf("Expected log message at level %s, got: %s", level, logContent)
			}
		})
	}
}

// TestMain_ContextHandler_MultipleAttributes tests logging with multiple attributes.
func TestMain_ContextHandler_MultipleAttributes(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_log_*.log")
	if err != nil {
		t.Fatalf("Failed to create temp log file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	opts := slog.HandlerOptions{AddSource: false}
	baseHandler := slog.NewTextHandler(tmpFile, &opts)
	handler := &ContextHandler{Handler: baseHandler}

	ctx := context.WithValue(context.Background(), traceIDKey, "multi-trace")

	record := slog.Record{
		Time:    time.Now(),
		Message: "Test with attributes",
		Level:   slog.LevelInfo,
	}
	record.AddAttrs(
		slog.String("key1", "value1"),
		slog.Int("key2", 42),
	)

	err = handler.Handle(ctx, record)
	if err != nil {
		t.Errorf("Handle failed: %v", err)
	}

	tmpFile.Seek(0, 0)
	content, _ := os.ReadFile(tmpFile.Name())
	logContent := string(content)

	if !strings.Contains(logContent, "multi-trace") {
		t.Errorf("Expected trace ID in log, got: %s", logContent)
	}
	if !strings.Contains(logContent, "Test with attributes") {
		t.Errorf("Expected message in log, got: %s", logContent)
	}
}

// TestMain_ContextHandler_EmptyTraceID tests with empty trace ID.
func TestMain_ContextHandler_EmptyTraceID(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_log_*.log")
	if err != nil {
		t.Fatalf("Failed to create temp log file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	opts := slog.HandlerOptions{AddSource: false}
	baseHandler := slog.NewTextHandler(tmpFile, &opts)
	handler := &ContextHandler{Handler: baseHandler}

	ctx := context.WithValue(context.Background(), traceIDKey, "")

	record := slog.Record{
		Time:    time.Now(),
		Message: "Test with empty trace",
		Level:   slog.LevelInfo,
	}

	err = handler.Handle(ctx, record)
	if err != nil {
		t.Errorf("Handle failed: %v", err)
	}

	tmpFile.Seek(0, 0)
	content, _ := os.ReadFile(tmpFile.Name())
	logContent := string(content)

	if !strings.Contains(logContent, "Test with empty trace") {
		t.Errorf("Expected message in log, got: %s", logContent)
	}
}

// TestMain_ContextHandler_ConcurrentWrites tests concurrent log writes.
func TestMain_ContextHandler_ConcurrentWrites(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_log_*.log")
	if err != nil {
		t.Fatalf("Failed to create temp log file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	opts := slog.HandlerOptions{AddSource: false}
	baseHandler := slog.NewTextHandler(tmpFile, &opts)
	handler := &ContextHandler{Handler: baseHandler}

	const numGoroutines = 10
	done := make(chan bool, numGoroutines)
	errChan := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(index int) {
			ctx := context.WithValue(context.Background(), traceIDKey, "trace-"+string(rune('0'+index)))
			record := slog.Record{
				Time:    time.Now(),
				Message: "Concurrent log message",
				Level:   slog.LevelInfo,
			}
			if err := handler.Handle(ctx, record); err != nil {
				errChan <- err
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < numGoroutines; i++ {
		<-done
	}
	close(errChan)

	for err := range errChan {
		t.Errorf("Concurrent write failed: %v", err)
	}
}

// TestMain_ContextHandler_WithSourceOption tests handler with AddSource option.
func TestMain_ContextHandler_WithSourceOption(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_log_*.log")
	if err != nil {
		t.Fatalf("Failed to create temp log file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	opts := slog.HandlerOptions{AddSource: true}
	baseHandler := slog.NewTextHandler(tmpFile, &opts)
	handler := &ContextHandler{Handler: baseHandler}

	ctx := context.WithValue(context.Background(), traceIDKey, "source-trace")

	record := slog.Record{
		Time:    time.Now(),
		Message: "Test with source",
		Level:   slog.LevelInfo,
	}

	err = handler.Handle(ctx, record)
	if err != nil {
		t.Errorf("Handle failed: %v", err)
	}
}

// TestMain_ContextKey tests the context key type.
func TestMain_ContextKey(t *testing.T) {
	key1 := ctxKey("key1")
	key2 := ctxKey("key2")

	if key1 == key2 {
		t.Error("Different context keys should not be equal")
	}

	ctx := context.Background()
	ctx = context.WithValue(ctx, key1, "value1")
	ctx = context.WithValue(ctx, key2, "value2")

	if ctx.Value(key1) != "value1" {
		t.Error("Expected to retrieve value1 for key1")
	}
	if ctx.Value(key2) != "value2" {
		t.Error("Expected to retrieve value2 for key2")
	}
}

// TestMain_Constants tests that constants are defined correctly.
func TestMain_Constants(t *testing.T) {
	if datafolder != "tododata" {
		t.Errorf("Expected datafolder to be 'tododata', got '%s'", datafolder)
	}
	if datafile != "todos.json" {
		t.Errorf("Expected datafile to be 'todos.json', got '%s'", datafile)
	}
	if logfile != "todos.log" {
		t.Errorf("Expected logfile to be 'todos.log', got '%s'", logfile)
	}
}

// TestMain_TraceIDKey tests the trace ID context key.
func TestMain_TraceIDKey(t *testing.T) {
	ctx := context.WithValue(context.Background(), traceIDKey, "test-id")

	value := ctx.Value(traceIDKey)
	if value == nil {
		t.Error("Expected trace ID to be set in context")
	}

	if strValue, ok := value.(string); !ok {
		t.Error("Expected trace ID to be a string")
	} else if strValue != "test-id" {
		t.Errorf("Expected trace ID 'test-id', got '%s'", strValue)
	}
}

// TestMain_RunModeConstants tests that RunMode constants are defined correctly.
func TestMain_RunModeConstants(t *testing.T) {
	if RunModeCLI != "CLI" {
		t.Errorf("Expected RunModeCLI to be 'CLI', got '%s'", RunModeCLI)
	}
	if RunModeServer != "SERVER" {
		t.Errorf("Expected RunModeServer to be 'SERVER', got '%s'", RunModeServer)
	}
}

// TestMain_RunModeVariable tests the runMode variable behavior.
func TestMain_RunModeVariable(t *testing.T) {
	// Store original value
	originalMode := runMode
	defer func() { runMode = originalMode }()

	// Test CLI mode
	runMode = RunModeCLI
	if runMode != "CLI" {
		t.Errorf("Expected runMode to be 'CLI', got '%s'", runMode)
	}

	// Test SERVER mode
	runMode = RunModeServer
	if runMode != "SERVER" {
		t.Errorf("Expected runMode to be 'SERVER', got '%s'", runMode)
	}
}
