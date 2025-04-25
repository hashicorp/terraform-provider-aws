// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package inspector2

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/inspector2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/inspector2/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_inspector2_filter", name="Filter")
// @Tags(identifierAttribute="arn")
// @Testing(tagsTest=true)
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/inspector2/types;types.Filter")
// @Testing(importStateIdAttribute="arn")
func newResourceFilter(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceFilter{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameFilter = "Filter"
)

type resourceFilter struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
}

func (r *resourceFilter) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_inspector2_filter"
}

func (r *resourceFilter) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	const (
		defaultFilterSchemaMaxSize = 20
	)
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"arn": framework.ARNAttributeComputedOnly(),
			"action": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.FilterAction](),
				Required:   true,
			},
			"description": schema.StringAttribute{
				Optional: true,
			},
			"name": schema.StringAttribute{
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
			"timeouts": timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
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
				"name": schema.ListNestedBlock{
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
				"version": schema.ListNestedBlock{
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

func (r *resourceFilter) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().Inspector2Client(ctx)

	var plan resourceFilterModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var input inspector2.CreateFilterInput
	input.Tags = getTagsIn(ctx)

	resp.Diagnostics.Append(flex.Expand(ctx, plan, &input)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := conn.CreateFilter(ctx, &input)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Inspector2, create.ErrActionCreating, ResNameFilter, plan.Name.String(), err),
			err.Error(),
		)
		return
	}
	if out == nil || out.Arn == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Inspector2, create.ErrActionCreating, ResNameFilter, plan.Name.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	plan.ARN = fwflex.StringToFramework(ctx, out.Arn)

	filter_out, err := findFilterByARN(ctx, conn, plan.ARN.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Inspector2, create.ErrActionSetting, ResNameFilter, plan.ARN.ValueString(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, filter_out, &plan, flex.WithFieldNamePrefix("filter_"))...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceFilter) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().Inspector2Client(ctx)

	var state resourceFilterModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findFilterByARN(ctx, conn, state.ARN.ValueString())

	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Inspector2, create.ErrActionSetting, ResNameFilter, state.ARN.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &state, flex.WithFieldNamePrefix("Filter"))...)
	if resp.Diagnostics.HasError() {
		return
	}

	setTagsOut(ctx, out.Tags)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceFilter) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {

	conn := r.Meta().Inspector2Client(ctx)

	var plan, state resourceFilterModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.Name.Equal(state.Name) ||
		!plan.Description.Equal(state.Description) ||
		!plan.Reason.Equal(state.Reason) ||
		!plan.Action.Equal(state.Action) ||
		!plan.FilterCriteria.Equal(state.FilterCriteria) {

		var input inspector2.UpdateFilterInput
		resp.Diagnostics.Append(flex.Expand(ctx, plan, &input, flex.WithFieldNamePrefix("Filter"))...)
		if resp.Diagnostics.HasError() {
			return
		}

		out, err := conn.UpdateFilter(ctx, &input)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.Inspector2, create.ErrActionUpdating, ResNameFilter, plan.ARN.String(), err),
				err.Error(),
			)
			return
		}
		if out == nil || out.Arn == nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.Inspector2, create.ErrActionUpdating, ResNameFilter, plan.ARN.String(), nil),
				errors.New("empty output").Error(),
			)
			return
		}

		filter_out, err := findFilterByARN(ctx, conn, state.ARN.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.Inspector2, create.ErrActionSetting, ResNameFilter, state.ARN.ValueString(), err),
				err.Error(),
			)
			return
		}

		resp.Diagnostics.Append(flex.Flatten(ctx, filter_out, &plan)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceFilter) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {

	conn := r.Meta().Inspector2Client(ctx)

	var state resourceFilterModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := inspector2.DeleteFilterInput{
		Arn: state.ARN.ValueStringPointer(),
	}

	_, err := conn.DeleteFilter(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}

		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Inspector2, create.ErrActionDeleting, ResNameFilter, state.ARN.String(), err),
			err.Error(),
		)
		return
	}

}

func (r *resourceFilter) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("arn"), req, resp)
}

func findFilterByARN(ctx context.Context, conn *inspector2.Client, arn string) (*awstypes.Filter, error) {
	in := &inspector2.ListFiltersInput{
		Arns: []string{arn},
	}

	out, err := conn.ListFilters(ctx, in)
	if err != nil {
		return nil, err
	}

	if out == nil || out.Filters == nil || len(out.Filters) == 0 {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}

	return &out.Filters[0], nil
}

type resourceFilterModel struct {
	ARN            types.String                                         `tfsdk:"arn"`
	Action         fwtypes.StringEnum[awstypes.FilterAction]            `tfsdk:"action"`
	Description    types.String                                         `tfsdk:"description"`
	Name           types.String                                         `tfsdk:"name"`
	Reason         types.String                                         `tfsdk:"reason"`
	Timeouts       timeouts.Value                                       `tfsdk:"timeouts"`
	FilterCriteria fwtypes.ListNestedObjectValueOf[filterCriteriaModel] `tfsdk:"filter_criteria"`
	Tags           tftags.Map                                           `tfsdk:"tags"`
	TagsAll        tftags.Map                                           `tfsdk:"tags_all"`
}

type filterCriteriaModel struct {
	AWSAccountID                   fwtypes.SetNestedObjectValueOf[stringFilterModel]    `tfsdk:"aws_account_id"`
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
