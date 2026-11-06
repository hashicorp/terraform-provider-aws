<!-- Copyright IBM Corp. 2014, 2026 -->
<!-- SPDX-License-Identifier: MPL-2.0 -->

# sdkmap test fixture

Miniature provider root for tests using `-root testdata`. Its `mock*` service packages are test fixtures: they use AWS SDK V2 clients matching the directory name, need not compile, and omit irrelevant helpers. Mock resources are named `Foo`, `Bar`, `Baz`, and `Qux`, in that order.

| Path | Fixture |
| --- | --- |
| `conns` | SDK client and non-client `AWSClient` methods |
| `service/mockbuildconstraints` | `!generate`, `generate`, and `ignore` files; internal/external tests |
| `service/mockcrossservice` | Exported finder used by another service |
| `service/mockcrud` | No Update; local/shared/cross-service finders, status/waiters, paginator, options, and retry closures |
| `service/mockcustomizediff` | References, literals, calls, variables, factories, and nested `customdiff` combinators |
| `service/mockfuncrefs` | Factory-local method values, call results, and a resource with a `schema.NoopContext` delete function |
| `service/mockimport` | Passthrough, named, literal, and absent importers |
| `service/mockinterface` | Interface dispatch and an SDK client wrapped by a package interface |
| `service/mockstructfield` | SDK clients in fields, including retry, direct/deferred orchestration, and a same-named foreign client |
| `service/mocktags` | Shared, branching tag handlers (direct/helper, grouped/default/missing cases), cross-service calls, multi-type factory, and `@Tags` without an identifier |
| `service/mocktagsgen` | Unbranched generated-style tag handlers and exported cross-service variables |
| `service/mocktagsupdateonly` | Generated-style UpdateTags; test-only ListTags; Read sets tags |
| `service/mocktagsnohandlers` | Tagged and untagged resources without tag handlers |
| `service/mockunanalyzed` | Analyzed paginator and unanalyzed waiter/presign clients |
| `service/mocknogofiles` | Only a build-excluded Go file |
| `service/mocktestonly` | Only test files |

## Golden files

`golden/<package>.json` is the sorted expected output for each fixture service. `TestFindSDKResources` compares it and requires a test case for every loaded service and golden file. Following intentional changes, regenerate then review the diff:

```bash
go test -run TestFindSDKResources -update
```
