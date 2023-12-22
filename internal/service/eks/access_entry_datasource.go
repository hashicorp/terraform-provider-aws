// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package eks

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_eks_access_entry")
func DataSourceAccessEntry() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceAccessEntryRead,

		Schema: map[string]*schema.Schema{
			"access_entry_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cluster_name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validClusterName,
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"kubernetes_group": {
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
			"user_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"type": {
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

	clusterName, principal_arn, err := AccessEntryParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EKS Access Entry (%s): %s", d.Id(), err)
	}
	output, err := FindAccessEntryByID(ctx, conn, clusterName, principal_arn)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EKS Access Entry (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EKS EKS Access Entry (%s): %s", d.Id(), err)
	}

	d.Set("access_entry_arn", output.AccessEntryArn)
	d.Set("cluster_name", output.ClusterName)
	d.Set("created_at", output.CreatedAt)
	// if err := d.Set("kubernetes_groups", aws.StringValueSlice(output.KubernetesGroups)); err != nil {
	// 	return sdkdiag.AppendErrorf(diags, "setting kubernetes_groups: %s", err)
	// }
	d.Set("kubernetes_groups", output.KubernetesGroups)
	d.Set("modified_at", output.ModifiedAt)
	d.Set("principal_arn", output.PrincipalArn)
	d.Set("user_name", output.Username)
	d.Set("type", output.Type)

	setTagsOut(ctx, output.Tags)

	return diags
}