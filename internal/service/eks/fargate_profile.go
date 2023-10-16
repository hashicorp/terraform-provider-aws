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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
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

// @SDKResource("aws_eks_fargate_profile", name="Fargate Profile")
// @Tags(identifierAttribute="arn")
func ResourceFargateProfile() *schema.Resource {
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
	client := meta.(*conns.AWSClient).EKSClient(ctx)

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

	err := retry.RetryContext(ctx, propagationTimeout, func() *retry.RetryError {
		_, err := client.CreateFargateProfile(ctx, input)

		// Retry for IAM eventual consistency on error:
		// InvalidParameterException: Misconfigured PodExecutionRole Trust Policy; Please add the eks-fargate-pods.amazonaws.com Service Principal
		if errs.IsA[*types.InvalidParameterException](err) {
			if strings.Contains(err.Error(), "Misconfigured PodExecutionRole Trust Policy") {
				return retry.RetryableError(err)
			}
		}

		if err != nil {
			return retry.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		_, err = client.CreateFargateProfile(ctx, input)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EKS Fargate Profile (%s): %s", profileID, err)
	}

	d.SetId(profileID)

	waiter := eks.NewFargateProfileActiveWaiter(client)
	waiterParams := &eks.DescribeFargateProfileInput{
		ClusterName:        aws.String(clusterName),
		FargateProfileName: aws.String(fargateProfileName),
	}

	err = waiter.Wait(ctx, waiterParams, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EKS Fargate Profile (%s) to create: %s", d.Id(), err)
	}

	return append(diags, resourceFargateProfileRead(ctx, d, meta)...)
}

func resourceFargateProfileRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := meta.(*conns.AWSClient).EKSClient(ctx)

	clusterName, fargateProfileName, err := FargateProfileParseResourceID(d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EKS Fargate Profile (%s): %s", d.Id(), err)
	}

	fargateProfile, err := FindFargateProfileByClusterNameAndFargateProfileName(ctx, client, clusterName, fargateProfileName)

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

	if err := d.Set("subnet_ids", fargateProfile.Subnets); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting subnet_ids: %s", err)
	}

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
	client := meta.(*conns.AWSClient).EKSClient(ctx)

	clusterName, fargateProfileName, err := FargateProfileParseResourceID(d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting EKS Fargate Profile (%s): %s", d.Id(), err)
	}

	// mutex lock for creation/deletion serialization
	mutexKey := fmt.Sprintf("%s-fargate-profiles", d.Get("cluster_name").(string))
	conns.GlobalMutexKV.Lock(mutexKey)
	defer conns.GlobalMutexKV.Unlock(mutexKey)

	log.Printf("[DEBUG] Deleting EKS Fargate Profile: %s", d.Id())
	_, err = client.DeleteFargateProfile(ctx, &eks.DeleteFargateProfileInput{
		ClusterName:        aws.String(clusterName),
		FargateProfileName: aws.String(fargateProfileName),
	})

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting EKS Fargate Profile (%s): %s", d.Id(), err)
	}

	waiter := eks.NewFargateProfileDeletedWaiter(client)
	waiterParams := &eks.DescribeFargateProfileInput{
		ClusterName:        aws.String(clusterName),
		FargateProfileName: aws.String(fargateProfileName),
	}

	err = waiter.Wait(ctx, waiterParams, d.Timeout(schema.TimeoutDelete))
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting EKS Fargate Profile (%s): waiting for completion: %s", d.Id(), err)
	}

	return diags
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
			fargateProfileSelector.Labels = make(map[string]string)
			for key, value := range flex.ExpandStringMap(v) {
				val := value
				fargateProfileSelector.Labels[key] = *val
			}
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
			"namespace": fargateProfileSelector.Namespace,
		}

		l = append(l, m)
	}

	return l
}
