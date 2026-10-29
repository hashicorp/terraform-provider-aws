// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package eks

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	awstypes "github.com/aws/aws-sdk-go-v2/service/eks/types"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tfiter "github.com/hashicorp/terraform-provider-aws/internal/iter"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKListResource("aws_eks_access_policy_association")
func newAccessPolicyAssociationResourceAsListResource() inttypes.ListResourceForSDK {
	return framework.NewSDKListResource(resourceAccessPolicyAssociation(), framework.SDKListSpec[awstypes.AssociatedAccessPolicy, listAccessPolicyAssociationModel]{
		ConfigSchema: listAccessPolicyAssociationConfigSchema,
		Items: func(ctx context.Context, c *conns.AWSClient, query listAccessPolicyAssociationModel) (iter.Seq2[awstypes.AssociatedAccessPolicy, error], diag.Diagnostics) {
			input := eks.ListAssociatedAccessPoliciesInput{
				ClusterName:  aws.String(fwflex.StringValueFromFramework(ctx, query.ClusterName)),
				PrincipalArn: aws.String(fwflex.StringValueFromFramework(ctx, query.PrincipalARN)),
			}

			return listAssociatedAccessPolicies(ctx, c.EKSClient(ctx), &input), nil
		},
		ID: func(p awstypes.AssociatedAccessPolicy, query listAccessPolicyAssociationModel) string {
			return accessPolicyAssociationCreateResourceID(query.ClusterName.ValueString(), query.PrincipalARN.ValueString(), aws.ToString(p.PolicyArn))
		},
		Prepare: func(_ context.Context, p awstypes.AssociatedAccessPolicy, query listAccessPolicyAssociationModel, rd *schema.ResourceData) {
			// The cluster and principal come from the query, and all three make
			// up this resource's identity.
			rd.Set(names.AttrClusterName, query.ClusterName.ValueString())
			rd.Set("policy_arn", aws.ToString(p.PolicyArn))
			rd.Set("principal_arn", query.PrincipalARN.ValueString())
		},
		Flatten:     resourceAccessPolicyAssociationFlatten,
		DisplayName: func(p awstypes.AssociatedAccessPolicy) string { return aws.ToString(p.PolicyArn) },
	})
}

type listAccessPolicyAssociationModel struct {
	framework.WithRegionModel
	ClusterName  types.String `tfsdk:"cluster_name"`
	PrincipalARN fwtypes.ARN  `tfsdk:"principal_arn"`
}

func listAssociatedAccessPolicies(ctx context.Context, conn *eks.Client, input *eks.ListAssociatedAccessPoliciesInput, optFns ...func(*eks.Options)) iter.Seq2[awstypes.AssociatedAccessPolicy, error] {
	return tfiter.ConcatValuesWithError(listAssociatedAccessPolicyPages(ctx, conn, input, optFns...))
}

func listAssociatedAccessPolicyPages(ctx context.Context, conn *eks.Client, input *eks.ListAssociatedAccessPoliciesInput, optFns ...func(*eks.Options)) iter.Seq2[[]awstypes.AssociatedAccessPolicy, error] {
	return func(yield func([]awstypes.AssociatedAccessPolicy, error) bool) {
		pages := eks.NewListAssociatedAccessPoliciesPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx, optFns...)
			if err != nil {
				yield(nil, fmt.Errorf("listing EKS Access Policy Associations: %w", err))
				return
			}

			if !yield(page.AssociatedAccessPolicies, nil) {
				return
			}
		}
	}
}

func listAccessPolicyAssociationConfigSchema(_ context.Context) listschema.Schema {
	return listschema.Schema{
		Attributes: map[string]listschema.Attribute{
			names.AttrClusterName: listschema.StringAttribute{
				Required:    true,
				Description: "Name of the cluster to list access policy associations from.",
			},
			"principal_arn": listschema.StringAttribute{
				CustomType:  fwtypes.ARNType,
				Required:    true,
				Description: "ARN of the IAM principal for the access entry.",
			},
		},
	}
}
