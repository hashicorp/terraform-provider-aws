# listpages

The `listpages` generator creates paginated variants of AWS Go SDK functions that return collections of objects where the SDK does not define them. It should typically be called using [`go generate`](https://golang.org/cmd/go/#hdr-Generate_Go_files_by_processing_source).

For example, the EC2 API defines both [`DescribeInstancesPages`](https://docs.aws.amazon.com/sdk-for-go/api/service/ec2/#EC2.DescribeInstancesPages) and  [`DescribeInstances`](https://docs.aws.amazon.com/sdk-for-go/api/service/ec2/#EC2.DescribeInstances), whereas the EventBridge API defines only [`ListEventBuses`](https://docs.aws.amazon.com/sdk-for-go/api/service/eventbridge/#EventBridge.ListEventBuses).

The `listpages` executable is called as follows:

```console
$ go run main.go -ListOps <function-name>[,<function-name>] [<generated-lister-file>]
```

* `<function-name>`: Name of a function to wrap
* `<generated-lister-file>`: Name of the generated lister source file, defaults to `list_pages_gen.go`

Optional Flags:

* `-Paginator`: Name of the pagination token field (default `NextToken`)
* `-Export`: Whether to export the generated functions

To use with `go generate`, add the following directive to a Go file

```go
//go:generate go run <relative-path-to-generators>/generate/listpages/main.go -ListOps=<comma-separated-list-of-functions>
```

For example, in the file `internal/service/events/generate.go`

```go
//go:generate go run ../../generate/listpages/main.go -ListOps=ListEventBuses,ListRules,ListTargetsByRule

package events
```

generates the file `internal/service/events/list_pages_gen.go` with the functions `listEventBusesPages`, `listRulesPages`, and `listTargetsByRulePages` as well as their `...WithContext` equivalents.
