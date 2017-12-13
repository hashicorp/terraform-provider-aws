package scanner

import (
	"testing"
)

func TestPeeker(t *testing.T) {
	ch := make(chan *Token)

	go func() {
		ch <- &Token{
			Type:    IDENTIFIER,
			Content: "foo",
		}
		ch <- &Token{
			Type:    INTEGER,
			Content: "1",
		}
		ch <- &Token{
			Type:    EOF,
			Content: "",
		}
		close(ch)
	}()

	peeker := NewPeeker(ch)

	if got, want := peeker.Peek().Type, IDENTIFIER; got != want {
		t.Fatalf("first peek returned %s; want %s", got, want)
	}
	if got, want := peeker.Read().Type, IDENTIFIER; got != want {
		t.Fatalf("first read returned %s; want %s", got, want)
	}
	if got, want := peeker.Peek().Type, INTEGER; got != want {
		t.Fatalf("second peek returned %s; want %s", got, want)
	}
	if got, want := peeker.Peek().Type, INTEGER; got != want {
		t.Fatalf("third peek returned %s; want %s", got, want)
	}
	if got, want := peeker.Read().Type, INTEGER; got != want {
		t.Fatalf("second read returned %s; want %s", got, want)
	}
	if got, want := peeker.Read().Type, EOF; got != want {
		t.Fatalf("third read returned %s; want %s", got, want)
	}
	// reading again after EOF just returns EOF again
	if got, want := peeker.Read().Type, EOF; got != want {
		t.Fatalf("final read returned %s; want %s", got, want)
	}
	if got, want := peeker.Peek().Type, EOF; got != want {
		t.Fatalf("final peek returned %s; want %s", got, want)
	}

	peeker.Close()
	if got, want := peeker.Peek().Type, EOF; got != want {
		t.Fatalf("peek after close returned %s; want %s", got, want)
	}
	if got, want := peeker.Read().Type, EOF; got != want {
		t.Fatalf("read after close returned %s; want %s", got, want)
	}
}
