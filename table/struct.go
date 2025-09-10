package table

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

// UnmarshalStruct unmarshals a table.Buffer into a slice of type T, matching fields by name or db tag.
// Field matching is case-sensitive. Extra columns are ignored. NULL values in non-pointer fields are set to zero values.
// If a field has a db tag with a "json" attribute (e.g., `db:"name,json"`), it is unmarshaled as a JSON array of objects.
func UnmarshalStruct[T any](buf *Buffer) ([]T, error) {
	if buf == nil {
		return nil, fmt.Errorf("buffer is nil")
	}

	var result []T
	tType := reflect.TypeOf((*T)(nil)).Elem()
	isPtr := tType.Kind() == reflect.Ptr
	if isPtr {
		tType = tType.Elem()
	}

	// Build field mapping: field index to column index.
	type fieldInfo struct {
		fieldIdx   int
		isJSON     bool
		jsonStruct reflect.Type
	}
	fieldMap := make(map[int]fieldInfo) // column index to field info.

	for i := 0; i < tType.NumField(); i++ {
		f := tType.Field(i)
		colIdx := buf.ColumnIndex(f.Name)
		if colIdx >= 0 {
			fieldMap[colIdx] = fieldInfo{fieldIdx: i}
			continue
		}

		// Check db tag.
		dbTag := f.Tag.Get("db")
		if dbTag != "" {
			parts := strings.Split(dbTag, ",")
			tagName := parts[0]
			if len(tagName) == 0 {
				tagName = f.Name
			}
			isJSON := false
			if len(parts) > 1 {
				for _, p := range parts[1:] {
					if p == "json" {
						isJSON = true
						break
					}
				}
			}
			colIdx = buf.ColumnIndex(tagName)
			if colIdx >= 0 {
				info := fieldInfo{fieldIdx: i, isJSON: isJSON}
				if isJSON {
					// Ensure field is a slice or array for JSON unmarshaling.
					if f.Type.Kind() == reflect.Slice || f.Type.Kind() == reflect.Array {
						info.jsonStruct = f.Type.Elem()
					} else {
						return nil, fmt.Errorf("field %s with db:%s,json tag must be a slice or array", f.Name, tagName)
					}
				}
				fieldMap[colIdx] = info
			}
		}
	}

	// Process each row.
	for _, row := range buf.Row {
		var obj T
		if isPtr {
			// Create new instance of T's underlying type.
			newObj := reflect.New(tType).Interface()
			obj = newObj.(T)
		} else {
			// Zero value of T.
			obj = reflect.New(tType).Elem().Interface().(T)
		}

		v := reflect.ValueOf(&obj).Elem()
		if isPtr {
			v = v.Elem()
		}

		for colIdx, info := range fieldMap {
			f := v.Field(info.fieldIdx)
			if !f.CanSet() {
				continue
			}

			val := row.Field[colIdx]
			if val.Null {
				continue
			}
			if info.isJSON {
				// Unmarshal JSON data.
				var jsonData []byte
				switch v := val.Value.(type) {
				default:
					return nil, fmt.Errorf("column %d: expected []byte for JSON field, got %T", colIdx, val.Value)
				case []byte:
					jsonData = v
				case string:
					jsonData = []byte(v)
				}
				sliceType := f.Type()
				slice := reflect.New(sliceType).Elem()
				tempSlice := reflect.New(reflect.SliceOf(info.jsonStruct)).Interface()
				if err := json.Unmarshal(jsonData, tempSlice); err != nil {
					return nil, fmt.Errorf("column %d: failed to unmarshal JSON: %w", colIdx, err)
				}
				slice.Set(reflect.ValueOf(tempSlice).Elem())
				f.Set(slice)
				continue
			}

			// Direct value assignment.
			fv := reflect.ValueOf(val.Value)
			if fv.Type().ConvertibleTo(f.Type()) {
				f.Set(fv.Convert(f.Type()))
			} else {
				return nil, fmt.Errorf("column %d: cannot convert %T to %s", colIdx, val.Value, f.Type())
			}
		}

		result = append(result, obj)
	}

	return result, nil
}
