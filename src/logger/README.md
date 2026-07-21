# Logger

A centralised log aggregation service for the simulation. Every other service in the platform forwards its structured log messages here via UDP. The logger stores them locally as JSONL files, keeps files from growing without bound through periodic rotation, and exposes a simple HTTP endpoint for downloading the full log archive as a zip when the simulation is done.

---

## How it works

The service starts three independent components and runs them concurrently:

1. **UDP server** — listens on `UDP_ADDR` (default `0.0.0.0:5000`) for incoming log messages. Each packet is parsed as a `LogMessage` JSON object and handed off to the processor's internal queue. Invalid JSON is silently discarded.
2. **Log processor** — a pool of 5 worker goroutines that drain the queue and persist each message to disk. The queue is bounded at 5000 entries; if it fills up the UDP server drops incoming messages rather than blocking.
3. **HTTP server** — listens on `HTTP_ADDR` (default `0.0.0.0:8080`) and serves a single endpoint for downloading all stored logs as a zip file.
4. **Log rotation** — a background goroutine that runs every 10 seconds and truncates any log file that has grown beyond 10 000 lines by dropping the oldest 50%.

All four run independently. Shutting down the service (SIGINT/SIGTERM) cancels the UDP read loop, drains the processor queue, stops the rotation worker, and gracefully shuts down the HTTP server.

### Storage layout

Log messages are persisted as JSONL (one JSON object per line) inside the directory configured by `LOGS_DIR` (default `./data/logs`). The filename for each file is derived from the message's `ServiceID` and `InstanceID` fields:

```
<ServiceID>_<InstanceID>.jsonl
```

Both fields are sanitised before use (only alphanumeric characters, dashes, and underscores are kept). If both fields are empty the file falls back to `unknown.jsonl`. Each write appends to the file; concurrent writes to the same file are protected by a per-file mutex.

Example filenames:

```
mission_uav_1.jsonl
followme_uav_2.jsonl
application_uav_1.jsonl
external_comms_uav_1.jsonl
```

---

## Inputs

### Log messages — received on UDP port 5000

Each message must be a flat JSON object. Unknown fields are ignored. The logger enriches it with a `ReceivedAt` timestamp before writing.

```json
{
  "InstanceID": "uav_1",
  "ServiceID": "mission",
  "Level": "INFO",
  "Timestamp": "2026-01-01T12:00:00Z",
  "Message": "Reached Waypoint 2",
  "EventID": "wp_reached"
}
```

| Field | Required | Description |
|-------|----------|-------------|
| `InstanceID` | No | Identifies the UAV instance. Used as part of the filename. |
| `ServiceID` | No | Identifies the service that generated the log. Used as part of the filename. |
| `Level` | No | Severity level (e.g. `INFO`, `WARN`, `ERROR`). Stored but not filtered. |
| `Timestamp` | No | Timestamp as set by the sender. |
| `Message` | No | The log message text. |
| `EventID` | No | Optional event identifier for structured querying. |

`ReceivedAt` is added automatically by the logger and is always present in the stored records.

---

## Outputs

### HTTP endpoint — GET `/api/logs/download`

Returns all stored JSONL files packaged into a single zip archive.

```
GET http://<logger>:8080/api/logs/download
Content-Type: application/zip
Content-Disposition: attachment; filename="logs.zip"
```

The zip contains one `.jsonl` file per unique `ServiceID`/`InstanceID` combination. If an error occurs mid-stream (e.g. a file disappears during the zip), the error is logged to stdout but the response cannot be rolled back since headers are already sent.

---

## Configuration

Configuration is loaded from `config.json` at startup. If the file is absent or a field is missing, built-in defaults are used.

| Parameter | Default | Description |
|-----------|---------|-------------|
| `udp_addr` | `0.0.0.0:5000` | Address and port the UDP server binds to. |
| `http_addr` | `0.0.0.0:8080` | Address and port the HTTP server listens on. |
| `logs_dir` | `./data/logs` | Directory where JSONL log files are stored. Created on startup if it doesn't exist. |
| `max_lines` | `10000` | Max lines per file before truncating the oldest 50%. |
| `buffer_size` | `5000` | Capacity of the internal log processing queue. |
| `workers_count` | `5` | Number of goroutines processing incoming logs to disk. |
| `truncation_interval_s` | `10` | Interval in seconds for the background log rotation job. |

---

## Running locally

```bash
go run ./cmd/logger config.json
```

---

## Code structure

```
logger/
├── cmd/logger/main.go                     Entry point. Wires all components and manages graceful shutdown.
├── domain/log_message.go                  Core type: LogMessage.
├── ports/
│   ├── server.go                          Server interface (Start/Stop).
│   └── storage.go                         LogStorage interface (SaveLog, WriteZippedLogs, RotateLogs).
├── usecases/
│   ├── process_log.go                     Async log processor with a worker pool and bounded queue.
│   ├── download_logs.go                   Zip generation use case.
│   └── truncate_logs.go                   Background log rotation (drops oldest 50% when file exceeds maxLines).
└── infra/
    ├── config/loader.go                   Loads and merges config.json with defaults.
    ├── udp/server.go                      UDP listener that enqueues incoming LogMessage packets.
    ├── http/server.go                     HTTP server with the /api/logs/download endpoint.
    └── storage/local_file_storage.go      JSONL file storage with per-file locking and rotation.
```

---

## Known limitations

- **The processor queue is bounded but drops silently.** When the queue of 5000 entries is full, new log messages are dropped. There is a `fmt.Println` warning, but no metrics or backpressure mechanism.
- **Log rotation reads entire files into memory.** The rotation logic reads every line of a file before truncating it. For very verbose services this could use significant memory, though the 10 000-line cap is intended to prevent this from becoming a problem in practice.
- **The HTTP download holds a global lock.** `WriteZippedLogs` acquires a global mutex on the storage for the duration of the zip operation. This blocks all concurrent writes until the download completes.
- **No authentication on the HTTP endpoint.** Any client that can reach port 8080 can download the full log archive.
- **`EventID` is stored but not indexed.** There is no way to query by event ID through the current API — the only export is the full raw JSONL dump.
