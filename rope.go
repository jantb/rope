package rope

import (
	"math"
	"unicode/utf8"
)

type Rope struct {
	height   int
	weight   int
	left     *Rope
	right    *Rope
	content  []byte
	balanced bool
}

var MaxLengthPerNode = 128

// NewFromBytes genearte new rope from bytes
func NewFromBytes(bs []byte) (ret *Rope) {
	if len(bs) == 0 {
		ret = &Rope{
			height:   0,
			weight:   0,
			content:  bs,
			balanced: false,
		}
		return
	}
	slots := make([]*Rope, 32)
	var slotIndex int
	var r *Rope
	for blockIndex := 0; blockIndex < len(bs)/MaxLengthPerNode; blockIndex++ {
		r = &Rope{
			height:   1,
			weight:   MaxLengthPerNode,
			content:  bs[blockIndex*MaxLengthPerNode : (blockIndex+1)*MaxLengthPerNode],
			balanced: true,
		}
		slotIndex = 0
		for slots[slotIndex] != nil {
			r = &Rope{
				height:   slotIndex + 2,
				weight:   (1 << uint(slotIndex)) * MaxLengthPerNode,
				left:     slots[slotIndex],
				right:    r,
				balanced: true,
			}
			slots[slotIndex] = nil
			slotIndex++
		}
		slots[slotIndex] = r
	}
	tailStart := len(bs) / MaxLengthPerNode * MaxLengthPerNode
	if tailStart < len(bs) {
		ret = &Rope{
			height:   1,
			weight:   len(bs) - tailStart,
			content:  bs[tailStart:],
			balanced: false,
		}
	}
	for _, c := range slots {
		if c != nil {
			if ret == nil {
				ret = c
			} else {
				ret = c.Concat(ret)
			}
		}
	}
	return
}

// Index returns byt at index
func (r *Rope) Index(i int) byte {
	if i >= r.weight {
		return r.right.Index(i - r.weight)
	}
	if r.left != nil { // non leaf
		return r.left.Index(i)
	}
	// leaf
	return r.content[i]
}

// Len returns the length of the rope
func (r *Rope) Len() int {
	if r == nil {
		return 0
	}
	return r.weight + r.right.Len()
}

// Bytes return all the bytes in the rope
func (r *Rope) Bytes() []byte {
	ret := make([]byte, r.Len())
	i := 0
	r.Iter(0, func(bs []byte) bool {
		copy(ret[i:], bs)
		i += len(bs)
		return true
	})
	return ret
}

func (r *Rope) Concat(r2 *Rope) (ret *Rope) {
	ret = &Rope{
		weight: r.Len(),
		left:   r,
		right:  r2,
	}
	if ret.left != nil {
		ret.height = ret.left.height
	}
	if ret.right != nil && ret.right.height > ret.height {
		ret.height = ret.right.height
	}
	if ret.left != nil && ret.left.balanced &&
		ret.right != nil && ret.right.balanced &&
		ret.left.height == ret.right.height {
		ret.balanced = true
	}
	ret.height++
	// check and rebalance
	if !ret.balanced {
		l := int((math.Ceil(math.Log2(float64((ret.Len()/MaxLengthPerNode)+1))) + 1) * 1.5)
		if ret.height > l {
			ret = ret.rebalance()
		}
	}
	return
}

func (r *Rope) rebalance() (ret *Rope) {
	var currentBytes []byte
	slots := make([]*Rope, 32)
	r.iterNodes(func(node *Rope) bool {
		var balancedNode *Rope
		iterSubNodes := true
		if len(currentBytes) == 0 && node.balanced { // balanced, insert to slots
			balancedNode = node
			iterSubNodes = false
		} else { // collect bytes
			currentBytes = append(currentBytes, node.content...)
			if len(currentBytes) >= MaxLengthPerNode { // a full leaf
				balancedNode = &Rope{
					height:   1,
					weight:   MaxLengthPerNode,
					balanced: true,
					content:  currentBytes[:MaxLengthPerNode],
				}
				currentBytes = currentBytes[MaxLengthPerNode:]
			}
		}
		if balancedNode != nil {
			slotIndex := balancedNode.height - 1
			for slots[slotIndex] != nil {
				balancedNode = &Rope{
					height:   balancedNode.height + 1,
					weight:   slots[slotIndex].Len(),
					left:     slots[slotIndex],
					right:    balancedNode,
					balanced: true,
				}
				slots[slotIndex] = nil
				slotIndex++
			}
			slots[slotIndex] = balancedNode
		}
		return iterSubNodes
	})
	if len(currentBytes) > 0 {
		ret = &Rope{
			height:   1,
			weight:   len(currentBytes),
			balanced: false,
			content:  currentBytes,
		}
	}
	for _, c := range slots {
		if c != nil {
			if ret == nil {
				ret = c
			} else {
				ret = c.Concat(ret)
			}
		}
	}
	return
}

func (r *Rope) Split(n int) (out1, out2 *Rope) {
	if r == nil {
		return
	}
	if len(r.content) > 0 { // leaf
		if n > len(r.content) { // offset overflow
			n = len(r.content)
		}
		out1 = NewFromBytes(r.content[:n])
		out2 = NewFromBytes(r.content[n:])
	} else { // non leaf
		var r1 *Rope
		if n >= r.weight { // at right subtree
			r1, out2 = r.right.Split(n - r.weight)
			out1 = r.left.Concat(r1)
		} else { // at left subtree
			out1, r1 = r.left.Split(n)
			out2 = r1.Concat(r.right)
		}
	}
	return
}

func (r *Rope) Insert(n int, bs []byte) *Rope {
	r1, r2 := r.Split(n)
	return r1.Concat(NewFromBytes(bs)).Concat(r2)
}

func (r *Rope) Delete(n, l int) *Rope {
	r1, r2 := r.Split(n)
	_, r2 = r2.Split(l)
	return r1.Concat(r2)
}

// Sub returns a substring of the rope
func (r *Rope) Sub(n, l int) []byte {
	ret := make([]byte, l)
	i := 0
	r.Iter(n, func(bs []byte) bool {
		if l >= len(bs) {
			copy(ret[i:], bs)
			i += len(bs)
			l -= len(bs)
			return true
		} else {
			copy(ret[i:], bs[:l])
			i += l
			return false
		}
	})
	return ret[:i]
}

func (r *Rope) Iter(offset int, fn func([]byte) bool) bool {
	if r == nil {
		return true
	}
	if len(r.content) > 0 { // leaf
		if offset < len(r.content) {
			if !fn(r.content[offset:]) {
				return false
			}
		}
	} else { // non leaf
		if offset >= r.weight { // start at right subtree
			if !r.right.Iter(offset-r.weight, fn) {
				return false
			}
		} else { // start at left subtree
			if !r.left.Iter(offset, fn) {
				return false
			}
			if !r.right.Iter(0, fn) {
				return false
			}
		}
	}
	return true
}

func (r *Rope) IterBackward(offset int, fn func([]byte) bool) bool {
	if r == nil {
		return true
	}
	if len(r.content) > 0 { // leaf
		content := r.content[:offset]
		if len(content) == 0 {
			return true
		}
		bs := reversedBytes(content)
		if !fn(bs) {
			return false
		}
	} else { // non leaf
		if offset >= r.weight { // start at right subtree
			if !r.right.IterBackward(offset-r.weight, fn) {
				return false
			}
			if !r.left.IterBackward(r.weight, fn) {
				return false
			}
		} else { // start at left subtree
			if !r.left.IterBackward(offset, fn) {
				return false
			}
		}
	}
	return true
}

func (r *Rope) iterNodes(fn func(*Rope) bool) {
	if r == nil {
		return
	}
	if fn(r) {
		r.left.iterNodes(fn)
		r.right.iterNodes(fn)
	}
}

func (r *Rope) IterRune(offset int, fn func(rune, int) bool) {
	var bs []byte
	stopped := false
	r.Iter(offset, func(slice []byte) bool {
		bs = append(bs, slice...)
		for len(bs) >= 4 {
			ru, l := utf8.DecodeRune(bs)
			bs = bs[l:]
			if ru == utf8.RuneError {
				return false
			}
			if !fn(ru, l) {
				stopped = true
				return false
			}
		}
		return true
	})
	if !stopped && len(bs) > 0 {
		for {
			ru, l := utf8.DecodeRune(bs)
			bs = bs[l:]
			if ru != utf8.RuneError {
				fn(ru, l)
			} else {
				break
			}
		}
	}
}
