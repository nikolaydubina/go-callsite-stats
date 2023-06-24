package main

import (
	"encoding/json"
	"flag"
	"go/ast"
	"go/token"
	"log"
	"os"

	"golang.org/x/tools/go/packages"
)

func main() {
	var (
		packagePattern string
		outJSON        bool
	)
	flag.BoolVar(&outJSON, "json", true, "output as JSONL to STDOUT")

	flag.Parse()
	if flag.NArg() == 0 {
		log.Fatal("missing package pattern (e.g. ./...)")
	}

	packagePattern = flag.Args()[0]

	var fset = token.NewFileSet()

	mode := packages.NeedName | packages.NeedTypes | packages.NeedSyntax | packages.NeedTypesInfo
	cfg := &packages.Config{Fset: fset, Mode: mode}
	pkgs, err := packages.Load(cfg, packagePattern)
	if err != nil {
		log.Fatal(err)
	}

	stats := make(map[FuncID]*FuncCallSiteStats)
	for _, pkg := range pkgs {
		for _, fileAst := range pkg.Syntax {
			for funcID, funcStats := range CollectFuncCallSiteStatsForFile(fileAst) {
				mergeFuncCallSiteStatsToMap(stats, funcID, funcStats)
			}
		}
	}

	if outJSON {
		encoder := json.NewEncoder(os.Stdout)
		for funcID, funcStat := range stats {
			type FuncStatRow struct {
				FuncID
				*FuncCallSiteStats
			}
			if err := encoder.Encode(FuncStatRow{FuncID: funcID, FuncCallSiteStats: funcStat}); err != nil {
				log.Printf("%s\n", err)
			}
		}
	}
}

func mergeFuncCallSiteStatsToMap(stats map[FuncID]*FuncCallSiteStats, funcID FuncID, funcStats *FuncCallSiteStats) {
	if _, ok := stats[funcID]; !ok {
		stats[funcID] = &FuncCallSiteStats{}
	}
	mergeFuncCallSiteStats(funcStats, stats[funcID])
}

// CollectFuncCallSiteStatsForFile can be used in analyzer and in other static analysis tools
// https://go.dev/ref/spec#Assignment_statements
func CollectFuncCallSiteStatsForFile(file *ast.File) map[FuncID]*FuncCallSiteStats {
	stats := map[FuncID]*FuncCallSiteStats{}

	ast.Inspect(file, func(n ast.Node) bool {
		switch n := n.(type) {
		case *ast.CallExpr:
			funcID, funcStats := analyzeFuncCallArguments(n)
			mergeFuncCallSiteStatsToMap(stats, funcID, &funcStats)
		case *ast.AssignStmt:
			for funcID, funcStats := range analyzeMultiFunctionAssignment(n.Lhs, n.Rhs) {
				mergeFuncCallSiteStatsToMap(stats, funcID, funcStats)
			}
		}
		return true
	})

	return stats
}

func analyzeFuncCallArguments(n *ast.CallExpr) (funcID FuncID, stats FuncCallSiteStats) {
	funcID = FinalCallerFuncIDFromCallExpr(n)

	for _, expr := range n.Args {
		if ident, ok := expr.(*ast.Ident); ok && ident != nil {
			stats.ArgumentNameCount = append(stats.ArgumentNameCount, map[string]uint{ident.Name: 1})
		}
	}

	return funcID, stats
}

func analyzeMultiFunctionAssignment(lhs []ast.Expr, rhs []ast.Expr) map[FuncID]*FuncCallSiteStats {
	if len(rhs) == 0 {
		return nil
	}

	// single assignment, one function on the right
	if len(rhs) == 1 {
		call, ok := rhs[0].(*ast.CallExpr)
		if !ok || call == nil {
			return nil
		}
		funcID, funcStats := analyzeSingleFunctionAssignment(lhs, call)
		return map[FuncID]*FuncCallSiteStats{funcID: &funcStats}
	}

	// multiple assignment
	stats := make(map[FuncID]*FuncCallSiteStats)

	// if function call is detected, then it is matched to single return on left hand side and analyzed as if single call
	for _, funcCall := range splitFuncCalls(lhs, rhs) {
		funcID, funcStats := analyzeSingleFunctionAssignment(funcCall.lhs, funcCall.call)

		funcStats.MultipleAssignmentWithOtherCount++

		if _, ok := stats[funcID]; !ok {
			stats[funcID] = &FuncCallSiteStats{}
		}
		mergeFuncCallSiteStats(&funcStats, stats[funcID])
	}

	return stats
}

type funcCallWithReturn struct {
	lhs  []ast.Expr
	call *ast.CallExpr
}

func splitFuncCalls(lhs []ast.Expr, rhs []ast.Expr) (funcs []funcCallWithReturn) {
	for i, expr := range rhs {
		if call, ok := expr.(*ast.CallExpr); ok && call != nil {
			funcs = append(funcs, funcCallWithReturn{lhs: []ast.Expr{lhs[i]}, call: call})
		}
	}
	return funcs
}

func analyzeSingleFunctionAssignment(lhs []ast.Expr, call *ast.CallExpr) (funcID FuncID, stats FuncCallSiteStats) {
	// TODO: returns are attributable only to final function in chain
	funcID = FinalCallerFuncIDFromCallExpr(call)

	if len(lhs) == 0 {
		stats.ReturnIgnoredCount++
	}
	if len(lhs) > 0 {
		stats.MultipleAssignmentCount = 1
	}
	for _, expr := range lhs {
		if ident, ok := expr.(*ast.Ident); ok && ident != nil {
			stats.ReturnNameCount = append(stats.ReturnNameCount, map[string]uint{ident.Name: 1})
		}
	}

	return funcID, stats
}

var NilFuncID = FuncID{}

// FuncID is an indexable type that identifies function
type FuncID struct {
	FunctionName string `json:"function_name"`
}

func NewFuncID(funcName string) FuncID { return FuncID{FunctionName: funcName} }

// FinalCallerFuncIDFromCallExpr extracts last function in call expression.
// If chain of calls and fields, then last function is only considered.
func FinalCallerFuncIDFromCallExpr(n *ast.CallExpr) FuncID {
	fnames := functionsNamesFromCallExpr(n)
	if len(fnames) == 0 {
		return NilFuncID
	}
	return NewFuncID(fnames[len(fnames)-1])
}

func functionsNamesFromCallExpr(n *ast.CallExpr) (funcs []string) {
	if ind, ok := n.Fun.(*ast.Ident); ok && ind != nil {
		return []string{ind.Name}
	}
	if sel, ok := n.Fun.(*ast.SelectorExpr); ok && sel != nil {
		funcs = append(funcs, sel.Sel.Name)
		if call, ok := sel.X.(*ast.CallExpr); ok && call != nil {
			funcs = append(funcs, functionsNamesFromCallExpr(call)...)
		}
		return funcs
	}
	return nil
}
