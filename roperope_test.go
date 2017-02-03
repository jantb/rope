package rope

import "testing"

func TestNewFromRopes(t *testing.T) {
	// nil bytes
	rr := NewFromRope([]Rope{})
	if rr != nil {
		t.Fatal()
	}

	if rr.Len() != 0 {
		t.Fatal()
	}
	r := NewFromBytes([]byte("Hello from rope"))
	if r == nil {
		t.Fatal()
	}
	rr = NewFromRope([]Rope{*r})

	bytes := rr.Bytes()
	if string(bytes) != "Hello from rope" {
		t.Fatal(string(bytes))
	}
}
func TestInsertRopeRope(t *testing.T) {
	r := NewFromBytes([]byte("Hello from rope"))
	if r == nil {
		t.Fatal()
	}
	rr := NewFromRope([]Rope{*r})
	rr = rr.Insert(0, []Rope{*r})
	if string(rr.Bytes()) != "Hello from ropeHello from rope" {
		t.Fatal(string(rr.Bytes()))
	}
}
