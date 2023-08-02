# µjson

[![GoDoc](http://img.shields.io/badge/go-documentation-blue.svg?style=flat-square)](https://pkg.go.dev/github.com/olvrng/ujson)
[![License](http://img.shields.io/badge/license-mit-blue.svg?style=flat-square)](https://raw.githubusercontent.com/olvrng/ujson/master/LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/olvrng/ujson?style=flat-square)](https://goreportcard.com/report/github.com/olvrng/ujson)
[![Build Status](http://img.shields.io/travis/olvrng/ujson.svg?style=flat-square)](https://travis-ci.org/olvrng/ujson)
[![Github code coverage](https://img.shields.io/badge/code%20coverage-97%25-brightgreen?style=flat-square)](https://gocover.io/github.com/olvrng/ujson)

A fast and minimal JSON parser and transformer that works on unstructured JSON.
It works by parsing input and calling the given callback function when
encountering each item.

Read more on [the introduction article](https://hackernoon.com/json-a-minimal-json-parser-and-transformer-in-go-8eo34lp).

## Motivation

Sometimes we just want to make some minimal changes to a JSON document, or do
some generic transformations without fully unmarshalling it. For example,
removing [blacklist fields](https://pkg.go.dev/github.com/olvrng/ujson#example-Walk-RemoveBlacklistFields2)
from response JSON. Why spend all the cost on unmarshalling into a `map[string]interface{}`
just to immediate marshal it again.

The following code is taken from [StackOverflow](https://stackoverflow.com/questions/35441254/making-minimal-modification-to-json-data-without-a-structure-in-golang?ref=hackernoon.com):

```json
{
  "responseHeader": {
    "status": 0,
    "QTime": 0,
    "params": {
      "q": "solo",
      "wt": "json"
    }
  },
  "response": {
    "numFound": 2,
    "start": 0,
    "docs": [
      { "name": "foo" },
      { "name": "bar" }
    ]
  }
}
```

With µjson, we can quickly write [a simple
transformation](https://pkg.go.dev/github.com/olvrng/ujson#example-Walk-RemoveBlacklistFields2)
to remove "responseHeader" completely from all responses, once and forever:

```go
func removeBlacklistFields(input []byte) []byte {
    blacklistFields := [][]byte{
		[]byte(`"responseHeader"`), // note the quotes
	}
	b := make([]byte, 0, 1024)
	err := ujson.Walk(input, func(_ int, key, value []byte) bool {
		if len(key) != 0 {
			for _, blacklist := range blacklistFields {
				if bytes.Equal(key, blacklist) {
					// remove the key and value from the output
					return false
				}
			}
		}
		// write to output
		if len(b) != 0 && ujson.ShouldAddComma(value, b[len(b)-1]) {
			b = append(b, ',')
		}
		if len(key) > 0 {
			b = append(b, key...)
			b = append(b, ':')
		}
		b = append(b, value...)
		return true
	})
	if err != nil {
		panic(err)
	}
	return b
 }
```

The original scenario that leads me to write the package is because of
***int64***. When working in Go and PostgreSQL, I use ***int64*** (instead of
***string***) for ***ids*** because it’s more effective and has enormous space
for randomly generated ids. It’s not as big as UUID, 128 bits, but still big
enough for production use. In PostgreSQL, those ids can be stored as
***bigint*** and being effectively indexed. But for JavaScript, it can only
process integer up to 53 bits (JavaScript has BigInt but that’s a different
story, and using it will make things even more complicated).

So we need to wrap those int64s into strings before sending them to JavaScript.
In Go and PostgreSQL, the JSON is `{"order_id": 12345678}` but JavaScript will
see it as `{"order_id": "12345678"}` (note that the value is quoted). In Go, we
can define a custom type and implement the
[`json.Marshaler`](https://golang.org/pkg/encoding/json/#Marshaler) interface. But
in PostgreSQL, that’s just not possible or too complicated. I wrote a service
that receives JSON from PostgreSQL and converts it to be consumable by
JavaScript. The service also removes some blacklisted keys or does some other
transformations (for example, change `orderId` to `order_id`).

### Example use cases:

1. Walk through unstructured JSON:
   - [Print all keys and values](https://pkg.go.dev/github.com/olvrng/ujson#example-Walk)
   - Extract some values
    
2. Transform unstructured JSON:
   - [Remove all spaces](https://pkg.go.dev/github.com/olvrng/ujson#example-Walk-Reconstruct)
   - [Reformat](https://pkg.go.dev/github.com/olvrng/ujson#example-Walk-Reformat)
   - [Remove blacklist fields](https://pkg.go.dev/github.com/olvrng/ujson#example-Walk-RemoveBlacklistFields2)
   - [Wrap int64 in string for processing by JavaScript](https://pkg.go.dev/github.com/olvrng/ujson#example-Walk-WrapInt64InString)

without fully unmarshalling it into a `map[string]interface{}`.

See usage and examples on
[pkg.go.dev](https://pkg.go.dev/github.com/olvrng/ujson).

**Important**: *Behaviour is undefined on invalid JSON. Use on trusted input
only. For untrusted input, you may want to run it through
[`json.Valid()`](https://golang.org/pkg/encoding/json/#Valid) first.*

## Usage

The single most important function is [`Walk(input,
callback)`](https://pkg.go.dev/github.com/olvrng/ujson#Walk), which parses the
`input` JSON and call `callback` function for each key/value pair processed.

The callback function is called when an object key/value or an array key is
encountered. It receives 3 params in order: `level`, `key` and `value`.

- `level` is the indentation level of the JSON, if you format it properly. It
  starts from 0. It increases after entering an object or array and decreases
  after leaving.

- `key` is the raw key of the current object or empty otherwise. It can be a
  double-quoted string or empty.

- `value` is the raw value of the current item or a bracket. It can be a string,
  number, boolean, null, or one of the following brackets: `{ } [ ]`.

It’s important to note that `key` and `value` are provided as raw. Strings are
always double-quoted. It’s there for keeping the library fast and ignoring
unnecessary operations. For example, when you only want to reformat the output
JSON properly; you don’t want to unquote those strings and then immediately
quote them again; you just need to output them unmodified. And there are
[`ujson.Unquote()`](https://pkg.go.dev/github.com/olvrng/ujson#Unquote) and
[`ujson.AppendQuote()`](https://pkg.go.dev/github.com/olvrng/ujson#AppendQuote)
when you need to get the original strings.

For valid JSON, values will never be empty. We can test the first byte of value
(`value[0]`) to get its type:

- `n`: Null (`null`)
- `f`, `t`: Boolean (`false`, `true`)
- `0`-`9`, `-`: Number
- `"`: String, see [`Unquote()`](https://pkg.go.dev/github.com/olvrng/ujson#Unquote)
- `[`, `]`: Array
- `{`, `}`: Object

When processing arrays and objects, first the open bracket (`[`, `{`) will be
provided as `value`, followed by its children, and finally the close bracket
(`]`, `}`). When encountering open brackets, You can make the callback function
return `false` to skip the array/object entirely.

## Examples

### 1. Print all keys and values in order

This example gives a quick idea about how *µjson* works.

```go 
 input := []byte(`{  
     "id": 12345,    
     "name": "foo",  
     "numbers": ["one", "two"],  
     "tags": {"color": "red", "priority": "high"},   
     "active": true  
 }`) 
 ujson.Walk(input, func(level int, key, value []byte) bool { 
     fmt.Printf("%2v% 12s : %s\n", level, key, value)    
     return true 
 })
```

```json
{
   "id": 12345,
   "name": "foo",
   "numbers": ["one", "two"],
   "tags": {"color": "red", "priority": "high"},
   "active": true
}
```

Calling `Walk()` with the above input will produce:

| level | key        | value   |
|:-----:|:----------:|:-------:|
|`0`    |            |`{`      |
|`1`    |`"id"`      |`12345`  |
|`1`    |`"name"`    |`"foo"`  |
|`1`    |`"numbers"` |`[`      |
|`2`    |            |`"one"`  |
|`2`    |            |`"two"`  |
|`1`    |            |`]`      |
|`1`    |`"tags"`    |`{`      |
|`2`    |`"color"`   |`"red"`  |
|`2`    |`"priority"`|`"high"` |
|`1`    |            |`}`      |
|`1`    |`"active"`  |`true`   |
|`0`    |            |`}`      |

### 0. The simplest examples

To easily get an idea on `level`, `key` and `value`, here are the simplest
examples:

```go
 input0 := []byte(`true`)
 ujson.Walk(input0, func(level int, key, value []byte) bool {
     fmt.Printf("level=%v key=%s value=%s\n", level, key, value)
     return true
 })
 // output:
 //   level=0 key= value=true

 input1 := []byte(`{ "key": 42 }`)
 ujson.Walk(input1, func(level int, key, value []byte) bool {
     fmt.Printf("level=%v key=%s value=%s\n", level, key, value)
     return true
 })
 // output:
 //   level=0 key= value={
 //   level=1 key="key" value=42
 //   level=0 key= value=}

 input2 := []byte(`[ true ]`)
 ujson.Walk(input2, func(level int, key, value []byte) bool {
     fmt.Printf("level=%v key=%s value=%s\n", level, key, value)
     return true
 })
 // output:
 //   level=0 key= value=[
 //   level=1 key= value=true
 //   level=0 key= value=]
```

In the first example, there is only a single boolean value. The callback
function is called once with `level=0`, `key` is empty and `value=true`.

In the second example, the callback function is called 3 times. Two times for
open and close brackets with `level=0`, `key` is empty and `value` is the
bracketed character. The other time for the only key with `level=1`, `key` is
`"key"` and `value=42`. Note that the key is quoted and you need to call
[`ujson.Unquote()`](https://pkg.go.dev/github.com/olvrng/ujson#Unquote) to
retrieve the unquoted string.

The last example is like the second, but with an array instead. Keys are always
empty inside arrays.

### 2. Reformat input

In this example, the input JSON is formatted with correct indentation. As
processing the input key by key, the callback function reconstructs the JSON. It
outputs each key/value pair in its own line, prefixed with spaces equal to the
param level. There is a catch, though. Valid JSON requires commas between values
in objects and arrays. So there is
[`ujson.ShouldAddComma()`](https://pkg.go.dev/github.com/olvrng/ujson#ShouldAddComma)
for checking whether a comma should be inserted.

```go
 input := []byte(`{"id":12345,"name":"foo","numbers":["one","two"],"tags":{"color":"red","priority":"high"},"active":true}`)

 b := make([]byte, 0, 1024)
 err := ujson.Walk(input, func(level int, key, value []byte) bool {
     if len(b) != 0 && ujson.ShouldAddComma(value, b[len(b)-1]) {
         b = append(b, ',')
     }
     b = append(b, '\n')
     for i := 0; i < level; i++ {
         b = append(b, '\t')
     }
     if len(key) > 0 {
         b = append(b, key...)
         b = append(b, `: `...)
     }
     b = append(b, value...)
     return true
 })
 if err != nil {
     panic(err)
 }
 fmt.Printf("%s\n", b)
```

#### Output:

```json
{
	"id": 12345,
	"name": "foo",
	"numbers": [
		"one",
		"two"
	],
	"tags": {
		"color": "red",
		"priority": "high"
	},
	"active": true
}
```

There is a built-in method
[`ujson.Reconstruct()`](https://pkg.go.dev/github.com/olvrng/ujson?ref=hackernoon.com#Reconstruct)
when you want to remove all the whitespaces.

### 3. Remove blacklisted keys

This example demonstrates removing some keys from the input JSON. The key param
is compared with a pre-defined list. If there is a match, the blacklisted key
and its value are dropped. The callback function returns false for skipping the
entire value (which may be an object or array). Note that the list is quoted,
i.e. "numbers" and "active" instead of number and active. For more advanced
checking, you may want to run [`ujson.Unquote()`](https://pkg.go.dev/github.com/olvrng/ujson#Unquote) on the key.

```go
 input := []byte(`{
     "id": 12345,
     "name": "foo",
     "numbers": ["one", "two"],
     "tags": {"color": "red", "priority": "high"},
     "active": true
 }`)

 blacklistFields := [][]byte{
     []byte(`"numbers"`), // note the quotes
     []byte(`"active"`),
 }
 b := make([]byte, 0, 1024)
 err := ujson.Walk(input, func(_ int, key, value []byte) bool {
     for _, blacklist := range blacklistFields {
         if bytes.Equal(key, blacklist) {
             // remove the key and value from the output
             return false
         }
     }

     // write to output
     if len(b) != 0 && ujson.ShouldAddComma(value, b[len(b)-1]) {
         b = append(b, ',')
     }
     if len(key) > 0 {
         b = append(b, key...)
         b = append(b, ':')
     }
     b = append(b, value...)
     return true
 })
 if err != nil {
     panic(err)
 }
 fmt.Printf("%s\n", b)
```

#### Output:

```json
{"id":12345,"name":"foo","tags":{"color":"red","priority":"high"}}
```

### 4. Wrap int64 in string

This is the original motivation behind µjson. The following example finds keys
ending with `_id"` (`"order_id"`, `"item_id"`, etc.) and converts their values
from numbers to strings, by simply wrapping them in double-quotes.

For valid JSON, values will never be empty. We can test the first byte of value
(`value[0]`) to get its type:

- `n`: Null (`null`)
- `f`, `t`: Boolean (`false`, `true`)
- `0`-`9`, `-`: Number
- `"`: String, see [`Unquote()`](https://pkg.go.dev/github.com/olvrng/ujson#Unquote)
- `[`, `]`: Array
- `{`, `}`: Object

In this case, we check `value[0]` within `0`…`9` to see whether it’s a number,
then insert double-quotes.

```go
 input := []byte(`{"order_id": 12345678901234, "number": 12, "item_id": 12345678905678, "counting": [1,"2",3]}`)

 suffix := []byte(`_id"`) // note the ending quote "
 b := make([]byte, 0, 256)
 err := ujson.Walk(input, func(_ int, key, value []byte) bool {
     // Test for keys with suffix _id" and value is an int64 number. For valid json,
     // values will never be empty, so we can safely test only the first byte.
     shouldWrap := bytes.HasSuffix(key, suffix) && value[0] > '0' && value[0] <= '9'

     // transform the input, wrap values in double quotes
     if len(b) != 0 && ujson.ShouldAddComma(value, b[len(b)-1]) {
         b = append(b, ',')
     }
     if len(key) > 0 {
         b = append(b, key...)
         b = append(b, ':')
     }
     if shouldWrap {
         b = append(b, '"')
     }
     b = append(b, value...)
     if shouldWrap {
         b = append(b, '"')
     }
     return true
 })
 if err != nil {
     panic(err)
 }
 fmt.Printf("%s\n", b)
```

#### Output:

```json
{"order_id":"12345678901234","number":12,"item_id":"12345678905678","counting":[1,"2",3]}
```

After processing, the numbers in `"order_id"` and `"item_id"` are quoted as strings. And JavaScript should be happy now! 🎉 🎉

## License

MIT
