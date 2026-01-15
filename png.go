package pngmeta

import (
	"fmt"
	"hash/crc32"
	"io"
)

// A collection of chunks with common operations.
type PNG struct {
	begin []Chunk
	data []Chunk
	end []Chunk
}

func LoadPNG(r io.Reader) (*PNG, error) {
	chunks, err := ReadChunks(r)
	if err != nil { return nil, err }

	// Determine chunk section ranges based off
	// of where IDAT chunks begin and end.
	var idatStart, idatEnd = 0, len(chunks) - 1
	for i, c := range chunks {
		if c.Type == TypeCodeIDAT {
			idatStart = i
			break
		}
	}
	i := idatEnd
	for i > 0 {
		if chunks[i].Type == TypeCodeIDAT {
			idatEnd = i + 1
			break
		}
		i -= 1
	}

	begin := make([]Chunk, idatStart)
	data := make([]Chunk, idatEnd - idatStart)
	end := make([]Chunk, len(chunks) - idatEnd)
	copy(begin, chunks[:idatStart])
	copy(data, chunks[idatStart:idatEnd])
	copy(end, chunks[idatEnd:])

	p := &PNG{
		begin: begin,
		data: data,
		end: end,
	}

	return p, nil
}

func (p *PNG) Chunks() []Chunk {
	chunks := make([]Chunk, len(p.begin) + len(p.data) + len(p.end))
	dataStart := len(p.begin)
	endStart := dataStart + len(p.data)
	copy(chunks[:dataStart], p.begin)
	copy(chunks[dataStart:endStart], p.data)
	copy(chunks[endStart:], p.end)
	return chunks
}

type Region = uint8

const (
	RegionBegin Region = iota
	RegionData
	RegionEnd
)

func (p *PNG) AddChunk(chunkType uint32, data []byte, region Region) error {
	// Ensure chunk data fits within PNG spec length limit.
	if len(data) > ((1 << 31) - 1) {
		return fmt.Errorf("Chunk data exceeds length limit of 2^31-1 bytes")
	}

	// Construct new chunk.
	newChunk := Chunk{
		Type: chunkType,
		Data: data,
		Length: uint32(len(data)),
	}

	// Compute newly-constructed chunk's CRC checksum.
	hasher := crc32.NewIEEE()
	sum, err := newChunk.Checksum(hasher)
	if err != nil { return err }
	newChunk.CRC = sum

	switch region {
	case RegionBegin:
		p.begin = append(p.begin, newChunk)
	case RegionData:
		p.data = append(p.data, newChunk)
	case RegionEnd:
		// Must prepend so IEND chunk retains the last spot.
		p.end = append([]Chunk{newChunk}, p.end...)
	}

	return nil
}

func (p *PNG) RegionSlice(region Region) []Chunk {
	switch region {
	case RegionBegin:
		return p.begin
	case RegionData:
		return p.data
	case RegionEnd:
		return p.end
	}
	// Should never happen.
	return nil
}

// Replaces the data for the first chunk of the given type, or adds
// the chunk as a new chunk if one does not yet exist.
func (p *PNG) SetChunk(chunkType uint32, data []byte, region Region) error {
	regionSlice := p.RegionSlice(region)

	for i, c := range regionSlice {
		if c.Type == chunkType {
			regionSlice[i].Data = data
			return nil
		}
	}

	return p.AddChunk(chunkType, data, region)
}

// Return the first chunk matching the given type code.
func (p *PNG) GetChunk(chunkType uint32, region Region) (Chunk, error) {
	regionSlice := p.RegionSlice(region)

	for _, c := range regionSlice {
		if c.Type == chunkType {
			return c, nil
		}
	}

	return Chunk{}, fmt.Errorf("Region does not contain a %s chunk", IntToChunkType(chunkType))
}

// Return all chunks which match a given type code.
func (p *PNG) GetChunks(chunkType uint32, region Region) (matched []Chunk) {
	regionSlice := p.RegionSlice(region)

	for _, c := range regionSlice {
		if c.Type == chunkType {
			matched = append(matched, c)
		}
	}

	return
}

func writeRegion(regionSlice []Chunk, w io.Writer) error {
	for _, c := range regionSlice {
		if err := c.Write(w); err != nil {
			return err
		}
	}
	return nil
}

func (p *PNG) Write(w io.Writer) error {
	if _, err := w.Write(PNGSignature); err != nil {
		return err
	}
	if err := writeRegion(p.begin, w); err != nil {
		return err
	}
	if err := writeRegion(p.data, w); err != nil {
		return err
	}
	if err := writeRegion(p.end, w); err != nil {
		return err
	}
	return nil
}
