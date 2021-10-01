package fall

import (
	"errors"
	"io"
	"fmt"
	"strings"
	"math"
	"math/rand"
)

var LoopInDag = errors.New("Loop in inputed DAG, cannot make fall graph.")

type node struct {
	to      map[string]struct{}
	from    map[string]struct{}
	balance int
}

func Svg(w io.Writer, graph map[string][]string) error {
	return errors.New("Not yet implemented")
}

func Text(w io.Writer, graph map[string][]string) error {
	nodes, order, err := prepare(graph)
	if err != nil {
		return err
	}

	write := func(s string) {
		if err != nil {
			return
		}
		_, err = io.WriteString(w, s)
	}

	shift := 0
	for i, name := range order {
		s := len(name) - i*2
		if s > shift {
			shift = s
		}
	}

	vert := map[string]struct{}{}
	for i, name := range order {
		padding := shift - len(name) + i*2
		for j := 0; j < padding; j++ {
			write(" ")
		}

		write(name)
		tos := len(nodes[name].to)
		for _, other := range order[i+1:] {
			if tos > 0 {
				if _, ok := nodes[name].to[other]; ok {
					vert[other] = struct{}{}
					write("━█")
					tos--
				} else if _, ok := vert[other]; ok {
					write("━╋")
				} else {
					write("━━")
				}
			} else {
				if _, ok := vert[other]; ok {
					write(" ┃")
				} else {
					write("  ")
				}
			}
		}

		write("\n")

		padding += len(name)
		for j := 0; j < padding; j++ {
			write(" ")
		}
		for _, other := range order[i+1:] {
			if _, ok := vert[other]; ok {
				write(" ┃")
			} else {
				write("  ")
			}
		}

		write("\n")
		_ = nodes
	}
	return err
}

// Dot exports the graph format used elsewhere in this package in the DOT
// format, so the same input can be used for comparison with other programs'
// output.
func Dot(w io.Writer, graph map[string][]string) error {
	var err error
	write := func(s string) {
		if err != nil {
			return
		}
		_, err = io.WriteString(w, s)
	}

	write("digraph {\n")
	nodeNames := make(map[string]string, len(graph))
	next := 0
	for name := range graph {
		nodeNames[name] = fmt.Sprintf("n%d", next)
		next++
		write("  ")
		write(nodeNames[name])
		write(" [label=\"")
		write(strings.ReplaceAll(name, "\"", "\\\""))
		write("\"];\n")
	}

	for from, edges := range graph {
		for _, to := range edges {
			write("  ")
			write(nodeNames[from])
			write(" -> ")
			write(nodeNames[to])
			write(";\n")
		}
	}

	write("}\n")

	return err
}

func prepare(graph map[string][]string) (map[string]*node, []string, error) {
	nodes := make(map[string]*node, len(graph))

	r := rand.New(rand.NewSource(0))

	for name := range graph {
		nodes[name] = &node{
			to:   make(map[string]struct{}),
			from: make(map[string]struct{}),
		}
	}

	for from := range graph {
		for _, to := range graph[from] {
			nodes[from].to[to] = struct{}{}
			nodes[to].from[from] = struct{}{}
		}
	}

	for i := range nodes {
		nodes[i].balance = len(nodes[i].to) - len(nodes[i].from)
	}

	var best []string
	var bestSum = math.MaxInt
	var bestIndex = 0
	for i := 0; (i < 100) || (i < bestIndex*2); i++ {
		shuffle, err := randomOrder(r, nodes)
		if err != nil {
			return nil, nil, err
		}
		simplify(nodes, shuffle)
		sum := sumDistances(nodes, shuffle)
		if sum < bestSum {
			bestSum = sum
			best = shuffle
			bestIndex = i
		}
	}

	return nodes, best, nil
}

func randomOrder(r *rand.Rand, nodes map[string]*node) ([]string, error) {
	nextFree := make([]string, 0, len(nodes))
	remaining := map[string]int{}

	for name := range nodes {
		remaining[name] = len(nodes[name].from)
		if remaining[name] == 0 {
			nextFree = append(nextFree, name)
		}
	}

	result := make([]string, 0, len(nodes))
	for {
		i := r.Intn(len(nextFree))
		next := nextFree[i]
		result = append(result, next)

		for other := range nodes[next].to {
			remaining[other]--
			if remaining[other] == 0 {
				nextFree = append(nextFree, other)
			}
		}

		if len(nextFree) <= 1 {
			break
		}
		last := len(nextFree) - 1
		nextFree[i] = nextFree[last]
		nextFree = nextFree[:last]
	}

	if len(result) != len(nodes) {
		return nil, LoopInDag
	}

	return result, nil
}

func simplify(nodes map[string]*node, result []string) {
	// for k := 0; k < 10; k++ {
	change := true
	for change {
		change = false
		for i := 0; i < len(result)-1; i++ {
			ri := result[i]
			rip := result[i+1]
			_, cantswap := nodes[ri].to[rip]
			if !cantswap && nodes[ri].balance > nodes[rip].balance {
				result[i], result[i+1] = result[i+1], result[i]
				change = true
			}
		}
	}
}

func sumDistances(nodes map[string]*node, result []string) int {
	pos := make(map[string]int, len(result))
	for i, name := range result {
		pos[name] = i
	}

	sum := 0

	for i, name := range result {
		for other := range nodes[name].to {
			sum += pos[other] - i
		}
	}
	return sum
}
