package parser

import (
	"net"
	"time"

	hstr "github.com/pingcap/ticat/pkg/cli/parser/dealstring"
)

func (f *FlagSet) GetBool(name string) (v bool) {
	flag := f.Lookup(name)
	if flag == nil {
		return
	}
	if flag.Type == "string" {
		val, err := hstr.ToBool(flag.Value.String())
		if err != nil {
			return
		}
		return val
	}
	if flag.Type != "bool" {
		return
	}
	return bool(*flag.Value.(*boolValue))
}

func (f *FlagSet) GetInt(name string) (v int) {
	flag := f.Lookup(name)
	if flag == nil {
		return
	}
	if flag.Type == "string" {
		val, err := hstr.ToInt(flag.Value.String())
		if err != nil {
			return
		}
		return val
	}
	if flag.Type != "int" {
		return
	}
	return int(*flag.Value.(*intValue))
}

func (f *FlagSet) GetUint(name string) (v uint) {
	flag := f.Lookup(name)
	if flag == nil {
		return
	}
	if flag.Type == "string" {
		val, err := hstr.ToUint(flag.Value.String())
		if err != nil {
			return
		}
		return val
	}
	if flag.Type != "uint" {
		return
	}
	return uint(*flag.Value.(*uintValue))
}

func (f *FlagSet) GetInt64(name string) (v int64) {
	flag := f.Lookup(name)
	if flag == nil {
		return
	}
	if flag.Type == "string" {
		val, err := hstr.ToInt64(flag.Value.String())
		if err != nil {
			return
		}
		return val
	}
	if flag.Type != "int64" {
		return
	}
	return int64(*flag.Value.(*int64Value))
}

func (f *FlagSet) GetInt32(name string) (v int32) {
	flag := f.Lookup(name)
	if flag == nil {
		return
	}
	if flag.Type == "string" {
		val, err := hstr.ToInt32(flag.Value.String())
		if err != nil {
			return
		}
		return val
	}
	if flag.Type != "int32" {
		return
	}
	return int32(*flag.Value.(*int32Value))
}

func (f *FlagSet) GetInt16(name string) (v int16) {
	flag := f.Lookup(name)
	if flag == nil {
		return
	}
	if flag.Type == "string" {
		val, err := hstr.ToInt16(flag.Value.String())
		if err != nil {
			return
		}
		return val
	}
	if flag.Type != "int16" {
		return
	}
	return int16(*flag.Value.(*int16Value))
}

func (f *FlagSet) GetInt8(name string) (v int8) {
	flag := f.Lookup(name)
	if flag == nil {
		return
	}
	if flag.Type == "string" {
		val, err := hstr.ToInt8(flag.Value.String())
		if err != nil {
			return
		}
		return val
	}
	if flag.Type != "int8" {
		return
	}
	return int8(*flag.Value.(*int8Value))
}

func (f *FlagSet) GetUint64(name string) (v uint64) {
	flag := f.Lookup(name)
	if flag == nil {
		return
	}
	if flag.Type == "string" {
		val, err := hstr.ToUint64(flag.Value.String())
		if err != nil {
			return
		}
		return val
	}
	if flag.Type != "uint64" {
		return
	}
	return uint64(*flag.Value.(*uint64Value))
}

func (f *FlagSet) GetUint32(name string) (v uint32) {
	flag := f.Lookup(name)
	if flag == nil {
		return
	}
	if flag.Type == "string" {
		val, err := hstr.ToUint32(flag.Value.String())
		if err != nil {
			return
		}
		return val
	}
	if flag.Type != "uint32" {
		return
	}
	return uint32(*flag.Value.(*uint32Value))
}

func (f *FlagSet) GetUint16(name string) (v uint16) {
	flag := f.Lookup(name)
	if flag == nil {
		return
	}
	if flag.Type == "string" {
		val, err := hstr.ToUint16(flag.Value.String())
		if err != nil {
			return
		}
		return val
	}
	if flag.Type != "uint16" {
		return
	}
	return uint16(*flag.Value.(*uint16Value))
}

func (f *FlagSet) GetUint8(name string) (v uint8) {
	flag := f.Lookup(name)
	if flag == nil {
		return
	}
	if flag.Type == "string" {
		val, err := hstr.ToUint8(flag.Value.String())
		if err != nil {
			return
		}
		return val
	}
	if flag.Type != "uint8" {
		return
	}
	return uint8(*flag.Value.(*uint8Value))
}

func (f *FlagSet) GetFloat64(name string) (v float64) {
	flag := f.Lookup(name)
	if flag == nil {
		return
	}
	if flag.Type == "string" {
		val, err := hstr.ToFloat64(flag.Value.String())
		if err != nil {
			return
		}
		return val
	}
	if flag.Type != "float64" {
		return
	}
	return float64(*flag.Value.(*float64Value))
}

func (f *FlagSet) GetFloat32(name string) (v float32) {
	flag := f.Lookup(name)
	if flag == nil {
		return
	}
	if flag.Type == "string" {
		val, err := hstr.ToFloat32(flag.Value.String())
		if err != nil {
			return
		}
		return val
	}
	if flag.Type != "float32" {
		return
	}
	return float32(*flag.Value.(*float32Value))
}

func (f *FlagSet) GetDuration(name string) (v time.Duration) {
	flag := f.Lookup(name)
	if flag == nil {
		return
	}
	if flag.Type == "string" {
		val, err := hstr.ToDuration(flag.Value.String())
		if err != nil {
			return
		}
		return val
	}
	if flag.Type != "duration" {
		return
	}
	return time.Duration(*flag.Value.(*durationValue))
}

func (f *FlagSet) GetTime(name string) (v time.Time) {
	flag := f.Lookup(name)
	if flag == nil {
		return
	}
	if flag.Type == "string" {
		val, err := hstr.ToTime(flag.Value.String())
		if err != nil {
			return
		}
		return val
	}
	if flag.Type != "time" {
		return
	}
	return time.Time(*flag.Value.(*timeValue))
}

func (f *FlagSet) GetIP(name string) (v net.IP) {
	flag := f.Lookup(name)
	if flag == nil {
		return
	}
	if flag.Type == "string" {
		val, err := hstr.ToIP(flag.Value.String())
		if err != nil {
			return
		}
		return val
	}
	if flag.Type != "ip" {
		return
	}
	return net.IP(*flag.Value.(*ipValue))
}

func (f *FlagSet) GetString(name string) string {
	flag := f.Lookup(name)
	if flag == nil {
		return ""
	}
	return flag.Value.String()
}

func (f *FlagSet) GetBoolSlice(name string) (v []bool) {
	flag := f.Lookup(name)
	if flag == nil {
		return
	}
	if flag.Type == "string" {
		val, err := hstr.ToBoolSlice(flag.Value.String())
		if err != nil {
			return
		}
		return val
	}
	if flag.Type != "[]bool" {
		return
	}
	return []bool(*flag.Value.(*boolSliceValue))
}

func (f *FlagSet) GetIntSlice(name string) (v []int) {
	flag := f.Lookup(name)
	if flag == nil {
		return
	}
	if flag.Type == "string" {
		val, err := hstr.ToIntSlice(flag.Value.String())
		if err != nil {
			return
		}
		return val
	}
	if flag.Type != "[]int" {
		return
	}
	return []int(*flag.Value.(*intSliceValue))
}

func (f *FlagSet) GetUintSlice(name string) (v []uint) {
	flag := f.Lookup(name)
	if flag == nil {
		return
	}
	if flag.Type == "string" {
		val, err := hstr.ToUintSlice(flag.Value.String())
		if err != nil {
			return
		}
		return val
	}
	if flag.Type != "[]uint" {
		return
	}
	return []uint(*flag.Value.(*uintSliceValue))
}

func (f *FlagSet) GetInt64Slice(name string) (v []int64) {
	flag := f.Lookup(name)
	if flag == nil {
		return
	}
	if flag.Type == "string" {
		val, err := hstr.ToInt64Slice(flag.Value.String())
		if err != nil {
			return
		}
		return val
	}
	if flag.Type != "[]int64" {
		return
	}
	return []int64(*flag.Value.(*int64SliceValue))
}

func (f *FlagSet) GetInt32Slice(name string) (v []int32) {
	flag := f.Lookup(name)
	if flag == nil {
		return
	}
	if flag.Type == "string" {
		val, err := hstr.ToInt32Slice(flag.Value.String())
		if err != nil {
			return
		}
		return val
	}
	if flag.Type != "[]int32" {
		return
	}
	return []int32(*flag.Value.(*int32SliceValue))
}

func (f *FlagSet) GetInt16Slice(name string) (v []int16) {
	flag := f.Lookup(name)
	if flag == nil {
		return
	}
	if flag.Type == "string" {
		val, err := hstr.ToInt16Slice(flag.Value.String())
		if err != nil {
			return
		}
		return val
	}
	if flag.Type != "[]int16" {
		return
	}
	return []int16(*flag.Value.(*int16SliceValue))
}

func (f *FlagSet) GetInt8Slice(name string) (v []int8) {
	flag := f.Lookup(name)
	if flag == nil {
		return
	}
	if flag.Type == "string" {
		val, err := hstr.ToInt8Slice(flag.Value.String())
		if err != nil {
			return
		}
		return val
	}
	if flag.Type != "[]int8" {
		return
	}
	return []int8(*flag.Value.(*int8SliceValue))
}

func (f *FlagSet) GetUint64Slice(name string) (v []uint64) {
	flag := f.Lookup(name)
	if flag == nil {
		return
	}
	if flag.Type == "string" {
		val, err := hstr.ToUint64Slice(flag.Value.String())
		if err != nil {
			return
		}
		return val
	}
	if flag.Type != "[]uint64" {
		return
	}
	return []uint64(*flag.Value.(*uint64SliceValue))
}

func (f *FlagSet) GetUint32Slice(name string) (v []uint32) {
	flag := f.Lookup(name)
	if flag == nil {
		return
	}
	if flag.Type == "string" {
		val, err := hstr.ToUint32Slice(flag.Value.String())
		if err != nil {
			return
		}
		return val
	}
	if flag.Type != "[]uint32" {
		return
	}
	return []uint32(*flag.Value.(*uint32SliceValue))
}

func (f *FlagSet) GetUint16Slice(name string) (v []uint16) {
	flag := f.Lookup(name)
	if flag == nil {
		return
	}
	if flag.Type == "string" {
		val, err := hstr.ToUint16Slice(flag.Value.String())
		if err != nil {
			return
		}
		return val
	}
	if flag.Type != "[]uint16" {
		return
	}
	return []uint16(*flag.Value.(*uint16SliceValue))
}

func (f *FlagSet) GetUint8Slice(name string) (v []uint8) {
	flag := f.Lookup(name)
	if flag == nil {
		return
	}
	if flag.Type == "string" {
		val, err := hstr.ToUint8Slice(flag.Value.String())
		if err != nil {
			return
		}
		return val
	}
	if flag.Type != "[]uint8" {
		return
	}
	return []uint8(*flag.Value.(*uint8SliceValue))
}

func (f *FlagSet) GetFloat64Slice(name string) (v []float64) {
	flag := f.Lookup(name)
	if flag == nil {
		return
	}
	if flag.Type == "string" {
		val, err := hstr.ToFloat64Slice(flag.Value.String())
		if err != nil {
			return
		}
		return val
	}
	if flag.Type != "[]float64" {
		return
	}
	return []float64(*flag.Value.(*float64SliceValue))
}

func (f *FlagSet) GetFloat32Slice(name string) (v []float32) {
	flag := f.Lookup(name)
	if flag == nil {
		return
	}
	if flag.Type == "string" {
		val, err := hstr.ToFloat32Slice(flag.Value.String())
		if err != nil {
			return
		}
		return val
	}
	if flag.Type != "[]float32" {
		return
	}
	return []float32(*flag.Value.(*float32SliceValue))
}

func (f *FlagSet) GetDurationSlice(name string) (v []time.Duration) {
	flag := f.Lookup(name)
	if flag == nil {
		return
	}
	if flag.Type == "string" {
		val, err := hstr.ToDurationSlice(flag.Value.String())
		if err != nil {
			return
		}
		return val
	}
	if flag.Type != "[]duration" {
		return
	}
	return []time.Duration(*flag.Value.(*durationSliceValue))
}

func (f *FlagSet) GetTimeSlice(name string) (v []time.Time) {
	flag := f.Lookup(name)
	if flag == nil {
		return
	}
	if flag.Type == "string" {
		val, err := hstr.ToTimeSlice(flag.Value.String())
		if err != nil {
			return
		}
		return val
	}
	if flag.Type != "[]time" {
		return
	}
	return []time.Time(*flag.Value.(*timeSliceValue))
}

func (f *FlagSet) GetIPSlice(name string) (v []net.IP) {
	flag := f.Lookup(name)
	if flag == nil {
		return
	}
	if flag.Type == "string" {
		val, err := hstr.ToIPSlice(flag.Value.String())
		if err != nil {
			return
		}
		return val
	}
	if flag.Type != "[]ip" {
		return
	}
	return []net.IP(*flag.Value.(*ipSliceValue))
}

func (f *FlagSet) GetStringSlice(name string) (v []string) {
	flag := f.Lookup(name)
	if flag == nil {
		return
	}
	if flag.Type == "string" {
		val, err := hstr.ToStringSlice(flag.Value.String())
		if err != nil {
			return
		}
		return val
	}
	if flag.Type != "[]string" {
		return
	}
	return []string(*flag.Value.(*stringSliceValue))
}
