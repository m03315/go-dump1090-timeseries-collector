package timeseries

import (
	"context"
	"fmt"
	"log"
	"net/http" // Ensure this is imported for http.Client
	"time"

	"github.com/InfluxCommunity/influxdb3-go/v2/influxdb3"
	"github.com/m03315/go-dump1090-timeseries-collector/models"
)

// InfluxDBWriter implements TimeSeriesWriter for InfluxDB 3.x.
type InfluxDBWriter struct {
	client   *influxdb3.Client
	database string
}

// NewInfluxDBWriter creates and returns a new InfluxDBWriter.
// It takes necessary connection details and an optional http.Client.
func NewInfluxDBWriter(host, token, database string, httpClient *http.Client) (*InfluxDBWriter, error) {
	client, err := influxdb3.New(influxdb3.ClientConfig{
		Host:       host,
		Token:      token,
		HTTPClient: httpClient, // Pass the standard net/http.Client directly
		Database:   database,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create InfluxDB 3.x client: %w", err)
	}
	return &InfluxDBWriter{
		client:   client,
		database: database,
	}, nil
}

// WriteBatch implements the TimeSeriesWriter interface for InfluxDB.
func (iw *InfluxDBWriter) WriteBatch(ctx context.Context, batch []models.AircraftData) error {
	if len(batch) == 0 {
		return nil
	}

	pointsToWrite := make([]*influxdb3.Point, 0, len(batch))

	for _, data := range batch {
		point := influxdb3.NewPointWithMeasurement("aircraft_sbs1").
			SetTimestamp(data.GeneratedTimestamp)

		// Set Tags
		point.SetTag("message_type", data.MessageType)
		if data.TransmissionType != "" {
			point.SetTag("transmission_type", data.TransmissionType)
		}
		point.SetTag("hex_ident", data.HexIdent)
		if data.Callsign != "" {
			point.SetTag("callsign", data.Callsign)
		}
		if data.Squawk != "" {
			point.SetTag("squawk", data.Squawk)
		}

		// Set Fields
		if data.SessionID != nil {
			point.SetField("session_id", *data.SessionID)
		}
		if data.AircraftID != nil {
			point.SetField("aircraft_id", *data.AircraftID)
		}
		if data.FlightID != nil {
			point.SetField("flight_id", *data.FlightID)
		}
		point.SetField("logged_timestamp_unix_ms", data.LoggedTimestamp.UnixNano()/int64(time.Millisecond))
		if data.Altitude != nil {
			point.SetField("altitude_ft", *data.Altitude)
		}
		if data.GroundSpeed != nil {
			point.SetField("ground_speed_kts", *data.GroundSpeed)
		}
		if data.Track != nil {
			point.SetField("track_deg", *data.Track)
		}
		if data.Latitude != nil {
			point.SetField("latitude", *data.Latitude)
		}
		if data.Longitude != nil {
			point.SetField("longitude", *data.Longitude)
		}
		if data.VerticalRate != nil {
			point.SetField("vertical_rate_fpm", *data.VerticalRate)
		}
		if data.Alert != nil {
			point.SetField("alert", *data.Alert)
		}
		if data.Emergency != nil {
			point.SetField("emergency", *data.Emergency)
		}
		if data.SPI != nil {
			point.SetField("spi", *data.SPI)
		}
		if data.IsOnGround != nil {
			point.SetField("is_on_ground", *data.IsOnGround)
		}

		if point.HasFields() {
			pointsToWrite = append(pointsToWrite, point)
		} else {
			log.Printf("Warning: Point for HexIdent %s has no fields and will be skipped. Raw message likely lacked relevant data or was not a position/status message.", data.HexIdent)
		}
	}

	log.Printf("Writing batch of %d points to InfluxDB 3.x (database: %s)...", len(pointsToWrite), iw.database)
	err := iw.client.WritePoints(ctx, pointsToWrite)
	if err != nil {
		return fmt.Errorf("influxdb write error: %w", err)
	}
	return nil
}

// Close implements the TimeSeriesWriter interface.
func (iw *InfluxDBWriter) Close() error {
	if iw.client != nil {
		return iw.client.Close()
	}
	return nil
}
