package conversions

import (
	"encoding/binary"
)

// Float32FromBytes convert bytes to float 32 value
func Float32FromBytes(bytes []byte) float32 {
	bits := binary.BigEndian.Uint16(bytes)
	return float32(bits)
}
