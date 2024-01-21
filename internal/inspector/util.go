package inspector

import "unsafe"


func unsafeS2B(s string) []byte {
	return unsafe.Slice(unsafe.StringData(s), len(s))
}

func unsafeB2S(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}
