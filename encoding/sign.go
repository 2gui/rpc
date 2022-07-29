
package encoding

import (
	"io"
	"reflect"
)

const (
	PointerS = '*'
	ArrayS = '['
	BoolS = 'A'
	Uint8S  = 'B'
	Uint16S = 'I'
	Uint32S = 'L'
	Uint64S = 'Q'
	Int8S  = 'b'
	Int16S = 'i'
	Int32S = 'l'
	Int64S = 'q'
	Float32S = 'F'
	Float64S = 'D'
	StringS = 'S'
)

var (
	refBoolT = reflect.TypeOf((bool)(false))
	refUint8T = reflect.TypeOf((uint8)(0))
	refUint16T = reflect.TypeOf((uint16)(0))
	refUint32T = reflect.TypeOf((uint32)(0))
	refUint64T = reflect.TypeOf((uint64)(0))
	refInt8T = reflect.TypeOf((int8)(0))
	refInt16T = reflect.TypeOf((int16)(0))
	refInt32T = reflect.TypeOf((int32)(0))
	refInt64T = reflect.TypeOf((int64)(0))
	refFloat32T = reflect.TypeOf((float32)(0))
	refFloat64T = reflect.TypeOf((float64)(0))
	refStringT = reflect.TypeOf((string)(""))
)

func EncodeTypeSign(typ reflect.Type)(s string){
	switch typ.Kind() {
	case reflect.Bool:
		return (string)(BoolS)
	case reflect.Uint8:
		return (string)(Uint8S)
	case reflect.Uint16:
		return (string)(Uint16S)
	case reflect.Uint32, reflect.Uint:
		return (string)(Uint32S)
	case reflect.Uint64:
		return (string)(Uint64S)
	case reflect.Int8:
		return (string)(Int8S)
	case reflect.Int16:
		return (string)(Int16S)
	case reflect.Int32, reflect.Int:
		return (string)(Int32S)
	case reflect.Int64:
		return (string)(Int64S)
	case reflect.Float32:
		return (string)(Float32S)
	case reflect.Float64:
		return (string)(Float64S)
	case reflect.Pointer:
		return (string)(PointerS) + EncodeTypeSign(typ.Elem())
	case reflect.Array, reflect.Slice:
		return (string)(ArrayS) + EncodeTypeSign(typ.Elem())
	case reflect.String:
		return (string)(StringS)
	default:
		panic("Unexpect type: " + typ.Kind().String())
	}
}

func EncodeFuncSign(fuc any)(in string, out string){
	typ := reflect.TypeOf(fuc)
	if typ.Kind() != reflect.Func {
		panic("require a funcion type, get " + typ.Kind().String())
	}
	ic, oc := typ.NumIn(), typ.NumOut()
	for i := 0; i < ic; i++ {
		in += EncodeTypeSign(typ.In(i))
	}
	for i := 0; i < oc; i++ {
		out += EncodeTypeSign(typ.Out(i))
	}
	return
}

func SplitSign(sign string)(signs []string){
	for i, j := 0, 0; j < len(sign); {
		j++
		switch sign[j - 1] {
		case PointerS, ArrayS:
		case BoolS, Uint8S, Uint16S, Uint32S, Uint64S, Int8S, Int16S, Int32S, Int64S, Float32S, Float64S:
			signs = append(signs, sign[i:j])
			i = j
		default:
			panic("Unknown sign " + (string)((rune)(sign[j])))
		}
	}
	return
}

func WriteType(w io.Writer, v reflect.Type)(n int64, err error){
	var n0 int
	n0, err = io.WriteString(w, EncodeTypeSign(v))
	n = (int64)(n0)
	return
}

func WriteValue(w io.Writer, v reflect.Value)(n int64, err error){
	switch v.Type().Kind() {
	case reflect.Bool:
		return WriteBool(w, v.Bool())
	case reflect.Uint8:
		return WriteUint8(w, (uint8)(v.Uint()))
	case reflect.Uint16:
		return WriteUint16(w, (uint16)(v.Uint()))
	case reflect.Uint32, reflect.Uint:
		return WriteUint32(w, (uint32)(v.Uint()))
	case reflect.Uint64:
		return WriteUint64(w, v.Uint())
	case reflect.Int8:
		return WriteUint8(w, (uint8)(v.Int()))
	case reflect.Int16:
		return WriteUint16(w, (uint16)(v.Int()))
	case reflect.Int32, reflect.Int:
		return WriteUint32(w, (uint32)(v.Int()))
	case reflect.Int64:
		return WriteUint64(w, (uint64)(v.Int()))
	case reflect.Float32:
		return WriteFloat32(w, (float32)(v.Float()))
	case reflect.Float64:
		return WriteFloat64(w, v.Float())
	case reflect.Pointer:
		var n0 int64
		if v.IsNil() {
			return WriteBool(w, false)
		}
		n, err = WriteBool(w, true)
		if err != nil {
			return
		}
		n0, err = WriteValue(w, v.Elem())
		n += n0
		return
	case reflect.Array, reflect.Slice:
		var n0 int64
		l := v.Len()
		n, err = WriteUint32(w, (uint32)(l))
		if err != nil {
			return
		}
		for i := 0; i < l; i++ {
			n0, err = WriteValue(w, v.Index(i))
			n += n0
			if err != nil {
				return
			}
		}
		return
	case reflect.String:
		return WriteString(w, v.String())
	default:
		panic("Unexpect type: " + v.Type().String())
	}
	return
}

func ReadType(r io.Reader)(typ reflect.Type, n int64, err error){
	var n0 int64
	var c byte
	c, n, err = ReadUint8(r)
	if err != nil {
		return
	}
	switch c {
	case PointerS:
		typ, n0, err = ReadType(r)
		n += n0
		if err != nil {
			return
		}
		typ = reflect.PointerTo(typ)
		return
	case ArrayS:
		typ, n0, err = ReadType(r)
		n += n0
		if err != nil {
			return
		}
		typ = reflect.SliceOf(typ)
		return
	case StringS:
		typ = refStringT
		return
	case BoolS:
		typ = refBoolT
		return
	case Int8S:
		typ = refInt8T
		return
	case Int16S:
		typ = refInt16T
		return
	case Int32S:
		typ = refInt32T
		return
	case Int64S:
		typ = refInt64T
		return
	case Uint8S:
		typ = refUint8T
		return
	case Uint16S:
		typ = refUint16T
		return
	case Uint32S:
		typ = refUint32T
		return
	case Uint64S:
		typ = refUint64T
		return
	case Float32S:
		typ = refFloat32T
		return
	case Float64S:
		typ = refFloat64T
		return
	default:
		panic("Unknown sign " + (string)((rune)(c)))
	}
}

func ReadValue(r io.Reader, typ reflect.Type)(val reflect.Value, n int64, err error){
	kind := typ.Kind()
	switch kind {
	case reflect.Pointer:
		var (
			v reflect.Value
			n0 int64
			ok bool
		)
		ok, n, err = ReadBool(r)
		if err != nil {
			return
		}
		if !ok {
			val = reflect.New(typ).Elem()
			return
		}
		v, n0, err = ReadValue(r, typ.Elem())
		n += n0
		if err != nil {
			return
		}
		val = reflect.New(typ.Elem())
		val.Elem().Set(v)
		return
	case reflect.Array, reflect.Slice:
		var l uint32
		var (
			v reflect.Value
			n0 int64
		)
		l, n, err = ReadUint32(r)
		if err != nil {
			return
		}
		if kind == reflect.Array{
			if typ.Len() != (int)(l) {
				panic("Array length not match")
			}
			val = reflect.New(typ).Elem()
		}else{
			val = reflect.MakeSlice(typ, (int)(l), (int)(l))
		}
		elem := typ.Elem()
		for i := 0; i < (int)(l); i++ {
			v, n0, err = ReadValue(r, elem)
			n += n0
			if err != nil {
				return
			}
			val.Index(i).Set(v)
		}
		return
	case reflect.String:
		var v0 string
		v0, n, err = ReadString(r)
		if err != nil {
			return
		}
		val = reflect.New(typ).Elem()
		val.SetString(v0)
		return
	case reflect.Bool:
		var v0 bool
		v0, n, err = ReadBool(r)
		if err != nil {
			return
		}
		val = reflect.New(typ).Elem()
		val.SetBool(v0)
		return
	case reflect.Int8, reflect.Uint8:
		var v0 uint8
		v0, n, err = ReadUint8(r)
		if err != nil {
			return
		}
		val = reflect.New(typ).Elem()
		if kind == reflect.Int8 {
			val.SetInt((int64)((int8)(v0)))
		}else{ // if kind == reflect.Uint8
			val.SetUint((uint64)(v0))
		}
		return
	case reflect.Int16, reflect.Uint16:
		var v0 uint16
		v0, n, err = ReadUint16(r)
		if err != nil {
			return
		}
		val = reflect.New(typ).Elem()
		if kind == reflect.Int16 {
			val.SetInt((int64)((int16)(v0)))
		}else{ // if kind == reflect.Uint8
			val.SetUint((uint64)(v0))
		}
		return
	case reflect.Int32, reflect.Uint32, reflect.Int, reflect.Uint:
		var v0 uint32
		v0, n, err = ReadUint32(r)
		if err != nil {
			return
		}
		val = reflect.New(typ).Elem()
		if kind == reflect.Int32 || kind == reflect.Int {
			val.SetInt((int64)((int32)(v0)))
		}else{ // if kind == reflect.Uint8
			val.SetUint((uint64)(v0))
		}
		return
	case reflect.Int64, reflect.Uint64:
		var v0 uint64
		v0, n, err = ReadUint64(r)
		if err != nil {
			return
		}
		val = reflect.New(typ).Elem()
		if kind == reflect.Int64 {
			val.SetInt((int64)(v0))
		}else{ // if kind == reflect.Uint8
			val.SetUint(v0)
		}
		return
	case reflect.Float32:
		var v0 float32
		v0, n, err = ReadFloat32(r)
		if err != nil {
			return
		}
		val = reflect.New(typ).Elem()
		val.SetFloat((float64)(v0))
		return
	case reflect.Float64:
		var v0 float64
		v0, n, err = ReadFloat64(r)
		if err != nil {
			return
		}
		val = reflect.New(typ).Elem()
		val.SetFloat(v0)
		return
	default:
		panic("Unexpect type: " + typ.String())
	}
}
