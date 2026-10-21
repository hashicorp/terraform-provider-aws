<!-- Copyright IBM Corp. 2014, 2026 -->
<!-- SPDX-License-Identifier: MPL-2.0 -->

# Principles for Declarative Resource Modeling

**Summary:** Principles for when a fact about a resource is written as data instead of code, what form that data takes, and what it must justify to be added.  
**Created**: 2026-10-08  
**Author**: [@YakDriver](https://github.com/YakDriver)  

---

## Decision

Any fact about a resource that we write as data instead of code (registration, tags, identity, test settings) has to clear these:

0. **Eliminate first.** Dissolve, derive, or default a fact before giving it syntax.
1. **Model the resource, not our code.** A model holds durable facts about the resource, or fills a slot every peer resource has.
2. **One fact, one authority.** Every other copy is generated from it or checked against it.
3. **Prefer typed Go.** References to code are checked by the compiler, never strings.
4. **Shared over generated.** Write common behavior once. Generate only when static types require it or it's clearly simpler.
5. **Bind behavior, don't encode it.** A model names a Go function. The algorithm stays in Go.
6. **New surface pays rent.** A new field answers five questions before it's added.

These govern what gets added and how proposals to change the syntax are judged. Nothing that exists today has to migrate.

**A step has to stand alone.** Most of these principles refuse something, so used as a checklist they can block every change. A change doesn't have to reach the end state. It has to leave the provider better even if everything planned after it is abandoned. One that only pays off once later work lands is a down payment, and "a step in the right direction" is how most `@Testing` keys got added. Principle 6 still charges rent for whatever a step adds.

## Background

Annotations arrived in 2023 as four bare markers. By October 2026 there were 25 marker kinds on about 7,000 lines, and about half the growth since 2024 was `@Testing`. Each addition was defensible on its own. Nothing asked whether it had earned its place, and *a new syntax won't ask either*.

In this provider the Go code is the resource. Metadata describes what the provider and its generators need and can't read from the code. [`terraform-provider-awscc`](https://github.com/hashicorp/terraform-provider-awscc) works the other way: it's generated from CloudFormation schemas, so the spec has to carry everything. Where the spec can't (CloudFormation doesn't expose defaults consistently), awscc marks optional values `Computed` and stops detecting drift on values a practitioner omits.

Case studies of how other systems model resources declaratively, and the provider measurements behind these principles, are in [YakDriver/declarative-modeling](https://github.com/YakDriver/declarative-modeling).

## The principles

### 0. Eliminate first

A fact can be *dissolved* (its consumer becomes generic and stops needing it), *derived* (computed from an existing authority), or *defaulted* (written only where it differs). Then there's nothing to author, review, or let go stale.

```go
Check: testAccCheckThingExists(ctx, resourceName, &v)       // v's type is authored separately
Check: acctest.CheckExists(ctx, resourceName, findThingByID) // the type follows from the finder
```

A derivation that's wrong has to fail loudly, at compile or generate time. A heuristic that's silently wrong some of the time is worse than an authored value, because the authored value at least gets reviewed. A default has to be the value whose mistakes show. Absence should mean "this test runs," never "this test is skipped," since nobody notices a skipped test.

- Upjet derives everything it can from the Terraform schema and asks for input only on what's left ([docs](https://github.com/crossplane/upjet/blob/main/docs/configuring-a-resource.md)).
- Magic Modules fills defaults centrally and omits them when writing YAML, so authored files shrink as defaults are added.
- Here, region: the default is real code, and 43 of 2,696 resources declare it. The generated registration still writes the value out, so what shrinks is the authored surface.

### 1. Model the resource, not our code

The test: is the fact still true if we rewrite the Go? Type name, taggability, and identity are. A factory's name, the Go type a finder returns, and whether tests skip a scenario aren't. Bind implementation details by checked reference (principle 3). Test policy belongs with the tests.

A fact also qualifies if it fills a slot every peer has. A list resource's operation can be modeled because every list resource has one, so a shared implementation can serve them. A flag that makes one resource behave differently from its peers doesn't qualify.

- CloudFormation's schema describes resource shape, and the handlers stay code ([schema](https://docs.aws.amazon.com/cloudformation-cli/latest/userguide/resource-type-schema.html)).
- Ours, negative: `@ArnFormat(attribute=...)` and `@IdentityAttribute(testNotNull=...)` are test-only keys on markers that otherwise describe production behavior.

### 2. One fact, one authority

Each fact has one owner, either the model or the implementation. *Model* means data with no control flow. *Implementation* means code that acts. A `resource.Spec{…}` literal is model and `findThingByID` is implementation, and both are Go. Splitting facts between the two is normal. The defect is a single fact owned in two places with nothing to notice when they disagree.

```text
model says  foo.Bar    finder returns  *foo.Bar    test repeats  foo.Bar
```

A hand-maintained index or registry of facts that already live elsewhere is that same defect. Generate it instead.

- Pulumi's AWS schema is generated by bridging this provider, so it's a projection of the provider and not a second authority.
- Ours, negative: `@Testing(existsType=...)` restates, as a string, the type the finder already returns, about 500 times.

### 3. Prefer typed Go

```go
factory = "newThingResource" // rename the function and this is silently wrong
Factory: newThingResource    // the compiler checks it, and renames carry it along
```

*Declarative* describes the artifact, not the language. `json:"name,omitempty"` is Go, and nobody calls it implementation. A resource model also never leaves the repository. Practitioner configuration has to be a text format because it crosses a process boundary, but a model doesn't cross one, so serializing it just gives up the compiler. A serialized format earns its place when the data names no Go symbols and is bulk data read by many consumers, the way `names_data.hcl` is.

- Kubernetes `controller-gen` silently ignores `validation:enum` (it should be `Enum`). A [production write-up](https://ahmet.im/blog/crd-generation-pitfalls/) concludes that markers need a separate linter. Ours: `@Testing(tagsTests=false)` in `elasticache/reserved_cache_node.go` has been wrong since March 2023, because the key is `tagsTest`. It's the only unrecognized key in the tree, and nothing has reported it in the three and a half years since.
- Protobuf got lint and [breaking-change detection](https://buf.build/docs/breaking/) from Buf more than a decade after the format existed, and it took a company to build them. A new format means building a new toolchain.
- Caveat: Upjet's configuration is typed Go and still carries `OmittedFields []string`. A reference to a schema attribute is still a string unless a constant exists, so use `names.Attr*`.

### 4. Shared over generated

*Generated* leaves N copies of code. *Shared* leaves one implementation plus N declarations. Shared is easier to change and harder to explain for any one resource; generated is the reverse. Prefer shared unless static types or methods need generation, or generation is clearly simpler.

Merging N implementations into one exposes differences nobody chose. Pick one answer instead of turning every difference into an option.

- controller-runtime: one `Reconciler` interface serves thousands of Kubernetes controllers.
- ent goes the other way and generates a typed client from Go schemas, because the static types require it. Representation (principle 3) and consumer (principle 4) are separate choices.
- [PR 34448](https://github.com/hashicorp/terraform-provider-aws/pull/34448) tried a generic Framework base here. It was closed unmerged with no recorded reason, but declaring one resource took `framework.NewResource[indexResource, indexResourceModel, *indexResource]()`. Go generics charge for their reach at the declaration site.

### 5. Bind behavior, don't encode it

```go
Read: findThingByID              // binds: checked, logic stays in Go
read_function = "findThingByID"  // encodes: unchecked, resolved by convention
```

Retries, waiters, pagination, and import each mix data (target states, retryable error codes, token fields, ID format) with an algorithm (polling, predicates, loops, parsing). Model the data and keep the algorithm in Go. A declaration that carries closures doing real work has turned back into implementation.

- Upjet binds per-resource behavior as Go closures (`SetIdentifierArgumentFn`), with no string indirection.
- Magic Modules covers behavior with about 90 resource fields, plus `custom_code` injection, plus a [handwritten escape hatch](https://googlecloudplatform.github.io/magic-modules/develop/custom-code/). That's what a spec looks like when it tries to hold the algorithm.

### 6. New surface pays rent

The change that adds a field answers:

- **Authority:** who owns it, and what keeps every other copy correct?
- **Consumers:** what reads it? If only one thing does, should the fact live there instead?
- **Validation:** what rejects a bad value, and when?
- **Failure mode:** what happens if it's missing, misspelled, or stale? "Nothing" is only acceptable when absence is a safe default (principle 0).
- **Lifetime:** is it permanent, compatibility history, or temporary, and how will we know when to remove it?

- Magic Modules validates its YAML against a schema and still has eleven `exclude_*` opt-outs and a field whose source comment reads `1 instance used`. A schema catches malformed values, not unneeded ones.
- Ours: in October 2026, `@Testing` had 53 keys, and 13 were used once.
- Upjet v2 has users apply the same configuration to two provider objects during a migration, and its docs concede the duplication. Migration facts outlive the migration unless something says when they end.

## Consequences

- Principle 6 adds friction for anyone adding metadata. That's intended, but the questions need to be in the contributor guide, not only in review comments.
- A serialized format that refers to Go symbols contradicts principle 3, so a proposal for one has to argue against principle 3 directly.
- Some facts fail principle 0 only because the tooling (a generic consumer, a derivation that fails loudly) doesn't exist yet. Record them that way rather than treating them as settled.
