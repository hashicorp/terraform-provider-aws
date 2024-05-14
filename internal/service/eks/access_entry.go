// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package eks

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/aws/aws-sdk-go-v2/service/eks/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
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
// @Tags(identifierAttribute="access_entry_arn")
func resourceAccessEntry() *schema.Resource {
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
			names.AttrClusterName: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validClusterName,
			},
			names.AttrCreatedAt: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"kubernetes_groups": {
				Type:     schema.TypeSet,
				Optional: true,
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
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			names.AttrType: {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      accessEntryTypeStandard,
				ValidateFunc: validation.StringInSlice(accessEntryType_Values(), false),
			},
			names.AttrUserName: {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
		},
	}
}

func resourceAccessEntryCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EKSClient(ctx)

	clusterName := d.Get(names.AttrClusterName).(string)
	principalARN := d.Get("principal_arn").(string)
	id := accessEntryCreateResourceID(clusterName, principalARN)
	input := &eks.CreateAccessEntryInput{
		ClusterName:  aws.String(clusterName),
		PrincipalArn: aws.String(principalARN),
		Tags:         getTagsIn(ctx),
		Type:         aws.String(d.Get(names.AttrType).(string)),
	}

	if v, ok := d.GetOk("kubernetes_groups"); ok {
		input.KubernetesGroups = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk(names.AttrUserName); ok {
		input.Username = aws.String(v.(string))
	}

	_, err := tfresource.RetryWhenIsAErrorMessageContains[*types.InvalidParameterException](ctx, propagationTimeout, func() (interface{}, error) {
		return conn.CreateAccessEntry(ctx, input)
	}, "The specified principalArn is invalid: invalid principal")

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EKS Access Entry (%s): %s", id, err)
	}

	d.SetId(id)

	return append(diags, resourceAccessEntryRead(ctx, d, meta)...)
}

func resourceAccessEntryRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EKSClient(ctx)

	clusterName, principalARN, err := accessEntryParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	output, err := findAccessEntryByTwoPartKey(ctx, conn, clusterName, principalARN)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EKS Access Entry (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EKS Access Entry (%s): %s", d.Id(), err)
	}

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

func resourceAccessEntryUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EKSClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		clusterName, principalARN, err := accessEntryParseResourceID(d.Id())
		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}

		input := &eks.UpdateAccessEntryInput{
			ClusterName:  aws.String(clusterName),
			PrincipalArn: aws.String(principalARN),
		}

		input.KubernetesGroups = flex.ExpandStringValueSet(d.Get("kubernetes_groups").(*schema.Set))
		input.Username = aws.String(d.Get(names.AttrUserName).(string))

		_, err = conn.UpdateAccessEntry(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating EKS Access Entry (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceAccessEntryRead(ctx, d, meta)...)
}

func resourceAccessEntryDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EKSClient(ctx)

	clusterName, principalARN, err := accessEntryParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[DEBUG] Deleting EKS Access Entry: %s", d.Id())
	_, err = conn.DeleteAccessEntry(ctx, &eks.DeleteAccessEntryInput{
		ClusterName:  aws.String(clusterName),
		PrincipalArn: aws.String(principalARN),
	})

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting EKS Access Entry (%s): %s", d.Id(), err)
	}

	return diags
}

const accessEntryResourceIDSeparator = ":"

func accessEntryCreateResourceID(clusterName, principal_arn string) string {
	parts := []string{clusterName, principal_arn}
	id := strings.Join(parts, accessEntryResourceIDSeparator)

	return id
}

func accessEntryParseResourceID(id string) (string, string, error) {
	parts := strings.SplitN(id, accessEntryResourceIDSeparator, 2)

	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected cluster-name%[2]sprincipal-arn", id, accessEntryResourceIDSeparator)
}

func findAccessEntryByTwoPartKey(ctx context.Context, conn *eks.Client, clusterName, principalARN string) (*types.AccessEntry, error) {
	input := &eks.DescribeAccessEntryInput{
		ClusterName:  aws.String(clusterName),
		PrincipalArn: aws.String(principalARN),
	}

	output, err := conn.DescribeAccessEntry(ctx, input)

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.AccessEntry == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.AccessEntry, nil
}
