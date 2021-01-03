package gocache

import (
	"fmt"
	"testing"
)

func TestEntry_SizeInBytes(t *testing.T) {
	testSizeInBytes(t, "key", 0, 75)
	testSizeInBytes(t, "k", 0, 73)
	testSizeInBytes(t, "k", "v", 66)
	testSizeInBytes(t, "k", true, 66)
	testSizeInBytes(t, "k", int8(1), 66)
	testSizeInBytes(t, "k", uint8(1), 66)
	testSizeInBytes(t, "k", true, 66)
	testSizeInBytes(t, "k", int16(1), 67)
	testSizeInBytes(t, "k", uint16(1), 67)
	testSizeInBytes(t, "k", int32(1), 69)
	testSizeInBytes(t, "k", uint32(1), 69)
	testSizeInBytes(t, "k", float32(1), 69)
	testSizeInBytes(t, "k", complex64(1), 69)
	testSizeInBytes(t, "k", int64(1), 73)
	testSizeInBytes(t, "k", uint64(1), 73)
	testSizeInBytes(t, "k", 1, 73)
	testSizeInBytes(t, "k", uint(1), 73)
	testSizeInBytes(t, "k", float64(1), 73)
	testSizeInBytes(t, "k", complex128(1), 73)
	testSizeInBytes(t, "k", []string{}, 65)
	testSizeInBytes(t, "k", []string{"what"}, 85)
	testSizeInBytes(t, "k", []string{"what", "the"}, 104)
	testSizeInBytes(t, "k", []int8{}, 65)
	testSizeInBytes(t, "k", []int8{1}, 66)
	testSizeInBytes(t, "k", []int8{1, 2}, 67)
	testSizeInBytes(t, "k", []uint8{1}, 66)
	testSizeInBytes(t, "k", []uint8{1, 2}, 67)
	testSizeInBytes(t, "k", []bool{true}, 66)
	testSizeInBytes(t, "k", []bool{true, false}, 67)
	testSizeInBytes(t, "k", []int16{1}, 67)
	testSizeInBytes(t, "k", []int16{1, 2}, 69)
	testSizeInBytes(t, "k", []uint16{1}, 67)
	testSizeInBytes(t, "k", []int32{1}, 69)
	testSizeInBytes(t, "k", []int32{1, 2}, 73)
	testSizeInBytes(t, "k", []uint32{1}, 69)
	testSizeInBytes(t, "k", []uint32{1, 2}, 73)
	testSizeInBytes(t, "k", []float32{1}, 69)
	testSizeInBytes(t, "k", []float32{1, 2}, 73)
	testSizeInBytes(t, "k", []complex64{1}, 69)
	testSizeInBytes(t, "k", []complex64{1, 2}, 73)
	testSizeInBytes(t, "k", []int64{1}, 73)
	testSizeInBytes(t, "k", []int64{1, 2}, 81)
	testSizeInBytes(t, "k", []uint64{1}, 73)
	testSizeInBytes(t, "k", []uint64{1, 2}, 81)
	testSizeInBytes(t, "k", []int{1}, 73)
	testSizeInBytes(t, "k", []int{1, 2}, 81)
	testSizeInBytes(t, "k", []uint{1}, 73)
	testSizeInBytes(t, "k", []uint{1, 2}, 81)
	testSizeInBytes(t, "k", []float64{1}, 73)
	testSizeInBytes(t, "k", []float64{1, 2}, 81)
	testSizeInBytes(t, "k", []complex128{1}, 73)
	testSizeInBytes(t, "k", []complex128{1, 2}, 81)
	testSizeInBytes(t, "k", struct{}{}, 67)
	testSizeInBytes(t, "k", struct{ A string }{A: "hello"}, 72)
	testSizeInBytes(t, "k", struct{ A, B string }{A: "hello", B: "world"}, 78)
	testSizeInBytes(t, "k", nil, 70)
	testSizeInBytes(t, "k", make([]interface{}, 5), 170)
}

func testSizeInBytes(t *testing.T, key string, value interface{}, expectedSize int) {
	t.Run(fmt.Sprintf("%T_%d", value, expectedSize), func(t *testing.T) {
		if size := (&Entry{Key: key, Value: value}).SizeInBytes(); size != expectedSize {
			t.Errorf("expected size of entry with key '%v' and value '%v' (%T) to be %d, got %d", key, value, value, expectedSize, size)
		}
	})
}
