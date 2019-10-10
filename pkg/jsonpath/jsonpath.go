// Simple Implementation of JSON Path for state machine
package jsonpath

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"
)

/*
The `data` must be from JSON Unmarshal, that way we can guarantee the types:

bool, for JSON booleans
float64, for JSON numbers
string, for JSON strings
[]interface{}, for JSON arrays
map[string]interface{}, for JSON objects
nil for JSON null

*/

var ErrNotFoundError = errors.New("Not Found")

type Path struct {
	path []string
}

// NewPath takes string returns JSONPath Object
func NewPath(pathString string) (*Path, error) {
	path := Path{}
	pathArray, err := ParsePathString(pathString)
	path.path = pathArray
	return &path, err
}

// UnmarshalJSON makes a path out of a json string
func (path *Path) UnmarshalJSON(b []byte) error {
	var pathString string
	err := json.Unmarshal(b, &pathString)

	if err != nil {
		return err
	}

	pathArray, err := ParsePathString(pathString)

	if err != nil {
		return err
	}

	path.path = pathArray
	return nil
}

// MarshalJSON converts path to json string
func (path *Path) MarshalJSON() ([]byte, error) {
	if len(path.path) == 0 {
		return json.Marshal("$")
	}
	return json.Marshal(path.String())
}

func (path *Path) String() string {
	return fmt.Sprintf("$.%v", strings.Join(path.path[:], "."))
}

// ParsePathString parses a path string
func ParsePathString(pathString string) ([]string, error) {
	// must start with $.<value> otherwise empty path
	if pathString == "" || pathString[0:1] != "$" {
		return nil, fmt.Errorf("Bad JSON path: must start with $")
	}

	if pathString == "$" {
		// Default is no path
		return []string{}, nil
	}

	if len(pathString) < 2 {
		// This handles the case for $. or $* which are invalid
		return nil, fmt.Errorf("Bad JSON path: cannot not be 2 characters")
	}

	head := pathString[2:]
	pathArray := strings.Split(head, ".")

	// if path contains an "" error
	for _, p := range pathArray {
		if p == "" {
			return nil, fmt.Errorf("Bad JSON path: has empty element")
		}
	}
	// Simple Path Builder
	return pathArray, nil
}

// PUBLIC METHODS

// GetTime returns Time from Path
func (path *Path) GetTime(input interface{}) (*time.Time, error) {
	outputValue, err := path.Get(input)

	if err != nil {
		return nil, fmt.Errorf("GetTime Error %q", err)
	}

	var output time.Time
	switch outputValueCast := outputValue.(type) {
	case string:
		output, err = time.Parse(time.RFC3339, outputValueCast)
		if err != nil {
			return nil, fmt.Errorf("GetTime Error: time error %q", err)
		}
	default:
		return nil, fmt.Errorf("GetTime Error: time must be string")
	}

	return &output, nil
}

// GetBool returns Bool from Path
func (path *Path) GetBool(input interface{}) (*bool, error) {
	outputValue, err := path.Get(input)

	if err != nil {
		return nil, fmt.Errorf("GetBool Error %q", err)
	}

	var output bool
	switch outputValueCast := outputValue.(type) {
	case bool:
		output = outputValueCast
	default:
		return nil, fmt.Errorf("GetBool Error: must return bool")
	}

	return &output, nil
}

// GetNumber returns Number from Path
func (path *Path) GetNumber(input interface{}) (*float64, error) {
	outputValue, err := path.Get(input)

	if err != nil {
		return nil, fmt.Errorf("GetFloat Error %q", err)
	}

	var output float64
	switch outputValueCast := outputValue.(type) {
	case float64:
		output = outputValueCast
	case int:
		output = float64(outputValue.(int))
	default:
		return nil, fmt.Errorf("GetFloat Error: must return float")
	}

	return &output, nil
}

// GetString returns String from Path
func (path *Path) GetString(input interface{}) (*string, error) {
	outputValue, err := path.Get(input)

	if err != nil {
		return nil, fmt.Errorf("GetString Error %q", err)
	}

	var output string
	switch outputValueCast := outputValue.(type) {
	case string:
		output = outputValueCast
	default:
		return nil, fmt.Errorf("GetString Error: must return string")
	}

	return &output, nil
}

// GetMap returns Map from Path
func (path *Path) GetMap(input interface{}) (output map[string]interface{}, err error) {
	outputValue, err := path.Get(input)

	if err != nil {
		return nil, fmt.Errorf("GetMap Error %q", err)
	}

	switch outputValueCast := outputValue.(type) {
	case map[string]interface{}:
		output = outputValueCast
	default:
		return nil, fmt.Errorf("GetMap Error: must return map")
	}

	return output, nil
}

// Get returns interface from Path
func (path *Path) Get(input interface{}) (value interface{}, err error) {
	if path == nil {
		return input, nil // Default is $
	}
	return recursiveGet(input, path.path)
}

// Set sets a Value in a map with Path
func (path *Path) Set(input interface{}, value interface{}) (output interface{}, err error) {
	var setPath []string
	if path == nil {
		setPath = []string{} // default "$"
	} else {
		setPath = path.path
	}

	if len(setPath) == 0 {
		// The output is the value
		switch valueCast := value.(type) {
		case map[string]interface{}:
			output = valueCast
			return output, nil
		case []interface{}:
			output = value.([]interface{})
			return output, nil
		default:
			return nil, fmt.Errorf("Cannot Set value %q type %q in root JSON path $", value, reflect.TypeOf(value))
		}
	}
	return recursiveSet(input, value, setPath), nil
}

// PRIVATE METHODS

func recursiveSet(data interface{}, value interface{}, path []string) (output interface{}) {
	var dataMap map[string]interface{}

	switch dataCast := data.(type) {
	case map[string]interface{}:
		dataMap = dataCast
	default:
		// Overwrite current data with new map
		// this will work for nil as well
		dataMap = make(map[string]interface{})
	}

	if len(path) == 1 {
		dataMap[path[0]] = value
	} else {
		dataMap[path[0]] = recursiveSet(dataMap[path[0]], value, path[1:])
	}

	return dataMap
}

func recursiveGet(data interface{}, path []string) (interface{}, error) {
	if len(path) == 0 {
		return data, nil
	}

	if data == nil {
		return nil, errors.New("Not Found")
	}

	switch dataCast := data.(type) {
	case map[string]interface{}:
		value, ok := dataCast[path[0]]

		if !ok {
			return data, ErrNotFoundError
		}

		return recursiveGet(value, path[1:])

	default:
		return data, ErrNotFoundError
	}
}
