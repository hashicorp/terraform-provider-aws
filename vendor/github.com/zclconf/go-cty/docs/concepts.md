# `cty` Concepts

`cty` is built around two fundamental concepts: _types_ and _values_.

A _type_ represents a particular way of storing some data in memory and a
set of operations that can be performed on that data. A _value_ is, therefore,
the combination of some raw data and a _type_ that describes how that data
should be interpreted.

The simplest types in `cty` are the _number_, _string_ and _bool_ types,
collectively known as the _primitive_ types.

Along with the primitive types, `cty` supports _compound_ types, which are
types that are constructed by assembling together other types in a particular
way. The _compound_ types are further subdivided into two categories:

* **Collection Types** represent collections of values that all have the same
  type (the _element type_) and permit access to those values in different
  ways. The collection type kinds are _list_, _set_, and _map_.

* **Structural Types** represent collections of values that may all have
  _different_ types, organized either by name or by position in a sequence.
  The structural type kinds are _object_ and _tuple_.

For example, "list of string" is a collection type that represents a
collection of string values (_elements_) that are each assigned a sequential
index starting at zero, while "map of string" instead assigns each of its
elements a name in the form of a string value.

The details of the specific types and type kinds are covered in
[the full description of the type system](./types.md); the remainder of _this_
document will discuss types and values in general, using specific types only
as examples.

## Value Operations

Each type defines a set of operations that are valid on its values. For
example, the _number_ type permits various arithmetic operations such as
addition, subtraction, and multiplication, but these are not permitted for
other types such as _bool_.

Since `cty` is a dynamic type system (from the perspective of the calling Go
program), the validity of an operation on a given value must be checked at
runtime. The documentation for each type defines what operations are valid
on it and what semantics each operation has.

## Unknown Values and the Dynamic Pseudo-Type

`cty` has some additional _optional_ concepts that may be useful in certain
applications.

An _unknown value_ is a value that carries a type but no value. It can serve
as a placeholder for a value to be resolved later, which can be useful when
implementing a static type checker for a language. Unknown values are special
because they support the same operations as a known value of the same type
but the result will itself be an unknown value. For example, the number 5
added to an unknown number yields another unknown number.

The dynamic pseudo-type is a special type that serves as a placeholder for
a type that isn't yet known. Whereas unknown values represent situations where
the type is known and the value is not, the dynamic pseudo-type represents
situations where neither is known, or where any value of any type is permitted.
It is referred to as a "pseudo-type" because while it can be used in many
places where types are permitted, it does not define any operations of its own.

These two concepts are related in that the dynamic pseudo-type has no non-null,
non-unknown values. It single non-null type is itself an unknown value.
_All_ operations are supported on non-null dynamic values, but the result
will always be an unknown value, possibly type-unknown itself.

Dealing with unknown values and the dynamic pseudo-type can cause additional
complexity for a calling application, although many details of it are handled
automatically by the `cty` internals. As a consequence, the main `cty` API
promises to never produce an unknown value for an operation unless one of the
operands is itself unknown, and so applications can opt out of this additional
complexity by never providing unknown values as operands.

## Type Equality and Type Conformance

Two types are said to be equal if they are exactly equivalent. Each type kind
defines its own equality rules, but the overall intent is to implement strict
type comparisons.

Type _conformance_ is a slightly-weaker concept that allows the dynamic
pseudo-type to be used as a placeholder to represent "any type". Therefore
a given type is equal only to itself but it is _conformant_ to either itself
or the dynamic pseudo-type.

Type conformance is not directly used by `cty`'s core, but it is used as
a building block for the `function` package and for JSON serialization.

## The `cty` Go API

The primary way a application works with `cty` values is via the API exposed
by the `cty` go package. The full details of this package are in
[its reference documentation](https://godoc.org/github.com/apparentlymart/go-cty/cty),
so this section will just cover the basic usage patterns.

The main features of the `cty` package are the Go types `cty.Type` and `cty.Value`,
which each represent the concept they are named after.

The package contains variables that represent the primitive types, `cty.Number`,
`cty.String` and `cty.Bool`. It also contains functions that allow the
construction of compound types, such as `cty.List`, `cty.Object`, etc. These
functions each take different arguments depending on the kind of compound type
in question.

Alongside the types and type factories, the package also contains variables
and functions for constructing _values_ of these types, which conventionally
have names that are the corresponding type or type kind with the suffix `Val`.
For example, the two boolean values are exported as `cty.True` and `cty.False`,
and string values can be constructed using the function `cty.StringVal`, given
a native Go string.

The `cty.Type` and `cty.Value` types are similar to the types of the same
name in the built-in Go `reflect` package. They expose methods that are the
union of all operations supported across all types, but each method has a
set of constraints associated with it, and failure to follow these will result
in a run-time panic.

The `cty.Value` object has two classes of methods:

* **Operation Methods** stay within the `cty` type system, dealing entirely
  with `cty.Value` instances. These methods fully deal with concerns such as
  unknown values, so the caller just needs to be sure to apply only operations
  that are valid for the receiving value's type.

* **Integration Methods** live on the boundary between `cty` and the native
  Go type system, and can be used by the calling application to integrate
  with non-`cty`-aware code. These methods often have constraints such as not
  supporting unknown values, which are covered in their documentation.

While the integration methods alone are sufficient for a calling application
to convert to and from `cty` values, the utility package
[`gocty`](./gocty.html) provides a more convenient way to convert between
Go native values and `cty` values.


