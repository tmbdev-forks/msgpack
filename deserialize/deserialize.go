package deserialize

import (
	"encoding/binary"
	"fmt"
	"reflect"

	"github.com/shamaton/msgpack/def"
)

type deserializer struct {
	data    []byte
	asArray bool
}

func Exec(data []byte, holder interface{}, asArray bool) error {
	d := deserializer{data: data, asArray: asArray}

	rv := reflect.ValueOf(holder)
	if rv.Kind() != reflect.Ptr {
		return fmt.Errorf("holder must set pointer value. but got: %t", holder)
	}

	rv = rv.Elem()
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}

	_, err := d.deserialize(rv, 0)
	return err
}

func (d *deserializer) deserialize(rv reflect.Value, offset int) (int, error) {
	var err error

	// TODO : offset use uint
	switch rv.Kind() {
	case reflect.Uint8:
		v, o, err := d.asUint8(rv, offset)
		if err != nil {
			return 0, err
		}
		rv.SetUint(uint64(v))
		offset = o

	case reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint:
		offset, err = d.readAsUint(rv, offset)
		if err != nil {
			return 0, err
		}

	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
	}
	return offset, nil
}

func (d *deserializer) asUint8(rv reflect.Value, offset int) (uint8, int, error) {
	code := d.data[offset]

	if d.isFixNum(code) {
		b, offset := d.readSize1(offset)
		return uint8(b), offset, nil
	} else if code == def.Uint8 || code == def.Int8 {
		offset++
		b, offset := d.readSize1(offset)
		return uint8(b), offset, nil
	} else if code == def.Uint16 || code == def.Int16 {
		offset++
		bs, offset := d.readSize2(offset)
		return uint8(bs[1]), offset, nil
	} else if code == def.Uint32 || code == def.Int32 || code == def.Float32 {
		offset++
		bs, offset := d.readSize4(offset)
		return uint8(bs[3]), offset, nil
	} else if code == def.Uint64 || code == def.Int64 || code == def.Float64 {
		offset++
		bs, offset := d.readSize8(offset)
		return uint8(bs[7]), offset, nil
	}
	return 0, 0, fmt.Errorf("mismatch code : %x", code)
}

func (d *deserializer) readAsUint(rv reflect.Value, offset int) (int, error) {
	code := d.data[offset]

	a := byte(0xff)
	b := int8(a)
	c := uint32(b)
	fmt.Println("c : ", c)

	if d.isFixNum(code) {
		b, o := d.readSize1(offset)
		rv.SetUint(uint64(b))
		offset = o
	} else if code == def.Uint8 {
		offset++
		b, o := d.readSize1(offset)
		rv.SetUint(uint64(b))
		offset = o
	} else if code == def.Uint16 {
		offset++
		b, o := d.readSize2(offset)
		rv.SetUint(uint64(binary.BigEndian.Uint16(b)))
		offset = o
	} else if code == def.Uint32 {
		offset++
		b, o := d.readSize4(offset)
		rv.SetUint(uint64(binary.BigEndian.Uint32(b)))
		offset = o
	} else if code == def.Uint64 {
		offset++
		b, o := d.readSize8(offset)
		rv.SetUint(binary.BigEndian.Uint64(b))
		offset = o
	}
	return offset, nil
}

func (d *deserializer) isFixNum(v byte) bool {
	if def.PositiveFixIntMin <= v && v <= def.PositiveFixIntMax {
		return true
	} else if def.NegativeFixintMin <= int8(v) && int8(v) <= def.NegativeFixintMax {
		return true
	}
	return false
}
