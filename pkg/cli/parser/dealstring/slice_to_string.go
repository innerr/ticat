package dealstring

import (
	"bytes"
	"net"
	"strings"
	"time"
)

func StringSliceTo(vs []string) string {
	return strings.Join(vs, ",")
}

func BoolSliceTo(vs []bool) string {
	var buffer bytes.Buffer
	for idx, v := range vs {
		buffer.WriteString(BoolTo(v))
		if idx != len(vs)-1 {
			buffer.WriteString(",")
		}
	}
	return buffer.String()
}

func IntSliceTo(vs []int) string {
	var buffer bytes.Buffer
	for idx, v := range vs {
		buffer.WriteString(IntTo(v))
		if idx != len(vs)-1 {
			buffer.WriteString(",")
		}
	}
	return buffer.String()
}

func UintSliceTo(vs []uint) string {
	var buffer bytes.Buffer
	for idx, v := range vs {
		buffer.WriteString(UintTo(v))
		if idx != len(vs)-1 {
			buffer.WriteString(",")
		}
	}
	return buffer.String()
}

func Int64SliceTo(vs []int64) string {
	var buffer bytes.Buffer
	for idx, v := range vs {
		buffer.WriteString(Int64To(v))
		if idx != len(vs)-1 {
			buffer.WriteString(",")
		}
	}
	return buffer.String()
}

func Int32SliceTo(vs []int32) string {
	var buffer bytes.Buffer
	for idx, v := range vs {
		buffer.WriteString(Int32To(v))
		if idx != len(vs)-1 {
			buffer.WriteString(",")
		}
	}
	return buffer.String()
}

func Int16SliceTo(vs []int16) string {
	var buffer bytes.Buffer
	for idx, v := range vs {
		buffer.WriteString(Int16To(v))
		if idx != len(vs)-1 {
			buffer.WriteString(",")
		}
	}
	return buffer.String()
}

func Int8SliceTo(vs []int8) string {
	var buffer bytes.Buffer
	for idx, v := range vs {
		buffer.WriteString(Int8To(v))
		if idx != len(vs)-1 {
			buffer.WriteString(",")
		}
	}
	return buffer.String()
}

func Uint64SliceTo(vs []uint64) string {
	var buffer bytes.Buffer
	for idx, v := range vs {
		buffer.WriteString(Uint64To(v))
		if idx != len(vs)-1 {
			buffer.WriteString(",")
		}
	}
	return buffer.String()
}

func Uint32SliceTo(vs []uint32) string {
	var buffer bytes.Buffer
	for idx, v := range vs {
		buffer.WriteString(Uint32To(v))
		if idx != len(vs)-1 {
			buffer.WriteString(",")
		}
	}
	return buffer.String()
}

func Uint16SliceTo(vs []uint16) string {
	var buffer bytes.Buffer
	for idx, v := range vs {
		buffer.WriteString(Uint16To(v))
		if idx != len(vs)-1 {
			buffer.WriteString(",")
		}
	}
	return buffer.String()
}

func Uint8SliceTo(vs []uint8) string {
	var buffer bytes.Buffer
	for idx, v := range vs {
		buffer.WriteString(Uint8To(v))
		if idx != len(vs)-1 {
			buffer.WriteString(",")
		}
	}
	return buffer.String()
}

func Float64SliceTo(vs []float64) string {
	var buffer bytes.Buffer
	for idx, v := range vs {
		buffer.WriteString(Float64To(v))
		if idx != len(vs)-1 {
			buffer.WriteString(",")
		}
	}
	return buffer.String()
}

func Float32SliceTo(vs []float32) string {
	var buffer bytes.Buffer
	for idx, v := range vs {
		buffer.WriteString(Float32To(v))
		if idx != len(vs)-1 {
			buffer.WriteString(",")
		}
	}
	return buffer.String()
}

func DurationSliceTo(vs []time.Duration) string {
	var buffer bytes.Buffer
	for idx, v := range vs {
		buffer.WriteString(DurationTo(v))
		if idx != len(vs)-1 {
			buffer.WriteString(",")
		}
	}
	return buffer.String()
}

func TimeSliceTo(vs []time.Time) string {
	var buffer bytes.Buffer
	for idx, v := range vs {
		buffer.WriteString(TimeTo(v))
		if idx != len(vs)-1 {
			buffer.WriteString(",")
		}
	}
	return buffer.String()
}

func IPSliceTo(vs []net.IP) string {
	var buffer bytes.Buffer
	for idx, v := range vs {
		buffer.WriteString(IPTo(v))
		if idx != len(vs)-1 {
			buffer.WriteString(",")
		}
	}
	return buffer.String()
}
