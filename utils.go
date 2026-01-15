package png

import (
	"encoding/binary"
	"math"
)

const (
	TypeCodeIHDR = 1229472850
	TypeCodeIDAT = 1229209940
	TypeCodeIEND = 1229278788
	TypeCodepHYs = 1883789683
)

// Fast conversion of type integer to equivalent string of 4 ASCII characters.
func IntToChunkType(i uint32) string {
	return string(rune(i & 0xff000000 >> 24)) +
		string(rune(i & 0x00ff0000 >> 16)) +
		string(rune(i & 0x0000ff00 >> 8)) +
		string(rune(i & 0x000000ff))
}

// Fast conversion of chunk type string to equivalent uint32.
func ChunkTypeToInt(str string) uint32 {
	return uint32(str[0]) << 24 &
		uint32(str[1] << 16) &
		uint32(str[2] << 8) &
		uint32(str[3])
}

// Returns chunk data for a pHYs (physical dimensions) chunk,
// given a value for pixels-per-meter.
func NewPhysData(ppm float64) []byte {
	data := make([]byte, 9)
	ppmQuantized := uint32(math.Round(ppm))
	binary.BigEndian.PutUint32(data[0:4], ppmQuantized)
	binary.BigEndian.PutUint32(data[4:8], ppmQuantized)
	data[8] = 1 // 1 means "meters" is the unit for pixels-per-unit.
	return data
}

type PNGHeader struct {
	Width, Height uint32
	BitDepth, ColorType uint8
	Compression, Filter, Interlace uint8
}

func ParseIHDRChunk(chunkIHDR Chunk) (PNGHeader, error) {
	var header PNGHeader

	chunkIHDR, err := p.GetChunk(TypeCodeIHDR)
	if err != nil { return PNGHeader{}, err }
}
