// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package eks

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_eks_access_entry", name="Access Entry")
func dataSourceAccessEntry() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceAccessEntryRead,

		Schema: map[string]*schema.Schema{
			"access_entry_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrClusterName: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validClusterName,
			},
			names.AttrCreatedAt: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"kubernetes_groups": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"modified_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"principal_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
			names.AttrUserName: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrType: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourceAccessEntryRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EKSClient(ctx)

	clusterName := d.Get(names.AttrClusterName).(string)
	principalARN := d.Get("principal_arn").(string)
	id := accessEntryCreateResourceID(clusterName, principalARN)
	output, err := findAccessEntryByTwoPartKey(ctx, conn, clusterName, principalARN)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EKS Access Entry (%s): %s", id, err)
	}

	d.SetId(id)
	d.Set("access_entry_arn", output.AccessEntryArn)
	d.Set(names.AttrClusterName, output.ClusterName)
	d.Set(names.AttrCreatedAt, aws.ToTime(output.CreatedAt).Format(time.RFC3339))
	d.Set("kubernetes_groups", output.KubernetesGroups)
	d.Set("modified_at", aws.ToTime(output.ModifiedAt).Format(time.RFC3339))
	d.Set("principal_arn", output.PrincipalArn)
	d.Set(names.AttrType, output.Type)
	d.Set(names.AttrUserName, output.Username)

	setTagsOut(ctx, output.Tags)

	return diags
}
