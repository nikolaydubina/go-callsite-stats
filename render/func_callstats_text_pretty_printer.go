package render

import (
	"fmt"
	"io"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/nikolaydubina/go-callsite-stats/analysis/callsitestats"
)

// FuncCallSiteStatsTextPrettyPrinter is pretty printer aimed for CLI output
type FuncCallSiteStatsTextPrettyPrinter struct {
	w *tabwriter.Writer
}

func NewFuncCallSiteStatsTextPrettyPrinter(w io.Writer) FuncCallSiteStatsTextPrettyPrinter {
	padding := 4
	return FuncCallSiteStatsTextPrettyPrinter{
		w: tabwriter.NewWriter(w, 0, 0, padding, ' ', 0),
	}
}

// PrintUsage describing format
func (s FuncCallSiteStatsTextPrettyPrinter) PrintUsage(w io.Writer) {
	w.Write([]byte("x<number of function calls>:  <var name>:<count>,<var name>:<count> = <func name>(<arg name>:<count>, <arg name>:<count>)\n"))
}

func newOrderedCounts[T uint](mp []map[string]T) [][]string {
	vs := make([][]string, len(mp))
	for i, m := range mp {
		for q := range m {
			vs[i] = append(vs[i], q)
		}
	}
	for k := 0; k < len(vs); k++ {
		sort.Slice(vs[k], func(i, j int) bool { return mp[k][vs[k][i]] > mp[k][vs[k][j]] })
	}
	return vs
}

func renderTuple[T uint](names []string, counts []T) string {
	if len(names) == 0 {
		return ""
	}
	var s string
	for i, name := range names {
		if name == "" {
			continue
		}
		if i > 0 {
			s += ", "
		}
		s += fmt.Sprintf("%s:%d", name, counts[i])
	}
	return s
}

func slice(wall [][]string, l int) (layer []string) {
	layer = make([]string, len(wall))
	for idx := 0; idx < len(wall); idx++ {
		if l < len(wall[idx]) {
			layer[idx] = wall[idx][l]
		}
	}
	return layer
}

func mapCounts(vs []string, mp []map[string]uint) (counts []uint) {
	for i, v := range vs {
		counts = append(counts, mp[i][v])
	}
	return counts
}

func renderTupleFromMap(vs []string, mp []map[string]uint) string {
	return renderTuple(vs, mapCounts(vs, mp))
}

func (s FuncCallSiteStatsTextPrettyPrinter) Flush() { s.w.Flush() }

// EncodeAll will print sorted with most called functions at the top
func (s FuncCallSiteStatsTextPrettyPrinter) EncodeAll(mp map[callsitestats.FuncID]*callsitestats.FuncCallSiteStats) {
	var ids []callsitestats.FuncID
	for id := range mp {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return mp[ids[i]].CallCount > mp[ids[j]].CallCount })
	for _, id := range ids {
		s.Encode(id, *mp[id])
	}
}

// Encode will add single function statistics as multiline, it will sort each return value and argument form most common to least
func (s FuncCallSiteStatsTextPrettyPrinter) Encode(id callsitestats.FuncID, stats callsitestats.FuncCallSiteStats) {
	rets := newOrderedCounts(stats.ReturnNameCount)
	args := newOrderedCounts(stats.ArgumentNameCount)

	var numLines int
	for _, q := range rets {
		if len(q) > numLines {
			numLines = len(q)
		}
	}
	for _, q := range args {
		if len(q) > numLines {
			numLines = len(q)
		}
	}

	retsStr := renderTupleFromMap(slice(rets, 0), stats.ReturnNameCount)
	if retsStr == "" {
		retsStr = "(no assignments)"
	}

	funcStr := " = " + id.FunctionName
	fmt.Fprintf(s.w, "x%d:\t %s\t%s(%s)\t\n", stats.CallCount, retsStr, funcStr, renderTupleFromMap(slice(args, 0), stats.ArgumentNameCount))

	funcPlaceholderStr := strings.Repeat(" ", len(funcStr))
	for i := 1; i < numLines; i++ {
		retsStr = renderTupleFromMap(slice(rets, i), stats.ReturnNameCount)
		argsStr := renderTupleFromMap(slice(args, i), stats.ArgumentNameCount)
		fmt.Fprintf(s.w, "\t %s\t%s(%s)\t\n", retsStr, funcPlaceholderStr, argsStr)
	}

	// TODO: place based on order of mean in return counts
	if stats.ReturnIgnoredCount > 0 {
		fmt.Fprintf(s.w, "\t (ignored)%d\t%s()\t\n", stats.ReturnIgnoredCount, funcPlaceholderStr)
	}
}
