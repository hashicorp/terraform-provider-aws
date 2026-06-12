// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package ssm

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_ssm_resource_policy", name="Resource Policy")
func newResourcePolicyDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &resourcePolicyDataSource{}, nil
}

type resourcePolicyDataSource struct {
	framework.DataSourceWithModel[resourcePolicyDataSourceModel]
}

func (d *resourcePolicyDataSource) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID: framework.IDAttribute(),
			names.AttrPolicy: schema.StringAttribute{
				CustomType: fwtypes.IAMPolicyType,
				Computed:   true,
			},
			"policy_hash": schema.StringAttribute{
				Computed: true,
			},
			"policy_id": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			names.AttrResourceARN: schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
			},
		},
	}
}

func (d *resourcePolicyDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data resourcePolicyDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().SSMClient(ctx)

	resourceARN := data.ResourceARN.ValueString()
	policyID := data.PolicyID.ValueString()

	var policy *awstypes.GetResourcePoliciesResponseEntry
	var err error
	if policyID != "" {
		policy, err = findResourcePolicyByTwoPartKey(ctx, conn, resourceARN, policyID)
	} else {
		policy, err = findResourcePolicyByARN(ctx, conn, resourceARN)
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading SSM Resource Policy (%s)", resourceARN), err.Error())

		return
	}

	data.Policy = fwtypes.IAMPolicyValue(aws.ToString(policy.Policy))
	data.PolicyHash = flex.StringToFramework(ctx, policy.PolicyHash)
	data.PolicyID = flex.StringToFramework(ctx, policy.PolicyId)
	data.ID = types.StringValue(resourcePolicyCreateID(resourceARN, aws.ToString(policy.PolicyId)))

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

// findResourcePolicyByARN returns the single resource policy attached to resourceARN.
// SSM resources (OpsItemGroup, advanced Parameter) currently support only one policy per resource;
// if multiple are returned, it errors so the caller can disambiguate with policy_id.
func findResourcePolicyByARN(ctx context.Context, conn *ssm.Client, resourceARN string) (*awstypes.GetResourcePoliciesResponseEntry, error) {
	input := &ssm.GetResourcePoliciesInput{
		ResourceArn: aws.String(resourceARN),
	}

	var policies []awstypes.GetResourcePoliciesResponseEntry
	pages := ssm.NewGetResourcePoliciesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.ResourceNotFoundException](err) || errs.IsA[*awstypes.ResourcePolicyNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError: err,
			}
		}

		if err != nil {
			return nil, err
		}

		policies = append(policies, page.Policies...)
	}

	switch len(policies) {
	case 0:
		return nil, &retry.NotFoundError{}
	case 1:
		return &policies[0], nil
	default:
		return nil, fmt.Errorf("multiple resource policies attached to %s; set policy_id to select one", resourceARN)
	}
}

type resourcePolicyDataSourceModel struct {
	framework.WithRegionModel
	ID          types.String      `tfsdk:"id"`
	Policy      fwtypes.IAMPolicy `tfsdk:"policy"`
	PolicyHash  types.String      `tfsdk:"policy_hash"`
	PolicyID    types.String      `tfsdk:"policy_id"`
	ResourceARN fwtypes.ARN       `tfsdk:"resource_arn"`
}
