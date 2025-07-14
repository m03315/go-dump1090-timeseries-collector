package parser

import (
	"fmt"
	"github.com/m03315/go-dump1090-timeseries-collector/models"
	"log"
	"strconv"
	"strings"
	"time"
)

// ParseSBS1Message decodes a raw dump1090 message string into an AircraftData struct.
// Returns (nil, nil) for non-"MSG" message types.
func ParseSBS1Message(line string) (*models.AircraftData, error) { // Returns models.AircraftData
	fields := strings.Split(strings.TrimSpace(line), ",")

	if len(fields) < 10 {
		return nil, fmt.Errorf("message too short, expected at least 10 fields for basic data: '%s'", line)
	}

	msgType := strings.TrimSpace(fields[0])
	if msgType != "MSG" {
		return nil, nil // Not an aircraft data message we're interested in, skip silently
	}

	var dateMsgGen, timeMsgGen, dateMsgLog, timeMsgLog string
	if len(fields) > 6 {
		dateMsgGen = strings.TrimSpace(fields[6])
	}
	if len(fields) > 7 {
		timeMsgGen = strings.TrimSpace(fields[7])
	}
	if len(fields) > 8 {
		dateMsgLog = strings.TrimSpace(fields[8])
	}
	if len(fields) > 9 {
		timeMsgLog = strings.TrimSpace(fields[9])
	}

	var generatedTimestamp time.Time
	if dateMsgGen != "" && timeMsgGen != "" {
		parsedTime, err := time.Parse("2006/01/02 15:04:05", fmt.Sprintf("%s %s", dateMsgGen, timeMsgGen))
		if err != nil {
			log.Printf("Warning: Could not parse Generated Timestamp (Fields 7&8) '%s %s': %v. Using current time.", dateMsgGen, timeMsgGen, err)
			generatedTimestamp = time.Now()
		} else {
			generatedTimestamp = parsedTime
		}
	} else {
		generatedTimestamp = time.Now() // Fallback
	}

	var loggedTimestamp time.Time
	if dateMsgLog != "" && timeMsgLog != "" {
		parsedTime, err := time.Parse("2006/01/02 15:04:05", fmt.Sprintf("%s %s", dateMsgLog, timeMsgLog))
		if err != nil {
			log.Printf("Warning: Could not parse Logged Timestamp (Fields 9&10) '%s %s': %v. Using current time.", dateMsgLog, timeMsgLog, err)
			loggedTimestamp = time.Now()
		} else {
			loggedTimestamp = parsedTime
		}
	} else {
		loggedTimestamp = time.Now() // Fallback
	}

	data := &models.AircraftData{ // Create a models.AircraftData struct
		MessageType:        msgType,
		HexIdent:           strings.TrimSpace(fields[4]),
		GeneratedTimestamp: generatedTimestamp,
		LoggedTimestamp:    loggedTimestamp,
	}

	if len(fields) > 1 && fields[1] != "" {
		data.TransmissionType = strings.TrimSpace(fields[1])
	}
	if len(fields) > 2 && fields[2] != "" {
		if val, err := strconv.Atoi(strings.TrimSpace(fields[2])); err == nil {
			data.SessionID = &val
		} else {
			log.Printf("Warning: Could not parse SessionID '%s': %v", fields[2], err)
		}
	}
	if len(fields) > 3 && fields[3] != "" {
		if val, err := strconv.Atoi(strings.TrimSpace(fields[3])); err == nil {
			data.AircraftID = &val
		} else {
			log.Printf("Warning: Could not parse AircraftID '%s': %v", fields[3], err)
		}
	}
	if len(fields) > 5 && fields[5] != "" {
		if val, err := strconv.Atoi(strings.TrimSpace(fields[5])); err == nil {
			data.FlightID = &val
		} else {
			log.Printf("Warning: Could not parse FlightID '%s': %v", fields[5], err)
		}
	}
	if len(fields) > 10 && fields[10] != "" {
		data.Callsign = strings.TrimSpace(fields[10])
	}
	if len(fields) > 11 && fields[11] != "" {
		if val, err := strconv.Atoi(strings.TrimSpace(fields[11])); err == nil {
			data.Altitude = &val
		} else {
			log.Printf("Warning: Could not parse Altitude '%s' for HexIdent %s: %v", fields[11], data.HexIdent, err)
		}
	}
	if len(fields) > 12 && fields[12] != "" {
		if val, err := strconv.ParseFloat(strings.TrimSpace(fields[12]), 64); err == nil {
			data.GroundSpeed = &val
		} else {
			log.Printf("Warning: Could not parse GroundSpeed '%s' for HexIdent %s: %v", fields[12], data.HexIdent, err)
		}
	}
	if len(fields) > 13 && fields[13] != "" {
		if val, err := strconv.ParseFloat(strings.TrimSpace(fields[13]), 64); err == nil {
			data.Track = &val
		} else {
			log.Printf("Warning: Could not parse Track '%s' for HexIdent %s: %v", fields[13], data.HexIdent, err)
		}
	}
	if len(fields) > 14 && fields[14] != "" {
		if val, err := strconv.ParseFloat(strings.TrimSpace(fields[14]), 64); err == nil {
			data.Latitude = &val
		} else {
			log.Printf("Warning: Could not parse Latitude '%s' for HexIdent %s: %v", fields[14], data.HexIdent, err)
		}
	}
	if len(fields) > 15 && fields[15] != "" {
		if val, err := strconv.ParseFloat(strings.TrimSpace(fields[15]), 64); err == nil {
			data.Longitude = &val
		} else {
			log.Printf("Warning: Could not parse Longitude '%s' for HexIdent %s: %v", fields[15], data.HexIdent, err)
		}
	}
	if len(fields) > 16 && fields[16] != "" {
		if val, err := strconv.Atoi(strings.TrimSpace(fields[16])); err == nil {
			data.VerticalRate = &val
		} else {
			log.Printf("Warning: Could not parse VerticalRate '%s' for HexIdent %s: %v", fields[16], data.HexIdent, err)
		}
	}
	if len(fields) > 17 && fields[17] != "" {
		data.Squawk = strings.TrimSpace(fields[17])
	}
	if len(fields) > 18 && fields[18] != "" {
		val := strings.TrimSpace(fields[18]) == "1"
		data.Alert = &val
	}
	if len(fields) > 19 && fields[19] != "" {
		val := strings.TrimSpace(fields[19]) == "1"
		data.Emergency = &val
	}
	if len(fields) > 20 && fields[20] != "" {
		val := strings.TrimSpace(fields[20]) == "1"
		data.SPI = &val
	}
	if len(fields) > 21 && fields[21] != "" {
		val := strings.TrimSpace(fields[21]) == "1"
		data.IsOnGround = &val
	}

	return data, nil
}
