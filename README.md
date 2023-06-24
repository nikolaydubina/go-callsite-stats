## go-callsite-stats: analyse function callsites

[![go-recipes](https://raw.githubusercontent.com/nikolaydubina/go-recipes/main/badge.svg?raw=true)](https://github.com/nikolaydubina/go-recipes)
[![Go Report Card](https://goreportcard.com/badge/github.com/nikolaydubina/go-callsite-stats)](https://goreportcard.com/report/github.com/nikolaydubina/go-callsite-stats)

> Useful for refactoring, better naming, better signatures, OOP

```
go install github.com/nikolaydubina/go-callsite-stats@latest
```

```
go-callsite-stats ./...
```

Output format
```
x<number of function calls>:  <var name>:<count>,<var name>:<count> = <func name>(<arg name>:<count>, <arg name>:<count>)
```

Example Kubernetes
```
x16:       (no assignments)                  = execHostnameTest(serviceAddress:7)
                                                               (nodePortAddress:3)
                                                               (nodePortAddress0:3)
                                                               (nodePortAddress1:2)
                                                               (clusterIPAddress:1)
x16:       pod:10, err:12                    = CreatePod(client:11, namespace:10, nil:9, pvclaims:6, false:7, execCommand:2)
           clientPod:1                                  (c:2, ns:2, podCount:2, true:3)
           _:1                                          (pod:1, pod:1, pvclaims:2, false:2)
           err:1                                        (ctx:1, nil:1, createdClaims:1, pvcClaims:1)
                                                        (namespace:1, nameSpace:1, podTemplate:1)
                                                        (, basePod:1)
x16:       (no assignments)                  = GET()
x16:       deployment:11, err:14             = UpdateDeploymentWithRetries(c:14, ns:14, deploymentName:3, applyUpdate:1, poll:1,pollShortTimeout:1)                                                         
           _:2                                                            (client:1, namespace:1, pollTimeout:1)
           deploymentWithUpdatedReplicas:1                                (applyUpdate:1, pollInterval:1, name:1)
x16:       err:16                            = waitForDefinition(schemaFoo:12
                                                                (schemaWaldo:3)
                                                                (expect:1)
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

Select specific functions and reformat output
```
go-callsite-stats -json ./... | grep createEndpoint | jq 
```

```
{
  "function_name": "createEndpoint",
  "call_count": 1,
  "argument_name_count": [
    {
      "i": 1
    }
  ],
  "return_name_count": [
    {
      "mu": 1
    },
    {
      "listener": 1
    },
    {
      "serverURL": 1
    },
    {
      "endpoint": 1
    },
    {
      "err": 1
    }
  ],
  "return_ignored_count": 0,
  "multiple_assignment_count": 1,
  "multiple_assignment_with_other_count": 0
}
```

## Further Improvements

- [ ] skip build files and not include directives
- [ ] skip generated once go@1.21 is released 
- [ ] asserting that type in methods in callsites matches target type
- [ ] filtering specific method receiver
- [ ] filtering specific declaration package name
- [ ] graph which other functions a function is called with in multiple assignments
- [ ] graph which other functions and fields a function is called in chain of call expression
- [ ] generated HTML nice visualization
- [ ] file and code location in UI
- [ ] file tree in UI
- [ ] website with open source analysis based on pkg.go.dev

## Appendix: A

Go analysis toolchain does not support multiple packages in pass.
They recommend either export to `STDOUT` and process in separate process or to use `go/packages`.

* https://github.com/golang/go/issues/53215
* https://github.com/golang/go/issues/50265
* https://eli.thegreenplace.net/2020/writing-multi-package-analysis-tools-for-go/
