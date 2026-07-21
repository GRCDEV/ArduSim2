package usecase

import (
	"fmt"
	"log"
	"math"

	"mission/domain"
	"mission/ports"
	"time"
)

type MissionManager struct {
	broker        ports.Broker
	configLoader  ports.ConfigLoader
	missionParser ports.MissionParser

	config    *domain.AppConfig
	waypoints []domain.Waypoint

	originalWaypoints []domain.Waypoint
	relativeHomeSet   bool
	lastPos           domain.Waypoint // Stores last known lat, lon for bearing calculation

	state          domain.State
	currentWpIndex int
}

func NewMissionManager(b ports.Broker, c ports.ConfigLoader, p ports.MissionParser) *MissionManager {
	return &MissionManager{
		broker:        b,
		configLoader:  c,
		missionParser: p,
		state:         domain.IDLE,
	}
}

func (m *MissionManager) GetConfig() *domain.AppConfig {
	return m.config
}

func (m *MissionManager) Initialize(configFile string) error {
	var err error
	m.config, err = m.configLoader.LoadAppConfig(configFile)
	if err != nil {
		return err
	}

	m.originalWaypoints, err = m.missionParser.ParseMission(m.config.MissionFile)
	if err != nil {
		return err
	}

	m.waypoints = make([]domain.Waypoint, len(m.originalWaypoints))
	copy(m.waypoints, m.originalWaypoints)

	if len(m.waypoints) > 0 {
		m.currentWpIndex = 0
	}

	m.relativeHomeSet = !m.config.RelativeMovement

	for {
		err := m.broker.Connect(m.config.BrokerIP, m.config.BrokerPort, m.config.SubscriptionTopic, m.config.TelemetryTopic)
		if err != nil {
			log.Printf(err.Error())
			time.Sleep(5 * time.Second)
			continue
		}
		break
	}

	return nil
}

func (m *MissionManager) Run() {
	msgChan, err := m.broker.Listen()
	if err != nil {
		log.Fatalf("Failed to listen to broker: %v", err)
	}

	log.Println("Mission Manager initialized. Waiting for commands...")

	for msg := range msgChan {
		m.handleMessage(msg)
	}
}

func (m *MissionManager) handleMessage(msg domain.BrokerMessage) {
	switch msg.Topic {
	case m.config.SubscriptionTopic:
		m.handleCommand(msg.Payload)
	case m.config.TelemetryTopic:
		m.handleTelemetry(msg.Payload)
	}
}

func (m *MissionManager) handleCommand(payload map[string]interface{}) {
	cmd, ok := payload["command"].(string)
	if !ok {
		return
	}

	switch cmd {
	case "start":
		if m.state == domain.IDLE {
			m.broker.Publish(m.config.PublishTopic, map[string]interface{}{"service_id": m.config.ServiceID, "endpoint": "Arm"})
			m.broker.Publish(m.config.PublishTopic, map[string]interface{}{"service_id": m.config.ServiceID, "endpoint": "SetFlightmode", "flightmode": "GUIDED"})

			takeoffAlt := m.waypoints[1].Altitude
			if takeoffAlt == 0 {
				takeoffAlt = m.config.MinimumWaypointRelativeAltitude
			}

			m.broker.Publish(m.config.PublishTopic, map[string]interface{}{
				"service_id": m.config.ServiceID,
				"endpoint":   "Takeoff",
				"altitude":   takeoffAlt,
			})
			m.state = domain.TAKEOFF
			log.Printf("Starting mission: taking off to %.2f\n", takeoffAlt)
		}
	case "resume":
		if m.state == domain.PAUSED {
			m.state = domain.FLYING
			log.Println("Resuming mission...")
			m.broker.Publish(m.config.PublishTopic, map[string]interface{}{"service_id": m.config.ServiceID, "endpoint": "SetFlightmode", "flightmode": "GUIDED"})
		}
	case "pause":
		if m.state == domain.FLYING {
			m.state = domain.PAUSED
			m.broker.Publish(m.config.PublishTopic, map[string]interface{}{"service_id": m.config.ServiceID, "endpoint": "SetFlightmode", "flightmode": "BRAKE"})
			log.Println("Mission paused.")
		}
	case "stop":
		if m.state != domain.IDLE && m.state != domain.FINISHED {
			m.state = domain.LANDING
			m.broker.Publish(m.config.PublishTopic, map[string]interface{}{"service_id": m.config.ServiceID, "endpoint": "Land"})
			log.Println("Mission stopped. Landing...")
		}
	case "rtl":
		m.state = domain.LANDING
		m.broker.Publish(m.config.PublishTopic, map[string]interface{}{"service_id": m.config.ServiceID, "endpoint": "SetFlightmode", "flightmode": "RTL"})
		log.Println("Returning to Home (RTL)...")
	case "emergency_land":
		m.state = domain.LANDING
		m.broker.Publish(m.config.PublishTopic, map[string]interface{}{"service_id": m.config.ServiceID, "endpoint": "Land"})
		log.Println("Emergency Land...")
	}
}

func (m *MissionManager) handleTelemetry(payload map[string]interface{}) {
	if m.state == domain.IDLE {
		return
	}

	// First, check for Relative Home if needed
	if m.config.RelativeMovement && !m.relativeHomeSet {
		pos, _ := payload["position"].(map[string]interface{})
		lat, ok1 := pos["lat"].(float64)
		lon, ok2 := pos["lon"].(float64)
		nrGpsOnline, _ := payload["nr_gps_online"].(float64)

		if ok1 && ok2 && lat != 0 && lon != 0 && nrGpsOnline > 0 {
			log.Printf("Relative Home set to: %.6f, %.6f\n", lat, lon)
			m.setRelativeWaypoints(lat, lon)
			m.relativeHomeSet = true
		}
		return
	}

	pos, _ := payload["position"].(map[string]interface{})
	lat, ok1 := pos["lat"].(float64)
	lon, ok2 := pos["lon"].(float64)

	if ok1 && ok2 {
		m.lastPos = domain.Waypoint{Latitude: lat, Longitude: lon}
	}

	switch m.state {
	case domain.TAKEOFF:
		alt, _ := pos["relative_alt"].(float64)
		takeoffAlt := m.waypoints[1].Altitude
		if takeoffAlt == 0 {
			takeoffAlt = m.config.MinimumWaypointRelativeAltitude
		}
		if math.Abs(alt-takeoffAlt) < 1.0 {
			m.state = domain.FLYING
			m.currentWpIndex = 0
			m.sendNextWaypoint()
		}
	case domain.FLYING:
		lat, ok1 := pos["lat"].(float64)
		lon, ok2 := pos["lon"].(float64)

		if ok1 && ok2 && m.currentWpIndex < len(m.waypoints) {
			target := m.waypoints[m.currentWpIndex]
			dist := haversine(lat, lon, target.Latitude, target.Longitude)

			// Convert configuration distance from cm to meters
			distThreshold := m.config.DistanceToWaypointReached / 100.0

			if dist < distThreshold {
				log.Printf("Reached Waypoint %d\n", m.currentWpIndex)

				// Implement Mission Delay
				if m.config.InputMissionDelay > 0 {
					m.state = domain.PAUSED // Use PAUSED as a temporary state for delay
					time.AfterFunc(time.Duration(m.config.InputMissionDelay*float64(time.Second)), func() {
						m.state = domain.FLYING
						m.advanceWaypoint()
					})
				} else {
					m.advanceWaypoint()
				}
			}
		}
	case domain.LANDING:
		alt, ok := pos["relative_alt"].(float64)
		if ok && alt <= 0.5 {
			m.state = domain.FINISHED
			m.broker.Publish(m.config.ExternalMessagesTopic, map[string]interface{}{"command": "finish", "source": "mission"})
			log.Println("UAV landed. Mission finished.")
		}
	}
}

func (m *MissionManager) advanceWaypoint() {
	m.currentWpIndex++
	if m.currentWpIndex < len(m.waypoints) {
		m.sendNextWaypoint()
	} else {
		m.handleMissionEnd()
	}
}

func (m *MissionManager) handleMissionEnd() {
	switch m.config.MissionEnd {
	case domain.LandMode:
		m.state = domain.LANDING
		m.broker.Publish(m.config.PublishTopic, map[string]interface{}{"service_id": m.config.ServiceID, "endpoint": "Land"})
		log.Println("Mission complete. Landing...")
	case domain.RTLMode:
		m.state = domain.LANDING
		m.broker.Publish(m.config.PublishTopic, map[string]interface{}{
			"service_id": m.config.ServiceID,
			"endpoint":   "SetFlightmode",
			"flightmode": "RTL",
		})
		// If RTL altitude is set, we might want to use it, but RTL mode usually handles its own altitude
		// Some systems allow setting RTL altitude beforehand.
		log.Println("Mission complete. RTL...")
	default:
		m.state = domain.FINISHED
		m.broker.Publish(m.config.ExternalMessagesTopic, map[string]interface{}{"command": "finish", "source": "mission"})
		log.Println("Mission complete. Staying at last waypoint (unmodified).")
	}
}

func (m *MissionManager) sendNextWaypoint() {
	m.broker.Publish(m.config.ExternalMessagesTopic, map[string]any{
		"command": fmt.Sprintf("Waypoint %d reached", m.currentWpIndex+1),
		"source":  "mission",
	})
	if m.currentWpIndex < len(m.waypoints) {
		wp := m.waypoints[m.currentWpIndex]

		altitude := wp.Altitude
		if m.config.OverrideIncludedAltitudeValues {
			altitude = m.config.WaypointsRelativeAltitude
		} else if altitude < m.config.MinimumWaypointRelativeAltitude {
			altitude = m.config.MinimumWaypointRelativeAltitude
		}

		payload := map[string]interface{}{
			"service_id": m.config.ServiceID,
			"endpoint":   "MoveToPosition",
			"latitude":   wp.Latitude,
			"longitude":  wp.Longitude,
			"altitude":   altitude,
		}

		if m.config.OverrideIncludedYawValues {
			payload["yaw"] = m.calculateYaw(wp)
		}

		m.broker.Publish(m.config.PublishTopic, payload)
		log.Printf("Moving to Waypoint %d (Alt: %.2f)\n", m.currentWpIndex, altitude)
	}
}

func (m *MissionManager) calculateYaw(target domain.Waypoint) float64 {
	switch m.config.YawValue {
	case domain.FaceNextWPStrategy, domain.FaceNextWPExceptRTLStrategy:
		if m.lastPos.Latitude != 0 || m.lastPos.Longitude != 0 {
			_, bearing := getDistanceAndBearing(m.lastPos.Latitude, m.lastPos.Longitude, target.Latitude, target.Longitude)
			if bearing < 0 {
				bearing += 360
			}
			return bearing
		}
	case domain.FaceAlongGPSCourseStrategy:
		// Course over ground is hard to calculate without more points or telemetry.
		// For now, use the same as FaceNextWP or 0.
		if m.lastPos.Latitude != 0 || m.lastPos.Longitude != 0 {
			_, bearing := getDistanceAndBearing(m.lastPos.Latitude, m.lastPos.Longitude, target.Latitude, target.Longitude)
			if bearing < 0 {
				bearing += 360
			}
			return bearing
		}
	case domain.FixedStrategy:
		return -1 // Often used as "no change" in ArduPilot MAVLink
	}
	return 0
}

func (m *MissionManager) setRelativeWaypoints(homeLat, homeLon float64) {
	if len(m.originalWaypoints) == 0 {
		return
	}

	firstWp := m.originalWaypoints[0]
	for i, wp := range m.originalWaypoints {
		dist, bearing := getDistanceAndBearing(firstWp.Latitude, firstWp.Longitude, wp.Latitude, wp.Longitude)
		newLat, newLon := destinationPoint(homeLat, homeLon, dist, bearing)
		m.waypoints[i] = domain.Waypoint{
			Latitude:  newLat,
			Longitude: newLon,
			Altitude:  wp.Altitude,
		}
	}
}

func getDistanceAndBearing(lat1, lon1, lat2, lon2 float64) (float64, float64) {
	const R = 6371e3
	phi1 := lat1 * math.Pi / 180
	phi2 := lat2 * math.Pi / 180
	deltaPhi := (lat2 - lat1) * math.Pi / 180
	deltaLambda := (lon2 - lon1) * math.Pi / 180

	a := math.Sin(deltaPhi/2)*math.Sin(deltaPhi/2) +
		math.Cos(phi1)*math.Cos(phi2)*
			math.Sin(deltaLambda/2)*math.Sin(deltaLambda/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	dist := R * c

	y := math.Sin(deltaLambda) * math.Cos(phi2)
	x := math.Cos(phi1)*math.Sin(phi2) -
		math.Sin(phi1)*math.Cos(phi2)*math.Cos(deltaLambda)
	bearing := math.Atan2(y, x) * 180 / math.Pi

	return dist, bearing
}

func destinationPoint(lat, lon, distance, bearing float64) (float64, float64) {
	const R = 6371e3
	phi1 := lat * math.Pi / 180
	lambda1 := lon * math.Pi / 180
	theta := bearing * math.Pi / 180
	delta := distance / R

	phi2 := math.Asin(math.Sin(phi1)*math.Cos(delta) +
		math.Cos(phi1)*math.Sin(delta)*math.Cos(theta))
	lambda2 := lambda1 + math.Atan2(math.Sin(theta)*math.Sin(delta)*math.Cos(phi1),
		math.Cos(delta)-math.Sin(phi1)*math.Sin(phi2))

	return phi2 * 180 / math.Pi, lambda2 * 180 / math.Pi
}

func haversine(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371e3
	phi1 := lat1 * math.Pi / 180
	phi2 := lat2 * math.Pi / 180
	deltaPhi := (lat2 - lat1) * math.Pi / 180
	deltaLambda := (lon2 - lon1) * math.Pi / 180

	a := math.Sin(deltaPhi/2)*math.Sin(deltaPhi/2) +
		math.Cos(phi1)*math.Cos(phi2)*
			math.Sin(deltaLambda/2)*math.Sin(deltaLambda/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return R * c
}
