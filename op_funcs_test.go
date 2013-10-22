package rhmrm

import "testing"

func TestInt10_To_Int16(t *testing.T) {
	t.Parallel()
	if x := int10_to_int16(0x3ff); x != -1 {
		t.Errorf("Result is %v, want %v", x, -1)
	}
	if x := int10_to_int16(0x7000); x != 0 {
		t.Errorf("Result is %v, want %v", x, 0)
	}
}
