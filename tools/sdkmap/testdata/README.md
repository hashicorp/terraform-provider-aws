<!-- Copyright IBM Corp. 2014, 2026 -->
<!-- SPDX-License-Identifier: MPL-2.0 -->

# sdkmap test fixture

This directory is shaped like the root of the terraform-provider-aws repository, so that sdkmap's tests can run it with `-root testdata`.

The service packages are mocks. Each is named for the behavior it covers and uses an AWS SDK for Go v2 client of the same name, e.g. `github.com/aws/aws-sdk-go-v2/service/examplecrud`. The Go source is parsed but never compiled. It follows the shapes used by real service packages, but it is not complete: helpers that sdkmap does not analyze may be missing.

| Directory | Covers |
| --- | --- |
| `internal/conns` | `AWSClient` methods that return AWS SDK clients, and methods that don't |
| `internal/service/examplebuildconstraints` | Files included and excluded by `//go:build !generate`, `//go:build generate`, and `//go:build ignore`; test files in the package and in an `_test` package |
| `internal/service/examplecrossservice` | An exported finder variable used by other service packages |
| `internal/service/examplecrud` | A resource without Update; finders, status functions, and waiters in separate files; a finder shared by resources; a cross-service finder; a paginator; an `Options` call; closures passed to `retry.Op` and generic `tfresource.RetryWhenIsA` |
| `internal/service/examplecustomizediff` | CustomizeDiff as an identifier, `customdiff.Sequence`, nested `customdiff.All`/`customdiff.Sequence` (with a call and a function literal that calls the SDK), a function literal, `customdiff.ForceNewIf`, a package-level variable, a call returning a `schema.CustomizeDiffFunc`, and a function that calls a resource factory |
| `internal/service/examplefuncrefs` | CRUD functions as method values on a factory-local handler, CRUD functions returned by calls, and `schema.NoopContext` as a CRUD function |
| `internal/service/exampleimport` | A passthrough importer, a custom import function, a function literal importer, and no importer |
| `internal/service/exampleinterface` | A method called through a package interface with two implementations, and a client passed as a package interface that wraps an SDK operation |
| `internal/service/examplestructfield` | Clients held in struct fields: a retryer whose method calls the SDK in a generic retry closure, an orchestrator whose methods are called directly and deferred, and a struct whose field has the same name as the orchestrator's but holds another service's client |
| `internal/service/exampletags` | `ListTags`/`UpdateTags` shared by resource types and branching on them, calling another service's exported variables; one factory with several `@SDKResource` type names; `@Tags` without `identifierAttribute` |
| `internal/service/exampletagsgen` | `ListTags`/`UpdateTags` shaped like generated code, which don't branch, with exported variables for other service packages |
| `internal/service/exampletagsupdateonly` | Generated-style `UpdateTags` with `ListTags` only in a test file, and a Read that sets tags itself |
| `internal/service/exampleunanalyzed` | A client passed to an SDK waiter and to an SDK presign client, which sdkmap doesn't analyze, and to an SDK paginator, which it does |
| `internal/service/examplenogofiles` | A directory whose only Go file is excluded by a build constraint |
| `internal/service/exampletestonly` | A directory containing only test files |

## Golden files

`golden/<package>.json` holds the resources that sdkmap finds in each service package, in sdkmap's output format, sorted by resource type. `TestFindSDKResources` fails if the results differ from them.

After a change to sdkmap or to the fixture that is meant to change the results, regenerate the golden files and review the diff before committing:

```bash
go test -run TestFindSDKResources -update
```

A new service package needs a test case in `TestFindSDKResources`; a golden file without one fails the test. JSON has no comments, so the notable results recorded in the golden files are listed here:

- `exampleinterface`: `aws_exampleinterface_thing`'s Read has no SDK calls, because its client is passed as a package interface that no type in the package implements.
- `examplecustomizediff`: `aws_examplecustomizediff_variable`'s CustomizeDiff, a package-level `customdiff.All`, has its SDK calls but no `functions`; `aws_examplecustomizediff_factory_reference`'s CustomizeDiff has no SDK calls, because the resource factory it calls isn't followed.
- `examplefuncrefs`: method values, e.g. `h.Upsert`, and `schema.NoopContext` have no `location`.
- `exampletags`: each tagged resource's `tags` has the SDK calls of every resource type that shares `ListTags` and `UpdateTags`.
- `exampleunanalyzed`: SDK calls made through the SDK waiter and presign client are missing. The warnings sdkmap logs for them aren't in the golden files and aren't tested.
