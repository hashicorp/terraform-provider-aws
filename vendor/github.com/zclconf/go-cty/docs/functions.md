# `cty` Functions system

Core `cty` is primarily concerned with types and values, with behavior
delegated to the calling application. However, writing functions that operate
on `cty.Value` is expected to be a common enough case for it to be worth
factoring out into a shared package, so
[the `function` package](https://godoc.org/github.com/apparentlymart/go-cty/cty/function)
serves that need.

The shared function abstraction is intended to help applications provide the
expected behavior in handling `cty` complexities such as unknown values and
dynamic types. The infrastructure in this package can do basic type checking
automatically, allowing applications to focus on the logic unique to each
function.

## Function Specifications

Functions are defined by calling applications via `FunctionSpec` instances.
These describe the parameters the function accepts, the return value it would
produce given a set of parameters, and the actual implementation of the
function.

The return type is defined by a function, allowing the definition of generic
functions whose return type depends on the given argument types or values.

### Function Parameters

Functions can have both fixed parameters and variadic arguments. Each fixed
parameter has its own separate specification, while the variadic arguments
together share a single parameter specification, meaning that they must
all be of the same type.

[`Parameter`](https://godoc.org/github.com/apparentlymart/go-cty/cty/function#Parameter)
represents the description of a parameter. The `Params` member of
`FunctionSpec` is a slice of positional parameters, while `VarParam` is
a pointer to the description of the variadic arguments, if supported.

Parameters have the following fields:

* `Name` is not used directly by this package but is intended to be useful
  in generating documentation based on function specifications.
* `Type` is a type specification that a given argument must _conform_ to.
  (see the `TestConformance` method of `cty.Type` for information on
  what exactly that means.)
* `AllowNull` can be set to `true` to permit the caller to provide null values.
  If not set, passing a null is treated as an immediate error and the
  implementation function is not called at all.
* `AllowUnknown` can be set to `true` if the implementation function is
  prepared to handle unknown values. If not set, calls with an unknown argument
  will immediately return an unknown value of the function's return type,
  and the implementation function is not called at all.
* `AllowDynamicType` can be set to `true` to allow not-yet-typed values to be
  passed. If not set, calls with a dynamic argument will immediately return
  `cty.DynamicVal`, and neither the type-checking function nor the
  implementation function will be called.
  
Since dynamic values are themselves unknown, `AllowUnknown` and
`AllowDynamicType` must be set together to permit `cty.DynamicVal` to be
passed as an argument to the implementation function, but setting
`AllowDynamicType` _without_ setting `AllowUnknown` has the special effect
of allowing dynamic values to be passed into the type checking function
_without_ also passing them to the implementation function, allowing a more
specific return type to be specified even if the input type isn't
known.

### Return Type

A function returns a single value when called. The return type function,
specified via the `Type` field in `FunctionSpec`, defines the type this
value will have for the given arguments.

The arguments are passed to the type function as _values_ rather than as
types, though in many cases they will be unknown values for which the only
useful operation is to call the `Type()` method on them. Unknown values
can be passed to the type function regardless of how the `AllowUnknown`
flag is set on the associated parameter specification.

If `AllowDynamicType` is set on a parameter specification, a corresponding
argument may be `cty.DynamicVal`. The return type function can then handle
this how it wishes. If the parameter _itself_ is typed as
`cty.DynamicPseudoType` then the corresponding argument may be a value of
_any_ type. These behaviors together allow the return type function to behave
as a full-fledged _type checking_ function, returning an error if the caller's
supplied types do not conform to some requirements that are not simple enough
to be expressed via the parameter specifications alone.

Returning `cty.DynamicPseudoType` from the type checking function signals that
the function is not able to determine its return type from the given
information. Hopefully -- but not necessarily -- the function _implementation_
will produce a value of a known type once the argument values are themselves
known.

Calling applications may elect to pass _known_ values for type checking, which
then allows for functions whose return type depends on argument _values_.
This is a relatively-rare situation, but one key example is a hypothetical
JSON decoding function, which takes a string value for the JSON structure to
decode. If given `cty.Unknown(cty.String)` as an argument, this function would
need to specify its return type as `cty.DynamicPseudoType`, but if given
a _known_ string it could infer an appropriate return type from that string.

### Function Implementation

The `Impl` field in `FunctionSpec` is used to specify the function's
implementation as a Go function pointer.

The implementation function takes a slice of `cty.Value` representing the
call arguments and the `cty.Type` that was returned from the return type
function. It must then either produce a value conforming to that given type
or return an error.

A function implementer can write any arbitrary Go code into the implementation
of a function, but `cty` functions are intended to behave as pure functions,
so side-effects should be avoided unless the function is specialized for a
particular calling application that is able to accept such side-effects.

If any of the given arguments are unknown and their corresponding parameter
specifications _permit_ unknowns, the function implementation must handle
this situation, normally by immediately returning an unknown value of the
required return type. A function should _not_ return unknown values unless
at least one of the arguments is unknown, since to do otherwise would
violate the `cty` guarantee that a caller can avoid dealing with the
complexity of unknown values by never passing any in.

## The `cty` Standard Library

The set of operations provided directly on `cty.Value` is intended to cover
the basic operators of a simple expression language, but there are several
higher-level operations that can be implemented in terms of `cty` values,
such as string manipulations, standard mathematical functions, etc.

[The standard library](https://godoc.org/github.com/apparentlymart/go-cty/cty/function/stdlib)
contains a set of `cty` functions that are intended to be generally useful.
For the convenience of calling applications, each function is provided both
as a first-class Go function _and_ as a `Function` instance; the former
could be useful for Go code dealing directly with `cty.Value` instances,
while the latter is likely more useful for exposing functions into a
language interpreter.

The standard library also includes some functions that are just thin wrappers
around the operations on `cty.Value`. These are somewhat redundant, but
exposing them as functions has the advantage that their operands can be
described as function parameters and so automatic type checking and error
handling is possible, whereas the `cty.Value` operations prefer to `panic`
when given invalid input.

