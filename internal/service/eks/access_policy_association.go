// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_eks_access_policy_association", name="Access Policy Association")
// @IdentityAttribute("cluster_name")
// @IdentityAttribute("principal_arn")
// @IdentityAttribute("policy_arn")
// @ImportIDHandler("accessPolicyAssociationImportID")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/eks/types;awstypes;awstypes.AssociatedAccessPolicy")
// @Testing(preIdentityVersion="v6.40.0")
func resourceAccessPolicyAssociation() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceAccessPolicyAssociationCreate,
		ReadWithoutTimeout:   resourceAccessPolicyAssociationRead,
		DeleteWithoutTimeout: resourceAccessPolicyAssociationDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"access_scope": {
				Type:     schema.TypeList,
				MinItems: 1,
				MaxItems: 1,
				Required: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"namespaces": {
							Type:     schema.TypeSet,
							Optional: true,
							ForceNew: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						names.AttrType: {
							Type:     schema.TypeString,
							ForceNew: true,
							Required: true,
						},
					},
				},
			},
			"associated_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrClusterName: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validClusterName,
			},
			"modified_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"policy_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"principal_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
		},
	}
}

func resourceAccessPolicyAssociationCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EKSClient(ctx)

	clusterName := d.Get(names.AttrClusterName).(string)
	principalARN := d.Get("principal_arn").(string)
	policyARN := d.Get("policy_arn").(string)
	id := accessPolicyAssociationCreateResourceID(clusterName, principalARN, policyARN)
	input := eks.AssociateAccessPolicyInput{
		AccessScope:  expandAccessScope(d.Get("access_scope").([]any)),
		ClusterName:  aws.String(clusterName),
		PolicyArn:    aws.String(policyARN),
		PrincipalArn: aws.String(principalARN),
	}

	_, err := tfresource.RetryWhenIsAErrorMessageContains[any, *types.ResourceNotFoundException](ctx, propagationTimeout, func(ctx context.Context) (any, error) {
		return conn.AssociateAccessPolicy(ctx, &input)
	}, "The specified principalArn could not be found")

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EKS Access Policy Association (%s): %s", id, err)
	}

	d.SetId(id)

	return append(diags, resourceAccessPolicyAssociationRead(ctx, d, meta)...)
}

func resourceAccessPolicyAssociationRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EKSClient(ctx)

	clusterName, principalARN, policyARN, err := accessPolicyAssociationParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	output, err := findAccessPolicyAssociationByThreePartKey(ctx, conn, clusterName, principalARN, policyARN)

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] EKS Access Policy Association (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EKS Access Policy Association (%s): %s", d.Id(), err)
	}

	d.Set("access_scope", flattenAccessScope(output.AccessScope))
	d.Set("associated_at", aws.ToTime(output.AssociatedAt).String())
	d.Set(names.AttrClusterName, clusterName)
	d.Set("modified_at", aws.ToTime(output.ModifiedAt).String())
	d.Set("policy_arn", policyARN)
	d.Set("principal_arn", principalARN)

	return diags
}

func resourceAccessPolicyAssociationDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EKSClient(ctx)

	clusterName, principalARN, policyARN, err := accessPolicyAssociationParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[DEBUG] Deleting EKS Access Policy Association: %s", d.Id())
	input := eks.DisassociateAccessPolicyInput{
		ClusterName:  aws.String(clusterName),
		PolicyArn:    aws.String(policyARN),
		PrincipalArn: aws.String(principalARN),
	}
	_, err = conn.DisassociateAccessPolicy(ctx, &input)

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting EKS Access Policy Association (%s): %s", d.Id(), err)
	}

	return diags
}

const accessPolicyAssociationResourceIDSeparator = "#"

func accessPolicyAssociationCreateResourceID(clusterName, principalARN, policyARN string) string {
	parts := []string{clusterName, principalARN, policyARN}
	id := strings.Join(parts, accessPolicyAssociationResourceIDSeparator)

	return id
}

func accessPolicyAssociationParseResourceID(id string) (string, string, string, error) {
	parts := strings.SplitN(id, accessPolicyAssociationResourceIDSeparator, 3)

	if len(parts) == 3 && parts[0] != "" && parts[1] != "" && parts[2] != "" {
		return parts[0], parts[1], parts[2], nil
	}

	return "", "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected cluster-name%[2]sprincipal-arn%[2]spolicy-arn", id, accessPolicyAssociationResourceIDSeparator)
}

func findAccessPolicyAssociationByThreePartKey(ctx context.Context, conn *eks.Client, clusterName, principalARN, policyARN string) (*types.AssociatedAccessPolicy, error) {
	input := eks.ListAssociatedAccessPoliciesInput{
		ClusterName:  aws.String(clusterName),
		PrincipalArn: aws.String(principalARN),
	}

	return findAssociatedAccessPolicy(ctx, conn, &input, func(v types.AssociatedAccessPolicy) bool {
		return aws.ToString(v.PolicyArn) == policyARN
	})
}

func findAssociatedAccessPolicy(ctx context.Context, conn *eks.Client, input *eks.ListAssociatedAccessPoliciesInput, filter tfslices.Predicate[types.AssociatedAccessPolicy]) (*types.AssociatedAccessPolicy, error) {
	output, err := findAssociatedAccessPolicies(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findAssociatedAccessPolicies(ctx context.Context, conn *eks.Client, input *eks.ListAssociatedAccessPoliciesInput, filter tfslices.Predicate[types.AssociatedAccessPolicy]) ([]types.AssociatedAccessPolicy, error) {
	var output []types.AssociatedAccessPolicy

	pages := eks.NewListAssociatedAccessPoliciesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*types.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError: err,
			}
		}

		if err != nil {
			return nil, err
		}

		for _, v := range page.AssociatedAccessPolicies {
			if filter(v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}

func expandAccessScope(tfList []any) *types.AccessScope {
	if len(tfList) == 0 {
		return nil
	}

	tfMap := tfList[0].(map[string]any)
	apiObject := &types.AccessScope{}

	if v, ok := tfMap["namespaces"]; ok {
		apiObject.Namespaces = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	if v, ok := tfMap[names.AttrType].(string); ok && v != "" {
		apiObject.Type = types.AccessScopeType(v)
	}

	return apiObject
}

func flattenAccessScope(apiObject *types.AccessScope) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{
		names.AttrType: (*string)(&apiObject.Type),
		"namespaces":   apiObject.Namespaces,
	}

	return []any{tfMap}
}

var (
	_ inttypes.SDKv2ImportID = accessPolicyAssociationImportID{}
)

type accessPolicyAssociationImportID struct{}

func (accessPolicyAssociationImportID) Parse(id string) (string, map[string]any, error) {
	clusterName, principalARN, policyARN, err := accessPolicyAssociationParseResourceID(id)
	if err != nil {
		return "", nil, err
	}

	result := map[string]any{
		names.AttrClusterName: clusterName,
		"policy_arn":          policyARN,
		"principal_arn":       principalARN,
	}

	return id, result, nil
}

func (accessPolicyAssociationImportID) Create(d *schema.ResourceData) string {
	return accessPolicyAssociationCreateResourceID(d.Get(names.AttrClusterName).(string), d.Get("principal_arn").(string), d.Get("policy_arn").(string))
}
