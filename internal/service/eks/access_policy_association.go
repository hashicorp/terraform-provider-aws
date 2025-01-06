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
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_eks_access_policy_association", name="Access Policy Association")
func resourceAccessPolicyAssociation() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceAccessPolicyAssociationCreate,
		ReadWithoutTimeout:   resourceAccessPolicyAssociationRead,
		DeleteWithoutTimeout: resourceAccessPolicyAssociationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

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

func resourceAccessPolicyAssociationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EKSClient(ctx)

	clusterName := d.Get(names.AttrClusterName).(string)
	principalARN := d.Get("principal_arn").(string)
	policyARN := d.Get("policy_arn").(string)
	id := accessPolicyAssociationCreateResourceID(clusterName, principalARN, policyARN)
	input := &eks.AssociateAccessPolicyInput{
		AccessScope:  expandAccessScope(d.Get("access_scope").([]interface{})),
		ClusterName:  aws.String(clusterName),
		PolicyArn:    aws.String(policyARN),
		PrincipalArn: aws.String(principalARN),
	}

	_, err := tfresource.RetryWhenIsAErrorMessageContains[*types.ResourceNotFoundException](ctx, propagationTimeout, func() (interface{}, error) {
		return conn.AssociateAccessPolicy(ctx, input)
	}, "The specified principalArn could not be found")

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EKS Access Policy Association (%s): %s", id, err)
	}

	d.SetId(id)

	return append(diags, resourceAccessPolicyAssociationRead(ctx, d, meta)...)
}

func resourceAccessPolicyAssociationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EKSClient(ctx)

	clusterName, principalARN, policyARN, err := accessPolicyAssociationParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	output, err := findAccessPolicyAssociationByThreePartKey(ctx, conn, clusterName, principalARN, policyARN)

	if !d.IsNewResource() && tfresource.NotFound(err) {
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

func resourceAccessPolicyAssociationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EKSClient(ctx)

	clusterName, principalARN, policyARN, err := accessPolicyAssociationParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[DEBUG] Deleting EKS Access Policy Association: %s", d.Id())
	_, err = conn.DisassociateAccessPolicy(ctx, &eks.DisassociateAccessPolicyInput{
		ClusterName:  aws.String(clusterName),
		PolicyArn:    aws.String(policyARN),
		PrincipalArn: aws.String(principalARN),
	})

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
	input := &eks.ListAssociatedAccessPoliciesInput{
		ClusterName:  aws.String(clusterName),
		PrincipalArn: aws.String(principalARN),
	}

	return findAssociatedAccessPolicy(ctx, conn, input, func(v *types.AssociatedAccessPolicy) bool {
		return aws.ToString(v.PolicyArn) == policyARN
	})
}

func findAssociatedAccessPolicy(ctx context.Context, conn *eks.Client, input *eks.ListAssociatedAccessPoliciesInput, filter tfslices.Predicate[*types.AssociatedAccessPolicy]) (*types.AssociatedAccessPolicy, error) {
	output, err := findAssociatedAccessPolicies(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findAssociatedAccessPolicies(ctx context.Context, conn *eks.Client, input *eks.ListAssociatedAccessPoliciesInput, filter tfslices.Predicate[*types.AssociatedAccessPolicy]) ([]types.AssociatedAccessPolicy, error) {
	var output []types.AssociatedAccessPolicy

	pages := eks.NewListAssociatedAccessPoliciesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*types.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		for _, v := range page.AssociatedAccessPolicies {
			if filter(&v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}

func expandAccessScope(l []interface{}) *types.AccessScope {
	if len(l) == 0 {
		return nil
	}

	m := l[0].(map[string]interface{})

	accessScope := &types.AccessScope{}

	if v, ok := m[names.AttrType].(string); ok && v != "" {
		accessScope.Type = types.AccessScopeType(v)
	}

	if v, ok := m["namespaces"]; ok {
		accessScope.Namespaces = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	return accessScope
}

func flattenAccessScope(apiObject *types.AccessScope) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		names.AttrType: (*string)(&apiObject.Type),
		"namespaces":   apiObject.Namespaces,
	}

	return []interface{}{tfMap}
}
