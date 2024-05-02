// Package noosexit Анализатор вызова функции os.Exit в пакете main функции main()
package noosexit

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

var OsExitAnalyzer = &analysis.Analyzer{
	Name: "noosexit",
	Doc:  "анализатор, запрещающий использовать прямой вызов os.Exit в функции main пакета main",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {

	for _, file := range pass.Files {

		if pass.Pkg.Name() != "main" {
			continue
		}

		ast.Inspect(file, func(n ast.Node) bool {

			mainFunc, ok := n.(*ast.FuncDecl)
			if !ok {
				return true // переходим к дочерним узлам
			}

			if mainFunc.Name.String() != "main" {
				return false // функиция НЕ main - к дочерним узлам смысла переходить нет
			}

			// переходим к разбору ф-ии main()
			ast.Inspect(mainFunc, func(node ast.Node) bool {
				call, okCall := node.(*ast.CallExpr)
				if !okCall { // если НЕ вызов ф-ии - пропускаем
					return true
				}

				s, okSelector := call.Fun.(*ast.SelectorExpr)
				if !okSelector {
					return true
				}

				if s.Sel.Name == "Exit" {
					ident := s.X.(*ast.Ident)
					if ident.Name == "os" {
						pass.Reportf(s.Pos(), "Вызов os.Exit в функции main")
					}
				}

				return false
			})

			return false
		})

	}

	return nil, nil
}
