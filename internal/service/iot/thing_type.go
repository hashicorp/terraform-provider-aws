// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iot

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iot"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_iot_thing_type", name="Thing Type")
// @Tags(identifierAttribute="arn")
func ResourceThingType() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceThingTypeCreate,
		ReadWithoutTimeout:   resourceThingTypeRead,
		UpdateWithoutTimeout: resourceThingTypeUpdate,
		DeleteWithoutTimeout: resourceThingTypeDelete,

		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				d.Set("name", d.Id())
				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validThingTypeName,
			},
			"properties": {
				Type:             schema.TypeList,
				Optional:         true,
				MaxItems:         1,
				DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"description": {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validThingTypeDescription,
						},
						"searchable_attributes": {
							Type:     schema.TypeSet,
							Optional: true,
							Computed: true,
							ForceNew: true,
							MaxItems: 3,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validThingTypeSearchableAttribute,
							},
						},
					},
				},
			},
			"deprecated": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceThingTypeCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IoTConn(ctx)

	input := &iot.CreateThingTypeInput{
		Tags:          getTagsIn(ctx),
		ThingTypeName: aws.String(d.Get("name").(string)),
	}

	if v, ok := d.GetOk("properties"); ok {
		configs := v.([]interface{})
		config, ok := configs[0].(map[string]interface{})

		if ok && config != nil {
			input.ThingTypeProperties = expandThingTypeProperties(config)
		}
	}

	out, err := conn.CreateThingTypeWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating IoT Thing Type (%s): %s", d.Get("name").(string), err)
	}

	d.SetId(aws.StringValue(out.ThingTypeName))

	if v := d.Get("deprecated").(bool); v {
		params := &iot.DeprecateThingTypeInput{
			ThingTypeName: aws.String(d.Id()),
			UndoDeprecate: aws.Bool(false),
		}

		_, err := conn.DeprecateThingTypeWithContext(ctx, params)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "creating IoT Thing Type (%s): deprecating Thing Type: %s", d.Get("name").(string), err)
		}
	}

	return append(diags, resourceThingTypeRead(ctx, d, meta)...)
}

func resourceThingTypeRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IoTConn(ctx)

	params := &iot.DescribeThingTypeInput{
		ThingTypeName: aws.String(d.Id()),
	}
	log.Printf("[DEBUG] Reading IoT Thing Type: %s", params)
	out, err := conn.DescribeThingTypeWithContext(ctx, params)

	if err != nil {
		if tfawserr.ErrCodeEquals(err, iot.ErrCodeResourceNotFoundException) {
			log.Printf("[WARN] IoT Thing Type (%s) not found, removing from state", d.Id())
			d.SetId("")
		}
		return sdkdiag.AppendErrorf(diags, "reading IoT Thing Type (%s): %s", d.Id(), err)
	}

	if out.ThingTypeMetadata != nil {
		d.Set("deprecated", out.ThingTypeMetadata.Deprecated)
	}

	if err := d.Set("properties", flattenThingTypeProperties(out.ThingTypeProperties)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting properties: %s", err)
	}

	d.Set("arn", out.ThingTypeArn)

	return diags
}

func resourceThingTypeUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IoTConn(ctx)

	if d.HasChange("deprecated") {
		params := &iot.DeprecateThingTypeInput{
			ThingTypeName: aws.String(d.Id()),
			UndoDeprecate: aws.Bool(!d.Get("deprecated").(bool)),
		}

		log.Printf("[DEBUG] Updating IoT Thing Type: %s", params)
		_, err := conn.DeprecateThingTypeWithContext(ctx, params)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating IoT Thing Type (%s): deprecating Thing Type: %s", d.Id(), err)
		}
	}

	return append(diags, resourceThingTypeRead(ctx, d, meta)...)
}

func resourceThingTypeDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IoTConn(ctx)

	// In order to delete an IoT Thing Type, you must deprecate it first and wait
	// at least 5 minutes.
	deprecateParams := &iot.DeprecateThingTypeInput{
		ThingTypeName: aws.String(d.Id()),
	}
	log.Printf("[DEBUG] Deprecating IoT Thing Type: %s", deprecateParams)
	_, err := conn.DeprecateThingTypeWithContext(ctx, deprecateParams)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting IoT Thing Type (%s): deprecating Thing Type: %s", d.Id(), err)
	}

	deleteParams := &iot.DeleteThingTypeInput{
		ThingTypeName: aws.String(d.Id()),
	}

	err = retry.RetryContext(ctx, 6*time.Minute, func() *retry.RetryError {
		_, err := conn.DeleteThingTypeWithContext(ctx, deleteParams)

		if err != nil {
			if tfawserr.ErrMessageContains(err, iot.ErrCodeInvalidRequestException, "Please wait for 5 minutes after deprecation and then retry") {
				return retry.RetryableError(err)
			}

			// As the delay post-deprecation is about 5 minutes, it may have been
			// deleted in between, thus getting a Not Found Exception.
			if tfawserr.ErrCodeEquals(err, iot.ErrCodeResourceNotFoundException) {
				return nil
			}

			return retry.NonRetryableError(err)
		}

		return nil
	})
	if tfresource.TimedOut(err) {
		_, err = conn.DeleteThingTypeWithContext(ctx, deleteParams)
		if tfawserr.ErrCodeEquals(err, iot.ErrCodeResourceNotFoundException) {
			return diags
		}
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting IoT Thing Type (%s): %s", d.Id(), err)
	}
	return diags
}
