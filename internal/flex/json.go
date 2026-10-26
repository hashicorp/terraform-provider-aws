// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package flex

import (
	"context"
	"fmt"
	"math/big"

	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	tfjson "github.com/hashicorp/terraform-provider-aws/internal/json"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
)

type TypeAtPather interface {
	TypeAtPath(context.Context, *tftypes.AttributePath) (tftypes.Type, error)
}

type JSONFlexer interface {
	Expand(context.Context, TypeAtPather, tftypes.Value) (string, error)
	Flatten(context.Context, TypeAtPather, string) (tftypes.Value, error)
}

var (
	_ JSONFlexer = jsonFlex{}
)

func NewJSONFlex(tfToJSONNameMap map[string]string) JSONFlexer {
	jsonToTFNameMap := make(map[string]string)
	for k, v := range tfToJSONNameMap {
		jsonToTFNameMap[v] = k
	}

	return jsonFlex{
		tfToJSONNameMap: tfToJSONNameMap,
		jsonToTFNameMap: jsonToTFNameMap,
	}
}

// JSONFlexer provides methods to expand and flatten JSON representation of Terraform values.
// The initial use case is for the Cloud Control API.
// Cribbed from terraform-provider-awscc/internal/generic/translate.go@v1.107.0.
type jsonFlex struct {
	tfToJSONNameMap map[string]string
	jsonToTFNameMap map[string]string
}

// Expand converts a Terraform value into its JSON representation.
func (j jsonFlex) Expand(ctx context.Context, typeAtPather TypeAtPather, val tftypes.Value) (string, error) {
	v, err := j.toRaw(ctx, typeAtPather, nil, val)
	if err != nil {
		return "", err
	}
	if v, ok := v.(map[string]any); ok {
		return tfjson.EncodeToString(v)
	}
	return "", fmt.Errorf("unexpected raw type: %T", v)
}

// Flatten converts a JSON representation into its Terraform value.
func (j jsonFlex) Flatten(ctx context.Context, typeAtPather TypeAtPather, s string) (tftypes.Value, error) {
	var v any
	if err := tfjson.DecodeFromString(s, &v); err != nil {
		return inttypes.Zero[tftypes.Value](), err
	}
	if v, ok := v.(map[string]any); ok {
		return j.fromRaw(ctx, typeAtPather, nil, v)
	}
	return inttypes.Zero[tftypes.Value](), fmt.Errorf("unexpected raw type: %T", v)
}

func (j jsonFlex) toRaw(ctx context.Context, typeAtPather TypeAtPather, path *tftypes.AttributePath, val tftypes.Value) (any, error) {
	if val.IsNull() || !val.IsKnown() {
		return make(map[string]any), nil
	}

	switch typ, err := typeAtPather.TypeAtPath(ctx, path); {
	case err != nil:
		return nil, err

	//
	// Primitive types.
	//
	case typ.Is(tftypes.Bool):
		var b bool
		if err := val.As(&b); err != nil {
			return nil, err
		}
		return b, nil

	case typ.Is(tftypes.Number):
		n := big.NewFloat(0)
		if err := val.As(&n); err != nil {
			return nil, err
		}
		f, _ := n.Float64()
		return f, nil

	case typ.Is(tftypes.String):
		var s string
		if err := val.As(&s); err != nil {
			return nil, err
		}
		return s, nil

		//
		// Complex types.
		//
	case typ.Is(tftypes.List{}), typ.Is(tftypes.Set{}), typ.Is(tftypes.Tuple{}):
		var vals []tftypes.Value
		if err := val.As(&vals); err != nil {
			return nil, err
		}
		vs := make([]any, 0)
		for idx, val := range vals {
			if typ.Is(tftypes.Set{}) {
				// No need to worry about a specific value here.
				path = path.WithElementKeyValue(tftypes.NewValue(typ.(tftypes.Set).ElementType, nil))
			} else {
				path = path.WithElementKeyInt(idx)
			}
			v, err := j.toRaw(ctx, typeAtPather, path, val)
			if err != nil {
				return nil, err
			}
			path = path.WithoutLastStep()
			if v == nil {
				continue
			}
			vs = append(vs, v)
		}
		if len(vs) == 0 {
			return nil, nil
		}
		return vs, nil

	case typ.Is(tftypes.Map{}), typ.Is(tftypes.Object{}):
		var vals map[string]tftypes.Value
		if err := val.As(&vals); err != nil {
			return nil, err
		}
		vs := make(map[string]any)
		for name, val := range vals {
			if typ.Is(tftypes.Object{}) {
				path = path.WithAttributeName(name)
			} else {
				path = path.WithElementKeyString(name)
			}

			var propertyName string
			if typ.Is(tftypes.Object{}) {
				var ok bool
				propertyName, ok = j.tfToJSONNameMap[name]
				if !ok {
					return nil, fmt.Errorf("attribute name mapping not found: %s", name)
				}
			}

			v, err := j.toRaw(ctx, typeAtPather, path, val)
			if err != nil {
				return nil, err
			}
			path = path.WithoutLastStep()
			if v == nil {
				continue
			}
			if typ.Is(tftypes.Object{}) {
				vs[propertyName] = v
			} else {
				vs[name] = v
			}
		}
		if len(vs) == 0 {
			return nil, nil
		}
		return vs, nil

	default:
		return nil, fmt.Errorf("unsupported value type: %s", typ)
	}
}

func (j jsonFlex) fromRaw(ctx context.Context, typeAtPather TypeAtPather, path *tftypes.AttributePath, v any) (tftypes.Value, error) {
	typ, err := typeAtPather.TypeAtPath(ctx, path)
	if err != nil {
		return inttypes.Zero[tftypes.Value](), err
	}

	switch v := v.(type) {
	//
	// Primitive types.
	//
	case bool:
		return tftypes.NewValue(tftypes.Bool, v), nil

	case float64:
		return tftypes.NewValue(tftypes.Number, v), nil

	case string:
		return tftypes.NewValue(tftypes.String, v), nil

	//
	// Complex types.
	//
	case []any:
		if len(v) == 0 {
			return tftypes.NewValue(typ, nil), nil
		}

		if typ.Is(tftypes.String) {
			// Value is JSON string.
			val, err := tfjson.EncodeToString(v)
			if err != nil {
				return inttypes.Zero[tftypes.Value](), err
			}
			return tftypes.NewValue(typ, val), nil
		}

		var vals []tftypes.Value
		for idx, v := range v {
			if typ.Is(tftypes.Set{}) {
				// No need to worry about a specific value here.
				path = path.WithElementKeyValue(tftypes.NewValue(typ.(tftypes.Set).ElementType, nil))
			} else {
				path = path.WithElementKeyInt(idx)
			}
			val, err := j.fromRaw(ctx, typeAtPather, path, v)
			if err != nil {
				return inttypes.Zero[tftypes.Value](), err
			}
			vals = append(vals, val)
			path = path.WithoutLastStep()
		}
		return tftypes.NewValue(typ, vals), nil

	case map[string]any:
		if typ.Is(tftypes.String) {
			// Value is JSON string.
			val, err := tfjson.EncodeToString(v)
			if err != nil {
				return inttypes.Zero[tftypes.Value](), err
			}
			return tftypes.NewValue(typ, val), nil
		}

		isObject := typ.Is(tftypes.Object{})
		vals := make(map[string]tftypes.Value)
		for key, v := range v {
			var attributeName string
			if isObject {
				var ok bool
				attributeName, ok = j.jsonToTFNameMap[key]
				if !ok {
					tflog.Info(ctx, "attribute name mapping not found", map[string]any{
						"key": key,
					})
					continue
				}
				path = path.WithAttributeName(attributeName)
			} else {
				path = path.WithElementKeyString(key)
			}

			val, err := j.fromRaw(ctx, typeAtPather, path, v)
			if err != nil {
				if isObject {
					tflog.Info(ctx, "not found in Terraform schema", map[string]any{
						"key":   key,
						"path":  path,
						"error": err.Error(),
					})
					path = path.WithoutLastStep()
					continue
				}
				return tftypes.Value{}, err
			}
			if isObject {
				vals[attributeName] = val
			} else {
				vals[key] = val
			}
			path = path.WithoutLastStep()
		}
		if isObject {
			// Set any missing attributes to Null.
			for k, t := range typ.(tftypes.Object).AttributeTypes {
				if _, ok := vals[k]; !ok {
					vals[k] = tftypes.NewValue(t, nil)
				}
			}
		}
		return tftypes.NewValue(typ, vals), nil

	default:
		return inttypes.Zero[tftypes.Value](), fmt.Errorf("unsupported raw type: %T", v)
	}
}
