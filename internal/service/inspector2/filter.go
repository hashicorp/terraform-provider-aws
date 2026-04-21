// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package inspector2

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/inspector2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/inspector2/types"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_inspector2_filter", name="Filter")
// @Tags(identifierAttribute="arn")
// @ArnIdentity
// @Testing(tagsTest=true)
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/inspector2/types;types.Filter")
// @Testing(importStateIdAttribute="arn")
// @Testing(preIdentityVersion="6.19.0")
func newFilterResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &filterResource{}, nil
}

type filterResource struct {
	framework.ResourceWithModel[filterResourceModel]
	framework.WithImportByIdentity
}

func (r *filterResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	const (
		defaultFilterSchemaMaxSize = 20
	)
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrAction: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.FilterAction](),
				Required:   true,
			},
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
			},
			names.AttrName: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"reason": schema.StringAttribute{
				Optional: true,
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			"filter_criteria": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[filterCriteriaModel](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtLeast(1),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						names.AttrAWSAccountID:               stringFilterSchemaFramework(ctx, defaultFilterSchemaMaxSize),
						"code_repository_project_name":       stringFilterSchemaFramework(ctx, defaultFilterSchemaMaxSize),
						"code_repository_provider_type":      stringFilterSchemaFramework(ctx, defaultFilterSchemaMaxSize),
						"code_vulnerability_detector_name":   stringFilterSchemaFramework(ctx, defaultFilterSchemaMaxSize),
						"code_vulnerability_detector_tags":   stringFilterSchemaFramework(ctx, defaultFilterSchemaMaxSize),
						"code_vulnerability_file_path":       stringFilterSchemaFramework(ctx, defaultFilterSchemaMaxSize),
						"component_id":                       stringFilterSchemaFramework(ctx, defaultFilterSchemaMaxSize),
						"component_type":                     stringFilterSchemaFramework(ctx, defaultFilterSchemaMaxSize),
						"ec2_instance_image_id":              stringFilterSchemaFramework(ctx, defaultFilterSchemaMaxSize),
						"ec2_instance_subnet_id":             stringFilterSchemaFramework(ctx, defaultFilterSchemaMaxSize),
						"ec2_instance_vpc_id":                stringFilterSchemaFramework(ctx, defaultFilterSchemaMaxSize),
						"ecr_image_architecture":             stringFilterSchemaFramework(ctx, defaultFilterSchemaMaxSize),
						"ecr_image_hash":                     stringFilterSchemaFramework(ctx, defaultFilterSchemaMaxSize),
						"ecr_image_in_use_count":             numberFilterSchemaFramework(ctx, defaultFilterSchemaMaxSize),
						"ecr_image_last_in_use_at":           dateFilterSchemaFramework(ctx, defaultFilterSchemaMaxSize),
						"ecr_image_pushed_at":                dateFilterSchemaFramework(ctx, defaultFilterSchemaMaxSize),
						"ecr_image_registry":                 stringFilterSchemaFramework(ctx, defaultFilterSchemaMaxSize),
						"ecr_image_repository_name":          stringFilterSchemaFramework(ctx, defaultFilterSchemaMaxSize),
						"ecr_image_tags":                     stringFilterSchemaFramework(ctx, defaultFilterSchemaMaxSize),
						"epss_score":                         numberFilterSchemaFramework(ctx, defaultFilterSchemaMaxSize),
						"exploit_available":                  stringFilterSchemaFramework(ctx, defaultFilterSchemaMaxSize),
						"finding_arn":                        stringFilterSchemaFramework(ctx, defaultFilterSchemaMaxSize),
						"finding_status":                     stringFilterSchemaFramework(ctx, defaultFilterSchemaMaxSize),
						"finding_type":                       stringFilterSchemaFramework(ctx, defaultFilterSchemaMaxSize),
						"first_observed_at":                  dateFilterSchemaFramework(ctx, defaultFilterSchemaMaxSize),
						"fix_available":                      stringFilterSchemaFramework(ctx, defaultFilterSchemaMaxSize),
						"inspector_score":                    numberFilterSchemaFramework(ctx, defaultFilterSchemaMaxSize),
						"lambda_function_execution_role_arn": stringFilterSchemaFramework(ctx, defaultFilterSchemaMaxSize),
						"lambda_function_last_modified_at":   dateFilterSchemaFramework(ctx, defaultFilterSchemaMaxSize),
						"lambda_function_layers":             stringFilterSchemaFramework(ctx, defaultFilterSchemaMaxSize),
						"lambda_function_name":               stringFilterSchemaFramework(ctx, defaultFilterSchemaMaxSize),
						"lambda_function_runtime":            stringFilterSchemaFramework(ctx, defaultFilterSchemaMaxSize),
						"last_observed_at":                   dateFilterSchemaFramework(ctx, defaultFilterSchemaMaxSize),
						"network_protocol":                   stringFilterSchemaFramework(ctx, defaultFilterSchemaMaxSize),
						"port_range":                         portRangeFilterSchemaFramework(ctx, defaultFilterSchemaMaxSize),
						"related_vulnerabilities":            stringFilterSchemaFramework(ctx, defaultFilterSchemaMaxSize),
						names.AttrResourceID:                 stringFilterSchemaFramework(ctx, defaultFilterSchemaMaxSize),
						names.AttrResourceTags:               mapFilterSchemaFramework(ctx, defaultFilterSchemaMaxSize),
						names.AttrResourceType:               stringFilterSchemaFramework(ctx, defaultFilterSchemaMaxSize),
						"severity":                           stringFilterSchemaFramework(ctx, defaultFilterSchemaMaxSize),
						"title":                              stringFilterSchemaFramework(ctx, defaultFilterSchemaMaxSize),
						"updated_at":                         dateFilterSchemaFramework(ctx, defaultFilterSchemaMaxSize),
						"vendor_severity":                    stringFilterSchemaFramework(ctx, defaultFilterSchemaMaxSize),
						"vulnerability_id":                   stringFilterSchemaFramework(ctx, defaultFilterSchemaMaxSize),
						"vulnerability_source":               stringFilterSchemaFramework(ctx, defaultFilterSchemaMaxSize),
						"vulnerable_packages":                packageFilterSchemaFramework(ctx, defaultFilterSchemaMaxSize),
					},
				},
			},
		},
	}
}

func dateFilterSchemaFramework(ctx context.Context, maxSize int) schema.SetNestedBlock {
	return schema.SetNestedBlock{
		CustomType: fwtypes.NewSetNestedObjectTypeOf[dateFilterModel](ctx),
		Validators: []validator.Set{
			setvalidator.SizeAtMost(maxSize),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"end_inclusive": schema.StringAttribute{
					CustomType: timetypes.RFC3339Type{},
					Optional:   true,
				},
				"start_inclusive": schema.StringAttribute{
					CustomType: timetypes.RFC3339Type{},
					Optional:   true,
				},
			},
		},
	}
}

func mapFilterSchemaFramework(ctx context.Context, maxSize int) schema.SetNestedBlock {
	return schema.SetNestedBlock{
		CustomType: fwtypes.NewSetNestedObjectTypeOf[mapFilterModel](ctx),
		Validators: []validator.Set{
			setvalidator.SizeAtMost(maxSize),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"comparison": schema.StringAttribute{
					CustomType: fwtypes.StringEnumType[awstypes.MapComparison](),
					Required:   true,
				},
				names.AttrKey: schema.StringAttribute{
					Required: true,
				},
				names.AttrValue: schema.StringAttribute{
					Required: true,
				},
			},
		},
	}
}

func numberFilterSchemaFramework(ctx context.Context, maxSize int) schema.SetNestedBlock {
	return schema.SetNestedBlock{
		CustomType: fwtypes.NewSetNestedObjectTypeOf[numberFilterModel](ctx),
		Validators: []validator.Set{
			setvalidator.SizeAtMost(maxSize),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"lower_inclusive": schema.Float64Attribute{
					Required: true,
				},
				"upper_inclusive": schema.Float64Attribute{
					Required: true,
				},
			},
		},
	}
}

func portRangeFilterSchemaFramework(ctx context.Context, maxSize int) schema.SetNestedBlock {
	return schema.SetNestedBlock{
		CustomType: fwtypes.NewSetNestedObjectTypeOf[portRangeFilterModel](ctx),
		Validators: []validator.Set{
			setvalidator.SizeAtMost(maxSize),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"begin_inclusive": schema.Int32Attribute{
					Required: true,
				},
				"end_inclusive": schema.Int32Attribute{
					Required: true,
				},
			},
		},
	}
}

func stringFilterSchemaFramework(ctx context.Context, maxSize int) schema.SetNestedBlock {
	return schema.SetNestedBlock{
		CustomType: fwtypes.NewSetNestedObjectTypeOf[stringFilterModel](ctx),
		Validators: []validator.Set{
			setvalidator.SizeAtMost(maxSize),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"comparison": schema.StringAttribute{
					CustomType: fwtypes.StringEnumType[awstypes.StringComparison](),
					Required:   true,
				},
				names.AttrValue: schema.StringAttribute{
					Required: true,
				},
			},
		},
	}
}

func packageFilterSchemaFramework(ctx context.Context, maxSize int) schema.SetNestedBlock {
	return schema.SetNestedBlock{
		CustomType: fwtypes.NewSetNestedObjectTypeOf[packageFilterModel](ctx),
		Validators: []validator.Set{
			setvalidator.SizeAtMost(maxSize),
		},
		NestedObject: schema.NestedBlockObject{
			Blocks: map[string]schema.Block{
				"architecture": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[stringFilterModel](ctx),
					Validators: []validator.List{
						listvalidator.SizeAtMost(1),
					},
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"comparison": schema.StringAttribute{
								CustomType: fwtypes.StringEnumType[awstypes.StringComparison](),
								Required:   true,
							},
							names.AttrValue: schema.StringAttribute{
								Required: true,
							},
						},
					},
				},
				"epoch": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[numberFilterModel](ctx),
					Validators: []validator.List{
						listvalidator.SizeAtMost(maxSize),
					},
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"lower_inclusive": schema.Float64Attribute{
								Required: true,
							},
							"upper_inclusive": schema.Float64Attribute{
								Required: true,
							},
						},
					},
				},
				"file_path": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[stringFilterModel](ctx),
					Validators: []validator.List{
						listvalidator.SizeAtMost(1),
					},
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"comparison": schema.StringAttribute{
								CustomType: fwtypes.StringEnumType[awstypes.StringComparison](),
								Required:   true,
							},
							names.AttrValue: schema.StringAttribute{
								Required: true,
							},
						},
					},
				},
				names.AttrName: schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[stringFilterModel](ctx),
					Validators: []validator.List{
						listvalidator.SizeAtMost(1),
					},
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"comparison": schema.StringAttribute{
								CustomType: fwtypes.StringEnumType[awstypes.StringComparison](),
								Required:   true,
							},
							names.AttrValue: schema.StringAttribute{
								Required: true,
							},
						},
					},
				},
				"release": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[stringFilterModel](ctx),
					Validators: []validator.List{
						listvalidator.SizeAtMost(1),
					},
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"comparison": schema.StringAttribute{
								CustomType: fwtypes.StringEnumType[awstypes.StringComparison](),
								Required:   true,
							},
							names.AttrValue: schema.StringAttribute{
								Required: true,
							},
						},
					},
				},
				"source_lambda_layer_arn": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[stringFilterModel](ctx),
					Validators: []validator.List{
						listvalidator.SizeAtMost(1),
					},
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"comparison": schema.StringAttribute{
								CustomType: fwtypes.StringEnumType[awstypes.StringComparison](),
								Required:   true,
							},
							names.AttrValue: schema.StringAttribute{
								Required: true,
							},
						},
					},
				},
				"source_layer_hash": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[stringFilterModel](ctx),
					Validators: []validator.List{
						listvalidator.SizeAtMost(1),
					},
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"comparison": schema.StringAttribute{
								CustomType: fwtypes.StringEnumType[awstypes.StringComparison](),
								Required:   true,
							},
							names.AttrValue: schema.StringAttribute{
								Required: true,
							},
						},
					},
				},
				names.AttrVersion: schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[stringFilterModel](ctx),
					Validators: []validator.List{
						listvalidator.SizeAtMost(1),
					},
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"comparison": schema.StringAttribute{
								CustomType: fwtypes.StringEnumType[awstypes.StringComparison](),
								Required:   true,
							},
							names.AttrValue: schema.StringAttribute{
								Required: true,
							},
						},
					},
				},
			},
		},
	}
}

func (r *filterResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data filterResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().Inspector2Client(ctx)

	name := fwflex.StringValueFromFramework(ctx, data.Name)
	var input inspector2.CreateFilterInput
	response.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	input.Tags = getTagsIn(ctx)

	output, err := conn.CreateFilter(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating Inspector2 Filter (%s)", name), err.Error())

		return
	}

	data.ARN = fwflex.StringToFramework(ctx, output.Arn)

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *filterResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data filterResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().Inspector2Client(ctx)

	output, err := findFilterByARN(ctx, conn, data.ARN.ValueString())

	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Inspector2 Filter (%s)", data.ARN.ValueString()), err.Error())

		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data, fwflex.WithFieldNamePrefix("Filter"))...)
	if response.Diagnostics.HasError() {
		return
	}

	setTagsOut(ctx, output.Tags)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *filterResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var new, old filterResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.State.Get(ctx, &old)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().Inspector2Client(ctx)

	if !new.Action.Equal(old.Action) ||
		!new.Description.Equal(old.Description) ||
		!new.FilterCriteria.Equal(old.FilterCriteria) ||
		!new.Name.Equal(old.Name) ||
		!new.Reason.Equal(old.Reason) {
		var input inspector2.UpdateFilterInput
		response.Diagnostics.Append(fwflex.Expand(ctx, new, &input, fwflex.WithFieldNamePrefix("Filter"))...)
		if response.Diagnostics.HasError() {
			return
		}

		// Additional fields.
		input.FilterArn = fwflex.StringFromFramework(ctx, new.ARN)

		_, err := conn.UpdateFilter(ctx, &input)

		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("updating Inspector2 Filter (%s)", new.ARN.ValueString()), err.Error())

			return
		}
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *filterResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data filterResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().Inspector2Client(ctx)

	input := inspector2.DeleteFilterInput{
		Arn: fwflex.StringFromFramework(ctx, data.ARN),
	}
	_, err := conn.DeleteFilter(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting Inspector2 Filter (%s)", data.ARN.ValueString()), err.Error())

		return
	}
}

func findFilterByARN(ctx context.Context, conn *inspector2.Client, arn string) (*awstypes.Filter, error) {
	input := inspector2.ListFiltersInput{
		Arns: []string{arn},
	}

	return findFilter(ctx, conn, &input)
}

func findFilter(ctx context.Context, conn *inspector2.Client, input *inspector2.ListFiltersInput) (*awstypes.Filter, error) {
	output, err := findFilters(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findFilters(ctx context.Context, conn *inspector2.Client, input *inspector2.ListFiltersInput) ([]awstypes.Filter, error) {
	var output []awstypes.Filter

	pages := inspector2.NewListFiltersPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		output = append(output, page.Filters...)
	}

	return output, nil
}

type filterResourceModel struct {
	framework.WithRegionModel
	Action         fwtypes.StringEnum[awstypes.FilterAction]            `tfsdk:"action"`
	ARN            types.String                                         `tfsdk:"arn"`
	Description    types.String                                         `tfsdk:"description"`
	Name           types.String                                         `tfsdk:"name"`
	Reason         types.String                                         `tfsdk:"reason"`
	FilterCriteria fwtypes.ListNestedObjectValueOf[filterCriteriaModel] `tfsdk:"filter_criteria"`
	Tags           tftags.Map                                           `tfsdk:"tags"`
	TagsAll        tftags.Map                                           `tfsdk:"tags_all"`
}

type filterCriteriaModel struct {
	AWSAccountID                   fwtypes.SetNestedObjectValueOf[stringFilterModel]    `tfsdk:"aws_account_id"`
	CodeRepositoryProjectName      fwtypes.SetNestedObjectValueOf[stringFilterModel]    `tfsdk:"code_repository_project_name"`
	CodeRepositoryProviderType     fwtypes.SetNestedObjectValueOf[stringFilterModel]    `tfsdk:"code_repository_provider_type"`
	CodeVulnerabilityDetectorName  fwtypes.SetNestedObjectValueOf[stringFilterModel]    `tfsdk:"code_vulnerability_detector_name"`
	CodeVulnerabilityDetectorTags  fwtypes.SetNestedObjectValueOf[stringFilterModel]    `tfsdk:"code_vulnerability_detector_tags"`
	CodeVulnerabilityFilePath      fwtypes.SetNestedObjectValueOf[stringFilterModel]    `tfsdk:"code_vulnerability_file_path"`
	ComponentId                    fwtypes.SetNestedObjectValueOf[stringFilterModel]    `tfsdk:"component_id"`
	ComponentType                  fwtypes.SetNestedObjectValueOf[stringFilterModel]    `tfsdk:"component_type"`
	EC2InstanceImageId             fwtypes.SetNestedObjectValueOf[stringFilterModel]    `tfsdk:"ec2_instance_image_id"`
	EC2InstanceSubnetId            fwtypes.SetNestedObjectValueOf[stringFilterModel]    `tfsdk:"ec2_instance_subnet_id"`
	EC2InstanceVpcId               fwtypes.SetNestedObjectValueOf[stringFilterModel]    `tfsdk:"ec2_instance_vpc_id"`
	ECRImageArchitecture           fwtypes.SetNestedObjectValueOf[stringFilterModel]    `tfsdk:"ecr_image_architecture"`
	ECRImageHash                   fwtypes.SetNestedObjectValueOf[stringFilterModel]    `tfsdk:"ecr_image_hash"`
	ECRImageInUseCount             fwtypes.SetNestedObjectValueOf[numberFilterModel]    `tfsdk:"ecr_image_in_use_count"`
	ECRImageLastInUseAt            fwtypes.SetNestedObjectValueOf[dateFilterModel]      `tfsdk:"ecr_image_last_in_use_at"`
	ECRImagePushedAt               fwtypes.SetNestedObjectValueOf[dateFilterModel]      `tfsdk:"ecr_image_pushed_at"`
	ECRImageRegistry               fwtypes.SetNestedObjectValueOf[stringFilterModel]    `tfsdk:"ecr_image_registry"`
	ECRImageRepositoryName         fwtypes.SetNestedObjectValueOf[stringFilterModel]    `tfsdk:"ecr_image_repository_name"`
	ECRImageTags                   fwtypes.SetNestedObjectValueOf[stringFilterModel]    `tfsdk:"ecr_image_tags"`
	EPSSScore                      fwtypes.SetNestedObjectValueOf[numberFilterModel]    `tfsdk:"epss_score"`
	ExploitAvailable               fwtypes.SetNestedObjectValueOf[stringFilterModel]    `tfsdk:"exploit_available"`
	FindingARN                     fwtypes.SetNestedObjectValueOf[stringFilterModel]    `tfsdk:"finding_arn"`
	FindingStatus                  fwtypes.SetNestedObjectValueOf[stringFilterModel]    `tfsdk:"finding_status"`
	FindingType                    fwtypes.SetNestedObjectValueOf[stringFilterModel]    `tfsdk:"finding_type"`
	FirstObservedAt                fwtypes.SetNestedObjectValueOf[dateFilterModel]      `tfsdk:"first_observed_at"`
	FixAvailable                   fwtypes.SetNestedObjectValueOf[stringFilterModel]    `tfsdk:"fix_available"`
	InspectorScore                 fwtypes.SetNestedObjectValueOf[numberFilterModel]    `tfsdk:"inspector_score"`
	LambdaFunctionExecutionRoleARN fwtypes.SetNestedObjectValueOf[stringFilterModel]    `tfsdk:"lambda_function_execution_role_arn"`
	LambdaFunctionLastModifiedAt   fwtypes.SetNestedObjectValueOf[dateFilterModel]      `tfsdk:"lambda_function_last_modified_at"`
	LambdaFunctionLayers           fwtypes.SetNestedObjectValueOf[stringFilterModel]    `tfsdk:"lambda_function_layers"`
	LambdaFunctionName             fwtypes.SetNestedObjectValueOf[stringFilterModel]    `tfsdk:"lambda_function_name"`
	LambdaFunctionRuntime          fwtypes.SetNestedObjectValueOf[stringFilterModel]    `tfsdk:"lambda_function_runtime"`
	LastObservedAt                 fwtypes.SetNestedObjectValueOf[dateFilterModel]      `tfsdk:"last_observed_at"`
	NetworkProtocol                fwtypes.SetNestedObjectValueOf[stringFilterModel]    `tfsdk:"network_protocol"`
	PortRange                      fwtypes.SetNestedObjectValueOf[portRangeFilterModel] `tfsdk:"port_range"`
	RelatedVulnerabilities         fwtypes.SetNestedObjectValueOf[stringFilterModel]    `tfsdk:"related_vulnerabilities"`
	ResourceId                     fwtypes.SetNestedObjectValueOf[stringFilterModel]    `tfsdk:"resource_id"`
	ResourceTags                   fwtypes.SetNestedObjectValueOf[mapFilterModel]       `tfsdk:"resource_tags"`
	ResourceType                   fwtypes.SetNestedObjectValueOf[stringFilterModel]    `tfsdk:"resource_type"`
	Severity                       fwtypes.SetNestedObjectValueOf[stringFilterModel]    `tfsdk:"severity"`
	Title                          fwtypes.SetNestedObjectValueOf[stringFilterModel]    `tfsdk:"title"`
	UpdatedAt                      fwtypes.SetNestedObjectValueOf[dateFilterModel]      `tfsdk:"updated_at"`
	VendorSeverity                 fwtypes.SetNestedObjectValueOf[stringFilterModel]    `tfsdk:"vendor_severity"`
	VulnerabilityId                fwtypes.SetNestedObjectValueOf[stringFilterModel]    `tfsdk:"vulnerability_id"`
	VulnerabilitySource            fwtypes.SetNestedObjectValueOf[stringFilterModel]    `tfsdk:"vulnerability_source"`
	VulnerablePackages             fwtypes.SetNestedObjectValueOf[packageFilterModel]   `tfsdk:"vulnerable_packages"`
}

type numberFilterModel struct {
	LowerInclusive types.Float64 `tfsdk:"lower_inclusive"`
	UpperInclusive types.Float64 `tfsdk:"upper_inclusive"`
}

type stringFilterModel struct {
	Comparison fwtypes.StringEnum[awstypes.StringComparison] `tfsdk:"comparison"`
	Value      types.String                                  `tfsdk:"value"`
}

type dateFilterModel struct {
	EndInclusive   timetypes.RFC3339 `tfsdk:"end_inclusive"`
	StartInclusive timetypes.RFC3339 `tfsdk:"start_inclusive"`
}

type portRangeFilterModel struct {
	BeginInclusive types.Int32 `tfsdk:"begin_inclusive"`
	EndInclusive   types.Int32 `tfsdk:"end_inclusive"`
}

type mapFilterModel struct {
	Comparison fwtypes.StringEnum[awstypes.MapComparison] `tfsdk:"comparison"`
	Key        types.String                               `tfsdk:"key"`
	Value      types.String                               `tfsdk:"value"`
}

type packageFilterModel struct {
	Architecture         fwtypes.ListNestedObjectValueOf[stringFilterModel] `tfsdk:"architecture"`
	Epoch                fwtypes.ListNestedObjectValueOf[numberFilterModel] `tfsdk:"epoch"`
	FilePath             fwtypes.ListNestedObjectValueOf[stringFilterModel] `tfsdk:"file_path"`
	Name                 fwtypes.ListNestedObjectValueOf[stringFilterModel] `tfsdk:"name"`
	Release              fwtypes.ListNestedObjectValueOf[stringFilterModel] `tfsdk:"release"`
	SourceLambdaLayerARN fwtypes.ListNestedObjectValueOf[stringFilterModel] `tfsdk:"source_lambda_layer_arn"`
	SourceLayerHash      fwtypes.ListNestedObjectValueOf[stringFilterModel] `tfsdk:"source_layer_hash"`
	Version              fwtypes.ListNestedObjectValueOf[stringFilterModel] `tfsdk:"version"`
}
