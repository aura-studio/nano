package hashtree

import (
	"encoding/json"
	"fmt"
	"math/big"
	"reflect"
	"strconv"
	"unsafe"

	"github.com/lonng/nano/x/virtualtime"
	"github.com/shopspring/decimal"
)

// timeStringToStamp converts a string time to int time
func timeStringToStamp(s string) (int64, error) {
	timestamp, err := strconv.ParseInt(s, 10, 64)
	if err == nil {
		return timestamp, nil
	}

	tm, err := virtualtime.Parse(s)
	if err == nil {
		return tm.Unix(), nil
	}
	return 0, err
}

// timeStampToString converts a int time to string time
func timeStampToString(n int64) string {
	return virtualtime.Format(virtualtime.Time(n))
}

// getValueByPath gets values by path
func getValueByPath(value reflect.Value, path []string) (reflect.Value, bool) {
	for _, field := range path {
		value = value.FieldByName(field)
		if !value.IsValid() {
			return value, false
		}
	}
	return value, true
}

// getValuePtr gets unsafe pointer for value
func getValuePtr(value reflect.Value) unsafe.Pointer {
	return unsafe.Pointer(value.UnsafeAddr())
}

// getValueFieldPtr gets unsafe pointer for field value
func getValueFieldPtr(value reflect.Value, name string) unsafe.Pointer {
	if value.Kind() != reflect.Struct {
		panic(fmt.Errorf("GetFieldValuePtr value is not struct"))
	}
	return unsafe.Pointer(value.FieldByName(name).UnsafeAddr())
}

// setValueField sets field value for reflect.Value
func setValueField(value reflect.Value, name string, v interface{}) {
	if value.Kind() != reflect.Struct {
		panic(fmt.Errorf("SetFieldValue value is not struct"))
	}
	setValue(value.FieldByName(name), v)
}

// setValue sets value for reflect.value
func setValue(value reflect.Value, v interface{}) {
	switch v := v.(type) {
	case int:
		*(*int)(unsafe.Pointer(value.UnsafeAddr())) = v
	case int8:
		*(*int8)(unsafe.Pointer(value.UnsafeAddr())) = v
	case int16:
		*(*int16)(unsafe.Pointer(value.UnsafeAddr())) = v
	case int32:
		*(*int32)(unsafe.Pointer(value.UnsafeAddr())) = v
	case int64:
		*(*int64)(unsafe.Pointer(value.UnsafeAddr())) = v
	case uint8:
		*(*uint8)(unsafe.Pointer(value.UnsafeAddr())) = v
	case uint:
		*(*uint)(unsafe.Pointer(value.UnsafeAddr())) = v
	case uint16:
		*(*uint16)(unsafe.Pointer(value.UnsafeAddr())) = v
	case uint32:
		*(*uint32)(unsafe.Pointer(value.UnsafeAddr())) = v
	case uint64:
		*(*uint64)(unsafe.Pointer(value.UnsafeAddr())) = v
	case float32:
		*(*float32)(unsafe.Pointer(value.UnsafeAddr())) = v
	case float64:
		*(*float64)(unsafe.Pointer(value.UnsafeAddr())) = v
	case complex64:
		*(*complex64)(unsafe.Pointer(value.UnsafeAddr())) = v
	case complex128:
		*(*complex128)(unsafe.Pointer(value.UnsafeAddr())) = v
	case big.Int:
		*(*big.Int)(unsafe.Pointer(value.UnsafeAddr())) = v
	case big.Rat:
		*(*big.Rat)(unsafe.Pointer(value.UnsafeAddr())) = v
	case big.Float:
		*(*big.Float)(unsafe.Pointer(value.UnsafeAddr())) = v
	case uintptr:
		*(*uintptr)(unsafe.Pointer(value.UnsafeAddr())) = v
	case bool:
		*(*bool)(unsafe.Pointer(value.UnsafeAddr())) = v
	case string:
		*(*string)(unsafe.Pointer(value.UnsafeAddr())) = v
	case []int:
		*(*[]int)(unsafe.Pointer(value.UnsafeAddr())) = v
	case []int8:
		*(*[]int8)(unsafe.Pointer(value.UnsafeAddr())) = v
	case []int16:
		*(*[]int16)(unsafe.Pointer(value.UnsafeAddr())) = v
	case []int32: // cover []rune
		*(*[]int32)(unsafe.Pointer(value.UnsafeAddr())) = v
	case []int64:
		*(*[]int64)(unsafe.Pointer(value.UnsafeAddr())) = v
	case []uint:
		*(*[]uint)(unsafe.Pointer(value.UnsafeAddr())) = v
	case []uint8: // cover []byte
		*(*[]uint8)(unsafe.Pointer(value.UnsafeAddr())) = v
	case []uint16:
		*(*[]uint16)(unsafe.Pointer(value.UnsafeAddr())) = v
	case []uint32:
		*(*[]uint32)(unsafe.Pointer(value.UnsafeAddr())) = v
	case []uint64:
		*(*[]uint64)(unsafe.Pointer(value.UnsafeAddr())) = v
	case []float32:
		*(*[]float32)(unsafe.Pointer(value.UnsafeAddr())) = v
	case []float64:
		*(*[]float64)(unsafe.Pointer(value.UnsafeAddr())) = v
	case []string:
		*(*[]string)(unsafe.Pointer(value.UnsafeAddr())) = v
	case *int:
		*(**int)(unsafe.Pointer(value.UnsafeAddr())) = v
	case *int8:
		*(**int8)(unsafe.Pointer(value.UnsafeAddr())) = v
	case *int16:
		*(**int16)(unsafe.Pointer(value.UnsafeAddr())) = v
	case *int32:
		*(**int32)(unsafe.Pointer(value.UnsafeAddr())) = v
	case *int64:
		*(**int64)(unsafe.Pointer(value.UnsafeAddr())) = v
	case *uint:
		*(**uint)(unsafe.Pointer(value.UnsafeAddr())) = v
	case *uint8:
		*(**uint8)(unsafe.Pointer(value.UnsafeAddr())) = v
	case *uint16:
		*(**uint16)(unsafe.Pointer(value.UnsafeAddr())) = v
	case *uint32:
		*(**uint32)(unsafe.Pointer(value.UnsafeAddr())) = v
	case *uint64:
		*(**uint64)(unsafe.Pointer(value.UnsafeAddr())) = v
	case *float32:
		*(**float32)(unsafe.Pointer(value.UnsafeAddr())) = v
	case *float64:
		*(**float64)(unsafe.Pointer(value.UnsafeAddr())) = v
	case *complex64:
		*(**complex64)(unsafe.Pointer(value.UnsafeAddr())) = v
	case *complex128:
		*(**complex128)(unsafe.Pointer(value.UnsafeAddr())) = v
	case *big.Int:
		*(**big.Int)(unsafe.Pointer(value.UnsafeAddr())) = v
	case *big.Rat:
		*(**big.Rat)(unsafe.Pointer(value.UnsafeAddr())) = v
	case *big.Float:
		*(**big.Float)(unsafe.Pointer(value.UnsafeAddr())) = v
	case *uintptr:
		*(**uintptr)(unsafe.Pointer(value.UnsafeAddr())) = v
	case *bool:
		*(**bool)(unsafe.Pointer(value.UnsafeAddr())) = v
	case *string:
		*(**string)(unsafe.Pointer(value.UnsafeAddr())) = v
	default:
		panic(fmt.Errorf("SetValue does not get suitable type %+v, %+v",
			v, reflect.TypeOf(v)))
	}
}

// interfaceToString convert sometype to string
func interfaceToString(i interface{}) (string, error) {
	var s string
	switch v := i.(type) {
	case int:
		s = strconv.FormatInt(int64(v), 10)
	case int8:
		s = strconv.FormatInt(int64(v), 10)
	case int16:
		s = strconv.FormatInt(int64(v), 10)
	case int32:
		s = strconv.FormatInt(int64(v), 10)
	case int64:
		s = strconv.FormatInt(int64(v), 10)
	case uint:
		s = strconv.FormatUint(uint64(v), 10)
	case uint8:
		s = strconv.FormatUint(uint64(v), 10)
	case uint16:
		s = strconv.FormatUint(uint64(v), 10)
	case uint32:
		s = strconv.FormatUint(uint64(v), 10)
	case uint64:
		s = strconv.FormatUint(uint64(v), 10)
	case float32:
		// Use decimal to fix precision issue, FormatFloat is instable.
		s = decimal.NewFromFloat32(v).String()
	case float64:
		// Use decimal to fix precision issue, FormatFloat is instable.
		s = decimal.NewFromFloat(v).String()
	case complex64:
		// New version of strconv required
		// s = strconv.FormatComplex(v, 10)

		// Use fmt.Sprint instead
		s = fmt.Sprint(v)
	case complex128:
		// New version of strconv required
		// s = strconv.FormatComplex(v, 10)

		// Use fmt.Sprint instead
		s = fmt.Sprint(v)
	case big.Int:
		s = v.String()
	case big.Rat:
		s = v.String()
	case big.Float:
		s = v.String()
	case *big.Int:
		s = v.String()
	case *big.Rat:
		s = v.String()
	case *big.Float:
		s = v.String()
	case uintptr:
		s = fmt.Sprint(v)
	case bool:
		s = strconv.FormatBool(v)
	case string:
		s = string(v)
	case []int:
		b, err := json.Marshal(v)
		if err != nil {
			return "", fmt.Errorf("json marshal failed:%+v, %+v", v, reflect.TypeOf(v))
		}
		s = string(b)
	case []int8:
		b, err := json.Marshal(v)
		if err != nil {
			return "", fmt.Errorf("json marshal failed:%+v, %+v", v, reflect.TypeOf(v))
		}
		s = string(b)
	case []int16:
		b, err := json.Marshal(v)
		if err != nil {
			return "", fmt.Errorf("json marshal failed:%+v, %+v", v, reflect.TypeOf(v))
		}
		s = string(b)
	case []int32: // cover rune
		b, err := json.Marshal(v)
		if err != nil {
			return "", fmt.Errorf("json marshal failed:%+v, %+v", v, reflect.TypeOf(v))
		}
		s = string(b)
	case []int64:
		b, err := json.Marshal(v)
		if err != nil {
			return "", fmt.Errorf("json marshal failed:%+v, %+v", v, reflect.TypeOf(v))
		}
		s = string(b)
	case []uint:
		b, err := json.Marshal(v)
		if err != nil {
			return "", fmt.Errorf("json marshal failed:%+v, %+v", v, reflect.TypeOf(v))
		}
		s = string(b)
	case []uint8:
		b, err := json.Marshal(v)
		if err != nil {
			return "", fmt.Errorf("json marshal failed:%+v, %+v", v, reflect.TypeOf(v))
		}
		s = string(b)
	case []uint16:
		b, err := json.Marshal(v)
		if err != nil {
			return "", fmt.Errorf("json marshal failed:%+v, %+v", v, reflect.TypeOf(v))
		}
		s = string(b)
	case []uint32:
		b, err := json.Marshal(v)
		if err != nil {
			return "", fmt.Errorf("json marshal failed:%+v, %+v", v, reflect.TypeOf(v))
		}
		s = string(b)
	case []uint64:
		b, err := json.Marshal(v)
		if err != nil {
			return "", fmt.Errorf("json marshal failed:%+v, %+v", v, reflect.TypeOf(v))
		}
		s = string(b)
	case []float32:
		b, err := json.Marshal(v)
		if err != nil {
			return "", fmt.Errorf("json marshal failed:%+v, %+v", v, reflect.TypeOf(v))
		}
		s = string(b)
	case []float64:
		b, err := json.Marshal(v)
		if err != nil {
			return "", fmt.Errorf("json marshal failed:%+v, %+v", v, reflect.TypeOf(v))
		}
		s = string(b)
	case []string:
		b, err := json.Marshal(v)
		if err != nil {
			return "", fmt.Errorf("json marshal failed:%+v, %+v", v, reflect.TypeOf(v))
		}
		s = string(b)
	default:
		return "", fmt.Errorf("invalid type to string:%+v, %+v", v, reflect.TypeOf(v))
	}
	return s, nil
}

// stringToValue convert string to value
func stringToValue(s string, v reflect.Value) error {
	switch v.Type().String() {
	case "int":
		n, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return err
		}
		setValue(v, int(n))
	case "int8":
		n, err := strconv.ParseInt(s, 10, 8)
		if err != nil {
			return err
		}
		setValue(v, int8(n))
	case "int16":
		n, err := strconv.ParseInt(s, 10, 16)
		if err != nil {
			return err
		}
		setValue(v, int16(n))
	case "int32":
		n, err := strconv.ParseInt(s, 10, 32)
		if err != nil {
			return err
		}
		setValue(v, int32(n))
	case "int64":
		n, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return err
		}
		setValue(v, int64(n))
	case "uint":
		n, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			return err
		}
		setValue(v, uint(n))
	case "uint8":
		n, err := strconv.ParseUint(s, 10, 8)
		if err != nil {
			return err
		}
		setValue(v, uint8(n))
	case "uint16":
		n, err := strconv.ParseUint(s, 10, 16)
		if err != nil {
			return err
		}
		setValue(v, uint16(n))
	case "uint32":
		n, err := strconv.ParseUint(s, 10, 32)
		if err != nil {
			return err
		}
		setValue(v, uint32(n))
	case "uint64":
		n, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			return err
		}
		setValue(v, uint64(n))
	case "float32":
		d, err := decimal.NewFromString(s)
		if err != nil {
			return err
		}
		f, _ := d.Float64()
		setValue(v, float32(f))
	case "float64":
		d, err := decimal.NewFromString(s)
		if err != nil {
			return err
		}
		f, _ := d.Float64()
		setValue(v, float64(f))
	case "complex64":
		var n complex64
		_, err := fmt.Sscan(s, &n)
		if err != nil {
			return err
		}
		setValue(v, n)
	case "complex128":
		var n complex128
		_, err := fmt.Sscan(s, &n)
		if err != nil {
			return err
		}
		setValue(v, n)
	case "big.Int":
		b, ok := new(big.Int).SetString(s, 10)
		if !ok {
			return fmt.Errorf("big.Int set string %s failed", s)
		}
		setValue(v, *b)
	case "big.Rat":
		b, ok := new(big.Rat).SetString(s)
		if !ok {
			return fmt.Errorf("big.Rat set string %s failed", s)
		}
		setValue(v, *b)
	case "big.Float":
		b, ok := new(big.Float).SetString(s)
		if !ok {
			return fmt.Errorf("big.Float set string %s failed", s)
		}
		setValue(v, *b)
	case "bool":
		t, err := strconv.ParseBool(s)
		if err != nil {
			return err
		}
		setValue(v, bool(t))
	case "string":
		setValue(v, s)
	case "[]int":
		n := make([]int, 0)
		err := json.Unmarshal([]byte(s), &n)
		if err != nil {
			return fmt.Errorf("json unmarshal failed:%+v, %+v", s,
				v.Type().String())
		}
		setValue(v, n)
	case "[]int8":
		n := make([]int8, 0)
		err := json.Unmarshal([]byte(s), &n)
		if err != nil {
			return fmt.Errorf("json unmarshal failed:%+v, %+v", s,
				v.Type().String())
		}
		setValue(v, n)
	case "[]int16":
		n := make([]int16, 0)
		err := json.Unmarshal([]byte(s), &n)
		if err != nil {
			return fmt.Errorf("json unmarshal failed:%+v, %+v", s,
				v.Type().String())
		}
		setValue(v, n)
	case "[]int32":
		n := make([]int32, 0)
		err := json.Unmarshal([]byte(s), &n)
		if err != nil {
			return fmt.Errorf("json unmarshal failed:%+v, %+v", s,
				v.Type().String())
		}
		setValue(v, n)
	case "[]int64":
		n := make([]int64, 0)
		err := json.Unmarshal([]byte(s), &n)
		if err != nil {
			return fmt.Errorf("json unmarshal failed:%+v, %+v", s,
				v.Type().String())
		}
		setValue(v, n)
	case "[]uint":
		n := make([]uint, 0)
		err := json.Unmarshal([]byte(s), &n)
		if err != nil {
			return fmt.Errorf("json unmarshal failed:%+v, %+v", s,
				v.Type().String())
		}
		setValue(v, n)
	case "[]uint8":
		n := make([]uint8, 0)
		err := json.Unmarshal([]byte(s), &n)
		if err != nil {
			return fmt.Errorf("json unmarshal failed:%+v, %+v", s,
				v.Type().String())
		}
		setValue(v, n)
	case "[]uint16":
		n := make([]uint16, 0)
		err := json.Unmarshal([]byte(s), &n)
		if err != nil {
			return fmt.Errorf("json unmarshal failed:%+v, %+v", s,
				v.Type().String())
		}
		setValue(v, n)
	case "[]uint32":
		n := make([]uint32, 0)
		err := json.Unmarshal([]byte(s), &n)
		if err != nil {
			return fmt.Errorf("json unmarshal failed:%+v, %+v", s,
				v.Type().String())
		}
		setValue(v, n)
	case "[]uint64":
		n := make([]uint64, 0)
		err := json.Unmarshal([]byte(s), &n)
		if err != nil {
			return fmt.Errorf("json unmarshal failed:%+v, %+v", s,
				v.Type().String())
		}
		setValue(v, n)
	case "[]float32":
		n := make([]float32, 0)
		err := json.Unmarshal([]byte(s), &n)
		if err != nil {
			return fmt.Errorf("json unmarshal failed:%+v, %+v", s,
				v.Type().String())
		}
		setValue(v, n)
	case "[]float64":
		n := make([]float64, 0)
		err := json.Unmarshal([]byte(s), &n)
		if err != nil {
			return fmt.Errorf("json unmarshal failed:%+v, %+v", s,
				v.Type().String())
		}
		setValue(v, n)
	case "[]string":
		n := make([]string, 0)
		err := json.Unmarshal([]byte(s), &n)
		if err != nil {
			return fmt.Errorf("json unmarshal failed:%+v, %+v", s,
				v.Type().String())
		}
		setValue(v, n)
	default:
		return fmt.Errorf("invalid type from string:%+v, %+v", s,
			v.Type().String())
	}
	return nil
}
