package usecases

import "context"

type activeSession struct {
	ctx                    context.Context
	composePath            string
	kubernetesManifestPath string
	dockerHubUser          string
	simulationName         string
	uavIDs                 []string
	algorithmIDs           map[string]bool
	stoppedIDs             map[string]bool
	loggingEnabled         bool
	loggerIP               string
	gatewayIP              string
	kubeConfigPath         string
}

func newActiveSession() *activeSession {
	return &activeSession{
		algorithmIDs: make(map[string]bool),
		stoppedIDs:   make(map[string]bool),
	}
}

func (s *activeSession) isRunning() bool {
	return s.composePath != "" || s.kubernetesManifestPath != ""
}

func (s *activeSession) isLocal() bool {
	return s.kubernetesManifestPath == ""
}

func (s *activeSession) loggerHost() string {
	if s.isLocal() {
		return "localhost"
	}
	return s.loggerIP
}

func (s *activeSession) markStopped(id string) (allStopped bool) {
	s.stoppedIDs[id] = true
	if len(s.algorithmIDs) == 0 {
		return false
	}
	allStopped = true
	for algoId := range s.algorithmIDs {
		if !s.stoppedIDs[algoId] {
			allStopped = false
			break
		}
	}
	return allStopped
}

func (s *activeSession) clear() {
	s.composePath = ""
	s.kubernetesManifestPath = ""
	s.dockerHubUser = ""
	s.simulationName = ""
	s.uavIDs = nil
	s.algorithmIDs = make(map[string]bool)
	s.stoppedIDs = make(map[string]bool)
	s.loggingEnabled = false
	s.loggerIP = ""
	s.gatewayIP = ""
	s.kubeConfigPath = ""
}
