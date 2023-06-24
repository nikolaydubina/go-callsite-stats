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
		tests          bool
		outJSON        bool
	)
	flag.BoolVar(&tests, "tests", false, "include tests")
	flag.BoolVar(&outJSON, "json", false, "output as JSONL to STDOUT")

	flag.Parse()
	if flag.NArg() == 0 {
		log.Fatal("missing package pattern (e.g. ./...)")
	}

	packagePattern = flag.Args()[0]

	var fset = token.NewFileSet()

	mode := packages.NeedName | packages.NeedTypes | packages.NeedSyntax | packages.NeedTypesInfo
	cfg := &packages.Config{
		Fset:  fset,
		Mode:  mode,
		Tests: tests,
	}
	pkgs, err := packages.Load(cfg, packagePattern)
	if err != nil {
		log.Fatal(err)
	}

	stats := NewFuncCallSiteStatsMapRepo()

	for _, pkg := range pkgs {
		for _, fileAst := range pkg.Syntax {
			CollectFuncCallSiteStatsForFile(fileAst, stats)
		}
	}

	if outJSON {
		encoder := json.NewEncoder(os.Stdout)
		for funcID, funcStat := range stats.GetAll() {
			type FuncStatRowJSON struct {
				FuncID
				*FuncCallSiteStats
			}
			if err := encoder.Encode(FuncStatRowJSON{FuncID: funcID, FuncCallSiteStats: funcStat}); err != nil {
				log.Printf("%s\n", err)
			}
		}
	}
}

// CollectFuncCallSiteStatsForFile can be used in analyzer and in other static analysis tools
// https://go.dev/ref/spec#Assignment_statements
func CollectFuncCallSiteStatsForFile(file *ast.File, stats FuncCallSiteStatsMapRepo) {
	ast.Inspect(file, func(n ast.Node) bool {
		switch n := n.(type) {
		case *ast.CallExpr:
			if funcID, funcStats := analyzeFuncCallArguments(n); funcID != NilFuncID {
				stats.Add(funcID, funcStats)
			}
		case *ast.AssignStmt:
			analyzeMultiFunctionAssignment(stats, n.Lhs, n.Rhs)
		}
		return true
	})
}

func analyzeFuncCallArguments(n *ast.CallExpr) (FuncID, FuncCallSiteStats) {
	stats := FuncCallSiteStats{CallCount: 1}

	for _, expr := range n.Args {
		if ident, ok := expr.(*ast.Ident); ok && ident != nil {
			stats.ArgumentNameCount = append(stats.ArgumentNameCount, map[string]uint{ident.Name: 1})
		}
	}

	return FinalCallerFuncIDFromCallExpr(n), stats
}

func analyzeMultiFunctionAssignment(stats FuncCallSiteStatsMapRepo, lhs []ast.Expr, rhs []ast.Expr) {
	if len(rhs) == 0 {
		return
	}

	// single assignment, one function on the right
	if len(rhs) == 1 {
		call, ok := rhs[0].(*ast.CallExpr)
		if !ok || call == nil {
			return
		}
		funcID, funcStats := analyzeSingleFunctionAssignment(lhs, call)
		if funcID == NilFuncID {
			return
		}
		stats.Add(funcID, funcStats)
		return
	}

	// multiple assignment
	// if function call is detected, then it is matched to single return on left hand side and analyzed as if single call
	for _, funcCall := range splitFuncCalls(lhs, rhs) {
		funcID, funcStats := analyzeSingleFunctionAssignment(funcCall.lhs, funcCall.call)
		if funcID == NilFuncID {
			continue
		}

		funcStats.MultipleAssignmentWithOtherCount++
		stats.Add(funcID, funcStats)
	}
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

func analyzeSingleFunctionAssignment(lhs []ast.Expr, call *ast.CallExpr) (FuncID, FuncCallSiteStats) {
	var stats FuncCallSiteStats

	if len(lhs) == 0 {
		stats.ReturnIgnoredCount++
	}
	if len(lhs) > 1 {
		stats.MultipleAssignmentCount = 1
	}
	for _, expr := range lhs {
		if ident, ok := expr.(*ast.Ident); ok && ident != nil {
			stats.ReturnNameCount = append(stats.ReturnNameCount, map[string]uint{ident.Name: 1})
		}
	}

	return FinalCallerFuncIDFromCallExpr(call), stats
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
	switch n := n.Fun.(type) {
	case *ast.Ident:
		return NewFuncID(n.Name)
	case *ast.SelectorExpr:
		return NewFuncID(n.Sel.Name)
	default:
		return NilFuncID
	}
}
