// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iot

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iot"
	awstypes "github.com/aws/aws-sdk-go-v2/service/iot/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_iot_thing", name="Thing")
func resourceThing() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceThingCreate,
		ReadWithoutTimeout:   resourceThingRead,
		UpdateWithoutTimeout: resourceThingUpdate,
		DeleteWithoutTimeout: resourceThingDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrAttributes: {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"default_client_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrName: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 128),
			},
			"thing_type_name": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 128),
			},
			names.AttrVersion: {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func resourceThingCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IoTClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &iot.CreateThingInput{
		ThingName: aws.String(name),
	}

	if v, ok := d.GetOk(names.AttrAttributes); ok && len(v.(map[string]interface{})) > 0 {
		input.AttributePayload = &awstypes.AttributePayload{
			Attributes: flex.ExpandStringValueMap(v.(map[string]interface{})),
		}
	}

	if v, ok := d.GetOk("thing_type_name"); ok {
		input.ThingTypeName = aws.String(v.(string))
	}

	output, err := conn.CreateThing(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating IoT Thing (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.ThingName))

	return append(diags, resourceThingRead(ctx, d, meta)...)
}

func resourceThingRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IoTClient(ctx)

	output, err := findThingByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] IoT Thing (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IoT Thing (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, output.ThingArn)
	d.Set("default_client_id", output.DefaultClientId)
	d.Set(names.AttrName, output.ThingName)
	d.Set(names.AttrAttributes, aws.StringMap(output.Attributes))
	d.Set("thing_type_name", output.ThingTypeName)
	d.Set(names.AttrVersion, output.Version)

	return diags
}

func resourceThingUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IoTClient(ctx)

	input := &iot.UpdateThingInput{
		ThingName: aws.String(d.Get(names.AttrName).(string)),
	}

	if d.HasChange(names.AttrAttributes) {
		attributes := map[string]string{}

		if v, ok := d.GetOk(names.AttrAttributes); ok && len(v.(map[string]interface{})) > 0 {
			attributes = flex.ExpandStringValueMap(v.(map[string]interface{}))
		}

		input.AttributePayload = &awstypes.AttributePayload{
			Attributes: attributes,
		}
	}

	if d.HasChange("thing_type_name") {
		if v, ok := d.GetOk("thing_type_name"); ok {
			input.ThingTypeName = aws.String(v.(string))
		} else {
			input.RemoveThingType = true
		}
	}

	_, err := conn.UpdateThing(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating IoT Thing (%s): %s", d.Id(), err)
	}

	return append(diags, resourceThingRead(ctx, d, meta)...)
}

func resourceThingDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IoTClient(ctx)

	log.Printf("[DEBUG] Deleting IoT Thing: %s", d.Id())
	_, err := conn.DeleteThing(ctx, &iot.DeleteThingInput{
		ThingName: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting IoT Thing (%s): %s", d.Id(), err)
	}

	return diags
}

func findThingByName(ctx context.Context, conn *iot.Client, name string) (*iot.DescribeThingOutput, error) {
	input := &iot.DescribeThingInput{
		ThingName: aws.String(name),
	}

	output, err := conn.DescribeThing(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}
