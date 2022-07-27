
package encoding

import (
	"io"
)

func WriteBytes(w io.Writer, v []byte)(n int64, err error){
	var n0 int
	n, err = WriteUint32(w, (uint32)(len(v)))
	if err != nil {
		return
	}
	n0, err = w.Write(v)
	n += (int64)(n0)
	return
}

func WriteString(w io.Writer, v string)(n int64, err error){
	return WriteBytes(w, ([]byte)(v))
}

func ReadBytes(r io.Reader)(v []byte, n int64, err error){
	var n0 int
	var l uint32
	l, n, err = ReadUint32(r)
	if err != nil {
		return
	}
	v = make([]byte, l)
	n0, err = io.ReadFull(r, v)
	n += (int64)(n0)
	return
}

func ReadString(r io.Reader)(v string, n int64, err error){
	var v0 []byte
	v0, n, err = ReadBytes(r)
	if err != nil {
		return
	}
	v = (string)(v0)
	return
}

func EncodeUint16s(buf []byte, v []uint16){
	for i, n := range v {
		EncodeUint16(buf[i * 2:], n)
	}
	return
}

func EncodeUint32s(buf []byte, v []uint32){
	for i, n := range v {
		EncodeUint32(buf[i * 4:], n)
	}
	return
}

func EncodeUint64s(buf []byte, v []uint64){
	for i, n := range v {
		EncodeUint64(buf[i * 8:], n)
	}
	return
}

func WriteUint16s(w io.Writer, v []uint16)(n int64, err error){
	var n0 int
	n, err = WriteUint32(w, (uint32)(len(v)))
	if err != nil {
		return
	}
	buf := make([]byte, len(v) * 2)
	EncodeUint16s(buf, v)
	n0, err = w.Write(buf)
	n += (int64)(n0)
	return
}

func WriteUint32s(w io.Writer, v []uint32)(n int64, err error){
	var n0 int
	n, err = WriteUint32(w, (uint32)(len(v)))
	if err != nil {
		return
	}
	buf := make([]byte, len(v) * 4)
	EncodeUint32s(buf, v)
	n0, err = w.Write(buf)
	n += (int64)(n0)
	return
}

func WriteUint64s(w io.Writer, v []uint64)(n int64, err error){
	var n0 int
	n, err = WriteUint32(w, (uint32)(len(v)))
	if err != nil {
		return
	}
	buf := make([]byte, len(v) * 8)
	EncodeUint64s(buf, v)
	n0, err = w.Write(buf)
	n += (int64)(n0)
	return
}
