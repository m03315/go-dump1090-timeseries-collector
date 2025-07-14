package main

import (
	"bufio"
	"context"
	"fmt"
	"github.com/m03315/go-dump1090-timeseries-collector/config"
	"github.com/m03315/go-dump1090-timeseries-collector/internal/parser"
	"github.com/m03315/go-dump1090-timeseries-collector/internal/timeseries"
	"github.com/m03315/go-dump1090-timeseries-collector/models"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// Dump1090Collector manages the connection to dump1090 and data ingestion.
type Dump1090Collector struct {
	config    *config.Config // !!! Changed type to use config.Config !!!
	writer    timeseries.TimeSeriesWriter
	conn      net.Conn
	running   bool
	dataChan  chan string
	batchChan chan []models.AircraftData
	errorChan chan error
	doneChan  chan struct{}
	flushChan chan struct{}
}

// NewDump1090Collector now accepts a *config.Config.
func NewDump1090Collector(cfg *config.Config, writer timeseries.TimeSeriesWriter) *Dump1090Collector {
	return &Dump1090Collector{
		config:    cfg,
		writer:    writer,
		dataChan:  make(chan string, 1000),
		batchChan: make(chan []models.AircraftData, 10),
		errorChan: make(chan error, 100),
		doneChan:  make(chan struct{}),
		flushChan: make(chan struct{}),
	}
}

// connectToDump1090 remains unchanged
func (d *Dump1090Collector) connectToDump1090() error {
	address := fmt.Sprintf("%s:%s", d.config.Dump1090Host, d.config.Dump1090Port)
	log.Printf("Attempting to connect to dump1090 at %s...", address)

	retries := 0
	for {
		conn, err := net.Dial("tcp", address)
		if err != nil {
			log.Printf("Error connecting to dump1090: %v. Retrying in %s...", err, d.config.ConnectRetryDelay)
			retries++
			if d.config.ConnectMaxRetries > 0 && retries > d.config.ConnectMaxRetries {
				return fmt.Errorf("max connection retries (%d) exceeded to dump1090 at %s", d.config.ConnectMaxRetries, address)
			}
			select {
			case <-d.doneChan:
				return fmt.Errorf("shutdown initiated during dump1090 connection attempt")
			case <-time.After(d.config.ConnectRetryDelay):
				continue
			}
		}
		d.conn = conn
		log.Printf("Successfully connected to dump1090 at %s.", address)
		return nil
	}
}

// readData remains unchanged
func (d *Dump1090Collector) readData() {
	defer func() {
		if d.conn != nil {
			log.Println("Closing dump1090 connection...")
			err := d.conn.Close()
			if err != nil {
				return
			}
		}
		close(d.dataChan)
		log.Println("readData goroutine stopped.")
	}()

	for d.running {
		if d.conn == nil {
			if err := d.connectToDump1090(); err != nil {
				d.errorChan <- err
				select {
				case <-d.doneChan:
					return
				case <-time.After(d.config.ConnectRetryDelay):
					continue
				}
			}
		}

		scanner := bufio.NewScanner(d.conn)
		scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

		for scanner.Scan() {
			select {
			case <-d.doneChan:
				log.Println("readData goroutine stopping due to shutdown.")
				return
			case d.dataChan <- scanner.Text():
			default:
				log.Println("Warning: Raw data channel full or slow consumer, dropping message to keep up with stream.")
			}
		}

		if err := scanner.Err(); err != nil {
			log.Printf("Error reading from dump1090 scanner: %v. Attempting to reconnect...", err)
			if d.conn != nil {
				err := d.conn.Close()
				if err != nil {
					return
				}
			}
			d.conn = nil
		} else {
			log.Println("Dump1090 connection appears to be closed by remote. Attempting to reconnect...")
			if d.conn != nil {
				err := d.conn.Close()
				if err != nil {
					return
				}
			}
			d.conn = nil
		}
	}
}

// parseAndBatchData now uses config.BatchSize and config.BatchInterval
func (d *Dump1090Collector) parseAndBatchData() {
	defer func() {
		close(d.batchChan)
		log.Println("parseAndBatchData goroutine stopped.")
	}()

	batch := make([]models.AircraftData, 0, d.config.BatchSize) // !!! Use config.BatchSize !!!
	ticker := time.NewTicker(d.config.BatchInterval)            // !!! Use config.BatchInterval !!!
	defer ticker.Stop()

	for {
		select {
		case <-d.doneChan:
			if len(batch) > 0 {
				log.Println("Parse and Batch: Flushing remaining data on shutdown.")
				d.batchChan <- batch
			}
			return
		case rawLine, ok := <-d.dataChan:
			if !ok {
				if len(batch) > 0 {
					log.Println("Parse and Batch: Flushing final data after raw data channel closed.")
					d.batchChan <- batch
				}
				return
			}

			data, err := parser.ParseSBS1Message(rawLine)
			if err != nil {
				log.Printf("Parse error for message '%s': %v", rawLine, err)
				continue
			}
			if data != nil {
				batch = append(batch, *data)
				if len(batch) >= d.config.BatchSize { // !!! Use config.BatchSize !!!
					d.batchChan <- batch
					batch = make([]models.AircraftData, 0, d.config.BatchSize)
					ticker.Reset(d.config.BatchInterval)
				}
			}
		case <-ticker.C:
			if len(batch) > 0 {
				d.batchChan <- batch
				batch = make([]models.AircraftData, 0, d.config.BatchSize)
			}
		case <-d.flushChan:
			if len(batch) > 0 {
				d.batchChan <- batch
				batch = make([]models.AircraftData, 0, d.config.BatchSize)
			}
		}
	}
}

// batchWriter remains unchanged (it uses the interface)
func (d *Dump1090Collector) batchWriter() {
	defer func() {
		if d.writer != nil {
			err := d.writer.Close()
			if err != nil {
				log.Printf("Error closing time-series writer: %v", err)
			} else {
				log.Println("Time-series writer closed.")
			}
		}
		log.Println("batchWriter goroutine stopped.")
	}()

	for {
		select {
		case <-d.doneChan:
			return
		case batch, ok := <-d.batchChan:
			if !ok {
				return
			}
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			err := d.writer.WriteBatch(ctx, batch)
			cancel()
			if err != nil {
				d.errorChan <- fmt.Errorf("time-series write error: %w", err)
			}
		}
	}
}

// errorHandler remains unchanged
func (d *Dump1090Collector) errorHandler() {
	defer log.Println("errorHandler goroutine stopped.")
	for {
		select {
		case <-d.doneChan:
			return
		case err := <-d.errorChan:
			log.Printf("ERROR: %v", err)
		}
	}
}

// Start method orchestration
func (d *Dump1090Collector) Start() error {
	log.Println("Starting dump1090 data collector...")

	d.running = true

	go d.errorHandler()
	go d.readData()
	go d.parseAndBatchData()
	go d.batchWriter()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	log.Println("Data collection started. Press Ctrl+C to stop.")
	<-sigChan

	log.Println("Received shutdown signal. Initiating graceful shutdown...")
	d.running = false
	close(d.doneChan)
	close(d.flushChan)

	time.Sleep(d.config.BatchInterval + 2*time.Second) // !!! Use config.BatchInterval !!!

	log.Println("Shutdown complete.")
	return nil
}

func main() {
	// !!! Load configuration using the config package !!!
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	var tsWriter timeseries.TimeSeriesWriter

	// Choose writer based on the loaded config
	switch cfg.OutputDBType {
	case "influxdb":
		httpClient := &http.Client{
			Timeout: 30 * time.Second,
		}
		tsWriter, err = timeseries.NewInfluxDBWriter(
			cfg.InfluxHost,
			cfg.InfluxToken,
			cfg.InfluxDatabase,
			httpClient,
		)
		if err != nil {
			log.Fatalf("Failed to initialize InfluxDB writer: %v", err)
		}
		log.Println("Initialized InfluxDB writer.")
	case "prometheus":
		// Example placeholder for Prometheus
		// tsWriter, err = timeseries.NewPrometheusWriter(cfg.PrometheusPushGatewayURL)
		// if err != nil { ... }
		// log.Println("Initialized Prometheus writer.")
		log.Fatalf("Prometheus output type not yet implemented.")
	default:
		log.Fatalf("Unsupported OUTPUT_DB_TYPE: %s", cfg.OutputDBType)
	}

	collector := NewDump1090Collector(cfg, tsWriter)
	if err := collector.Start(); err != nil {
		log.Fatalf("Collector exited with error: %v", err)
	}
}
