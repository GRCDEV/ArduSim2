# NetSim

A distributed network simulator node for ArduSim2. It models how radio messages propagate between UAVs in a swarm — accounting for range limits, channel contention (CSMA), and packet loss — so that algorithms running in simulation experience realistic inter-UAV communication constraints rather than a perfect broadcast bus.

Each NetSim instance is responsible for a subset of the UAV fleet, assigned by the gateway through consistent hashing. Multiple NetSim nodes can run in parallel to scale the simulation.

---

## How it works

The simulator receives two types of input from the gateway: **telemetry updates** (a UAV's current position) and **broadcast requests** (a UAV wants to send a message to its peers). For each broadcast, it decides which other UAVs can receive it and when, then sends the deliverable messages back to the gateway for forwarding to the target UAVs.

### Spatial partitioning

UAVs are bucketed into 3D chunks of configurable size (`chunk_size_m`). When processing a broadcast, the simulator only considers UAVs in neighbouring chunks (within `chunk_radius` chunks of the sender) as potential receivers. This reduces the per-broadcast cost from O(N) to roughly O(local density).

### Loss modes

Configured via `loss_mode`:

**`realistic` (default)** — Uses a CSMA/CA-inspired model with a probabilistic distance-based loss function:

1. Checks distance: receiver must be within `max_range_m` of the sender.
2. Applies a quadratic probability model — loss probability increases with distance squared. Even within range, distant packets may be dropped.
3. Checks receiver availability: if the receiver's radio is busy (still processing a previous message), the new message is dropped or delayed.
4. Checks receiver buffer: if the receiver's buffer (`buffer_size_bytes`) is full, the message is dropped.
5. Computes transmission time based on payload size.
6. If a collision is detected (overlapping transmissions within `csma_range_m`), the message is retried with exponential backoff up to `max_csma_retries` times. After that it is dropped.

**`fixed_range`** — Hard range cutoff with no probabilistic component. A message is delivered if and only if the receiver is within `max_range_m` of the sender. All other CSMA/buffer/busy checks still apply. Useful when you want deterministic range behaviour without the stochastic loss of `realistic`.

**`unrestricted`** — Delivers every message to every UAV with a known position, regardless of distance, channel state, or buffer. Useful for testing algorithm logic in isolation from network effects.

### Snapshot emission

Every `snapshot_interval_s` seconds, the simulator publishes the last known telemetry of every UAV it tracks to the gateway with topic `snapshot`. The gateway aggregates snapshots from all nodes and uses them to answer UI queries about swarm state.

### Multi-node federation

When a UAV broadcast arrives, the simulator also sends a `broadcast_notify` message to the gateway. The gateway then forwards this as a `peer_broadcast` to all other NetSim nodes. Each peer applies its own CSMA/range checks for the UAVs it manages, preventing duplicate deliveries and preserving realistic range constraints across node boundaries.

---

## Message protocol

The simulator communicates with the netsim_gateway exclusively via UDP. All messages are JSON with a top-level `topic` field.

### Received from gateway

| Topic | Description |
|-------|-------------|
| `uav_telemetry` | Position update for a UAV. Updates the internal UAV registry. |
| `uav_broadcast` | Broadcast request from a UAV. Triggers CSMA/range processing. |
| `peer_broadcast` | Broadcast notification from another NetSim node. Processed only if `origin_node` differs from this node's ID. |

**`uav_telemetry` payload:**
```json
{
  "uav_id": "1",
  "payload": {
    "lat": 39.4816,
    "lon": -0.3492,
    "alt": 512.3,
    "relative_alt": 15.0,
    "heading": 270.0
  }
}
```

**`uav_broadcast` payload:**
```json
{
  "uav_id": "1",
  "payload": { "topic": "algo/followme", "payload": { ... } }
}
```

### Sent to gateway

| Topic | Description |
|-------|-------------|
| `deliver` | A message that should be forwarded to a specific UAV. |
| `broadcast_notify` | Notification that a UAV broadcast was processed, for federation to other NetSim nodes. |
| `snapshot` | Periodic telemetry snapshot of all tracked UAVs. |

**`deliver` payload:**
```json
{
  "topic": "deliver",
  "payload": {
    "target_uav_id": "2",
    "sender_id": "1",
    "payload": "{ ... }"
  }
}
```

**`snapshot` payload:**
```json
{
  "topic": "snapshot",
  "payload": {
    "node_id": "node-a",
    "uavs": {
      "1": { "lat": 39.48, "lon": -0.34, "alt": 512.3, "heading": 270.0 },
      "2": { "lat": 39.49, "lon": -0.35, "alt": 510.0, "heading": 90.0 }
    }
  }
}
```

---

## Configuration

Configuration is loaded from `config.json` at startup. If the file is absent or a field is missing, built-in defaults are used.

| Parameter | Default | Description |
|-----------|---------|-------------|
| `listen_port` | `3000` | UDP port to receive messages from the gateway. |
| `gateway_addr` | `netsim_gateway:3001` | Address of the gateway's netsim-facing port. Telemetry snapshots and delivery packets are sent here. |
| `loss_mode` | `realistic` | `realistic` (CSMA/range) or `unrestricted` (no losses). |
| `buffer_size_bytes` | `163840` | Per-UAV receive buffer size. Messages that would exceed it are dropped. |
| `csma_range_m` | `700.0` | Distance within which two simultaneous transmissions are considered colliding. |
| `max_csma_retries` | `5` | Maximum retry attempts for a colliding message before it is dropped. |
| `chunk_size_m` | `350.0` | Size of each spatial chunk in metres (used for neighbour lookup optimisation). |
| `chunk_radius` | `2` | How many chunks away from the sender to consider as potential receivers. |
| `max_range_m` | `1350.0` | Maximum distance at which a message can be received at all. |
| `snapshot_interval_s` | `1` | How often (seconds) the node emits a telemetry snapshot. |
| `flush_interval_ms` | `1` | How often (milliseconds) the simulator flushes pending messages to the gateway. |
| `level` | `info` | Log level for the internal UDP logger (`debug`, `info`, `warn`, `error`). |
| `logger_addr` | `logger:5000` | Address of the logger microservice for sending internal logs. |

`NODE_ID` is read from the environment variable of the same name at startup. It is used to identify this node in `peer_broadcast` messages and snapshot payloads, preventing echo loops in the multi-node federation.

---

## Running locally

```bash
NODE_ID=node-a go run .
```

---

## Code structure

```
netsim/
├── main.go                                Entry point. Wires config, connections, simulator, and goroutines.
├── domain/
│   ├── model/
│   │   ├── message.go                     Message and DelayedMessage types.
│   │   ├── position.go                    UAV position and telemetry data types.
│   │   └── uav.go                         UAV state (position, buffer, channel availability).
│   └── service/
│       ├── spatial.go                     SpatialGrid: chunk-based UAV neighbour lookup.
│       ├── loss.go                        Signal propagation and packet loss model.
│       └── csma.go                        CSMA channel busy-time calculation.
├── usecase/
│   ├── simulator.go                       Core Simulator struct, message dispatch, gateway notification.
│   ├── enqueue_broadcast.go               Broadcast intake and per-receiver delivery logic.
│   ├── flush_messages.go                  Dequeues pending messages and retries delayed ones.
│   ├── emit_snapshot.go                   Periodic snapshot emitter goroutine.
│   ├── update_telemetry.go                Handles telemetry updates from the gateway.
│   ├── broadcast_strategy.go              BroadcastStrategy interface.
│   ├── strategy_csma.go                   Realistic CSMA strategy implementation.
│   └── strategy_unrestricted.go           Unrestricted (no-loss) strategy implementation.
├── ports/
│   ├── input/                             RawPacket and handler interfaces.
│   └── output/                            PacketSender and Logger interfaces.
└── infra/
    ├── config/loader.go                   Loads and merges config.json with defaults.
    ├── udp/                               UDP connection, sender, and receiver implementations.
    └── logger/                            UDP logger client for forwarding internal logs to the logger service.
```

---

## Known limitations

- **Retry logic has a locking gap.** `flush_messages.go` unlocks the mutex before calling `EnqueueBroadcast` for retried messages. There is a brief window between unlock and re-lock where the state could change.
