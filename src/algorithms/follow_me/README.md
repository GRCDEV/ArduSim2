# Follow Me Algorithm

A master/slave swarm behaviour where one UAV (the master) broadcasts its position at a fixed rate and all others (slaves) autonomously take off and mirror its movements in real time. The role is determined at startup from `config.json` — a single binary handles both sides of the protocol.

---

## How it works

Each instance of the algorithm reads its `role` from configuration and instantiates either the master or the slave logic. Both share the same connection and message loop; only the handling differs.

### Master

The master is passive from a movement perspective — it doesn't send any suggestions to its own UAV controller. Its job is to watch its own telemetry and forward it to the swarm.

1. On `start`, the master activates a `ChronJob` that fires every `send_period_ms` milliseconds.
2. Each tick takes the last received telemetry snapshot and publishes it as a `FollowMeMessage` to `broadcast_topic`. The message is wrapped with the `subscription_topic` as the inner topic so that slave instances listening on that topic receive it directly.
3. The master keeps an eye on its own altitude. When it detects the vehicle is landing (altitude below 0.5 m and decreasing relative to the previous sample), it stops the broadcast ticker, sends a `Land` suggestion to its own UAV controller, and publishes a `finish` command to `broadcast_topic`.
4. `pause` stops the ticker without resetting it. `stop` lands the master UAV and tears down the ticker entirely.

### Slave

The slave sits in two modes: waiting for commands, and following.

1. On `start`, the slave transitions to `RUNNING`. It doesn't arm or take off at this point.
2. On each position broadcast from the master (arriving on `subscription_topic`):
   - If the master is on the ground (`relative_alt < 0.5 m`) and the slave is flying, it lands immediately.
   - If the master is flying and the slave hasn't taken off yet, it arms, sets `GUIDED` mode, and takes off to `slaves_takeoff_altitude`. State is set to `TAKEOFF`.
   - Once the slave's own telemetry confirms it has exceeded `slaves_takeoff_altitude`, state transitions to `RUNNING`.
   - While in `RUNNING`, every master broadcast triggers a `MoveToPosition` suggestion with the master's exact lat/lon/alt.
3. `pause` transitions back to `IDLE` — no movement commands are sent while paused.
4. `stop` lands and resets the `alreadyTakeOff` flag so a subsequent `start` will trigger a fresh takeoff sequence.

### Message routing

Both master and slave subscribe to the same two topics at startup: `telemetry_topic` (own UAV telemetry) and `subscription_topic` (control commands and master position broadcasts). The slave's `HandleSubscriptionTopic` tries to parse incoming payloads as a command first; if that fails (missing or empty `command` field), it treats the payload as a master telemetry message. This is how commands from the GUI and position broadcasts from the master share the same topic without ambiguity.

---

## States

| State | Master | Slave |
|-------|--------|-------|
| `IDLE` | Waiting for `start`. Broadcast ticker stopped. | Waiting for `start`. Ignores master telemetry. |
| `RUNNING` | Broadcasting position on schedule. | Sending `MoveToPosition` on every master broadcast. |
| `TAKEOFF` | *(not used)* | Arms, switches to GUIDED, takes off. Transitions to `RUNNING` when altitude target is reached. |

---

## Inputs and outputs

### Messages consumed

**Control commands** — topic: `subscription_topic` (default `algo/followme`)

```json
{ "topic": "algo/followme", "payload": { "command": "<cmd>" } }
```

| Command | Effect |
|---------|--------|
| `start` | Activates the algorithm. Master starts the broadcast ticker; slave transitions to RUNNING. |
| `resume` | Same as `start`. |
| `pause` | Master stops the ticker. Slave stops tracking. |
| `stop` | Both roles land and reset. |

**Master position broadcasts** — topic: `subscription_topic` (slave only)

The slave distinguishes these from commands because they lack a `command` field:

```json
{
  "topic": "algo/followme",
  "payload": {
    "lat": 39.4816,
    "lon": -0.3492,
    "alt": 512.3,
    "relative_alt": 15.0,
    "heading": 270.0,
    "timestamp": 1748000000000
  }
}
```

**Own telemetry** — topic: `telemetry_topic` (default `uav/telemetry`)

```json
{
  "topic": "uav/telemetry",
  "payload": {
    "nr_gps_online": 1,
    "position": {
      "lat": 39.4816,
      "lon": -0.3492,
      "alt": 512.3,
      "relative_alt": 15.0,
      "heading": 270.0
    },
    "speed": {
      "vx": 1.2,
      "vy": 0.3,
      "vz": 0.0
    }
  }
}
```

### Messages published

**UAV suggestions** — topic: `suggestions_topic` (default `uav/suggestions`)

```json
{ "topic": "uav/suggestions", "payload": { "endpoint": "Arm" } }
{ "topic": "uav/suggestions", "payload": { "endpoint": "SetFlightmode", "flightmode": "GUIDED" } }
{ "topic": "uav/suggestions", "payload": { "endpoint": "Takeoff", "altitude": 5.0 } }
{ "topic": "uav/suggestions", "payload": { "endpoint": "MoveToPosition", "latitude": 39.48, "longitude": -0.34, "altitude": 512.3 } }
{ "topic": "uav/suggestions", "payload": { "endpoint": "Land" } }
```

**Master position broadcasts** — topic: `broadcast_topic` (master only, default `external/messages`)

The master wraps its position in an outer envelope so the broker can route it back to instances subscribed to `subscription_topic`:

```json
{
  "topic": "external/messages",
  "payload": {
    "topic": "algo/followme",
    "payload": {
      "lat": 39.4816,
      "lon": -0.3492,
      "alt": 512.3,
      "relative_alt": 15.0,
      "heading": 270.0,
      "timestamp": 1748000000000
    }
  }
}
```

**Finish event** — topic: `broadcast_topic` (master only)

Published when the master detects it has landed:

```json
{ "topic": "external/messages", "payload": { "command": "finish", "source": "followme" } }
```

**Logs** — topic: `logs_topic` (default `uav/logs`)

```json
{
  "topic": "uav/logs",
  "payload": {
    "InstanceID": "uav_1",
    "ServiceID": "followme",
    "Level": "INFO",
    "Timestamp": "2026-01-01T12:00:00Z",
    "Message": "FollowMe Algorithm started as master"
  }
}
```

`InstanceID` is `uav_<UAV_ID>` where `UAV_ID` comes from the environment variable of the same name (defaults to `"unknown"`).

---

## Configuration

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `role` | enum | `master` | Either `master` or `slave`. Determines which behaviour is instantiated at startup. Panics on any other value. |
| `slaves_takeoff_altitude` | number | `5.0` | Target altitude (m) for slave takeoff. The slave transitions from `TAKEOFF` to `RUNNING` once its own telemetry exceeds this. |
| `send_period_ms` | integer | `1000` | How often the master broadcasts its position, in milliseconds. |
| `broker_ip` | string | `communication_module` | Hostname or IP of the UDP broker. |
| `broker_port` | number | `3400` | UDP port of the broker. |
| `telemetry_topic` | string | `uav/telemetry` | Topic delivering the UAV's own telemetry. |
| `suggestions_topic` | string | `uav/suggestions` | Topic for publishing UAV control suggestions. |
| `broadcast_topic` | string | `external/messages` | Topic the master uses to send position broadcasts and the finish event. |
| `subscription_topic` | string | `algo/followme` | Topic for receiving control commands (all roles) and master position messages (slave only). |
| `logs_topic` | string | `uav/logs` | Topic for structured log messages. |

---

## Running locally

```bash
go run cmd/main.go config.json
```

---

## Code structure

```
follow_me/
├── cmd/main.go                              Entry point. Loads config, wires broker and logger, delegates to factory.
├── domain/models.go                         Core types: State, Config, Telemetry, FollowMeMessage, Suggestion.
├── ports/interfaces.go                      Interfaces: CommunicationProvider, FollowMeManager, FollowMeHandler.
├── usecase/
│   ├── follow_me_base.go                    Shared connection loop and command dispatcher. Both roles embed this.
│   ├── follow_me_as_master.go               Master logic: periodic broadcast, landing detection.
│   └── follow_me_as_slave.go                Slave logic: autonomous takeoff and position tracking.
└── infrastructure/
    ├── broker/udp_broker.go                 UDP pub/sub client.
    ├── chron-job/chron-job.go               Tick-based job runner used by the master for periodic broadcasts.
    ├── follow_me/follow_me_factory.go       Factory that instantiates master or slave based on config.role.
    └── broker_log_writer.go                 io.Writer that forwards log output to the broker.
```

The role-specific behaviour is cleanly isolated: `FollowMeBase` holds the connection and message loop, and delegates per-topic handling to whichever `FollowMeHandler` implementation is embedded. The factory in `infrastructure/follow_me/` is the only place that knows about the two concrete types.

---

## Known limitations

- **The slave tracks the master's absolute coordinates, not an offset.** There is no formation logic: all slaves converge on the master's exact position simultaneously. The `MetersToDegrees` helper in `domain/models.go` exists but is not used by the current movement logic.
- **Landing detection on the master is heuristic.** It checks that `relative_alt < 0.5 m` AND that altitude is decreasing compared to the previous sample. This means it won't trigger if the master was never flying or if telemetry arrives out of order.
- **`ChronJob.Stop` can block.** If the stop channel is full (capacity 1), `Stop()` will block until the goroutine reads from it. In practice this is fine as long as `Stop` is not called in rapid succession.
- **No explicit TAKEOFF state for the master.** The master skips the armed/takeoff sequence entirely — it assumes its own UAV controller handles that separately before `start` is sent.
