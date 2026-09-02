---
name: go-conventions
description: "Fundamental Go conventions for the Terraform AWS provider. Use whenever writing or editing Go in internal/**/*.go — any resource, data source, ephemeral resource, action, test, or helper — before making the change, not only when asked about Go."
---

<!-- Copyright IBM Corp. 2014, 2026 -->
<!-- SPDX-License-Identifier: MPL-2.0 -->

# Skill: Go Conventions

The provider already contains anti-patterns. Three forces pull code away from Go's conventions: existing code that violates them, human habits carried from other languages, and agent instincts trained on other ecosystems. Follow established provider practices where they do not contradict these conventions; where existing code violates them, do not treat it as precedent — apply these conventions to new and edited code regardless of what surrounds it.

## Code comments

Clarity comes from code first. Comments document the API or explain the non-obvious. They never compensate for code that is hard to read.

- Make the code self-explanatory before reaching for a comment. Prefer good names, small functions, and obvious structure over implementation comments.
- Comment intent, constraints, invariants, surprising behavior, or context the code cannot express. Never narrate syntax.
- Document every exported declaration. Doc comments are the package's public API — write them even when the implementation is obvious. Begin with the declaration's name.
- Delete comments that restate the code.

## Code organization: function, file, package

Three units, three costs. Match the boundary to the cost — noticing a separately nameable concept does not imply it deserves its own structural boundary.

- **Function** — cheap. Create freely when it improves the code.
- **File** — lightweight organization inside a package. Default to modifying an existing file; a new file creates no encapsulation, ownership, or architecture.
- **Package** — a significant API and dependency boundary (visibility, imports, call-site vocabulary, docs). Creation is uncommon and strongly justified.

Files:

- Do not introduce `helpers.go`, `common.go`, `utils.go`, or similar generic catch-alls for a small amount of shared code.
- Two or three callers, or "separation of concerns," do not justify a new file.
- Do not carry one-class-per-file or one-concern-per-file habits from Java or Python into Go. Keep closely related code physically close; do not trade locality for organizational purity.

Packages:

- Prefer adding cohesive functionality to an existing package over creating a narrowly scoped one. A healthy package holds multiple types, files, and responsibilities that naturally belong together (`net/http`, `os`, `flag`).
- A nameable concept or "this code does one thing" is not sufficient justification for a package.
- Decisive test: if callers will almost always need the new package together with its parent or neighbor, it should not be separate.
- Do not recursively decompose code into smaller packages. Create one only when it forms a meaningful standalone API or dependency boundary.
- Diagnostic signal: if ordinary implementation code routinely imports several sibling packages sharing a project-specific path prefix, those are probably not meaningful independent boundaries. Code normally used together should live together. Prefer a cohesive package with unexported implementation details over a constellation of tiny packages that callers must assemble.

## Abstraction

Every abstraction must earn its existence. Prefer concrete code until multiple real uses demonstrate the need.

- Do not introduce an abstraction merely because one could exist; stay concrete until there is an actual need to generalize.
- Premature interfaces are a well-known Go mistake — do not add an interface before there are multiple real implementations.
- Unlike Java, do not create an interface alongside an implementation by default. Define interfaces where they are consumed, and keep them as small as the consumer requires.

## Duplication

Go tolerates modest repetition when the alternative obscures control flow or introduces machinery. Favor simple, top-to-bottom-readable code over unnecessary abstraction.

- Do not deduplicate mechanically.
- Small amounts of obvious duplication are preferable to an abstraction that makes the code harder to follow.

## Functions over frameworks

Functions are cheap; frameworks are expensive. Where another ecosystem invents an object hierarchy or subsystem, Go uses a couple of plain functions. Write code, don't design types.

- Prefer functions and ordinary data structures over new architectural constructs.
- Do not reach for types, builders, managers, registries, or frameworks when a function will do.

## Control flow

Control flow should be boring and visible. Favor early returns, especially for errors, over deep nesting or elaborate control-flow abstractions. The normal path continues down the left edge.

- Prefer explicit, linear control flow.
- Handle exceptional conditions early and return.
- Do not hide simple branching behind abstractions or unnecessary `else` blocks.

## Errors are values

Errors are ordinary values, not an exception subsystem. Return an error, wrap it with useful context where appropriate, and handle it at the level that can act on it.

- Return, wrap, inspect, or handle errors deliberately.
- Do not invent exception-like infrastructure or custom error hierarchies around routine failures.

## Commit messages

Same ethos, applied to prose: state what changed and why, not every thought that led there.

- Concise subject; add a short body only for intent or non-obvious constraints. Skip narration the diff already shows — a large mechanical change may need barely a word.

## Above all: don't import other languages' architecture

Go prefers concrete code, explicit control flow, locality, small consumer-driven interfaces, and modest repetition over abstraction, indirection, decomposition, and machinery. When two implementations are equally correct, prefer the one with fewer concepts.

Fewer moving parts is not missing architecture — in Go it often *is* the architecture. Absence of machinery is not unfinished work; added machinery is usually the defect.

Resist these imported instincts:

- **Generics to remove duplication** — use them only when the problem itself is genuinely generic.
- **DI frameworks** — a function parameter, struct field, or tiny consumer-defined interface already is dependency injection. No containers, providers, factories, registries, or service locators.
- **Redesigning production code to enable mocking** — test real behavior; add seams only when they have genuine production value.
- **Tiny single-use helpers** — extract only when the result is a meaningful operation, not merely to shorten a function.
- **Named wrapper types** (`Config`, `Options`, `Manager`, `Processor`, `Result`) for tiny concepts — leave a string a string. Add a type when it provides behavior, semantics, safety, or API clarity.
- **Reflex accessors and constructors** — prefer useful zero values; add methods for behavior or a real API boundary, not to wrap fields (and it's `Owner()`, not `GetOwner()`).
- **Clever reuse** (map/filter/reduce pipelines, reflection, functional composition) where a straightforward loop reads directly.
