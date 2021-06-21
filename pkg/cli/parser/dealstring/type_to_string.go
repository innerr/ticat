package dealstring

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"time"

	jsoniter "github.com/json-iterator/go"
	"gopkg.in/yaml.v2"
)

func Indent(prefix, str string) string {
	r := bufio.NewReader(bytes.NewBuffer([]byte(str)))
	w := &bytes.Buffer{}
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				w.WriteString(prefix)
				w.WriteString(line)
			}
			break
		}
		w.WriteString(prefix)
		w.WriteString(line)
	}

	return strings.TrimRight(w.String(), "\n ")
}

var json = jsoniter.ConfigCompatibleWithStandardLibrary

func ToJsonString(v interface{}) string {
	buf, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(buf)
}

func ToYamlString(v interface{}) string {
	buf, err := yaml.Marshal(v)
	if err != nil {
		return err.Error()
	}
	return string(buf)
}

func ToString(v interface{}) string {
	switch v.(type) {
	case bool:
		return BoolTo(v.(bool))
	case int:
		return IntTo(v.(int))
	case uint:
		return UintTo(v.(uint))
	case int64:
		return Int64To(v.(int64))
	case int32:
		return Int32To(v.(int32))
	case int16:
		return Int16To(v.(int16))
	case int8:
		return Int8To(v.(int8))
	case uint64:
		return Uint64To(v.(uint64))
	case uint32:
		return Uint32To(v.(uint32))
	case uint16:
		return Uint16To(v.(uint16))
	case uint8:
		return Uint8To(v.(uint8))
	case float64:
		return Float64To(v.(float64))
	case float32:
		return Float32To(v.(float32))
	case time.Duration:
		return DurationTo(v.(time.Duration))
	case time.Time:
		return TimeTo(v.(time.Time))
	case net.IP:
		return IPTo(v.(net.IP))
	case []string:
		return StringSliceTo(v.([]string))
	case []bool:
		return BoolSliceTo(v.([]bool))
	case []int:
		return IntSliceTo(v.([]int))
	case []uint:
		return UintSliceTo(v.([]uint))
	case []int64:
		return Int64SliceTo(v.([]int64))
	case []int32:
		return Int32SliceTo(v.([]int32))
	case []int16:
		return Int16SliceTo(v.([]int16))
	case []int8:
		return Int8SliceTo(v.([]int8))
	case []uint64:
		return Uint64SliceTo(v.([]uint64))
	case []uint32:
		return Uint32SliceTo(v.([]uint32))
	case []uint16:
		return Uint16SliceTo(v.([]uint16))
	case []uint8:
		return Uint8SliceTo(v.([]uint8))
	case []float64:
		return Float64SliceTo(v.([]float64))
	case []float32:
		return Float32SliceTo(v.([]float32))
	case []time.Duration:
		return DurationSliceTo(v.([]time.Duration))
	case []time.Time:
		return TimeSliceTo(v.([]time.Time))
	case []net.IP:
		return IPSliceTo(v.([]net.IP))
	default:
		return fmt.Sprintf("%v", v)
	}
}

func BoolTo(v bool) string {
	return fmt.Sprintf("%v", v)
}

func IntTo(v int) string {
	return strconv.Itoa(v)
}

func Int64To(v int64) string {
	return strconv.FormatInt(v, 10)
}

func Int32To(v int32) string {
	return strconv.FormatInt(int64(v), 10)
}

func Int16To(v int16) string {
	return strconv.FormatInt(int64(v), 10)
}

func Int8To(v int8) string {
	return strconv.FormatInt(int64(v), 10)
}

func UintTo(v uint) string {
	return strconv.FormatUint(uint64(v), 10)
}

func Uint64To(v uint64) string {
	return strconv.FormatUint(v, 10)
}

func Uint32To(v uint32) string {
	return strconv.FormatUint(uint64(v), 10)
}

func Uint16To(v uint16) string {
	return strconv.FormatUint(uint64(v), 10)
}

func Uint8To(v uint8) string {
	return strconv.FormatUint(uint64(v), 10)
}

func Float64To(v float64) string {
	return fmt.Sprintf("%f", v)
}

func Float32To(v float32) string {
	return fmt.Sprintf("%f", v)
}

func DurationTo(v time.Duration) string {
	return v.String()
}

func TimeTo(v time.Time) string {
	return v.Format(time.RFC3339)
}

func IPTo(v net.IP) string {
	if v == nil {
		return ""
	}
	return v.String()
}
