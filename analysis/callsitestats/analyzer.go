package callsitestats

import "go/ast"

// CollectFuncCallSiteStatsForFile can be used in analyzer and in other static analysis tools.
// Not thread safe.
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

	return finalCallerFuncIDFromCallExpr(n), stats
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

	return finalCallerFuncIDFromCallExpr(call), stats
}

// NilFuncID is zero value
var NilFuncID FuncID

// FuncID is an indexable type that identifies function
type FuncID struct {
	FunctionName string `json:"function_name"`
}

// finalCallerFuncIDFromCallExpr extracts last function in call expression.
// If chain of calls and fields, then last function is only considered.
func finalCallerFuncIDFromCallExpr(n *ast.CallExpr) FuncID {
	switch n := n.Fun.(type) {
	case *ast.Ident:
		return FuncID{FunctionName: n.Name}
	case *ast.SelectorExpr:
		return FuncID{FunctionName: n.Sel.Name}
	default:
		return NilFuncID
	}
}
