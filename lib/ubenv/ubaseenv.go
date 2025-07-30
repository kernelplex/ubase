package ubenv

import (
	"encoding/base64"
	"fmt"
	"os"
	"reflect"
	"strconv"
)

const (
	DefaultTag  = "default"
	EnvTag      = "env"
	RequiredTag = "required"
)

func ConfigFromEnv(cfg any) error {
	// Get the value pointed to by the interface
	cfgValue := reflect.ValueOf(cfg)
	if cfgValue.Kind() == reflect.Ptr {
		cfgValue = cfgValue.Elem()
	}

	cfgType := cfgValue.Type()

	for i := range cfgType.NumField() {
		field := cfgType.Field(i)
		fieldValue := cfgValue.Field(i)

		kind := fieldValue.Kind()
		// Process sub structs
		if kind == reflect.Struct {
			err := ConfigFromEnv(fieldValue.Addr().Interface())
			if err != nil {
				return fmt.Errorf("failed to load config from environment: %w", err)
			}
			continue
		}

		envTag := field.Tag.Get(EnvTag)
		if envTag == "" {
			continue
		}

		required := field.Tag.Get(RequiredTag) == "true"
		defaultValue := field.Tag.Get(DefaultTag)
		envValue, exists := getEnv(envTag, defaultValue)

		if !exists && envValue == "" {
			if required {
				return fmt.Errorf("required environment variable %s not found", envTag)
			}
			continue
		}

		switch fieldValue.Kind() {
		case reflect.String:
			if !fieldValue.CanSet() {
				return fmt.Errorf("cannot set string field %s", field.Name)
			}
			fieldValue.SetString(envValue)
		case reflect.Int:
		case reflect.Int64:
			{
				intValue, err := strconv.Atoi(envValue)
				if err != nil {
					return fmt.Errorf("invalid %s value: %w", envTag, err)
				}
				fieldValue.SetInt(int64(intValue))
			}
		case reflect.Slice: // Handle []byte fields (Pepper and SecretKey)
			if fieldValue.Type().Elem().Kind() == reflect.Uint8 && envValue != "" {
				decoded, err := base64.StdEncoding.DecodeString(envValue)
				if err != nil {
					return fmt.Errorf("invalid %s value: %w", envTag, err)
				}
				fieldValue.SetBytes(decoded)
			} else {
				return fmt.Errorf("invalid %s value: must be base64 encoded", envTag)
			}
		default:
			return fmt.Errorf("unsupported field type %s for field %s", fieldValue.Kind(), field.Name)
		}
	}

	return nil
}

// getEnv remains as a private helper function
func getEnv(key, defaultValue string) (string, bool) {
	value := defaultValue
	exists := false
	newValue, exists := os.LookupEnv(key)
	if exists {
		value = newValue
	}

	return value, exists
}
