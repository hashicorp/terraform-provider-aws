// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package cloudfront

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/datasourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
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

const (
	DSNameDistributionTenant = "Distribution Tenant Data Source"
)

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
			"distribution_id": schema.StringAttribute{
				Computed: true,
			},
			names.AttrDomain: schema.StringAttribute{
				Optional: true,
			},
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
			"last_modified_time": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			names.AttrName: schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			names.AttrStatus: schema.StringAttribute{
				Computed: true,
			},
			names.AttrTags: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			"customizations": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[customizationsDataSourceModel](ctx),
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"geo_restriction": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[geoRestrictionDataSourceModel](ctx),
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"locations": schema.SetAttribute{
										ElementType: types.StringType,
										Computed:    true,
									},
									"restriction_type": schema.StringAttribute{
										Computed:   true,
										CustomType: fwtypes.StringEnumType[awstypes.GeoRestrictionType](),
									},
								},
							},
						},
						names.AttrCertificate: schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[certificateDataSourceModel](ctx),
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrARN: schema.StringAttribute{
										Computed:   true,
										CustomType: fwtypes.ARNType,
									},
								},
							},
						},
						"web_acl": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[webAclDataSourceModel](ctx),
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrAction: schema.StringAttribute{
										Computed:   true,
										CustomType: fwtypes.StringEnumType[awstypes.CustomizationActionType](),
									},
									names.AttrARN: schema.StringAttribute{
										Computed:   true,
										CustomType: fwtypes.ARNType,
									},
								},
							},
						},
					},
				},
			},
			"managed_certificate_request": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[managedCertificateRequestDataSourceModel](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"certificate_transparency_logging_preference": schema.StringAttribute{
							Computed:   true,
							CustomType: fwtypes.StringEnumType[awstypes.CertificateTransparencyLoggingPreference](),
						},
						"primary_domain_name": schema.StringAttribute{
							Computed: true,
						},
						"validation_token_host": schema.StringAttribute{
							Computed:   true,
							CustomType: fwtypes.StringEnumType[awstypes.ValidationTokenHost](),
						},
					},
				},
			},
			"domains": schema.SetNestedBlock{
				CustomType: fwtypes.NewSetNestedObjectTypeOf[domainItemDataSourceModel](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrDomain: schema.StringAttribute{
							Computed: true,
						},
					},
				},
			},
			names.AttrParameters: schema.SetNestedBlock{
				CustomType: fwtypes.NewSetNestedObjectTypeOf[parameterDataSourceModel](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrName: schema.StringAttribute{
							Computed: true,
						},
						names.AttrValue: schema.StringAttribute{
							Computed: true,
						},
					},
				},
			},
		},
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

	var output any
	var err error

	// Try each lookup strategy until we find a non-null, non-unknown value
	for _, strategy := range lookupStrategies {
		if !strategy.value.IsNull() && !strategy.value.IsUnknown() {
			output, err = strategy.fn(ctx, conn, strategy.value.ValueString())
			break
		}
	}

	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.CloudFront, create.ErrActionReading, DSNameDistributionTenant, data.ID.String(), err),
			err.Error(),
		)
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

	// Manually flatten domains and parameters using helper functions from resource file
	var diags diag.Diagnostics
	data.Domains, diags = flattenDomainsDataSource(ctx, tenant.Domains)
	response.Diagnostics.Append(diags...)
	data.Parameters, diags = flattenParametersDataSource(ctx, tenant.Parameters)
	response.Diagnostics.Append(diags...)
	data.Customizations, diags = flattenCustomizationsDataSource(ctx, tenant.Customizations)
	response.Diagnostics.Append(diags...)

	// Set computed fields that need special handling
	data.ID = fwflex.StringToFramework(ctx, tenant.Id)
	data.ETag = fwflex.StringToFramework(ctx, etag)
	if tenant.LastModifiedTime != nil {
		data.LastModifiedTime = fwflex.TimeToFramework(ctx, tenant.LastModifiedTime)
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

type distributionTenantDataSourceModel struct {
	ARN                       types.String                                                              `tfsdk:"arn"`
	ConnectionGroupID         types.String                                                              `tfsdk:"connection_group_id"`
	Customizations            fwtypes.ListNestedObjectValueOf[customizationsDataSourceModel]            `tfsdk:"customizations"`
	DistributionID            types.String                                                              `tfsdk:"distribution_id"`
	Domain                    types.String                                                              `tfsdk:"domain"`
	Domains                   fwtypes.SetNestedObjectValueOf[domainItemDataSourceModel]                 `tfsdk:"domains"`
	Enabled                   types.Bool                                                                `tfsdk:"enabled"`
	ETag                      types.String                                                              `tfsdk:"etag"`
	ID                        types.String                                                              `tfsdk:"id"`
	LastModifiedTime          timetypes.RFC3339                                                         `tfsdk:"last_modified_time"`
	ManagedCertificateRequest fwtypes.ListNestedObjectValueOf[managedCertificateRequestDataSourceModel] `tfsdk:"managed_certificate_request"`
	Name                      types.String                                                              `tfsdk:"name"`
	Parameters                fwtypes.SetNestedObjectValueOf[parameterDataSourceModel]                  `tfsdk:"parameters"`
	Status                    types.String                                                              `tfsdk:"status"`
	Tags                      tftags.Map                                                                `tfsdk:"tags"`
}

type customizationsDataSourceModel struct {
	GeoRestriction fwtypes.ListNestedObjectValueOf[geoRestrictionDataSourceModel] `tfsdk:"geo_restriction"`
	Certificate    fwtypes.ListNestedObjectValueOf[certificateDataSourceModel]    `tfsdk:"certificate"`
	WebAcl         fwtypes.ListNestedObjectValueOf[webAclDataSourceModel]         `tfsdk:"web_acl"`
}

// Implement fwflex.Flattener interface
var (
	_ fwflex.Flattener = &customizationsDataSourceModel{}
	_ fwflex.Flattener = &geoRestrictionDataSourceModel{}
	_ fwflex.Flattener = &certificateDataSourceModel{}
	_ fwflex.Flattener = &webAclDataSourceModel{}
	_ fwflex.Flattener = &managedCertificateRequestDataSourceModel{}
	_ fwflex.Flattener = &parameterDataSourceModel{}
	_ fwflex.Flattener = &domainItemDataSourceModel{}
)

type domainItemDataSourceModel struct {
	Domain types.String `tfsdk:"domain"`
}

type geoRestrictionDataSourceModel struct {
	Locations       fwtypes.SetOfString                             `tfsdk:"locations"`
	RestrictionType fwtypes.StringEnum[awstypes.GeoRestrictionType] `tfsdk:"restriction_type"`
}

type certificateDataSourceModel struct {
	ARN fwtypes.ARN `tfsdk:"arn"`
}

type webAclDataSourceModel struct {
	Action fwtypes.StringEnum[awstypes.CustomizationActionType] `tfsdk:"action"`
	ARN    fwtypes.ARN                                          `tfsdk:"arn"`
}

type managedCertificateRequestDataSourceModel struct {
	CertificateTransparencyLoggingPreference fwtypes.StringEnum[awstypes.CertificateTransparencyLoggingPreference] `tfsdk:"certificate_transparency_logging_preference"`
	PrimaryDomainName                        types.String                                                          `tfsdk:"primary_domain_name"`
	ValidationTokenHost                      fwtypes.StringEnum[awstypes.ValidationTokenHost]                      `tfsdk:"validation_token_host"`
}

type parameterDataSourceModel struct {
	Name  types.String `tfsdk:"name"`
	Value types.String `tfsdk:"value"`
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
	input := &cloudfront.GetDistributionTenantByDomainInput{
		Domain: &domain,
	}

	output, err := conn.GetDistributionTenantByDomain(ctx, input)
	if err != nil {
		return nil, err
	}

	if output == nil || output.DistributionTenant == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

// Implement fwflex.Flattener interface methods
func (m *customizationsDataSourceModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics

	if v == nil {
		return diags
	}

	if t, ok := v.(*awstypes.Customizations); ok {
		if t.GeoRestrictions != nil {
			var geoModel geoRestrictionDataSourceModel
			diags.Append(geoModel.Flatten(ctx, t.GeoRestrictions)...)
			if diags.HasError() {
				return diags
			}
			geoList, d := fwtypes.NewListNestedObjectValueOfPtr(ctx, &geoModel)
			diags.Append(d...)
			if diags.HasError() {
				return diags
			}
			m.GeoRestriction = geoList
		} else {
			m.GeoRestriction = fwtypes.NewListNestedObjectValueOfNull[geoRestrictionDataSourceModel](ctx)
		}

		if t.Certificate != nil {
			var certModel certificateDataSourceModel
			diags.Append(certModel.Flatten(ctx, t.Certificate)...)
			if diags.HasError() {
				return diags
			}
			certList, d := fwtypes.NewListNestedObjectValueOfPtr(ctx, &certModel)
			diags.Append(d...)
			if diags.HasError() {
				return diags
			}
			m.Certificate = certList
		} else {
			m.Certificate = fwtypes.NewListNestedObjectValueOfNull[certificateDataSourceModel](ctx)
		}

		if t.WebAcl != nil {
			var webAclModel webAclDataSourceModel
			diags.Append(webAclModel.Flatten(ctx, t.WebAcl)...)
			if diags.HasError() {
				return diags
			}
			webAclList, d := fwtypes.NewListNestedObjectValueOfPtr(ctx, &webAclModel)
			diags.Append(d...)
			if diags.HasError() {
				return diags
			}
			m.WebAcl = webAclList
		} else {
			m.WebAcl = fwtypes.NewListNestedObjectValueOfNull[webAclDataSourceModel](ctx)
		}
	}

	return diags
}

func (m *geoRestrictionDataSourceModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics

	if v == nil {
		return diags
	}

	if t, ok := v.(*awstypes.GeoRestrictionCustomization); ok {
		m.RestrictionType = fwtypes.StringEnumValue(t.RestrictionType)

		// Convert locations slice to SetOfString
		if len(t.Locations) > 0 {
			// Filter out empty strings
			filteredLocations := make([]string, 0, len(t.Locations))
			for _, location := range t.Locations {
				if location != "" {
					filteredLocations = append(filteredLocations, location)
				}
			}

			if len(filteredLocations) > 0 {
				// Convert strings to attr.Value slice
				elements := make([]attr.Value, len(filteredLocations))
				for i, location := range filteredLocations {
					elements[i] = basetypes.NewStringValue(location)
				}
				setVal, d := fwtypes.NewSetValueOf[basetypes.StringValue](ctx, elements)
				diags.Append(d...)
				m.Locations = setVal
			} else {
				m.Locations = fwtypes.NewSetValueOfNull[basetypes.StringValue](ctx)
			}
		} else {
			m.Locations = fwtypes.NewSetValueOfNull[basetypes.StringValue](ctx)
		}
	}

	return diags
}

func (m *certificateDataSourceModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics

	if v == nil {
		return diags
	}

	if t, ok := v.(*awstypes.Certificate); ok {
		m.ARN = fwtypes.ARNValue(aws.ToString(t.Arn))
	}

	return diags
}

func (m *webAclDataSourceModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics

	if v == nil {
		return diags
	}

	if t, ok := v.(*awstypes.WebAclCustomization); ok {
		m.Action = fwtypes.StringEnumValue(t.Action)
		m.ARN = fwtypes.ARNValue(aws.ToString(t.Arn))
	}

	return diags
}

func (m *managedCertificateRequestDataSourceModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics

	if v == nil {
		return diags
	}

	if t, ok := v.(*awstypes.ManagedCertificateRequest); ok {
		m.CertificateTransparencyLoggingPreference = fwtypes.StringEnumValue(t.CertificateTransparencyLoggingPreference)
		m.PrimaryDomainName = fwflex.StringToFramework(ctx, t.PrimaryDomainName)
		m.ValidationTokenHost = fwtypes.StringEnumValue(t.ValidationTokenHost)
	}

	return diags
}

func (m *parameterDataSourceModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics

	if v == nil {
		return diags
	}

	if t, ok := v.(*awstypes.Parameter); ok {
		m.Name = fwflex.StringToFramework(ctx, t.Name)
		m.Value = fwflex.StringToFramework(ctx, t.Value)
	}

	return diags
}

func (m *domainItemDataSourceModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics

	if v == nil {
		return diags
	}

	if t, ok := v.(*awstypes.DomainResult); ok {
		m.Domain = fwflex.StringToFramework(ctx, t.Domain)
	}

	return diags
}

// flattenDomainsDataSource converts AWS SDK DomainResult slice to framework ListNestedObjectValueOf for data source
func flattenDomainsDataSource(ctx context.Context, domains []awstypes.DomainResult) (fwtypes.SetNestedObjectValueOf[domainItemDataSourceModel], diag.Diagnostics) {
	var diags diag.Diagnostics

	domainModels := make([]*domainItemDataSourceModel, 0, len(domains))
	for _, domainResult := range domains {
		domainModels = append(domainModels, &domainItemDataSourceModel{
			Domain: fwflex.StringToFramework(ctx, domainResult.Domain),
		})
	}

	domainsSet, d := fwtypes.NewSetNestedObjectValueOfSlice(ctx, domainModels, nil)
	diags.Append(d...)

	return domainsSet, diags
}

// flattenParametersDataSource converts AWS SDK Parameter slice to framework ListNestedObjectValueOf for data source
func flattenParametersDataSource(ctx context.Context, parameters []awstypes.Parameter) (fwtypes.SetNestedObjectValueOf[parameterDataSourceModel], diag.Diagnostics) {
	var diags diag.Diagnostics

	parameterModels := make([]*parameterDataSourceModel, 0, len(parameters))
	for _, param := range parameters {
		parameterModels = append(parameterModels, &parameterDataSourceModel{
			Name:  fwflex.StringToFramework(ctx, param.Name),
			Value: fwflex.StringToFramework(ctx, param.Value),
		})
	}

	parametersSet, d := fwtypes.NewSetNestedObjectValueOfSlice(ctx, parameterModels, nil)
	diags.Append(d...)

	return parametersSet, diags
}

// flattenCustomizationsDataSource converts AWS SDK Customizations to framework ListNestedObjectValueOf for data source
func flattenCustomizationsDataSource(ctx context.Context, customizations *awstypes.Customizations) (fwtypes.ListNestedObjectValueOf[customizationsDataSourceModel], diag.Diagnostics) {
	var diags diag.Diagnostics

	if customizations == nil {
		return fwtypes.NewListNestedObjectValueOfNull[customizationsDataSourceModel](ctx), diags
	}

	var customModel customizationsDataSourceModel
	diags.Append(customModel.Flatten(ctx, customizations)...)
	if diags.HasError() {
		return fwtypes.NewListNestedObjectValueOfNull[customizationsDataSourceModel](ctx), diags
	}

	customList, d := fwtypes.NewListNestedObjectValueOfPtr(ctx, &customModel)
	diags.Append(d...)

	return customList, diags
}
