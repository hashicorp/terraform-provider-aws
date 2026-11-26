// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ecs

import (
	"context"
	"strings"
	"testing"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestTaskDefinitionTrustedExecutionConfigurationDiff(t *testing.T) {
	t.Parallel()

	const containerDefinitions = `[{"name":"app","image":"busybox","essential":true}]`

	r := resourceTaskDefinition()
	schemaType := r.CoreConfigSchema().ImpliedType()
	blockType := schemaType.AttributeType("trusted_execution_configuration")
	blockElemType := blockType.ElementType()

	block := func(isolationMode string) cty.Value {
		return cty.ListVal([]cty.Value{cty.ObjectVal(map[string]cty.Value{
			"attestation_policy": cty.NullVal(cty.String),
			"isolation_mode":     cty.StringVal(isolationMode),
		})})
	}

	config := func(tec cty.Value) cty.Value {
		attrs := make(map[string]cty.Value)
		for name, ty := range schemaType.AttributeTypes() {
			switch {
			case ty.IsListType():
				attrs[name] = cty.ListValEmpty(ty.ElementType())
			case ty.IsSetType():
				attrs[name] = cty.SetValEmpty(ty.ElementType())
			default:
				attrs[name] = cty.NullVal(ty)
			}
		}
		attrs["container_definitions"] = cty.StringVal(containerDefinitions)
		attrs[names.AttrFamily] = cty.StringVal("test")
		attrs["trusted_execution_configuration"] = tec

		return cty.ObjectVal(attrs)
	}

	state := func(isolationMode string) *terraform.InstanceState {
		attrs := map[string]string{
			"container_definitions":             containerDefinitions,
			names.AttrFamily:                    "test",
			"trusted_execution_configuration.#": "0",
		}
		if isolationMode != "" {
			attrs["trusted_execution_configuration.#"] = "1"
			attrs["trusted_execution_configuration.0.attestation_policy"] = ""
			attrs["trusted_execution_configuration.0.isolation_mode"] = isolationMode
		}

		return &terraform.InstanceState{ID: "test", Attributes: attrs}
	}

	testCases := map[string]struct {
		state           string
		config          cty.Value
		wantRequiresNew bool
	}{
		"absent to SHARED": {
			config: block("SHARED"),
		},
		"SHARED to absent": {
			state:  "SHARED",
			config: cty.ListValEmpty(blockElemType),
		},
		"absent to ISOLATED_SANDBOX": {
			config:          block("ISOLATED_SANDBOX"),
			wantRequiresNew: true,
		},
		"SHARED to ISOLATED_SANDBOX": {
			state:           "SHARED",
			config:          block("ISOLATED_SANDBOX"),
			wantRequiresNew: true,
		},
		"ISOLATED_SANDBOX to absent": {
			state:           "ISOLATED_SANDBOX",
			config:          cty.ListValEmpty(blockElemType),
			wantRequiresNew: true,
		},
		"absent to unknown": {
			config:          cty.UnknownVal(blockType),
			wantRequiresNew: true,
		},
		"SHARED to unknown": {
			state:           "SHARED",
			config:          cty.UnknownVal(blockType),
			wantRequiresNew: true,
		},
		"SHARED to unknown isolation mode": {
			state: "SHARED",
			config: cty.ListVal([]cty.Value{cty.ObjectVal(map[string]cty.Value{
				"attestation_policy": cty.NullVal(cty.String),
				"isolation_mode":     cty.UnknownVal(cty.String),
			})}),
			wantRequiresNew: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// As in PlanResourceChange, the raw configuration is passed on the prior state.
			configVal := config(tc.config)
			s := state(tc.state)
			s.RawConfig = configVal

			diff, err := r.SimpleDiff(context.Background(), s, terraform.NewResourceConfigShimmed(configVal, r.CoreConfigSchema()), nil)
			if err != nil {
				t.Fatal(err)
			}

			var gotDiff, gotRequiresNew bool
			if diff != nil {
				for k, v := range diff.Attributes {
					if !strings.HasPrefix(k, "trusted_execution_configuration") {
						continue
					}
					gotDiff = true
					if v.RequiresNew {
						gotRequiresNew = true
					}
				}
			}

			if gotRequiresNew != tc.wantRequiresNew {
				t.Errorf("trusted_execution_configuration RequiresNew = %t, want %t (diff: %t)", gotRequiresNew, tc.wantRequiresNew, gotDiff)
			}
			if !tc.wantRequiresNew && gotDiff {
				t.Errorf("unexpected trusted_execution_configuration diff: %v", diff.Attributes)
			}
		})
	}
}
