package main

import (
	"encoding/json"
	"flag"
	"go/token"
	"log"
	"os"

	"github.com/nikolaydubina/go-callsite-stats/analysis/callsitestats"
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

	stats := callsitestats.NewFuncCallSiteStatsMapRepo()

	for _, pkg := range pkgs {
		for _, fileAst := range pkg.Syntax {
			callsitestats.CollectFuncCallSiteStatsForFile(fileAst, stats)
		}
	}

	if outJSON {
		encoder := json.NewEncoder(os.Stdout)
		for funcID, funcStat := range stats.GetAll() {
			type FuncStatRowJSON struct {
				callsitestats.FuncID
				*callsitestats.FuncCallSiteStats
			}
			if err := encoder.Encode(FuncStatRowJSON{FuncID: funcID, FuncCallSiteStats: funcStat}); err != nil {
				log.Printf("%s\n", err)
			}
		}
	}
}
