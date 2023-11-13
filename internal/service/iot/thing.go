// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iot

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iot"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @SDKResource("aws_iot_thing")
func ResourceThing() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceThingCreate,
		ReadWithoutTimeout:   resourceThingRead,
		UpdateWithoutTimeout: resourceThingUpdate,
		DeleteWithoutTimeout: resourceThingDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"attributes": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"default_client_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
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
			"version": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func resourceThingCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IoTConn(ctx)

	name := d.Get("name").(string)
	input := &iot.CreateThingInput{
		ThingName: aws.String(name),
	}

	if v, ok := d.GetOk("attributes"); ok && len(v.(map[string]interface{})) > 0 {
		input.AttributePayload = &iot.AttributePayload{
			Attributes: flex.ExpandStringMap(v.(map[string]interface{})),
		}
	}

	if v, ok := d.GetOk("thing_type_name"); ok {
		input.ThingTypeName = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating IoT Thing: %s", input)
	output, err := conn.CreateThingWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating IoT Thing (%s): %s", name, err)
	}

	d.SetId(aws.StringValue(output.ThingName))

	return append(diags, resourceThingRead(ctx, d, meta)...)
}

func resourceThingRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IoTConn(ctx)

	output, err := FindThingByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] IoT Thing (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IoT Thing (%s): %s", d.Id(), err)
	}

	d.Set("arn", output.ThingArn)
	d.Set("default_client_id", output.DefaultClientId)
	d.Set("name", output.ThingName)
	d.Set("attributes", aws.StringValueMap(output.Attributes))
	d.Set("thing_type_name", output.ThingTypeName)
	d.Set("version", output.Version)

	return diags
}

func resourceThingUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IoTConn(ctx)

	input := &iot.UpdateThingInput{
		ThingName: aws.String(d.Get("name").(string)),
	}

	if d.HasChange("attributes") {
		attributes := map[string]*string{}

		if v, ok := d.GetOk("attributes"); ok && len(v.(map[string]interface{})) > 0 {
			attributes = flex.ExpandStringMap(v.(map[string]interface{}))
		}

		input.AttributePayload = &iot.AttributePayload{
			Attributes: attributes,
		}
	}

	if d.HasChange("thing_type_name") {
		if v, ok := d.GetOk("thing_type_name"); ok {
			input.ThingTypeName = aws.String(v.(string))
		} else {
			input.RemoveThingType = aws.Bool(true)
		}
	}

	log.Printf("[DEBUG] Updating IoT Thing: %s", input)
	_, err := conn.UpdateThingWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating IoT Thing (%s): %s", d.Id(), err)
	}

	return append(diags, resourceThingRead(ctx, d, meta)...)
}

func resourceThingDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IoTConn(ctx)

	log.Printf("[DEBUG] Deleting IoT Thing: %s", d.Id())
	_, err := conn.DeleteThingWithContext(ctx, &iot.DeleteThingInput{
		ThingName: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, iot.ErrCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting IoT Thing (%s): %s", d.Id(), err)
	}

	return diags
}
