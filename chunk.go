package pngmeta

import (
	"encoding/binary"
	"fmt"
	"hash"
	"io"
)

var PNGSignature = []byte{137, 80, 78, 71, 13, 10, 26, 10}

type Chunk struct {
	Length, Type, CRC uint32
	Data []byte
}

func (c Chunk) String() string {
	return fmt.Sprintf(
		"Chunk(type: %s, length: %d)",
		c.TypeName(),
		c.Length,
	)
}

func (c Chunk) TypeName() string {
	return IntToChunkType(c.Type)
}

// Return values of chunk type's property bits.
func (c Chunk) Properties() (anc, priv, res, copy bool) {
	anc = c.Type & 0x8000000 == 1
	priv = c.Type & 0x80000 == 1
	res = c.Type & 0x800 == 1
	copy = c.Type & 0x8 == 1
	return
}

func (c Chunk) Checksum(hasher hash.Hash32) (uint32, error) {
	hasher.Reset()
	
	// Digest the chunk type field first.
	typeBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(typeBytes, c.Type)
	_, err := hasher.Write(typeBytes)
	if err != nil { return 0, err }

	// Digest all data bytes.
	n, err := hasher.Write(c.Data)
	if err != nil { return 0, err }
	if n != int(c.Length) {
		return 0, fmt.Errorf("Read %d bytes of chunk data but expected %d", n, c.Length)
	}

	return hasher.Sum32(), nil
}

func (c Chunk) Write(w io.Writer) error {
	u32Slice := make([]byte, 4)

	// Write length.
	binary.BigEndian.PutUint32(u32Slice, c.Length)
	if _, err := w.Write(u32Slice); err != nil {
		return err
	}

	// Write type code.
	binary.BigEndian.PutUint32(u32Slice, c.Type)
	if _, err := w.Write(u32Slice); err != nil {
		return err
	}

	// Write data.
	if _, err := w.Write(c.Data); err != nil {
		return err
	}

	// Write checksum.
	binary.BigEndian.PutUint32(u32Slice, c.CRC)
	if _, err := w.Write(u32Slice); err != nil {
		return err
	}

	return nil
}
