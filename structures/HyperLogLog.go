package structures

import (
	"hash/fnv"
	"math"
	"math/bits"
)

func firstKbits(value, k uint64) uint64 {
	return value >> (64 - k)
}

func trailingZeroBits(value uint64) int {
	return bits.TrailingZeros64(value)
}

type HLL struct {
	m   uint64
	p   uint8
	reg []uint8
}

func NewHyperLogLog(p uint8) *HLL {
	m := uint64(1) << p
	return &HLL{
		m:   m,
		p:   p,
		reg: make([]uint8, m),
	}
}

func (h *HLL) Add(item string) {
	hash := h.hash(item)
	p := uint64(h.p)
	index := firstKbits(hash, p)
	value := uint8(trailingZeroBits(hash) + 1)

	if value > h.reg[index] {
		h.reg[index] = value
	}
}

func (h *HLL) hash(item string) uint64 {
	hasher := fnv.New64a()
	hasher.Write([]byte(item))
	return hasher.Sum64()
}

func (hll *HLL) Estimate() float64 {
	sum := 0.0
	for _, val := range hll.reg {
		sum += math.Pow(math.Pow(2.0, float64(val)), -1)
	}

	alpha := 0.7213 / (1.0 + 1.079/float64(hll.m))
	estimation := alpha * math.Pow(float64(hll.m), 2.0) / sum
	emptyRegs := hll.emptyCount()
	if estimation <= 2.5*float64(hll.m) { // do small range correction
		if emptyRegs > 0 {
			estimation = float64(hll.m) * math.Log(float64(hll.m)/float64(emptyRegs))
		}
	} else if estimation > 1/30.0*math.Pow(2.0, 32.0) { // do large range correction
		estimation = -math.Pow(2.0, 32.0) * math.Log(1.0-estimation/math.Pow(2.0, 32.0))
	}
	return estimation
}

func (hll *HLL) emptyCount() int {
	sum := 0
	for _, val := range hll.reg {
		if val == 0 {
			sum++
		}
	}
	return sum
}
