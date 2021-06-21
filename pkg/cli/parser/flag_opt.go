package parser

import (
	"fmt"
	"net"
	"time"

	hstr "github.com/pingcap/ticat/pkg/cli/parser/dealstring"
)

func (f *FlagSet) Lookup(name string) *Flag {
	flag, ok := f.nameToFlag[name]
	if ok {
		return flag
	}
	k, ok := f.shorthandToName[name]
	if ok {
		return f.nameToFlag[k]
	}
	return nil
}

func (f *FlagSet) Set(name string, val string) error {
	flag := f.Lookup(name)
	if flag == nil {
		return fmt.Errorf("no such flag, Name [%v]", name)
	}

	return flag.Set(val)
}

func (f *FlagSet) Visit(callback func(f *Flag)) {
	for _, flag := range f.nameToFlag {
		if flag.Assigned {
			callback(flag)
		}
	}
}

func (f *FlagSet) VisitAll(callback func(f *Flag)) {
	for _, flag := range f.nameToFlag {
		callback(flag)
	}
}

func (f *FlagSet) Parsed() bool {
	return f.parsed
}

func (f *FlagSet) NArg() int {
	return len(f.args)
}

func (f *FlagSet) Args() []string {
	return f.args
}

func (f *FlagSet) Arg(i int) string {
	if i >= len(f.args) || i < 0 {
		return ""
	}

	return f.args[i]
}

func (f *FlagSet) NFlag() int {
	n := 0
	for _, flag := range f.nameToFlag {
		if flag.Assigned {
			n++
		}
	}
	return n
}

func (f *FlagSet) PrintDefaults() {
	fmt.Println(f.Usage())
}

func (f *FlagSet) addFlagAutoShorthand(name string, usage string, typeStr string, defaultValue string) error {
	if len(name) == 1 {
		return f.addFlag(name, usage, name, typeStr, false, defaultValue)
	}
	return f.addFlag(name, usage, "", typeStr, false, defaultValue)
}

func (f *FlagSet) Bool(name string, defaultValue bool, usage string) *bool {
	if err := f.addFlagAutoShorthand(name, usage, "bool", hstr.BoolTo(defaultValue)); err != nil {
		panic(err)
	}
	return (*bool)(f.nameToFlag[name].Value.(*boolValue))
}

func (f *FlagSet) Int(name string, defaultValue int, usage string) *int {
	if err := f.addFlagAutoShorthand(name, usage, "int", hstr.IntTo(defaultValue)); err != nil {
		panic(err)
	}
	return (*int)(f.nameToFlag[name].Value.(*intValue))
}

func (f *FlagSet) Uint(name string, defaultValue uint, usage string) *uint {
	if err := f.addFlagAutoShorthand(name, usage, "uint", hstr.UintTo(defaultValue)); err != nil {
		panic(err)
	}
	return (*uint)(f.nameToFlag[name].Value.(*uintValue))
}

func (f *FlagSet) Int64(name string, defaultValue int64, usage string) *int64 {
	if err := f.addFlagAutoShorthand(name, usage, "int64", hstr.Int64To(defaultValue)); err != nil {
		panic(err)
	}
	return (*int64)(f.nameToFlag[name].Value.(*int64Value))
}

func (f *FlagSet) Int32(name string, defaultValue int32, usage string) *int32 {
	if err := f.addFlagAutoShorthand(name, usage, "int32", hstr.Int32To(defaultValue)); err != nil {
		panic(err)
	}
	return (*int32)(f.nameToFlag[name].Value.(*int32Value))
}

func (f *FlagSet) Int16(name string, defaultValue int16, usage string) *int16 {
	if err := f.addFlagAutoShorthand(name, usage, "int16", hstr.Int16To(defaultValue)); err != nil {
		panic(err)
	}
	return (*int16)(f.nameToFlag[name].Value.(*int16Value))
}

func (f *FlagSet) Int8(name string, defaultValue int8, usage string) *int8 {
	if err := f.addFlagAutoShorthand(name, usage, "int8", hstr.Int8To(defaultValue)); err != nil {
		panic(err)
	}
	return (*int8)(f.nameToFlag[name].Value.(*int8Value))
}

func (f *FlagSet) Uint64(name string, defaultValue uint64, usage string) *uint64 {
	if err := f.addFlagAutoShorthand(name, usage, "uint64", hstr.Uint64To(defaultValue)); err != nil {
		panic(err)
	}
	return (*uint64)(f.nameToFlag[name].Value.(*uint64Value))
}

func (f *FlagSet) Uint32(name string, defaultValue uint32, usage string) *uint32 {
	if err := f.addFlagAutoShorthand(name, usage, "uint32", hstr.Uint32To(defaultValue)); err != nil {
		panic(err)
	}
	return (*uint32)(f.nameToFlag[name].Value.(*uint32Value))
}

func (f *FlagSet) Uint16(name string, defaultValue uint16, usage string) *uint16 {
	if err := f.addFlagAutoShorthand(name, usage, "uint16", hstr.Uint16To(defaultValue)); err != nil {
		panic(err)
	}
	return (*uint16)(f.nameToFlag[name].Value.(*uint16Value))
}

func (f *FlagSet) Uint8(name string, defaultValue uint8, usage string) *uint8 {
	if err := f.addFlagAutoShorthand(name, usage, "uint8", hstr.Uint8To(defaultValue)); err != nil {
		panic(err)
	}
	return (*uint8)(f.nameToFlag[name].Value.(*uint8Value))
}

func (f *FlagSet) Float64(name string, defaultValue float64, usage string) *float64 {
	if err := f.addFlagAutoShorthand(name, usage, "float64", hstr.Float64To(defaultValue)); err != nil {
		panic(err)
	}
	return (*float64)(f.nameToFlag[name].Value.(*float64Value))
}

func (f *FlagSet) Float32(name string, defaultValue float32, usage string) *float32 {
	if err := f.addFlagAutoShorthand(name, usage, "float32", hstr.Float32To(defaultValue)); err != nil {
		panic(err)
	}
	return (*float32)(f.nameToFlag[name].Value.(*float32Value))
}

func (f *FlagSet) Duration(name string, defaultValue time.Duration, usage string) *time.Duration {
	if err := f.addFlagAutoShorthand(name, usage, "duration", hstr.DurationTo(defaultValue)); err != nil {
		panic(err)
	}
	return (*time.Duration)(f.nameToFlag[name].Value.(*durationValue))
}

func (f *FlagSet) Time(name string, defaultValue time.Time, usage string) *time.Time {
	if err := f.addFlagAutoShorthand(name, usage, "time", hstr.TimeTo(defaultValue)); err != nil {
		panic(err)
	}
	return (*time.Time)(f.nameToFlag[name].Value.(*timeValue))
}

func (f *FlagSet) IP(name string, defaultValue net.IP, usage string) *net.IP {
	if err := f.addFlagAutoShorthand(name, usage, "ip", hstr.IPTo(defaultValue)); err != nil {
		panic(err)
	}
	return (*net.IP)(f.nameToFlag[name].Value.(*ipValue))
}

func (f *FlagSet) String(name string, defaultValue string, usage string) *string {
	if err := f.addFlagAutoShorthand(name, usage, "string", defaultValue); err != nil {
		panic(err)
	}
	return (*string)(f.nameToFlag[name].Value.(*stringValue))
}

func (f *FlagSet) BoolSlice(name string, defaultValue []bool, usage string) *[]bool {
	if err := f.addFlagAutoShorthand(name, usage, "[]bool", hstr.BoolSliceTo(defaultValue)); err != nil {
		panic(err)
	}
	return (*[]bool)(f.nameToFlag[name].Value.(*boolSliceValue))
}

func (f *FlagSet) IntSlice(name string, defaultValue []int, usage string) *[]int {
	if err := f.addFlagAutoShorthand(name, usage, "[]int", hstr.IntSliceTo(defaultValue)); err != nil {
		panic(err)
	}
	return (*[]int)(f.nameToFlag[name].Value.(*intSliceValue))
}

func (f *FlagSet) UintSlice(name string, defaultValue []uint, usage string) *[]uint {
	if err := f.addFlagAutoShorthand(name, usage, "[]uint", hstr.UintSliceTo(defaultValue)); err != nil {
		panic(err)
	}
	return (*[]uint)(f.nameToFlag[name].Value.(*uintSliceValue))
}

func (f *FlagSet) Int64Slice(name string, defaultValue []int64, usage string) *[]int64 {
	if err := f.addFlagAutoShorthand(name, usage, "[]int64", hstr.Int64SliceTo(defaultValue)); err != nil {
		panic(err)
	}
	return (*[]int64)(f.nameToFlag[name].Value.(*int64SliceValue))
}

func (f *FlagSet) Int32Slice(name string, defaultValue []int32, usage string) *[]int32 {
	if err := f.addFlagAutoShorthand(name, usage, "[]int32", hstr.Int32SliceTo(defaultValue)); err != nil {
		panic(err)
	}
	return (*[]int32)(f.nameToFlag[name].Value.(*int32SliceValue))
}

func (f *FlagSet) Int16Slice(name string, defaultValue []int16, usage string) *[]int16 {
	if err := f.addFlagAutoShorthand(name, usage, "[]int16", hstr.Int16SliceTo(defaultValue)); err != nil {
		panic(err)
	}
	return (*[]int16)(f.nameToFlag[name].Value.(*int16SliceValue))
}

func (f *FlagSet) Int8Slice(name string, defaultValue []int8, usage string) *[]int8 {
	if err := f.addFlagAutoShorthand(name, usage, "[]int8", hstr.Int8SliceTo(defaultValue)); err != nil {
		panic(err)
	}
	return (*[]int8)(f.nameToFlag[name].Value.(*int8SliceValue))
}

func (f *FlagSet) Uint64Slice(name string, defaultValue []uint64, usage string) *[]uint64 {
	if err := f.addFlagAutoShorthand(name, usage, "[]uint64", hstr.Uint64SliceTo(defaultValue)); err != nil {
		panic(err)
	}
	return (*[]uint64)(f.nameToFlag[name].Value.(*uint64SliceValue))
}

func (f *FlagSet) Uint32Slice(name string, defaultValue []uint32, usage string) *[]uint32 {
	if err := f.addFlagAutoShorthand(name, usage, "[]uint32", hstr.Uint32SliceTo(defaultValue)); err != nil {
		panic(err)
	}
	return (*[]uint32)(f.nameToFlag[name].Value.(*uint32SliceValue))
}

func (f *FlagSet) Uint16Slice(name string, defaultValue []uint16, usage string) *[]uint16 {
	if err := f.addFlagAutoShorthand(name, usage, "[]uint16", hstr.Uint16SliceTo(defaultValue)); err != nil {
		panic(err)
	}
	return (*[]uint16)(f.nameToFlag[name].Value.(*uint16SliceValue))
}

func (f *FlagSet) Uint8Slice(name string, defaultValue []uint8, usage string) *[]uint8 {
	if err := f.addFlagAutoShorthand(name, usage, "[]uint8", hstr.Uint8SliceTo(defaultValue)); err != nil {
		panic(err)
	}
	return (*[]uint8)(f.nameToFlag[name].Value.(*uint8SliceValue))
}

func (f *FlagSet) Float64Slice(name string, defaultValue []float64, usage string) *[]float64 {
	if err := f.addFlagAutoShorthand(name, usage, "[]float64", hstr.Float64SliceTo(defaultValue)); err != nil {
		panic(err)
	}
	return (*[]float64)(f.nameToFlag[name].Value.(*float64SliceValue))
}

func (f *FlagSet) Float32Slice(name string, defaultValue []float32, usage string) *[]float32 {
	if err := f.addFlagAutoShorthand(name, usage, "[]float32", hstr.Float32SliceTo(defaultValue)); err != nil {
		panic(err)
	}
	return (*[]float32)(f.nameToFlag[name].Value.(*float32SliceValue))
}

func (f *FlagSet) DurationSlice(name string, defaultValue []time.Duration, usage string) *[]time.Duration {
	if err := f.addFlagAutoShorthand(name, usage, "[]duration", hstr.DurationSliceTo(defaultValue)); err != nil {
		panic(err)
	}
	return (*[]time.Duration)(f.nameToFlag[name].Value.(*durationSliceValue))
}

func (f *FlagSet) TimeSlice(name string, defaultValue []time.Time, usage string) *[]time.Time {
	if err := f.addFlagAutoShorthand(name, usage, "[]time", hstr.TimeSliceTo(defaultValue)); err != nil {
		panic(err)
	}
	return (*[]time.Time)(f.nameToFlag[name].Value.(*timeSliceValue))
}

func (f *FlagSet) IPSlice(name string, defaultValue []net.IP, usage string) *[]net.IP {
	if err := f.addFlagAutoShorthand(name, usage, "[]ip", hstr.IPSliceTo(defaultValue)); err != nil {
		panic(err)
	}
	return (*[]net.IP)(f.nameToFlag[name].Value.(*ipSliceValue))
}

func (f *FlagSet) StringSlice(name string, defaultValue []string, usage string) *[]string {
	if err := f.addFlagAutoShorthand(name, usage, "[]string", hstr.StringSliceTo(defaultValue)); err != nil {
		panic(err)
	}
	return (*[]string)(f.nameToFlag[name].Value.(*stringSliceValue))
}

func (f *FlagSet) BoolVar(v *bool, name string, defaultValue bool, usage string) {
	*v = defaultValue
	if err := f.addFlagAutoShorthand(name, usage, "bool", hstr.BoolTo(defaultValue)); err != nil {
		panic(err)
	}
	f.nameToFlag[name].Value = (*boolValue)(v)
}

func (f *FlagSet) IntVar(v *int, name string, defaultValue int, usage string) {
	*v = defaultValue
	if err := f.addFlagAutoShorthand(name, usage, "int", hstr.IntTo(defaultValue)); err != nil {
		panic(err)
	}
	f.nameToFlag[name].Value = (*intValue)(v)
}

func (f *FlagSet) UintVar(v *uint, name string, defaultValue uint, usage string) {
	*v = defaultValue
	if err := f.addFlagAutoShorthand(name, usage, "uint", hstr.UintTo(defaultValue)); err != nil {
		panic(err)
	}
	f.nameToFlag[name].Value = (*uintValue)(v)
}

func (f *FlagSet) Int64Var(v *int64, name string, defaultValue int64, usage string) {
	*v = defaultValue
	if err := f.addFlagAutoShorthand(name, usage, "int64", hstr.Int64To(defaultValue)); err != nil {
		panic(err)
	}
	f.nameToFlag[name].Value = (*int64Value)(v)
}

func (f *FlagSet) Int32Var(v *int32, name string, defaultValue int32, usage string) {
	*v = defaultValue
	if err := f.addFlagAutoShorthand(name, usage, "int32", hstr.Int32To(defaultValue)); err != nil {
		panic(err)
	}
	f.nameToFlag[name].Value = (*int32Value)(v)
}

func (f *FlagSet) Int16Var(v *int16, name string, defaultValue int16, usage string) {
	*v = defaultValue
	if err := f.addFlagAutoShorthand(name, usage, "int16", hstr.Int16To(defaultValue)); err != nil {
		panic(err)
	}
	f.nameToFlag[name].Value = (*int16Value)(v)
}

func (f *FlagSet) Int8Var(v *int8, name string, defaultValue int8, usage string) {
	*v = defaultValue
	if err := f.addFlagAutoShorthand(name, usage, "int8", hstr.Int8To(defaultValue)); err != nil {
		panic(err)
	}
	f.nameToFlag[name].Value = (*int8Value)(v)
}

func (f *FlagSet) Uint64Var(v *uint64, name string, defaultValue uint64, usage string) {
	*v = defaultValue
	if err := f.addFlagAutoShorthand(name, usage, "uint64", hstr.Uint64To(defaultValue)); err != nil {
		panic(err)
	}
	f.nameToFlag[name].Value = (*uint64Value)(v)
}

func (f *FlagSet) Uint32Var(v *uint32, name string, defaultValue uint32, usage string) {
	*v = defaultValue
	if err := f.addFlagAutoShorthand(name, usage, "uint32", hstr.Uint32To(defaultValue)); err != nil {
		panic(err)
	}
	f.nameToFlag[name].Value = (*uint32Value)(v)
}

func (f *FlagSet) Uint16Var(v *uint16, name string, defaultValue uint16, usage string) {
	*v = defaultValue
	if err := f.addFlagAutoShorthand(name, usage, "uint16", hstr.Uint16To(defaultValue)); err != nil {
		panic(err)
	}
	f.nameToFlag[name].Value = (*uint16Value)(v)
}

func (f *FlagSet) Uint8Var(v *uint8, name string, defaultValue uint8, usage string) {
	*v = defaultValue
	if err := f.addFlagAutoShorthand(name, usage, "uint8", hstr.Uint8To(defaultValue)); err != nil {
		panic(err)
	}
	f.nameToFlag[name].Value = (*uint8Value)(v)
}

func (f *FlagSet) Float64Var(v *float64, name string, defaultValue float64, usage string) {
	*v = defaultValue
	if err := f.addFlagAutoShorthand(name, usage, "float64", hstr.Float64To(defaultValue)); err != nil {
		panic(err)
	}
	f.nameToFlag[name].Value = (*float64Value)(v)
}

func (f *FlagSet) Float32Var(v *float32, name string, defaultValue float32, usage string) {
	*v = defaultValue
	if err := f.addFlagAutoShorthand(name, usage, "float32", hstr.Float32To(defaultValue)); err != nil {
		panic(err)
	}
	f.nameToFlag[name].Value = (*float32Value)(v)
}

func (f *FlagSet) DurationVar(v *time.Duration, name string, defaultValue time.Duration, usage string) {
	*v = defaultValue
	if err := f.addFlagAutoShorthand(name, usage, "duration", hstr.DurationTo(defaultValue)); err != nil {
		panic(err)
	}
	f.nameToFlag[name].Value = (*durationValue)(v)
}

func (f *FlagSet) TimeVar(v *time.Time, name string, defaultValue time.Time, usage string) {
	*v = defaultValue
	if err := f.addFlagAutoShorthand(name, usage, "time", hstr.TimeTo(defaultValue)); err != nil {
		panic(err)
	}
	f.nameToFlag[name].Value = (*timeValue)(v)
}

func (f *FlagSet) IPVar(v *net.IP, name string, defaultValue net.IP, usage string) {
	*v = defaultValue
	if err := f.addFlagAutoShorthand(name, usage, "ip", hstr.IPTo(defaultValue)); err != nil {
		panic(err)
	}
	f.nameToFlag[name].Value = (*ipValue)(v)
}

func (f *FlagSet) StringVar(v *string, name string, defaultValue string, usage string) {
	*v = defaultValue
	if err := f.addFlagAutoShorthand(name, usage, "string", defaultValue); err != nil {
		panic(err)
	}
	f.nameToFlag[name].Value = (*stringValue)(v)
}

func (f *FlagSet) BoolSliceVar(v *[]bool, name string, defaultValue []bool, usage string) {
	*v = defaultValue
	if err := f.addFlagAutoShorthand(name, usage, "[]bool", hstr.BoolSliceTo(defaultValue)); err != nil {
		panic(err)
	}
	f.nameToFlag[name].Value = (*boolSliceValue)(v)
}

func (f *FlagSet) IntSliceVar(v *[]int, name string, defaultValue []int, usage string) {
	*v = defaultValue
	if err := f.addFlagAutoShorthand(name, usage, "[]int", hstr.IntSliceTo(defaultValue)); err != nil {
		panic(err)
	}
	f.nameToFlag[name].Value = (*intSliceValue)(v)
}

func (f *FlagSet) UintSliceVar(v *[]uint, name string, defaultValue []uint, usage string) {
	*v = defaultValue
	if err := f.addFlagAutoShorthand(name, usage, "[]uint", hstr.UintSliceTo(defaultValue)); err != nil {
		panic(err)
	}
	f.nameToFlag[name].Value = (*uintSliceValue)(v)
}

func (f *FlagSet) Int64SliceVar(v *[]int64, name string, defaultValue []int64, usage string) {
	*v = defaultValue
	if err := f.addFlagAutoShorthand(name, usage, "[]int64", hstr.Int64SliceTo(defaultValue)); err != nil {
		panic(err)
	}
	f.nameToFlag[name].Value = (*int64SliceValue)(v)
}

func (f *FlagSet) Int32SliceVar(v *[]int32, name string, defaultValue []int32, usage string) {
	*v = defaultValue
	if err := f.addFlagAutoShorthand(name, usage, "[]int32", hstr.Int32SliceTo(defaultValue)); err != nil {
		panic(err)
	}
	f.nameToFlag[name].Value = (*int32SliceValue)(v)
}

func (f *FlagSet) Int16SliceVar(v *[]int16, name string, defaultValue []int16, usage string) {
	*v = defaultValue
	if err := f.addFlagAutoShorthand(name, usage, "[]int16", hstr.Int16SliceTo(defaultValue)); err != nil {
		panic(err)
	}
	f.nameToFlag[name].Value = (*int16SliceValue)(v)
}

func (f *FlagSet) Int8SliceVar(v *[]int8, name string, defaultValue []int8, usage string) {
	*v = defaultValue
	if err := f.addFlagAutoShorthand(name, usage, "[]int8", hstr.Int8SliceTo(defaultValue)); err != nil {
		panic(err)
	}
	f.nameToFlag[name].Value = (*int8SliceValue)(v)
}

func (f *FlagSet) Uint64SliceVar(v *[]uint64, name string, defaultValue []uint64, usage string) {
	*v = defaultValue
	if err := f.addFlagAutoShorthand(name, usage, "[]uint64", hstr.Uint64SliceTo(defaultValue)); err != nil {
		panic(err)
	}
	f.nameToFlag[name].Value = (*uint64SliceValue)(v)
}

func (f *FlagSet) Uint32SliceVar(v *[]uint32, name string, defaultValue []uint32, usage string) {
	*v = defaultValue
	if err := f.addFlagAutoShorthand(name, usage, "[]uint32", hstr.Uint32SliceTo(defaultValue)); err != nil {
		panic(err)
	}
	f.nameToFlag[name].Value = (*uint32SliceValue)(v)
}

func (f *FlagSet) Uint16SliceVar(v *[]uint16, name string, defaultValue []uint16, usage string) {
	*v = defaultValue
	if err := f.addFlagAutoShorthand(name, usage, "[]uint16", hstr.Uint16SliceTo(defaultValue)); err != nil {
		panic(err)
	}
	f.nameToFlag[name].Value = (*uint16SliceValue)(v)
}

func (f *FlagSet) Uint8SliceVar(v *[]uint8, name string, defaultValue []uint8, usage string) {
	*v = defaultValue
	if err := f.addFlagAutoShorthand(name, usage, "[]uint8", hstr.Uint8SliceTo(defaultValue)); err != nil {
		panic(err)
	}
	f.nameToFlag[name].Value = (*uint8SliceValue)(v)
}

func (f *FlagSet) Float64SliceVar(v *[]float64, name string, defaultValue []float64, usage string) {
	*v = defaultValue
	if err := f.addFlagAutoShorthand(name, usage, "[]float64", hstr.Float64SliceTo(defaultValue)); err != nil {
		panic(err)
	}
	f.nameToFlag[name].Value = (*float64SliceValue)(v)
}

func (f *FlagSet) Float32SliceVar(v *[]float32, name string, defaultValue []float32, usage string) {
	*v = defaultValue
	if err := f.addFlagAutoShorthand(name, usage, "[]float32", hstr.Float32SliceTo(defaultValue)); err != nil {
		panic(err)
	}
	f.nameToFlag[name].Value = (*float32SliceValue)(v)
}

func (f *FlagSet) DurationSliceVar(v *[]time.Duration, name string, defaultValue []time.Duration, usage string) {
	*v = defaultValue
	if err := f.addFlagAutoShorthand(name, usage, "[]duration", hstr.DurationSliceTo(defaultValue)); err != nil {
		panic(err)
	}
	f.nameToFlag[name].Value = (*durationSliceValue)(v)
}

func (f *FlagSet) TimeSliceVar(v *[]time.Time, name string, defaultValue []time.Time, usage string) {
	*v = defaultValue
	if err := f.addFlagAutoShorthand(name, usage, "[]time", hstr.TimeSliceTo(defaultValue)); err != nil {
		panic(err)
	}
	f.nameToFlag[name].Value = (*timeSliceValue)(v)
}

func (f *FlagSet) IPSliceVar(v *[]net.IP, name string, defaultValue []net.IP, usage string) {
	*v = defaultValue
	if err := f.addFlagAutoShorthand(name, usage, "[]ip", hstr.IPSliceTo(defaultValue)); err != nil {
		panic(err)
	}
	f.nameToFlag[name].Value = (*ipSliceValue)(v)
}

func (f *FlagSet) StringSliceVar(v *[]string, name string, defaultValue []string, usage string) {
	*v = defaultValue
	if err := f.addFlagAutoShorthand(name, usage, "[]string", hstr.StringSliceTo(defaultValue)); err != nil {
		panic(err)
	}
	f.nameToFlag[name].Value = (*stringSliceValue)(v)
}
