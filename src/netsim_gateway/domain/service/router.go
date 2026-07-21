package service

import (
	"hash/fnv"
)

func ConsistentHash(uavID string, n int) int {
	if n == 0 {
		return 0
	}
	h := fnv.New32a()
	h.Write([]byte(uavID))
	return int(h.Sum32() % uint32(n))
}
