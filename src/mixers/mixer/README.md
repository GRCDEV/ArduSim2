# Mixer

The central orchestrator of ArduSim2. Every UAV in the simulation has one instance of this service running alongside it. It sits between the algorithms and the UAV controller: it collects movement suggestions from all running algorithms, arbitrates between them through a time-windowed mixing strategy, and forwards the result directly to the UAV controller. It also receives raw telemetry from the UAV controller and rebroadcasts it to the broker so that every subscribed algorithm can see the UAV's current state.

---

## How it works

On startup the service connects to two separate transports:

1. **The UDP broker** (the communication module) — for receiving suggestions from algorithms and publishing telemetry back to them.
2. **A direct UDP link to the UAV controller** — for sending commands and receiving raw telemetry.

Both connections use a retry loop with a 5-second sleep between attempts. The service does not start the main loop until both are established.

Once connected, it runs a single `select` loop with three cases:

### Telemetry bridge

Telemetry packets arriving directly from the UAV controller are forwarded to the broker under `telemetry_topic` as-is. No transformation is applied — whatever JSON the UAV controller sends lands verbatim on the broker topic.

### Suggestion intake

Messages arriving on `suggestions_topic` are decoded into a `Suggestion` struct and appended to an in-memory buffer. The buffer is protected by a mutex because it is written from the broker goroutine and read from the ticker goroutine.

### Mix window evaluation

A `time.Ticker` fires every `mix_window_ms` milliseconds. On each tick the current buffer is swapped out for an empty one and the accumulated suggestions are processed by the movement mixer.

---

## The movement mixer

The mixer's job is to reduce N suggestions from potentially multiple concurrent algorithms into a minimal ordered sequence of commands that is safe to send to the UAV controller.

**Classification.** Each suggestion is first classified:

| Category | Action types |
|----------|-------------|
| Structural | `Arm`, `Disarm`, `SetFlightmode`, `Takeoff`, `Land`, `RecoverControl`, `SetMessageInterval`, `RequestMessage` |
| Movement | `MoveToPosition`, `MoveByVector`, `Rotate` |
| Critical | `Land`, or `SetFlightmode` with `flightmode` = `RTL`, `BRAKE`, or `LAND` |

**Deduplication.** For each movement category only the *last* suggestion in the window survives. If five `MoveToPosition` commands arrived in a 200 ms window, only the most recent one is sent. This reflects the assumption that later suggestions supersede earlier ones from the same source.

**Critical override.** If any suggestion in the window is critical (Land/RTL/BRAKE), all movement suggestions for that window are discarded entirely. Only structural commands proceed.

**Ordering.** The surviving suggestions are sorted by their original arrival index and sent to the UAV controller in that order. Structural commands (including the critical ones) keep their relative ordering to respect sequencing that matters — e.g. `Arm` before `Takeoff`.

**Global stop.** After forwarding a critical command to the UAV controller, the mixer publishes `{"command": "stop"}` to `global_commands`. This lets all algorithm containers react to a landing or abort event without polling the UAV's state.

---

## Inputs and outputs

### Messages consumed

**Algorithm suggestions** — topic: `suggestions_topic` (default `uav/suggestions`)

```json
{
  "topic": "uav/suggestions",
  "payload": { 
    "endpoint": "MoveToPosition", 
    "latitude": 39.48, 
    "longitude": -0.34, 
    "altitude": 15.0 
  }
}
```

All `endpoint` values the mixer recognises:

| Endpoint | Category | Additional fields |
|----------|----------|-------------------|
| `Arm` | Structural | — |
| `Disarm` | Structural | — |
| `Takeoff` | Structural | `altitude` |
| `Land` | Structural + Critical | — |
| `SetFlightmode` | Structural (Critical if `RTL`/`BRAKE`/`LAND`) | `flightmode` |
| `RecoverControl` | Structural | — |
| `SetMessageInterval` | Structural | `messageID` |
| `RequestMessage` | Structural | `messageID` |
| `MoveToPosition` | Movement | `latitude`, `longitude`, `altitude`, `yaw` |
| `MoveByVector` | Movement | `vx`, `vy`, `vz` |
| `Rotate` | Movement | `yaw`, `speed`, `direction`, `relative` |

**Raw telemetry from UAV controller** — received on `uav_telemetry_port` (UDP, default `3500`)

The telemetry payload is an opaque JSON object — its structure is defined by the UAV controller. The application forwards it without modification.

### Messages published

**Telemetry broadcast** — topic: `telemetry_topic` (default `uav/telemetry`)

The raw telemetry packet from the UAV controller, republished to the broker. All algorithms subscribe to this to track the UAV's position and state.

**Global stop** — topic: `global_commands` (default `global/commands`)

```json
{ "topic": "global/commands", "payload": { "command": "stop" } }
```

Published once per critical command that makes it through the mixer. Algorithms that subscribe to `global_commands` can use this to halt their own execution without needing to monitor UAV altitude.

**Commands to UAV controller** — sent directly over UDP to `uav_controller_ip:uav_controller_port` (default port `9876`)

The mixer sends each surviving suggestion directly to the UAV controller as a JSON-encoded `Suggestion` struct. This bypasses the broker entirely.

**Logs** — topic: `logs_topic` (default `uav/logs`)

```json
{
  "topic": "uav/logs",
  "payload": {
    "InstanceID": "uav_<UAV_ID>",
    "ServiceID": "application",
    "Level": "INFO",
    "Timestamp": "2026-01-01T12:00:00Z",
    "Message": "Executing Suggestion (Idx 0): MoveToPosition"
  }
}
```

`InstanceID` is `uav_<UAV_ID>` where `UAV_ID` comes from the environment variable of the same name (defaults to `"unknown"`).

---

## Configuration

The service does not use a `schema.json` — it is not an algorithm and is not configured through the GUI. Its `config.json` is baked into the container at build time.

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `broker_ip` | string | `communication_module` | Hostname or IP of the UDP broker. Resolves via Docker networking. |
| `broker_port` | number | `3400` | UDP port of the broker. |
| `uav_controller_ip` | string | `uav_controller` | Hostname or IP of the UAV controller container. |
| `uav_controller_port` | number | `9876` | Port for sending commands to the UAV controller. |
| `uav_telemetry_port` | number | `3500` | Local port the application binds to in order to receive telemetry from the UAV controller. |
| `telemetry_topic` | string | `uav/telemetry` | Broker topic where raw telemetry is rebroadcast. |
| `suggestions_topic` | string | `uav/suggestions` | Broker topic the mixer listens on for algorithm suggestions. |
| `external_telemetry` | string | `external/telemetry` | Topic declared in config but currently not published to by the mixer. |
| `external_messages` | string | `external/messages` | Topic declared in config but currently not published to by the mixer. |
| `global_commands` | string | `global/commands` | Broker topic used to broadcast stop events after critical commands. |
| `logs_topic` | string | `uav/logs` | Topic for structured log messages. |
| `mix_window_ms` | number | `200` | Width of the mixing window in milliseconds. If zero or missing, defaults to 200 ms. |

---

## Running locally

```bash
go run cmd/main.go config.json
```

The service requires both the broker and the UAV controller to be reachable. In local development without Docker, the retry loops will keep the service alive until both become available, but it won't do anything useful until they are.

---

## Code structure

```
application/
├── cmd/main.go                            Entry point. Loads config, connects broker and UAV link, starts mixer.
├── domain/models.go                       Core types: Suggestion, AppConfig, ActionType constants.
├── ports/interfaces.go                    Interfaces: Broker, UAVLink, ConfigLoader.
├── usecase/movement_mixer.go              Mixing logic: intake, classification, deduplication, critical override.
└── infrastructure/
    ├── broker/udp_broker.go               UDP pub/sub client for the communication module.
    ├── broker/broker_log_writer.go        io.Writer that forwards log output to the broker.
    ├── config/file_loader.go              Reads and unmarshals config.json.
    └── uav/uav_link.go                    Direct UDP client for the UAV controller (send + receive).
```

The service maintains two separate UDP sockets: one managed by the broker client (ephemeral local port, sends to the broker) and one managed by the UAV link (binds to `uav_telemetry_port`, sends to the UAV controller). They are independent and can fail independently.

---

## Known limitations

- **No averaging of movement suggestions.** Despite what the original README described, the mixer does not average multiple `MoveToPosition` suggestions — it picks the last one. If two algorithms disagree on where to move, the one that sent its suggestion most recently in the window wins unconditionally.
- **The telemetry bridge does not filter or validate.** Any parseable JSON arriving on `uav_telemetry_port` is forwarded to the broker. A malformed or unexpected payload from the UAV controller will be passed through without error.
- **The broker channel is buffered at 100.** If the mixer's ticker is delayed (e.g. the UAV link is slow) and suggestions arrive faster than they are consumed, the channel will eventually fill and the broker goroutine will silently drop messages.
