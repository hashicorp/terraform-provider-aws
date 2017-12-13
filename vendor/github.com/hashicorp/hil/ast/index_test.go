package ast

import (
	"strings"
	"testing"
)

func TestIndexTypeMap_empty(t *testing.T) {
	i := &Index{
		Target: &VariableAccess{Name: "foo"},
		Key: &LiteralNode{
			Typex: TypeString,
			Value: "bar",
		},
	}

	scope := &BasicScope{
		VarMap: map[string]Variable{
			"foo": Variable{
				Type:  TypeMap,
				Value: map[string]Variable{},
			},
		},
	}

	actual, err := i.Type(scope)
	if err == nil || !strings.Contains(err.Error(), "does not have any elements") {
		t.Fatalf("bad err: %s", err)
	}
	if actual != TypeInvalid {
		t.Fatalf("bad: %s", actual)
	}
}

func TestIndexTypeMap_string(t *testing.T) {
	i := &Index{
		Target: &VariableAccess{Name: "foo"},
		Key: &LiteralNode{
			Typex: TypeString,
			Value: "bar",
		},
	}

	scope := &BasicScope{
		VarMap: map[string]Variable{
			"foo": Variable{
				Type: TypeMap,
				Value: map[string]Variable{
					"baz": Variable{
						Type:  TypeString,
						Value: "Hello",
					},
					"bar": Variable{
						Type:  TypeString,
						Value: "World",
					},
				},
			},
		},
	}

	actual, err := i.Type(scope)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if actual != TypeString {
		t.Fatalf("bad: %s", actual)
	}
}

func TestIndexTypeMap_int(t *testing.T) {
	i := &Index{
		Target: &VariableAccess{Name: "foo"},
		Key: &LiteralNode{
			Typex: TypeString,
			Value: "bar",
		},
	}

	scope := &BasicScope{
		VarMap: map[string]Variable{
			"foo": Variable{
				Type: TypeMap,
				Value: map[string]Variable{
					"baz": Variable{
						Type:  TypeInt,
						Value: 1,
					},
					"bar": Variable{
						Type:  TypeInt,
						Value: 2,
					},
				},
			},
		},
	}

	actual, err := i.Type(scope)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if actual != TypeInt {
		t.Fatalf("bad: %s", actual)
	}
}

func TestIndexTypeMap_nonHomogenous(t *testing.T) {
	i := &Index{
		Target: &VariableAccess{Name: "foo"},
		Key: &LiteralNode{
			Typex: TypeString,
			Value: "bar",
		},
	}

	scope := &BasicScope{
		VarMap: map[string]Variable{
			"foo": Variable{
				Type: TypeMap,
				Value: map[string]Variable{
					"bar": Variable{
						Type:  TypeString,
						Value: "Hello",
					},
					"baz": Variable{
						Type:  TypeInt,
						Value: 43,
					},
				},
			},
		},
	}

	_, err := i.Type(scope)
	if err == nil || !strings.Contains(err.Error(), "homogenous") {
		t.Fatalf("expected error")
	}
}

func TestIndexTypeList_empty(t *testing.T) {
	i := &Index{
		Target: &VariableAccess{Name: "foo"},
		Key: &LiteralNode{
			Typex: TypeInt,
			Value: 1,
		},
	}

	scope := &BasicScope{
		VarMap: map[string]Variable{
			"foo": Variable{
				Type:  TypeList,
				Value: []Variable{},
			},
		},
	}

	actual, err := i.Type(scope)
	if err == nil || !strings.Contains(err.Error(), "does not have any elements") {
		t.Fatalf("bad err: %s", err)
	}
	if actual != TypeInvalid {
		t.Fatalf("bad: %s", actual)
	}
}

func TestIndexTypeList_string(t *testing.T) {
	i := &Index{
		Target: &VariableAccess{Name: "foo"},
		Key: &LiteralNode{
			Typex: TypeInt,
			Value: 1,
		},
	}

	scope := &BasicScope{
		VarMap: map[string]Variable{
			"foo": Variable{
				Type: TypeList,
				Value: []Variable{
					Variable{
						Type:  TypeString,
						Value: "Hello",
					},
					Variable{
						Type:  TypeString,
						Value: "World",
					},
				},
			},
		},
	}

	actual, err := i.Type(scope)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if actual != TypeString {
		t.Fatalf("bad: %s", actual)
	}
}

func TestIndexTypeList_int(t *testing.T) {
	i := &Index{
		Target: &VariableAccess{Name: "foo"},
		Key: &LiteralNode{
			Typex: TypeInt,
			Value: 1,
		},
	}

	scope := &BasicScope{
		VarMap: map[string]Variable{
			"foo": Variable{
				Type: TypeList,
				Value: []Variable{
					Variable{
						Type:  TypeInt,
						Value: 34,
					},
					Variable{
						Type:  TypeInt,
						Value: 54,
					},
				},
			},
		},
	}

	actual, err := i.Type(scope)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if actual != TypeInt {
		t.Fatalf("bad: %s", actual)
	}
}

func TestIndexTypeList_nonHomogenous(t *testing.T) {
	i := &Index{
		Target: &VariableAccess{Name: "foo"},
		Key: &LiteralNode{
			Typex: TypeInt,
			Value: 1,
		},
	}

	scope := &BasicScope{
		VarMap: map[string]Variable{
			"foo": Variable{
				Type: TypeList,
				Value: []Variable{
					Variable{
						Type:  TypeString,
						Value: "Hello",
					},
					Variable{
						Type:  TypeInt,
						Value: 43,
					},
				},
			},
		},
	}

	_, err := i.Type(scope)
	if err == nil || !strings.Contains(err.Error(), "homogenous") {
		t.Fatalf("expected error")
	}
}
