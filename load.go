package png

import (
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"io"
	"slices"
)

func LoadChunks(r io.Reader) (chunks []Chunk, err error) {
	// First, verify signature.
	signature := make([]byte, 8)
	_, err = r.Read(signature)
	if err != nil { return nil, err }
	if !slices.Equal(signature, PNGSignature) {
		return nil, fmt.Errorf("Not a PNG; invalid signature")
	}

	hasher := crc32.NewIEEE()

	// Read chunks.
	var chunk Chunk
	for {
		// Read chunk data length.
		if err = binary.Read(r, binary.BigEndian, &chunk.Length); err != nil {
			break
		}

		// Read chunk type.
		if err = binary.Read(r, binary.BigEndian, &chunk.Type); err != nil {
			break
		}

		// Read chunk data via prior length.
		chunk.Data = make([]byte, chunk.Length)
		_, err := io.ReadFull(r, chunk.Data)
		if err != nil { return nil, err }

		// Read chunk CRC checksum.
		if err = binary.Read(r, binary.BigEndian, &chunk.CRC); err != nil {
			break
		}

		// Verify checksum.
		sum, err := chunk.Checksum(hasher)
		if err != nil { return nil, err }

		if sum != chunk.CRC {
			return nil, fmt.Errorf(
				"Checksum mismatch in %s chunk",
				chunk.TypeName(),
			)
		}
		
		chunks = append(chunks, chunk)

		// If current chunk is IEND, break.
		if chunk.Type == TypeCodeIEND {
			break
		}
	}

	return
}
