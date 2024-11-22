package mapvalue

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	uuid "github.com/satori/go.uuid"
)

func IsString(sourceMap map[string]interface{}, key string) error {
	// Check that key exists in map
	val, ok := sourceMap[key]
	if !ok {
		return fmt.Errorf("%s must be provided", key)
	}

	// Check that value for key is a string
	_, ok = val.(string)
	if !ok {
		return fmt.Errorf("%s found with wrong type: expected string", key)
	}
	return nil
}

func String(sourceMap map[string]interface{}, key string) string {
	valAsString, ok := sourceMap[key].(string)
	if !ok {
		return ""
	}
	return valAsString
}
func Bool(sourceMap map[string]interface{}, key string, deflt bool) bool {
	valAsBool, ok := sourceMap[key].(bool)
	if !ok {
		return deflt
	}
	return valAsBool
}
func Integer(sourceMap map[string]interface{}, key string) int {
	valAsFloat, ok := sourceMap[key].(float64)
	if !ok {
		return 0
	}
	return int(valAsFloat)
}

func IsMapSlice(sourceMap map[string]interface{}, key string) error {
	// Check that key exists in map
	val, ok := sourceMap[key]
	if !ok {
		return fmt.Errorf("%s must be provided", key)
	}

	// Check that value of key is a slice
	valAsSlice, ok := val.([]interface{})
	if !ok {
		return fmt.Errorf("%s found with wrong type: expected JSON array", key)
	}

	// Check that the value of each element is a map
	for index, nestedValue := range valAsSlice {
		_, ok = nestedValue.(map[string]interface{})
		if !ok {
			return fmt.Errorf("Object at index %d in array found with wrong type: expected JSON object", index)
		}
	}

	return nil
}

func MapSlice(sourceMap map[string]interface{}, key string) []map[string]interface{} {
	var valAsMapSlice []map[string]interface{}
	// The type switch statement here complicates things, but it allows us to handle unmarshaled []interface{} slices
	// as well as already strongly-typed []map[string]interface{}, without worrying about the underlying abstraction.
	switch sourceMap[key].(type) {
	case []interface{}:
		// If sourceMap has the type []interface{}, then cast each value into a more strongly typed container.
		temp := sourceMap[key].([]interface{})
		valAsMapSlice = make([]map[string]interface{}, len(temp))
		for i, v := range temp {
			castVal, ok := v.(map[string]interface{})
			if !ok {
				return []map[string]interface{}{}
			}
			valAsMapSlice[i] = castVal
		}
	case []map[string]interface{}:
		// If sourceMap[key] is already typed as []map[string]interface{}, then we're fine and can proceed without casting each value.
		valAsMapSlice = sourceMap[key].([]map[string]interface{})
	default:
		return []map[string]interface{}{}
	}
	return valAsMapSlice
}

func StringSlice(sourceMap map[string]interface{}, key string) []string {
	valAsSlice, ok := sourceMap[key].([]interface{})
	if !ok {
		return []string{}
	}

	valAsStringSlice := make([]string, len(valAsSlice))
	for index, nestedValue := range valAsSlice {
		valAsStringSlice[index], ok = nestedValue.(string)
		if !ok {
			return []string{}
		}
	}
	return valAsStringSlice
}

func CastInterface(in interface{}, out interface{}) error {
	if reflect.ValueOf(out).Kind() != reflect.Ptr {
		return fmt.Errorf("out must be a pointer")
	}

	data, err := json.Marshal(in)

	if err != nil {
		return err
	}

	err = json.Unmarshal(data, out)

	return err
}

func IsMap(sourceMap map[string]interface{}, key string) error {
	// Check that key exists in map
	val, ok := sourceMap[key]
	if !ok {
		return fmt.Errorf("%s must be provided", key)
	}

	// Check that value of key is a map
	_, ok = val.(map[string]interface{})
	if !ok {
		return fmt.Errorf("%s found with wrong type: expected JSON object", key)
	}

	return nil
}

func Map(sourceMap map[string]interface{}, key string) map[string]interface{} {
	val, ok := sourceMap[key].(map[string]interface{})
	if !ok {
		return map[string]interface{}{}
	}
	return val
}

func StringSliceContainsKey(items []string, key string) bool {
	for _, item := range items {
		if item == key {
			return true
		}
	}
	return false
}

func StringSliceContainsKeyCaseInsensitive(items []string, item string) (bool, string) {
	for _, it := range items {
		if strings.EqualFold(it, item) {
			return true, it
		}
	}
	return false, ""
}

func IsValidUUID(u string) bool {
	_, err := uuid.FromString(u)
	return err == nil
}

func CopyMap(in map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{})

	for k, v := range in {
		out[k] = v
	}

	return out
}

func Pop(slice []string) (string, []string) {
	return slice[len(slice)-1], slice[:len(slice)-1]
}

// CombineStructWithMap takes a map and adds its values into the struct behind an interface,
// creating a new struct of that type
func CombineStructWithMap(s interface{}, m map[string]interface{}) (interface{}, error) {
	typeof := reflect.TypeOf(s)
	newval := reflect.New(typeof).Elem()
	sourceval := reflect.ValueOf(s)
	for i := 0; i < typeof.NumField(); i++ {
		typeField := typeof.Field(i)
		names := []string{
			typeField.Name,
			typeField.Tag.Get("json"),
		}
		vf := newval.Field(i)
		setMaybe(m, names, sourceval.Field(i), &vf)
	}
	return newval.Interface(), nil
}

// setMaybe will take a map, an array of keys, a default reflect.Value, and a val *reflect.Value
// If any of the keys are in map m, and the type matches, it will set val to m[name], otherwise default.
func setMaybe(m map[string]interface{}, names []string, defaultVal reflect.Value, val *reflect.Value) {
	for _, name := range names {
		if mf, ok := m[name]; ok && val.Type() == reflect.TypeOf(mf) && val.CanSet() {
			val.Set(reflect.ValueOf(mf))
			return
		}
	}
	val.Set(defaultVal)
}

func GetValue(in interface{}, key string) (value string, err error) {
	valOf := reflect.ValueOf(in)
	err = fmt.Errorf("key not found")

	for i := 0; i < valOf.NumField(); i++ {
		if key != valOf.Type().Field(i).Name {
			continue
		}

		valueField := valOf.Field(i)

		if valueField.Kind() == reflect.Ptr {
			value = fmt.Sprintf("%v", valueField.Elem())
		} else {
			value = fmt.Sprintf("%v", valueField)
		}

		err = nil
		break
	}

	return
}
