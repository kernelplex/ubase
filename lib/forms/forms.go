package forms

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/kernelplex/ubase/lib/ubvalidation"
)

var timeParsers = []string{
	time.RFC3339Nano,
	time.RFC3339,
	"2006-01-02T15:04",
	"2006-01-02 15:04:05",
	"2006-01-02",
}

// ParseFormToStruct parses r.Form (URL query + body form) into dest,
// using each field's `json:"name"` tag for lookup. Fields without a json
// tag fall back to the field name (case-sensitive). Supported types:
// string, all signed ints, and time.Time (via common layouts above).
// Pointer fields to these types are also supported.
// dest must be a pointer to a struct.
func ParseFormToStruct(r *http.Request, dest any) error {
	validationErrTracker := ubvalidation.NewValidationTracker()
	if dest == nil {
		return errors.New("dest cannot be nil")
	}
	rv := reflect.ValueOf(dest)
	if rv.Kind() != reflect.Ptr || rv.Elem().Kind() != reflect.Struct {
		return errors.New("dest must be a pointer to a struct")
	}
	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("parse form: %w", err)
	}

	// collect all (trimmed) values for each key to support slices
	form := map[string][]string{}
	for k, vs := range r.Form {
		cleaned := make([]string, 0, len(vs))
		for _, v := range vs {
			v = strings.TrimSpace(v)
			if v != "" {
				cleaned = append(cleaned, v)
			}
		}
		if len(cleaned) > 0 {
			form[k] = cleaned
		}
	}

	return populateStruct(form, rv.Elem(), validationErrTracker, "")
}

func populateStruct(form map[string][]string, sv reflect.Value, tracker *ubvalidation.ValidationTracker, prefix string) error {
	st := sv.Type()
	for i := 0; i < st.NumField(); i++ {
		sf := st.Field(i)
		fv := sv.Field(i)

		// Skip unexported or explicitly ignored fields.
		if sf.PkgPath != "" { // unexported
			continue
		}
		if tag := sf.Tag.Get("json"); tag == "-" {
			continue
		}

		// If it's an embedded struct, recurse (no tag lookup for embedding).
		if sf.Anonymous && fv.Kind() == reflect.Struct && fv.CanSet() {
			if err := populateStruct(form, fv, tracker, prefix); err != nil {
				return err
			}
			continue
		}

		// Determine full key for this field.
		fieldKey := prefix
		if prefix != "" {
			fieldKey += "."
		}
		jsonName := jsonFieldName(sf)
		if jsonName != "" {
			fieldKey += jsonName
		} else {
			fieldKey += sf.Name
		}

		// Handle struct fields (including time.Time and nested structs).
		if fv.Kind() == reflect.Struct || (fv.Kind() == reflect.Pointer && fv.Type().Elem().Kind() == reflect.Struct) {
			if fv.Kind() == reflect.Pointer {
				if fv.IsNil() {
					et := fv.Type().Elem()
					fv.Set(reflect.New(et))
				}
				fv = fv.Elem()
			}
			// Now fv is the struct value.
			if fv.Type() == reflect.TypeOf(time.Time{}) {
				rawVals, ok := form[fieldKey]
				if !ok || len(rawVals) == 0 {
					continue
				}
				t, err := parseTime(rawVals[0])
				if err != nil {
					tracker.AddIssue(fieldKey, err.Error())
					return fmt.Errorf("field %q: %w", fieldKey, err)
				}
				fv.Set(reflect.ValueOf(t))
				continue
			}
			// Recurse into nested struct.
			if err := populateStruct(form, fv, tracker, fieldKey); err != nil {
				return err
			}
			continue
		}

		// For other fields (primitives, pointers to primitives, slices), look up and set.
		rawVals, ok := form[fieldKey]
		if !ok || len(rawVals) == 0 {
			continue // nothing to set
		}

		if err := setValue(fv, sf, rawVals); err != nil {
			tracker.AddIssue(fieldKey, err.Error())
			return fmt.Errorf("field %q: %w", fieldKey, err)
		}
	}
	return nil
}

func jsonFieldName(sf reflect.StructField) string {
	tag := sf.Tag.Get("json")
	if tag == "" {
		return ""
	}
	name, _, _ := strings.Cut(tag, ",")
	return name
}

func setValue(fv reflect.Value, sf reflect.StructField, rawVals []string) error {
	// Handle pointers by allocating and setting the element.
	if fv.Kind() == reflect.Pointer {
		elemT := fv.Type().Elem()
		elemV := reflect.New(elemT).Elem()
		if err := setValue(elemV, reflect.StructField{Type: elemT, Name: sf.Name, Tag: sf.Tag}, rawVals); err != nil {
			return err
		}
		fv.Set(elemV.Addr())
		return nil
	}

	// Helper to select first value for scalar fields
	first := func() string {
		if len(rawVals) == 0 {
			return ""
		}
		return rawVals[0]
	}

	switch fv.Kind() {
	case reflect.String:
		fv.SetString(first())
		return nil

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		// Special-case time.Time which is a struct, so not here.
		val, err := strconv.ParseInt(first(), 10, fv.Type().Bits())
		if err != nil {
			return fmt.Errorf("invalid integer %q: %w", first(), err)
		}
		fv.SetInt(val)
		return nil

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		val, err := strconv.ParseUint(first(), 10, fv.Type().Bits())
		if err != nil {
			return fmt.Errorf("invalid unsigned integer %q: %w", first(), err)
		}
		fv.SetUint(val)
		return nil

	case reflect.Float32, reflect.Float64:
		val, err := strconv.ParseFloat(first(), fv.Type().Bits())
		if err != nil {
			return fmt.Errorf("invalid float %q: %w", first(), err)
		}
		fv.SetFloat(val)
		return nil

	case reflect.Bool:
		b, ok := parseBool(first())
		if !ok {
			return fmt.Errorf("invalid bool %q", first())
		}
		fv.SetBool(b)
		return nil

	case reflect.Slice:
		elemT := fv.Type().Elem()
		// allocate result slice
		out := reflect.MakeSlice(fv.Type(), 0, len(rawVals))
		for idx, svRaw := range rawVals {
			// create a new element value
			ev := reflect.New(elemT).Elem()
			if err := setValue(ev, reflect.StructField{Type: elemT, Name: sf.Name, Tag: sf.Tag}, []string{svRaw}); err != nil {
				return fmt.Errorf("invalid slice element at index %d: %w", idx, err)
			}
			out = reflect.Append(out, ev)
		}
		fv.Set(out)
		return nil
	}

	return fmt.Errorf("unsupported type %s", fv.Type())
}

func parseTime(s string) (time.Time, error) {
	for _, layout := range timeParsers {
		if t, err := time.Parse(layout, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("invalid time %q (expected RFC3339, RFC3339Nano, '2006-01-02 15:04:05', or '2006-01-02')", s)
}

func parseBool(s string) (bool, bool) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "1", "t", "true", "yes", "y", "on":
		return true, true
	case "0", "f", "false", "no", "n", "off":
		return false, true
	default:
		return false, false
	}
}
