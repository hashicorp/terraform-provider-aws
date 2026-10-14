<!-- Copyright IBM Corp. 2014, 2026 -->
<!-- SPDX-License-Identifier: MPL-2.0 -->

# Principles for Declarative Resource Modeling

**Summary:** Establish principles governing when a fact about a resource should be expressed as declarative data rather than as code, which representation that data should take, and what it must justify before being admitted.  
**Created**: 2026-10-08  
**Author**: [@YakDriver](https://github.com/YakDriver)  

---

The provider carries some information about resources as data rather than behavior: the Terraform type name, whether a resource supports tags, how its identity is composed. That data has lived in a central map and in comment annotations, and it could live in typed Go declarations, a serialized specification, or a mix.

In this provider, **the Go code is the resource**. Declarative data describes facts about the code that the provider and its generators need and can't get from the code itself. A provider generated from a specification works the other way: the spec is the resource and the code is output, so the spec has to carry everything, down to URLs, verbs, update masks, and retry behavior. HashiCorp ships both for AWS. [`terraform-provider-awscc`](https://github.com/hashicorp/terraform-provider-awscc) is generated from CloudFormation schemas; this provider is hand-written, and these principles assume that.

They don't adopt or retire any syntax. They're the constraints a representation should satisfy and the questions a reviewer should ask when someone proposes adding to the data surface. A proposal that meets them with comment annotations is acceptable; one that violates them with a well-designed schema isn't.

## Background

**The surface grows on its own.** Annotations arrived in 2023 as four bare markers with no arguments. By 2026 they spanned dozens of marker kinds and thousands of lines, and about half of the growth since 2024 was a single family for test generation. Each addition was defensible on its own, and nothing asked whether a new fact had earned its place.

**Changing the format doesn't change that.** [Magic Modules](https://googlecloudplatform.github.io/magic-modules/), which generates the Google Cloud providers from schema-validated YAML, shows the same symptoms in the territory annotations cover. In October 2026 its resource definition had eleven separate opt-out flags and a field whose own source comment says one resource uses it. A schema checks syntax, not judgment.

**A specification can't always express the contract.** CloudFormation doesn't expose property defaults consistently, so awscc marks optional values `Computed`. That avoids spurious diffs, but it stops detecting drift when a practitioner omits a value. Generating a provider from an upstream specification doesn't automatically keep the Terraform semantics a hand-written provider implements.

**Comment markers go unchecked unless someone builds the checker.** Kubernetes `controller-gen` generates resource definitions from Go marker comments and silently ignores markers it doesn't recognize, so `validation:enum` instead of `validation:Enum` does nothing at all. A [production write-up](https://ahmet.im/blog/crd-generation-pitfalls/) links a code search finding many projects that hit exactly that, counts four different ways to mark a field optional, and concludes that a weakly typed marker system needs a separate static analysis tool beside it. That project has more resources behind it than this one and the same failure mode. Unchecked references are a property of the mechanism, not of the team using it.

## Terminology

Three distinctions carry most of the weight here, and each is commonly read narrower than intended.

**Model and implementation.** A model is data describing a resource; an implementation is code that acts. The test is control flow. A composite literal has none, and something else decides what to do with it. A function with loops and branches is the thing doing. Both are written in Go in this provider, and that isn't a contradiction.

**Declarative describes the artifact, not the language hosting it.** Go struct tags are the existence proof: `json:"name,omitempty"` is metadata nobody mistakes for implementation, and it is Go. The converse holds too — a CI pipeline in YAML with conditional steps and inline shell is imperative code in a format that merely looks declarative. So "it's Go, therefore it's implementation" doesn't follow. What matters is whether the artifact describes or executes. Typed Go declarations are a specification whose file extension is `.go`.

There is a real objection inside that one, and principle 5 is the guard for it: a Go declaration *can* slide into implementation by carrying closures that do work rather than name it. That's a failure mode to prevent, not evidence the category doesn't exist.

**Three things get called configuration, and only one is this document's subject.**

| | Written by | Read | Must it serialize? |
|---|---|---|---|
| Practitioner configuration | practitioners | at Terraform runtime | yes — it crosses a process and trust boundary |
| **Resource model** | us | at build time, by a generator or the compiler | no — it never leaves the repository |
| Implementation | us | executed | not applicable |

Intuitions from the first row don't transfer to the second. Practitioner configuration has to be a text format because it crosses a boundary and is written by people without our toolchain. A resource model crosses no boundary, so serializing it buys nothing and gives up the checking a compiler does for free.

**Generated implementation and shared implementation.** Both remove duplication. They differ in where the behavior ends up.

| | Generated | Shared |
|---|---|---|
| What exists after the build | N copies of similar code | one implementation plus N declarations |
| Where behavior lives | in each resource | in one place, away from the resource |
| Changing that behavior | edit the generator, regenerate everything | edit one function |
| Explaining one resource | read its generated code | read the shared code and the declaration |
| Merge conflicts | frequent, resolved by regenerating | rare, resolved by hand |

**These two choices are independent,** which is why principles 3 and 4 are separate principles rather than one. Principle 3 picks a row below; principle 4 picks a column. Every cell is occupied by a real system:

| | Consumed by a generator | Consumed by a shared implementation |
|---|---|---|
| **Typed host-language values** | ent: Go schema to typed client | Upjet: Go configuration to one generic reconciler |
| **Serialized format or markers** | Magic Modules: YAML to provider | Kubernetes: resource definition to one API server |

## Decision

A resource model is worth having only after elimination, only for durable resource semantics, with one authority, with checked references into the implementation, and without absorbing imperative behavior.

### 0. Eliminate before modeling

A fact that can be safely dissolved, derived, or defaulted needs no per-resource home, syntax, or migration. In order of preference:

- **Dissolved:** the fact stops being needed because its consumer becomes generic.
- **Derived:** the fact is computed from an existing authority.
- **Defaulted:** the fact is written only where it differs from the common case.

Dissolving is strongest, because nothing is authored or read and nothing can go stale. A test helper that needs the Go type a finder returns can take the finder instead:

```go
// The type is data that has to be authored, kept correct, and validated.
Check: testAccCheckThingExists(ctx, resourceName, &v) // v's type declared separately

// The type follows from the finder. Nothing to author.
Check: acctest.CheckExists(ctx, resourceName, findThingByID)
```

**Making a consumer generic exposes differences that were never decided.** When one implementation replaces several, every difference between them must be resolved or promoted to a field. A difference nobody chose is not a requirement, and the shared implementation should pick one answer. A field is warranted when two resources *need* to differ, not when two *happen* to differ. Harvesting every observed difference into options produces the same accreted surface as adding fields one at a time, reached faster and all at once.

**"Safely" is a requirement.** A derivation is acceptable only if a wrong result fails loudly, preferably as a compile or generation error. A heuristic that's usually right and silently wrong the rest of the time is worse than an authored value, which at least gets reviewed. An ambiguous derivation refuses rather than guesses.

The bar is that failure is *certain*, not that it's immediate. A wrong value that becomes a compile error only once generated code exists is still checked, because regenerating and compiling are enforced rather than remembered. Latency is a cost worth reducing, not what makes a derivation unsafe. What fails this test is a wrong result that compiles and merely behaves incorrectly, with nothing asserting the difference.

Derivation rules and defaults are still written somewhere, as generator logic or policy. What goes away is the per-resource declaration. Region behavior works this way: the default is real code, and most registrations never mention region.

A default pays off most when adopting it also deletes existing authored values. One that only helps new resources leaves the surface as large as it was.

**A default decides what absence means, so it has to be the value whose wrong application shows.** Ask what a wrong default produces. If it produces activity — behavior that runs and can fail, an assertion that can break — absence is self-checking, and "intentionally default" versus "forgot to set it" is a distinction without a consequence. If it produces inactivity, nothing fails and the two are indistinguishable.

Defaults that omit are therefore the hazardous direction. One that skips a test scenario loses coverage invisibly, because nothing runs and no one can tell an unsupported scenario from an overlooked one. Orient the default so silence is the safe answer: absence means the scenario is exercised, and skipping is authored where a reviewer sees it. Where the common case genuinely is "off," defaulting can't be made safe, and the fact should be derived from something already true about the resource.

### 1. Model the resource, not our implementation

A fact belongs in a model when it describes the resource rather than our implementation, is durable, and either is part of the public contract or fills a slot that a shared consumer already has. A special case added to one consumer doesn't qualify.

**This asks a different question than principle 0.** Principle 0 asks whether a fact needs to be authored at all. Principle 1 asks where an authored fact belongs. A fact can survive elimination and still not belong in a resource model — test policy is the clearest instance.

The first test does most of the work. A fact about the resource stays true if we rewrite the Go; a fact about our implementation doesn't. The Terraform type name, taggability, and identity composition are resource facts. A factory's name, or the Go type a finder happens to return, is implementation. If a model needs an implementation detail, it binds to it by checked reference (principle 3) and never transcribes it.

Test policy fails the first test too. Whether a resource's tests skip a scenario describes our tests, not the resource, so it belongs with the tests. Keeping it there also means a test-only change stays inside test files, so a reviewer can see at a glance that nothing a practitioner depends on moved.

The last test is about the consumer's shape, not how many consumers there are. Practitioners depend on a resource's type name and identity whether one consumer reads them or five do. And a single consumer is enough when it's shared: a list resource's operation and result collection fill slots every list resource has, so modeling them lets one implementation serve many resources. That's what makes principle 4 possible. A flag that makes one resource behave differently from its peers is the special case to refuse. "Important to us" isn't enough; "fills a uniform slot" or "visible to practitioners" is.

### 2. One fact, one authority

Each fact has exactly one authority, either the declarative model or the imperative implementation. Every other copy is generated from it or checked against it by a compiler, a failing generator, a lint pass, or a test.

A `resource.Spec{…}` literal in a `.go` file is model; `findThingByID` is implementation. Both are Go.

A view that aggregates facts from many authorities — an index, a census, a registry — is a derived copy and belongs to whatever generates it. Authoring one alongside the declarations it summarizes creates the duplication this principle exists to prevent.

Having some facts in the model and others in the implementation is normal. Magic Modules has handwritten resources and injected code, and a CloudFormation schema describes shape while its handlers stay implementations. The defect is one fact owned in several places with nothing to notice when they disagree:

```text
model says the result type is  foo.Bar
finder actually returns        *foo.Bar
test code repeats              foo.Bar
a generator assumes a naming convention links them
```

That defect survives any change of syntax. A new representation that doesn't settle authority brings it along.

### 3. Prefer typed Go

Anything that refers to a Go symbol or type uses a checked Go reference, not a string:

```go
factory = "newThingResource" // nothing checks it exists; rename and delete leave it wrong
Factory: newThingResource    // checked by the compiler; renames with the function
```

The default comes from an asymmetry. Typed declarations can be serialized to JSON or YAML for an outside consumer with a little code. A serialized format gets type checking back only by adding a schema, a validator, positional diagnostics, editor support, and schema migrations. Typed Go has its own cost: generators have to interpret literals or load packages to read values, and that has to stay fast enough for `make gen PKG=xyz`.

Protobuf is the measure of that second cost. The format existed for over a decade before Buf supplied the linting, formatting, and [breaking-change detection](https://buf.build/docs/breaking/) that make it pleasant to work in, and that took a company. Choosing Go starts with the toolchain already built.

Typed Go checks references to Go symbols. It does not automatically check references to schema attributes, which stay strings unless constants exist for them — Upjet's per-resource configuration is typed Go and still carries `OmittedFields []string`. This provider generates `names.Attr*` constants, and a model naming attributes should use them rather than literals.

A serialized format earns the trade when the data has no references into code and either needs language-independent consumption or is naturally maintained as bulk domain data rather than program structure. Service-level naming data is the second kind: a table of scalars for every service, read by many generators and projected into Go constants, CI configuration, repository labels, and website metadata. Nothing in it names a Go symbol, so Go's type system has nothing to check, and writing it as composite literals would add thousands of lines of syntax and verify none of it.

### 4. Prefer shared implementation over generated implementation

Express common behavior once and drive it from data where practical. Generate code when static types or methods require it, or when generating produces a materially simpler or safer system than the runtime indirection it would replace.

Generated code is volume that has to be stored, reviewed, regenerated, and kept consistent. A generic implementation parameterized by typed data avoids that, and the provider does this where behavior is uniform. Two things argue the other way. Generated code can be read, diffed, debugged, and compared against its inputs when behavior is surprising. That isn't the same as being reviewed (nobody reads a two-thousand-line generated test file line by line), but it's what makes a failure tractable. And some things need static types or methods that no runtime indirection supplies. Almost anything can be made generic with enough indirection, so the question is whether it's simpler and safer, and sometimes it isn't.

This provider has attempted it before. [PR 34448](https://github.com/hashicorp/terraform-provider-aws/pull/34448) proposed a generic embedded structure supplying the Framework CRUD methods and upcalling to simplified per-resource methods, across 21 files. It was closed unmerged with no recorded rationale, so it says nothing about whether the approach works. Its ergonomics do say something: instantiating the shared type read `framework.NewResource[indexResource, indexResourceModel, *indexResource]()` — three type parameters to declare one resource. Generic shared implementations in Go tend to pay for their reach at the declaration site, and that cost belongs in the comparison.

### 5. Keep imperative behavior in Go

A model may select or bind behavior. It must not encode the behavior itself.

```go
Read: findThingByID              // binding: checked, logic stays in Go
read_function = "findThingByID"  // encoding: unchecked, resolved by convention
```

Contested cases usually mix the two. Retry, waiter, import, and pagination logic each combine declarative policy with an imperative algorithm:

| Potentially data | Behavior |
|---|---|
| waiter target and failure states | the polling algorithm |
| retryable error codes | a custom retry predicate |
| pagination token fields | a custom pagination loop |
| import ID format | parsing with conditionals |

Model the declarative half when it passes principles 0 and 1, and keep the algorithm in Go. A model that expresses the algorithm is on its way to being a programming language with no type system.

### 6. Make new model surface pay rent

A new field or capability answers five questions in the change that proposes it:

- **Authority:** who owns this fact, and what keeps every other copy correct?
- **Consumers:** what reads it? If one thing, is that consumer shared across many resources, or should the fact live inside it?
- **Validation:** what rejects a bad value, and at build, generate, or run time?
- **Failure mode:** what happens when it's absent, misspelled, or stale? Silence isn't an acceptable answer, unless absence is a default whose wrong application is visible (principle 0).
- **Lifetime:** is this permanent resource semantics, compatibility history, or temporary machinery, and how will we know when it can be removed?

Lifetime matters because the fields most likely to become permanent are the ones meant to be temporary. Compatibility and migration facts record a moment. Without a stated end condition, nothing ever says they're obsolete, and they look exactly like permanent resource semantics.

This is the only principle that keeps working after adoption. Data surfaces like this one tend to start well designed and accumulate anyway, one defensible field at a time. A field that can't answer the questions isn't necessarily a bad idea, but it isn't ready, and asking usually ends with principle 0 eliminating it.

## Applying the Principles

The principles form a sequence, and each step matters only if the previous answer was yes:

```text
0  Does this fact need to exist as authored data?   usually no, stop here
1  Does it belong in the model, or somewhere else?
2  What is its single authority?
3  What representation keeps its references checked?
4  Should the data drive shared behavior or generated behavior?
5  Where does data end and behavior begin?
6  What keeps this surface from accreting forever?
```

- **They govern what enters a model, not everything that exists.** A fact that fails them isn't scheduled for removal, and many will stay where they are. The principles constrain growth and guide change; they don't mandate migration.
- **They apply to proposals as well as code:** reviewing a change that adds metadata, and evaluating a proposal to change the representation. Moving facts to a new syntax without reducing their number, settling their authority, or improving their checking doesn't improve anything these principles care about.

## Consequences/Future Work

- **Principle 6 adds friction for contributors adding metadata.** That's intended, but contributors may not know why, so the questions belong in the contributor guide, not only in review comments.
- **Principle 3 rules out a serialized format that carries references to Go symbols.** A proposal for one has to argue against principle 3 directly.
- **Principles 1 and 4 need judgment.** "Describes the resource" and "materially simpler or safer" are clear in common cases and arguable at the margins. Arguing the margins case by case is better than a precise rule that gets them wrong.
- **Principle 0 can depend on tooling that doesn't exist yet,** such as a generic consumer or a derivation that fails loudly. A fact that fails principle 0 only for that reason should be recorded as such, not treated as settled.

Follow-up work, none of it required by this decision:

- Audit the metadata surface against principles 0 and 2 for facts that are derivable, dissolvable, or duplicated without a check.
- Decide where validation lives, so principle 6's failure-mode question has a standard answer.
- Put the five rent questions where contributors will see them.

## References

- [CloudFormation resource provider definition schema](https://docs.aws.amazon.com/cloudformation-cli/latest/userguide/resource-type-schema.html)
- [Magic Modules resource reference](https://googlecloudplatform.github.io/magic-modules/reference/resource/)
- [Magic Modules custom code](https://googlecloudplatform.github.io/magic-modules/develop/custom-code/): the escape hatches a spec-generated provider needs
- [`terraform-provider-awscc`](https://github.com/hashicorp/terraform-provider-awscc)
- [Kubernetes CRD generation pitfalls](https://ahmet.im/blog/crd-generation-pitfalls/): what an unchecked marker system costs at scale
- [Upjet resource configuration](https://github.com/crossplane/upjet/blob/main/docs/configuring-a-resource.md): typed Go declarations driving a generic runtime
- [PR 34448](https://github.com/hashicorp/terraform-provider-aws/pull/34448): this provider's earlier shared-implementation proposal
