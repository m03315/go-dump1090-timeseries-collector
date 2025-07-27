# go-dump1090-timeseries-collector

A flexible and robust Go client for collecting and ingesting real-time aviation data from a [dump1090](https://github.com/flightaware/dump1090) server and writing it to various time-series databases.

## Project Vision

This project's goal is to be a versatile data collector that can act as the "glue" between a `dump1090` data stream and a diverse set of time-series analysis platforms. It is designed with a modular and extensible architecture to easily support new data formats, protocols, and database backends.

## Features

* **Modular Architecture:** The codebase is structured with clear separation of concerns, with dedicated packages for data models, parsing logic, and time-series database writers.
* **Flexible Data Sinks:** Implements a `TimeSeriesWriter` interface, allowing the program to easily switch between different time-series databases with minimal code changes.
    * **Current Implementation:** InfluxDB 3.x is fully supported.
    * **Planned Implementations:** Prometheus, TimescaleDB, and others.
* **Protocol Support:** Currently parses data using the **SBS-1 protocol**, specifically from dump1090's port `30003`.
* **Robust & Resilient:** Includes built-in reconnection and retry logic to maintain a stable connection to the dump1090 server.
* **Efficient Data Handling:** Utilizes Go channels and batch processing to efficiently parse and write large volumes of data without overwhelming the destination database.

## Getting Started

### Prerequisites

* Go (version 1.22 or higher)
* A running `dump1090` server (e.g., `dump1090-fa` or `dump1090-mutability`)
* An instance of a supported time-series database (e.g., InfluxDB 3.x)

### Installation

1.  Clone the repository:
    ```bash
    git clone [https://github.com/m03315/go-dump1090-timeseries-collector.git](https://github.com/m03315/go-dump1090-timeseries-collector.git)
    cd go-dump1090-timeseries-collector
    ```
2.  Initialize the Go module and download dependencies:
    ```bash
    go mod tidy
    ```

## Configuration

The application is configured using environment variables. This allows for flexible deployment in containers, on servers, or in a local development environment.

| Variable | Description | Default | Required for InfluxDB |
| :--- | :--- | :--- | :--- |
| `DUMP1090_HOST` | Hostname or IP of the dump1090 server. | `localhost` | No |
| `DUMP1090_PORT` | Port for the SBS-1 protocol. | `30003` | No |
| `OUTPUT_DB_TYPE` | The type of time-series database to write to. | `influxdb` | No |
| `INFLUX_URL` | The URL of your InfluxDB 3.x instance. | (none) | Yes |
| `INFLUXDB_TOKEN` | The authentication token for InfluxDB. | (none) | Yes |
| `INFLUXDB_DATABASE` | The target database name in InfluxDB. | (none) | Yes |
| `BATCH_SIZE` | The number of messages to batch before writing to the database. | `50` | No |
| `BATCH_INTERVAL` | The maximum time to wait before flushing a batch, even if it's not full (e.g., `5s`). | `5s` | No |
| `CONNECT_RETRY_DELAY` | Time to wait between connection attempts to dump1090 (e.g., `5s`). | `5s` | No |
| `CONNECT_MAX_RETRIES` | Max number of connection attempts to dump1090 (`0` for infinite). | `0` | No |

## Usage

Set the required environment variables and run the application.

**Example with InfluxDB:**

```bash
export DUMP1090_HOST="localhost"
export DUMP1090_PORT="30003"
export OUTPUT_DB_TYPE="influxdb"
export INFLUX_URL="http://your-influxdb-host:8186"
export INFLUXDB_TOKEN="your-super-secret-token"
export INFLUXDB_DATABASE="adsb_data"

go run main.go
```

The program will connect to your dump1090 server, start collecting and parsing messages, and write them in batches to your specified InfluxDB instance.

## Roadmap
The project is actively being developed. Future plans include:

* Additional Protocols: Adding parsers for other dump1090 data formats, such as the JSON data from port 30106 or the raw data from 30001.

* New Database Sinks: Implementing the TimeSeriesWriter interface for other popular time-series databases like Prometheus, TimescaleDB, and ClickHouse.

* Enhanced Configuration: Adding support for configuration files (e.g., YAML) to manage settings more easily than with environment variables.

* Advanced Data Processing: Implementing optional filters or transformations for the data stream before it's ingested.

## Documentation

For a detailed breakdown of the data formats handled by this project, including the SBS-1 and BST file formats, please see our dedicated documentation:

* [Socket Data and BST File Formats](docs/sbs-bst-formats.md)

## Contributing
We welcome contributions from the community!

* Reporting Bugs: If you find a bug, please open an issue on the GitHub repository.

* Feature Requests: If you have an idea for a new feature, please open an issue to discuss it.

* Pull Requests: Feel free to submit pull requests for bug fixes, new features, or improvements to the documentation.

## License
This project is licensed under the MIT License — see the [LICENSE file](LICENSE) for details.
