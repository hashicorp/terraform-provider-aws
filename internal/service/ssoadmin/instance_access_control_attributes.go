// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssoadmin

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssoadmin"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
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
						"key": {
							Type:     schema.TypeString,
							Required: true,
						},
						"value": {
							Type:     schema.TypeSet,
							Required: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"source": {
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
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status_reason": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAccessControlAttributesCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SSOAdminConn(ctx)

	instanceARN := d.Get("instance_arn").(string)
	input := &ssoadmin.CreateInstanceAccessControlAttributeConfigurationInput{
		InstanceArn: aws.String(instanceARN),
		InstanceAccessControlAttributeConfiguration: &ssoadmin.InstanceAccessControlAttributeConfiguration{
			AccessControlAttributes: expandAccessControlAttributes(d),
		},
	}

	_, err := conn.CreateInstanceAccessControlAttributeConfigurationWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating SSO Instance Access Control Attributes (%s): %s", instanceARN, err)
	}

	d.SetId(instanceARN)

	return append(diags, resourceAccessControlAttributesRead(ctx, d, meta)...)
}

func resourceAccessControlAttributesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SSOAdminConn(ctx)

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
	d.Set("status", output.Status)
	d.Set("status_reason", output.StatusReason)

	return diags
}

func resourceAccessControlAttributesUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SSOAdminConn(ctx)

	input := &ssoadmin.UpdateInstanceAccessControlAttributeConfigurationInput{
		InstanceArn: aws.String(d.Id()),
		InstanceAccessControlAttributeConfiguration: &ssoadmin.InstanceAccessControlAttributeConfiguration{
			AccessControlAttributes: expandAccessControlAttributes(d),
		},
	}

	_, err := conn.UpdateInstanceAccessControlAttributeConfigurationWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating SSO Instance Access Control Attributes (%s): %s", d.Id(), err)
	}

	return append(diags, resourceAccessControlAttributesRead(ctx, d, meta)...)
}

func resourceAccessControlAttributesDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SSOAdminConn(ctx)

	_, err := conn.DeleteInstanceAccessControlAttributeConfigurationWithContext(ctx, &ssoadmin.DeleteInstanceAccessControlAttributeConfigurationInput{
		InstanceArn: aws.String(d.Id()),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting SSO Instance Access Control Attributes (%s): %s", d.Id(), err)
	}

	return diags
}

func FindInstanceAttributeControlAttributesByARN(ctx context.Context, conn *ssoadmin.SSOAdmin, arn string) (*ssoadmin.DescribeInstanceAccessControlAttributeConfigurationOutput, error) {
	input := &ssoadmin.DescribeInstanceAccessControlAttributeConfigurationInput{
		InstanceArn: aws.String(arn),
	}

	output, err := conn.DescribeInstanceAccessControlAttributeConfigurationWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, ssoadmin.ErrCodeResourceNotFoundException) {
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

func expandAccessControlAttributes(d *schema.ResourceData) (attributes []*ssoadmin.AccessControlAttribute) {
	attInterface := d.Get("attribute").(*schema.Set).List()
	for _, attrMap := range attInterface {
		attr := attrMap.(map[string]interface{})
		var attribute ssoadmin.AccessControlAttribute
		if key, ok := attr["key"].(string); ok {
			attribute.Key = aws.String(key)
		}
		val := attr["value"].(*schema.Set).List()[0].(map[string]interface{})
		if v, ok := val["source"].(*schema.Set); ok && len(v.List()) > 0 {
			attribute.Value = &ssoadmin.AccessControlAttributeValue{
				Source: flex.ExpandStringSet(v),
			}
		}
		attributes = append(attributes, &attribute)
	}
	return
}

func flattenAccessControlAttributes(attributes []*ssoadmin.AccessControlAttribute) []interface{} {
	var results []interface{}
	if len(attributes) == 0 {
		return []interface{}{}
	}
	for _, attr := range attributes {
		if attr == nil {
			continue
		}
		var val []interface{}
		val = append(val, map[string]interface{}{
			"source": flex.FlattenStringSet(attr.Value.Source),
		})
		results = append(results, map[string]interface{}{
			"key":   aws.StringValue(attr.Key),
			"value": val,
		})
	}
	return results
}
