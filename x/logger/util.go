package logger

import (
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
)

// GetPackageFile gets package/file.go style return
func GetPackageFile(s string) string {
	fileIndex := strings.LastIndex(s, "/")
	packageIndex := strings.LastIndex(s[:fileIndex], "/")
	atIndex := strings.LastIndex(s[packageIndex+1:fileIndex], "@")
	if atIndex == -1 {
		return s[packageIndex+1:]
	}
	return s[packageIndex+1:packageIndex+atIndex+1] + "" + s[fileIndex:]
}

// ToString convert some type to string
func ToString(i interface{}) (string, error) {
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
	case []byte:
		s = string(v)
	case []rune:
		s = string(v)
	default:
		s = fmt.Sprint(v)
	}
	return s, nil
}

// LastPart splits s with sep, and get last piece
func LastPart(s string, sep string) string {
	lastIndex := strings.LastIndex(s, sep)
	if lastIndex < 0 {
		return s
	}
	return s[lastIndex+len(sep):]
}

// MustString calls ToString, panics if error
func MustString(i interface{}) string {
	s, err := ToString(i)
	if err != nil {
		panic(err)
	}
	return s
}

func parseLevel(level string) logrus.Level {
	switch strings.ToLower(level) {
	case "panic":
		return logrus.PanicLevel
	case "fatal":
		return logrus.FatalLevel
	case "error":
		return logrus.ErrorLevel
	case "warn", "warning":
		return logrus.WarnLevel
	case "info", "print":
		return logrus.InfoLevel
	case "debug":
		return logrus.DebugLevel
	case "trace":
		return logrus.TraceLevel
	default:
		panic(fmt.Errorf("unknown type of log level %s", level))
	}
}
