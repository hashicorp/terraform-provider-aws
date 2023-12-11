// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package eks

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/aws/aws-sdk-go-v2/service/eks/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_eks_fargate_profile", name="Fargate Profile")
// @Tags(identifierAttribute="arn")
func resourceFargateProfile() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceFargateProfileCreate,
		ReadWithoutTimeout:   resourceFargateProfileRead,
		UpdateWithoutTimeout: resourceFargateProfileUpdate,
		DeleteWithoutTimeout: resourceFargateProfileDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		CustomizeDiff: verify.SetTagsDiff,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cluster_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validClusterName,
			},
			"fargate_profile_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			"pod_execution_role_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			"selector": {
				Type:     schema.TypeSet,
				Required: true,
				ForceNew: true,
				MinItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"labels": {
							Type:     schema.TypeMap,
							Optional: true,
							ForceNew: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"namespace": {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validation.NoZeroValues,
						},
					},
				},
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"subnet_ids": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				MinItems: 1,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func resourceFargateProfileCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EKSClient(ctx)

	clusterName := d.Get("cluster_name").(string)
	fargateProfileName := d.Get("fargate_profile_name").(string)
	profileID := FargateProfileCreateResourceID(clusterName, fargateProfileName)
	input := &eks.CreateFargateProfileInput{
		ClientRequestToken:  aws.String(id.UniqueId()),
		ClusterName:         aws.String(clusterName),
		FargateProfileName:  aws.String(fargateProfileName),
		PodExecutionRoleArn: aws.String(d.Get("pod_execution_role_arn").(string)),
		Selectors:           expandFargateProfileSelectors(d.Get("selector").(*schema.Set).List()),
		Subnets:             flex.ExpandStringValueSet(d.Get("subnet_ids").(*schema.Set)),
		Tags:                getTagsIn(ctx),
	}

	// mutex lock for creation/deletion serialization
	mutexKey := fmt.Sprintf("%s-fargate-profiles", clusterName)
	conns.GlobalMutexKV.Lock(mutexKey)
	defer conns.GlobalMutexKV.Unlock(mutexKey)

	// Retry for IAM eventual consistency on error:
	// InvalidParameterException: Misconfigured PodExecutionRole Trust Policy; Please add the eks-fargate-pods.amazonaws.com Service Principal
	_, err := tfresource.RetryWhenIsAErrorMessageContains[*types.InvalidParameterException](ctx, propagationTimeout, func() (interface{}, error) {
		return conn.CreateFargateProfile(ctx, input)
	}, "Misconfigured PodExecutionRole Trust Policy")

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EKS Fargate Profile (%s): %s", profileID, err)
	}

	d.SetId(profileID)

	if _, err := waitFargateProfileCreated(ctx, conn, clusterName, fargateProfileName, d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EKS Fargate Profile (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceFargateProfileRead(ctx, d, meta)...)
}

func resourceFargateProfileRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EKSClient(ctx)

	clusterName, fargateProfileName, err := FargateProfileParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	fargateProfile, err := findFargateProfileByTwoPartKey(ctx, conn, clusterName, fargateProfileName)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EKS Fargate Profile (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EKS Fargate Profile (%s): %s", d.Id(), err)
	}

	d.Set("arn", fargateProfile.FargateProfileArn)
	d.Set("cluster_name", fargateProfile.ClusterName)
	d.Set("fargate_profile_name", fargateProfile.FargateProfileName)
	d.Set("pod_execution_role_arn", fargateProfile.PodExecutionRoleArn)
	if err := d.Set("selector", flattenFargateProfileSelectors(fargateProfile.Selectors)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting selector: %s", err)
	}
	d.Set("status", fargateProfile.Status)
	d.Set("subnet_ids", fargateProfile.Subnets)

	setTagsOut(ctx, fargateProfile.Tags)

	return diags
}

func resourceFargateProfileUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	// Tags only.
	return append(diags, resourceFargateProfileRead(ctx, d, meta)...)
}

func resourceFargateProfileDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EKSClient(ctx)

	clusterName, fargateProfileName, err := FargateProfileParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	// mutex lock for creation/deletion serialization
	mutexKey := fmt.Sprintf("%s-fargate-profiles", d.Get("cluster_name").(string))
	conns.GlobalMutexKV.Lock(mutexKey)
	defer conns.GlobalMutexKV.Unlock(mutexKey)

	log.Printf("[DEBUG] Deleting EKS Fargate Profile: %s", d.Id())
	_, err = conn.DeleteFargateProfile(ctx, &eks.DeleteFargateProfileInput{
		ClusterName:        aws.String(clusterName),
		FargateProfileName: aws.String(fargateProfileName),
	})

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting EKS Fargate Profile (%s): %s", d.Id(), err)
	}

	if _, err := waitFargateProfileDeleted(ctx, conn, clusterName, fargateProfileName, d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EKS Fargate Profile (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func findFargateProfileByTwoPartKey(ctx context.Context, conn *eks.Client, clusterName, fargateProfileName string) (*types.FargateProfile, error) {
	input := &eks.DescribeFargateProfileInput{
		ClusterName:        aws.String(clusterName),
		FargateProfileName: aws.String(fargateProfileName),
	}

	output, err := conn.DescribeFargateProfile(ctx, input)

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.FargateProfile == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.FargateProfile, nil
}

func statusFargateProfile(ctx context.Context, conn *eks.Client, clusterName, fargateProfileName string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findFargateProfileByTwoPartKey(ctx, conn, clusterName, fargateProfileName)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func waitFargateProfileCreated(ctx context.Context, conn *eks.Client, clusterName, fargateProfileName string, timeout time.Duration) (*types.FargateProfile, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.FargateProfileStatusCreating),
		Target:  enum.Slice(types.FargateProfileStatusActive),
		Refresh: statusFargateProfile(ctx, conn, clusterName, fargateProfileName),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.FargateProfile); ok {
		return output, err
	}

	return nil, err
}

func waitFargateProfileDeleted(ctx context.Context, conn *eks.Client, clusterName, fargateProfileName string, timeout time.Duration) (*types.FargateProfile, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.FargateProfileStatusActive, types.FargateProfileStatusDeleting),
		Target:  []string{},
		Refresh: statusFargateProfile(ctx, conn, clusterName, fargateProfileName),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.FargateProfile); ok {
		return output, err
	}

	return nil, err
}

func expandFargateProfileSelectors(l []interface{}) []types.FargateProfileSelector {
	if len(l) == 0 {
		return nil
	}

	fargateProfileSelectors := make([]types.FargateProfileSelector, 0, len(l))

	for _, mRaw := range l {
		m, ok := mRaw.(map[string]interface{})

		if !ok {
			continue
		}

		fargateProfileSelector := types.FargateProfileSelector{}

		if v, ok := m["labels"].(map[string]interface{}); ok && len(v) > 0 {
			fargateProfileSelector.Labels = flex.ExpandStringValueMap(v)
		}

		if v, ok := m["namespace"].(string); ok && v != "" {
			fargateProfileSelector.Namespace = aws.String(v)
		}

		fargateProfileSelectors = append(fargateProfileSelectors, fargateProfileSelector)
	}

	return fargateProfileSelectors
}

func flattenFargateProfileSelectors(fargateProfileSelectors []types.FargateProfileSelector) []map[string]interface{} {
	if len(fargateProfileSelectors) == 0 {
		return []map[string]interface{}{}
	}

	l := make([]map[string]interface{}, 0, len(fargateProfileSelectors))

	for _, fargateProfileSelector := range fargateProfileSelectors {
		m := map[string]interface{}{
			"labels":    fargateProfileSelector.Labels,
			"namespace": aws.ToString(fargateProfileSelector.Namespace),
		}

		l = append(l, m)
	}

	return l
}
