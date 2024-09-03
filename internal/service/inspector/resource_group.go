// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package inspector

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/inspector"
	awstypes "github.com/aws/aws-sdk-go-v2/service/inspector/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_inspector_resource_group")
func ResourceResourceGroup() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceResourceGroupCreate,
		ReadWithoutTimeout:   resourceResourceGroupRead,
		DeleteWithoutTimeout: resourceResourceGroupDelete,

		Schema: map[string]*schema.Schema{
			names.AttrTags: {
				ForceNew: true,
				Required: true,
				Type:     schema.TypeMap,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceResourceGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).InspectorClient(ctx)

	req := &inspector.CreateResourceGroupInput{
		ResourceGroupTags: expandResourceGroupTags(d.Get(names.AttrTags).(map[string]interface{})),
	}
	log.Printf("[DEBUG] Creating Inspector Classic Resource Group: %#v", req)
	resp, err := conn.CreateResourceGroup(ctx, req)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Inspector Classic Resource Group: %s", err)
	}

	d.SetId(aws.ToString(resp.ResourceGroupArn))

	return append(diags, resourceResourceGroupRead(ctx, d, meta)...)
}

func resourceResourceGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).InspectorClient(ctx)

	resp, err := conn.DescribeResourceGroups(ctx, &inspector.DescribeResourceGroupsInput{
		ResourceGroupArns: []string{d.Id()},
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Inspector Classic Resource Group (%s): %s", d.Id(), err)
	}

	if len(resp.ResourceGroups) == 0 {
		if failedItem, ok := resp.FailedItems[d.Id()]; ok {
			if failedItem.FailureCode == awstypes.FailedItemErrorCodeItemDoesNotExist {
				log.Printf("[WARN] Inspector Classic Resource Group (%s) not found, removing from state", d.Id())
				d.SetId("")
				return diags
			}

			return sdkdiag.AppendErrorf(diags, "reading Inspector Classic Resource Group (%s): %s", d.Id(), string(failedItem.FailureCode))
		}

		return sdkdiag.AppendErrorf(diags, "reading Inspector Classic Resource Group (%s): %v", d.Id(), resp.FailedItems)
	}

	resourceGroup := resp.ResourceGroups[0]
	d.Set(names.AttrARN, resourceGroup.Arn)

	//lintignore:AWSR002
	if err := d.Set(names.AttrTags, flattenResourceGroupTags(resourceGroup.Tags)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	return diags
}

func resourceResourceGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	return diags
}

func expandResourceGroupTags(m map[string]interface{}) []awstypes.ResourceGroupTag {
	var result []awstypes.ResourceGroupTag

	for k, v := range m {
		result = append(result, awstypes.ResourceGroupTag{
			Key:   aws.String(k),
			Value: aws.String(v.(string)),
		})
	}

	return result
}

func flattenResourceGroupTags(tags []awstypes.ResourceGroupTag) map[string]interface{} {
	m := map[string]interface{}{}

	for _, tag := range tags {
		m[aws.ToString(tag.Key)] = aws.ToString(tag.Value)
	}

	return m
}
