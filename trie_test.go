package trie

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"
)

func TestBranchHasPrefix(t *testing.T) {
	for i, test := range []struct {
		val string
		sep string
		exp int
	}{
		{
			val: "hello",
			sep: "hell",
			exp: 4,
		},
	} {
		t.Run(fmt.Sprintf("#%d", i), func(t *testing.T) {
			b := branch{val: test.val}
			if i := b.hasPrefix(test.sep); i != test.exp {
				t.Errorf("b{%s}.hasPrefix(%s) = %d; want %d", test.val, test.sep, i, test.exp)
			}
		})
	}
}

func TestTrieAdd(t *testing.T) {
	type text struct {
		val string
		ok  bool
		exp string
	}
	for i, test := range []struct {
		patterns []string
		text     []text
	}{
		{
			patterns: []string{
				"ananas",
				"and",
				"banderole",
				"banana",
				"bandana",
				"card",
				"canary",
				"carry",
			},
			text: []text{
				{"ananafcardio", true, "card"},
				{"ananafcario", false, ""},
			},
		},
	} {
		t.Run(fmt.Sprintf("#%d", i), func(t *testing.T) {
			trie := New(test.patterns...)
			for _, text := range test.text {
				m, ok := trie.Match(text.val)
				nm, nok := naive(test.patterns, text.val)
				if m != nm || ok != nok {
					t.Logf(
						"warning: trie.Match(%s) = %q, %v; naive is %q, %v\n%s\n",
						text.val, m, ok, nm, nok, trie.String(),
					)
				}
				if text.ok != ok || text.exp != m {
					t.Errorf(
						"trie.Match(%s) = %q, %v; want %q, %v\n%s\n",
						text.val, m, ok, text.exp, text.ok, trie.String(),
					)
				}
			}
		})
	}
}

func patterns(n, m int) []string {
	ret := make([]string, n)
	uniq := make(map[string]bool)
	b := make([]byte, m)
	for i := 0; i < n; i++ {
		for {
			for j := 0; j < m; j++ {
				b[j] = byte(rand.Intn(int('x'-'a')) + 'a')
			}
			str := string(b)
			if _, ok := uniq[str]; !ok {
				ret[i] = str
				break
			}
		}
	}
	return ret
}

func BenchmarkTrieMatch(b *testing.B) {
	for i, bench := range []struct {
		patterns []string
		text     string
	}{
		{
			patterns: []string{
				"ananas",
				"and",
				"banderole",
				"banana",
				"bandana",
				"card",
				"canary",
				"carry",
			},
			text: "ananafcardreader",
		},
		{
			patterns: []string{
				"ananas",
				"banderole",
				"card",
			},
			text: "ananafcardreader",
		},
		{
			patterns: []string{
				"ananas",
				"banderole",
				"card",
			},
			text: strings.Repeat("x", 1024),
		},
		{
			patterns: patterns(2048, 16),
			text:     strings.Repeat("x", 1024) + "ananafcardreader",
		},
	} {
		b.Run(fmt.Sprintf("#%d_trie", i), func(b *testing.B) {
			trie := New(bench.patterns...)
			b.StartTimer()
			for i := 0; i < b.N; i++ {
				_, _ = trie.Match(bench.text)
			}
		})
		b.Run(fmt.Sprintf("#%d_naive", i), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, _ = naive(bench.patterns, bench.text)
			}
		})
	}
}

func naive(patterns []string, text string) (m string, ok bool) {
	index := -1
	for _, p := range patterns {
		if i := strings.Index(text, p); i != -1 && (index == -1 || i < index) {
			index = i
			m = p
		}
	}
	return m, index > -1
}
