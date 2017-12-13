package ast

import (
	"fmt"
	"reflect"
	"testing"
)

func TestLiteralNodeType(t *testing.T) {
	c := &LiteralNode{Typex: TypeString}
	actual, err := c.Type(nil)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if actual != TypeString {
		t.Fatalf("bad: %s", actual)
	}
}

func TestNewLiteralNode(t *testing.T) {
	tests := []struct {
		Value    interface{}
		Expected *LiteralNode
	}{
		{
			1,
			&LiteralNode{
				Value: 1,
				Typex: TypeInt,
			},
		},
		{
			1.0,
			&LiteralNode{
				Value: 1.0,
				Typex: TypeFloat,
			},
		},
		{
			true,
			&LiteralNode{
				Value: true,
				Typex: TypeBool,
			},
		},
		{
			"hi",
			&LiteralNode{
				Value: "hi",
				Typex: TypeString,
			},
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%T", test.Value), func(t *testing.T) {
			inPos := Pos{
				Column:   2,
				Line:     3,
				Filename: "foo",
			}
			node, err := NewLiteralNode(test.Value, inPos)

			if err != nil {
				t.Fatalf("error: %s", err)
			}

			if got, want := node.Typex, test.Expected.Typex; want != got {
				t.Errorf("got type %s; want %s", got, want)
			}
			if got, want := node.Value, test.Expected.Value; want != got {
				t.Errorf("got value %#v; want %#v", got, want)
			}
			if got, want := node.Posx, inPos; !reflect.DeepEqual(got, want) {
				t.Errorf("got position %#v; want %#v", got, want)
			}
		})
	}
}
