package strs

import (
	"fmt"
	"math"
	"strings"
)

func ParseIntWithCharMap(s string, charMap string, radix int) (int64, error) {
	if len(charMap) != radix {
		return 0, fmt.Errorf("错误的字符表")
	}

	var out float64
	for idx := range s {
		i := strings.IndexByte(charMap, s[idx])
		if i == -1 {
			return 0, fmt.Errorf("字符%s不在字符表内", string(s[idx]))
		}
		out += float64(i) * math.Pow(float64(radix), float64(idx))
	}

	return int64(out), nil
}

func FormatIntWithCharMap(v int64, charMap string, radix int) (string, error) {
	if len(charMap) != radix {
		return "", fmt.Errorf("错误的字符表")
	}

	var s string
	var left = v
	for left != 0 {
		i := int(left % int64(radix))
		s = charMap[i:i+1] + s
		left = left / int64(radix)
	}
	return s, nil
}

func FormatInt(v int64, radix int) string {
	s, _ := FormatIntWithCharMap(v, "ABCDEFGHIJKLMNOPQRSTUVWXYZ234567", radix)
	return s
}

func ParseInt(s string, radix int)  (int64, error) {
	return ParseIntWithCharMap(s, "ABCDEFGHIJKLMNOPQRSTUVWXYZ234567", radix)
}

