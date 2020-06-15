package define

import (
	"fmt"
	"regexp"
	"strconv"
)

func Range(min, max int64, message string) VFunc {
	return func(value string) error {
		if value == "" {
			return nil
		}
		i, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return err
		}
		if min <= i && i < max {
			return nil
		}
		return fmt.Errorf(message)
	}
}

func Int(message string) VFunc {
	return func(value string) error {
		if value == "" {
			return nil
		}
		_, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return fmt.Errorf(message)
		}
		return nil
	}
}

func Float(message string) VFunc {
	return func(value string) error {
		if value == "" {
			return nil
		}
		_, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fmt.Errorf(message)
		}
		return nil
	}
}

func Bool(message string) VFunc {
	return func(value string) error {
		if value == "" {
			return nil
		}
		_, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf(message)
		}
		return nil
	}
}

func Regex(pattern string, message string) VFunc {
	return func(value string) error {
		if value == "" {
			return nil
		}
		r, err := regexp.Compile(pattern)
		if err != nil {
			return err
		}
		if !r.MatchString(value) {
			return fmt.Errorf(message)
		}
		return nil
	}
}

func Length(min, max int, message string) VFunc {
	err := fmt.Errorf(message)
	return func(value string) error {
		l := len(value)
		if (min > 0 && l < min) || (max > 0 && l > max) {
			return err
		}
		return nil
	}
}
