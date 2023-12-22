// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package eks

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// @SDKResource("aws_eks_access_policy_association", name="Access Policy Association")
func ResourceAccessPolicyAssociation() *schema.Resource {
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
			"associated_access_policy": {
				Type:     schema.TypeList,
				MinItems: 1,
				MaxItems: 1,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"associated_at": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"modified_at": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},

			"access_scope": {
				Type:     schema.TypeList,
				MinItems: 1,
				MaxItems: 1,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:     schema.TypeString,
							Required: true,
						},
						"namespaces": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},
			"cluster_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validClusterName,
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
	conn := meta.(*conns.AWSClient).EKSConn(ctx)

	clusterName := d.Get("cluster_name").(string)
	principal_arn := d.Get("principal_arn").(string)
	policy_arn := d.Get("policy_arn").(string)
	associateID := AssociatePolicyCreateResourceID(clusterName, principal_arn, policy_arn)
	input := &eks.AssociateAccessPolicyInput{
		ClusterName:  aws.String(clusterName),
		PrincipalArn: aws.String(principal_arn),
		PolicyArn:    aws.String(policy_arn),
		AccessScope:  expandAccessScope(d.Get("access_Scope").([]interface{})),
	}

	_, err := conn.AssociateAccessPolicyWithContext(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EKS Access Policy Association: %s", err)
	}

	d.SetId(associateID)

	return append(diags, resourceAccessPolicyAssociationRead(ctx, d, meta)...)
}

func resourceAccessPolicyAssociationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EKSConn(ctx)

	clusterName, principal_arn, policy_arn, err := AssociatePolicyParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EKS Associate Policy (%s): %s", d.Id(), err)
	}
	output, err := FindAccessPolicyByID(ctx, conn, clusterName, principal_arn, policy_arn)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EKS Access Policy Association (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EKS Access Policy Association (%s): %s", d.Id(), err)
	}

	if output == nil {
		if d.IsNewResource() {
			return sdkdiag.AppendErrorf(diags, "reading EKS Associated Policy Attachment (%s): not found after creation", d.Id())
		}

		log.Printf("[WARN] EKS Associated Policy Attachment (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	return diags
}

func resourceAccessPolicyAssociationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EKSConn(ctx)

	clusterName, principal_arn, policy_arn, err := AssociatePolicyParseResourceID(d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting EKS Policy Association (%s): %s", d.Id(), err)
	}

	log.Printf("[DEBUG] Deleting EKS Policy Associaion: %s", d.Id())

	input := &eks.DisassociateAccessPolicyInput{
		ClusterName:  aws.String(clusterName),
		PrincipalArn: aws.String(principal_arn),
		PolicyArn:    aws.String(policy_arn),
	}
	_, err = conn.DisassociateAccessPolicyWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, eks.ErrCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting EKS Policy Associattion (%s): %s", d.Id(), err)
	}

	return diags
}

func expandAccessScope(l []interface{}) *eks.AccessScope {
	if len(l) == 0 {
		return nil
	}

	m := l[0].(map[string]interface{})

	accessScope := &eks.AccessScope{
		Type: aws.String(m["type"].(string)),
	}

	if v, ok := m["namespaces"].(*schema.Set); ok && v.Len() > 0 {
		accessScope.Namespaces = flex.ExpandStringSet(v)
	}

	return accessScope
}

func FindAccessPolicyByID(ctx context.Context, conn *eks.EKS, clusterName string, principal_arn string, policy_arn string) (*eks.AssociatedAccessPolicy, error) {
	input := &eks.ListAssociatedAccessPoliciesInput{
		ClusterName:  aws.String(clusterName),
		PrincipalArn: aws.String(principal_arn),
	}

	var result *eks.AssociatedAccessPolicy

	err := conn.ListAssociatedAccessPoliciesPagesWithContext(ctx, input, func(page *eks.ListAssociatedAccessPoliciesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, attachedPolicy := range page.AssociatedAccessPolicies {
			if attachedPolicy == nil {
				continue
			}

			if aws.StringValue(attachedPolicy.PolicyArn) == policy_arn {
				result = attachedPolicy
				return false
			}
		}

		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}