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
	"github.com/hashicorp/terraform-plugin-framework/list"
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tfiter "github.com/hashicorp/terraform-provider-aws/internal/iter"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKListResource("aws_eks_access_policy_association")
func newAccessPolicyAssociationResourceAsListResource() inttypes.ListResourceForSDK {
	l := accessPolicyAssociationListResource{}
	l.SetResourceSchema(resourceAccessPolicyAssociation())
	return &l
}

var _ list.ListResource = &accessPolicyAssociationListResource{}

type accessPolicyAssociationListResource struct {
	framework.ListResourceWithSDKv2Resource
}

func (l *accessPolicyAssociationListResource) ListResourceConfigSchema(ctx context.Context, request list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = listschema.Schema{
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

func (l *accessPolicyAssociationListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().EKSClient(ctx)

	var query listAccessPolicyAssociationModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	clusterName, principalARN := fwflex.StringValueFromFramework(ctx, query.ClusterName), fwflex.StringValueFromFramework(ctx, query.PrincipalARN)

	tflog.Info(ctx, "Listing Resources", map[string]any{
		logging.ResourceAttributeKey(names.AttrClusterName): clusterName,
		logging.ResourceAttributeKey("principal_arn"):       principalARN,
	})

	stream.Results = func(yield func(list.ListResult) bool) {
		input := eks.ListAssociatedAccessPoliciesInput{
			ClusterName:  aws.String(clusterName),
			PrincipalArn: aws.String(principalARN),
		}
		for item, err := range listAssociatedAccessPolicies(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			result := request.NewListResult(ctx)

			policyARN := aws.ToString(item.PolicyArn)
			rd := l.ResourceData()
			rd.SetId(accessPolicyAssociationCreateResourceID(clusterName, principalARN, policyARN))
			rd.Set(names.AttrClusterName, clusterName)
			rd.Set("policy_arn", policyARN)
			rd.Set("principal_arn", principalARN)

			if request.IncludeResource {
				if err := resourceAccessPolicyAssociationFlatten(ctx, clusterName, principalARN, &item, rd); err != nil {
					tflog.Error(ctx, "Reading EKS Access Policy Association", map[string]any{
						"error": err.Error(),
					})
					continue
				}
			}

			result.DisplayName = policyARN

			l.SetResult(ctx, l.Meta(), request.IncludeResource, rd, &result)
			if result.Diagnostics.HasError() {
				yield(result)
				return
			}

			if !yield(result) {
				return
			}
		}
	}
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
