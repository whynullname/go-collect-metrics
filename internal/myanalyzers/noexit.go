package myanalyzers

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
)

var NoExitAnalyzer = &analysis.Analyzer{
	Name: "noexit",
	Doc:  "Запрещает использование os.Exit внутри функции main пакета main",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	if pass.Pkg.Name() != "main" {
		return nil, nil
	}

	for _, file := range pass.Files {
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Name.Name != "main" || fn.Recv != nil {
				continue
			}

			ast.Inspect(fn.Body, func(n ast.Node) bool {
				call, ok := n.(*ast.CallExpr)
				if !ok {
					return true
				}

				sel, ok := call.Fun.(*ast.SelectorExpr)
				if !ok {
					return true
				}

				pkgIdent, ok := sel.X.(*ast.Ident)
				if !ok {
					return true
				}

				if sel.Sel.Name == "Exit" {
					obj := pass.TypesInfo.Uses[pkgIdent]
					if pkgName, ok := obj.(*types.PkgName); ok && pkgName.Imported().Path() == "os" {
						pass.Reportf(call.Pos(), "запрещено использовать os.Exit в функции main")
					}
				}

				return true
			})
		}
	}

	return nil, nil
}
