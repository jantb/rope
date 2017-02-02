package rope

import "math"

type RopeRope struct {
	height   int
	weight   int
	left     *RopeRope
	right    *RopeRope
	content  []Rope
	balanced bool
}

var MaxLengthPerNodeRope = 128

// NewFromBytes genearte new rope from bytes
func NewFromBytesRope(bs []Rope) (ret *RopeRope) {
	if len(bs) == 0 {
		return nil
	}
	slots := make([]*RopeRope, 32)
	var slotIndex int
	var r *RopeRope
	for blockIndex := 0; blockIndex < len(bs)/MaxLengthPerNodeRope; blockIndex++ {
		r = &RopeRope{
			height:   1,
			weight:   MaxLengthPerNodeRope,
			content:  bs[blockIndex*MaxLengthPerNodeRope : (blockIndex+1)*MaxLengthPerNodeRope],
			balanced: true,
		}
		slotIndex = 0
		for slots[slotIndex] != nil {
			r = &RopeRope{
				height:   slotIndex + 2,
				weight:   (1 << uint(slotIndex)) * MaxLengthPerNodeRope,
				left:     slots[slotIndex],
				right:    r,
				balanced: true,
			}
			slots[slotIndex] = nil
			slotIndex++
		}
		slots[slotIndex] = r
	}
	tailStart := len(bs) / MaxLengthPerNodeRope * MaxLengthPerNodeRope
	if tailStart < len(bs) {
		ret = &RopeRope{
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

// Index returns rope at index
func (r *RopeRope) Index(i int) Rope {
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
func (r *RopeRope) Len() int {
	if r == nil {
		return 0
	}
	return r.weight + r.right.Len()
}

// Bytes return all the bytes in the rope
func (r *RopeRope) Bytes() []byte {
	ret := make([]byte, r.Len())
	i := 0
	r.Iter(0, func(bs []Rope) bool {
		for _, r := range bs {
			copy(ret[i:], r.Bytes())
			i += len(r.Bytes())
		}

		return true
	})
	return ret
}

// Concat concatinates two roperopes
func (r *RopeRope) Concat(r2 *RopeRope) (ret *RopeRope) {
	ret = &RopeRope{
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
		l := int((math.Ceil(math.Log2(float64((ret.Len()/MaxLengthPerNodeRope)+1))) + 1) * 1.5)
		if ret.height > l {
			ret = ret.rebalance()
		}
	}
	return
}

func (r *RopeRope) rebalance() (ret *RopeRope) {
	var currentBytes []Rope
	slots := make([]*RopeRope, 32)
	r.iterNodes(func(node *RopeRope) bool {
		var balancedNode *RopeRope
		iterSubNodes := true
		if len(currentBytes) == 0 && node.balanced { // balanced, insert to slots
			balancedNode = node
			iterSubNodes = false
		} else { // collect bytes
			currentBytes = append(currentBytes, node.content...)
			if len(currentBytes) >= MaxLengthPerNodeRope { // a full leaf
				balancedNode = &RopeRope{
					height:   1,
					weight:   MaxLengthPerNodeRope,
					balanced: true,
					content:  currentBytes[:MaxLengthPerNodeRope],
				}
				currentBytes = currentBytes[MaxLengthPerNodeRope:]
			}
		}
		if balancedNode != nil {
			slotIndex := balancedNode.height - 1
			for slots[slotIndex] != nil {
				balancedNode = &RopeRope{
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
		ret = &RopeRope{
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

func (r *RopeRope) Split(n int) (out1, out2 *RopeRope) {
	if r == nil {
		return
	}
	if len(r.content) > 0 { // leaf
		if n > len(r.content) { // offset overflow
			n = len(r.content)
		}
		out1 = NewFromBytesRope(r.content[:n])
		out2 = NewFromBytesRope(r.content[n:])
	} else { // non leaf
		var r1 *RopeRope
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

func (r *RopeRope) Insert(n int, bs []Rope) *RopeRope {
	r1, r2 := r.Split(n)
	return r1.Concat(NewFromBytesRope(bs)).Concat(r2)
}

func (r *RopeRope) Delete(n, l int) *RopeRope {
	r1, r2 := r.Split(n)
	_, r2 = r2.Split(l)
	return r1.Concat(r2)
}

// Sub returns a substring of the rope
// func (r *RopeRope) Sub(n, l int) []Rope {
// 	ret := make([]byte, l)
// 	i := 0
// 	r.Iter(n, func(bs []Rope) bool {
// 		if l >= len(bs) {
// 			copy(ret[i:], bs)
// 			i += len(bs)
// 			l -= len(bs)
// 			return true
// 		}
// 		copy(ret[i:], bs[:l])
// 		i += l
// 		return false
// 	})
// 	return ret[:i]
// }

func (r *RopeRope) Iter(offset int, fn func([]Rope) bool) bool {
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

func (r *RopeRope) IterBackward(offset int, fn func([]Rope) bool) bool {
	if r == nil {
		return true
	}
	if len(r.content) > 0 { // leaf
		content := r.content[:offset]
		if len(content) == 0 {
			return true
		}
		bs := reversedRopes(content)
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

func (r *RopeRope) iterNodes(fn func(*RopeRope) bool) {
	if r == nil {
		return
	}
	if fn(r) {
		r.left.iterNodes(fn)
		r.right.iterNodes(fn)
	}
}

func (r *RopeRope) IterRune(offset int, fn func(rune, int) bool) {
	var bs []Rope
	r.Iter(offset, func(slice []Rope) bool {
		bs = append(bs, slice...)
		return true
	})
}
