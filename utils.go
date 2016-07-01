package pkgs

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

func MustDefaultBool(obj map[string]interface{}, path string, d bool) bool {
	v, err := DefaultBool(obj, path, d)
	if err != nil {
		log.Fatal(err)
	}
	return v
}

func DefaultBool(obj map[string]interface{}, path string, d bool) (bool, error) {
	v := TryLookupMap(obj, path)
	switch b := v.(type) {
	case bool:
		return b, nil
	case nil:
		return d, nil
	}
	return d, errors.New(path + " is not bool value")
}

func TryLookupMap(obj map[string]interface{}, path string) interface{} {
	v, err := GetProperty(obj, path)
	if err != nil {
		log.Debugln(err)
	}
	return v
}

// GetProperty returns a property if it exist.
// return boolean if come across boolean on path
//
//    property, err := GetProperty(document, "one.two.three[0]")
//    property, err := GetProperty(document, "one.two.three[0]", ".")
//    property, err := GetProperty(document, "one/two/three[0]", "/")
//
// Property type is `interface{}`
func GetProperty(original_data map[string]interface{}, path string, separator_arr ...string) (path_parsed interface{}, err error) {
	var separator = "."
	if len(separator_arr) > 0 {
		if len(separator_arr[0]) > 0 {
			separator = separator_arr[0]
		}
	}

	// Protect the original map :D
	data := make(map[string]interface{})
	for k, v := range original_data {
		data[k] = v
	}
	err = fmt.Errorf("Property %s does not exist", path)

	if len(path) == 0 {
		path = separator
	}

	levels_tmp := strings.Split(path, separator)
	levels := make([]string, 0)
	for _, level_tmp := range levels_tmp {
		if len(level_tmp) > 0 {
			levels = append(levels, level_tmp)
		}
	}

	if len(levels) > 0 && path != separator {
		path_level_one := levels[0]

		// If we have a level in path_level_one

		re := regexp.MustCompile(`\w+\[\d+\]{1}`)
		if matched := re.FindString(path_level_one); len(matched) > 0 {
			property_re := regexp.MustCompile(`\w+`)
			index_re := regexp.MustCompile(`\[\d+\]{1}`)
			// Get a property
			// avatars
			property := property_re.FindString(path_level_one)

			// Get an index
			index_found := index_re.FindString(path_level_one)

			// If index > 0 - check if this property is array
			if len(index_found) > 0 {
				if len(property) > 0 {
					path_level_one = property
				}
				index_found = strings.Trim(index_found, "[]")
				if index, err := strconv.Atoi(index_found); err == nil {
					if v, ok := data[property]; ok {
						if isKind(v, reflect.Slice) {
							slice := reflect.ValueOf(v)
							if index >= 0 && index < slice.Len() {
								value := slice.Index(index).Interface()

								data[property] = value
							} else {
								err = fmt.Errorf(
									"%s: Min index is 0, Max index is %d. You passed index %d", property, slice.Len(), index,
								)
								return path_parsed, err
							}
						} else {
							// TODO change to buissess
							if b, ok := v.(bool); ok {
								return b, nil
							}
							err = fmt.Errorf(
								"%s: is not an array", property,
							)
							return path_parsed, err
						}
					} else {
						err = fmt.Errorf(
							"Property %s does not exist", property,
						)
						return path_parsed, err
					}
				} else {
					err = fmt.Errorf(
						"%s must be of type %s",
						fmt.Sprintf("%s[%d]", property, index_found),
						"number",
					)
					return path_parsed, err
				}
			}
		}

		if len(levels[1:]) >= 1 {
			if level_one_value, ok := data[path_level_one]; ok {
				if level_one_value != nil {
					switch reflect.TypeOf(level_one_value).Kind() {
					case reflect.Map:
						if mapped_level_one_value, ok := level_one_value.(map[string]interface{}); ok {
							return GetProperty(mapped_level_one_value, strings.Join(levels[1:], separator), separator)
						}
					case reflect.Bool:
						// TODO change to buissess
						return level_one_value, nil
					default:
						// pass
					}
				}
			} else {
				err = fmt.Errorf(
					"Property %s does not exist", path_level_one,
				)
				return path_parsed, err
			}
		} else {
			if v, ok := data[path_level_one]; ok {
				path_parsed = v
				err = nil
			}
		}
	} else if path == separator {
		path_parsed = data
		err = nil
	}
	return
}

func isKind(what interface{}, kind reflect.Kind) bool {
	return reflect.ValueOf(what).Kind() == kind
}
