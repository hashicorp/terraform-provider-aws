// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package inspector2

// **PLEASE DELETE THIS AND ALL TIP COMMENTS BEFORE SUBMITTING A PR FOR REVIEW!**
//
// TIP: ==== INTRODUCTION ====
// Thank you for trying the skaff tool!
//
// You have opted to include these helpful comments. They all include "TIP:"
// to help you find and remove them when you're done with them.
//
// While some aspects of this file are customized to your input, the
// scaffold tool does *not* look at the AWS API and ensure it has correct
// function, structure, and variable names. It makes guesses based on
// commonalities. You will need to make significant adjustments.
//
// In other words, as generated, this is a rough outline of the work you will
// need to do. If something doesn't make sense for your situation, get rid of
// it.

import (
	// TIP: ==== IMPORTS ====
	// This is a common set of imports but not customized to your code since
	// your code hasn't been written yet. Make sure you, your IDE, or
	// goimports -w <file> fixes these imports.
	//
	// The provider linter wants your imports to be in two groups: first,
	// standard library (i.e., "fmt" or "strings"), second, everything else.
	//
	// Also, AWS Go SDK v2 may handle nested structures differently than v1,
	// using the services/inspector2/types package. If so, you'll
	// need to import types and reference the nested types, e.g., as
	// awstypes.<Type Name>.
	"context"
	"errors"
	// "fmt"
	"time"

	// "github.com/aws/aws-sdk-go-v2/aws"
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
	// tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// TIP: ==== FILE STRUCTURE ====
// All resources should follow this basic outline. Improve this resource's
// maintainability by sticking to it.
//
// 1. Package declaration
// 2. Imports
// 3. Main resource struct with schema method
// 4. Create, read, update, delete methods (in that order)
// 5. Other functions (flatteners, expanders, waiters, finders, etc.)

// Function annotations are used for resource registration to the Provider. DO NOT EDIT.
// @FrameworkResource("aws_inspector2_filter", name="Filter")
// @Tags(identifierAttribute="arn")
// @Testing(tagsTest=true)
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
	// framework.WithImportByID
}

func (r *resourceFilter) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_inspector2_filter"
}

// TIP: ==== SCHEMA ====
// In the schema, add each of the attributes in snake case (e.g.,
// delete_automated_backups).
//
// Formatting rules:
// * Alphabetize attributes to make them easier to find.
// * Do not add a blank line between attributes.
//
// Attribute basics:
//   - If a user can provide a value ("configure a value") for an
//     attribute (e.g., instances = 5), we call the attribute an
//     "argument."
//   - You change the way users interact with attributes using:
//   - Required
//   - Optional
//   - Computed
//   - There are only four valid combinations:
//
// 1. Required only - the user must provide a value
// Required: true,
//
//  2. Optional only - the user can configure or omit a value; do not
//     use Default or DefaultFunc
//
// Optional: true,
//
//  3. Computed only - the provider can provide a value but the user
//     cannot, i.e., read-only
//
// Computed: true,
//
//  4. Optional AND Computed - the provider or user can provide a value;
//     use this combination if you are using Default
//
// Optional: true,
// Computed: true,
//
// You will typically find arguments in the input struct
// (e.g., CreateDBInstanceInput) for the create operation. Sometimes
// they are only in the input struct (e.g., ModifyDBInstanceInput) for
// the modify operation.
//
// For more about schema options, visit
// https://developer.hashicorp.com/terraform/plugin/framework/handling-data/schemas?page=schemas
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
			// TIP: ==== "ID" ATTRIBUTE ====
			// When using the Terraform Plugin Framework, there is no required "id" attribute.
			// This is different from the Terraform Plugin SDK.
			//
			// Only include an "id" attribute if the AWS API has an "Id" field, such as "FilterId"
			// "id": framework.IDAttribute(),
			"name": schema.StringAttribute{
				Required: true,
				// TIP: ==== PLAN MODIFIERS ====
				// Plan modifiers were introduced with Plugin-Framework to provide a mechanism
				// for adjusting planned changes prior to apply. The planmodifier subpackage
				// provides built-in modifiers for many common use cases such as
				// requiring replacement on a value change ("ForceNew: true" in Plugin-SDK
				// resources).
				//
				// See more:
				// https://developer.hashicorp.com/terraform/plugin/framework/resources/plan-modification
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
				"end": schema.StringAttribute{
					CustomType: timetypes.RFC3339Type{},
					Optional:   true,
				},
				"start": schema.StringAttribute{
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
				"begin_inclusive": schema.Float64Attribute{
					Required: true,
				},
				"end_inclusive": schema.Float64Attribute{
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
				"epoch": schema.SetNestedBlock{
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
	// TIP: ==== RESOURCE CREATE ====
	// Generally, the Create function should do the following things. Make
	// sure there is a good reason if you don't do one of these.
	//
	// 1. Get a client connection to the relevant service
	// 2. Fetch the plan
	// 3. Populate a create input structure
	// 4. Call the AWS create/put function
	// 5. Using the output from the create function, set the minimum arguments
	//    and attributes for the Read function to work, as well as any computed
	//    only attributes.
	// 6. Use a waiter to wait for create to complete
	// 7. Save the request plan to response state

	// TIP: -- 1. Get a client connection to the relevant service
	conn := r.Meta().Inspector2Client(ctx)

	// TIP: -- 2. Fetch the plan
	var plan resourceFilterModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// TIP: -- 3. Populate a Create input structure
	var input inspector2.CreateFilterInput
	// TIP: Using a field name prefix allows mapping fields such as `ID` to `FilterId`
	resp.Diagnostics.Append(flex.Expand(ctx, plan, &input)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// TIP: -- 4. Call the AWS Create function
	out, err := conn.CreateFilter(ctx, &input)
	if err != nil {
		// TIP: Since ID has not been set yet, you cannot use plan.ARN.String()
		// in error messages at this point.
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

	// TIP: -- 5. Using the output from the create function, set attributes
	resp.Diagnostics.Append(flex.Flatten(ctx, filter_out, &plan, flex.WithFieldNamePrefix("filter_"))...)
	if resp.Diagnostics.HasError() {
		return
	}

	// TIP: -- 7. Save the request plan to response state
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceFilter) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// TIP: ==== RESOURCE READ ====
	// Generally, the Read function should do the following things. Make
	// sure there is a good reason if you don't do one of these.
	//
	// 1. Get a client connection to the relevant service
	// 2. Fetch the state
	// 3. Get the resource from AWS
	// 4. Remove resource from state if it is not found
	// 5. Set the arguments and attributes
	// 6. Set the state

	// TIP: -- 1. Get a client connection to the relevant service
	conn := r.Meta().Inspector2Client(ctx)

	// TIP: -- 2. Fetch the state
	var state resourceFilterModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// TIP: -- 3. Get the resource from AWS using an API Get, List, or Describe-
	// type function, or, better yet, using a finder.
	out, err := findFilterByARN(ctx, conn, state.ARN.ValueString())
	// TIP: -- 4. Remove resource from state if it is not found
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

	// TIP: -- 5. Set the arguments and attributes
	// The WithFieldNamePrefix option is needed because while the create API uses filter_criteria, the list_functions returns only "criteria" without the "filter_" prefix
	resp.Diagnostics.Append(flex.Flatten(ctx, out, &state, flex.WithFieldNamePrefix("Filter"))...)
	if resp.Diagnostics.HasError() {
		return
	}

	// TIP: -- 6. Set the state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceFilter) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// TIP: ==== RESOURCE UPDATE ====
	// Not all resources have Update functions. There are a few reasons:
	// a. The AWS API does not support changing a resource
	// b. All arguments have RequiresReplace() plan modifiers
	// c. The AWS API uses a create call to modify an existing resource
	//
	// In the cases of a. and b., the resource will not have an update method
	// defined. In the case of c., Update and Create can be refactored to call
	// the same underlying function.
	//
	// The rest of the time, there should be an Update function and it should
	// do the following things. Make sure there is a good reason if you don't
	// do one of these.
	//
	// 1. Get a client connection to the relevant service
	// 2. Fetch the plan and state
	// 3. Populate a modify input structure and check for changes
	// 4. Call the AWS modify/update function
	// 5. Use a waiter to wait for update to complete
	// 6. Save the request plan to response state
	// TIP: -- 1. Get a client connection to the relevant service
	conn := r.Meta().Inspector2Client(ctx)

	// TIP: -- 2. Fetch the plan
	var plan, state resourceFilterModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// TIP: -- 3. Populate a modify input structure and check for changes
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

		// TIP: -- 4. Call the AWS modify/update function
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

		// TIP: Using the output from the update function, re-set any computed attributes
		resp.Diagnostics.Append(flex.Flatten(ctx, filter_out, &plan)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// TIP: -- 6. Save the request plan to response state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceFilter) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// TIP: ==== RESOURCE DELETE ====
	// Most resources have Delete functions. There are rare situations
	// where you might not need a delete:
	// a. The AWS API does not provide a way to delete the resource
	// b. The point of your resource is to perform an action (e.g., reboot a
	//    server) and deleting serves no purpose.
	//
	// The Delete function should do the following things. Make sure there
	// is a good reason if you don't do one of these.
	//
	// 1. Get a client connection to the relevant service
	// 2. Fetch the state
	// 3. Populate a delete input structure
	// 4. Call the AWS delete function
	// 5. Use a waiter to wait for delete to complete
	// TIP: -- 1. Get a client connection to the relevant service
	conn := r.Meta().Inspector2Client(ctx)

	// TIP: -- 2. Fetch the state
	var state resourceFilterModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// TIP: -- 3. Populate a delete input structure
	input := inspector2.DeleteFilterInput{
		Arn: state.ARN.ValueStringPointer(),
	}

	// TIP: -- 4. Call the AWS delete function
	_, err := conn.DeleteFilter(ctx, &input)
	// TIP: On rare occassions, the API returns a not found error after deleting a
	// resource. If that happens, we don't want it to show up as an error.
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

// TIP: ==== TERRAFORM IMPORTING ====
// If Read can get all the information it needs from the Identifier
// (i.e., path.Root("id")), you can use the PassthroughID importer. Otherwise,
// you'll need a custom import function.
//
// See more:
// https://developer.hashicorp.com/terraform/plugin/framework/resources/import
func (r *resourceFilter) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("arn"), req, resp)
}

// TIP: ==== FINDERS ====
// The find function is not strictly necessary. You could do the API
// request from the status function. However, we have found that find often
// comes in handy in other places besides the status function. As a result, it
// is good practice to define it separately.
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

// TIP: ==== DATA STRUCTURES ====
// With Terraform Plugin-Framework configurations are deserialized into
// Go types, providing type safety without the need for type assertions.
// These structs should match the schema definition exactly, and the `tfsdk`
// tag value should match the attribute name.
//
// Nested objects are represented in their own data struct. These will
// also have a corresponding attribute type mapping for use inside flex
// functions.
//
// See more:
// https://developer.hashicorp.com/terraform/plugin/framework/handling-data/accessing-values
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
	BeginInclusive types.Float64 `tfsdk:"begin_inclusive"`
	EndInclusive   types.Float64 `tfsdk:"end_inclusive"`
}

type mapFilterModel struct {
	Comparison fwtypes.StringEnum[awstypes.MapComparison] `tfsdk:"comparison"`
	Key        types.String                               `tfsdk:"key"`
	Value      types.String                               `tfsdk:"value"`
}

type packageFilterModel struct {
	Architecture         fwtypes.SetNestedObjectValueOf[stringFilterModel] `tfsdk:"architecture"`
	Epoch                fwtypes.SetNestedObjectValueOf[numberFilterModel] `tfsdk:"epoch"`
	FilePath             fwtypes.SetNestedObjectValueOf[stringFilterModel] `tfsdk:"file_path"`
	Name                 fwtypes.SetNestedObjectValueOf[stringFilterModel] `tfsdk:"name"`
	Release              fwtypes.SetNestedObjectValueOf[stringFilterModel] `tfsdk:"release"`
	SourceLambdaLayerARN fwtypes.SetNestedObjectValueOf[stringFilterModel] `tfsdk:"source_lambda_layer_arn"`
	SourceLayerHash      fwtypes.SetNestedObjectValueOf[stringFilterModel] `tfsdk:"source_layer_hash"`
	Version              fwtypes.SetNestedObjectValueOf[stringFilterModel] `tfsdk:"version"`
}
