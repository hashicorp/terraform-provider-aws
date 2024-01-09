// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package eks

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
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
func dataSourceAccessEntry() *schema.Resource {
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

	principalArn := d.Get("principal_arn").(string)
	clusterName := d.Get("cluster_name").(string)
	id := AccessEntryCreateResourceID(clusterName, principalArn)
	output, err := FindAccessEntryByID(ctx, conn, clusterName, principalArn)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EKS Access Entry (%s) not found, removing from state", id)
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EKS Access Entry (%s): %s", id, err)
	}
	d.SetId(id)
	d.Set("access_entry_arn", output.AccessEntryArn)
	d.Set("cluster_name", output.ClusterName)
	d.Set("created_at", aws.ToTime(output.CreatedAt).String())
	d.Set("kubernetes_groups", output.KubernetesGroups)
	d.Set("modified_at", aws.ToTime(output.ModifiedAt).String())
	d.Set("principal_arn", output.PrincipalArn)
	d.Set("user_name", output.Username)
	d.Set("type", output.Type)

	setTagsOut(ctx, output.Tags)

	return diags
}
