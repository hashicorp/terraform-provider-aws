// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package timestreaminfluxdb

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-provider-aws/names"
)

type modifyPlanTestCase struct {
	config    map[string]tftypes.Value
	state     map[string]tftypes.Value // nil means the resource is being created.
	wantError bool
}

type modifyPlanResource interface {
	resource.Resource
	resource.ResourceWithModifyPlan
}

func runModifyPlanTests(t *testing.T, factory func(context.Context) (resource.ResourceWithConfigure, error), testCases map[string]modifyPlanTestCase) {
	t.Helper()

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx := t.Context()

			r, err := factory(ctx)
			if err != nil {
				t.Fatalf("creating resource: %s", err)
			}

			var schemaResp resource.SchemaResponse
			r.Schema(ctx, resource.SchemaRequest{}, &schemaResp)
			if schemaResp.Diagnostics.HasError() {
				t.Fatalf("getting schema: %s", schemaResp.Diagnostics)
			}
			s := schemaResp.Schema
			// The region attribute is normally injected by the provider's schema interceptor.
			s.Attributes[names.AttrRegion] = schema.StringAttribute{Optional: true, Computed: true}

			config := testModifyPlanObjectValue(ctx, s, testCase.config)
			state := tftypes.NewValue(s.Type().TerraformType(ctx), nil)
			if testCase.state != nil {
				state = testModifyPlanObjectValue(ctx, s, testCase.state)
			}

			req := resource.ModifyPlanRequest{
				Config: tfsdk.Config{Schema: s, Raw: config},
				Plan:   tfsdk.Plan{Schema: s, Raw: config},
				State:  tfsdk.State{Schema: s, Raw: state},
			}
			resp := resource.ModifyPlanResponse{
				Plan: req.Plan,
			}

			r.(modifyPlanResource).ModifyPlan(ctx, req, &resp)

			if got, want := resp.Diagnostics.HasError(), testCase.wantError; got != want {
				t.Errorf("HasError() = %t, want %t: %s", got, want, resp.Diagnostics)
			}
		})
	}
}

// testModifyPlanObjectValue returns an object value for the schema with all attributes null except those specified.
func testModifyPlanObjectValue(ctx context.Context, s schema.Schema, values map[string]tftypes.Value) tftypes.Value {
	typ := s.Type().TerraformType(ctx).(tftypes.Object)

	vals := make(map[string]tftypes.Value, len(typ.AttributeTypes))
	for name, attrType := range typ.AttributeTypes {
		if v, ok := values[name]; ok {
			vals[name] = v
		} else {
			vals[name] = tftypes.NewValue(attrType, nil)
		}
	}

	return tftypes.NewValue(typ, vals)
}

func testModifyPlanRestoreValue(ctx context.Context, s schema.Schema, sourceDBBackupID string) tftypes.Value {
	listType := s.Type().TerraformType(ctx).(tftypes.Object).AttributeTypes["restore"].(tftypes.List)
	elemType := listType.ElementType.(tftypes.Object)

	return tftypes.NewValue(listType, []tftypes.Value{
		tftypes.NewValue(elemType, map[string]tftypes.Value{
			"restore_mode":        tftypes.NewValue(tftypes.String, nil),
			"restore_to_time":     tftypes.NewValue(tftypes.String, nil),
			"source_db_backup_id": tftypes.NewValue(tftypes.String, sourceDBBackupID),
		}),
	})
}

func testModifyPlanStringSet(values ...string) tftypes.Value {
	elems := make([]tftypes.Value, len(values))
	for i, v := range values {
		elems[i] = tftypes.NewValue(tftypes.String, v)
	}

	return tftypes.NewValue(tftypes.Set{ElementType: tftypes.String}, elems)
}

func testModifyPlanUnknownString() tftypes.Value {
	return tftypes.NewValue(tftypes.String, tftypes.UnknownValue)
}

func testDBInstanceModifyPlanConfig(overrides map[string]tftypes.Value) map[string]tftypes.Value {
	config := map[string]tftypes.Value{
		"allocated_storage":      tftypes.NewValue(tftypes.Number, 20),
		"bucket":                 tftypes.NewValue(tftypes.String, "initial"),
		"db_instance_type":       tftypes.NewValue(tftypes.String, "db.influx.medium"),
		names.AttrName:           tftypes.NewValue(tftypes.String, "test"),
		"organization":           tftypes.NewValue(tftypes.String, "organization"),
		"password":               tftypes.NewValue(tftypes.String, "testpassword"),
		"username":               tftypes.NewValue(tftypes.String, "admin"),
		"vpc_security_group_ids": testModifyPlanStringSet("sg-12345678"),
		"vpc_subnet_ids":         testModifyPlanStringSet("subnet-12345678"),
	}

	for k, v := range overrides {
		config[k] = v
	}

	return config
}

func TestDBInstanceModifyPlan_requiredArguments(t *testing.T) {
	t.Parallel()

	runModifyPlanTests(t, newDBInstanceResource, map[string]modifyPlanTestCase{
		"all known": {
			config: testDBInstanceModifyPlanConfig(nil),
		},
		"unknown password": {
			config: testDBInstanceModifyPlanConfig(map[string]tftypes.Value{
				"password": testModifyPlanUnknownString(),
			}),
		},
		"unknown vpc_subnet_ids": {
			config: testDBInstanceModifyPlanConfig(map[string]tftypes.Value{
				"vpc_subnet_ids": tftypes.NewValue(tftypes.Set{ElementType: tftypes.String}, tftypes.UnknownValue),
			}),
		},
		"missing password": {
			config: testDBInstanceModifyPlanConfig(map[string]tftypes.Value{
				"password": tftypes.NewValue(tftypes.String, nil),
			}),
			wantError: true,
		},
	})
}

func testDBClusterModifyPlanConfig(overrides map[string]tftypes.Value) map[string]tftypes.Value {
	config := testDBInstanceModifyPlanConfig(overrides)
	config["vpc_subnet_ids"] = testModifyPlanStringSet("subnet-12345678", "subnet-87654321")
	if v, ok := overrides["vpc_subnet_ids"]; ok {
		config["vpc_subnet_ids"] = v
	}

	return config
}

func TestDBClusterModifyPlan_requiredArguments(t *testing.T) {
	t.Parallel()

	runModifyPlanTests(t, newDBClusterResource, map[string]modifyPlanTestCase{
		"all known": {
			config: testDBClusterModifyPlanConfig(nil),
		},
		"unknown password": {
			config: testDBClusterModifyPlanConfig(map[string]tftypes.Value{
				"password": testModifyPlanUnknownString(),
			}),
		},
		"unknown db_instance_type": {
			config: testDBClusterModifyPlanConfig(map[string]tftypes.Value{
				"db_instance_type": testModifyPlanUnknownString(),
			}),
		},
		"missing password": {
			config: testDBClusterModifyPlanConfig(map[string]tftypes.Value{
				"password": tftypes.NewValue(tftypes.String, nil),
			}),
			wantError: true,
		},
		"missing db_instance_type": {
			config: testDBClusterModifyPlanConfig(map[string]tftypes.Value{
				"db_instance_type": tftypes.NewValue(tftypes.String, nil),
			}),
			wantError: true,
		},
	})
}
