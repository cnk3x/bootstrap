package mux

import (
    "fmt"
    "go.shu.run/bootstrap/utils/namec"
    "net/http"
    "net/url"
    "reflect"
    "strconv"
    "strings"
    "time"
)

var _ R = H(nil)

type H map[string]interface{}

func (m H) WriteTo(c *C, w http.ResponseWriter) {
    c.JSON(m).WriteTo(c, w)
}

func (m H) Error() string {
    if len(m) > 0 {
        return fmt.Sprintf("%v", m["err"])
    }
    return ""
}

func (m H) IsError() bool {
    err, find := m["err"]
    return find && err != nil
}

func (m H) Pop(name string) interface{} {
    if v, find := m[name]; find {
        delete(m, name)
        return v
    }
    return nil
}

func (m H) Extend(values ...interface{}) H {
    for _, value := range values {
        if value == nil {
            continue
        }

        var refV reflect.Value

        switch v := value.(type) {
        case url.Values:
            for k, v := range v {
                if len(v) > 0 {
                    if len(v) > 1 {
                        _ = m.Set(k, v)
                    } else {
                        _ = m.Set(k, v[0])
                    }
                }
            }
            continue
        case map[string]interface{}:
            for k, v := range v {
                _ = m.Set(k, v)
            }
            continue
        case reflect.Value:
            refV = v
        default:
            refV = reflect.Indirect(reflect.ValueOf(value))
        }

        refT := refV.Type()
        switch refT.Kind() {
        case reflect.Map:
            iter := refV.MapRange()
            for iter.Next() {
                _ = m.Set(toString(iter.Key()), iter.Value())
            }
        case reflect.Struct:
            for i := 0; i < refT.NumField(); i++ {
                _ = m.Set(namec.SnakeCase(refT.Field(i).Name), refV.Field(i))
            }
        default:
            fmt.Printf("H.Extend -> 不支持: %v\n", value)
        }
    }

    return m
}

func (m H) Set(name string, value interface{}) H {
    if val := extendValue(value); val != nil {
        m[name] = val
    }
    return m
}

func (m H) Del(names ...string) H {
    for _, name := range names {
        delete(m, name)
    }
    return m
}

func extendValue(value interface{}) interface{} {
    if value == nil {
        return nil
    }

    refV, ok := value.(reflect.Value)
    if ok {
        if refV.IsZero() {
            return nil
        }
        value = refV.Interface()
    }

    if refV = reflect.Indirect(reflect.ValueOf(value)); refV.IsZero() {
        return nil
    }

    switch v := value.(type) {
    case time.Time, string, bool, float32, float64, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, BigInt:
        return v
    }

    refT := refV.Type()
    switch refT.Kind() {
    case reflect.Struct:
        m := H{}
        for i := 0; i < refT.NumField(); i++ {
            if val := extendValue(refV.Field(i)); val != nil {
                m[namec.SnakeCase(refT.Field(i).Name)] = val
            }
        }
        if len(m) == 0 {
            return nil
        }
        return m
    case reflect.Array, reflect.Slice:
        var array []interface{}
        for i := 0; i < refV.Len(); i++ {
            if val := extendValue(refV.Index(i)); val != nil {
                array = append(array, val)
            }
        }
        if len(array) == 0 {
            return nil
        }
        return array
    case reflect.Map:
        m := H{}
        iter := refV.MapRange()
        for iter.Next() {
            if val := extendValue(iter.Value()); val != nil {
                m[toString(iter.Key())] = val
            }
        }
        if len(m) == 0 {
            return nil
        }
        return m
    default:
        fmt.Printf("H.extendValue -> 不支持: %v\n", value)
        return nil
    }
}

func toString(value interface{}) string {
    switch v := value.(type) {
    case string:
        return v
    case reflect.Value:
        return toString(v.Interface())
    case time.Time:
        return v.Format(time.RFC3339)
    case bool:
        return strconv.FormatBool(v)
    case float32, float64:
        return strconv.FormatFloat(reflect.ValueOf(value).Float(), 'f', -1, 64)
    case int, int8, int16, int64, BigInt:
        return strconv.FormatInt(reflect.ValueOf(value).Int(), 10)
    case uint, uint16, uint32, uint64:
        return strconv.FormatUint(reflect.ValueOf(value).Uint(), 10)
    case []byte:
        return string(v)
    case []string:
        return strings.Join(v, "")
    case uint8:
        return string(v)
    case int32:
        return string(v)
    case fmt.Stringer:
        return v.String()
    case error:
        return v.Error()
    default:
        return fmt.Sprintf("%v", v)
    }
}
