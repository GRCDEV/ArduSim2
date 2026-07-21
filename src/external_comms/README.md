# External Communications

A gateway bridge that connects a single UAV's local messaging bus to the swarm-wide network simulator. Every UAV in the simulation has one instance of this service running alongside it. It handles two independent data flows: outbound (local telemetry and algorithm messages → NetSim) and inbound (broadcasts from other UAVs → local broker). It also acts as the log forwarding agent, collecting log messages from all local services and delivering them directly to the centralised logger microservice.

---

## How it works

On startup the service establishes three connections, each with a retry loop:

1. **Local broker** — subscribes to `sub_telemetry_topic`, `sub_messages_topic`, and `sub_logs_topic`.
2. **Network simulator (NetSim)** — a direct UDP socket used for sending and receiving swarm-wide messages.
3. **Logger** — an optional UDP connection to the logger microservice. If `logger_ip` and `logger_port` are not configured, this connection is skipped and logs go to stdout only.

The logger connection is the only one that is truly optional. If it can't be established at startup, the service continues without it. Once all connections are up, the bridge enters a `select` loop that processes two channels in parallel: messages from the local broker and messages arriving from NetSim.

### Outbound path (local → swarm)

| Local topic | NetSim envelope topic | Description |
|-------------|----------------------|-------------|
| `sub_telemetry_topic` | `"telemetry"` | Own UAV position data forwarded to the swarm. |
| `sub_messages_topic` | `"broadcast"` | Algorithm-generated events (e.g. waypoint notifications, follow-me broadcasts) forwarded to the swarm. |
| `sub_logs_topic` | *(direct to logger)* | Log messages forwarded directly to the logger service, not to NetSim. |

Messages sent to NetSim are wrapped in an outer envelope:

```json
{
  "topic": "telemetry",
  "payload": {
    "uav_id": "<UAV_ID>",
    "payload": { ... }
  }
}
```

`UAV_ID` is read from the `UAV_ID` environment variable at startup.

### Inbound path (swarm → local)

When a message arrives from NetSim:

1. If `uav_id` in the message matches the local `UAV_ID`, it is silently dropped (own-echo filter).
2. If the payload has no `topic` field, it is also dropped.
3. Otherwise, the payload is published as-is to the local broker. The `topic` field inside the payload determines which broker topic it lands on.

### Log forwarding

Log messages from all local services arrive on `sub_logs_topic` as structured JSON. The bridge forwards them directly to the logger microservice via the `UDPLoggerLink`. This link has a built-in queue: if the logger becomes temporarily unreachable, messages are buffered in memory and flushed on the next successful send. Malformed payloads are silently discarded from the queue.

The service also logs its own internal messages (startup events, forward confirmations, errors). These go through the same `DirectLogWriter` that wraps the logger link, so they are forwarded to the logger too — or just to stdout if the logger link is not configured.

---

## Inputs and outputs

### Messages consumed

**Local telemetry** — topic: `sub_telemetry_topic` (default `uav/telemetry`)

The UAV's own telemetry payload, published by the `application` service. Forwarded to NetSim with topic `"telemetry"`.

**Local messages** — topic: `sub_messages_topic` (default `external/messages`)

Algorithm-generated events — waypoint notifications, follow-me position broadcasts, finish events, etc. Forwarded to NetSim with topic `"broadcast"`.

**Log messages** — topic: `sub_logs_topic` (default `uav/logs`)

Structured log payloads from any local service:

```json
{
  "InstanceID": "uav_1",
  "ServiceID": "mission",
  "Level": "INFO",
  "Timestamp": "2026-01-01T12:00:00Z",
  "Message": "Reached Waypoint 2"
}
```

Forwarded directly to the logger microservice via UDP — not to NetSim.

**Messages from NetSim** — received on the NetSim UDP socket

```json
{
  "uav_id": "2",
  "payload": {
    "topic": "algo/followme",
    "payload": {
      "lat": 39.48,
      "lon": -0.34,
      "relative_alt": 15.0,
      "heading": 270.0,
      "timestamp": 1748000000000
    }
  }
}
```

After the own-echo and topic checks, the inner `payload` map is published directly to the local broker.

### Messages published

**To the local broker** — inbound NetSim messages re-published verbatim

The `Broker.Publish` method on this service does not add a topic wrapper — it sends the raw payload map directly to the broker. The routing is determined by the `topic` field that the NetSim embedded in the message body.

**To NetSim** — outbound telemetry and broadcast messages

See the envelope format described in the [Outbound path](#outbound-path-local--swarm) section above.

**To the logger** — log payloads forwarded over UDP

The logger link sends raw JSON payloads directly. No additional wrapping is applied beyond what the originating service already put in the log message.

**Own process logs** — sent to logger via `DirectLogWriter` and stdout

```json
{
  "InstanceID": "uav_<UAV_ID>",
  "ServiceID": "external_comms",
  "Level": "INFO",
  "Timestamp": "2026-01-01T12:00:00Z",
  "Message": "[Internal->External] Forwarded Telemetry"
}
```

---

## Configuration

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `broker_ip` | string | `communication_module` | Hostname or IP of the local UDP broker. |
| `broker_port` | number | `3400` | UDP port of the local broker. |
| `simulator_ip` | string | `netsim_gateway` | Hostname or IP of the network simulator. |
| `simulator_port` | number | `3000` | UDP port of the network simulator. |
| `logger_ip` | string | `logger` | Hostname or IP of the logger microservice. Leave empty to disable log forwarding. |
| `logger_port` | number | `5000` | UDP port of the logger microservice. |
| `sub_telemetry_topic` | string | `uav/telemetry` | Local broker topic to subscribe for own UAV telemetry. |
| `sub_messages_topic` | string | `external/messages` | Local broker topic to subscribe for outbound algorithm messages. |
| `sub_logs_topic` | string | `uav/logs` | Local broker topic to subscribe for log messages. |
| `pub_ext_telemetry_topic` | string | `external/telemetry` | *Declared in config but currently not used by the bridge.* |
| `pub_ext_messages_topic` | string | `global` | *Declared in config but currently not used by the bridge.* |

`UAV_ID` is not in `config.json` — it is read from the `UAV_ID` environment variable at startup and injected into `AppConfig.UAVId`.

---

## Running locally

```bash
go run cmd/main.go config.json
```

Both the local broker and the NetSim must be reachable. The logger is optional. Without Docker, you can set `UAV_ID` in the environment:

```bash
UAV_ID=1 go run cmd/main.go config.json
```

If `UAV_ID` is not set, the own-echo filter will use an empty string as the source ID, which means no inbound messages will be filtered out.

---

## Code structure

```
external_comms/
├── cmd/main.go                            Entry point. Loads config, wires all three connections, starts bridge.
├── domain/models.go                       Core types: AppConfig, SendedNetSimMessage, ReceivedNetSimMessage.
├── ports/interfaces.go                    Interfaces: Broker, NetSimLink, LoggerLink.
├── usecase/bridge.go                      Bridge logic: inbound/outbound routing and log forwarding.
└── infrastructure/
    ├── udp_broker.go                      UDP pub/sub client for the local communication module.
    ├── net_sim_link.go                    UDP client for sending and receiving NetSim messages.
    ├── udp_logger_link.go                 UDP client for the logger, with in-memory queuing on disconnect.
    ├── direct_log_writer.go               io.Writer that forwards Go log output to the logger link.
    └── file_loader.go                     Reads and unmarshals config.json.
```

---

## Known limitations

- **The own-echo filter relies on string equality.** If `UAV_ID` is unset, it defaults to `""`. Any message arriving from NetSim with `uav_id: ""` would be dropped as an echo even if it came from a different (misconfigured) peer.
- **The logger queue is unbounded.** If the logger is unreachable for an extended period, the in-memory queue will grow without limit. In a long-running simulation with verbose logging, this could become a memory problem.
- **Inbound NetSim messages are published without a topic envelope.** The `Broker.Publish` here sends the raw payload directly, relying on the `topic` field already embedded in the payload by the NetSim. This is intentional but differs from how other services publish to the broker — those always wrap in `{"topic": "...", "payload": {...}}`. If a future NetSim implementation changes how it embeds the topic, the routing will silently break.
