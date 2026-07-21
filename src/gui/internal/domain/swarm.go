package domain

type Swarm struct {
	ID                  string  `json:"id"`
	UAVs                []UAV   `json:"uavs"`
	GroundFormation     string  `json:"groundFormation"` // LINEAR, MATRIX, CIRCLE, RANDOM
	FormationCenterLat  float64 `json:"formationCenterLat"`
	FormationCenterLon  float64 `json:"formationCenterLon"`
	FormationSpacing    float64 `json:"formationSpacing"`
	FormationCenterMode string  `json:"formationCenterMode"`
}
