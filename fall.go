package fall

import (
	"errors"
	"io"
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
	return text(w, graph, func(int) string { return "" }, "")
}

// TODO: Needs to call into windows and set the terminal mode.
func WinTerm(w io.Writer, graph map[string][]string) error {
	return text(w, graph, winTermColor, "\x1B[[0m")
}

func winTermColor(n int) string {
	switch n % 6 {
	case 0:
		return "\x1B[[97m" // White
	case 1:
		return "\x1B[[91m" // Red
	case 2:
		return "\x1B[[92m" // Green
	case 3:
		return "\x1B[[93m" // Yellow
		//Skipping blue because it's hard to read.
	case 4:
		return "\x1B[[95m" // Magenta
	case 5:
		fallthrough
	default:
		return "\x1B[[96m" // Cyan
	}
}

func text(w io.Writer, graph map[string][]string, colorCode func(int) string, resetColor string) error {
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

	colorCodeMap := make(map[string]string)
	for i, name := range order {
		colorCodeMap[name] = colorCode(i)
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
					write(colorCodeMap[name])
					write("━")
					write(resetColor)
					write("█")
					tos--
				} else if _, ok := vert[other]; ok {
					write(colorCodeMap[name])
					write("━╋")
					write(resetColor)
				} else {
					write(colorCodeMap[name])
					write("━━")
					write(resetColor)
				}
			} else {
				if _, ok := vert[other]; ok {
					write(colorCodeMap[other])
					write(" ┃")
					write(resetColor)
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
		println(i)
		println(bestIndex)
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
