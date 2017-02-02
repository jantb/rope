package rope

import (
	"bytes"
	"fmt"
	"strings"
)

var (
	p = fmt.Printf
)

func (r *Rope) StructEqual(r2 *Rope) bool {
	if r == nil && r2 == nil {
		return true
	}
	if r == nil && r2 != nil || r != nil && r2 == nil {
		return false
	}
	if r.weight != r2.weight {
		return false
	}
	if !(bytes.Equal(r.content, r2.content)) {
		return false
	}
	if !r.left.StructEqual(r2.left) {
		return false
	}
	if !r.right.StructEqual(r2.right) {
		return false
	}
	return true
}

func (r *Rope) Dump() {
	r.dump(0, "")
}

func (r *Rope) dump(level int, prefix string) {
	p("%s%s%d |%s|\n", strings.Repeat("  ", level), prefix, r.weight, r.content)
	if r.left != nil {
		r.left.dump(level+1, "<")
	}
	if r.right != nil {
		r.right.dump(level+1, ">")
	}
}

func reversedBytes(bs []byte) []byte {
	ret := make([]byte, len(bs))
	for i, b := range bs {
		ret[len(bs)-i-1] = b
	}
	return ret
}
func reversedRopes(bs []Rope) []Rope {
	ret := make([]Rope, len(bs))
	for i, b := range bs {
		ret[len(bs)-i-1] = b
	}
	return ret
}
