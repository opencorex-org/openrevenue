package id

import (
	"crypto/rand"
	"fmt"
)

type ID[T any] string

func New[T any]() ID[T] {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		panic("secure ID generation failed: " + err.Error())
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return ID[T](fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16]))
}
func (v ID[T]) String() string { return string(v) }
