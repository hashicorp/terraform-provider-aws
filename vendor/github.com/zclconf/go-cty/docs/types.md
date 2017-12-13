# `cty` types

This page covers in detail all of the primitive types and compound type kinds supported within `cty`. For more general background information, see
[the `cty` overview](./concepts.md).

## Common Operations and Integration Methods

The following methods apply to all values:

* `Type` returns the type of the value, as a `cty.Type` instance.
* `Equals` returns a `cty.Bool` that is `cty.True` if the receiver and the
  given other value are equal. This is an operation method, so its result
  will be unknown if either argument is unknown.
* `RawEquals` is similar to `Equals` except that it doesn't implement the
  usual special behavior for unknowns and dynamic values, and it returns a
  native Go `bool` as its result. This method is intended for use in tests;
  `Equals` should be preferred for most uses.
* `IsKnown` returns `true` if the receiver is a known value, or `false`
  if it is an unknown value.
* `IsNull` returns `true` if the receiver is a null value, or `false`
  otherwise.

All values except capsule-typed values can be seralized with the builtin
Go package `encoding/gob`. Values can also be used with the `%#v` pattern
in the `fmt` package to print out a Go-oriented serialization of the
value.

## Primitive Types

### `cty.Number`

The number type represents arbitrary-precision floating point numbers.

Since numbers are arbitrary-precision, there is no need to worry about
integer overflow/underflow or loss of precision during arithmetic operations.
However, eventually a calling application will probably want to convert a
number to one of the Go numeric types, at which point its range will be
constrained to fit within that type, generating an error if it does not fit.

The following additional operations are supported on numbers:

* `Absolute` converts a negative value to a positive value of the same
  magnitude.
* `Add` computes the sum of two numbers.
* `Divide` divides the receiver by another number.
* `GreaterThan` returns `cty.True` if the receiver is greater than the other
  given number.
* `GreaterThanOrEqualTo` returns `cty.True` if the receiver is greater than or
  equal to the other given number.
* `LessThan` returns `cty.True` if the receiver is less than the other given
  number.
* `LessThanOrEqualTo` returns `cty.True` if the receiver is less than or
  equal to the other given number.
* `Modulo` computes the remainder from integer division of the receiver by
  the other given number.
* `Multiply` computes the product of two numbers.
* `Negate` inverts the sign of the number.
* `Subtract` computes the difference between two numbers.

`cty.Number` values can be constructed using several different factory
functions:

* `NumberVal` creates a number value from a `*big.Float`, from the `math/big` package.
* `NumberIntVal` creates a number value from a native `int64` value.
* `NumberUIntVal` creates a number value from a native `uint64` value.
* `NumberFloatVal` creates a number value from a native `float64` value.

The core API only allows extracting the value from a known number as a
`*big.Float` using the `AsBigFloat` method. However,
[the `gocty` package](./gocty.md) provides a more convenient way to convert
numbers to any native Go number type, with automatic range checking to ensure
that the value fits into the target type.

The following numbers are provided as package variables for convenience:

* `cty.Zero` is the number zero.
* `cty.PositiveInfinity` represents positive infinity as a number. All other
  numbers are less than this value.
* `cty.NegativeInfinity` represents negative infinity as a number. All other
  numbers are greater than this value.

### `cty.String`

The string type represents a sequence of unicode codepoints.

There are no additional operations supported for strings.

`cty.String` values can be constructed using the `cty.StringVal` factory
function. The native Go `string` passed in must be a valid UTF-8 sequence,
and it will be normalized such that any combining diacritics are converted
to precomposed forms where available. (Technically-speaking, the mapping
applied is the NFC normalization as defined in the relevant Unicode
specifications.)

The `AsString` method can be called on a string value to obtain the native
Go `string` representation of a known string, after normalization.

### `cty.Bool`

The bool type represents boolean (true of false) values.

The following additional operations are supported on bool values:

* `And` computes the logical AND operation for two boolean values.
* `Not` returns the boolean opposite of the receiver.
* `Or` computes the ligical OR operation for two boolean values.

Calling applications may either work directly with the predefined `cty.True`
and `cty.False` variables, or dynamically create a boolean value using
`cty.BoolVal`.

The `True` method returns a native Go `bool` representing a known boolean
value. The `False` method returns the opposite of it.

## Collection Type Kinds

`cty` has three different kinds of collection type. All three of them are
parameterized with a single _element type_ to produce a collection type.
The difference between the kinds is how the elements are internally organized
and what operations are used to retrieve them.

### `cty.List` types

List types are ordered sequences of values, accessed using consecutive
integers starting at zero.

The following operations apply to values of a list type:

* `Index` can be passed an integer number less than the list's length to
  retrieve one of the list elements.
* `HasIndex` can be used to determine if a particular call to `Index` would
  succeed.
* `Length` returns a number representing the number of elements in the list.
  The highest integer that can be passed to `Index` is one less than this
  number.

List types are created by passing an element type to the function `cty.List`.
List _values_ can be created by passing a type-homogenous `[]cty.Value`
to `cty.ListVal`, or by passing an element type to `cty.ListValEmpty`.

The following integration methods can be used with known list-typed values:

* `LengthInt` returns the length of the list as a native Go `int`.
* `ElementIterator` returns an object that can be used to iterate over the
  list elements.
* `ForEachElement` runs a given callback function for each element.

### `cty.Map` types

Map types are collection values that are each assigned a unique string key.

The following operations apply to values of a map type:

* `Index` can be passed a string value to retrieve the corresponding
  element.
* `HasIndex` can be used to determine if a particular call to `Index` would
  succeed.
* `Length` returns a number representing the number of elements in the map.

Map types are created by passing an element type to the function `cty.Map`.
Map _values_ can be created by passing a type-homogenous `map[string]cty.Value`
to `cty.MapVal`, or by passing an element type to `cty.MapValEmpty`.

The following integration methods can be used with known map-typed values:

* `LengthInt` returns the number of elements as a native Go `int`.
* `ElementIterator` returns an object that can be used to iterate over the
  map elements in lexicographical order by key.
* `ForEachElement` runs a given callback function for each element in the
  same order as the `ElementIterator`.

### `cty.Set` types

Set types are collection values that model a mathematical set, where every
possible value is either in or out of the set. Thus each set element value is
its own identity in the set, and a given value cannot appear twice in the same
set.

The following operations apply to values of a set type:

* `HasIndex` can be used to determine whether a particular value is in the
  receiving set..
* `Length` returns a number representing the number of elements in the set.
* `Index` is not particularly useful for sets, but for symmetry with the
  other collection types it may be passed a value that is in the set and it
  will then return that same value.

Set types are created by passing an element type to the function `cty.Set`.
Set _values_ can be created by passing a type-homogenous `[]cty.Value` to
`cty.SetVal`, though the result is undefined if two values in the slice are
equal. Alternatively, an empty set can be constructed using `cty.SetValEmpty`.

The following integration methods can be used with known set-typed values:

* `LengthInt` returns the number of elements as a native Go `int`.
* `ElementIterator` returns an object that can be used to iterate over the
  set elements in an undefined (but consistent) order.
* `ForEachElement` runs a given callback function for each element in the
  same order as the `ElementIterator`.

Set membership is determined by equality, which has an interesting consequence
for unknown values. Since unknown values are never equal to one another,
theoretically an infinite number of unknown values can be in a set (constrained
by available memory) but can never be detected by calls to `HasIndex`. However,
they _can_ be seen in the set's length and by iterating over its members.

## Structural Types

`cty` has two different kinds of structural type. They have in common that
they combine a number of values of arbitrary types together into a single
value, but differ in how those values are internally organized and in which
operations are used to retreive them.

### `cty.Object` types

Object types each have zero or more named attributes that each in turn have
their own type.

The following operation applies to values of an object type:

* `GetAttr` returns the value of an attribute given its name.

The set of valid attributes for an object type can be inspected using the
following methods on the type itself:

* `AttributeTypes` returns a `map[string]Type` describing the types of all of
  the attributes.
* `AttributeType` returns the type of a single attribute given its name.
* `HasAttribute` returns `true` if the type has an attribute with the given
  name.

Object types are constructed by passing a `map[string]Type` to `cty.Object`.
Object _values_ can be created by passing a `map[string]Value` to
`cty.ObjectVal`, in which the keys and value types define the object type
that is implicitly created for that value.

The variable `cty.EmptyObject` contains the object type with no attributes,
and `cty.EmptyObjectVal` is the only non-null, known value of that type.

### `cty.Tuple` types

Tuple types each have zero or more elements, each with its own type, arranged
in a sequence and accessed by integer numbers starting at zero.

A tuple type is therefore somewhat similar to a list type, but rather than
representing an arbitrary number of values of a single type it represents a
fixed number of values that may have _different_ types.

The following operations apply to values of a tuple type:

* `Index` can be passed an integer number less than the tuple's length to
  retrieve one of the tuple elements.
* `HasIndex` can be used to determine if a particular call to `Index` would
  succeed.
* `Length` returns a number representing the number of elements in the tuple.
  The highest integer that can be passed to `Index` is one less than this
  number.

Tuple types are created by passing a `[]cty.Type` to the function `cty.Tuple`.
Tuple _values_ can be created by passing a `[]cty.Value` to `cty.TupleVal`,
in which the value types define the tuple type that is implicitly created
for that value.

The variable `cty.EmptyTuple` contains the tuple type with no elements,
and `cty.EmptyTupleVal` is the only non-null, known value of that type.

The following integration methods can be used with known tuple-typed values:

* `LengthInt` returns the length of the tuple as a native Go `int`.
* `ElementIterator` returns an object that can be used to iterate over the
  tuple elements.
* `ForEachElement` runs a given callback function for each element.

## The Dynamic Pseudo-Type

The dynamic pseudo-type is not a real type but is rather a _placeholder_ for
a type that isn't known.

One consequence of this being a "pseudo-type" is that there is no known,
non-null value of this type, but `cty.DynamicVal` is the unknown value of
this type, and a null value without a known type can be represented by
`cty.NullVal(cty.DynamicPseudoType)`.

This pseudo-type serves two similar purposes as a placeholder type:

* When `cty.DynamicVal` is used in an operation with another value, the
  result is either itself `cty.DynamicVal` or it is an unknown value of
  some suitable type. This allows the dynamic pseudo-type to be used as
  a placeholder during type checking, optimistically assuming that the
  eventually-determined type will be compatible and failing at that
  later point if not.
* `cty.DynamicPseudoType` can be used with the type `TestConformance` method
  to declare that any type is permitted in the type being tested for
  conformance.

`cty` doesn't have _sum types_ (i.e. union types), so `cty.DynamicPseudoType`
can be used also to represent situations where two or more specific types are
allowed, under the assumption that more specific type checking will be done
within the calling application's own logic even though it cannot be expressed
directly within the `cty` type system.

## Capsule Types

Capsule types are a special kind of type that allows a calling application to
"smuggle" otherwise-unsupported Go values through the `cty` type system.

Such types and their associated values have no defined meaning in `cty`. The
interpreter for a language building on `cty` might use capsule types for
passing language-specific objects between functions provided in that language.

A capsule type is created using the function `cty.Capsule`, which takes a
"friendly name" for the type along with a `reflect.Type` that defines what
type of Go value will be encapsulated in values of this type. A capsule-typed
value can then be created by passing the capsule type and a pointer to a
native value of the encapsulated type to `cty.CapsuleVal`.

The integration method `EncapsulatedValue` allows the encapsulated data to
then later be retrieved.

Capsule types compare by reference, so each call to `cty.Capsule` produces
a distinct type. Capsule _values_ compare by pointer equality, so two
capsule values are equal if they have the same capsule type and they
encapsulate a pointer to the same object.

Due to the strange nature of capsule types, they are not totally supported
by all of the other packages that build on the core `cty` API. They should
be used with care and the documentation for other packages should be consulted
for information about caveats and constraints relating to their use.

