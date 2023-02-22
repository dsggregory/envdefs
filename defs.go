package defaults

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/iancoleman/strcase"
)

// ReadDefaults Fill struct with defaults from struct tags if not already set.
// Struct tags are of the form:
//    `env:"<ENVVAR | ->[,opts...]"`
//    Hyphen (-) instead of an environment variable name means the struct field is to be ignored
//    opts are:
//      required: a value from the struct itself or the environment var must be set
func ReadDefaults(defs interface{}) error {
	return readDefaults(defs)
}

// Lookup the env from |key| renamed to uppercase, hyphen is underscore, and return it or
// the |defaultVal| in the type of |defaultVal|
func lookupEnv(envNm string, defaultVal interface{}) (interface{}, error) {
	var res interface{}

	if val, ok := os.LookupEnv(envNm); ok {
		res = val
	} else {
		return defaultVal, nil
	}
	switch t := defaultVal.(type) {
	case int:
		v, err := strconv.Atoi(res.(string))
		if err != nil {
			return nil, fmt.Errorf("%w, lookupEnv[%s]: %v\n", err, envNm, res)
		}
		return v, nil
	case int64:
		v, err := strconv.ParseInt(res.(string), 10, 64)
		if err != nil {
			return nil, fmt.Errorf("%w, lookupEnv[%s]: %v\n", err, envNm, res)
		}
		return v, nil
	case float64:
		v, err := strconv.ParseFloat(res.(string), 64)
		if err != nil {
			return nil, fmt.Errorf("%w, lookupEnv[%s]: %v\n", err, envNm, res)
		}
		return v, nil
	case bool:
		bstr := strings.ToUpper(res.(string))
		if bstr == "TRUE" || bstr == "1" {
			return true, nil
		}
		return false, nil
	case string:
		return res, nil
	case time.Duration:
		v, err := time.ParseDuration(res.(string))
		if err != nil {
			return nil, fmt.Errorf("%w, lookupEnv[%s]: %v\n", err, envNm, res)
		}
		return v, nil
	default:
		return nil, fmt.Errorf("lookupEnv[%s]: unsupported type %v", envNm, t)
	}
}

func readDefaults(defs interface{}) error {
	v := reflect.ValueOf(defs)
	if v.Kind() != reflect.Ptr || v.IsNil() {
		return fmt.Errorf("argument is not a struct pointer")
	}

	if err := reflectStruct(v, ""); err != nil {
		return err
	}

	return nil
}

func setDefault(fValue reflect.Value, field reflect.StructField, pfx string) error {
	fTag := field.Tag

	// flag struct tag
	flagName := strcase.ToKebab(pfx) + strcase.ToKebab(field.Name)
	flagTag, flagTagOK := fTag.Lookup("flag")
	if flagTag != "" {
		if flagTag == "-" {
			// the ignore tag
			return nil
		}
		flagName = strcase.ToKebab(pfx) + flagTag
	}

	// for a nested struct or struct pointer
	if fValue.Kind() == reflect.Ptr || fValue.Kind() == reflect.Struct {
		if fValue.Kind() == reflect.Ptr && fValue.IsNil() {
			return nil
		}
		fpfx := flagName + "-"
		// an explicitly empty flagName on the nested structure means no prefix for its fields
		if flagTagOK && flagTag == "" {
			fpfx = ""
		}
		addr := fValue
		if fValue.Kind() != reflect.Ptr {
			addr = fValue.Addr()
		} else if addr.Elem().Kind() != reflect.Struct {
			return nil
		}
		if err := reflectStruct(addr, fpfx); err != nil {
			return fmt.Errorf("%w; %s: field failure", err, field.Name)
		}
		return nil
	}

	// env struct tag and default value
	defaultVal := fValue.Interface()
	envName := ""
	envTag, envTagOK := fTag.Lookup("env")
	if envTagOK {
		envName = envTag
	} else {
		envName = strcase.ToScreamingSnake(flagName)
	}
	// opts can be comma-sep (required)
	envRequired := false
	opts := strings.Split(envTag, ",")
	if len(opts) > 1 {
		envTag = opts[0]
		envName = opts[0]
		for i := range opts {
			if opts[i] == "required" {
				envRequired = true
			}
		}
	}
	// envTag of "-" means do not consider OS environment variable
	if envTag != "-" {
		d, err := lookupEnv(envName, defaultVal)
		if err != nil {
			return err
		}
		defaultVal = d
	}

	if envRequired && defaultVal == nil {
		return fmt.Errorf("environment %s is required", envName)
	}

	if !fValue.CanAddr() {
		return fmt.Errorf("unable to address field %s", field.Name)
	}

	switch field.Type.String() {
	case "int":
		x := fValue.Addr().Interface().(*int)
		*x = defaultVal.(int)
		//flagset.IntVar(x, flagName, defaultVal.(int), flagUsage)
	case "int64":
		x := fValue.Addr().Interface().(*int64)
		*x = defaultVal.(int64)
		//flagset.Int64Var(x, flagName, defaultVal.(int64), flagUsage)
	case "float64":
		x := fValue.Addr().Interface().(*float64)
		*x = defaultVal.(float64)
		//flagset.Float64Var(x, flagName, defaultVal.(float64), flagUsage)
	case "string":
		x := fValue.Addr().Interface().(*string)
		*x = defaultVal.(string)
		if envRequired && *x == "" {
			return fmt.Errorf("environment %s is required", envTag)
		}
		//flagset.StringVar(x, flagName, defaultVal.(string), flagUsage)
	case "bool":
		x := fValue.Addr().Interface().(*bool)
		*x = defaultVal.(bool)
		//flagset.BoolVar(x, flagName, defaultVal.(bool), flagUsage)
	case "time.Duration":
		x := fValue.Addr().Interface().(*time.Duration)
		*x = defaultVal.(time.Duration)
		//flagset.DurationVar(x, flagName, defaultVal.(time.Duration), flagUsage)
	default:
		return fmt.Errorf("unsuported struct type %s", field.Type.String())
	}

	return nil
}

func reflectStruct(v reflect.Value, pfx string) error {
	val := v.Elem()

	for i := 0; i < val.NumField(); i++ {
		fValue := val.Field(i)
		field := val.Type().Field(i)
		if !fValue.CanInterface() || !fValue.CanSet() {
			// Unexported struct field, can't reflect the interface.
			// Quietly ignore like json marshalling.
			continue
		}

		if err := setDefault(fValue, field, pfx); err != nil {
			return err
		}
	}

	return nil
}
