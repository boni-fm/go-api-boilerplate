package database

import (
	"context"
	"testing"
	"time"

	logger "github.com/boni-fm/go-libsd3/pkg/log"
)

func TestDcAdapter_GetOrInit(t *testing.T) {
	// Initialize logger
	log := logger.NewLoggerWithFilename("test-adapter")

	adapter := GetDcAdapter("test-app", log)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	kodeDc := "G026SIM"

	// Attempt to GetOrInit the connection
	db, err := adapter.GetOrInit(ctx, kodeDc)
	if err != nil {
		t.Fatalf("Failed to GetOrInit for %s: %v", kodeDc, err)
	}

	if db == nil {
		t.Fatalf("Expected db to not be nil")
	}

	t.Logf("Successfully connected to %s", kodeDc)

	// Verify stats
	stats := adapter.Stats(ctx)
	if _, exists := stats[kodeDc]; !exists {
		t.Errorf("Expected stats to contain key %s", kodeDc)
	}

	// Reset and clean up
	adapter.Reset(kodeDc)
	adapter.CloseAll()
}
