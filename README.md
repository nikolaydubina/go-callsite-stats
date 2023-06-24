## go-callsite-stats: analyse function callsites

[![go-recipes](https://raw.githubusercontent.com/nikolaydubina/go-recipes/main/badge.svg?raw=true)](https://github.com/nikolaydubina/go-recipes)
[![Go Report Card](https://goreportcard.com/badge/github.com/nikolaydubina/go-callsite-stats)](https://goreportcard.com/report/github.com/nikolaydubina/go-callsite-stats)

> Useful for refactoring, better naming, better signatures, OOP

```
go install github.com/nikolaydubina/go-callsite-stats@latest
```

```
go-callsite-stats -json ./...
```

```
{"function_name":"AddReference","return_ignored_count":0,"argument_name_count":[{"name":1}],"multiple_assignment_count":0,"multiple_assignment_with_other_count":0}
{"function_name":"extractRawLog","return_ignored_count":0,"return_name_count":[{"err":1}],"multiple_assignment_count":1,"multiple_assignment_with_other_count":0}
{"function_name":"CreatePVC","return_ignored_count":0,"return_name_count":[{"err":3,"pvc":8,"pvclaim":14,"tmpClaim":1},{"err":23}],"argument_name_count":[{"c":6,"client":15,"cs":2,"ns":1,"pvc":2},{"namespace":15,"ns":8},{"claim":1,"pvc":6,"pvclaimSpec":1}],"multiple_assignment_count":26,"multiple_assignment_with_other_count":0}
{"function_name":"testNotReachableHTTP","return_ignored_count":0,"argument_name_count":[{"nodeIP":2,"tcpIngressIP":1},{"svcPort":1,"tcpNodePort":1,"tcpNodePortOld":1},{"loadBalancerLagTimeout":1}],"multiple_assignment_count":0,"multiple_assignment_with_other_count":0}
```

## Further Improvements

- [ ] skip build files and not include directives
- [ ] skip generated once go@1.21 is released 
- [ ] asserting that type in methods in callsites matches target type
- [ ] filtering specific method receiver
- [ ] filtering specific declaration package name
- [ ] graph which other functions a function is called with in multiple assignments
- [ ] graph which other functions and fields a function is called in chain of call expression
- [ ] generated HTML nice visualisation
- [ ] file and code location in UI
- [ ] file tree in UI
- [ ] website with open source analysis based on pkg.go.dev

## Appendix: A

Go analysis toolchain does not support multiple packages in pass.
They recommend either export to `STDOUT` and process in separate process or to use `go/packages`.

* https://github.com/golang/go/issues/53215
* https://github.com/golang/go/issues/50265
* https://eli.thegreenplace.net/2020/writing-multi-package-analysis-tools-for-go/
