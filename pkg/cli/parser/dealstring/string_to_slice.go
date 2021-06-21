package dealstring

import (
	"net"
	"strings"
	"time"
)

func ToStringSlice(str string) ([]string, error) {
	if str == "" {
		return []string{}, nil
	}
	return strings.Split(str, ","), nil
}

func ToBoolSlice(str string) ([]bool, error) {
	vals := strings.Split(str, ",")
	res := make([]bool, 0, len(vals))
	for _, val := range vals {
		v, err := ToBool(val)
		if err != nil {
			return nil, err
		}
		res = append(res, v)
	}
	return res, nil
}

func ToIntSlice(str string) ([]int, error) {
	vals := strings.Split(str, ",")
	res := make([]int, 0, len(vals))
	for _, val := range vals {
		v, err := ToInt(val)
		if err != nil {
			return nil, err
		}
		res = append(res, v)
	}
	return res, nil
}

func ToUintSlice(str string) ([]uint, error) {
	vals := strings.Split(str, ",")
	res := make([]uint, 0, len(vals))
	for _, val := range vals {
		v, err := ToUint(val)
		if err != nil {
			return nil, err
		}
		res = append(res, v)
	}
	return res, nil
}

func ToInt64Slice(str string) ([]int64, error) {
	vals := strings.Split(str, ",")
	res := make([]int64, 0, len(vals))
	for _, val := range vals {
		v, err := ToInt64(val)
		if err != nil {
			return nil, err
		}
		res = append(res, v)
	}
	return res, nil
}

func ToInt32Slice(str string) ([]int32, error) {
	vals := strings.Split(str, ",")
	res := make([]int32, 0, len(vals))
	for _, val := range vals {
		v, err := ToInt32(val)
		if err != nil {
			return nil, err
		}
		res = append(res, v)
	}
	return res, nil
}

func ToInt16Slice(str string) ([]int16, error) {
	vals := strings.Split(str, ",")
	res := make([]int16, 0, len(vals))
	for _, val := range vals {
		v, err := ToInt16(val)
		if err != nil {
			return nil, err
		}
		res = append(res, v)
	}
	return res, nil
}

func ToInt8Slice(str string) ([]int8, error) {
	vals := strings.Split(str, ",")
	res := make([]int8, 0, len(vals))
	for _, val := range vals {
		v, err := ToInt8(val)
		if err != nil {
			return nil, err
		}
		res = append(res, v)
	}
	return res, nil
}

func ToUint64Slice(str string) ([]uint64, error) {
	vals := strings.Split(str, ",")
	res := make([]uint64, 0, len(vals))
	for _, val := range vals {
		v, err := ToUint64(val)
		if err != nil {
			return nil, err
		}
		res = append(res, v)
	}
	return res, nil
}

func ToUint32Slice(str string) ([]uint32, error) {
	vals := strings.Split(str, ",")
	res := make([]uint32, 0, len(vals))
	for _, val := range vals {
		v, err := ToUint32(val)
		if err != nil {
			return nil, err
		}
		res = append(res, v)
	}
	return res, nil
}

func ToUint16Slice(str string) ([]uint16, error) {
	vals := strings.Split(str, ",")
	res := make([]uint16, 0, len(vals))
	for _, val := range vals {
		v, err := ToUint16(val)
		if err != nil {
			return nil, err
		}
		res = append(res, v)
	}
	return res, nil
}

func ToUint8Slice(str string) ([]uint8, error) {
	vals := strings.Split(str, ",")
	res := make([]uint8, 0, len(vals))
	for _, val := range vals {
		v, err := ToUint8(val)
		if err != nil {
			return nil, err
		}
		res = append(res, v)
	}
	return res, nil
}

func ToFloat64Slice(str string) ([]float64, error) {
	vals := strings.Split(str, ",")
	res := make([]float64, 0, len(vals))
	for _, val := range vals {
		v, err := ToFloat64(val)
		if err != nil {
			return nil, err
		}
		res = append(res, v)
	}
	return res, nil
}

func ToFloat32Slice(str string) ([]float32, error) {
	vals := strings.Split(str, ",")
	res := make([]float32, 0, len(vals))
	for _, val := range vals {
		v, err := ToFloat32(val)
		if err != nil {
			return nil, err
		}
		res = append(res, v)
	}
	return res, nil
}

func ToDurationSlice(str string) ([]time.Duration, error) {
	vals := strings.Split(str, ",")
	res := make([]time.Duration, 0, len(vals))
	for _, val := range vals {
		v, err := ToDuration(val)
		if err != nil {
			return nil, err
		}
		res = append(res, v)
	}
	return res, nil
}

func ToTimeSlice(str string) ([]time.Time, error) {
	vals := strings.Split(str, ",")
	res := make([]time.Time, 0, len(vals))
	for _, val := range vals {
		v, err := ToTime(val)
		if err != nil {
			return nil, err
		}
		res = append(res, v)
	}
	return res, nil
}

func ToIPSlice(str string) ([]net.IP, error) {
	vals := strings.Split(str, ",")
	res := make([]net.IP, 0, len(vals))
	for _, val := range vals {
		v, err := ToIP(val)
		if err != nil {
			return nil, err
		}
		res = append(res, v)
	}
	return res, nil
}
