# go-callsite-stats

what
+ skip tests
+ skip generated
+ one arg function
+ all functions
+ export in JSONL
+ basic text CLI visualisation
+ generated HTML nice visualisation
+ file and code location in UI
+ file tree in UI
+ website with open source analysis based on pkg.go.dev

This can be useful for refactoring, better naming, better signatures, OOP design ideas.

## Further Improvements

- [ ] asserting that type in methods in callsites matches target type
- [ ] filtering specific method receiver
- [ ] filtering specific declaration package name
- [ ] graph which other functions a function is called with in multiple assignments
- [ ] graph which other functions and fields a function is called in chain of call expression

## Appendix: A

Go analysis toolchain does not support multiple packages in pass.
They recommend either export to `STDOUT` and process in separate process or to use `go/packages`.

* https://github.com/golang/go/issues/53215
* https://github.com/golang/go/issues/50265
* https://eli.thegreenplace.net/2020/writing-multi-package-analysis-tools-for-go/
