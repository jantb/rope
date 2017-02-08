package rope

import "math"

type Unium struct {
	height   int
	weight   int
	left     *Unium
	right    *Unium
	content  []interface{}
	balanced bool
}

var MaxLengthPerNodeinterface = 512

// NewFrominterface{} genearte new interface{} from bytes
func NewFrominterface(bs []interface{}) (ret *Unium) {
	if len(bs) == 0 {
		return nil
	}
	slots := make([]*Unium, 32)
	var slotIndex int
	var r *Unium
	for blockIndex := 0; blockIndex < len(bs)/MaxLengthPerNodeinterface; blockIndex++ {
		r = &Unium{
			height:   1,
			weight:   MaxLengthPerNodeinterface,
			content:  bs[blockIndex*MaxLengthPerNodeinterface : (blockIndex+1)*MaxLengthPerNodeinterface],
			balanced: true,
		}
		slotIndex = 0
		for slots[slotIndex] != nil {
			r = &Unium{
				height:   slotIndex + 2,
				weight:   (1 << uint(slotIndex)) * MaxLengthPerNodeinterface,
				left:     slots[slotIndex],
				right:    r,
				balanced: true,
			}
			slots[slotIndex] = nil
			slotIndex++
		}
		slots[slotIndex] = r
	}
	tailStart := len(bs) / MaxLengthPerNodeinterface * MaxLengthPerNodeinterface
	if tailStart < len(bs) {
		ret = &Unium{
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

// Index returns interface{} at index
func (r *Unium) Index(row int) interface{} {
	if row >= r.weight {
		return r.right.Index(row - r.weight)
	}
	if r.left != nil { // non leaf
		return r.left.Index(row)
	}
	// leaf
	return r.content[row]
}

// Len returns the length of the interface{}
func (r *Unium) Len() int {
	if r == nil {
		return 0
	}
	return r.weight + r.right.Len()
}

// Concat concatinates two Uniums
func (r *Unium) Concat(r2 *Unium) (ret *Unium) {
	ret = &Unium{
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
		l := int((math.Ceil(math.Log2(float64((ret.Len()/MaxLengthPerNodeinterface)+1))) + 1) * 1.5)
		if ret.height > l {
			ret = ret.rebalance()
		}
	}
	return
}

func (r *Unium) rebalance() (ret *Unium) {
	var currentBytes []interface{}
	slots := make([]*Unium, 32)
	r.iterNodes(func(node *Unium) bool {
		var balancedNode *Unium
		iterSubNodes := true
		if len(currentBytes) == 0 && node.balanced { // balanced, insert to slots
			balancedNode = node
			iterSubNodes = false
		} else { // collect bytes
			currentBytes = append(currentBytes, node.content...)
			if len(currentBytes) >= MaxLengthPerNodeinterface { // a full leaf
				balancedNode = &Unium{
					height:   1,
					weight:   MaxLengthPerNodeinterface,
					balanced: true,
					content:  currentBytes[:MaxLengthPerNodeinterface],
				}
				currentBytes = currentBytes[MaxLengthPerNodeinterface:]
			}
		}
		if balancedNode != nil {
			slotIndex := balancedNode.height - 1
			for slots[slotIndex] != nil {
				balancedNode = &Unium{
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
		ret = &Unium{
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

func (r *Unium) Split(n int) (out1, out2 *Unium) {
	if r == nil {
		return
	}
	if len(r.content) > 0 { // leaf
		if n > len(r.content) { // offset overflow
			n = len(r.content)
		}
		out1 = NewFrominterface(r.content[:n])
		out2 = NewFrominterface(r.content[n:])
	} else { // non leaf
		var r1 *Unium
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

func (r *Unium) Insert(n int, bs []interface{}) *Unium {
	r1, r2 := r.Split(n)
	return r1.Concat(NewFrominterface(bs)).Concat(r2)
}

func (r *Unium) Delete(n, l int) *Unium {
	r1, r2 := r.Split(n)
	_, r2 = r2.Split(l)
	return r1.Concat(r2)
}

// Sub returns a substring of the interface{}
func (r *Unium) Sub(n, l int) []interface{} {
	ret := make([]interface{}, l)
	i := 0
	r.Iter(n, func(bs []interface{}) bool {
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

func (r *Unium) Iter(offset int, fn func([]interface{}) bool) bool {
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

func (r *Unium) IterBackward(offset int, fn func([]interface{}) bool) bool {
	if r == nil {
		return true
	}
	if len(r.content) > 0 { // leaf
		content := r.content[:offset]
		if len(content) == 0 {
			return true
		}
		bs := reversedInterfaces(content)
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

func (r *Unium) iterNodes(fn func(*Unium) bool) {
	if r == nil {
		return
	}
	if fn(r) {
		r.left.iterNodes(fn)
		r.right.iterNodes(fn)
	}
}

func (r *Unium) IterRune(offset int, fn func(rune, int) bool) {
	var bs []interface{}
	r.Iter(offset, func(slice []interface{}) bool {
		bs = append(bs, slice...)
		return true
	})
}
