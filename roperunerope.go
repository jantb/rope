package rope

import "math"

type RopeRuneRope struct {
	height   int
	weight   int
	left     *RopeRuneRope
	right    *RopeRuneRope
	content  []RuneRope
	balanced bool
}

var MaxLengthPerNodeRuneRope = 512

// NewFromBytes genearte new rope from bytes
func NewFromRuneRope(bs []RuneRope) (ret *RopeRuneRope) {
	if len(bs) == 0 {
		return nil
	}
	slots := make([]*RopeRuneRope, 32)
	var slotIndex int
	var r *RopeRuneRope
	for blockIndex := 0; blockIndex < len(bs)/MaxLengthPerNodeRope; blockIndex++ {
		r = &RopeRuneRope{
			height:   1,
			weight:   MaxLengthPerNodeRope,
			content:  bs[blockIndex*MaxLengthPerNodeRope : (blockIndex+1)*MaxLengthPerNodeRope],
			balanced: true,
		}
		slotIndex = 0
		for slots[slotIndex] != nil {
			r = &RopeRuneRope{
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
		ret = &RopeRuneRope{
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
func (r *RopeRuneRope) Index(row int) RuneRope {
	if row >= r.weight {
		return r.right.Index(row - r.weight)
	}
	if r.left != nil { // non leaf
		return r.left.Index(row)
	}
	// leaf
	return r.content[row]
}

// Len returns the length of the rope
func (r *RopeRuneRope) Len() int {
	if r == nil {
		return 0
	}
	return r.weight + r.right.Len()
}

// Bytes return all the bytes in the rope
func (r *RopeRuneRope) Bytes() []byte {
	i := 0
	l := 0
	r.Iter(0, func(bs []RuneRope) bool {
		for _, r := range bs {
			l += r.Len()
		}

		return true
	})
	ret := make([]byte, l)

	r.Iter(0, func(bs []RuneRope) bool {
		for _, r := range bs {
			b := []byte(string(r.runes()))
			copy(ret[i:], b)
			i += len(b)
		}

		return true
	})
	return ret
}

// Concat concatinates two RopeRuneRopes
func (r *RopeRuneRope) Concat(r2 *RopeRuneRope) (ret *RopeRuneRope) {
	ret = &RopeRuneRope{
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

func (r *RopeRuneRope) rebalance() (ret *RopeRuneRope) {
	var currentBytes []RuneRope
	slots := make([]*RopeRuneRope, 32)
	r.iterNodes(func(node *RopeRuneRope) bool {
		var balancedNode *RopeRuneRope
		iterSubNodes := true
		if len(currentBytes) == 0 && node.balanced { // balanced, insert to slots
			balancedNode = node
			iterSubNodes = false
		} else { // collect bytes
			currentBytes = append(currentBytes, node.content...)
			if len(currentBytes) >= MaxLengthPerNodeRope { // a full leaf
				balancedNode = &RopeRuneRope{
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
				balancedNode = &RopeRuneRope{
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
		ret = &RopeRuneRope{
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

func (r *RopeRuneRope) Split(n int) (out1, out2 *RopeRuneRope) {
	if r == nil {
		return
	}
	if len(r.content) > 0 { // leaf
		if n > len(r.content) { // offset overflow
			n = len(r.content)
		}
		out1 = NewFromRuneRope(r.content[:n])
		out2 = NewFromRuneRope(r.content[n:])
	} else { // non leaf
		var r1 *RopeRuneRope
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

func (r *RopeRuneRope) Insert(n int, bs []RuneRope) *RopeRuneRope {
	r1, r2 := r.Split(n)
	return r1.Concat(NewFromRuneRope(bs)).Concat(r2)
}

func (r *RopeRuneRope) Delete(n, l int) *RopeRuneRope {
	r1, r2 := r.Split(n)
	_, r2 = r2.Split(l)
	return r1.Concat(r2)
}

// Sub returns a substring of the rope
func (r *RopeRuneRope) Sub(n, l int) []RuneRope {
	ret := make([]RuneRope, l)
	i := 0
	r.Iter(n, func(bs []RuneRope) bool {
		if l >= len(bs) {
			copy(ret[i:], bs)
			i += len(bs)
			l -= len(bs)
			return true
		}
		copy(ret[i:], bs[:l])
		i += l
		return false
	})
	return ret[:i]
}

func (r *RopeRuneRope) Iter(offset int, fn func([]RuneRope) bool) bool {
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

func (r *RopeRuneRope) IterBackward(offset int, fn func([]RuneRope) bool) bool {
	if r == nil {
		return true
	}
	if len(r.content) > 0 { // leaf
		content := r.content[:offset]
		if len(content) == 0 {
			return true
		}
		bs := reversedRuneRopes(content)
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

func (r *RopeRuneRope) iterNodes(fn func(*RopeRuneRope) bool) {
	if r == nil {
		return
	}
	if fn(r) {
		r.left.iterNodes(fn)
		r.right.iterNodes(fn)
	}
}

func (r *RopeRuneRope) IterRune(offset int, fn func(rune, int) bool) {
	var bs []RuneRope
	r.Iter(offset, func(slice []RuneRope) bool {
		bs = append(bs, slice...)
		return true
	})
}
