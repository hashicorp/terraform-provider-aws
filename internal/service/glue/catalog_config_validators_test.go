// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package glue_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfglue "github.com/hashicorp/terraform-provider-aws/internal/service/glue"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestCatalogConfigValidators(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	res, err := tfglue.ResourceCatalog(ctx)
	if err != nil {
		t.Fatalf("ResourceCatalog: %s", err)
	}

	withValidators, ok := res.(resource.ResourceWithConfigValidators)
	if !ok {
		t.Fatalf("%T does not implement resource.ResourceWithConfigValidators", res)
	}

	schemaResp := &resource.SchemaResponse{}
	withValidators.Schema(ctx, resource.SchemaRequest{}, schemaResp)
	if schemaResp.Diagnostics.HasError() {
		t.Fatalf("Schema: %s", schemaResp.Diagnostics)
	}

	rootType := schemaResp.Schema.Type().TerraformType(ctx)
	validators := withValidators.ConfigValidators(ctx)
	if len(validators) == 0 {
		t.Fatal("expected config validators")
	}

	testCases := map[string]struct {
		spec    map[string]any
		wantErr bool
	}{
		"no blocks": {
			spec: map[string]any{
				names.AttrName: acctest.Ct12Digit,
			},
			wantErr: true,
		},
		"federated catalog only": {
			spec: map[string]any{
				names.AttrName: "s3tablescatalog",
				"federated_catalog": []any{
					map[string]any{},
				},
			},
		},
		"data lake access properties only": {
			spec: map[string]any{
				names.AttrName: acctest.Ct12Digit,
				"catalog_properties": []any{
					map[string]any{
						"data_lake_access_properties": []any{
							map[string]any{},
						},
					},
				},
			},
		},
		"iceberg optimization properties only": {
			spec: map[string]any{
				names.AttrName: acctest.Ct12Digit,
				"catalog_properties": []any{
					map[string]any{
						"iceberg_optimization_properties": []any{
							map[string]any{},
						},
					},
				},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			raw, err := catalogTerraformValue(rootType, tc.spec)
			if err != nil {
				t.Fatalf("building config: %s", err)
			}

			cfg := tfsdk.Config{
				Raw:    raw,
				Schema: schemaResp.Schema,
			}

			var diags resource.ValidateConfigResponse
			for _, validator := range validators {
				resp := &resource.ValidateConfigResponse{}
				validator.ValidateResource(ctx, resource.ValidateConfigRequest{Config: cfg}, resp)
				diags.Diagnostics.Append(resp.Diagnostics...)
			}

			if !tc.wantErr {
				if diags.Diagnostics.HasError() {
					t.Fatalf("unexpected diagnostics: %s", diags.Diagnostics)
				}
				return
			}

			if !diags.Diagnostics.HasError() {
				t.Fatal("expected Missing Attribute Configuration")
			}
			for _, d := range diags.Diagnostics {
				if d.Summary() == "Missing Attribute Configuration" {
					return
				}
			}
			t.Fatalf("expected Missing Attribute Configuration, got: %s", diags.Diagnostics)
		})
	}
}

func catalogTerraformValue(typ tftypes.Type, spec any) (tftypes.Value, error) {
	if spec == nil {
		return tftypes.NewValue(typ, nil), nil
	}

	switch t := typ.(type) {
	case tftypes.Object:
		m, ok := spec.(map[string]any)
		if !ok {
			return tftypes.Value{}, fmt.Errorf("object wants map[string]any, got %T", spec)
		}
		for name := range m {
			if _, ok := t.AttributeTypes[name]; !ok {
				return tftypes.Value{}, fmt.Errorf("unknown attribute %q", name)
			}
		}
		attrs := make(map[string]tftypes.Value, len(t.AttributeTypes))
		for name, attrType := range t.AttributeTypes {
			val, err := catalogTerraformValue(attrType, m[name])
			if err != nil {
				return tftypes.Value{}, fmt.Errorf("%s: %w", name, err)
			}
			attrs[name] = val
		}
		return tftypes.NewValue(t, attrs), nil

	case tftypes.List:
		items, ok := spec.([]any)
		if !ok {
			return tftypes.Value{}, fmt.Errorf("list wants []any, got %T", spec)
		}
		vals := make([]tftypes.Value, len(items))
		for i, item := range items {
			val, err := catalogTerraformValue(t.ElementType, item)
			if err != nil {
				return tftypes.Value{}, fmt.Errorf("[%d]: %w", i, err)
			}
			vals[i] = val
		}
		return tftypes.NewValue(t, vals), nil

	case tftypes.Map:
		m, ok := spec.(map[string]any)
		if !ok {
			return tftypes.Value{}, fmt.Errorf("map wants map[string]any, got %T", spec)
		}
		vals := make(map[string]tftypes.Value, len(m))
		for key, item := range m {
			val, err := catalogTerraformValue(t.ElementType, item)
			if err != nil {
				return tftypes.Value{}, fmt.Errorf("[%s]: %w", key, err)
			}
			vals[key] = val
		}
		return tftypes.NewValue(t, vals), nil

	default:
		switch {
		case typ.Is(tftypes.String):
			s, ok := spec.(string)
			if !ok {
				return tftypes.Value{}, fmt.Errorf("string wants string, got %T", spec)
			}
			return tftypes.NewValue(typ, s), nil
		case typ.Is(tftypes.Bool):
			b, ok := spec.(bool)
			if !ok {
				return tftypes.Value{}, fmt.Errorf("bool wants bool, got %T", spec)
			}
			return tftypes.NewValue(typ, b), nil
		default:
			return tftypes.Value{}, fmt.Errorf("unsupported terraform type %s", typ)
		}
	}
}
