/*

https://golang.org/pkg/strings/
https://golang.org/pkg/bufio/#Scanner.Scan
https://golang.org/pkg/container/list/
https://blog.golang.org/go-maps-in-action
https://golang.org/pkg/context/

*/

package main

import (
	"fmt"
	"strings"
	"testing"
)

func Test_map(t *testing.T) {
	// map is not concurrent safe and reference type
	m := make(map[string]string)

	m["c"] = "value c"
}

func Test_solution_func(t *testing.T) {
	gram := NewGrammar()
	gram.terminal = NewSymbolSet("r d o a e i")
	gram.nonterminal = NewSymbolSet("S X Z")
	gram.start = Symbol(strings.Trim("S", " "))

	gram.rules = append(gram.rules, Rule{NewSymbolSet("S"), NewSymbolSet("r X d")})
	gram.rules = append(gram.rules, Rule{NewSymbolSet("X"), NewSymbolSet("o a")})

	begin := Stack{}
	begin.current = NewSymbolSet(gram.start.str())
	ret := begin

	// test left derivation
	for _, s := range derive_leftmost_grammar(gram, begin) {
		fmt.Println(s.current.str(), s.path)
	}

	// test find-non-terminal
	_, str, index := contains_non_terminal(gram, begin.current)

	// test find_grammar
	rule := find_grammar(gram, str)
	for _, r := range rule {
		fmt.Println(r.str())
	}

	// test expand grammar
	fmt.Println(begin.current.join(" "))
	for _, e := range expand_grammar(begin, rule, index) {
		fmt.Println(e.current.join(" "))
		for _, p := range e.path {
			fmt.Println(p.str())
		}
	}

	if ret.success {
		for _, r := range ret.path {
			fmt.Println(r.str())
		}
		fmt.Println("SUCCESS")
	} else {
		fmt.Println("FAILED")
	}
}

func Test_stack_clone(t *testing.T) {
	src := Stack{}
	src.current = NewSymbolSet("r o a d")
	src.path = []Rule{Rule{NewSymbolSet("S"), NewSymbolSet("r X d")}}

	dst := Stack{}
	clone(&src, &dst)

	dst.current = NewSymbolSet("r e a d")
	//dst.path = []Rule{}
	dst.path = append(dst.path, Rule{NewSymbolSet("S"), NewSymbolSet("r Z")})

	t.Log(src)
	t.Log(dst)

}