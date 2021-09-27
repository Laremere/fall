package main

import (
	"github.com/Laremere/fall"
	"os"
)

func main() {
	dag := map[string][]string{
		"iron_ore":   {},
		"copper_ore": {},
		"iron":       {"iron_ore"},
		"copper":     {"copper_ore"},
		"circuit":    {"iron", "copper"},
	}

	fall.Text(os.Stdout, dag)
}
