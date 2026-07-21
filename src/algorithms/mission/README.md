# Mission Algorithm

Executes a predefined flight route for a UAV. You hand it a KML file with a sequence of waypoints and it handles everything from arm and takeoff to navigating between points and triggering whatever end behaviour you configured. The algorithm communicates with the rest of the platform exclusively through the UDP broker, publishing suggestions to the UAV controller and receiving back telemetry.

---

## How it works

The algorithm is a straightforward state machine driven by two inputs: commands arriving on the control topic, and periodic telemetry updates from the UAV.

States:

| State | Description |
|-------|-------------|
| `IDLE` | Waiting for a `start` command. |
| `TAKEOFF` | Motors armed, UAV climbing to the first waypoint's altitude. Transitions to `FLYING` once the target altitude is within 1 m. |
| `FLYING` | Navigating between waypoints. On each telemetry update the algorithm checks the horizontal distance (haversine) to the current target. When it falls below `distance_to_waypoint_reached`, the waypoint is marked as reached and the next one is sent. |
| `PAUSED` | The `pause` command puts the UAV into `BRAKE` mode and freezes waypoint progression. `resume` switches back to `GUIDED` and continues. Also used internally as a transient state when `input_mission_delay` is non-zero. |
| `LANDING` | Triggered by `stop`, `rtl`, `emergency_land`, or the mission end handler. Transitions to `FINISHED` once the UAV's relative altitude drops below 0.5 m. |
| `FINISHED` | Terminal state. A `finish` event is published to `external_messages_topic`. |

On startup, the algorithm loads `config.json`, parses the KML file into a list of waypoints, then connects to the broker with a retry loop (5 s between attempts) until it succeeds. Both `subscription_topic` and `telemetry_topic` subscriptions are registered at that point via the `$subscribe` reserved topic.

### Relative movement mode

When `relative_movement` is `true`, the waypoints in the KML file are treated as offsets from the UAV's takeoff position rather than absolute coordinates. The algorithm waits until the first valid GPS fix arrives (non-zero lat/lon, at least one GPS online), sets that as the new origin, and translates all waypoints by applying the same bearing and distances from the KML's first waypoint.

This lets you design a route once and replay it from any starting location.

### Altitude handling

Waypoint altitudes come from the KML coordinates field. Two options can override them:

- **`override_include_altitude_values`** — ignores all KML altitudes and uses `waypoints_relative_altitude` for every waypoint.
- **`minimum_waypoint_relative_altitude`** — if a waypoint's altitude in the KML is zero or below this threshold, it is clamped up to this value. The same threshold is used as the takeoff altitude when the first waypoint has no altitude.

### Yaw control

If `override_include_yaw_values` is `true`, each `MoveToPosition` command includes a `yaw` field computed according to `yaw_value`:

| Strategy | Behaviour |
|----------|-----------|
| `Face next Waypoint` | Points the UAV toward the target waypoint. |
| `Face next Waypoint except RTL` | Same, but yaw is not applied during RTL. |
| `Face along GPS course` | Currently equivalent to `Face next Waypoint` (course-over-ground requires additional telemetry context). |
| `Fixed` | Sends `-1`, which in ArduPilot MAVLink means "don't change yaw". |

### Mission end behaviour

Configured via `mission_end`:

| Value | What happens |
|-------|--------------|
| `land` | Issues a `Land` command and transitions to `LANDING`. |
| `rtl` | Switches to `RTL` flight mode and transitions to `LANDING`. |
| `unmodified` *(default)* | Does nothing — the UAV holds position at the last waypoint. A `finish` event is still published to `external_messages_topic`. |

---

## Inputs and outputs

### Messages consumed

**Control commands** — topic: `subscription_topic` (default `algo/mission`)

```json
{ "topic": "algo/mission", "command": "<cmd>" }
```

| Command | Valid from state | Effect |
|---------|-----------------|--------|
| `start` | `IDLE` | Arms motors, sets `GUIDED` mode, initiates takeoff. |
| `resume` | `PAUSED` | Restores `GUIDED` mode and continues waypoint sequence. |
| `pause` | `FLYING` | Switches to `BRAKE` mode, freezes progression. |
| `stop` | any except `IDLE`, `FINISHED` | Lands immediately. |
| `rtl` | any | Activates `RTL` flight mode. |
| `emergency_land` | any | Lands immediately with no RTL. |

**Telemetry** — topic: `telemetry_topic` (default `uav/telemetry`)

The expected payload structure:

```json
{
  "topic": "uav/telemetry",
  "payload": {
    "position": {
      "lat": 39.4816,
      "lon": -0.3492,
      "relative_alt": 12.3
    },
    "nr_gps_online": 1
  }
}
```

`nr_gps_online` is only relevant when `relative_movement` is `true` — the algorithm waits for a non-zero value before locking in the home position.

### Messages published

All outgoing messages go through the broker with the following shapes.

**UAV suggestions** — topic: `publish_topic` (default `uav/suggestions`)

```json
{ "topic": "uav/suggestions", "payload": { "endpoint": "Arm" } }
{ "topic": "uav/suggestions", "payload": { "endpoint": "SetFlightmode", "flightmode": "GUIDED" } }
{ "topic": "uav/suggestions", "payload": { "endpoint": "Takeoff", "altitude": 10.0 } }
{ "topic": "uav/suggestions", "payload": { "endpoint": "MoveToPosition", "latitude": 39.48, "longitude": -0.34, "altitude": 15.0 } }
{ "topic": "uav/suggestions", "payload": { "endpoint": "MoveToPosition", "latitude": 39.48, "longitude": -0.34, "altitude": 15.0, "yaw": 45.0 } }
{ "topic": "uav/suggestions", "payload": { "endpoint": "Land" } }
```

**External events** — topic: `external_messages_topic` (default `external/messages`)

Progress notifications sent when a waypoint is reached and when the mission ends:

```json
{ "topic": "external/messages", "payload": { "command": "Waypoint 3 reached", "source": "mission" } }
{ "topic": "external/messages", "payload": { "command": "finish", "source": "mission" } }
```

**Logs** — topic: `logs_topic` (default `uav/logs`)

```json
{
  "topic": "uav/logs",
  "payload": {
    "InstanceID": "uav_1",
    "ServiceID": "mission",
    "Level": "INFO",
    "Timestamp": "2026-01-01T12:00:00Z",
    "Message": "Reached Waypoint 2"
  }
}
```

`InstanceID` is derived from the `UAV_ID` environment variable (falls back to `"unknown"` if not set).

---

## Configuration

The full set of parameters injected via `config.json`. All of these are generated by the GUI from `schema.json`.

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `mission_file` | string | — | Path to the `.kml` file containing the route. Required. |
| `distance_to_waypoint_reached` | number | `200.0` | Distance threshold in **centimetres** to consider a waypoint reached. |
| `minimum_waypoint_relative_altitude` | number | `5.0` | Minimum altitude (m) for any waypoint, also used as takeoff altitude if the KML doesn't specify one. |
| `waypoints_relative_altitude` | number | `5.0` | Uniform altitude (m) applied to all waypoints when `override_include_altitude_values` is `true`. |
| `override_include_altitude_values` | boolean | `false` | Ignore KML altitudes and use `waypoints_relative_altitude` everywhere. |
| `override_include_yaw_values` | boolean | `false` | Attach a computed yaw angle to every `MoveToPosition` command. |
| `yaw_value` | enum | `Face next Waypoint except RTL` | Yaw computation strategy. See [Yaw control](#yaw-control). |
| `mission_end` | enum | `unmodified` | What the UAV does after the last waypoint. `land`, `rtl`, or `unmodified`. |
| `final_altitude_for_rtl` | number | `5.0` | Altitude (m) for RTL. Currently informational — ArduPilot's RTL mode handles its own altitude. |
| `input_mission_delay` | number | `0.0` | Seconds to wait at each waypoint before advancing to the next one. |
| `relative_movement` | boolean | `false` | Treat KML waypoints as relative offsets from the actual takeoff position. |
| `broker_ip` | string | `communication_module` | Hostname or IP of the UDP broker. Resolves via Docker networking. |
| `broker_port` | number | `3400` | UDP port of the broker. |
| `subscription_topic` | string | `algo/mission` | Topic to receive control commands on. |
| `publish_topic` | string | `uav/suggestions` | Topic to publish UAV suggestions to. |
| `telemetry_topic` | string | `uav/telemetry` | Topic to subscribe to for UAV position updates. |
| `external_messages_topic` | string | `external/messages` | Topic for waypoint progress and mission finish events. |
| `logs_topic` | string | `uav/logs` | Topic for structured log messages. |

---

## Running locally

```bash
go run cmd/main.go config.json
```

---

## Code structure

```
mission/
├── cmd/main.go                        Entry point. Wires dependencies and starts the manager.
├── domain/models.go                   Core types: Waypoint, State, AppConfig, BrokerMessage.
├── ports/interfaces.go                Interfaces for Broker, ConfigLoader, MissionParser.
├── usecase/mission_manager.go         State machine logic. Everything algorithmic lives here.
└── infrastructure/
    ├── broker/udp_broker.go           UDP pub/sub client.
    ├── config/file_loader.go          Reads and unmarshals config.json.
    ├── parser/kml_parser.go           Extracts waypoints from a KML LineString.
    └── broker_log_writer.go           io.Writer that forwards log output to the broker.
```

The architecture follows a ports-and-adapters pattern. `MissionManager` depends only on the interfaces in `ports/`; the concrete UDP, KML, and file implementations are injected in `main.go`. Swapping any of them out for tests or alternative transports requires no changes to the use case layer.

---

## Known limitations

- **KML structure is rigid.** The parser expects a single `Document > Placemark > LineString` element. KML files exported from Google Earth usually fit this shape, but anything with multiple placemarks, folders, or point geometries will be silently ignored or fail to parse.
- **No altitude awareness when paused.** When `input_mission_delay` fires, the state is temporarily set to `PAUSED` and restored asynchronously via `time.AfterFunc`. If a `pause` command arrives during that window, the goroutine will set state back to `FLYING` regardless.
- **`FaceAlongGPSCourse` is equivalent to `FaceNextWaypoint`.** Computing a true course-over-ground heading requires more telemetry history than is currently tracked.
- **`final_altitude_for_rtl` is not enforced.** When `mission_end` is `rtl`, the algorithm simply switches to `RTL` mode. ArduPilot handles the RTL altitude internally based on its own parameters.