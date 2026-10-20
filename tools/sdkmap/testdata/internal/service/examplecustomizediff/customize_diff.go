// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package examplecustomizediff

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

// @SDKResource("aws_examplecustomizediff_identifier", name="Identifier")
func resourceIdentifier() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceThingCreate,
		ReadWithoutTimeout:   resourceThingRead,
		UpdateWithoutTimeout: resourceThingUpdate,
		DeleteWithoutTimeout: resourceThingDelete,

		CustomizeDiff: customizeDiffChecksum,

		SchemaFunc: thingSchema,
	}
}

// @SDKResource("aws_examplecustomizediff_sequence", name="Sequence")
func resourceSequence() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceThingCreate,
		ReadWithoutTimeout:   resourceThingRead,
		UpdateWithoutTimeout: resourceThingUpdate,
		DeleteWithoutTimeout: resourceThingDelete,

		CustomizeDiff: customdiff.Sequence(
			customizeDiffChecksum,
			validateSize,
		),

		SchemaFunc: thingSchema,
	}
}

// @SDKResource("aws_examplecustomizediff_nested", name="Nested")
func resourceNested() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceThingCreate,
		ReadWithoutTimeout:   resourceThingRead,
		UpdateWithoutTimeout: resourceThingUpdate,
		DeleteWithoutTimeout: resourceThingDelete,

		CustomizeDiff: customdiff.All(
			customdiff.Sequence(
				validateSize,
				validateParentExists,
			),
			customdiff.ForceNewIfChange("size", sizeDecreased),
			func(ctx context.Context, d *schema.ResourceDiff, meta any) error {
				if !d.HasChange("size") {
					return nil
				}

				quota, err := findQuotaByName(ctx, meta.(*conns.AWSClient).ExampleCustomizeDiffClient(ctx), "size")

				if err != nil {
					return err
				}

				if v := d.Get("size").(int); int32(v) > quota.Value {
					return fmt.Errorf("size (%d) exceeds quota (%d)", v, quota.Value)
				}

				return nil
			},
		),

		SchemaFunc: thingSchema,
	}
}

// @SDKResource("aws_examplecustomizediff_function_literal", name="Function Literal")
func resourceFunctionLiteral() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceThingCreate,
		ReadWithoutTimeout:   resourceThingRead,
		UpdateWithoutTimeout: resourceThingUpdate,
		DeleteWithoutTimeout: resourceThingDelete,

		CustomizeDiff: func(_ context.Context, d *schema.ResourceDiff, meta any) error {
			if d.HasChange("content") && d.Get("size").(int) == 0 {
				return errors.New("size must be set when content changes")
			}

			return nil
		},

		SchemaFunc: thingSchema,
	}
}

// @SDKResource("aws_examplecustomizediff_force_new_if", name="Force New If")
func resourceForceNewIf() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceThingCreate,
		ReadWithoutTimeout:   resourceThingRead,
		UpdateWithoutTimeout: resourceThingUpdate,
		DeleteWithoutTimeout: resourceThingDelete,

		CustomizeDiff: customdiff.ForceNewIf("parent_id", func(_ context.Context, d *schema.ResourceDiff, meta any) bool {
			return d.HasChange("size")
		}),

		SchemaFunc: thingSchema,
	}
}

// @SDKResource("aws_examplecustomizediff_variable", name="Variable")
func resourceVariable() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceThingCreate,
		ReadWithoutTimeout:   resourceThingRead,
		UpdateWithoutTimeout: resourceThingUpdate,
		DeleteWithoutTimeout: resourceThingDelete,

		CustomizeDiff: thingCustomizeDiff,

		SchemaFunc: thingSchema,
	}
}

// @SDKResource("aws_examplecustomizediff_call", name="Call")
func resourceCall() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceThingCreate,
		ReadWithoutTimeout:   resourceThingRead,
		UpdateWithoutTimeout: resourceThingUpdate,
		DeleteWithoutTimeout: resourceThingDelete,

		CustomizeDiff: customizeDiffValidateQuota("size"),

		SchemaFunc: thingSchema,
	}
}

// @SDKResource("aws_examplecustomizediff_factory_reference", name="Factory Reference")
func resourceFactoryReference() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceThingCreate,
		ReadWithoutTimeout:   resourceThingRead,
		UpdateWithoutTimeout: resourceThingUpdate,
		DeleteWithoutTimeout: resourceThingDelete,

		CustomizeDiff: customizeDiffComputedOnUpdate,

		SchemaFunc: thingSchema,
	}
}

var thingCustomizeDiff = customdiff.All(
	customizeDiffChecksum,
	validateParentExists,
)

func customizeDiffChecksum(_ context.Context, d *schema.ResourceDiff, meta any) error {
	if d.HasChange("content") {
		return d.SetNewComputed("checksum")
	}

	return nil
}

func validateSize(_ context.Context, d *schema.ResourceDiff, meta any) error {
	if v := d.Get("size").(int); v%2 != 0 {
		return errors.New("size must be even")
	}

	return nil
}

func validateParentExists(ctx context.Context, d *schema.ResourceDiff, meta any) error {
	parentID, ok := d.GetOk("parent_id")
	if !ok || d.Id() != "" {
		return nil
	}

	_, err := findParentByID(ctx, meta.(*conns.AWSClient).ExampleCustomizeDiffClient(ctx), parentID.(string))

	return err
}

func sizeDecreased(_ context.Context, old, new, meta any) bool {
	return new.(int) < old.(int)
}

func customizeDiffValidateQuota(key string) schema.CustomizeDiffFunc {
	return func(ctx context.Context, d *schema.ResourceDiff, meta any) error {
		if !d.HasChange(key) {
			return nil
		}

		quota, err := findQuotaByName(ctx, meta.(*conns.AWSClient).ExampleCustomizeDiffClient(ctx), key)

		if err != nil {
			return err
		}

		if v := d.Get(key).(int); int32(v) > quota.Value {
			return fmt.Errorf("%s (%d) exceeds quota (%d)", key, v, quota.Value)
		}

		return nil
	}
}

func customizeDiffComputedOnUpdate(_ context.Context, d *schema.ResourceDiff, meta any) error {
	if d.Id() != "" && thingNeedsUpdate(d) {
		return d.SetNewComputed("checksum")
	}

	return nil
}

func thingNeedsUpdate(d *schema.ResourceDiff) bool {
	for k, attr := range resourceFactoryReference().SchemaMap() {
		if attr.ForceNew || (attr.Computed && !attr.Optional) {
			continue
		}

		if d.HasChange(k) {
			return true
		}
	}

	return false
}
