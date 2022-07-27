
package encoding

import (
	"io"
	"math"
)

func EncodeUint16(buf []byte, v uint16)([]byte){
	buf[0] = (byte)(v & 0xff)
	buf[1] = (byte)(v >> 8 & 0xff)
	return buf
}

func EncodeUint32(buf []byte, v uint32)([]byte){
	buf[0] = (byte)(v & 0xff)
	buf[1] = (byte)(v >> 8 & 0xff)
	buf[2] = (byte)(v >> 16 & 0xff)
	buf[3] = (byte)(v >> 24 & 0xff)
	return buf
}

func EncodeUint64(buf []byte, v uint64)([]byte){
	buf[0] = (byte)(v & 0xff)
	buf[1] = (byte)(v >> 8 & 0xff)
	buf[2] = (byte)(v >> 16 & 0xff)
	buf[3] = (byte)(v >> 24 & 0xff)
	buf[4] = (byte)(v >> 32 & 0xff)
	buf[5] = (byte)(v >> 40 & 0xff)
	buf[6] = (byte)(v >> 48 & 0xff)
	buf[7] = (byte)(v >> 56 & 0xff)
	return buf
}

func EncodeFloat32(buf []byte, v float32)([]byte){
	return EncodeUint32(buf, math.Float32bits(v))
}

func EncodeFloat64(buf []byte, v float64)([]byte){
	return EncodeUint64(buf, math.Float64bits(v))
}

func DecodeUint16(buf []byte)(uint16){
	return (uint16)(buf[0]) | (uint16)(buf[1]) << 8
}

func DecodeUint32(buf []byte)(uint32){
	return (uint32)(buf[0]) | (uint32)(buf[1]) << 8 | (uint32)(buf[2]) << 16 | (uint32)(buf[3]) << 24
}

func DecodeUint64(buf []byte)(uint64){
	return (uint64)(buf[0]) | (uint64)(buf[1]) << 8 | (uint64)(buf[2]) << 16 | (uint64)(buf[3]) << 24 |
		(uint64)(buf[4]) << 32 | (uint64)(buf[5]) << 40 | (uint64)(buf[6]) << 48 | (uint64)(buf[7]) << 56
}

func DecodeFloat32(buf []byte)(float32){
	return math.Float32frombits(DecodeUint32(buf))
}

func DecodeFloat64(buf []byte)(float64){
	return math.Float64frombits(DecodeUint64(buf))
}

func WriteBool(w io.Writer, v bool)(n int64, err error){
	var buf [1]byte
	if v {
		buf[0] = 1
	}
	var n0 int
	n0, err = w.Write(buf[:])
	n = (int64)(n0)
	return
}

func WriteUint8(w io.Writer, v uint8)(n int64, err error){
	var n0 int
	n0, err = w.Write([]byte{v})
	n = (int64)(n0)
	return
}

func WriteUint16(w io.Writer, v uint16)(n int64, err error){
	var buf [2]byte
	var n0 int
	n0, err = w.Write(EncodeUint16(buf[:], v))
	n = (int64)(n0)
	return
}

func WriteUint32(w io.Writer, v uint32)(n int64, err error){
	var buf [4]byte
	var n0 int
	n0, err = w.Write(EncodeUint32(buf[:], v))
	n = (int64)(n0)
	return
}

func WriteUint64(w io.Writer, v uint64)(n int64, err error){
	var buf [8]byte
	var n0 int
	n0, err = w.Write(EncodeUint64(buf[:], v))
	n = (int64)(n0)
	return
}

func WriteFloat32(w io.Writer, v float32)(n int64, err error){
	return WriteUint32(w, math.Float32bits(v))
}

func WriteFloat64(w io.Writer, v float64)(n int64, err error){
	return WriteUint64(w, math.Float64bits(v))
}

func ReadBool(r io.Reader)(v bool, n int64, err error){
	var buf [1]byte
	var n0 int
	n0, err = io.ReadFull(r, buf[:])
	n = (int64)(n0)
	if err != nil {
		return
	}
	v = buf[0] != 0
	return
}

func ReadUint8(r io.Reader)(v uint8, n int64, err error){
	var buf [1]byte
	var n0 int
	n0, err = io.ReadFull(r, buf[:])
	n = (int64)(n0)
	if err != nil {
		return
	}
	v = buf[0]
	return
}

func ReadUint16(r io.Reader)(v uint16, n int64, err error){
	var buf [2]byte
	var n0 int
	n0, err = io.ReadFull(r, buf[:])
	n = (int64)(n0)
	if err != nil {
		return
	}
	v = DecodeUint16(buf[:])
	return
}

func ReadUint32(r io.Reader)(v uint32, n int64, err error){
	var buf [4]byte
	var n0 int
	n0, err = io.ReadFull(r, buf[:])
	n = (int64)(n0)
	if err != nil {
		return
	}
	v = DecodeUint32(buf[:])
	return
}

func ReadUint64(r io.Reader)(v uint64, n int64, err error){
	var buf [8]byte
	var n0 int
	n0, err = io.ReadFull(r, buf[:])
	n = (int64)(n0)
	if err != nil {
		return
	}
	v = DecodeUint64(buf[:])
	return
}

func ReadFloat32(r io.Reader)(v float32, n int64, err error){
	var v0 uint32
	v0, n, err = ReadUint32(r)
	if err != nil {
		return
	}
	v = math.Float32frombits(v0)
	return
}

func ReadFloat64(r io.Reader)(v float64, n int64, err error){
	var v0 uint64
	v0, n, err = ReadUint64(r)
	if err != nil {
		return
	}
	v = math.Float64frombits(v0)
	return
}
