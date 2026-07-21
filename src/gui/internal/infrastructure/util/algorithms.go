package util

import "ui/internal/domain"

// CollectAlgorithmIDs returns the unique algorithm service IDs across all UAVs.
func CollectAlgorithmIDs(uavs []domain.UAV) []string {
	seen := make(map[string]bool)
	var ids []string
	for _, uav := range uavs {
		for _, svc := range uav.Services {
			if !seen[svc.ServiceId] {
				seen[svc.ServiceId] = true
				ids = append(ids, svc.ServiceId)
			}
		}
	}
	return ids
}
