package rhmrm

import "testing"

func TestByteString(t *testing.T) {
	b := Byte(9001)
	if bs := b.String(); bs != "2329" {
		t.Errorf("Bytestring is %v, want 2329", bs)
	}
}

func TestWordString(t *testing.T) {
	w := Word(9001)
	if ws := w.String(); ws != "2329" {
		t.Errorf("Wordstring is %v, want 2329", ws)
	}
}

func TestRegisterString(t *testing.T) {
	var r Register
	r.Set(9001)
	
	if rs := r.String(); rs != "2329" {
		t.Errorf("Registerstring is %v, want 2329", rs)
	}
}
