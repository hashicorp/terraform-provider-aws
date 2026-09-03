---
name: go-conventions
description: "Fundamental Go conventions for the Terraform AWS provider. Use whenever writing or editing Go in internal/**/*.go (any resource, data source, ephemeral resource, action, test, or helper) before making the change, not only when asked about Go."
---

<!-- Copyright IBM Corp. 2014, 2026 -->
<!-- SPDX-License-Identifier: MPL-2.0 -->

# Skill: Go Conventions

Three forces pull this repository away from Go's conventions: existing code that violates them, human habits from other languages, and agent instincts trained on other ecosystems. Follow provider practice where it doesn't contradict this skill; where existing code violates this skill, it isn't precedent. Rationale and evidence: [docs/go-for-contributors.md](../../../docs/go-for-contributors.md).

## Naming

- Initialisms keep one case: `ID`, `ARN`, `API`, `VPC`, `KMS`, `URL`, `HTTP`. Write `applicationID`, never `applicationId`, `Arn`, or `Url`.
- `MixedCaps`, not underscores: `maxRetries`, not `MAX_RETRIES`. Test names (`TestAccFoo_basic`) are the exception.
- Getters drop `Get`: `Owner()`, not `GetOwner()`.
- Short locals (`c`, `i`). Receivers are one or two letters, consistent across the type, never `this` or `self`.
- Don't create packages named `util`, `common`, `misc`, `api`, `types`, or `interfaces`. If callers must alias your package, the name failed. (We ship `internal/types`; it's a known fail, not precedent.)

## Comments

- Names and structure first. A comment never compensates for code that is hard to read.
- Delete a comment that restates the line, names the obvious operation, acts as an in-function section header, paraphrases the signature, teaches Go, or explains a name you should rename instead.
- Keep comments that record constraints, invariants, surprising AWS behavior, or why an obvious approach was rejected.
- Document every exported declaration: full sentence, begins with the name, ends with a period.

## Organization: function, file, package

Three units, three costs. A separately nameable concept does not earn a boundary.

- **Function**: cheap. Create freely when it improves the code.
- **File**: organization for humans. Default to editing an existing file; a new file creates no encapsulation, ownership, or API boundary.
- **Package**: a real API and dependency boundary. Rare, and strongly justified.

Then:

- No `helpers.go`, `common.go`, or `utils.go`. Two or three callers, or "separation of concerns," don't justify a new file.
- No one-class-per-file or one-concern-per-file habits from Java or Python. Keep related code physically close.
- Prefer growing an existing package. A healthy one holds several types, files, and responsibilities (`net/http`, `os`, `flag`).
- If callers will almost always need the new package alongside its parent or neighbor, don't split it.
- If implementation code routinely imports several sibling packages sharing a path prefix, those aren't meaningful boundaries.
- If a split forces you to export what used to be unexported, the split is the mistake.

## Abstraction

- Stay concrete until multiple real uses demand otherwise. Similarity is evidence to examine, not an instruction to abstract.
- Define interfaces where they're consumed, as small as the consumer needs. Never pair an interface with its implementation by default, and never add one so it "could be mocked."
- A little copying beats a little dependency. Modest duplication is better than an abstraction that obscures control flow.
- Prefer functions and ordinary data structures over types, builders, managers, registries, and frameworks.

## Control flow and errors

- Linear and top-to-bottom. Handle the exceptional case early and return. No unnecessary `else`.
- Errors are values: return them, wrap with useful context, inspect deliberately. No exception-like infrastructure or custom error hierarchies.
- Error strings are lowercase and unpunctuated: `"reading bucket policy"`.
- Never discard an error with `_`. Never panic for an ordinary failure.

## Context

- `ctx context.Context` is the first parameter, always.
- Never store a `Context` in a struct field; pass it to each method that needs it.

## Commit messages

- Concise subject. Add a body only for intent or non-obvious constraints. A large mechanical change may need barely a word.

## Above all: don't import other languages' architecture

Go prefers concrete code, explicit control flow, locality, small consumer-driven interfaces, and modest repetition over abstraction, indirection, and machinery. When two implementations are equally correct, choose the one with fewer concepts. Absence of machinery is not unfinished work; the machinery is usually the defect.

Resist:

- Generics to remove duplication. Use them when the problem itself is generic.
- DI infrastructure. A function parameter or struct field already is dependency injection.
- Redesigning production code to enable mocking.
- Extracting tiny single-use helpers merely to shorten a function.
- Wrapper types (`Config`, `Options`, `Manager`, `Result`) for tiny concepts. Leave a string a string.
- Constructors where a useful zero value works.
- map/filter/reduce pipelines, reflection, or functional composition where a plain loop reads directly.
