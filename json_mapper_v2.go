// Package jsonmapper_v2 provides a library for manipulating JSON structures.
// It allows for navigating, adding, and removing values within a JSON structure
// through specified key paths.
package jsonmapper_v2

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

// JsonMapper is a struct that implements the JsonMapper interface.
// It is used for manipulating JSON structures.
type JsonMapper struct {
	m map[string]interface{}
}

// NewJsonMapFromFile initializes a new JsonMapper instance from a JSON file.
// It reads the file, unmarshals its content into a map[string]interface{}, and returns a new JsonMapper instance for manipulation.
// Returns an error if reading the file or parsing the JSON fails.
func NewJsonMapStr(s string) (*JsonMapper, error) {
	var m map[string]interface{}
	if err := json.Unmarshal([]byte(s), &m); err != nil {
		return nil, err
	}
	return &JsonMapper{m: m}, nil
}

// NewJsonMapFromFile initializes a new JsonMapper instance from a JSON file.
// It reads the file, unmarshals its content into a map[string]interface{}, and returns a new JsonMapper instance for manipulation.
// Returns an error if reading the file or parsing the JSON fails.
func NewJsonMapFile(filePath string) (*JsonMapper, error) {
	byteValue, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var m map[string]interface{}
	if err := json.Unmarshal(byteValue, &m); err != nil {
		return nil, err
	}

	return &JsonMapper{m: m}, nil
}

// NewJsonMapFromBytes initializes a new JsonMapper instance from a slice of bytes containing JSON data.
// It unmarshals the byte slice into a map[string]interface{} for manipulation.
// Useful for processing JSON data received from APIs or other byte streams.
// Returns an error if unmarshaling fails.
func NewJsonMapBytes(data []byte) (*JsonMapper, error) {
	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	return &JsonMapper{m: m}, nil
}

// Find retrieves the value located at the specified keyPath within the JSON structure.
// The keyPath is a dot-separated string indicating the path to the value.
// Supports array indexing using the notation [index] or .index.
// Returns the value as an interface{} or an error if the path is invalid or the key does not exist.
func (j *JsonMapper) Find(keyPath string) (interface{}, error) {
	if keyPath == "" {
		return j.m, nil
	}

	convertedKeyPath := convertBracketsToDots(keyPath)
	keys := strings.Split(convertedKeyPath, ".")
	var current interface{} = j.m

	for _, key := range keys {
		switch currentType := current.(type) {
		case map[string]interface{}:
			if value, ok := currentType[key]; ok {
				current = value
			} else {
				return nil, fmt.Errorf("key not found: %s", key)
			}
		case []interface{}:
			index, err := strconv.Atoi(key)
			if err != nil {
				return nil, fmt.Errorf("invalid array index: %s", key)
			}
			if index < 0 || index >= len(currentType) {
				return nil, fmt.Errorf("array index out of range: %d", index)
			}
			current = currentType[index]
		default:
			return current, nil
		}
	}

	return current, nil
}

// Add inserts or updates a value at the specified keyPath within the JSON structure.
// If the path does not exist, it creates the necessary structures (maps or slices) along the path.
// If the keyPath ends with an array index, the value is inserted at the specified index, replacing existing values if necessary.
// Supports negative indexing with -1 to append to slices.
// Returns an error if the path is invalid or if the operation cannot be completed.
func (j *JsonMapper) Add(keyPath string, value interface{}) error {
	convertedKeyPath := convertBracketsToDots(keyPath)
	keys := strings.Split(convertedKeyPath, ".")
	var current interface{} = j.m

	for i := 0; i < len(keys); i++ {
		key := keys[i]
		lastKey := i == len(keys)-1

		if lastKey {
			switch parent := current.(type) {
			case map[string]interface{}:
				parent[key] = value
			case []interface{}:
				index, err := strconv.Atoi(key)
				if err != nil {
					return fmt.Errorf("invalid array index '%s': %v", key, err)
				}
				if index == -1 {
					current = append(parent, value)
				} else if index >= 0 && index < len(parent) {
					parent[index] = value
				} else {
					return fmt.Errorf("array index '%d' is out of range", index)
				}

				if i > 0 {
					parentKey := keys[i-1]
					grandParent, _ := j.m[keys[0]].(map[string]interface{})
					for _, k := range keys[1 : i-1] {
						grandParent = grandParent[k].(map[string]interface{})
					}
					grandParent[parentKey] = current
				}
			}
			break
		} else {
			if next, ok := current.(map[string]interface{})[key]; ok {
				current = next
			} else if index, err := strconv.Atoi(key); err == nil {
				if nextSlice, ok := current.([]interface{}); ok && index >= 0 && index < len(nextSlice) {
					current = nextSlice[index]
				} else {
					return fmt.Errorf("invalid array index '%s': %v", key, err)
				}
			} else {
				current.(map[string]interface{})[key] = make(map[string]interface{})
				current = current.(map[string]interface{})[key]
			}
		}
	}

	return nil
}

// Remove deletes the value located at the specified keyPath within the JSON structure.
// If the keyPath points to an array index, it removes the element at that index and shifts subsequent elements.
// Supports negative indexing with -1 to remove the last element of a slice.
// Returns an error if the path is invalid or the key does not exist.
func (j *JsonMapper) Remove(keyPath string) error {
	convertedKeyPath := convertBracketsToDots(keyPath)
	keys := strings.Split(convertedKeyPath, ".")
	current := j.m
	var parent map[string]interface{} = nil
	var parentKey string

	for i, key := range keys {
		if i == len(keys)-1 {
			break
		}

		if i == len(keys)-2 {
			parent = current
			parentKey = key
		}

		switch currentElement := current[key].(type) {
		case map[string]interface{}:
			current = currentElement
		case []interface{}:
			index, err := strconv.Atoi(keys[i+1])
			if err == nil && index == -1 {
				index = len(currentElement) - 1
			}
			if index < 0 || index >= len(currentElement) {
				return fmt.Errorf("array index '%d' is out of range", index)
			}
			if i == len(keys)-2 {
				updatedSlice := append(currentElement[:index], currentElement[index+1:]...)
				current[parentKey] = updatedSlice
				return nil
			}
			if nextElement, ok := currentElement[index].(map[string]interface{}); ok {
				current = nextElement
			} else {
				return fmt.Errorf("expected a map at '%s', but found a different type", keys[i+1])
			}
		default:
			return fmt.Errorf("unexpected type %T at '%s'", currentElement, key)
		}
	}

	if parent != nil {
		delete(parent, keys[len(keys)-1])
	}

	return nil
}

// Print returns the JSON structure as a compact string.
// Useful for logging or debugging purposes.
func (j *JsonMapper) Print() string {
	jsonString, err := json.Marshal(j.m)
	if err != nil {
		return ""
	}

	return string(jsonString)
}

// PrettyPrint returns the JSON structure as a well-formatted string with indentation.
// Enhances readability for logging or debugging.
func (j *JsonMapper) PrettyPrint() string {
	jsonString, err := json.MarshalIndent(j.m, "", "  ")
	if err != nil {
		return ""
	}

	return string(jsonString)
}

// Find<Type> functions (e.g., FindBool, FindString, FindInt) retrieve values of specific types from the JSON structure.
// Each function targets a specific type and returns the value at the given keyPath if it matches the expected type.
// The 'Or' variant of each function (e.g., FindBoolOr, FindStringOr) returns a default value if the target value does not exist or does not match the expected type.
// These functions simplify type assertions and error handling when accessing JSON data.

// FindBool searches for a boolean value at the given keyPath.
func (j *JsonMapper) FindBool(k string) (bool, error) {
	tmp, err := j.Find(k)
	if err != nil {
		return false, err
	}
	if boolValue, ok := tmp.(bool); ok {
		return boolValue, nil
	}
	return false, fmt.Errorf("value at %s is not a bool", k)
}

// FindBoolOr is similar to FindBool but returns a defaultValue if the value is not found.
func (j *JsonMapper) FindBoolOr(k string, defaultValue bool) bool {
	boolValue, err := j.FindBool(k)
	if err != nil {
		return defaultValue
	}
	return boolValue
}

// FindString searches for a string value at the given keyPath.
// It returns the string value found, or an error if the path does not exist or the value is not a string.
func (j *JsonMapper) FindString(k string) (string, error) {
	tmp, err := j.Find(k)
	if err != nil {
		return "", err
	}
	if strValue, ok := tmp.(string); ok {
		return strValue, nil
	}
	return "", fmt.Errorf("value at %s is not a string", k)
}

// FindStringOr is similar to FindString but returns the defaultValue if the value is not found or not a string.
func (j *JsonMapper) FindStringOr(k string, defaultValue string) string {
	strValue, err := j.FindString(k)
	if err != nil {
		return defaultValue
	}
	return strValue
}

// FindInt searches for an integer value at the given keyPath.
// It returns the integer value found, or an error if the path does not exist or the value is not an integer.
func (j *JsonMapper) FindInt(k string) (int, error) {
	tmp, err := j.Find(k)
	if err != nil {
		return 0, err
	}
	if intValue, ok := tmp.(float64); ok {
		return int(intValue), nil
	}
	return 0, fmt.Errorf("value at %s is not an int", k)
}

// FindIntOr is similar to FindInt but returns the defaultValue if the value is not found or not an integer.
func (j *JsonMapper) FindIntOr(k string, defaultValue int) int {
	intValue, err := j.FindInt(k)
	if err != nil {
		return defaultValue
	}
	return intValue
}

// FindFloat searches for a float value at the given keyPath.
// It returns the float value found, or an error if the path does not exist or the value is not a float.
func (j *JsonMapper) FindFloat(k string) (float64, error) {
	tmp, err := j.Find(k)
	if err != nil {
		return 0.0, err
	}
	if floatValue, ok := tmp.(float64); ok {
		return floatValue, nil
	}
	return 0.0, fmt.Errorf("value at %s is not a float", k)
}

// FindFloatOr is similar to FindFloat but returns the defaultValue if the value is not found or not a float.
func (j *JsonMapper) FindFloatOr(k string, defaultValue float64) float64 {
	floatValue, err := j.FindFloat(k)
	if err != nil {
		return defaultValue
	}
	return floatValue
}

// FindSlice searches for a slice at the given keyPath.
// It returns the slice found, or an error if the path does not exist or the value is not a slice.
func (j *JsonMapper) FindSlice(k string) ([]interface{}, error) {
	tmp, err := j.Find(k)
	if err != nil {
		return nil, err
	}
	if sliceValue, ok := tmp.([]interface{}); ok {
		return sliceValue, nil
	}
	return nil, fmt.Errorf("value at %s is not a slice", k)
}

// FindSliceOr is similar to FindSlice but returns the defaultValue if the value is not found or not a slice.
func (j *JsonMapper) FindSliceOr(k string, defaultValue []interface{}) []interface{} {
	sliceValue, err := j.FindSlice(k)
	if err != nil {
		return defaultValue
	}
	return sliceValue
}

// FindMap searches for a map at the given keyPath.
// It returns the map found, or an error if the path does not exist or the value is not a map.
func (j *JsonMapper) FindMap(k string) (map[string]interface{}, error) {
	tmp, err := j.Find(k)
	if err != nil {
		return nil, err
	}
	if mapValue, ok := tmp.(map[string]interface{}); ok {
		return mapValue, nil
	}
	return nil, fmt.Errorf("value at %s is not a map", k)
}

// FindMapOr is similar to FindMap but returns the defaultValue if the value is not found or not a map.
func (j *JsonMapper) FindMapOr(k string, defaultValue map[string]interface{}) map[string]interface{} {
	mapValue, err := j.FindMap(k)
	if err != nil {
		return defaultValue
	}
	return mapValue
}

// FindUint searches for an unsigned integer value at the given keyPath.
// It returns the unsigned integer value found, or an error if the path does not exist or the value is not an unsigned integer.
func (j *JsonMapper) FindUint(k string) (uint, error) {
	tmp, err := j.Find(k)
	if err != nil {
		return 0, err
	}
	if floatValue, ok := tmp.(float64); ok {
		return uint(floatValue), nil
	}
	return 0, fmt.Errorf("value at %s is not an uint", k)
}

// FindUintOr is similar to FindUint but returns the defaultValue if the value is not found or not an unsigned integer.
func (j *JsonMapper) FindUintOr(k string, defaultValue uint) uint {
	uintValue, err := j.FindUint(k)
	if err != nil {
		return defaultValue
	}
	return uintValue
}

// FindUint32 searches for an unsigned 32-bit integer value at the given keyPath.
// It returns the unsigned 32-bit integer value found, or an error if the path does not exist or the value is not an unsigned 32-bit integer.
func (j *JsonMapper) FindUint32(k string) (uint32, error) {
	tmp, err := j.Find(k)
	if err != nil {
		return 0, err
	}
	if floatValue, ok := tmp.(float64); ok {
		return uint32(floatValue), nil
	}
	return 0, fmt.Errorf("value at %s is not an uint32", k)
}

// FindUint32Or is similar to FindUint32 but returns the defaultValue if the value is not found or not an unsigned 32-bit integer.
func (j *JsonMapper) FindUint32Or(k string, defaultValue uint32) uint32 {
	uint32Value, err := j.FindUint32(k)
	if err != nil {
		return defaultValue
	}
	return uint32Value
}

// FindUint64 searches for an unsigned 64-bit integer value at the given keyPath.
// It returns the unsigned 64-bit integer value found, or an error if the path does not exist or the value is not an unsigned 64-bit integer.
func (j *JsonMapper) FindUint64(k string) (uint64, error) {
	tmp, err := j.Find(k)
	if err != nil {
		return 0, err
	}
	if floatValue, ok := tmp.(float64); ok {
		return uint64(floatValue), nil
	}
	return 0, fmt.Errorf("value at %s is not an uint64", k)
}

// FindUint64Or is similar to FindUint64 but returns the defaultValue if the value is not found or not an unsigned 64-bit integer.
func (j *JsonMapper) FindUint64Or(k string, defaultValue uint64) uint64 {
	uint64Value, err := j.FindUint64(k)
	if err != nil {
		return defaultValue
	}
	return uint64Value
}

// FindSliceOfMaps searches for a slice of maps at the given keyPath.
// It returns the slice of maps found, or an error if the path does not exist or the value is not a slice of maps.
func (j *JsonMapper) FindSliceOfMaps(k string) ([]map[string]interface{}, error) {
	tmp, err := j.Find(k)
	if err != nil {
		return nil, err
	}
	if slice, ok := tmp.([]interface{}); ok {
		var sliceOfMaps []map[string]interface{}
		for _, item := range slice {
			if m, ok := item.(map[string]interface{}); ok {
				sliceOfMaps = append(sliceOfMaps, m)
			} else {
				return nil, fmt.Errorf("element in slice at %s is not a map", k)
			}
		}
		return sliceOfMaps, nil
	}
	return nil, fmt.Errorf("value at %s is not a slice of maps", k)
}

// FindMapOfSlices searches for a map of slices at the given keyPath.
// It returns the map of slices found, or an error if the path does not exist or the value is not a map of slices.
func (j *JsonMapper) FindMapOfSlices(k string) (map[string][]interface{}, error) {
	tmp, err := j.Find(k)
	if err != nil {
		return nil, err
	}
	if m, ok := tmp.(map[string]interface{}); ok {
		mapOfSlices := make(map[string][]interface{})
		for key, value := range m {
			if slice, ok := value.([]interface{}); ok {
				mapOfSlices[key] = slice
			} else {
				return nil, fmt.Errorf("value for key %s in map at %s is not a slice", key, k)
			}
		}
		return mapOfSlices, nil
	}
	return nil, fmt.Errorf("value at %s is not a map of slices", k)
}

// WriteFile saves the current JSON structure to a file at the specified filePath.
// The 'pretty' parameter controls whether the JSON is formatted with indentation.
// Overwrites the file if it already exists, or creates a new file if it does not.
// Returns an error if writing to the file fails.
func (j *JsonMapper) WriteFile(filePath string, pretty bool) error {
	var data []byte
	var err error

	if pretty {
		data, err = json.MarshalIndent(j.m, "", "  ")
	} else {
		data, err = json.Marshal(j.m)
	}
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %v", err)
	}

	err = os.WriteFile(filePath, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write file: %v", err)
	}

	return nil
}

// convertBracketsToDots transforms array index accessors from bracket notation [index] to dot notation .index in a keyPath.
// Facilitates uniform handling of array indexes in keyPaths, aligning with the dot-separated keyPath format used by other functions.
// This internal function supports the parsing and manipulation of keyPaths with array indexes.
func convertBracketsToDots(keyPath string) string {
	re := regexp.MustCompile(`\[\-?(\d+)\]`)
	return re.ReplaceAllStringFunc(keyPath, func(match string) string {
		index := strings.Trim(match, "[]")
		return "." + index
	})
}

// TODO: go version 1.18 + update gopls
// func (j *JsonMapper) FindCustomType[T any](k string) (T, error) {
//     var result T
//     tmp, err := j.Find(k)
//     if err != nil {
//         return result, err
//     }
//     tmpBytes, err := json.Marshal(tmp)
//     if err != nil {
//         return result, err
//     }
//     if err := json.Unmarshal(tmpBytes, &result); err != nil {
//         return result, fmt.Errorf("value at %s cannot be converted to the desired type: %v", k, err)
//     }
//     return result, nil
// }
