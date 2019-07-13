package aws

import (
	"testing"
)

var base64encodingTests = []struct {
	in  []byte
	out string
}{
	// normal encoding case
	{[]byte("data should be encoded"), "ZGF0YSBzaG91bGQgYmUgZW5jb2RlZA=="},
	// base64 encoded input should result in no change of output
	{[]byte("ZGF0YSBzaG91bGQgYmUgZW5jb2RlZA=="), "ZGF0YSBzaG91bGQgYmUgZW5jb2RlZA=="},
}

func TestBase64Encode(t *testing.T) {
	for _, tt := range base64encodingTests {
		out := base64Encode(tt.in)
		if out != tt.out {
			t.Errorf("base64Encode(%s) => %s, want %s", tt.in, out, tt.out)
		}
	}
}

func TestLooksLikeJsonString(t *testing.T) {
	looksLikeJson := ` {"abc":"1"} `
	doesNotLookLikeJson := `abc: 1`

	if !looksLikeJsonString(looksLikeJson) {
		t.Errorf("Expected looksLikeJson to return true for %s", looksLikeJson)
	}
	if looksLikeJsonString(doesNotLookLikeJson) {
		t.Errorf("Expected looksLikeJson to return false for %s", doesNotLookLikeJson)
	}
}

func TestJsonBytesEqualQuotedAndUnquoted(t *testing.T) {
	unquoted := `{"test": "test"}`
	quoted := "{\"test\": \"test\"}"

	if !jsonBytesEqual([]byte(unquoted), []byte(quoted)) {
		t.Errorf("Expected jsonBytesEqual to return true for %s == %s", unquoted, quoted)
	}

	unquotedDiff := `{"test": "test"}`
	quotedDiff := "{\"test\": \"tested\"}"

	if jsonBytesEqual([]byte(unquotedDiff), []byte(quotedDiff)) {
		t.Errorf("Expected jsonBytesEqual to return false for %s == %s", unquotedDiff, quotedDiff)
	}
}

func TestJsonBytesEqualWhitespaceAndNoWhitespace(t *testing.T) {
	noWhitespace := `{"test":"test"}`
	whitespace := `
{
  "test": "test"
}`

	if !jsonBytesEqual([]byte(noWhitespace), []byte(whitespace)) {
		t.Errorf("Expected jsonBytesEqual to return true for %s == %s", noWhitespace, whitespace)
	}

	noWhitespaceDiff := `{"test":"test"}`
	whitespaceDiff := `
{
  "test": "tested"
}`

	if jsonBytesEqual([]byte(noWhitespaceDiff), []byte(whitespaceDiff)) {
		t.Errorf("Expected jsonBytesEqual to return false for %s == %s", noWhitespaceDiff, whitespaceDiff)
	}
}

func TestInterfaceStringSlice(t *testing.T) {
	data := []string{"a", "b", "c"}
	var tests = []struct {
		input []interface{}
		want  []*string
	}{
		{nil, nil},
		{
			[]interface{}{"a", "b", "c"},
			[]*string{&data[0], &data[1], &data[2]},
		},
	}
	for _, test := range tests {
		got := interfaceStringSlice(test.input)
		if len(test.want) != len(got) {
			t.Errorf("interfaceStringSlice(%v) returned len(%d), want len(%d)", test.input, len(got), len(test.want))
		}
		for j, v := range test.want {
			if *v != *got[j] {
				t.Errorf("interfaceStringSlice(%v) returned unexpected output", test.input)
				break
			}
		}
	}
}
