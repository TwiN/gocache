package gocache

import (
	"reflect"
	"testing"
)

// Test for Array types
func TestDeepCopy_Array(t *testing.T) {
	// Test array of primitives
	original := [3]int{1, 2, 3}
	copied := deepCopy(original)

	if !reflect.DeepEqual(original, copied) {
		t.Errorf("Array copy failed: expected %v, got %v", original, copied)
	}

	// Verify it's a deep copy by checking type
	if reflect.TypeOf(copied) != reflect.TypeOf(original) {
		t.Errorf("Array copy type mismatch: expected %T, got %T", original, copied)
	}

	// Test array of structs
	type TestStruct struct {
		Name  string
		Value int
		Data  []string
	}
	structArray := [2]TestStruct{
		{Name: "first", Value: 1, Data: []string{"a", "b"}},
		{Name: "second", Value: 2, Data: []string{"c", "d"}},
	}
	copiedStructArray := deepCopy(structArray)

	if !reflect.DeepEqual(structArray, copiedStructArray) {
		t.Errorf("Struct array copy failed: expected %v, got %v", structArray, copiedStructArray)
	}

	// Verify deep copy by modifying original
	copiedArray := copiedStructArray.([2]TestStruct)
	copiedArray[0].Data[0] = "modified"
	if structArray[0].Data[0] == "modified" {
		t.Error("Array was not deep copied - original was modified")
	}
}

// Test for Channel types
func TestDeepCopy_Chan(t *testing.T) {
	// Test unbuffered channel
	original := make(chan int)
	copied := deepCopy(original)

	// Channels cannot be deep copied meaningfully, should copy reference
	if copied != original {
		t.Error("Channel should copy reference, not create new channel")
	}

	// Test buffered channel
	bufferedChan := make(chan string, 5)
	bufferedChan <- "test"
	copiedBuffered := deepCopy(bufferedChan)

	if copiedBuffered != bufferedChan {
		t.Error("Buffered channel should copy reference")
	}

	// Verify same channel by receiving from copied
	select {
	case val := <-copiedBuffered.(chan string):
		if val != "test" {
			t.Errorf("Expected 'test', got %s", val)
		}
	default:
		t.Error("Should be able to receive from copied channel")
	}
}

// Test for Function types
func TestDeepCopy_Func(t *testing.T) {
	// Test simple function
	original := func(x int) int { return x * 2 }
	copied := deepCopy(original)

	// Functions are immutable, should copy reference
	originalVal := reflect.ValueOf(original)
	copiedVal := reflect.ValueOf(copied)

	if originalVal.Pointer() != copiedVal.Pointer() {
		t.Error("Function should copy reference, not create new function")
	}

	// Test function execution
	result := copied.(func(int) int)(5)
	if result != 10 {
		t.Errorf("Expected 10, got %d", result)
	}

	// Test nil function
	var nilFunc func()
	copiedNilFunc := deepCopy(nilFunc)
	if !reflect.DeepEqual(nilFunc, copiedNilFunc) {
		t.Error("Nil function should remain equivalent to nil")
	}
	if !reflect.ValueOf(copiedNilFunc).IsNil() {
		t.Error("Copied nil function should be nil")
	}
}

// Test for Pointer types
func TestDeepCopy_Ptr(t *testing.T) {
	// Test pointer to primitive
	original := 42
	ptr := &original
	copied := deepCopy(ptr)

	copiedPtr := copied.(*int)
	if *copiedPtr != 42 {
		t.Errorf("Expected 42, got %d", *copiedPtr)
	}

	// Verify it's a deep copy - modifying copied shouldn't affect original
	*copiedPtr = 100
	if original != 42 {
		t.Error("Original value was modified - not a deep copy")
	}

	// Test pointer to struct
	type TestStruct struct {
		Name  string
		Value int
		Data  []string
	}
	structPtr := &TestStruct{
		Name:  "test",
		Value: 123,
		Data:  []string{"x", "y", "z"},
	}
	copiedStructPtr := deepCopy(structPtr)

	copied_struct := copiedStructPtr.(*TestStruct)
	if !reflect.DeepEqual(*structPtr, *copied_struct) {
		t.Error("Struct pointer content not copied correctly")
	}

	// Verify deep copy of nested slice
	copied_struct.Data[0] = "modified"
	if structPtr.Data[0] == "modified" {
		t.Error("Original struct slice was modified - not a deep copy")
	}

	// Test nil pointer
	var nilPtr *int
	copiedNilPtr := deepCopy(nilPtr)
	if !reflect.DeepEqual(nilPtr, copiedNilPtr) {
		t.Error("Nil pointer should remain equivalent to nil")
	}
	if !reflect.ValueOf(copiedNilPtr).IsNil() {
		t.Error("Copied nil pointer should be nil")
	}
	// Verify type is preserved for nil pointers
	if reflect.TypeOf(copiedNilPtr) != reflect.TypeOf(nilPtr) {
		t.Error("Nil pointer type not preserved")
	}

	// Test typed nil pointer
	var typedNilPtr *TestStruct
	copiedTypedNil := deepCopy(typedNilPtr)
	if !reflect.DeepEqual(typedNilPtr, copiedTypedNil) {
		t.Error("Typed nil pointer should remain equivalent to nil")
	}
	if !reflect.ValueOf(copiedTypedNil).IsNil() {
		t.Error("Copied typed nil pointer should be nil")
	}
	// Verify type is preserved for typed nil pointers
	if reflect.TypeOf(copiedTypedNil) != reflect.TypeOf(typedNilPtr) {
		t.Error("Typed nil pointer type not preserved")
	}
}

// Test for Interface types
func TestDeepCopy_Interface(t *testing.T) {
	// Test nil interface
	var nilInterface interface{}
	copiedNilInterface := deepCopy(nilInterface)
	if copiedNilInterface != nil {
		t.Error("Nil interface should remain nil")
	}

	// Test empty interface with various types
	var emptyInterface interface{}

	// Test with struct
	type TestStruct struct {
		Name  string
		Value int
		Data  []string
	}
	emptyInterface = TestStruct{Name: "test", Value: 456, Data: []string{"a", "b"}}
	copiedEmpty := deepCopy(emptyInterface)
	if !reflect.DeepEqual(emptyInterface, copiedEmpty) {
		t.Error("Empty interface with struct not copied correctly")
	}

	// Test with slice
	emptyInterface = []int{1, 2, 3, 4}
	copiedSlice := deepCopy(emptyInterface)
	slice := copiedSlice.([]int)
	slice[0] = 999
	if emptyInterface.([]int)[0] == 999 {
		t.Error("Original slice in interface was modified - not a deep copy")
	}

	// Test with map
	emptyInterface = map[string]int{"a": 1, "b": 2}
	copiedMap := deepCopy(emptyInterface)
	mapVal := copiedMap.(map[string]int)
	mapVal["a"] = 999
	if emptyInterface.(map[string]int)["a"] == 999 {
		t.Error("Original map in interface was modified - not a deep copy")
	}

	// Test interface with concrete type containing pointer
	type StructWithPointer struct {
		Value *int
	}
	val := 42
	emptyInterface = StructWithPointer{Value: &val}
	copiedStructWithPtr := deepCopy(emptyInterface)
	structPtr := copiedStructWithPtr.(StructWithPointer)
	*structPtr.Value = 999
	if *emptyInterface.(StructWithPointer).Value == 999 {
		t.Error("Original struct pointer in interface was modified - not a deep copy")
	}
}

// Test complex nested structures combining multiple types
func TestDeepCopy_ComplexNested(t *testing.T) {
	type TestImpl struct {
		Data string
	}

	type ComplexStruct struct {
		StringArray [3]string
		IntPtr      *int
		Interface   interface{}
		Channel     chan bool
		Function    func(int) int
	}

	intVal := 42
	original := ComplexStruct{
		StringArray: [3]string{"a", "b", "c"},
		IntPtr:      &intVal,
		Interface:   TestImpl{Data: "interface data"},
		Channel:     make(chan bool, 1),
		Function:    func(x int) int { return x * 3 },
	}

	original.Channel <- true

	copied := deepCopy(original)
	copiedStruct := copied.(ComplexStruct)

	// Verify array was copied
	if !reflect.DeepEqual(original.StringArray, copiedStruct.StringArray) {
		t.Error("Array in complex struct not copied correctly")
	}

	// Verify pointer was deep copied
	*copiedStruct.IntPtr = 999
	if *original.IntPtr == 999 {
		t.Error("Pointer in complex struct was not deep copied")
	}

	// Verify interface was deep copied
	copiedImpl := copiedStruct.Interface.(TestImpl)
	copiedImpl.Data = "modified"
	if original.Interface.(TestImpl).Data == "modified" {
		t.Error("Interface in complex struct was not deep copied")
	}

	// Verify channel reference was copied (not deep copied)
	if copiedStruct.Channel != original.Channel {
		t.Error("Channel should copy reference")
	}

	// Verify function reference was copied
	originalFuncPtr := reflect.ValueOf(original.Function).Pointer()
	copiedFuncPtr := reflect.ValueOf(copiedStruct.Function).Pointer()
	if originalFuncPtr != copiedFuncPtr {
		t.Error("Function should copy reference")
	}
}

// Test edge cases
func TestDeepCopy_EdgeCases(t *testing.T) {
	// Test nil input
	result := deepCopy(nil)
	if result != nil {
		t.Error("deepCopy(nil) should return nil")
	}

	// Test empty array
	emptyArray := [0]int{}
	copiedEmpty := deepCopy(emptyArray)
	if !reflect.DeepEqual(emptyArray, copiedEmpty) {
		t.Error("Empty array not copied correctly")
	}

	// Test pointer to array
	arrayPtr := &[2]string{"x", "y"}
	copiedArrayPtr := deepCopy(arrayPtr)
	copiedArray := copiedArrayPtr.(*[2]string)
	copiedArray[0] = "modified"
	if arrayPtr[0] == "modified" {
		t.Error("Array through pointer was not deep copied")
	}

	// Test array of pointers
	val1, val2 := 10, 20
	ptrArray := [2]*int{&val1, &val2}
	copiedPtrArray := deepCopy(ptrArray).([2]*int)
	*copiedPtrArray[0] = 999
	if val1 == 999 {
		t.Error("Array of pointers was not deep copied")
	}
}
