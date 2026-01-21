// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package cloudfront

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/datasourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_cloudfront_distribution_tenant", name="Distribution Tenant")
// @Tags(identifierAttribute="arn")
func newDistributionTenantDataSource(_ context.Context) (datasource.DataSourceWithConfigure, error) {
	d := &distributionTenantDataSource{}
	return d, nil
}

type distributionTenantDataSource struct {
	framework.DataSourceWithModel[distributionTenantDataSourceModel]
}

func (d *distributionTenantDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"connection_group_id": schema.StringAttribute{
				Computed: true,
			},
			"customizations": framework.DataSourceComputedListOfObjectAttribute[customizationsModel](ctx),
			"distribution_id": schema.StringAttribute{
				Computed: true,
			},
			names.AttrDomain: schema.StringAttribute{
				Optional: true,
			},
			"domains": framework.DataSourceComputedListOfObjectAttribute[domainResultModel](ctx),
			names.AttrEnabled: schema.BoolAttribute{
				Computed: true,
			},
			"etag": schema.StringAttribute{
				Computed: true,
			},
			names.AttrID: schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"managed_certificate_request": framework.DataSourceComputedListOfObjectAttribute[managedCertificateRequestModel](ctx),
			names.AttrName: schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			names.AttrParameters: framework.DataSourceComputedListOfObjectAttribute[parameterModel](ctx),
			names.AttrStatus: schema.StringAttribute{
				Computed: true,
			},
			names.AttrTags: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{},
	}
}

func (d *distributionTenantDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data distributionTenantDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().CloudFrontClient(ctx)

	// Define lookup strategies using config values
	lookupStrategies := []struct {
		value types.String
		fn    func(context.Context, *cloudfront.Client, string) (any, error)
	}{
		{data.ID, func(ctx context.Context, conn *cloudfront.Client, id string) (any, error) {
			return findDistributionTenantByIdentifier(ctx, conn, id)
		}},
		{data.ARN, func(ctx context.Context, conn *cloudfront.Client, arn string) (any, error) {
			return findDistributionTenantByIdentifier(ctx, conn, arn)
		}},
		{data.Name, func(ctx context.Context, conn *cloudfront.Client, name string) (any, error) {
			return findDistributionTenantByIdentifier(ctx, conn, name)
		}},
		{data.Domain, func(ctx context.Context, conn *cloudfront.Client, domain string) (any, error) {
			return findDistributionTenantByDomain(ctx, conn, domain)
		}},
	}

	var (
		output any
		err    error
		key    string
	)

	// Try each lookup strategy until we find a non-null, non-unknown value
	for _, strategy := range lookupStrategies {
		if !strategy.value.IsNull() && !strategy.value.IsUnknown() {
			key = fwflex.StringValueFromFramework(ctx, strategy.value)
			output, err = strategy.fn(ctx, conn, key)
			break
		}
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading CloudFront Distribution Tenant (%s)", key), err.Error())
		return
	}

	// Handle both output types
	var tenant *awstypes.DistributionTenant
	var etag *string
	switch v := output.(type) {
	case *cloudfront.GetDistributionTenantOutput:
		tenant = v.DistributionTenant
		etag = v.ETag
	case *cloudfront.GetDistributionTenantByDomainOutput:
		tenant = v.DistributionTenant
		etag = v.ETag
	}

	// Flatten the distribution tenant data into the model
	response.Diagnostics.Append(fwflex.Flatten(ctx, tenant, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Use AutoFlex to flatten the response
	response.Diagnostics.Append(fwflex.Flatten(ctx, tenant, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Set computed fields that need special handling
	data.ID = fwflex.StringToFramework(ctx, tenant.Id)
	data.ETag = fwflex.StringToFramework(ctx, etag)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

type distributionTenantDataSourceModel struct {
	ARN                       types.String                                                    `tfsdk:"arn"`
	ConnectionGroupID         types.String                                                    `tfsdk:"connection_group_id"`
	Customizations            fwtypes.ListNestedObjectValueOf[customizationsModel]            `tfsdk:"customizations"`
	DistributionID            types.String                                                    `tfsdk:"distribution_id"`
	Domain                    types.String                                                    `tfsdk:"domain"`
	Domains                   fwtypes.ListNestedObjectValueOf[domainResultModel]              `tfsdk:"domains" autoflex:",xmlwrapper=Items"`
	Enabled                   types.Bool                                                      `tfsdk:"enabled"`
	ETag                      types.String                                                    `tfsdk:"etag"`
	ID                        types.String                                                    `tfsdk:"id"`
	ManagedCertificateRequest fwtypes.ListNestedObjectValueOf[managedCertificateRequestModel] `tfsdk:"managed_certificate_request"`
	Name                      types.String                                                    `tfsdk:"name"`
	Parameters                fwtypes.ListNestedObjectValueOf[parameterModel]                 `tfsdk:"parameters" autoflex:",xmlwrapper=Items"`
	Status                    types.String                                                    `tfsdk:"status"`
	Tags                      tftags.Map                                                      `tfsdk:"tags"`
}

func (d *distributionTenantDataSource) ConfigValidators(_ context.Context) []datasource.ConfigValidator {
	return []datasource.ConfigValidator{
		datasourcevalidator.ExactlyOneOf(
			path.MatchRoot(names.AttrID),
			path.MatchRoot(names.AttrARN),
			path.MatchRoot(names.AttrName),
			path.MatchRoot(names.AttrDomain),
		),
	}
}
func findDistributionTenantByDomain(ctx context.Context, conn *cloudfront.Client, domain string) (*cloudfront.GetDistributionTenantByDomainOutput, error) {
	input := cloudfront.GetDistributionTenantByDomainInput{
		Domain: aws.String(domain),
	}
	output, err := conn.GetDistributionTenantByDomain(ctx, &input)

	if errs.IsA[*awstypes.EntityNotFound](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.DistributionTenant == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output, nil
}
