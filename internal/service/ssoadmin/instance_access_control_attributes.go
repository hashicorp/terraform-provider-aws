// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssoadmin

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssoadmin"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ssoadmin/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_ssoadmin_instance_access_control_attributes")
func ResourceAccessControlAttributes() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceAccessControlAttributesCreate,
		ReadWithoutTimeout:   resourceAccessControlAttributesRead,
		UpdateWithoutTimeout: resourceAccessControlAttributesUpdate,
		DeleteWithoutTimeout: resourceAccessControlAttributesDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"attribute": {
				Type:     schema.TypeSet,
				Required: true,
				MinItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrKey: {
							Type:     schema.TypeString,
							Required: true,
						},
						names.AttrValue: {
							Type:     schema.TypeSet,
							Required: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrSource: {
										Type:     schema.TypeSet,
										Required: true,
										MinItems: 1,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
								},
							},
						},
					},
				},
			},
			"instance_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			names.AttrStatus: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrStatusReason: {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAccessControlAttributesCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SSOAdminClient(ctx)

	instanceARN := d.Get("instance_arn").(string)
	input := &ssoadmin.CreateInstanceAccessControlAttributeConfigurationInput{
		InstanceArn: aws.String(instanceARN),
		InstanceAccessControlAttributeConfiguration: &awstypes.InstanceAccessControlAttributeConfiguration{
			AccessControlAttributes: expandAccessControlAttributes(d),
		},
	}

	_, err := conn.CreateInstanceAccessControlAttributeConfiguration(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating SSO Instance Access Control Attributes (%s): %s", instanceARN, err)
	}

	d.SetId(instanceARN)

	return append(diags, resourceAccessControlAttributesRead(ctx, d, meta)...)
}

func resourceAccessControlAttributesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SSOAdminClient(ctx)

	output, err := FindInstanceAttributeControlAttributesByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] SSO Instance Access Control Attributes %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SSO Instance Access Control Attributes (%s): %s", d.Id(), err)
	}

	d.Set("instance_arn", d.Id())
	if err := d.Set("attribute", flattenAccessControlAttributes(output.InstanceAccessControlAttributeConfiguration.AccessControlAttributes)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting attribute: %s", err)
	}
	d.Set(names.AttrStatus, output.Status)
	d.Set(names.AttrStatusReason, output.StatusReason)

	return diags
}

func resourceAccessControlAttributesUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SSOAdminClient(ctx)

	input := &ssoadmin.UpdateInstanceAccessControlAttributeConfigurationInput{
		InstanceArn: aws.String(d.Id()),
		InstanceAccessControlAttributeConfiguration: &awstypes.InstanceAccessControlAttributeConfiguration{
			AccessControlAttributes: expandAccessControlAttributes(d),
		},
	}

	_, err := conn.UpdateInstanceAccessControlAttributeConfiguration(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating SSO Instance Access Control Attributes (%s): %s", d.Id(), err)
	}

	return append(diags, resourceAccessControlAttributesRead(ctx, d, meta)...)
}

func resourceAccessControlAttributesDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SSOAdminClient(ctx)

	_, err := conn.DeleteInstanceAccessControlAttributeConfiguration(ctx, &ssoadmin.DeleteInstanceAccessControlAttributeConfigurationInput{
		InstanceArn: aws.String(d.Id()),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting SSO Instance Access Control Attributes (%s): %s", d.Id(), err)
	}

	return diags
}

func FindInstanceAttributeControlAttributesByARN(ctx context.Context, conn *ssoadmin.Client, arn string) (*ssoadmin.DescribeInstanceAccessControlAttributeConfigurationOutput, error) {
	input := &ssoadmin.DescribeInstanceAccessControlAttributeConfigurationInput{
		InstanceArn: aws.String(arn),
	}

	output, err := conn.DescribeInstanceAccessControlAttributeConfiguration(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.InstanceAccessControlAttributeConfiguration == nil || len(output.InstanceAccessControlAttributeConfiguration.AccessControlAttributes) == 0 {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func expandAccessControlAttributes(d *schema.ResourceData) []awstypes.AccessControlAttribute {
	var attributes []awstypes.AccessControlAttribute

	attInterface := d.Get("attribute").(*schema.Set).List()
	for _, attrMap := range attInterface {
		attr := attrMap.(map[string]interface{})
		var attribute awstypes.AccessControlAttribute
		if key, ok := attr[names.AttrKey].(string); ok {
			attribute.Key = aws.String(key)
		}
		val := attr[names.AttrValue].(*schema.Set).List()[0].(map[string]interface{})
		if v, ok := val[names.AttrSource].(*schema.Set); ok && len(v.List()) > 0 {
			attribute.Value = &awstypes.AccessControlAttributeValue{
				Source: flex.ExpandStringValueSet(v),
			}
		}
		attributes = append(attributes, attribute)
	}

	return attributes
}

func flattenAccessControlAttributes(attributes []awstypes.AccessControlAttribute) []interface{} {
	var results []interface{}
	if len(attributes) == 0 {
		return []interface{}{}
	}

	for _, attr := range attributes {
		var val []interface{}
		val = append(val, map[string]interface{}{
			names.AttrSource: flex.FlattenStringValueSet(attr.Value.Source),
		})
		results = append(results, map[string]interface{}{
			names.AttrKey:   aws.ToString(attr.Key),
			names.AttrValue: val,
		})
	}

	return results
}
