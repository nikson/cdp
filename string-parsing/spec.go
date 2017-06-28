/**
Author: Nikson Kanti Paul
*/

package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"context"
	"runtime"
	"sync"
	// _ "github.com/pkg/profile"
)

// ------------- Data types -------------------- //

type Symbol string

type SymbolSet []string

func NewSymbolSet(s string) SymbolSet {
	return SymbolSet(strings.Fields(s))
}

type Rule struct {
	left  SymbolSet
	right SymbolSet
}

// Data type for the solution
type Grammar struct {
	terminal    SymbolSet
	nonterminal SymbolSet
	start       Symbol
	rules       []Rule
}

// constructor of datatype
func NewGrammar() Grammar {
	return Grammar{terminal: make([]string, 0),
		nonterminal: make([]string, 0),
		start: Symbol(""),
		rules: make([]Rule, 0) }
}
/*/
func (src Grammar) copy() Grammar {
	dest := NewGrammar()
	copy(src.terminal, dest.terminal)
	copy(src.nonterminal, dest.nonterminal)
	for key := range src.rules {
		dest.rules[key] = src.rules[key]
	}
	return dest
}
//*/

type Stack struct {
	current SymbolSet // evaluation rule, will apply in 'input'
	input   SymbolSet // input data
	path    []Rule    // rules set used until 'current'
	success bool      // status of Grammar
}

func NewStack() Stack {
	return Stack{current: make([]string, 0),
		input: make([]string, 0),
		path: make([]Rule, 0) }
}

func (src Stack) clone() Stack {
	dst := Stack{input: src.input, current: src.current, success: src.success}
	dst.path = append(dst.path, src.path...)

	// dst.path = make([]Rule, len(src.path))	
	// for i, rule := range src.path {
	// 	dst.path[i] = rule
	// }

	return dst
}

// ------------- data type -----------

// ------------ functions ------------
// merge array into a single string use separator 'sep'
func (s SymbolSet) str() string {
	return strings.Join(s, "")
}
func (s SymbolSet) join(sep string) string {
	return strings.Join(s, sep)
}

func (dst SymbolSet) append(src SymbolSet) SymbolSet {
	var sym []string
	for _, x := range dst {
		sym = append(sym, x)
	}
	for _, x := range src {
		sym = append(sym, x)
	}
	return SymbolSet(sym)
}

func (s SymbolSet) empty() bool {
	return (len(s) == 0)
}

func (s Symbol) str() string {
	return string(s)
}
func (r Rule) str() string {
	return r.left.join(" ") + " : " + r.right.join(" ")
}

func (s Stack) addRule(r Rule) Stack {
	s.path = append(s.path, r)
	return s

}
// ------------ functions ------------

// ------------- problem solving functions --------------

// Pseudo:
// 1. step 1: eliminate common prefix
// 2. step 2: first symbol of current is terminal, return FAILED
// 3. step 3: input and current are empty : return SUCCESS
// 4. step 4: input empyt but current contains a terminal: return FAILED
// 5. step 5: p = Aq, apply all production rules for A=>R
// 6. step 6: recursive call on R, if any SUCCESS found terminate the execution

func evaluate_grammar(ctx context.Context, output chan <- Stack, wg *sync.WaitGroup, g Grammar, data Stack) Stack {
	// cancel all recursive call/goroutine thread while "cancel" triggered

	defer wg.Done()

	select {
	case <-ctx.Done():
		return NewStack()
	default:
	}

	// 1. step 1: eliminate common prefix
	data = reduce(data);

	// 2. step 2: first symbol of current is terminal, return FAILED
	if ( !data.current.empty() && is_terminal(g, data.current[0])) {
		data.success = false
		//output <- data
		return data
	}

	if (data.input.empty()) {
		// 3. step 3: input and current are empty : return SUCCESS
		if (data.current.empty()) {
			data.success = true;
			// write the SUCCESS stack in output channel, which will trigger terminate program
			output <- data
			return data;
		}
		// 4. step 4: input empty but current contains a terminal: return FAILED
		if (contains_terminal(g, data.current)) {
			data.success = false
			//output <- data
			return data
		}
	}

	// 5. step 5: p = Aq, apply all production rules for A
	new_stacks := derive_leftmost_grammar(g, data)

	// 6. step 6: recursive call on R, if any SUCCESS found terminate the execution
	for _, s := range new_stacks {
		// if cancel trigger, exit
		select {
		case <-ctx.Done():
			return NewStack()
		default:
		}

		wg.Add(1)
		go evaluate_grammar(ctx, output, wg, g, s)
		//temp := evaluate_grammar(ctx, output, g, s)
		//if temp.success {
		//	return temp
		//}
	}

	// release underlaying memory for GC
	new_stacks = nil

	//output <- data
	return data
}

// Pseudo:
// step 1: check is there any non terminal symbol, if not return empty
// step 2: find first left most non-terminal symbol (X)
// step 3: find all production rules of X => R
// step 4: apply and expand the stack using R
func derive_leftmost_grammar(g Grammar, s Stack) []Stack {
	// step 1 & 2
	// find left most non terminal symbol(X), if not found return empty
	found, non_terminal_symbol, index := contains_non_terminal(g, s.current)
	if !found {
		return []Stack{}
	}

	// step 3: find all production rules of X
	rules := find_grammar(g, non_terminal_symbol)
	// step 4: apply and expand the stack using R
	stacks := expand_grammar(s, rules, index)

	// release slice
	rules = nil
	return stacks
}

func find_grammar(g Grammar, left_rule string) []Rule {
	rules := make([]Rule, 0)

	for _, r := range g.rules {
		// flaten the array
		if r.left.str() == left_rule {
			rules = append(rules, r)
		}
	}
	return rules
}

func expand_grammar(s Stack, rules []Rule, index_of_non_terminal int) []Stack {
	// step 1: split s.current = left - middle - right
	left := s.current[:index_of_non_terminal]
	right := s.current[index_of_non_terminal + 1:]
	// step 2: replace middle using Rule->right
	stacks := make([]Stack, len(rules))
	for i, rule := range rules {
		// clone current stack
		new_stack := s.clone()
		// add new rules
		new_stack = new_stack.addRule(rule)
		// apply rules and extend current
		middle := rule.right
		new_stack.current = left.append(middle).append(right)
		// add it in array
		stacks[i] = new_stack
	}

	return stacks
}

func is_terminal(g Grammar, sym string) bool {
	return find(g.terminal, sym)
}

func is_non_terminal(g Grammar, sym string) bool {
	return !is_terminal(g, sym)
}

func find(ds SymbolSet, lookup string) bool {
	for _, item := range ds {
		if item == lookup {
			return true
		}
	}
	return false
}

func contains_terminal(g Grammar, val SymbolSet) bool {
	for _, sym := range val {
		if is_terminal(g, sym) {
			return true
		}
	}
	return false
}

// check current rule contian non-terminal, and return it with position
func contains_non_terminal(g Grammar, val SymbolSet) (bool, string, int) {
	for pos, sym := range val {
		if is_non_terminal(g, sym) {
			return true, sym, pos
		}
	}
	return false, "", 0
}

func reduce(s Stack) Stack {
	if (s.current.empty() || s.input.empty() ) {
		return s
	}

	index := 0
	for i, match := range s.current {
		if (i < len(s.input) && s.input[i] == match) {
			index = i + 1;
		} else {
			break
		}
	}

	s.current = s.current[index:]
	s.input = s.input[index:]
	return s
}


// ------------------------ problem solving function -------------------

// ------------  input, output -----------

func read_input(rd io.Reader) (Grammar, Stack) {
	scanner := bufio.NewScanner(rd)
	// data structure for input data types
	grm := NewGrammar()
	stk := NewStack()

	// read input
	if scanner.Scan() {
		stk.input = NewSymbolSet(scanner.Text())
	}
	// read terminal
	if scanner.Scan() {
		grm.terminal = NewSymbolSet(scanner.Text())
	}
	// read nonterminal
	if scanner.Scan() {
		grm.nonterminal = NewSymbolSet(scanner.Text())
	}
	// read starting rules
	if scanner.Scan() {
		grm.start = Symbol(strings.Trim(scanner.Text(), " "))
	}

	// read production rules
	for scanner.Scan() {
		prod_rule := scanner.Text()
		if prod_rule == "-" {
			break
		}
		r := strings.Split(prod_rule, ":")
		left, right := r[0], r[1]
		grm.rules = append(grm.rules, Rule{NewSymbolSet(left), NewSymbolSet(right)})
	}

	return grm, stk
}

func print_output(ret Stack) {
	if ret.success {
		for _, r := range ret.path {
			fmt.Println(r.str())
		}
		fmt.Println("SUCCESS")
	} else {
		fmt.Println("FAILED")
	}
}

func print(grm Grammar, st Stack) {
	//fmt.Println("Input word:")
	fmt.Println(st.input.str())

	//fmt.Println("Terminal:")
	fmt.Println(grm.terminal.str())
	fmt.Println(grm.nonterminal.str())
	fmt.Println(grm.start)

	for _, r := range grm.rules {
		//fmt.Println(r.left.str() + ":" + r.right.str())
		fmt.Println(r.str())
	}
}

func main() {
	// defer profile.Start(profile.CPUProfile, profile.ProfilePath("."), profile.NoShutdownHook).Stop()
	// defer profile.Start(profile.MemProfile, profile.ProfilePath("."), profile.NoShutdownHook).Stop()

	runtime.GOMAXPROCS(runtime.NumCPU())

	// read input from stdin
	gram, input := read_input(os.Stdin)
	//print(gram, input)

	begin := input
	begin.current = NewSymbolSet(gram.start.str())

	// worker counter
	wg := sync.WaitGroup{}
	// recursive call cancel handler
	ctx, cancel := context.WithCancel(context.Background())
	// result channel
	result_chan := make(chan Stack)                // single buffer for 1 success value
	result := NewStack()

	// count
	wg.Add(1)
	// start main go routine
	go evaluate_grammar(ctx, result_chan, &wg, gram, begin)

	// Special routine: if all evaluate_grammar routine finish without SUCCESS, detect it and push empty stack
	go func(out chan <- Stack, wg1 *sync.WaitGroup) {
		wg1.Wait()
		out <- NewStack()
	}(result_chan, &wg)

	// Wait for result, if result found cancel all recurisve call and print result
	result = <-result_chan
	cancel()
	print_output(result)
}
