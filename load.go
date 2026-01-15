package pngmeta

import (
	"encoding/binary"
	"fmt"
	"hash"
	"hash/crc32"
	"io"
	"slices"
)

// Attempt to read the PNG signature bytes from `r`.
func readSignature(r io.Reader) error {
	// First, verify signature.
	signature := make([]byte, 8)
	_, err := r.Read(signature)
	if err != nil { return err }
	if !slices.Equal(signature, PNGSignature) {
		return fmt.Errorf("Not a PNG; invalid signature")
	}
	return nil
}

// Only read the IHDR chunk from the file, and nothing more.
// Can be beneficial for high-performance metadata checking.
func ReadIHDRChunk(r io.Reader) (c Chunk, err error) {
	if err = readSignature(r); err != nil {
		return Chunk{}, err
	}

	hasher := crc32.NewIEEE()
	
	c, err = ReadChunk(r, hasher)
	if err != nil { return Chunk{}, err }

	if c.Type != TypeCodeIHDR {
		return Chunk{}, fmt.Errorf("IHDR chunk expected at beginning, got %s", IntToChunkType(c.Type))
	}

	return
}

// Read all chunks from given reader.
func ReadChunks(r io.Reader) (chunks []Chunk, err error) {
	if err = readSignature(r); err != nil {
		return nil, err
	}
	
	hasher := crc32.NewIEEE()

	for {
		chunk, err := ReadChunk(r, hasher)
		if err != nil { return nil, err }
		
		chunks = append(chunks, chunk)

		// If current chunk is IEND, break.
		if chunk.Type == TypeCodeIEND {
			break
		}
	}

	return
}

// Read a single chunk from `r`, using `hasher` which must be an IEEE CRC-32 instance.
func ReadChunk(r io.Reader, hasher hash.Hash32) (c Chunk, err error) {
	// Read chunk data length.
	if err = binary.Read(r, binary.BigEndian, &c.Length); err != nil {
		return Chunk{}, err
	}

	// Read c type.
	if err = binary.Read(r, binary.BigEndian, &c.Type); err != nil {
		return Chunk{}, err
	}

	// Read c data via prior length.
	c.Data = make([]byte, c.Length)
	_, err = io.ReadFull(r, c.Data)
	if err != nil { return Chunk{}, err }

	// Read c CRC checksum.
	if err = binary.Read(r, binary.BigEndian, &c.CRC); err != nil {
		return Chunk{}, err
	}

	// Verify checksum.
	sum, err := c.Checksum(hasher)
	if err != nil { return Chunk{}, err }

	if sum != c.CRC {
		return Chunk{}, fmt.Errorf(
			"Checksum mismatch in %s c",
			c.TypeName(),
		)
	}

	return
}
