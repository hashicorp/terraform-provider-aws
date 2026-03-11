// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package statecheck

import (
	"context"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	r "github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestStateValue_ValuesSame(t *testing.T) { //nolint:paralleltest // false positive
	ctx := acctest.Context(t)

	stateValue := StateValue()

	acctest.ParallelTest(ctx, t, r.TestCase{
		ProviderFactories: map[string]func() (*schema.Provider, error){
			"test": func() (*schema.Provider, error) { //nolint:unparam // required signature
				return testProvider(), nil
			},
		},
		Steps: []r.TestStep{
			{
				Config: `resource "test_resource" "one" {
  string_attribute = "same"
}
				`,
				ConfigStateChecks: []statecheck.StateCheck{
					stateValue.GetStateValue("test_resource.one", tfjsonpath.New("string_attribute")),
				},
			},
			{
				Config: `resource "test_resource" "one" {
  string_attribute = "same"
}
				`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("test_resource.one", tfjsonpath.New("string_attribute"), stateValue.Value()),
				},
			},
		},
	})
}

func TestStateValue_ValuesNotSame(t *testing.T) { //nolint:paralleltest // false positive
	ctx := acctest.Context(t)

	stateValue := StateValue()

	acctest.ParallelTest(ctx, t, r.TestCase{
		ProviderFactories: map[string]func() (*schema.Provider, error){
			"test": func() (*schema.Provider, error) { //nolint:unparam // required signature
				return testProvider(), nil
			},
		},
		Steps: []r.TestStep{
			{
				Config: `resource "test_resource" "one" {
  string_attribute = "same"
}
				`,
				ConfigStateChecks: []statecheck.StateCheck{
					stateValue.GetStateValue("test_resource.one", tfjsonpath.New("string_attribute")),
				},
			},
			{
				Config: `resource "test_resource" "one" {
  string_attribute = "not same"
}
				`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("test_resource.one", tfjsonpath.New("string_attribute"), stateValue.Value()),
				},
				ExpectError: regexache.MustCompile(`expected value same for StateValue check, got: not same`),
			},
		},
	})
}

func TestStateValue_NotInitialized(t *testing.T) { //nolint:paralleltest // false positive
	ctx := acctest.Context(t)

	stateValue := StateValue()

	acctest.ParallelTest(ctx, t, r.TestCase{
		ProviderFactories: map[string]func() (*schema.Provider, error){
			"test": func() (*schema.Provider, error) { //nolint:unparam // required signature
				return testProvider(), nil
			},
		},
		Steps: []r.TestStep{
			{
				Config: `resource "test_resource" "one" {
  string_attribute = "value"
}
				`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("test_resource.one", tfjsonpath.New("string_attribute"), stateValue.Value()),
				},
				ExpectError: regexache.MustCompile(`state value has not been set`),
			},
		},
	})
}

// Copied from https://github.com/hashicorp/terraform-plugin-testing/blob/main/statecheck/expect_known_value_test.go
func testProvider() *schema.Provider {
	return &schema.Provider{
		ResourcesMap: map[string]*schema.Resource{
			"test_resource": {
				CreateContext: func(_ context.Context, d *schema.ResourceData, _ any) diag.Diagnostics {
					d.SetId("test")

					err := d.Set("string_computed_attribute", "computed")
					if err != nil {
						return diag.Errorf("error setting string_computed_attribute: %s", err) // nosemgrep:ci.semgrep.errors.no-diag.Errorf-leading-error,ci.semgrep.pluginsdk.avoid-diag_Errorf
					}

					return nil
				},
				UpdateContext: func(_ context.Context, _ *schema.ResourceData, _ any) diag.Diagnostics {
					return nil
				},
				DeleteContext: func(_ context.Context, _ *schema.ResourceData, _ any) diag.Diagnostics {
					return nil
				},
				ReadContext: func(_ context.Context, _ *schema.ResourceData, _ any) diag.Diagnostics {
					return nil
				},
				Schema: map[string]*schema.Schema{
					"bool_attribute": {
						Optional: true,
						Type:     schema.TypeBool,
					},
					"float_attribute": {
						Optional: true,
						Type:     schema.TypeFloat,
					},
					"int_attribute": {
						Optional: true,
						Type:     schema.TypeInt,
					},
					"list_attribute": {
						Type: schema.TypeList,
						Elem: &schema.Schema{
							Type: schema.TypeString,
						},
						Optional: true,
					},
					"list_nested_block": {
						Type:     schema.TypeList,
						Optional: true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"list_nested_block_attribute": {
									Type:     schema.TypeString,
									Optional: true,
								},
							},
						},
					},
					"map_attribute": {
						Type: schema.TypeMap,
						Elem: &schema.Schema{
							Type: schema.TypeString,
						},
						Optional: true,
					},
					"set_attribute": {
						Type: schema.TypeSet,
						Elem: &schema.Schema{
							Type: schema.TypeString,
						},
						Optional: true,
					},
					"set_nested_block": {
						Type:     schema.TypeSet,
						Optional: true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"set_nested_block_attribute": {
									Type:     schema.TypeString,
									Optional: true,
								},
							},
						},
					},
					"set_nested_nested_block": {
						Type:     schema.TypeSet,
						Optional: true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"set_nested_block": {
									Type:     schema.TypeSet,
									Optional: true,
									Elem: &schema.Resource{
										Schema: map[string]*schema.Schema{
											"set_nested_block_attribute": {
												Type:     schema.TypeString,
												Optional: true,
											},
										},
									},
								},
							},
						},
					},
					"string_attribute": {
						Optional: true,
						Type:     schema.TypeString,
					},
					"string_computed_attribute": {
						Computed: true,
						Type:     schema.TypeString,
					},
				},
			},
		},
	}
}
