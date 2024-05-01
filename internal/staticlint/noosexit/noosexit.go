package main

import (
	"fmt"
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

var OsExitAnalyzer = &analysis.Analyzer{
	Name: "noosexit",
	Doc:  "анализатор, запрещающий использовать прямой вызов os.Exit в функции main пакета main",
	Run:  run,
}

func main() {

}

func run(pass *analysis.Pass) (interface{}, error) {

	//fset := token.NewFileSet()
	//// получаем дерево разбора
	//f, err := parser.ParseFile(fset, "", src, parser.AllErrors)
	//if err != nil {
	//	fmt.Println(err)
	//}

	for _, file := range pass.Files {
		ast.Inspect(file, func(n ast.Node) bool {

			fmt.Println(file.Name)

			return true
		})

	}

	return nil, nil
}
