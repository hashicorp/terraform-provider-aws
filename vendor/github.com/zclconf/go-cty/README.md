# cty

`cty` (pronounced "see-tie") is a dynamic type system for applications written
in Go that need to represent user-supplied values without losing type
information. The primary intended use is for implementing configuration
languages, but other uses may be possible too.

One could think of `cty` as being the reflection API for a language that
doesn't exist, or that doesn't exist _yet_. It provides a set of value types
and an API for working with values of that type.

Fundamentally what `cty` provides is equivalent to an `interface{}` with some
dynamic type information attached, but `cty` encapsulates this to ensure that
invariants are preserved and to provide a more convenient API.

As well as primitive types, basic collection types (lists, maps and sets) and
structural types (object, tuple), the `cty` type and value system has some
additional, optional features that may be useful to certain applications:

* Representation of "unknown" values, which serve as a typed placeholder for
  a value that has yet to be determined. This can be a useful building-block
  for a type checker. Unknown values support all of the same operations as
  known values of their type, but the result will often itself be unknown.

* Representation of values whose _types_ aren't even known yet. This can
  represent, for example, the result of a JSON-decoding function before the
  JSON data is known.

Along with the type system itself, a number of utility packages are provided
that build on the basics to help integrate `cty` into calling applications.
For example, `cty` values can be automatically converted to other types,
converted to and from native Go data structures, or serialized as JSON.

For more details, see the following documentation:

* [Concepts](./docs/concepts.md)
* [Full Description of the `cty` Types](./docs/types.md)
* [API Reference](https://godoc.org/github.com/apparentlymart/go-cty/cty) (godoc)
* [Conversion between `cty` types](./docs/convert.md)
* [Conversion to and from native Go values](./docs/gocty.md)
* [JSON serialization](./docs/json.md)
* [`cty` Functions system](./docs/functions.md)
* [diff and patch for `cty` values](./docs/diff.md)

---

## License

Copyright 2017 Martin Atkins

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
