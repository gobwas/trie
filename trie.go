package trie

import (
	"bytes"
	"fmt"
	"io"
	"reflect"
	"sort"
	"strings"
	"unsafe"
)

type branch struct {
	val     string
	sub     map[int][]*branch
	offsets []int
	offset  int
	offsetv uint64
}

func newBranch(p string) *branch {
	return &branch{
		val:    p,
		sub:    make(map[int][]*branch),
		offset: len(p),
	}
}

var bits [8]uint64

func init() {
	bits[0] = ^(uint64(0))
	for i := 1; i < 8; i++ {
		bits[i] = ^(uint64(0xff) << (8 * uint64(8-i))) & bits[i-1]
	}
}

func bits64(a string, n int) uint64 {
	ah := *(*reflect.StringHeader)(unsafe.Pointer(&a))
	av := *(*uint64)(unsafe.Pointer(ah.Data))
	if n < 8 {
		av &= bits[8-n]
	}
	return av
}

func match64(a, b string, n int) bool {
	return bits64(a, n) == bits64(b, n)
}

func (b *branch) couldMatch(s string) bool {
	n := b.offset
	if n > 8 {
		n = 8
	}
	return b.matchFirstN(s, n)
}

func (b *branch) matchFirstN(s string, n int) bool {
	if n == 1 {
		return b.val[0] == s[0]
	}
	if n <= 8 {
		return b.offsetv == bits64(s, n)
	}
	return strings.HasPrefix(s, b.val[:n])
}

func (b *branch) hasPrefix(s string) int {
	var i, n int

	if len(s) < len(b.val) {
		n = len(s)
	} else {
		n = len(b.val)
	}

	for i = 0; i < n; i++ {
		if b.val[i] != s[i] {
			break
		}
	}
	return i
}

func (b *branch) finalize() {
	b.offsets = make([]int, 0, len(b.sub))
	for n, sb := range b.sub {
		b.offsets = append(b.offsets, n)
		for _, sub := range sb {
			sub.finalize()
		}
	}
	sort.Ints(b.offsets)

	if b.val != "" {
		b.offsetv = bits64(b.val, b.offset)
	}
}

func (b *branch) append(offset int, p string) {
	var max int
	var maxb *branch
	var maxj int
	for j, br := range b.sub[offset] {
		if n := br.hasPrefix(p); n != 0 && n > max {
			max = n
			maxb = br
			maxj = j
		}
	}

	if maxb == nil {
		b.sub[offset] = append(b.sub[offset], newBranch(p))
		if b.offset > offset {
			b.offset = offset
		}
		return
	}

	if len(maxb.val) < len(p) {
		nb := newBranch(p)
		nb.sub[max] = []*branch{maxb}
		nb.offset = max

		b.sub[offset][maxj] = nb
		maxb.val = maxb.val[max:]
		maxb.offset = len(maxb.val)
		return
	}

	maxb.append(max, p[max:])
}

func (b *branch) matchPrefix(s string) (eq int) {
	eq = -1
	for _, offset := range b.offsets {
		if !b.matchFirstN(s, offset) {
			return
		}
		eq = offset
	}
	if b.matchFirstN(s, len(b.val)) {
		eq = len(b.val)
	}
	return
}

func (b *branch) match(offset int, s string) (n int, ok bool) {
	min := -1
	if s == "" {
		return 0, true
	}
	for _, sb := range b.sub[offset] {
		if sb.val == "" {
			return offset, true
		}
		if sb.couldMatch(s) {
			m := sb.matchPrefix(s)
			if m == len(sb.val) {
				return offset + m, true
			}
			if m != -1 {
				if n, ok := sb.match(m, s[m:]); ok {
					return offset + n, true
				}
			}
		}
		if min == -1 || sb.offset < min {
			min = sb.offset
		}
	}

	return min, false
}

type Trie struct {
	root *branch
}

func New(patterns ...string) *Trie {
	t := &Trie{
		root: &branch{
			sub: make(map[int][]*branch),
		},
	}
	for _, p := range patterns {
		t.root.append(0, p)
	}
	t.root.finalize()
	return t
}

func (t *Trie) String() string {
	buf := &bytes.Buffer{}
	draw(buf, 0, t.root)
	return buf.String()
}

func (t *Trie) Match(s string) (string, bool) {
	var pos int
	for pos < len(s)-1 {
		n, ok := t.root.match(0, s[pos:])
		if ok {
			return s[pos : pos+n], true
		}
		if n <= 0 {
			panic("trie fatal error")
		}
		pos += n
	}
	return "", false
}

func draw(w io.Writer, tab int, b *branch) {
	var space string
	if tab >= 1 {
		space = strings.Repeat(" ", tab-1) + "└"
	}
	if b.val == "" {
		if tab != 0 {
			fmt.Fprintln(w, space+"$")
		}
	} else {
		fmt.Fprintln(w, space+b.val)
	}

	subs := make([]int, 0, len(b.sub))
	for n := range b.sub {
		subs = append(subs, n)
	}
	sort.Ints(subs)
	for _, n := range subs {
		for _, b := range b.sub[n] {
			draw(w, tab+n, b)
		}
	}
	return
}
