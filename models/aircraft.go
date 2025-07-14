package models

import "time"

// AircraftData represents the parsed information from an SBS-1 message.
type AircraftData struct {
	MessageType        string
	TransmissionType   string
	SessionID          *int
	AircraftID         *int
	HexIdent           string
	FlightID           *int
	GeneratedTimestamp time.Time
	LoggedTimestamp    time.Time

	Callsign     string
	Altitude     *int
	GroundSpeed  *float64
	Track        *float64
	Latitude     *float64
	Longitude    *float64
	VerticalRate *int
	Squawk       string
	Alert        *bool
	Emergency    *bool
	SPI          *bool
	IsOnGround   *bool
}
