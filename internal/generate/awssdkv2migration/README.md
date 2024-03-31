# awssdkv2migration

The `awssdkv2migration` generator creates [gopatch](https://github.com/uber-go/gopatch) patches that aid in the migration to AWS SDK v2.

The `awssdkv2migration` executable is called as follows:

To use with `go generate`, add the following directive to a Go file

```go
//go:generate go run <relative-path-to-generators>/generate/awssdkv2migration/main.go
```

For example, in the file `internal/service/events/generate.go`

```go
//go:generate go run ../../generate/awssdkv2migration/main.go

package events
```

generates the file `internal/service/events/aws_sdk_v2.patch`.

That patch can then be applied using:

```sh
gopatch -p internal/service/events/aws_sdk_v2.patch internal/service/events/...
```
