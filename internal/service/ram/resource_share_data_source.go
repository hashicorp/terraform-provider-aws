// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ram

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ram"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ram/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_ram_resource_share", name="Resource Share")
// @Tags
func dataSourceResourceShare() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceResourceShareRead,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrFilter: {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrName: {
							Type:     schema.TypeString,
							Required: true,
						},
						names.AttrValues: {
							Type:     schema.TypeList,
							Required: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"owning_account_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"resource_arns": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			names.AttrResourceOwner: {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: enum.Validate[awstypes.ResourceOwner](),
			},
			"resource_share_status": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: enum.Validate[awstypes.ResourceShareStatus](),
			},
			names.AttrStatus: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourceResourceShareRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RAMClient(ctx)

	resourceOwner := d.Get(names.AttrResourceOwner).(string)
	inputG := &ram.GetResourceSharesInput{
		ResourceOwner: awstypes.ResourceOwner(resourceOwner),
	}

	if v, ok := d.GetOk(names.AttrName); ok {
		inputG.Name = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrFilter); ok && v.(*schema.Set).Len() > 0 {
		inputG.TagFilters = expandTagFilters(v.(*schema.Set).List())
	}

	if v, ok := d.GetOk("resource_share_status"); ok {
		inputG.ResourceShareStatus = awstypes.ResourceShareStatus(v.(string))
	}

	share, err := findResourceShare(ctx, conn, inputG)

	if err != nil {
		return sdkdiag.AppendFromErr(diags, tfresource.SingularDataSourceFindError("RAM Resource Share", err))
	}

	arn := aws.ToString(share.ResourceShareArn)
	d.SetId(arn)
	d.Set(names.AttrARN, arn)
	d.Set(names.AttrName, share.Name)
	d.Set("owning_account_id", share.OwningAccountId)
	d.Set(names.AttrStatus, share.Status)

	setTagsOut(ctx, share.Tags)

	inputL := &ram.ListResourcesInput{
		ResourceOwner:     awstypes.ResourceOwner(resourceOwner),
		ResourceShareArns: []string{arn},
	}
	resources, err := findResources(ctx, conn, inputL)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading RAM Resource Share (%s) resources: %s", arn, err)
	}

	resourceARNs := tfslices.ApplyToAll(resources, func(r awstypes.Resource) string {
		return aws.ToString(r.Arn)
	})
	d.Set("resource_arns", resourceARNs)

	return diags
}

func expandTagFilter(tfMap map[string]interface{}) awstypes.TagFilter {
	apiObject := awstypes.TagFilter{}

	if v, ok := tfMap[names.AttrName].(string); ok && v != "" {
		apiObject.TagKey = aws.String(v)
	}

	if v, ok := tfMap[names.AttrValues].([]interface{}); ok && len(v) > 0 {
		apiObject.TagValues = flex.ExpandStringValueList(v)
	}

	return apiObject
}

func expandTagFilters(tfList []interface{}) []awstypes.TagFilter {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.TagFilter

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObjects = append(apiObjects, expandTagFilter(tfMap))
	}

	return apiObjects
}
