package jsonmapper_v2

import (
	"fmt"
	"reflect"
)

// FindAllWithCondition searches through the JSON structure starting from the given keyPath
// and returns all paths that satisfy the specified conditions. The conditions parameter
// should be a map or nested maps with logical and comparison operators as keys.
// Supported logical operators include "and", "or", "xor", and "nor".
// Supported comparison operators include "eq" (equal), "neq" (not equal),
// "lt" (less than), "lte" (less than or equal), "gt" (greater than), and "gte" (greater than or equal).
// The function recursively traverses the JSON structure, evaluating each value against the conditions.
// If a value satisfies the conditions, its path is added to the results.
//
// Parameters:
//   - keyPath: A dot-separated string specifying the starting point within the JSON structure.
//     If empty, the search starts from the root of the JSON structure.
//   - conditions: A map or nested maps specifying the conditions that values must satisfy.
//     The keys are logical or comparison operators, and the values are the operands.
//
// Returns:
// - A slice of strings containing the paths of all values that satisfy the conditions.
// - An error if the conditions are invalid or if an error occurs during the evaluation.
//
// Example:
// To find all paths where the "id" is greater than 2, you could use:
// conditions := map[string]interface{}{"gt": 2}
// paths, err := jm.FindAllWithCondition("testData.s2", conditions)
func (j *JsonMapper) FindAllWithCondition(keyPath string, conditions interface{}) ([]string, error) {
	var results []string

	var evaluate func(interface{}, string) error
	evaluate = func(current interface{}, currentPath string) error {
		switch currentType := current.(type) {
		case map[string]interface{}:
			for k, v := range currentType {
				newPath := currentPath
				if newPath != "" {
					newPath += "."
				}
				newPath += k
				evaluate(v, newPath)
			}
		case []interface{}:
			for i, v := range currentType {
				newPath := fmt.Sprintf("%s[%d]", currentPath, i)
				evaluate(v, newPath)
			}
		default:
			satisfied, err := j.evaluateCondition(current, conditions)
			if err != nil {
				return err
			}
			if satisfied {
				results = append(results, currentPath)
			}
		}
		return nil
	}

	var startValue interface{}
	var err error

	if keyPath == "" {
		startValue = j.m // Use the entire map if the keyPath is root
	} else {
		startValue, err = j.Find(keyPath)
		if err != nil {
			return nil, err
		}
	}

	err = evaluate(startValue, keyPath)
	if err != nil {
		return nil, err
	}

	return results, nil
}

// evaluateCondition checks if the given value satisfies the specified conditions.
// The conditions parameter can be a map containing comparison operations
// or a map of logical operations that contain comparison operations.
// This function supports handling complex logical expressions using "and", "or", "xor", and "nor" logical operations,
// and it supports "eq" (equal), "neq" (not equal), "lt" (less than), "lte" (less than or equal),
// "gt" (greater than), and "gte" (greater than or equal) comparison operations.
//
// Parameters:
//   - value: The value to be evaluated against the conditions.
//   - conditions: A map or nested maps specifying the conditions. The keys represent the operators,
//     and the values represent the operands or further nested conditions.
//
// Returns:
// - A boolean indicating whether the value satisfies the conditions.
// - An error if an unsupported operation is encountered or if there's an issue evaluating the conditions.
func (j *JsonMapper) evaluateCondition(value interface{}, conditions interface{}) (bool, error) {
	switch cond := conditions.(type) {
	case map[string]interface{}:
		for op, conditionValue := range cond {
			return j.checkCondition(value, op, conditionValue)
		}
	case map[string][]map[string]interface{}:
		for logicalOp, subConditions := range cond {
			switch logicalOp {
			case "and", "AND":
				for _, conditionMap := range subConditions {
					for op, conditionValue := range conditionMap {
						satisfied, err := j.checkCondition(value, op, conditionValue)
						if err != nil || !satisfied {
							return false, err
						}
					}
				}
				return true, nil
			case "or", "OR":
				satisfiedAny := false
				for _, conditionMap := range subConditions {
					for op, conditionValue := range conditionMap {
						satisfied, err := j.checkCondition(value, op, conditionValue)
						if err != nil {
							return false, err
						}
						if satisfied {
							satisfiedAny = true
							break
						}
					}
					if satisfiedAny {
						break
					}
				}
				return satisfiedAny, nil
			case "xor", "XOR":
				satisfiedCount := 0
				for _, conditionMap := range subConditions {
					for op, conditionValue := range conditionMap {
						satisfied, err := j.checkCondition(value, op, conditionValue)
						if err != nil {
							return false, err
						}
						if satisfied {
							satisfiedCount++
						}
					}
				}
				return satisfiedCount == 1, nil
			case "nor", "NOR":
				for _, conditionMap := range subConditions {
					for op, conditionValue := range conditionMap {
						satisfied, err := j.checkCondition(value, op, conditionValue)
						if err != nil {
							return false, err
						}
						if satisfied {
							return false, nil
						}
					}
				}
				return true, nil
			default:
				return false, fmt.Errorf("unsupported logical operation: %s", logicalOp)
			}
		}
	default:
		return false, fmt.Errorf("invalid conditions format")
	}
	return false, fmt.Errorf("no valid condition found")
}

// checkCondition evaluates a single comparison operation between a value and a threshold.
// This function supports "eq" (equal), "neq" (not equal), "lt" (less than), "lte" (less than or equal),
// "gt" (greater than), and "gte" (greater than or equal) operations. The function is designed
// to work with numeric values but also supports equality and inequality checks for other data types.
//
// Parameters:
// - value: The value to be compared.
// - op: A string representing the comparison operation.
// - threshold: The value to compare against.
//
// Returns:
// - A boolean indicating the result of the comparison.
// - An error if the operation is not supported for the given value types or if an error occurs during comparison.
func (j *JsonMapper) checkCondition(value interface{}, op string, threshold interface{}) (bool, error) {
	vValue := reflect.ValueOf(value)
	vThreshold := reflect.ValueOf(threshold)

	switch op {
	case "eq":
		if isNumeric(value) && isNumeric(threshold) {
			valueFloat, err := convertToFloat64(value)
			if err != nil {
				return false, err
			}
			thresholdFloat, err := convertToFloat64(threshold)
			if err != nil {
				return false, err
			}
			return valueFloat == thresholdFloat, nil
		}

		return reflect.DeepEqual(value, threshold), nil
	case "neq":
		if reflect.TypeOf(value) != reflect.TypeOf(threshold) {
			return true, nil
		}
		return !reflect.DeepEqual(value, threshold), nil

	case "lt", "lte", "gt", "gte":
		if vValue.Kind().String() == "int" || vValue.Kind().String() == "float64" &&
			(vThreshold.Kind().String() == "int" || vThreshold.Kind().String() == "float64") {
			return compareNumericUsingReflect(vValue, vThreshold, op)
		} else {
			return false, fmt.Errorf("comparison %s not supported for non-numeric types", op)
		}
	default:
		return false, fmt.Errorf("unsupported operation: %s", op)
	}
}

// compareNumericUsingReflect performs a numeric comparison between two reflect.Value instances
// based on the specified operation. This function is utilized internally by checkCondition
// to handle numeric comparisons using reflection. Supported operations include
// "lt" (less than), "lte" (less than or equal), "gt" (greater than), and "gte" (greater than or equal).
//
// Parameters:
// - vValue: The first value wrapped in a reflect.Value.
// - vThreshold: The second value wrapped in a reflect.Value to compare against.
// - op: A string indicating the comparison operation.
//
// Returns:
// - A boolean indicating the result of the comparison.
// - An error if the operation is not supported or if an error occurs during comparison.
func compareNumericUsingReflect(vValue, vThreshold reflect.Value, op string) (bool, error) {
	valueFloat, err := convertToFloat64(vValue.Interface())
	if err != nil {
		return false, err
	}

	thresholdFloat, err := convertToFloat64(vThreshold.Interface())
	if err != nil {
		return false, err
	}

	switch op {
	case "lt":
		return valueFloat < thresholdFloat, nil
	case "lte":
		return valueFloat <= thresholdFloat, nil
	case "gt":
		return valueFloat > thresholdFloat, nil
	case "gte":
		return valueFloat >= thresholdFloat, nil
	default:
		return false, fmt.Errorf("unsupported numeric comparison operation: %s", op)
	}
}

// convertToFloat64 attempts to convert various numeric types to float64. This function supports
// conversion from integer types (int, int8, int16, int32, int64) and unsigned integer types
// (uint, uint8, uint16, uint32, uint64), as well as float32 and float64 types. It is used internally
// to normalize numeric values for comparison operations.
//
// Parameters:
// - value: The value to be converted to float64.
//
// Returns:
// - The converted float64 value.
// - An error if the value cannot be converted to float64 due to unsupported type.
func convertToFloat64(value interface{}) (float64, error) {
	switch v := value.(type) {
	case float64:
		return v, nil
	case float32:
		return float64(v), nil
	case int, int8, int16, int32, int64:
		return float64(reflect.ValueOf(value).Int()), nil
	case uint, uint8, uint16, uint32, uint64:
		return float64(reflect.ValueOf(value).Uint()), nil
	default:
		return 0, fmt.Errorf("unsupported type for numeric comparison: %T", value)
	}
}

// isNumeric checks if the given value is of a numeric type. This function supports checking
// against integer types (int, int8, int16, int32, int64), unsigned integer types (uint, uint8, uint16, uint32, uint64),
// and floating-point types (float32, float64). It is used internally to determine if a value
// can be used in numeric comparison operations.
//
// Parameters:
// - value: The value to check.
//
// Returns:
// - A boolean indicating whether the value is of a numeric type.
func isNumeric(value interface{}) bool {
	switch value.(type) {
	case float64, float32, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return true
	default:
		return false
	}
}
