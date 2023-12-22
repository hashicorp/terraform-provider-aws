// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package eks

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/aws/aws-sdk-go-v2/service/eks/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_eks_access_entry", name="Access Entry")
// @Tags(identifierAttribute="arn")
func ResourceAccessEntry() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceAccessEntryCreate,
		ReadWithoutTimeout:   resourceAccessEntryRead,
		UpdateWithoutTimeout: resourceAccessEntryUpdate,
		DeleteWithoutTimeout: resourceAccessEntryDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		CustomizeDiff: verify.SetTagsDiff,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"access_entry_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cluster_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validClusterName,
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"kubernetes_group": {
				Type:     schema.TypeSet,
				Optional: true,
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
				ForceNew:     true,
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

func resourceAccessEntryCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EKSClient(ctx)

	clusterName := d.Get("cluster_name").(string)
	principal_arn := d.Get("principal_arn").(string)
	accessID := AccessEntryCreateResourceID(clusterName, principal_arn)
	input := &eks.CreateAccessEntryInput{
		ClusterName:      aws.String(clusterName),
		PrincipalArn:     aws.String(principal_arn),
		KubernetesGroups: flex.ExpandStringValueSet(d.Get("kubernetes_groups").(*schema.Set)),
		Tags:             getTagsIn(ctx),
	}

	_, err := conn.CreateAccessEntry(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EKS Access Config: %s", err)
	}

	d.SetId(accessID)

	return append(diags, resourceAccessEntryRead(ctx, d, meta)...)
}

func resourceAccessEntryRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

func resourceAccessEntryUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EKSClient(ctx)
	clusterName, principal_arn, err := AccessEntryParseResourceID(d.Id())

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}
	if d.HasChangesExcept("tags", "tags_all") {
		input := &eks.UpdateAccessEntryInput{
			ClusterName:  aws.String(clusterName),
			PrincipalArn: aws.String(principal_arn),
		}

		if d.HasChange("kubernetes_group") {
			input.KubernetesGroups = flex.ExpandStringValueSet(d.Get("kubernetes_groups").(*schema.Set))
		}

		_, err := conn.UpdateAccessEntry(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Access Entry (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceFargateProfileRead(ctx, d, meta)...)
}

func resourceAccessEntryDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EKSClient(ctx)

	clusterName, principal_arn, err := AccessEntryParseResourceID(d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting EKS Access Entry (%s): %s", d.Id(), err)
	}

	log.Printf("[DEBUG] Deleting EKS Access Entry: %s", d.Id())
	_, err = conn.DeleteAccessEntry(ctx, &eks.DeleteAccessEntryInput{
		ClusterName:  aws.String(clusterName),
		PrincipalArn: aws.String(principal_arn),
	})

	if errs.IsAErrorMessageContains[*types.ResourceNotFoundException](err, "No Access Entry found for Id:") {
		return nil
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting EKS Access Entry (%s): %s", d.Id(), err)
	}

	return diags
}