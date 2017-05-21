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
)

func test_map() {
	// map is not concurrent safe and reference type
	m := make(map[string]string)

	m["c"] = "value c"
}

func test_solution_func() {
	gram := NewGrammar()
	grm.terminal = NewSymbolSet("r d o a e i")
	grm.nonterminal = NewSymbolSet("S X Z")
	grm.start = Symbol(strings.Trim("S", " "))

	grm.rules = []Rule{Rule{NewSymbolSet("S"), NewSymbolSet("r X d")}}

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
