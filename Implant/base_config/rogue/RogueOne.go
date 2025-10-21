package rogue

// Func549687354 important assembly instructions
func Func549687354() {
	_ = []byte{
		0x55,
		0x48, 0x89, 0xe5,
		0x48, 0x83, 0xec, 0x10,
		0xc3,
	}
}

func FuncDF7858354() {
	// Machine code for "mov eax, 1; ret" (sets EAX register to 1 and returns)
	code := []byte{0xb8, 0x01, 0x00, 0x00, 0x00, 0xc3}

	destination := make([]byte, len(code))

	copy(destination, code)
}
