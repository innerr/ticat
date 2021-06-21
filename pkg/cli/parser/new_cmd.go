package parser

import (
	"fmt"
	"net"
	"os"
	"time"
)

var CommandLine = NewFlagSet(os.Args[0])

func Lookup(name string) *Flag {
	return CommandLine.Lookup(name)
}

func Set(name string, val string) error {
	return CommandLine.Set(name, val)
}

func Visit(callback func(f *Flag)) {
	CommandLine.Visit(callback)
}

func VisitAll(callback func(f *Flag)) {
	CommandLine.VisitAll(callback)
}

func Parsed() bool {
	return CommandLine.Parsed()
}

func NArg() int {
	return CommandLine.NArg()
}

func Args() []string {
	return CommandLine.Args()
}

func Arg(i int) string {
	return CommandLine.Arg(i)
}

func NFlag() int {
	return CommandLine.NFlag()
}

func PrintDefaults() {
	CommandLine.PrintDefaults()
}

func Parse() error {
	if CommandLine.Lookup("help") == nil && CommandLine.Lookup("h") == nil {
		CommandLine.AddFlag("help", "show usage", Shorthand("h"), Type("bool"))
	}
	if err := CommandLine.Parse(os.Args[1:]); err != nil {
		if CommandLine.GetBool("help") {
			fmt.Println(CommandLine.Usage())
			os.Exit(0)
		}
		return err
	}
	if CommandLine.GetBool("help") {
		fmt.Println(CommandLine.Usage())
		os.Exit(0)
	}

	return nil
}

func AddFlag(name string, usage string, opts ...FlagOption) {
	CommandLine.AddFlag(name, usage, opts...)
}

func AddPosFlag(name string, usage string, opts ...FlagOption) {
	CommandLine.AddPosFlag(name, usage, opts...)
}

func Usage() string {
	return CommandLine.Usage()
}

func Bind(v interface{}) error {
	return CommandLine.Bind(v)
}

func Unmarshal(v interface{}) error {
	return CommandLine.Unmarshal(v)
}

func Bool(name string, defaultValue bool, usage string) *bool {
	return CommandLine.Bool(name, defaultValue, usage)
}

func Int(name string, defaultValue int, usage string) *int {
	return CommandLine.Int(name, defaultValue, usage)
}

func Uint(name string, defaultValue uint, usage string) *uint {
	return CommandLine.Uint(name, defaultValue, usage)
}

func Int64(name string, defaultValue int64, usage string) *int64 {
	return CommandLine.Int64(name, defaultValue, usage)
}

func Int32(name string, defaultValue int32, usage string) *int32 {
	return CommandLine.Int32(name, defaultValue, usage)
}

func Int16(name string, defaultValue int16, usage string) *int16 {
	return CommandLine.Int16(name, defaultValue, usage)
}

func Int8(name string, defaultValue int8, usage string) *int8 {
	return CommandLine.Int8(name, defaultValue, usage)
}

func Uint64(name string, defaultValue uint64, usage string) *uint64 {
	return CommandLine.Uint64(name, defaultValue, usage)
}

func Uint32(name string, defaultValue uint32, usage string) *uint32 {
	return CommandLine.Uint32(name, defaultValue, usage)
}

func Uint16(name string, defaultValue uint16, usage string) *uint16 {
	return CommandLine.Uint16(name, defaultValue, usage)
}

func Uint8(name string, defaultValue uint8, usage string) *uint8 {
	return CommandLine.Uint8(name, defaultValue, usage)
}

func Float64(name string, defaultValue float64, usage string) *float64 {
	return CommandLine.Float64(name, defaultValue, usage)
}

func Float32(name string, defaultValue float32, usage string) *float32 {
	return CommandLine.Float32(name, defaultValue, usage)
}

func Duration(name string, defaultValue time.Duration, usage string) *time.Duration {
	return CommandLine.Duration(name, defaultValue, usage)
}

func Time(name string, defaultValue time.Time, usage string) *time.Time {
	return CommandLine.Time(name, defaultValue, usage)
}

func IP(name string, defaultValue net.IP, usage string) *net.IP {
	return CommandLine.IP(name, defaultValue, usage)
}

func String(name string, defaultValue string, usage string) *string {
	return CommandLine.String(name, defaultValue, usage)
}

func BoolSlice(name string, defaultValue []bool, usage string) *[]bool {
	return CommandLine.BoolSlice(name, defaultValue, usage)
}

func IntSlice(name string, defaultValue []int, usage string) *[]int {
	return CommandLine.IntSlice(name, defaultValue, usage)
}

func UintSlice(name string, defaultValue []uint, usage string) *[]uint {
	return CommandLine.UintSlice(name, defaultValue, usage)
}

func Int64Slice(name string, defaultValue []int64, usage string) *[]int64 {
	return CommandLine.Int64Slice(name, defaultValue, usage)
}

func Int32Slice(name string, defaultValue []int32, usage string) *[]int32 {
	return CommandLine.Int32Slice(name, defaultValue, usage)
}

func Int16Slice(name string, defaultValue []int16, usage string) *[]int16 {
	return CommandLine.Int16Slice(name, defaultValue, usage)
}

func Int8Slice(name string, defaultValue []int8, usage string) *[]int8 {
	return CommandLine.Int8Slice(name, defaultValue, usage)
}

func Uint64Slice(name string, defaultValue []uint64, usage string) *[]uint64 {
	return CommandLine.Uint64Slice(name, defaultValue, usage)
}

func Uint32Slice(name string, defaultValue []uint32, usage string) *[]uint32 {
	return CommandLine.Uint32Slice(name, defaultValue, usage)
}

func Uint16Slice(name string, defaultValue []uint16, usage string) *[]uint16 {
	return CommandLine.Uint16Slice(name, defaultValue, usage)
}

func Uint8Slice(name string, defaultValue []uint8, usage string) *[]uint8 {
	return CommandLine.Uint8Slice(name, defaultValue, usage)
}

func Float64Slice(name string, defaultValue []float64, usage string) *[]float64 {
	return CommandLine.Float64Slice(name, defaultValue, usage)
}

func Float32Slice(name string, defaultValue []float32, usage string) *[]float32 {
	return CommandLine.Float32Slice(name, defaultValue, usage)
}

func DurationSlice(name string, defaultValue []time.Duration, usage string) *[]time.Duration {
	return CommandLine.DurationSlice(name, defaultValue, usage)
}

func TimeSlice(name string, defaultValue []time.Time, usage string) *[]time.Time {
	return CommandLine.TimeSlice(name, defaultValue, usage)
}

func IPSlice(name string, defaultValue []net.IP, usage string) *[]net.IP {
	return CommandLine.IPSlice(name, defaultValue, usage)
}

func StringSlice(name string, defaultValue []string, usage string) *[]string {
	return CommandLine.StringSlice(name, defaultValue, usage)
}

func BoolVar(v *bool, name string, defaultValue bool, usage string) {
	CommandLine.BoolVar(v, name, defaultValue, usage)
}

func IntVar(v *int, name string, defaultValue int, usage string) {
	CommandLine.IntVar(v, name, defaultValue, usage)
}

func UintVar(v *uint, name string, defaultValue uint, usage string) {
	CommandLine.UintVar(v, name, defaultValue, usage)
}

func Int64Var(v *int64, name string, defaultValue int64, usage string) {
	CommandLine.Int64Var(v, name, defaultValue, usage)
}

func Int32Var(v *int32, name string, defaultValue int32, usage string) {
	CommandLine.Int32Var(v, name, defaultValue, usage)
}

func Int16Var(v *int16, name string, defaultValue int16, usage string) {
	CommandLine.Int16Var(v, name, defaultValue, usage)
}

func Int8Var(v *int8, name string, defaultValue int8, usage string) {
	CommandLine.Int8Var(v, name, defaultValue, usage)
}

func Uint64Var(v *uint64, name string, defaultValue uint64, usage string) {
	CommandLine.Uint64Var(v, name, defaultValue, usage)
}

func Uint32Var(v *uint32, name string, defaultValue uint32, usage string) {
	CommandLine.Uint32Var(v, name, defaultValue, usage)
}

func Uint16Var(v *uint16, name string, defaultValue uint16, usage string) {
	CommandLine.Uint16Var(v, name, defaultValue, usage)
}

func Uint8Var(v *uint8, name string, defaultValue uint8, usage string) {
	CommandLine.Uint8Var(v, name, defaultValue, usage)
}

func Float64Var(v *float64, name string, defaultValue float64, usage string) {
	CommandLine.Float64Var(v, name, defaultValue, usage)
}

func Float32Var(v *float32, name string, defaultValue float32, usage string) {
	CommandLine.Float32Var(v, name, defaultValue, usage)
}

func DurationVar(v *time.Duration, name string, defaultValue time.Duration, usage string) {
	CommandLine.DurationVar(v, name, defaultValue, usage)
}

func TimeVar(v *time.Time, name string, defaultValue time.Time, usage string) {
	CommandLine.TimeVar(v, name, defaultValue, usage)
}

func IPVar(v *net.IP, name string, defaultValue net.IP, usage string) {
	CommandLine.IPVar(v, name, defaultValue, usage)
}

func StringVar(v *string, name string, defaultValue string, usage string) {
	CommandLine.StringVar(v, name, defaultValue, usage)
}

func BoolSliceVar(v *[]bool, name string, defaultValue []bool, usage string) {
	CommandLine.BoolSliceVar(v, name, defaultValue, usage)
}

func IntSliceVar(v *[]int, name string, defaultValue []int, usage string) {
	CommandLine.IntSliceVar(v, name, defaultValue, usage)
}

func UintSliceVar(v *[]uint, name string, defaultValue []uint, usage string) {
	CommandLine.UintSliceVar(v, name, defaultValue, usage)
}

func Int64SliceVar(v *[]int64, name string, defaultValue []int64, usage string) {
	CommandLine.Int64SliceVar(v, name, defaultValue, usage)
}

func Int32SliceVar(v *[]int32, name string, defaultValue []int32, usage string) {
	CommandLine.Int32SliceVar(v, name, defaultValue, usage)
}

func Int16SliceVar(v *[]int16, name string, defaultValue []int16, usage string) {
	CommandLine.Int16SliceVar(v, name, defaultValue, usage)
}

func Int8SliceVar(v *[]int8, name string, defaultValue []int8, usage string) {
	CommandLine.Int8SliceVar(v, name, defaultValue, usage)
}

func Uint64SliceVar(v *[]uint64, name string, defaultValue []uint64, usage string) {
	CommandLine.Uint64SliceVar(v, name, defaultValue, usage)
}

func Uint32SliceVar(v *[]uint32, name string, defaultValue []uint32, usage string) {
	CommandLine.Uint32SliceVar(v, name, defaultValue, usage)
}

func Uint16SliceVar(v *[]uint16, name string, defaultValue []uint16, usage string) {
	CommandLine.Uint16SliceVar(v, name, defaultValue, usage)
}

func Uint8SliceVar(v *[]uint8, name string, defaultValue []uint8, usage string) {
	CommandLine.Uint8SliceVar(v, name, defaultValue, usage)
}

func Float64SliceVar(v *[]float64, name string, defaultValue []float64, usage string) {
	CommandLine.Float64SliceVar(v, name, defaultValue, usage)
}

func Float32SliceVar(v *[]float32, name string, defaultValue []float32, usage string) {
	CommandLine.Float32SliceVar(v, name, defaultValue, usage)
}

func DurationSliceVar(v *[]time.Duration, name string, defaultValue []time.Duration, usage string) {
	CommandLine.DurationSliceVar(v, name, defaultValue, usage)
}

func TimeSliceVar(v *[]time.Time, name string, defaultValue []time.Time, usage string) {
	CommandLine.TimeSliceVar(v, name, defaultValue, usage)
}

func IPSliceVar(v *[]net.IP, name string, defaultValue []net.IP, usage string) {
	CommandLine.IPSliceVar(v, name, defaultValue, usage)
}

func StringSliceVar(v *[]string, name string, defaultValue []string, usage string) {
	CommandLine.StringSliceVar(v, name, defaultValue, usage)
}

func GetBool(name string) bool {
	return CommandLine.GetBool(name)
}

func GetInt(name string) int {
	return CommandLine.GetInt(name)
}

func GetUint(name string) uint {
	return CommandLine.GetUint(name)
}

func GetInt64(name string) int64 {
	return CommandLine.GetInt64(name)
}

func GetInt32(name string) int32 {
	return CommandLine.GetInt32(name)
}

func GetInt16(name string) int16 {
	return CommandLine.GetInt16(name)
}

func GetInt8(name string) int8 {
	return CommandLine.GetInt8(name)
}

func GetUint64(name string) uint64 {
	return CommandLine.GetUint64(name)
}

func GetUint32(name string) uint32 {
	return CommandLine.GetUint32(name)
}

func GetUint16(name string) uint16 {
	return CommandLine.GetUint16(name)
}

func GetUint8(name string) uint8 {
	return CommandLine.GetUint8(name)
}

func GetFloat64(name string) float64 {
	return CommandLine.GetFloat64(name)
}

func GetFloat32(name string) float32 {
	return CommandLine.GetFloat32(name)
}

func GetDuration(name string) time.Duration {
	return CommandLine.GetDuration(name)
}

func GetTime(name string) time.Time {
	return CommandLine.GetTime(name)
}

func GetIP(name string) net.IP {
	return CommandLine.GetIP(name)
}

func GetString(name string) string {
	return CommandLine.GetString(name)
}

func GetBoolSlice(name string) []bool {
	return CommandLine.GetBoolSlice(name)
}

func GetIntSlice(name string) []int {
	return CommandLine.GetIntSlice(name)
}

func GetUintSlice(name string) []uint {
	return CommandLine.GetUintSlice(name)
}

func GetInt64Slice(name string) []int64 {
	return CommandLine.GetInt64Slice(name)
}

func GetInt32Slice(name string) []int32 {
	return CommandLine.GetInt32Slice(name)
}

func GetInt16Slice(name string) []int16 {
	return CommandLine.GetInt16Slice(name)
}

func GetInt8Slice(name string) []int8 {
	return CommandLine.GetInt8Slice(name)
}

func GetUint64Slice(name string) []uint64 {
	return CommandLine.GetUint64Slice(name)
}

func GetUint32Slice(name string) []uint32 {
	return CommandLine.GetUint32Slice(name)
}

func GetUint16Slice(name string) []uint16 {
	return CommandLine.GetUint16Slice(name)
}

func GetUint8Slice(name string) []uint8 {
	return CommandLine.GetUint8Slice(name)
}

func GetFloat64Slice(name string) []float64 {
	return CommandLine.GetFloat64Slice(name)
}

func GetFloat32Slice(name string) []float32 {
	return CommandLine.GetFloat32Slice(name)
}

func GetDurationSlice(name string) []time.Duration {
	return CommandLine.GetDurationSlice(name)
}

func GetTimeSlice(name string) []time.Time {
	return CommandLine.GetTimeSlice(name)
}

func GetIPSlice(name string) []net.IP {
	return CommandLine.GetIPSlice(name)
}

func GetStringSlice(name string) []string {
	return CommandLine.GetStringSlice(name)
}
