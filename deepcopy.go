package gocache

import (
	"reflect"
)

// deepCopy creates a deep copy of the given value to prevent mutation of cached data
func deepCopy(src interface{}) interface{} {
	if src == nil {
		return nil
	}
	// Fast path: skip deep copy for immutable types
	switch src.(type) {
	case bool, int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64, uintptr,
		float32, float64, complex64, complex128,
		string:
		return src // These types are immutable, no need to copy
	}
	// Get the value and type
	original := reflect.ValueOf(src)
	// Handle nil pointers - preserve the typed nil
	if original.Kind() == reflect.Ptr && original.IsNil() {
		// Return a typed nil (same type, but nil value)
		return reflect.Zero(original.Type()).Interface()
	}
	// Handle pointers specially to maintain pointer semantics
	if original.Kind() == reflect.Ptr {
		// Create a new pointer to a copy of the pointed value
		pointedCopy := reflect.New(original.Elem().Type())
		deepCopyRecursive(pointedCopy.Elem(), original.Elem())
		return pointedCopy.Interface()
	}
	// For non-pointer types, create a copy
	copied := reflect.New(original.Type()).Elem()
	deepCopyRecursive(copied, original)
	return copied.Interface()
}

func deepCopyRecursive(dst, src reflect.Value) {
	switch src.Kind() {
	case reflect.Ptr:
		if !src.IsNil() {
			// Create a new pointer and copy the pointed value
			dst.Set(reflect.New(src.Type().Elem()))
			deepCopyRecursive(dst.Elem(), src.Elem())
		}
	case reflect.Interface:
		if !src.IsNil() {
			// Copy the underlying value
			dst.Set(reflect.ValueOf(deepCopy(src.Interface())))
		}
	case reflect.Struct:
		// Check if this struct has any fields that can be copied individually
		hasSettableFields := false
		hasUnsettableFields := false
		for i := 0; i < src.NumField(); i++ {
			if dst.Field(i).CanSet() {
				hasSettableFields = true
			} else {
				hasUnsettableFields = true
			}
		}
		// If struct has only unexported/unsettable fields, use direct assignment
		if hasUnsettableFields && !hasSettableFields {
			dst.Set(src)
		} else if hasSettableFields && !hasUnsettableFields {
			// All fields can be copied individually - do deep copy
			for i := 0; i < src.NumField(); i++ {
				deepCopyRecursive(dst.Field(i), src.Field(i))
			}
		} else {
			// Mixed case: has both settable and unsettable fields
			// This is tricky - we can't do proper deep copying while preserving unexported fields
			// For now, use direct assignment to preserve data integrity
			dst.Set(src)
		}
	case reflect.Slice:
		if !src.IsNil() {
			// Create a new slice with the same length and capacity
			dst.Set(reflect.MakeSlice(src.Type(), src.Len(), src.Cap()))
			for i := 0; i < src.Len(); i++ {
				deepCopyRecursive(dst.Index(i), src.Index(i))
			}
		}
	case reflect.Array:
		// Copy each element
		for i := 0; i < src.Len(); i++ {
			deepCopyRecursive(dst.Index(i), src.Index(i))
		}
	case reflect.Map:
		if !src.IsNil() {
			// Create a new map
			dst.Set(reflect.MakeMapWithSize(src.Type(), src.Len()))
			for _, key := range src.MapKeys() {
				// Deep copy both key and value
				keyType := src.Type().Key()
				valueType := src.Type().Elem()
				copiedKey := reflect.New(keyType).Elem()
				deepCopyRecursive(copiedKey, key)
				copiedValue := reflect.New(valueType).Elem()
				deepCopyRecursive(copiedValue, src.MapIndex(key))
				dst.SetMapIndex(copiedKey, copiedValue)
			}
		}
	case reflect.Chan:
		// Channels cannot be deep copied meaningfully
		// Just copy the channel reference
		dst.Set(src)
	case reflect.Func:
		// Functions are immutable, safe to copy reference
		dst.Set(src)
	default:
		// For primitive types (int, string, bool, etc.), just copy the value
		if src.IsValid() && dst.CanSet() {
			dst.Set(src)
		}
	}
}
