package main

import (
	"github.com/Laremere/fall"
	"go/build"
	"os"
)

func main() {
	dag := make(map[string][]string)

	todo := []string{"github.com/Laremere/fall/godeps"}
	checking := map[string]struct{}{
		todo[0]: struct{}{},
	}

	for len(todo) > 0 {
		last := len(todo) - 1
		path := todo[last]
		todo = todo[:last]

		dag[path] = nil

		p, err := build.Import(path, "", 0)
		if err != nil {
			panic(err)
		}

		for _, depPath := range p.Imports {
			dag[path] = append(dag[path], depPath)
			if _, ok := checking[depPath]; !ok {
				checking[depPath] = struct{}{}
				todo = append(todo, depPath)
			}
		}
	}

	fall.Text(os.Stdout, dag)
	fall.Dot(os.Stdout, dag)
}
