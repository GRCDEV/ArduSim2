# NetSim Gateway

The central router between the UAV fleet and the network simulation layer. It sits at the boundary between the per-UAV `external_comms` containers and the one or more `netsim` nodes that model radio propagation. Every broadcast a UAV sends passes through here on its way into the simulator, and every delivered message comes back here before being forwarded to the right UAV.

---

## How it works

The gateway opens two UDP ports and runs two independent receive loops:

- **UAV-facing port** (`uav_listen_port`, default `3000`) — receives messages from `external_comms` instances.
- **NetSim-facing port** (`netsim_listen_port`, default `3001`) — receives messages from `netsim` nodes.

Each incoming packet is parsed and dispatched based on its `topic` field.

### UAV → Gateway

When a UAV message arrives the gateway first checks the topic:

- **`subscribe`** — the sender is registered as a UI subscriber. Future broadcast notifications will be forwarded to it.
- **`telemetry`** / **`broadcast`** — the `uav_id` field in the payload is extracted. The sender's UDP address is stored in a registry keyed by that ID (so the gateway knows where to send replies later).
  - The UAV is assigned to a specific NetSim node using consistent hashing on `uav_id`. All traffic for a given UAV always goes to the same node, which prevents a UAV's state from being spread across multiple simulators.
  - The message is re-wrapped with a `uav_<topic>` prefix (`uav_telemetry`, `uav_broadcast`) and forwarded to the assigned node.
- **`broadcast` with empty `uav_id`** — treated as a direct broadcast to all registered UAVs, bypassing the simulator. Used internally for platform-level messages.

### NetSim → Gateway

Three topics arrive from the simulator:

- **`snapshot`** — telemetry data for all UAVs managed by that node. The gateway caches the latest telemetry for each UAV (used for UI subscribers and aggregated snapshots).
- **`deliver`** — a message that the simulator has decided should reach a specific UAV. The gateway looks up the UAV's address in the registry and forwards the message directly.
- **`broadcast_notify`** — the simulator is notifying the gateway that it processed a broadcast. The gateway:
  1. Forwards the broadcast payload to all registered UI subscribers.
  2. Propagates it as a `peer_broadcast` to all other NetSim nodes (excluding the one that sent the notification), enabling multi-node federation.

### NetSim discovery

The gateway discovers NetSim node addresses from the `ADDRS` environment variable (a comma-separated list). At startup it tries to resolve all addresses and blocks in a loop until all of them are reachable. It then passes the resolved addresses to the Gateway use case.

---

## Message protocol

### Received from UAVs (port 3000)

All messages are JSON with a top-level `topic` and `payload`.

| Topic | Description |
|-------|-------------|
| `subscribe` | Registers the sender as a UI subscriber. |
| `telemetry` | UAV position update. Forwarded to the assigned NetSim node as `uav_telemetry`. |
| `broadcast` | UAV wants to broadcast to peers. Forwarded to the assigned NetSim node as `uav_broadcast`. |

**Telemetry message (from `external_comms`):**
```json
{
  "topic": "telemetry",
  "payload": {
    "uav_id": "1",
    "payload": { 
      "lat": 39.48, 
      "lon": -0.34, 
      "alt": 512.3, 
      "relative_alt": 15.0, 
      "heading": 270.0 
    }
  }
}
```

**Broadcast message (from `external_comms`):**
```json
{
  "topic": "broadcast",
  "payload": {
    "uav_id": "1",
    "payload": { "topic": "algo/followme", "payload": { ... } }
  }
}
```

### Received from NetSim nodes (port 3001)

| Topic | Description |
|-------|-------------|
| `snapshot` | Telemetry snapshot for all UAVs on that node. Cached and forwarded to UI subscribers. |
| `deliver` | A message to be delivered to a specific UAV. |
| `broadcast_notify` | Notification to forward to UI subscribers and other NetSim nodes. |

### Sent to UAVs (port 3000)

Delivered messages are sent directly to the target UAV's registered UDP address:

```json
{
  "uav_id": "1",
  "payload": { "topic": "algo/followme", "payload": { ... } }
}
```

### Sent to NetSim nodes (port 3001)

Forwarded UAV messages (with `uav_` prefix) and peer broadcast propagation:

```json
{
  "topic": "uav_broadcast",
  "payload": {
    "uav_id": "2",
    "payload": { "topic": "algo/followme", "payload": { ... } }
  }
}
```

```json
{
  "topic": "peer_broadcast",
  "payload": {
    "sender_id": "1",
    "sender_position": { "lat": 39.48, "lon": -0.34, "alt": 512.3 },
    "payload": { ... },
    "origin_node": "192.168.1.10:3001"
  }
}
```

---

## Configuration

| Parameter | Default | Description |
|-----------|---------|-------------|
| `uav_listen_port` | `3000` | UDP port for incoming UAV messages. |
| `netsim_listen_port` | `3001` | UDP port for incoming NetSim messages. |
| `snapshot_interval_s` | `1` | How often (seconds) the aggregated snapshot emitter runs. |
| `logger_addr` | `logger:5000` | Address of the logger microservice for internal logs. |

`ADDRS` is read from the environment variable of the same name. It must be a comma-separated list of `host:port` strings pointing to all NetSim nodes (e.g. `netsim:3001`). The gateway blocks until all addresses resolve successfully.

---

## Running locally

```bash
ADDRS=localhost:3001 go run .
```

Both the NetSim node(s) and the logger must be reachable before the gateway starts routing traffic.

---

## Code structure

```
netsim_gateway/
├── main.go                                Entry point. Loads config, discovers NetSim nodes, wires all components.
├── domain/
│   ├── model/
│   │   ├── registry.go                    UAV registry type.
│   │   └── snapshot.go                    TelemetrySnapshot type (received from NetSim nodes).
│   └── service/
│       └── consistent_hash.go             Maps a UAV ID to a NetSim node index.
├── usecase/
│   ├── gateway.go                         Core Gateway: UAV/NetSim message handlers, UAV registry, telemetry cache.
│   └── snapshot_emitter.go               Aggregated snapshot emitter goroutine.
├── ports/
│   ├── input/                             RawPacket, UAVMessageHandler, NetsimMessageHandler interfaces.
│   └── output/                            PacketSender and Logger interfaces.
└── infra/
    ├── config/loader.go                   Loads config.json, reads ADDRS env var, resolves NetSim addresses.
    ├── udp/                               UDP connection, sender, and receiver implementations.
    └── logger/                            UDP logger client.
```

---

## Known limitations

- **NetSim addresses are resolved once at startup.** If a NetSim node restarts and gets a new IP, the gateway won't notice unless it itself restarts. The discovery loop in `main.go` retries until all configured addresses resolve, but does not re-resolve them after startup.
- **The UAV registry is never pruned.** Once a UAV registers its address, the entry stays forever. If a UAV container restarts on a different port, it will re-register with the new address, overwriting the old one. But stale entries from UAVs that disappeared without re-registering accumulate silently.
- **Consistent hashing assumes a fixed number of NetSim nodes.** If nodes are added or removed after startup (`UpdateNetsims` is called), UAVs may be remapped to different nodes, losing their accumulated state on the old node.
- **UI subscribers are never removed.** Any client that sends a `subscribe` message stays registered until the gateway restarts, even if it has long since disconnected. Stale subscribers will receive delivery errors on every broadcast notification.
