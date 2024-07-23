// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iot

import (
	"context"
	"log"
	"time"

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
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_iot_billing_group", name="Billing Group")
// @Tags(identifierAttribute="arn")
func resourceBillingGroup() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceBillingGroupCreate,
		ReadWithoutTimeout:   resourceBillingGroupRead,
		UpdateWithoutTimeout: resourceBillingGroupUpdate,
		DeleteWithoutTimeout: resourceBillingGroupDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"metadata": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrCreationDate: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			names.AttrName: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 128),
			},
			names.AttrProperties: {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrDescription: {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			names.AttrVersion: {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceBillingGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IoTClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &iot.CreateBillingGroupInput{
		BillingGroupName: aws.String(name),
		Tags:             getTagsIn(ctx),
	}

	if v, ok := d.GetOk(names.AttrProperties); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.BillingGroupProperties = expandBillingGroupProperties(v.([]interface{})[0].(map[string]interface{}))
	}

	output, err := conn.CreateBillingGroup(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating IoT Billing Group (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.BillingGroupName))

	return append(diags, resourceBillingGroupRead(ctx, d, meta)...)
}

func resourceBillingGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IoTClient(ctx)

	output, err := findBillingGroupByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] IoT Billing Group (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IoT Billing Group (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, output.BillingGroupArn)
	d.Set(names.AttrName, output.BillingGroupName)

	if output.BillingGroupMetadata != nil {
		if err := d.Set("metadata", []interface{}{flattenBillingGroupMetadata(output.BillingGroupMetadata)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting metadata: %s", err)
		}
	} else {
		d.Set("metadata", nil)
	}
	if v := flattenBillingGroupProperties(output.BillingGroupProperties); len(v) > 0 {
		if err := d.Set(names.AttrProperties, []interface{}{v}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting properties: %s", err)
		}
	} else {
		d.Set(names.AttrProperties, nil)
	}
	d.Set(names.AttrVersion, output.Version)

	return diags
}

func resourceBillingGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IoTClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := &iot.UpdateBillingGroupInput{
			BillingGroupName: aws.String(d.Id()),
			ExpectedVersion:  aws.Int64(int64(d.Get(names.AttrVersion).(int))),
		}

		if v, ok := d.GetOk(names.AttrProperties); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			input.BillingGroupProperties = expandBillingGroupProperties(v.([]interface{})[0].(map[string]interface{}))
		} else {
			input.BillingGroupProperties = &awstypes.BillingGroupProperties{}
		}

		_, err := conn.UpdateBillingGroup(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating IoT Billing Group (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceBillingGroupRead(ctx, d, meta)...)
}

func resourceBillingGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IoTClient(ctx)

	log.Printf("[DEBUG] Deleting IoT Billing Group: %s", d.Id())
	_, err := conn.DeleteBillingGroup(ctx, &iot.DeleteBillingGroupInput{
		BillingGroupName: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting IoT Billing Group (%s): %s", d.Id(), err)
	}

	return diags
}

func findBillingGroupByName(ctx context.Context, conn *iot.Client, name string) (*iot.DescribeBillingGroupOutput, error) {
	input := &iot.DescribeBillingGroupInput{
		BillingGroupName: aws.String(name),
	}

	output, err := conn.DescribeBillingGroup(ctx, input)

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

func expandBillingGroupProperties(tfMap map[string]interface{}) *awstypes.BillingGroupProperties {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.BillingGroupProperties{}

	if v, ok := tfMap[names.AttrDescription].(string); ok && v != "" {
		apiObject.BillingGroupDescription = aws.String(v)
	}

	return apiObject
}

func flattenBillingGroupMetadata(apiObject *awstypes.BillingGroupMetadata) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.CreationDate; v != nil {
		tfMap[names.AttrCreationDate] = aws.ToTime(v).Format(time.RFC3339)
	}

	return tfMap
}

func flattenBillingGroupProperties(apiObject *awstypes.BillingGroupProperties) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.BillingGroupDescription; v != nil {
		tfMap[names.AttrDescription] = aws.ToString(v)
	}

	return tfMap
}
