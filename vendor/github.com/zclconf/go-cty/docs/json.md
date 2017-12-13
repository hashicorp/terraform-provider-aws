# JSON serialization of `cty` values

[The `json` package](https://godoc.org/github.com/apparentlymart/go-cty/cty/json)
allows `cty` values to be serialized as JSON and decoded back into `cty` values.

Since the `cty` type system is a superset of the JSON type system, two modes
of operation are possible:

The recommended approach is to define the intended `cty` data structure as
a `cty.Type` -- possibly involving `cty.DynamicPseudoType` placeholders --
which then allows full recovery of the original values with correct type
information, assuming that the same type description can be provided at
decoding time.

Alternatively, this package can decode an arbitrary JSON data structure into
the corresponding `cty` types, which means that it is possible to serialize
a `cty` value without type information and then decode into a value that
contains the same data but possibly uses different types to represent
that data. This allows direct integration with the standard library
`encoding/json` package, at the expense of type-lossy deserialization.

## Type-preserving JSON Serialization

The `Marshal` and `Unmarshal` functions together provide for type-preserving
serialization and deserialization (respectively) of `cty` values.

The pattern for using these functions is to define the intended `cty` type
as a `cty.Type` instance and then pass an identical type as the second argument
to both `Marshal` and `Unmarshal`. Assuming an identical type is used for both
functions, it is guaranteed that values will round-trip through JSON
serialization to produce a value of the same type.

The `cty.Type` passed to `Unmarshal` is used as a hint to resolve ambiguities
in the mapping to JSON. For example, `cty` list, set and tuple types all
lower to JSON arrays, so additional type information is needed to decide
which type to use when unmarshaling.

The `cty.Type` passed to `Marshal` serves a more subtle purpose. Any
`cty.DynamicPseudoType` placeholders in the type will cause extra type
information to be saved in the JSON data structure, which is then used
by `Unmarshal` to recover the original type.

Type-preserving JSON serialization is able to serialize and deserialize
capsule-typed values whose encapsulated Go types are JSON-serializable, except
when those values are conformed to a `cty.DynamicPseudoType`. However, since
capsule values compare by pointer equality, a decoded value will not be
equal (as `cty` defines it) with the value that produced it.

## Type-lossy JSON Serialization

If a given application does not need to exactly preserve the type of a value,
the `SimpleJSONValue` type provides a simpler method for JSON serialization
that works with the `encoding/json` package in Go's standard library.

`SimpleJSONValue` is a wrapper struct around `cty.Value`, which can be
embedded into another struct used with the standard library `Marshal` and
`Unmarshal` functions:

```go
type Example struct {
    Name  string          `json:"name"`
    Value SimpleJSONValue `json:"value"`
}

var example Example
example.Name = "Ermintrude"
example.Value = SimpleJSONValue{cty.NumberIntVal(43)}
```

Since no specific `cty` type is available when unmarshalling into
`SimpleJSONValue`, a straightforward mapping is used:

* JSON strings become `cty.String` values.
* JSON numbers become `cty.Number` values.
* JSON booleans become `cty.Bool` values.
* JSON arrays become `cty.Tuple`-typed values whose element types are selected via this mapping.
* JSON objects become `cty.Object`-typed values whose attribute types are selected via this mapping.
* Any JSON `null` is mapped to `cty.NullVal(cty.DynamicPseudoType)`.

The above mapping is unambiguous and lossless, so any valid JSON buffer can be
decoded into an equally-expressive `cty` value, but the type may not exactly
match that of the value used to produce the JSON buffer in the first place.
