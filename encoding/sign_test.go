
package encoding_test

import (
	"testing"
	"reflect"
	. "github.com/2gui/rpc/encoding"
)

func TestEncodeTypeSign(t *testing.T){
	type T struct{
		t reflect.Type
		except string
	}
	types := []T{
		{reflect.TypeOf((bool)(false)), "A"},
		{reflect.TypeOf((uint8)(0)), "B"},
		{reflect.TypeOf((uint16)(0)), "I"},
		{reflect.TypeOf((uint32)(0)), "L"},
		{reflect.TypeOf((uint64)(0)), "Q"},
		{reflect.TypeOf((int8)(0)), "b"},
		{reflect.TypeOf((int16)(0)), "i"},
		{reflect.TypeOf((int32)(0)), "l"},
		{reflect.TypeOf((int64)(0)), "q"},
		{reflect.TypeOf((float32)(0)), "F"},
		{reflect.TypeOf((float64)(0)), "D"},
		{reflect.TypeOf((*bool)(nil)), "*A"},
		{reflect.TypeOf((*int8)(nil)), "*b"},
		{reflect.TypeOf((*int16)(nil)), "*i"},
		{reflect.TypeOf((*int32)(nil)), "*l"},
		{reflect.TypeOf((*int64)(nil)), "*q"},
		{reflect.TypeOf((string)("")), "[B"},
		{reflect.TypeOf(([]bool)(nil)), "[A"},
		{reflect.TypeOf(([]int8)(nil)), "[b"},
		{reflect.TypeOf(([]int16)(nil)), "[i"},
		{reflect.TypeOf(([]int32)(nil)), "[l"},
		{reflect.TypeOf(([]int64)(nil)), "[q"},
		{reflect.TypeOf([0]bool{}), "[A"},
		{reflect.TypeOf([0]int8{}), "[b"},
		{reflect.TypeOf([0]int16{}), "[i"},
		{reflect.TypeOf([0]int32{}), "[l"},
		{reflect.TypeOf([0]int64{}), "[q"},
		{reflect.TypeOf((*[]byte)(nil)), "*[B"},
		{reflect.TypeOf([]*byte{}), "[*B"},
		{reflect.TypeOf([][]byte{}), "[[B"},
		{reflect.TypeOf([][]*byte{}), "[[*B"},
		{reflect.TypeOf([][][]byte{}), "[[[B"},
		{reflect.TypeOf([][0]byte{}), "[[B"},
		{reflect.TypeOf([0][]byte{}), "[[B"},
		{reflect.TypeOf([0][0]byte{}), "[[B"},
	}
	for _, d := range types {
		s := EncodeTypeSign(d.t)
		t.Logf("%-10s %s", d.t.String(), s)
		if s != d.except {
			t.Errorf("'%s' is %s, but except %s", d.t.String(), s, d.except)
		}
	}
}
