package dealstring

import (
	"fmt"
	"net"
	"reflect"
	"strconv"
	"time"
)

func ToInterface(str string, v interface{}) error {
	switch v.(type) {
	case *string:
		*v.(*string) = str
		return nil
	case *bool:
		val, err := ToBool(str)
		if err != nil {
			return err
		}
		*v.(*bool) = val
		return nil
	case *int:
		val, err := ToInt(str)
		if err != nil {
			return err
		}
		*v.(*int) = val
		return nil
	case *uint:
		val, err := ToUint(str)
		if err != nil {
			return err
		}
		*v.(*uint) = val
		return nil
	case *int64:
		val, err := ToInt64(str)
		if err != nil {
			return err
		}
		*v.(*int64) = val
		return nil
	case *int32:
		val, err := ToInt32(str)
		if err != nil {
			return err
		}
		*v.(*int32) = val
		return nil
	case *int16:
		val, err := ToInt16(str)
		if err != nil {
			return err
		}
		*v.(*int16) = val
		return nil
	case *int8:
		val, err := ToInt8(str)
		if err != nil {
			return err
		}
		*v.(*int8) = val
		return nil
	case *uint64:
		val, err := ToUint64(str)
		if err != nil {
			return err
		}
		*v.(*uint64) = val
		return nil
	case *uint32:
		val, err := ToUint32(str)
		if err != nil {
			return err
		}
		*v.(*uint32) = val
		return nil
	case *uint16:
		val, err := ToUint16(str)
		if err != nil {
			return err
		}
		*v.(*uint16) = val
		return nil
	case *uint8:
		val, err := ToUint8(str)
		if err != nil {
			return err
		}
		*v.(*uint8) = val
		return nil
	case *float64:
		val, err := ToFloat64(str)
		if err != nil {
			return err
		}
		*v.(*float64) = val
		return nil
	case *float32:
		val, err := ToFloat32(str)
		if err != nil {
			return err
		}
		*v.(*float32) = val
		return nil
	case *time.Duration:
		val, err := ToDuration(str)
		if err != nil {
			return err
		}
		*v.(*time.Duration) = val
		return nil
	case *time.Time:
		val, err := ToTime(str)
		if err != nil {
			return err
		}
		*v.(*time.Time) = val
		return nil
	case *net.IP:
		val, err := ToIP(str)
		if err != nil {
			return err
		}
		*v.(*net.IP) = val
		return nil
	case *[]string:
		val, err := ToStringSlice(str)
		if err != nil {
			return err
		}
		*v.(*[]string) = val
		return nil
	case *[]bool:
		val, err := ToBoolSlice(str)
		if err != nil {
			return err
		}
		*v.(*[]bool) = val
		return nil
	case *[]int:
		val, err := ToIntSlice(str)
		if err != nil {
			return err
		}
		*v.(*[]int) = val
		return nil
	case *[]uint:
		val, err := ToUintSlice(str)
		if err != nil {
			return err
		}
		*v.(*[]uint) = val
		return nil
	case *[]int64:
		val, err := ToInt64Slice(str)
		if err != nil {
			return err
		}
		*v.(*[]int64) = val
		return nil
	case *[]int32:
		val, err := ToInt32Slice(str)
		if err != nil {
			return err
		}
		*v.(*[]int32) = val
		return nil
	case *[]int16:
		val, err := ToInt16Slice(str)
		if err != nil {
			return err
		}
		*v.(*[]int16) = val
		return nil
	case *[]int8:
		val, err := ToInt8Slice(str)
		if err != nil {
			return err
		}
		*v.(*[]int8) = val
		return nil
	case *[]uint64:
		val, err := ToUint64Slice(str)
		if err != nil {
			return err
		}
		*v.(*[]uint64) = val
		return nil
	case *[]uint32:
		val, err := ToUint32Slice(str)
		if err != nil {
			return err
		}
		*v.(*[]uint32) = val
		return nil
	case *[]uint16:
		val, err := ToUint16Slice(str)
		if err != nil {
			return err
		}
		*v.(*[]uint16) = val
		return nil
	case *[]uint8:
		val, err := ToUint8Slice(str)
		if err != nil {
			return err
		}
		*v.(*[]uint8) = val
		return nil
	case *[]float64:
		val, err := ToFloat64Slice(str)
		if err != nil {
			return err
		}
		*v.(*[]float64) = val
		return nil
	case *[]float32:
		val, err := ToFloat32Slice(str)
		if err != nil {
			return err
		}
		*v.(*[]float32) = val
		return nil
	case *[]time.Duration:
		val, err := ToDurationSlice(str)
		if err != nil {
			return err
		}
		*v.(*[]time.Duration) = val
		return nil
	case *[]time.Time:
		val, err := ToTimeSlice(str)
		if err != nil {
			return err
		}
		*v.(*[]time.Time) = val
		return nil
	case *[]net.IP:
		val, err := ToIPSlice(str)
		if err != nil {
			return err
		}
		*v.(*[]net.IP) = val
		return nil
	default:
		return fmt.Errorf("unsupport type [%v]", reflect.TypeOf(v))
	}
}

func SetValue(v reflect.Value, str string) error {
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	switch v.Interface().(type) {
	case string:
		v.Set(reflect.ValueOf(str))
		return nil
	case bool:
		val, err := ToBool(str)
		if err != nil {
			return err
		}
		v.Set(reflect.ValueOf(val))
		return nil
	case int:
		val, err := ToInt(str)
		if err != nil {
			return err
		}
		v.Set(reflect.ValueOf(val))
		return nil
	case uint:
		val, err := ToUint(str)
		if err != nil {
			return err
		}
		v.Set(reflect.ValueOf(val))
		return nil
	case int64:
		val, err := ToInt64(str)
		if err != nil {
			return err
		}
		v.Set(reflect.ValueOf(val))
		return nil
	case int32:
		val, err := ToInt32(str)
		if err != nil {
			return err
		}
		v.Set(reflect.ValueOf(val))
		return nil
	case int16:
		val, err := ToInt16(str)
		if err != nil {
			return err
		}
		v.Set(reflect.ValueOf(val))
		return nil
	case int8:
		val, err := ToInt8(str)
		if err != nil {
			return err
		}
		v.Set(reflect.ValueOf(val))
		return nil
	case uint64:
		val, err := ToUint64(str)
		if err != nil {
			return err
		}
		v.Set(reflect.ValueOf(val))
		return nil
	case uint32:
		val, err := ToUint32(str)
		if err != nil {
			return err
		}
		v.Set(reflect.ValueOf(val))
		return nil
	case uint16:
		val, err := ToUint16(str)
		if err != nil {
			return err
		}
		v.Set(reflect.ValueOf(val))
		return nil
	case uint8:
		val, err := ToUint8(str)
		if err != nil {
			return err
		}
		v.Set(reflect.ValueOf(val))
		return nil
	case float64:
		val, err := ToFloat64(str)
		if err != nil {
			return err
		}
		v.Set(reflect.ValueOf(val))
		return nil
	case float32:
		val, err := ToFloat32(str)
		if err != nil {
			return err
		}
		v.Set(reflect.ValueOf(val))
		return nil
	case time.Duration:
		val, err := ToDuration(str)
		if err != nil {
			return err
		}
		v.Set(reflect.ValueOf(val))
		return nil
	case time.Time:
		val, err := ToTime(str)
		if err != nil {
			return err
		}
		v.Set(reflect.ValueOf(val))
		return nil
	case net.IP:
		val, err := ToIP(str)
		if err != nil {
			return err
		}
		v.Set(reflect.ValueOf(val))
		return nil
	case []string:
		val, err := ToStringSlice(str)
		if err != nil {
			return err
		}
		v.Set(reflect.ValueOf(val))
		return nil
	case []bool:
		val, err := ToBoolSlice(str)
		if err != nil {
			return err
		}
		v.Set(reflect.ValueOf(val))
		return nil
	case []int:
		val, err := ToIntSlice(str)
		if err != nil {
			return err
		}
		v.Set(reflect.ValueOf(val))
		return nil
	case []uint:
		val, err := ToUintSlice(str)
		if err != nil {
			return err
		}
		v.Set(reflect.ValueOf(val))
		return nil
	case []int64:
		val, err := ToInt64Slice(str)
		if err != nil {
			return err
		}
		v.Set(reflect.ValueOf(val))
		return nil
	case []int32:
		val, err := ToInt32Slice(str)
		if err != nil {
			return err
		}
		v.Set(reflect.ValueOf(val))
		return nil
	case []int16:
		val, err := ToInt16Slice(str)
		if err != nil {
			return err
		}
		v.Set(reflect.ValueOf(val))
		return nil
	case []int8:
		val, err := ToInt8Slice(str)
		if err != nil {
			return err
		}
		v.Set(reflect.ValueOf(val))
		return nil
	case []uint64:
		val, err := ToUint64Slice(str)
		if err != nil {
			return err
		}
		v.Set(reflect.ValueOf(val))
		return nil
	case []uint32:
		val, err := ToUint32Slice(str)
		if err != nil {
			return err
		}
		v.Set(reflect.ValueOf(val))
		return nil
	case []uint16:
		val, err := ToUint16Slice(str)
		if err != nil {
			return err
		}
		v.Set(reflect.ValueOf(val))
		return nil
	case []uint8:
		val, err := ToUint8Slice(str)
		if err != nil {
			return err
		}
		v.Set(reflect.ValueOf(val))
		return nil
	case []float64:
		val, err := ToFloat64Slice(str)
		if err != nil {
			return err
		}
		v.Set(reflect.ValueOf(val))
		return nil
	case []float32:
		val, err := ToFloat32Slice(str)
		if err != nil {
			return err
		}
		v.Set(reflect.ValueOf(val))
		return nil
	case []time.Duration:
		val, err := ToDurationSlice(str)
		if err != nil {
			return err
		}
		v.Set(reflect.ValueOf(val))
		return nil
	case []time.Time:
		val, err := ToTimeSlice(str)
		if err != nil {
			return err
		}
		v.Set(reflect.ValueOf(val))
		return nil
	case []net.IP:
		val, err := ToIPSlice(str)
		if err != nil {
			return err
		}
		v.Set(reflect.ValueOf(val))
		return nil
	default:
		return fmt.Errorf("unsupport type [%v]", v)
	}
}

func ToBool(str string) (bool, error) {
	return strconv.ParseBool(str)
}

func ToInt(str string) (int, error) {
	return strconv.Atoi(str)
}

func ToUint(str string) (uint, error) {
	i, err := strconv.ParseUint(str, 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(i), nil
}

func ToInt64(str string) (int64, error) {
	return strconv.ParseInt(str, 10, 64)
}

func ToInt32(str string) (int32, error) {
	i, err := strconv.ParseInt(str, 10, 32)
	if err != nil {
		return 0, err
	}
	return int32(i), nil
}

func ToInt16(str string) (int16, error) {
	i, err := strconv.ParseInt(str, 10, 16)
	if err != nil {
		return 0, err
	}
	return int16(i), nil
}

func ToInt8(str string) (int8, error) {
	i, err := strconv.ParseInt(str, 10, 8)
	if err != nil {
		return 0, err
	}
	return int8(i), nil
}

func ToUint64(str string) (uint64, error) {
	return strconv.ParseUint(str, 10, 64)
}

func ToUint32(str string) (uint32, error) {
	i, err := strconv.ParseUint(str, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint32(i), nil
}

func ToUint16(str string) (uint16, error) {
	i, err := strconv.ParseUint(str, 10, 16)
	if err != nil {
		return 0, err
	}
	return uint16(i), nil
}

func ToUint8(str string) (uint8, error) {
	i, err := strconv.ParseUint(str, 10, 8)
	if err != nil {
		return 0, err
	}
	return uint8(i), nil
}

func ToFloat64(str string) (float64, error) {
	return strconv.ParseFloat(str, 64)
}

func ToFloat32(str string) (float32, error) {
	f, err := strconv.ParseFloat(str, 32)
	if err != nil {
		return 0, err
	}
	return float32(f), nil
}

func ToDuration(str string) (time.Duration, error) {
	return time.ParseDuration(str)
}

func ToTime(str string) (time.Time, error) {
	if str == "now" {
		return time.Now(), nil
	} else if len(str) == 10 {
		return time.Parse("2006-01-02", str)
	} else if len(str) == 19 {
		if str[10] == ' ' {
			return time.Parse("2006-01-02 15:04:05", str)
		}
		return time.Parse("2006-01-02T15:04:05", str)
	}

	return time.Parse(time.RFC3339, str)
}

func ToIP(str string) (net.IP, error) {
	ip := net.ParseIP(str)
	if ip == nil {
		return nil, fmt.Errorf("invalid ip [%v]", str)
	}
	return ip, nil
}
