# Converting between `cty` types

[The `convert` package](https://godoc.org/github.com/apparentlymart/go-cty/cty/convert)
provides a standard set of type conversion routines for moving between
types in the `cty` type system.

_Conversion_ in this context means taking a given value and producing a
new value of a different type that, in some sense, contains the same
information. For example, the number `5` can be converted to a string as
`"5"`.

Specific conversion operations are represented by type `Conversion`, which
is a function type that takes a single value as input and returns a value
or an error.

## "Safe" and "Unsafe" conversions

The `convert` package broadly organizes its supported conversions into two
types.

"Safe" conversions are ones where all values of the source type can
be represented in the target type, and thus the conversion is guaranteed to
succeed for any value of the source type.

"Unsafe" conversions, on the other hand, are able to convert only a subset
of values of the source type. Values outside of that subset will cause the
conversion function to return an error.

Converting from number to string is safe because an unambiguous string
representation can be created for any number. The converse is _unsafe_,
because while a string like `"2.5"` can be converted to a number, a string
like `"bananas"` cannot.

The calling application must choose whether to attempt unsafe conversions,
depending on whether it is willing to tolerate conversions returning errors
even though they ostensibly passed type checking. Operations that have both
safe and unsafe modes come in couplets, with the unsafe version's name
having the suffix `Unsafe`.

## Getting a Conversion

To find out if a conversion is available between two types, an application can
call either `GetConversion` or `GetConversionUnsafe`. These functions return
a valid `Conversion` if one is available, or `nil` if not.

Note that there are no conversions from a type to itself. Callers should check
if two types are equal before attempting to obtain a conversion between them.

As usual, `cty.DynamicPseudoType` serves as a special-case placeholder. It is
used in two ways, depending on whether it appears in the source or the
destination type:

* When a source type is dynamic, a special unsafe conversion is available that
  takes any value and passes it through verbatim if it matches the destination
  type, or returns an error if it does not. This can be used as part of handling
  dynamic values during a type-checking procedure, with the generated
  conversion serving as a run-time type check.

* When a _destination_ type is dynamic, a simple passthrough conversion is
  generated that does not transform the source value at all. This is supported
  so that a destination type can behave similarly to a type description used
  for a conformance check, thus allowing this package to be used to attempt
  to _make_ a type conformant, rather than merely check whether it already
  is.

## Converting a Value

A value can be converted by passing it as the argument to any conversion whose
source type matches the value's type. If the conversion is an unsafe one, the
conversion function may return an error, in which case the returned value is
invalid and must not be used.

As a convenience, the `Convert` function takes a value and a target type and
returns a converted value if a conversion is available. This is equivalent
to testing for an unsafe conversion for the value's type and then immediately
calling any discovered conversion. An error is returned if a conversion is not
available.

## Type Unification

A related idea to type _conversion_ is type _unification_. While conversion
is concerned with going from a specific source type to a specific target type,
unification is instead concerned with finding a single type that several other
types can be converted to, without any specific preference as to what the
final type is.

A good example of this would be to take a set of values provided to initialize
a list and choose a single type that all of those values can be
converted to, which then decides the element type of the final list.

The `Unify` and `UnifyUnsafe` functions are used for type unification. They
both take as input a slice of types and then return, if possible, a single
target type along with a slice of conversions corresponding to each
of the input types.

Since many type pairs support type conversions in both directions, the unify
functions must apply a preference for which direction to follow given such a
pair of types. These functions prefer safe conversions over unsafe ones
(assuming that `UnifyUnsafe` was called), and prefer lossless conversions
over potentially-lossy ones.

Type unification is a potentially-expensive operation, depending on the
complexity of the passed types and whether they are mutually conformant.

## Conversion Charts

The foundation of the available conversions is the matrix of conversions
between the primitive types. String is the most general type, since the
other two primitive types have safe conversions to string. The full
matrix for primitive types is as follows:

|         | string | number | boolean |
|---------|:------:|:------:|:-------:|
| string  |   n/a  | unsafe |  unsafe |
| number  |  safe  |   n/a  |   none  |
| boolean |  safe  |  none  |   n/a   |

The conversions for compound types are then derived from the above foundation.
For example, a list of numbers can convert to a list of strings
because a number can convert to a string.

The compound type kinds themselves have some available conversions, though:

|        |  tuple | object | list |   map  |     set    |
|--------|:------:|:------:|:----:|:------:|:----------:|
| tuple  |   n/a  |  none  | safe |  none  | safe+lossy |
| object |  none  |   n/a  | none |  safe  |    none    |
| list   | unsafe |  none  |  n/a |  none  | safe+lossy |
| map    |  none  | unsafe | none |   n/a  |    none    |
| set    | unsafe |  none  | safe |  none  |     n/a    |

Conversions between compound kinds, as shown above, are possible only
if their respective elements/attributes also have conversions available.

The conversions from structural types to collection types rely on
type unification to identify a single element type for the final collection,
and so conversion is possible only if unification is possible.
