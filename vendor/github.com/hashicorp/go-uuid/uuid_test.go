package uuid

import (
	"crypto/rand"
	"reflect"
	"regexp"
	"testing"
)

func TestGenerateUUID(t *testing.T) {
	prev, err := GenerateUUID()
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 100; i++ {
		id, err := GenerateUUID()
		if err != nil {
			t.Fatal(err)
		}
		if prev == id {
			t.Fatalf("Should get a new ID!")
		}

		matched, err := regexp.MatchString(
			"[\\da-f]{8}-[\\da-f]{4}-[\\da-f]{4}-[\\da-f]{4}-[\\da-f]{12}", id)
		if !matched || err != nil {
			t.Fatalf("expected match %s %v %s", id, matched, err)
		}
	}
}

func TestParseUUID(t *testing.T) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		t.Fatalf("failed to read random bytes: %v", err)
	}

	uuidStr, err := FormatUUID(buf)
	if err != nil {
		t.Fatal(err)
	}

	parsedStr, err := ParseUUID(uuidStr)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(parsedStr, buf) {
		t.Fatalf("mismatched buffers")
	}
}

func BenchmarkGenerateUUID(b *testing.B) {
	for n := 0; n < b.N; n++ {
		_, _ = GenerateUUID()
	}
}
