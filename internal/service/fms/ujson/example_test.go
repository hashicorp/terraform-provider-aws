package ujson_test

import (
	"bytes"
	"fmt"

	"github.com/olvrng/ujson"
)

func ExampleWalk() {
	input := []byte(`{"order_id": 12345678901234, "number": 12, "item_id": 12345678905678, "counting": [1,"2",3]}`)

	err := ujson.Walk(input, func(level int, key, value []byte) bool {
		fmt.Println(level, string(key), string(value))
		return true
	})
	if err != nil {
		panic(err)
	}
	// Output:
	// 0  {
	// 1 "order_id" 12345678901234
	// 1 "number" 12
	// 1 "item_id" 12345678905678
	// 1 "counting" [
	// 2  1
	// 2  "2"
	// 2  3
	// 1  ]
	// 0  }
}

func ExampleWalk_reconstruct() {
	input := []byte(`{"order_id": 12345678901234, "number": 12, "item_id": 12345678905678, "counting": [1,"2",3]}`)

	b := make([]byte, 0, 256)
	err := ujson.Walk(input, func(level int, key, value []byte) bool {
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
	fmt.Printf("%s", b)
	// Output: {"order_id":12345678901234,"number":12,"item_id":12345678905678,"counting":[1,"2",3]}
}

func ExampleWalk_reformat() {
	input := []byte(`{"order_id": 12345678901234, "number": 12, "item_id": 12345678905678, "counting": [1,"2",3]}`)

	b := make([]byte, 0, 256)
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
	fmt.Printf("%s", b)
	// Output:
	// {
	//	"order_id": 12345678901234,
	//	"number": 12,
	//	"item_id": 12345678905678,
	//	"counting": [
	//		1,
	//		"2",
	//		3
	//	]
	// }
}

func ExampleWalk_wrapInt64InString() {
	input := []byte(`{"order_id": 12345678901234, "number": 12, "item_id": 12345678905678, "counting": [1,"2",3]}`)

	suffix := []byte(`_id`)
	b := make([]byte, 0, 256)
	err := ujson.Walk(input, func(_ int, key, value []byte) bool {
		// unquote key
		if len(key) != 0 {
			key = key[1 : len(key)-1]
		}

		// Test for field with suffix _id and value is an int64 number. For
		// valid json, value will never be empty, so we can safely test only the
		// first byte.
		wrap := bytes.HasSuffix(key, suffix) && value[0] > '0' && value[0] <= '9'

		// transform the input, wrap values in double quote
		if len(b) != 0 && ujson.ShouldAddComma(value, b[len(b)-1]) {
			b = append(b, ',')
		}
		if len(key) > 0 {
			b = append(b, '"')
			b = append(b, key...)
			b = append(b, '"')
			b = append(b, ':')
		}
		if wrap {
			b = append(b, '"')
		}
		b = append(b, value...)
		if wrap {
			b = append(b, '"')
		}
		return true
	})
	if err != nil {
		panic(err)
	}
	fmt.Printf("%s", b)
	// Output: {"order_id":"12345678901234","number":12,"item_id":"12345678905678","counting":[1,"2",3]}
}

func ExampleWalk_removeBlacklistFields() {
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
	fmt.Printf("%s", b)
	// Output: {"id":12345,"name":"foo","tags":{"color":"red","priority":"high"}}
}

// This example was taken from StackOverflow:
// https://stackoverflow.com/questions/35441254/making-minimal-modification-to-json-data-without-a-structure-in-golang
func ExampleWalk_removeBlacklistFields2() {
	input := []byte(`
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
	}`)

	blacklistFields := [][]byte{
		[]byte(`"responseHeader"`), // note the quotes
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
	fmt.Printf("%s", b)
	// Output: {"response":{"numFound":2,"start":0,"docs":[{"name":"foo"},{"name":"bar"}]}}
}
