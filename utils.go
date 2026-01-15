package pngmeta

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

// See: https://exiftool.org/TagNames/PNG.html#ImageHeader
type PNGHeader struct {
	Width, Height uint32
	BitDepth, ColorType uint8
	Compression, Filter, Interlace uint8
}

func ParseIHDRChunk(chunkIHDR Chunk) (h PNGHeader) {
	h.Width = binary.BigEndian.Uint32(chunkIHDR.Data[0:4])
	h.Height = binary.BigEndian.Uint32(chunkIHDR.Data[4:8])
	h.BitDepth = chunkIHDR.Data[8]
	h.ColorType = chunkIHDR.Data[9]
	h.Compression = chunkIHDR.Data[10]
	h.Filter = chunkIHDR.Data[11]
	h.Interlace = chunkIHDR.Data[12]
	return
}
