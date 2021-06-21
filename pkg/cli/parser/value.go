package parser

import (
	"net"
	"time"

	hstr "github.com/pingcap/ticat/pkg/cli/parser/dealstring"
)

func NewValueType(typeStr string) Value {
	switch typeStr {
	case "bool":
		return new(boolValue)
	case "int":
		return new(intValue)
	case "uint":
		return new(uintValue)
	case "int64":
		return new(int64Value)
	case "int32":
		return new(int32Value)
	case "int16":
		return new(int16Value)
	case "int8":
		return new(int8Value)
	case "uint64":
		return new(uint64Value)
	case "uint32":
		return new(uint32Value)
	case "uint16":
		return new(uint16Value)
	case "uint8":
		return new(uint8Value)
	case "float64":
		return new(float64Value)
	case "float32":
		return new(float32Value)
	case "duration":
		return new(durationValue)
	case "time":
		return new(timeValue)
	case "ip":
		return new(ipValue)
	case "string":
		return new(stringValue)
	case "[]bool":
		return new(boolSliceValue)
	case "[]int":
		return new(intSliceValue)
	case "[]uint":
		return new(uintSliceValue)
	case "[]int64":
		return new(int64SliceValue)
	case "[]int32":
		return new(int32SliceValue)
	case "[]int16":
		return new(int16SliceValue)
	case "[]int8":
		return new(int8SliceValue)
	case "[]uint64":
		return new(uint64SliceValue)
	case "[]uint32":
		return new(uint32SliceValue)
	case "[]uint16":
		return new(uint16SliceValue)
	case "[]uint8":
		return new(uint8SliceValue)
	case "[]float64":
		return new(float64SliceValue)
	case "[]float32":
		return new(float32SliceValue)
	case "[]duration":
		return new(durationSliceValue)
	case "[]time":
		return new(timeSliceValue)
	case "[]ip":
		return new(ipSliceValue)
	case "[]string":
		return new(stringSliceValue)
	default:
		return nil
	}
}

type Value interface {
	Set(string) error
	String() string
}

type boolValue bool
type intValue int
type uintValue uint
type int64Value int64
type int32Value int32
type int16Value int16
type int8Value int8
type uint64Value uint64
type uint32Value uint32
type uint16Value uint16
type uint8Value uint8
type float64Value float64
type float32Value float32
type durationValue time.Duration
type timeValue time.Time
type ipValue net.IP
type stringValue string
type boolSliceValue []bool
type intSliceValue []int
type uintSliceValue []uint
type int64SliceValue []int64
type int32SliceValue []int32
type int16SliceValue []int16
type int8SliceValue []int8
type uint64SliceValue []uint64
type uint32SliceValue []uint32
type uint16SliceValue []uint16
type uint8SliceValue []uint8
type float64SliceValue []float64
type float32SliceValue []float32
type durationSliceValue []time.Duration
type timeSliceValue []time.Time
type ipSliceValue []net.IP
type stringSliceValue []string

func (v *boolValue) Set(str string) error {
	val, err := hstr.ToBool(str)
	if err != nil {
		return err
	}
	*v = boolValue(val)
	return nil
}

func (v *intValue) Set(str string) error {
	val, err := hstr.ToInt(str)
	if err != nil {
		return err
	}
	*v = intValue(val)
	return nil
}

func (v *uintValue) Set(str string) error {
	val, err := hstr.ToUint(str)
	if err != nil {
		return err
	}
	*v = uintValue(val)
	return nil
}

func (v *int64Value) Set(str string) error {
	val, err := hstr.ToInt64(str)
	if err != nil {
		return err
	}
	*v = int64Value(val)
	return nil
}

func (v *int32Value) Set(str string) error {
	val, err := hstr.ToInt32(str)
	if err != nil {
		return err
	}
	*v = int32Value(val)
	return nil
}

func (v *int16Value) Set(str string) error {
	val, err := hstr.ToInt16(str)
	if err != nil {
		return err
	}
	*v = int16Value(val)
	return nil
}

func (v *int8Value) Set(str string) error {
	val, err := hstr.ToInt8(str)
	if err != nil {
		return err
	}
	*v = int8Value(val)
	return nil
}

func (v *uint64Value) Set(str string) error {
	val, err := hstr.ToUint64(str)
	if err != nil {
		return err
	}
	*v = uint64Value(val)
	return nil
}

func (v *uint32Value) Set(str string) error {
	val, err := hstr.ToUint32(str)
	if err != nil {
		return err
	}
	*v = uint32Value(val)
	return nil
}

func (v *uint16Value) Set(str string) error {
	val, err := hstr.ToUint16(str)
	if err != nil {
		return err
	}
	*v = uint16Value(val)
	return nil
}

func (v *uint8Value) Set(str string) error {
	val, err := hstr.ToUint8(str)
	if err != nil {
		return err
	}
	*v = uint8Value(val)
	return nil
}

func (v *float64Value) Set(str string) error {
	val, err := hstr.ToFloat64(str)
	if err != nil {
		return err
	}
	*v = float64Value(val)
	return nil
}

func (v *float32Value) Set(str string) error {
	val, err := hstr.ToFloat32(str)
	if err != nil {
		return err
	}
	*v = float32Value(val)
	return nil
}

func (v *durationValue) Set(str string) error {
	val, err := hstr.ToDuration(str)
	if err != nil {
		return err
	}
	*v = durationValue(val)
	return nil
}

func (v *timeValue) Set(str string) error {
	val, err := hstr.ToTime(str)
	if err != nil {
		return err
	}
	*v = timeValue(val)
	return nil
}

func (v *ipValue) Set(str string) error {
	val, err := hstr.ToIP(str)
	if err != nil {
		return err
	}
	*v = ipValue(val)
	return nil
}

func (v *stringValue) Set(str string) error {
	*v = stringValue(str)
	return nil
}

func (v *boolSliceValue) Set(str string) error {
	val, err := hstr.ToBoolSlice(str)
	if err != nil {
		return err
	}
	*v = boolSliceValue(val)
	return nil
}

func (v *intSliceValue) Set(str string) error {
	val, err := hstr.ToIntSlice(str)
	if err != nil {
		return err
	}
	*v = intSliceValue(val)
	return nil
}

func (v *uintSliceValue) Set(str string) error {
	val, err := hstr.ToUintSlice(str)
	if err != nil {
		return err
	}
	*v = uintSliceValue(val)
	return nil
}

func (v *int64SliceValue) Set(str string) error {
	val, err := hstr.ToInt64Slice(str)
	if err != nil {
		return err
	}
	*v = int64SliceValue(val)
	return nil
}

func (v *int32SliceValue) Set(str string) error {
	val, err := hstr.ToInt32Slice(str)
	if err != nil {
		return err
	}
	*v = int32SliceValue(val)
	return nil
}

func (v *int16SliceValue) Set(str string) error {
	val, err := hstr.ToInt16Slice(str)
	if err != nil {
		return err
	}
	*v = int16SliceValue(val)
	return nil
}

func (v *int8SliceValue) Set(str string) error {
	val, err := hstr.ToInt8Slice(str)
	if err != nil {
		return err
	}
	*v = int8SliceValue(val)
	return nil
}

func (v *uint64SliceValue) Set(str string) error {
	val, err := hstr.ToUint64Slice(str)
	if err != nil {
		return err
	}
	*v = uint64SliceValue(val)
	return nil
}

func (v *uint32SliceValue) Set(str string) error {
	val, err := hstr.ToUint32Slice(str)
	if err != nil {
		return err
	}
	*v = uint32SliceValue(val)
	return nil
}

func (v *uint16SliceValue) Set(str string) error {
	val, err := hstr.ToUint16Slice(str)
	if err != nil {
		return err
	}
	*v = uint16SliceValue(val)
	return nil
}

func (v *uint8SliceValue) Set(str string) error {
	val, err := hstr.ToUint8Slice(str)
	if err != nil {
		return err
	}
	*v = uint8SliceValue(val)
	return nil
}

func (v *float64SliceValue) Set(str string) error {
	val, err := hstr.ToFloat64Slice(str)
	if err != nil {
		return err
	}
	*v = float64SliceValue(val)
	return nil
}

func (v *float32SliceValue) Set(str string) error {
	val, err := hstr.ToFloat32Slice(str)
	if err != nil {
		return err
	}
	*v = float32SliceValue(val)
	return nil
}

func (v *durationSliceValue) Set(str string) error {
	val, err := hstr.ToDurationSlice(str)
	if err != nil {
		return err
	}
	*v = durationSliceValue(val)
	return nil
}

func (v *timeSliceValue) Set(str string) error {
	val, err := hstr.ToTimeSlice(str)
	if err != nil {
		return err
	}
	*v = timeSliceValue(val)
	return nil
}

func (v *ipSliceValue) Set(str string) error {
	val, err := hstr.ToIPSlice(str)
	if err != nil {
		return err
	}
	*v = ipSliceValue(val)
	return nil
}

func (v *stringSliceValue) Set(str string) error {
	val, err := hstr.ToStringSlice(str)
	if err != nil {
		return err
	}
	*v = stringSliceValue(val)
	return nil
}

func (v boolValue) String() string {
	return hstr.BoolTo(bool(v))
}

func (v intValue) String() string {
	return hstr.IntTo(int(v))
}

func (v uintValue) String() string {
	return hstr.UintTo(uint(v))
}

func (v int64Value) String() string {
	return hstr.Int64To(int64(v))
}

func (v int32Value) String() string {
	return hstr.Int32To(int32(v))
}

func (v int16Value) String() string {
	return hstr.Int16To(int16(v))
}

func (v int8Value) String() string {
	return hstr.Int8To(int8(v))
}

func (v uint64Value) String() string {
	return hstr.Uint64To(uint64(v))
}

func (v uint32Value) String() string {
	return hstr.Uint32To(uint32(v))
}

func (v uint16Value) String() string {
	return hstr.Uint16To(uint16(v))
}

func (v uint8Value) String() string {
	return hstr.Uint8To(uint8(v))
}

func (v float64Value) String() string {
	return hstr.Float64To(float64(v))
}

func (v float32Value) String() string {
	return hstr.Float32To(float32(v))
}

func (v durationValue) String() string {
	return hstr.DurationTo(time.Duration(v))
}

func (v timeValue) String() string {
	return hstr.TimeTo(time.Time(v))
}

func (v ipValue) String() string {
	return hstr.IPTo(net.IP(v))
}

func (v stringValue) String() string {
	return string(v)
}

func (v boolSliceValue) String() string {
	return hstr.BoolSliceTo([]bool(v))
}

func (v intSliceValue) String() string {
	return hstr.IntSliceTo([]int(v))
}

func (v uintSliceValue) String() string {
	return hstr.UintSliceTo([]uint(v))
}

func (v int64SliceValue) String() string {
	return hstr.Int64SliceTo([]int64(v))
}

func (v int32SliceValue) String() string {
	return hstr.Int32SliceTo([]int32(v))
}

func (v int16SliceValue) String() string {
	return hstr.Int16SliceTo([]int16(v))
}

func (v int8SliceValue) String() string {
	return hstr.Int8SliceTo([]int8(v))
}

func (v uint64SliceValue) String() string {
	return hstr.Uint64SliceTo([]uint64(v))
}

func (v uint32SliceValue) String() string {
	return hstr.Uint32SliceTo([]uint32(v))
}

func (v uint16SliceValue) String() string {
	return hstr.Uint16SliceTo([]uint16(v))
}

func (v uint8SliceValue) String() string {
	return hstr.Uint8SliceTo([]uint8(v))
}

func (v float64SliceValue) String() string {
	return hstr.Float64SliceTo([]float64(v))
}

func (v float32SliceValue) String() string {
	return hstr.Float32SliceTo([]float32(v))
}

func (v durationSliceValue) String() string {
	return hstr.DurationSliceTo([]time.Duration(v))
}

func (v timeSliceValue) String() string {
	return hstr.TimeSliceTo([]time.Time(v))
}

func (v ipSliceValue) String() string {
	return hstr.IPSliceTo([]net.IP(v))
}

func (v stringSliceValue) String() string {
	return hstr.StringSliceTo([]string(v))
}
