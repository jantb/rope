package rope

import "bytes"

type Rope struct {
	weight  int
	left    *Rope
	right   *Rope
	content []byte
}

var MaxLengthPerNode = 128

func NewFromBytes(bs []byte) *Rope {
	if len(bs) == 0 {
		return nil
	}
	if len(bs) < MaxLengthPerNode {
		return &Rope{
			weight:  len(bs),
			content: bs,
		}
	}
	leftLen := len(bs) / 2
	return &Rope{
		weight: leftLen,
		left:   NewFromBytes(bs[:leftLen]),
		right:  NewFromBytes(bs[leftLen:]),
	}
}

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

func (r *Rope) Len() int {
	if r == nil {
		return 0
	}
	return r.weight + r.right.Len()
}

func (r *Rope) Bytes() []byte {
	buf := new(bytes.Buffer)
	r.collectBytes(buf)
	return buf.Bytes()
}

func (r *Rope) collectBytes(buf *bytes.Buffer) {
	if r == nil {
		return
	}
	if len(r.content) > 0 {
		buf.Write(r.content)
	} else {
		r.left.collectBytes(buf)
		r.right.collectBytes(buf)
	}
}

func (r *Rope) Concat(r2 *Rope) *Rope {
	return &Rope{
		weight: r.Len(),
		left:   r,
		right:  r2,
	}
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

func (r *Rope) Sub(n, l int) []byte {
	buf := new(bytes.Buffer)
	r.sub(n, l, buf)
	return buf.Bytes()
}

func (r *Rope) sub(n, l int, buf *bytes.Buffer) {
	if len(r.content) > 0 { // leaf
		end := n + l
		if end > len(r.content) {
			end = len(r.content)
		}
		buf.Write(r.content[n:end])
	} else { // non leaf
		if n >= r.weight { // start at right subtree
			r.right.sub(n-r.weight, l, buf)
		} else { // start at left subtree
			r.left.sub(n, l, buf)
			if n+l > r.weight {
				r.right.sub(0, n+l-r.weight, buf)
			}
		}
	}
}
