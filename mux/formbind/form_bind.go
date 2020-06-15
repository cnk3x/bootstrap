package formbind

import (
	"fmt"
	"net/url"
	"reflect"
	"strconv"
	"strings"

	"go.shu.run/bootstrap/utils/json"
	"go.shu.run/bootstrap/utils/namec"
)

func Bind(form url.Values, i interface{}) error {
	if len(form) == 0 || i == nil {
		return nil
	}

	errs := make(ErrorMap)
	refV := reflect.Indirect(reflect.ValueOf(i))
	refT := refV.Type()

	for i := 0; i < refT.NumField(); i++ {
		var (
			fieldS = refT.Field(i) //StaticField
			fieldV = refV.Field(i) //Value
			fieldT = fieldS.Type   //Type
		)

		if fieldT.Kind() == reflect.Ptr {
			fieldT = fieldT.Elem()
		}

		nameTag := lookUpTag(fieldS, "form", "json")
		val := form[nameTag]
		if len(val) == 0 {
			nameTag = fieldS.Name
			val = form[nameTag]
		}

		if len(val) > 0 {
			if err := fieldSet(fieldT, fieldV, val); err != nil {
				errs.Add(fieldS.Name, err.Error())
			}
		}
	}

	return errs.Err()
}

func fieldSet(fieldT reflect.Type, fieldV reflect.Value, val []string) (err error) {
	defer func() {
		if re := recover(); re != nil {
			err = fmt.Errorf("%v", re)
		}
	}()

	if fieldT.Kind() != reflect.Slice {
		return fieldSet2(fieldT, fieldV, val[0])
	}

	var (
		distV = reflect.MakeSlice(fieldT, 0, len(val))
		itemT = fieldT.Elem()
		itemP = false
	)

	if itemT.Kind() == reflect.Ptr {
		itemT = itemT.Elem()
		itemP = true
	}

	var errs = make(ErrorMap)
	for _, v := range val {
		itemV := reflect.New(itemT)
		if ex := fieldSet2(itemT, itemV.Elem(), v); ex != nil {
			errs.Add(v, ex.Error())
		}

		if itemP {
			distV = reflect.Append(distV, itemV)
		} else {
			distV = reflect.Append(distV, itemV.Elem())
		}
	}

	if fieldV.Kind() == reflect.Ptr {
		pV := reflect.New(fieldT)
		pV.Elem().Set(distV)
		distV = pV
	}
	fieldV.Set(distV)

	return errs.Err()
}

func fieldSet2(fieldT reflect.Type, fieldV reflect.Value, val string) (err error) {
	defer func() {
		if re := recover(); re != nil {
			err = fmt.Errorf("%v", re)
		}
	}()

	if fieldV.Kind() == reflect.Ptr {
		tempV := reflect.New(fieldT)
		if err := fieldSet3(fieldT, tempV.Elem(), val); err != nil {
			return err
		}
		fieldV.Set(tempV)
		return nil
	}

	return fieldSet3(fieldT, fieldV, val)
}

func fieldSet3(fieldT reflect.Type, fieldV reflect.Value, val string) (err error) {
	defer func() {
		if re := recover(); re != nil {
			err = fmt.Errorf("%v", re)
		}
	}()

	switch fieldT.Kind() {
	case reflect.String:
		fieldV.SetString(val)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if v, ex := strconv.ParseInt(val, 10, 64); ex == nil {
			fieldV.SetInt(v)
		} else {
			return ex
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if v, ex := strconv.ParseUint(val, 10, 64); ex == nil {
			fieldV.SetUint(v)
		} else {
			return ex
		}
	case reflect.Bool:
		if v, ex := strconv.ParseBool(val); ex == nil {
			fieldV.SetBool(v)
		} else {
			return ex
		}
	case reflect.Float32, reflect.Float64:
		if v, ex := strconv.ParseFloat(val, 64); ex == nil {
			fieldV.SetFloat(v)
		} else {
			return ex
		}
	default:
		return fmt.Errorf("不支持的类型: %s", fieldT)
	}

	return nil
}

func lookUpTag(field reflect.StructField, names ...string) string {
	for _, name := range names {
		if s, ok := field.Tag.Lookup(name); ok {
			if i := strings.IndexRune(s, ','); i != -1 {
				s = s[:i]
			}
			if s != "" {
				return strings.TrimSpace(s)
			}
		}
	}
	return namec.SnakeCase(field.Name)
}

type ErrorMap map[string]interface{}

func (em ErrorMap) Error() string {
	v, _ := json.MarshalIndent(em, "", "  ")
	return string(v)
}

func (em ErrorMap) Add(name string, err string) {
	em[name] = err
}

func (em ErrorMap) Err() error {
	if len(em) > 0 {
		return em
	}
	return nil
}
