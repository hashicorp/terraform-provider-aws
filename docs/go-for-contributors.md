<!-- Copyright IBM Corp. 2014, 2026 -->
<!-- SPDX-License-Identifier: MPL-2.0 -->

# Go for Terraform AWS Provider Contributors

*Applying Go's design philosophy in a very large provider codebase.*

Most of us didn't start in Go, and the provider remembers. We came from Java, Python, TypeScript, C#, and Ruby, and those habits still shape the code we write here. Some of them don't belong in Go. You don't have to take that on faith: the [Hall of Shame](#hall-of-shame) below counts the damage in our own repository, with file names. This guide is about catching the pattern before you add to it.

It isn't one maintainer's opinion about Go. Every rule below traces to Go's own guidance (*Effective Go*, *Organizing Go Code*, *Package Names*, *Go Code Review Comments*), translated into what it means inside a three-million-line provider. Where we get it wrong today, we say so.

!!! note "These are defaults, not laws"
    This document is opinionated about defaults, not absolute prohibitions. It exists because unnecessary structure is far easier to add than to remove in a repository this size. Deviate when you have a concrete reason, not because another design is theoretically cleaner. If you think a rule here is wrong, or wrong for one case, say so in review. That's how the rule gets fixed or the exception gets recorded.

## How to read this in 10, 30, or 60 seconds

- **10 seconds:** Prefer the simplest code that fits into the package you already have. Concrete over abstract, visible over clever, together over scattered. When two designs are equally correct, choose the one with fewer concepts.
- **30 seconds:** Read the table below. It's the highest-value thing here.
- **60 seconds:** Keep going. Every principle is short, and the one your reviewer keeps citing is probably in it.

## Instincts from other languages that don't transfer to Go

Read down the left column and wince at the ones you recognize. Everyone recognizes at least three.

| The instinct you brought with you | What Go asks for instead |
| --- | --- |
| Extract repeated code on sight | First decide whether the abstraction improves clarity |
| An interface for every implementation | Consumer-defined interfaces, only where needed |
| One class or concept per file | A cohesive package, files sized for humans |
| Tiny packages for "separation of concerns" | Packages that are real API or dependency boundaries |
| A `helpers`, `common`, or `utils` package | Keep code with the domain it belongs to |
| More layers means better architecture | Fewer concepts and visible control flow |
| A comment on every operation | Names and structure first, comments for the non-obvious |
| Mocking requires interfaces everywhere | Introduce seams only where they earn their keep |
| A constructor for every type | Prefer useful zero values where practical |
| Two similar things must be generalized | Generalize when the problem is general |

If you absorb only this table, you're already writing more idiomatic provider code than a fair amount of what's already merged. No names, no shame. Just read the diffs sometime.

## Hall of Shame

Everything above is easy to agree with in the abstract, so here are five current examples from this repository. Each was chosen because the pattern outlives the next refactor. Nobody involved was careless, and every one of them looked reasonable the day it was written. These instincts feel right in the moment. The provider is what they accumulate into.

**Exhibit A: `Id` in a provider made of IDs.** Go's rule on initialisms is unambiguous. Acronyms keep one case, so it's `ID`, `ARN`, `API`, `VPC`, `URL`, never `Id`, `Arn`, `Api`. In non-generated, non-test service code we have 61 local variables and 38 function names using `Id`, among them `findBaiduChannelByApplicationId` and `findADMChannelByApplicationId`. The best of them appears verbatim in several Pinpoint resources:

```go
applicationId := d.Get(names.AttrApplicationID).(string)
```

Right on the right, wrong on the left, one statement. Nobody was being sloppy. `Id` is what muscle memory types in most other languages, which is the reason this guide exists. Look at that second function name too: `ADMChannel` gets the initialism right while `ApplicationId` doesn't.

!!! tip "What we should have done"
    `applicationID := d.Get(names.AttrApplicationID).(string)`, and `findADMChannelByApplicationID`. Every one of these is an unexported local or an unexported function, so renaming breaks nothing outside its own package. This is the cheapest fix in the document. Do it in files you're already editing, and skip the 51-file rename PR; that's churn, and reviewers can't see the real change underneath it.

**Exhibit B: the package name we rename 1,372 times.** *Go Code Review Comments* lists the package names to avoid, by name: "util, common, misc, api, types, and interfaces." We ship three packages named `types`.

You don't have to take Go's word for it, because our own import blocks testify against us. Our `types` collides with the AWS SDK's `types` and with the Plugin Framework's, so nobody imports ours under the name we gave it:

```go
inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
fwtypes  "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
sdktypes "github.com/hashicorp/terraform-provider-aws/internal/sdkv2/types"
```

Those account for 708, 662, and 2 import lines, and all 1,372 of them are aliased. Not one file in the repository imports any of the three by its real name. The same document that warns against the name also says to avoid renaming imports, because "good package names should not require renaming." So the codebase reports the defect once per file, 1,372 times, with no dissent.

This one belongs to maintainers rather than to some first-time contributor, and it's the most expensive mistake in this guide, because a package name is load-bearing at every call site. Renaming now would touch well over a thousand files, so it stays.

!!! tip "What we should have done"
    Name the package what every caller already calls it. 1,372 import lines voted independently for `inttypes`, `fwtypes`, and `sdktypes`; those are the names, discovered the hard way. A package that every consumer renames on the way in has told you its real name, so believe it. Better still, much of `internal/types` is specific enough to live in the package that uses it, which deletes the import instead of renaming it. The test is simple: if you have to alias a package to use it, the name failed.

**Exhibit C: the generic filename with no generic content.** `internal/service/networkfirewall/helpers.go` exists, and its name promises the miscellany this guide warns you off. Open it, and every function (`expandEncryptionConfiguration`, `flattenCustomActions`, `expandIPSets`) is Network Firewall-specific expand, flatten, and schema code that belongs to that resource as much as any other line in the package. The problem isn't the file, it's the label. The name tells the next reader "generic, skippable," and the contents aren't either. `internal/service/comprehend/common_model.go` earns the same citation.

!!! tip "What we should have done"
    Name the file for what's in it: `encryption_configuration.go`, `custom_action.go`, `ip_set.go`. Better still, move each function into the resource file that uses it and delete the extra file. A file name is free documentation that shows up in every diff, every `git log --stat`, and every editor tab. `helpers.go` spends that budget on nothing. If you can't name a file for its contents, the contents probably don't belong together.

**Exhibit D: the package family that's never used apart.** `internal/provider/framework/resourceattribute`, `listresourceattribute`, and `datasourceattribute` are three separate packages. Search for their consumers and you find one file, `region.go`, importing all three, plus one more that pulls in `datasourceattribute` alone. Nothing imports one of these without needing a neighbor. That's the package diagnostic from later in this guide, confirmed rather than hypothetical.

This exhibit is nuance, not a verdict. Each package wraps a different Plugin Framework type (`resource/schema.StringAttribute`, `list/schema.StringAttribute`, `datasource/schema.StringAttribute`), so collapsing them isn't free. Go won't let three unrelated concrete return types share one function without an interface or generics, and this guide asks you to justify both. Whether the fix is one package with three constructors, a shared generic helper, or nothing at all is a real design conversation. Sometimes the diagnostic fires and the honest answer is "yes, and it's still the least-bad option."

!!! tip "What we should have done"
    One package, three plainly named constructors: `attribute.ResourceRegion()`, `attribute.DataSourceRegion()`, `attribute.ListResourceRegion()`. Returning three different types from three functions in one package is ordinary Go. It needs no interface, no generics, and no directory per type. `strconv` settles the question: `ParseInt`, `ParseBool`, and `ParseFloat` live together rather than in `intconv`, `boolconv`, and `floatconv`. Put the type in the function name, not the package name.

**Exhibit E: forty percent of our packages are one file.** Exhibit D isn't an isolated case, it's a sample. Excluding `internal/service/**` and `internal/generate/**`, the provider has **90 packages**. Of those, **36 are a single file** and **44 are under 150 lines**. Half of our internal architecture is a directory wrapped around one file.

The extremes make the point faster than the average:

| Package | Non-test LOC | Exported identifiers |
| --- | --- | --- |
| `internal/unique` | 19 | 1 |
| `internal/io` | 28 | 1 |
| `internal/logging` | 36 | 2 |
| `internal/yaml` | 37 | 3 |
| `internal/smithy` | 39 | 3 |

Each of those is a full API boundary: an import path, a name at every call site, a documentation unit, an exported-or-not decision. All of it wrapped around less code than the function you're probably writing right now. Several are genuinely reusable and fine. But 36 single-file packages isn't 36 independent decisions that each happened to be right. It's a habit. When the next one feels obviously necessary, that feeling is what this exhibit is documenting.

Forty percent is also the conservative floor, because counting single-file packages only catches the ones that are too small. The harder question is how many of the remaining 54 would be better merged with a neighbor, not because either is tiny but because no caller ever wants one without the other. Size is the easy metric. Cohesion is the real one.

!!! tip "What we should have done"
    Ask which existing package this belongs in before you reach for `mkdir`. Concretely: `internal/io`, `internal/yaml`, `internal/json`, and `internal/smithy` are all encoding and serialization concerns that would make one coherent package with an obvious name, instead of four import paths that each answer a fraction of one question. For the truly tiny ones, 19 lines and a single exported function, the honest home is usually the one package that calls them, unexported. A new package should be the conclusion of a design discussion, not the reflex that starts one.

## External Hall of Shame

We didn't invent all of this. Some of it we inherited, and it's worth naming, because "the library does it this way" is the most common defense for fragmentation in our own code. Sometimes that defense is fair, which is what makes it dangerous.

Here is the real import block from `internal/service/timestreaminfluxdb/db_cluster.go`. One resource, **35 imports**:

```go
"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
...
"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32planmodifier"
"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
```

Ten packages, one job: validating and plan-modifying attributes. Upstream ships **11 separate `*planmodifier` packages** and a matching spread of `*default` packages. Across the provider we import 18 distinct `resource/schema/*` packages, and 42 distinct `terraform-plugin-framework` packages in total. The split is per type rather than per concept, so `stringplanmodifier.RequiresReplace` and `boolplanmodifier.RequiresReplace` live in different packages for no reason a caller cares about. It's a type-parameter problem solved with a directory tree.

Two things follow, and the second matters more.

**We mirrored it.** `internal/framework/planmodifiers` reproduces the same shape: `boolplanmodifier`, `int32planmodifier`, `int64planmodifier`, `listplanmodifier`, `setplanmodifier`, `stringplanmodifier`. Seven packages, 15 files, 703 lines, with `LegacyValue()` defined four separate times, differing only by return type. There's even a `planmodifiers/internal/` package holding the implementation the split forced us to re-share. Consistency with the library it extends is a legitimate reason for that layout, which is exactly what makes it the strongest argument here for thinking about a package boundary before copying one.

**It's why Exhibit B hurts.** Look at that import block again: `awstypes`, the Framework's `types`, and our `fwtypes` are all in scope at once. Three packages named `types` in one file, two of which we rename to compile. Fragmented ecosystems produce name collisions, and name collisions produce the alias sprawl measured above. The costs compound.

None of this is a reason to fork a dependency or fight its idioms. Use the libraries as designed. It is a reason to stop treating upstream layout as proof that a layout is good.

### Where does the per-type package split come from?

Fair question, and the answer isn't "Java." Nothing in the enterprise-Java playbook produces `int32validator`. This is an over-application of one of Go's own rules, under a constraint that has since expired.

The rule: let the package qualify the name, so you write `bytes.Buffer` rather than `bytes.BytesBuffer`. Applied here, `stringvalidator.LengthBetween(1, 64)` reads better than `validator.StringLengthBetween(1, 64)`. The constraint: the Plugin Framework predates usable generics, and each type genuinely needs its own interface, since `validator.String` and `validator.Int64` have different method signatures. One `LengthBetween` couldn't return the right thing for both.

So the trade was 13 import paths in exchange for roughly six characters per call site. That's the entire saving. We import `actionvalidator`, `boolvalidator`, `datasourcevalidator`, `float32validator`, `float64validator`, `int32validator`, `int64validator`, `listvalidator`, `mapvalidator`, `objectvalidator`, `resourcevalidator`, `setvalidator`, and `stringvalidator`. `SizeAtMost` alone appears 1,187 times across the list, set, and map variants: the same function name in three packages.

The standard library hit the identical problem and answered it the other way, which is what makes this settled rather than a matter of taste.

- **`strconv`** converts ints, floats, bools, and uints in one package: `ParseInt`, `ParseBool`, `ParseFloat`, `FormatInt`. Not `intconv`, `boolconv`, `floatconv`.
- **`sort`** sorts every type in one package: `sort.Ints`, `sort.Strings`, `sort.Float64s`.
- **`slices`** does it generically now. One `slices.Sort` covers all of them.

Same problem, one package, slightly longer function names. Nobody has ever complained that `strconv.ParseBool` is hard to read.

!!! tip "What should have been done, and what we do now"
    Upstream: one `validator` package with type-qualified function names (`validator.StringLengthBetween`, `validator.Int64Between`), or one generic API now that generics exist. Thirteen packages to save six characters is a bad trade, and it pushes the cost onto every caller forever.

    Us: stop mirroring it. Our `internal/framework/planmodifiers` copied a shape we had no obligation to copy. One `planmodifiers` package exposing `LegacyValueString()`, `LegacyValueBool()`, `LegacyValueInt32()`, and `LegacyValueInt64()` would have fit in a single file. Consume a fragmented library at its edges, and don't reproduce its fragmentation in the code you own. Upstream's layout is a constraint to absorb, not a template to follow.

## Clarity comes from the code, not the commentary

*Effective Go* treats a good name as the beginning of good documentation, and reserves comments for what the code can't say itself. Get the name right and you may not need the comment.

The provider's own [Naming](naming.md) guidance makes the case better than a comment could. It calls out `helper` by name as a package name that "conveys zero information," and holds up `verify` as good because it tells you what the package does. Apply that test before reaching for a comment. If you have to explain what a name means, the name is the bug.

Implementation comments are where the over-commenting habit lives. 

**If a comment is doing any of the following, _delete it_:**

- Restating the line below it. "Check if the ARN is valid" above an `arn.IsARN` call, "increment the counter" above `i++`.
- Naming the obvious operation. "Create the client," "return the error," "close the file."
- Acting as a section header inside a function. "// Validation," "// Step 2." If a function needs internal signposting, the signal is about the function, not the missing comment.
- Explaining a name instead of fixing it. Rename the variable and the comment disappears.
- Paraphrasing the signature. "Takes a bucket name and returns an error."
- Repeating the AWS API reference with no provider-specific consequence. Link it or omit it.
- Teaching Go. "`defer` runs when the function returns."
- Sitting above commented-out code. Delete the code; git remembers it.
- Carrying a bare `TODO` with no issue number. In a repository this size that's archaeology, not a plan.

What survives that list is worth keeping: a constraint, an invariant, an AWS behavior that contradicts the obvious reading, or the reason an obvious approach was rejected. Comment the surprise, not the syntax.

**Exported declarations are the exception to all of it.** Their comments are the package's public documentation rather than commentary on the implementation, so document every exported type, function, method, and variable, starting with its name, even when the implementation is obvious.

## The small stuff reviewers will flag

None of these are judgment calls. They're settled Go conventions, and they come up in review constantly.

**Names**

- **Initialisms keep one case:** `ID`, `ARN`, `API`, `VPC`, `KMS`, `URL`, `HTTP`, so `applicationID`, `vpcARN`, `ServeHTTP`. Never `Id`, `Arn`, `Url`. This matters more here than in most repositories, because the provider is mostly acronyms.
- **`MixedCaps`, never underscores.** An unexported constant is `maxRetries`, not `MAX_RETRIES` or `max_retries`. Test names are the deliberate exception: `TestAccDocDBElasticCluster_basic` uses the underscore to separate subject from scenario, which is the Go test convention rather than a violation of this one.
- **Short local names.** Prefer `c` to `lineCount` and `i` to `sliceIndex`. The further a name travels from its declaration, the more descriptive it needs to be, so loop indices stay tiny and package-level vars don't.
- **Receivers are short and consistent.** One or two letters (`r` for a resource, `c` for a client), the same letter on every method of the type. Never `this`, `me`, or `self`.
- **Getters drop the `Get`.** Write `Owner()`, not `GetOwner()`. A setter is `SetOwner()`.

**Context**

- `ctx context.Context` is the **first parameter**, always.
- **Never store a `Context` in a struct field.** Pass it to each method that needs it. Provider CRUD, finders, waiters, and sweepers all take it explicitly, which keeps the call chain visible.
- Don't define custom context types, and don't accept anything but `context.Context` in that position.

**Errors**

- **Error strings are lowercase and unpunctuated:** `fmt.Errorf("reading bucket policy")`, not `"Reading bucket policy."`. They get wrapped into longer messages, where a mid-sentence capital reads like a bug. Proper nouns and acronyms stay capitalized.
- **Never discard an error with `_`.** Handle it, return it, or panic in the rare case that warrants it. `fi, _ := os.Stat(path)` is how you get a nil-pointer panic three lines later.
- **Don't panic** for ordinary failures. Errors are values, so return them.

**Data**

- **`var x []string`, not `x := []string{}`.** The nil slice is the idiomatic empty slice.
- **Know when the distinction is load-bearing.** A `nil` slice marshals to JSON `null` and an empty one marshals to `[]`, and flatten functions feeding Terraform state care about the difference. Choose deliberately instead of by accident.
- **Don't pass a pointer to save bytes.** If a function only ever reads `*s`, take a `string`. A `*string` or `*io.Reader` parameter is almost always a mistake. This doesn't apply to the AWS SDK's own `*string` fields, which use pointers to model absence.

**Tests**

- **Fail with useful messages**, `got` before `want`: `t.Errorf("Foo(%q) = %d, want %d", in, got, want)`. Assume whoever debugs it is neither you nor on your team.
- Reach for [table-driven tests](https://go.dev/wiki/TableDrivenTests) when a helper would otherwise need six call sites.

**Doc comments**

- Full sentences, starting with the name and ending with a period: `// FindBucketByName returns the bucket matching name.` That's what reads correctly in `go doc` output, which is the whole reason for the convention.

## Files organize; packages architect

This is where imported instincts do the most damage, so it gets the most space.

A Go package is every source file in one directory, compiled together, and a declaration in one file is visible across the whole package. A new file therefore creates **no** encapsulation, ownership, visibility, or API boundary. Files exist to keep a package readable by humans, nothing more.

!!! tip "Default: use an existing file"
    Creating a new `.go` file is the exception, not the routine. Noticing that some code is shared, separately nameable, or "logically a different concern" is not by itself a reason for a new file.

For files:

- No `helpers.go`, `common.go`, or `utils.go` holding a handful of functions. *Package Names* calls out `util`, `common`, and `misc` as smells.
- Two resources sharing a helper doesn't justify a new file. Keep resource-specific helpers with the resource.
- Split a file when its size genuinely hurts navigation, not because another noun can be invented.

Packages are the real boundary, and the more expensive one. *Organizing Go Code* warns plainly that it's easy to go too far splitting code into small packages and end up "bogged down in interface design" instead of getting work done.

- Don't create a package just because functionality can be independently named. "Single responsibility" doesn't mean one responsibility per package.
- Closely related functionality that callers use together usually belongs together.

!!! warning "Known fails"
    We ship three packages named `types`: `internal/types`, `internal/framework/types`, and `internal/sdkv2/types`. Go's own guidance lists `types` among the names to avoid, and every one of our 1,372 imports of them is aliased. See [Exhibit B](#hall-of-shame) for the accounting. They're load-bearing and renaming them now would touch over a thousand files, so they stay. These are known fails, not precedent.

!!! tip "Two diagnostics for over-decomposition"
    1. If ordinary code repeatedly imports several narrow sibling packages sharing one repo path prefix, the functionality was fragmented rather than decomposed.
    2. If splitting code into a new package forces you to export identifiers that used to be unexported, the split itself may be the mistake. When in doubt, leave it out and keep implementation details unexported.

The thread tying all of it together is **locality**. Code that changes together and is understood together should stay together: a resource's schema and implementation, its flatten and expand functions, its tests. Don't relocate details into another file or package because they can be labeled "validation," "conversion," or "helpers."

## Earn your abstractions

Start concrete. Add the abstraction once real usage proves it makes the program better, not because you can picture the version where it would.

- Two nearly identical expand functions don't automatically demand a generic framework.
- A single implementation doesn't need an interface so that it "could be mocked" someday.
- **Similarity is evidence to examine, not an instruction to abstract.** The second call site tells you something. Go build it.
- **A little copying is better than a little dependency.** If sharing five lines means importing a package, adding an interface, or reaching across a boundary that doesn't otherwise exist, the copy is usually the better trade. That isn't laziness. It's a real Go value rather than a consolation prize for skipping the abstraction.

Three traps, because we've all set at least one of them:

- **Interfaces are consumer-defined capabilities.** Don't pair an interface with every concrete type. Define one when a consumer needs a smaller capability than the concrete type exposes. Turning a concrete `*ec2.Client` into a twelve-method `EC2API` interface so it "could be mocked" is a Java reflex, not a Go one.
- **Generics solve generic problems.** Use standard-library generics like `slices` and `maps` where they simplify code directly. Don't build a generic abstraction because two concrete types have a similar shape. Modern Go isn't anti-generics, it's against mechanism without payoff.
- **Tests don't dictate production architecture.** Don't add interfaces, factories, wrappers, or packages to make a unit test easier to write. And don't split one test file into ten because each function "deserves its own." Organize tests along the same boundaries as the code they test.

## Keep control flow visible

Terraform resources already carry real lifecycle complexity. Hiding it behind callbacks and clever indirection makes it worse.

- Prefer straightforward, top-to-bottom control flow.
- Handle errors and exceptional cases early, then return.
- Don't extract three obvious lines into `buildRequest` and `processResponse` to make a function shorter. You've only made the reader travel further to learn less.

Here, `if err != nil` isn't boilerplate to be ashamed of. It's the plot.

## Errors are values

The provider is a translation layer between Terraform and AWS, which makes error handling load-bearing. Treat errors as ordinary values: return them, wrap them with useful context, and preserve their identity when callers need to inspect AWS errors. Don't wrap merely to produce more verbose prose, and don't build an exception-like subsystem around routine failures.

See [Error Handling](error-handling.md) for the provider's concrete patterns (`smerr` and `smarterr`, `retry.NotFound`, `errs.IsA`).

## Reach for what already exists

The provider is enormous and already solves most problems once. Before adding a package, a helper framework, a dependency, or an abstraction, check whether the standard library, the AWS SDK, the Terraform plugin libraries, or existing provider utilities already do the job. The standard library isn't showing off. It's showing you the idiom.

## When in doubt, fewer concepts

If you remember one sentence, make it this one:

> When two implementations are equally correct, prefer the one that introduces fewer concepts.

Not fewer lines. Not fewer functions. Fewer things the next contributor has to learn: fewer packages, fewer interfaces, fewer indirections, fewer exported identifiers, fewer files without a real boundary, fewer comments standing in for names that could have spoken for themselves.

Agents and humans both tend to read the absence of machinery as an unfinished job. Go reads the machinery as the defect. Fewer moving parts isn't a lack of architecture. In Go it often is the architecture.

## Sources

The guidance above is drawn from Go's own canon. When in doubt, these outrank this page:

- [Effective Go](https://go.dev/doc/effective_go)
- [Organizing Go Code](https://go.dev/blog/organizing-go-code)
- [Package Names](https://go.dev/blog/package-names)
- [How to Write Go Code](https://go.dev/doc/code)
- [Go Code Review Comments](https://go.dev/wiki/CodeReviewComments)
- [Go Test Comments](https://go.dev/wiki/TestComments)
- [Go Doc Comments](https://go.dev/doc/comment)

For provider-specific application, see also [Provider Design](provider-design.md), [Naming](naming.md), and [Error Handling](error-handling.md).
