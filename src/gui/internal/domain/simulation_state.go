package domain

type SimulationState struct {
	Swarms        []Swarm       `json:"swarms"`
	GeneralConfig GeneralConfig `json:"generalConfig"`
	ActiveMode    string        `json:"activeMode"`
}
