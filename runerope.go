package rope

import "math"

type RuneRope struct {
	height   int
	weight   int
	left     *RuneRope
	right    *RuneRope
	content  []rune
	balanced bool
}

var MaxLengthPerNodeRune = 128

// NewFromrunes genearte new Runerope from runes
func NewFromrunes(bs []rune) (ret *RuneRope) {
	if len(bs) == 0 {
		ret = &RuneRope{
			height:   0,
			weight:   0,
			content:  bs,
			balanced: false,
		}
		return
	}
	slots := make([]*RuneRope, 32)
	var slotIndex int
	var r *RuneRope
	for blockIndex := 0; blockIndex < len(bs)/MaxLengthPerNodeRune; blockIndex++ {
		r = &RuneRope{
			height:   1,
			weight:   MaxLengthPerNodeRune,
			content:  bs[blockIndex*MaxLengthPerNodeRune : (blockIndex+1)*MaxLengthPerNodeRune],
			balanced: true,
		}
		slotIndex = 0
		for slots[slotIndex] != nil {
			r = &RuneRope{
				height:   slotIndex + 2,
				weight:   (1 << uint(slotIndex)) * MaxLengthPerNodeRune,
				left:     slots[slotIndex],
				right:    r,
				balanced: true,
			}
			slots[slotIndex] = nil
			slotIndex++
		}
		slots[slotIndex] = r
	}
	tailStart := len(bs) / MaxLengthPerNodeRune * MaxLengthPerNodeRune
	if tailStart < len(bs) {
		ret = &RuneRope{
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
func (r *RuneRope) Index(i int) rune {
	if i >= r.weight {
		return r.right.Index(i - r.weight)
	}
	if r.left != nil { // non leaf
		return r.left.Index(i)
	}
	// leaf
	return r.content[i]
}

// Len returns the length of the Runerope
func (r *RuneRope) Len() int {
	if r == nil {
		return 0
	}
	return r.weight + r.right.Len()
}

// runes return all the runes in the Runerope
func (r *RuneRope) runes() []rune {
	ret := make([]rune, r.Len())
	i := 0
	r.Iter(0, func(bs []rune) bool {
		copy(ret[i:], bs)
		i += len(bs)
		return true
	})
	return ret
}

func (r *RuneRope) Concat(r2 *RuneRope) (ret *RuneRope) {
	ret = &RuneRope{
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
		l := int((math.Ceil(math.Log2(float64((ret.Len()/MaxLengthPerNodeRune)+1))) + 1) * 1.5)
		if ret.height > l {
			ret = ret.rebalance()
		}
	}
	return
}

func (r *RuneRope) rebalance() (ret *RuneRope) {
	var currentrunes []rune
	slots := make([]*RuneRope, 32)
	r.iterNodes(func(node *RuneRope) bool {
		var balancedNode *RuneRope
		iterSubNodes := true
		if len(currentrunes) == 0 && node.balanced { // balanced, insert to slots
			balancedNode = node
			iterSubNodes = false
		} else { // collect runes
			currentrunes = append(currentrunes, node.content...)
			if len(currentrunes) >= MaxLengthPerNodeRune { // a full leaf
				balancedNode = &RuneRope{
					height:   1,
					weight:   MaxLengthPerNodeRune,
					balanced: true,
					content:  currentrunes[:MaxLengthPerNodeRune],
				}
				currentrunes = currentrunes[MaxLengthPerNodeRune:]
			}
		}
		if balancedNode != nil {
			slotIndex := balancedNode.height - 1
			for slots[slotIndex] != nil {
				balancedNode = &RuneRope{
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
	if len(currentrunes) > 0 {
		ret = &RuneRope{
			height:   1,
			weight:   len(currentrunes),
			balanced: false,
			content:  currentrunes,
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

func (r *RuneRope) Split(n int) (out1, out2 *RuneRope) {
	if r == nil {
		return
	}
	if len(r.content) > 0 { // leaf
		if n > len(r.content) { // offset overflow
			n = len(r.content)
		}
		out1 = NewFromrunes(r.content[:n])
		out2 = NewFromrunes(r.content[n:])
	} else { // non leaf
		var r1 *RuneRope
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

func (r *RuneRope) Insert(n int, bs []rune) *RuneRope {
	r1, r2 := r.Split(n)
	return r1.Concat(NewFromrunes(bs)).Concat(r2)
}

func (r *RuneRope) Delete(n, l int) *RuneRope {
	r1, r2 := r.Split(n)
	_, r2 = r2.Split(l)
	return r1.Concat(r2)
}

// Sub returns a substring of the Runerope
func (r *RuneRope) Sub(n, l int) []rune {
	ret := make([]rune, l)
	i := 0
	r.Iter(n, func(bs []rune) bool {
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

func (r *RuneRope) Iter(offset int, fn func([]rune) bool) bool {
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

func (r *RuneRope) IterBackward(offset int, fn func([]rune) bool) bool {
	if r == nil {
		return true
	}
	if len(r.content) > 0 { // leaf
		content := r.content[:offset]
		if len(content) == 0 {
			return true
		}
		bs := reversedrunes(content)
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

func (r *RuneRope) iterNodes(fn func(*RuneRope) bool) {
	if r == nil {
		return
	}
	if fn(r) {
		r.left.iterNodes(fn)
		r.right.iterNodes(fn)
	}
}
