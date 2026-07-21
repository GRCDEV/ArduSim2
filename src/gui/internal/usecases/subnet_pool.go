package usecases

import "fmt"

type subnetPool struct {
	nextBlock uint32
}

func newSubnetPool() *subnetPool {
	// 10.10.0.0 en uint32
	return &subnetPool{nextBlock: (10<<24 | 10<<16)}
}

func (p *subnetPool) Next(nContainers int) string {
	prefix := prefixForContainers(nContainers)
	blockSize := uint32(1) << (32 - prefix)

	if p.nextBlock%blockSize != 0 {
		p.nextBlock = (p.nextBlock/blockSize + 1) * blockSize
	}

	a := p.nextBlock
	p.nextBlock += blockSize

	return fmt.Sprintf("%d.%d.%d.%d/%d",
		(a>>24)&0xFF, (a>>16)&0xFF, (a>>8)&0xFF, a&0xFF, prefix)
}

func prefixForContainers(n int) int {
	needed := n + 3
	prefix := 32
	for (1 << (32 - prefix)) < needed {
		prefix--
	}
	if prefix > 28 {
		prefix = 28
	}
	return prefix
}
