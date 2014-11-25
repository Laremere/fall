package main

import (
	"fmt"
	"go/build"
	"os"
	"sort"
	"text/template"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: fall [import path]")
		return
	}
	var si SvgInfo
	si.Height = 3
	si.Width = 100
	var err error
	si.Listing, err = getImports(os.Args[1])
	if err != nil {
		panic(err)
	}

	for _, ident := range si.Listing {
		rowHeight := 17
		if rowHeight < 5+len(ident.Imports)*3+10 {
			rowHeight = 5 + len(ident.Imports)*3 + 10
		}
		if si.Width < len(ident.Name)*7 {
			si.Width = len(ident.Name) * 7
		}
		ident.Y = si.Height
		ident.Ytext = ident.Y + 13
		si.Height += rowHeight
		for _, depIdent := range ident.Imports {
			depIdent.TimesImported += 1
		}
	}

	for i, ident := range si.Listing {
		ident.Index = i

		si.Width += ident.TimesImported * 3
		ident.Vertical = si.Width - 2
		ident.TopWidth = si.Width
		ident.Xtext = si.Width - 10
		si.Width += 3
	}

	for _, ident := range si.Listing {
		sort.Sort(ident)
		yspot := ident.Y + 2 + len(ident.Imports)*3 - 3
		for _, dep := range ident.Imports {
			si.Curves = append(si.Curves, DepCurve{
				ident.TopWidth - 10,
				yspot,
				dep.Vertical - ident.TopWidth,
				dep.Y - yspot - 11,
				positionToColor(dep.Index),
			})
			yspot -= 3
			dep.Vertical -= 3
		}
	}

	// si.Curves = []DepCurve{DepCurve{
	// 	leftSide, 20, 40, 30,
	// }}

	file, err := os.Create("out.html")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	t := template.Must(template.New("svg").Parse(SvgTemplate))
	err = t.Execute(file, &si)
	if err != nil {
		panic(err)
	}

	fmt.Println("Number of packages:", len(si.Listing))
	fmt.Println("Number of dependancy relations:", len(si.Curves))
}

func positionToColor(pos int) string {
	const goldenAngle = 137.508
	angle := int(goldenAngle*float64(pos)) % 360
	return fmt.Sprintf("hsl(%d, 75%%, 50%%)", angle)
}

type SvgInfo struct {
	Listing       []*PkgImports
	Width, Height int
	Curves        []DepCurve
}

type DepCurve struct {
	XStart     int
	YStart     int
	Horizontal int
	Vertical   int
	Color      string
}

type PkgImports struct {
	Name          string
	Imports       []*PkgImports
	Y             int
	Ytext         int
	Xtext         int
	TopWidth      int
	TimesImported int
	Vertical      int
	Index         int
}

func (p *PkgImports) Len() int { return len(p.Imports) }
func (p *PkgImports) Swap(i, j int) {
	p.Imports[i], p.Imports[j] = p.Imports[j], p.Imports[i]
}
func (p *PkgImports) Less(i, j int) bool {
	return p.Imports[i].Index < p.Imports[j].Index
}

const SvgTemplate = `
<html>
<body>
<svg height="{{.Height}}" width="{{.Width}}">
<g font-size="15" font="sans-serif" fill="black" stroke="none">
{{range .Listing}}
	<text x="{{.Xtext}}" y="{{.Ytext}}" text-anchor="end">{{.Name}}</text>
{{end}}
</g>
{{range .Listing}}
	<path d="M 0 {{.Y}} h {{.TopWidth}}" stroke="blue" stroke-width="2" fill="none"/>
{{end}}
{{range .Curves}}
	<path d="M {{.XStart}} {{.YStart}} h {{.Horizontal}} q 10 0 10 10 v {{.Vertical}}"
	stroke="{{.Color}}" stroke-width="3" fill="none"/>
{{end}}
</svg> 
</body>
</html>
`

func (p *PkgImports) String() string {
	result := "Pkg " + p.Name + " imports:"
	for _, ident := range p.Imports {
		result += " " + ident.Name
	}
	return result + "\n"
}

func getImports(pkgRoot string) ([]*PkgImports, error) {
	type importsCount struct {
		imports []string
		count   int
		ident   *PkgImports
	}

	todo := []string{pkgRoot}
	pkgs := make(map[string]*importsCount)
	pkgs[pkgRoot] = new(importsCount)
	for len(todo) > 0 {
		pkgName := todo[len(todo)-1]
		todo = todo[0 : len(todo)-1]

		pkgDetails, err := build.Import(pkgName, "", 0)
		if err != nil {
			return nil, err
		}
		pkgs[pkgName].imports = pkgDetails.Imports
		for _, i := range pkgDetails.Imports {
			if _, ok := pkgs[i]; !ok {
				pkgs[i] = new(importsCount)
				todo = append(todo, i)
			}
			pkgs[i].count += 1
		}
	}
	for pkgName := range pkgs {
		var ident PkgImports
		ident.Name = pkgName
		ident.Imports = make([]*PkgImports, len(pkgs[pkgName].imports))
		pkgs[pkgName].ident = &ident
	}

	todo = append(todo, pkgRoot)
	result := make([]*PkgImports, 0, len(pkgs))
	lastEntry := ""

	for len(todo) > 0 {
		bestSemanticSimularity := -1
		bssIndex := 0
		for i, pkgName := range todo {
			var semSim int
			for semSim = 0; semSim < len(lastEntry) &&
				semSim < len(pkgName) &&
				lastEntry[semSim] == pkgName[semSim]; semSim++ {
			}
			if semSim > bestSemanticSimularity {
				bestSemanticSimularity = semSim
				bssIndex = i
			}
		}

		pkgName := todo[bssIndex]
		lastEntry = pkgName
		todo[bssIndex] = todo[len(todo)-1]
		todo = todo[0 : len(todo)-1]
		ident := pkgs[pkgName].ident

		result = append(result, ident)
		for i, pkgDep := range pkgs[pkgName].imports {
			pkgs[pkgDep].count -= 1
			if pkgs[pkgDep].count <= 0 {
				todo = append(todo, pkgDep)
			}
			ident.Imports[i] = pkgs[pkgDep].ident
		}
	}

	return result, nil
}

// func getImports(pkgRoot string) (map[string][]string, error) {
// 	remaining := []string{pkgRoot}

// 	result := make(map[string][]string)
// 	for len(remaining) > 0 {
// 		pkg := remaining[len(remaining)-1]
// 		remaining = remaining[0 : len(remaining)-1]

// 		pkgDetails, err := build.Import(pkg, "", 0)
// 		if err != nil {
// 			return nil, err
// 		}
// 		result[pkg] = pkgDetails.Imports
// 		for _, i := range pkgDetails.Imports {
// 			if _, ok := result[i]; !ok {
// 				result[i] = nil
// 				remaining = append(remaining, i)
// 			}
// 		}
// 	}
// 	return result, nil
// }

// func sortImports(imports map[string][]string) []string {
// 	result = make([]string, len(imports))
// 	for pkg := range imports {
// 		minIndex := 0

// 	}
// }
