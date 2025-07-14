package timeseries

import (
	"context"
	"github.com/m03315/go-dump1090-timeseries-collector/models"
)

// TimeSeriesWriter defines the interface for writing batches of aircraft data
// to any time-series database.
type TimeSeriesWriter interface {
	// WriteBatch writes a slice of AircraftData to the time-series database.
	// It takes a context for cancellation/timeouts.
	WriteBatch(ctx context.Context, batch []models.AircraftData) error

	// Close cleans up resources (e.g., closes database connections).
	Close() error
}
