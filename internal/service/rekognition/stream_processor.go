// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rekognition

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
	// using the services/rekognition/types package. If so, you'll
	// need to import types and reference the nested types, e.g., as
	// awstypes.<Type Name>.
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rekognition"
	awstypes "github.com/aws/aws-sdk-go-v2/service/rekognition/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
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
// @FrameworkResource("aws_rekognition_stream_processor", name="Stream Processor")
func newResourceStreamProcessor(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceStreamProcessor{}

	// TIP: ==== CONFIGURABLE TIMEOUTS ====
	// Users can configure timeout lengths but you need to use the times they
	// provide. Access the timeout they configure (or the defaults) using,
	// e.g., r.CreateTimeout(ctx, plan.Timeouts) (see below). The times here are
	// the defaults if they don't configure timeouts.
	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameStreamProcessor = "Stream Processor"
)

type resourceStreamProcessor struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
}

func (r *resourceStreamProcessor) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_rekognition_stream_processor"
}

func (r *resourceStreamProcessor) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"arn": framework.ARNAttributeComputedOnly(),
			"description": schema.StringAttribute{
				Optional: true,
			},
			"id": framework.IDAttribute(),
			"name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"type": schema.StringAttribute{
				Required: true,
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		// Blocks: map[string]schema.Block{
		// 	"complex_argument": schema.ListNestedBlock{
		// 		// TIP: ==== LIST VALIDATORS ====
		// 		// List and set validators take the place of MaxItems and MinItems in
		// 		// Plugin-Framework based resources. Use listvalidator.SizeAtLeast(1) to
		// 		// make a nested object required. Similar to Plugin-SDK, complex objects
		// 		// can be represented as lists or sets with listvalidator.SizeAtMost(1).
		// 		//
		// 		// For a complete mapping of Plugin-SDK to Plugin-Framework schema fields,
		// 		// see:
		// 		// https://developer.hashicorp.com/terraform/plugin/framework/migrating/attributes-blocks/blocks
		// 		Validators: []validator.List{
		// 			listvalidator.SizeAtMost(1),
		// 		},
		// 		NestedObject: schema.NestedBlockObject{
		// 			Attributes: map[string]schema.Attribute{
		// 				"nested_required": schema.StringAttribute{
		// 					Required: true,
		// 				},
		// 				"nested_computed": schema.StringAttribute{
		// 					Computed: true,
		// 					PlanModifiers: []planmodifier.String{
		// 						stringplanmodifier.UseStateForUnknown(),
		// 					},
		// 				},
		// 			},
		// 		},
		// 	},
		// 	"timeouts": timeouts.Block(ctx, timeouts.Opts{
		// 		Create: true,
		// 		Update: true,
		// 		Delete: true,
		// 	}),
		// },
	}
}

func (r *resourceStreamProcessor) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
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
	conn := r.Meta().RekognitionClient(ctx)

	// TIP: -- 2. Fetch the plan
	var plan resourceStreamProcessorData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// TIP: -- 3. Populate a create input structure
	in := &rekognition.CreateStreamProcessorInput{
		Name: aws.String(plan.Name.ValueString()),
	}

	// if !plan.Description.IsNull() {
	// 	// TIP: Optional fields should be set based on whether or not they are
	// 	// used.
	// 	in.Description = aws.String(plan.Description.ValueString())
	// }
	// if !plan.ComplexArgument.IsNull() {
	// 	// TIP: Use an expander to assign a complex argument. The elements must be
	// 	// deserialized into the appropriate struct before being passed to the expander.
	// 	var tfList []complexArgumentData
	// 	resp.Diagnostics.Append(plan.ComplexArgument.ElementsAs(ctx, &tfList, false)...)
	// 	if resp.Diagnostics.HasError() {
	// 		return
	// 	}

	// 	in.ComplexArgument = expandComplexArgument(tfList)
	// }

	// TIP: -- 4. Call the AWS create function
	out, err := conn.CreateStreamProcessor(ctx, in)
	if err != nil {
		// TIP: Since ID has not been set yet, you cannot use plan.ID.String()
		// in error messages at this point.
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Rekognition, create.ErrActionCreating, ResNameStreamProcessor, plan.Name.String(), err),
			err.Error(),
		)
		return
	}
	if out == nil || out.StreamProcessorArn == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Rekognition, create.ErrActionCreating, ResNameStreamProcessor, plan.Name.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	// TIP: -- 5. Using the output from the create function, set the minimum attributes
	plan.ARN = flex.StringToFramework(ctx, out.StreamProcessorArn)

	// TIP: -- 6. Use a waiter to wait for create to complete
	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	_, err = waitStreamProcessorCreated(ctx, conn, plan.Name.ValueString(), createTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Rekognition, create.ErrActionWaitingForCreation, ResNameStreamProcessor, plan.Name.String(), err),
			err.Error(),
		)
		return
	}

	// TIP: -- 7. Save the request plan to response state
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceStreamProcessor) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().RekognitionClient(ctx)

	// TIP: -- 2. Fetch the state
	var state resourceStreamProcessorData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findStreamProcessorByID(ctx, conn, state.Name.ValueString())
	// TIP: -- 4. Remove resource from state if it is not found
	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Rekognition, create.ErrActionSetting, ResNameStreamProcessor, state.Name.String(), err),
			err.Error(),
		)
		return
	}

	state.Name = flex.StringToFramework(ctx, out.Name)

	// TIP: Setting a complex type.
	// complexArgument, d := flattenComplexArgument(ctx, out.ComplexArgument)
	// resp.Diagnostics.Append(d...)
	// state.ComplexArgument = complexArgument

	// TIP: -- 6. Set the state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceStreamProcessor) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().RekognitionClient(ctx)

	// TIP: -- 2. Fetch the plan
	var plan, state resourceStreamProcessorData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// TIP: -- 3. Populate a modify input structure and check for changes
	if !plan.Name.Equal(state.Name) {

		in := &rekognition.UpdateStreamProcessorInput{
			// TIP: Mandatory or fields that will always be present can be set when
			// you create the Input structure. (Replace these with real fields.)
			Name: aws.String(plan.Name.ValueString()),
		}

		// TIP: -- 4. Call the AWS modify/update function
		_, err := conn.UpdateStreamProcessor(ctx, in)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.Rekognition, create.ErrActionUpdating, ResNameStreamProcessor, plan.Name.String(), err),
				err.Error(),
			)
			return
		}

		// we have to call describe to get the new values

		// if out == nil || out.ResultMetadata == nil {
		// 	resp.Diagnostics.AddError(
		// 		create.ProblemStandardMessage(names.Rekognition, create.ErrActionUpdating, ResNameStreamProcessor, plan.Name.String(), nil),
		// 		errors.New("empty output").Error(),
		// 	)
		// 	return
		// }

		// TIP: Using the output from the update function, re-set any computed attributes
		// plan.ARN = flex.StringToFramework(ctx, out.Arn)
		// plan.ID = flex.StringToFramework(ctx, out.StreamProcessor.StreamProcessorId)
	}

	// TIP: -- 5. Use a waiter to wait for update to complete
	updateTimeout := r.UpdateTimeout(ctx, plan.Timeouts)
	_, err := waitStreamProcessorUpdated(ctx, conn, plan.Name.ValueString(), updateTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Rekognition, create.ErrActionWaitingForUpdate, ResNameStreamProcessor, plan.Name.String(), err),
			err.Error(),
		)
		return
	}

	// TIP: -- 6. Save the request plan to response state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceStreamProcessor) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().RekognitionClient(ctx)

	// TIP: -- 2. Fetch the state
	var state resourceStreamProcessorData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// TIP: -- 3. Populate a delete input structure
	in := &rekognition.DeleteStreamProcessorInput{
		Name: aws.String(state.Name.ValueString()),
	}

	_, err := conn.DeleteStreamProcessor(ctx, in)

	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Rekognition, create.ErrActionDeleting, ResNameStreamProcessor, state.Name.String(), err),
			err.Error(),
		)
		return
	}

	// TIP: -- 5. Use a waiter to wait for delete to complete
	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	_, err = waitStreamProcessorDeleted(ctx, conn, state.Name.ValueString(), deleteTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Rekognition, create.ErrActionWaitingForDeletion, ResNameStreamProcessor, state.Name.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceStreamProcessor) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func waitStreamProcessorCreated(ctx context.Context, conn *rekognition.Client, id string, timeout time.Duration) (*rekognition.DescribeStreamProcessorOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{},
		Target: []string{
			string(awstypes.StreamProcessorStatusStarting),
			string(awstypes.StreamProcessorStatusRunning),
			string(awstypes.StreamProcessorStatusFailed)},
		Refresh:                   statusStreamProcessor(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*rekognition.DescribeStreamProcessorOutput); ok {
		return out, err
	}

	return nil, err
}

func waitStreamProcessorUpdated(ctx context.Context, conn *rekognition.Client, id string, timeout time.Duration) (*rekognition.DescribeStreamProcessorOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{string(awstypes.StreamProcessorStatusUpdating)},
		Target: []string{
			string(awstypes.StreamProcessorStatusStarting),
			string(awstypes.StreamProcessorStatusRunning),
			string(awstypes.StreamProcessorStatusFailed)},
		Refresh:                   statusStreamProcessor(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*rekognition.DescribeStreamProcessorOutput); ok {
		return out, err
	}

	return nil, err
}

func waitStreamProcessorDeleted(ctx context.Context, conn *rekognition.Client, id string, timeout time.Duration) (*rekognition.DescribeStreamProcessorOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{
			string(awstypes.StreamProcessorStatusStopped),
			string(awstypes.StreamProcessorStatusStarting),
			string(awstypes.StreamProcessorStatusRunning),
			string(awstypes.StreamProcessorStatusFailed),
			string(awstypes.StreamProcessorStatusStopping),
			string(awstypes.StreamProcessorStatusUpdating),
		},
		Target:  []string{},
		Refresh: statusStreamProcessor(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*rekognition.DescribeStreamProcessorOutput); ok {
		return out, err
	}

	return nil, err
}

func statusStreamProcessor(ctx context.Context, conn *rekognition.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := findStreamProcessorByID(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, aws.ToString((*string)(&out.Status)), nil
	}
}

func findStreamProcessorByID(ctx context.Context, conn *rekognition.Client, name string) (*rekognition.DescribeStreamProcessorOutput, error) {
	in := &rekognition.DescribeStreamProcessorInput{
		Name: aws.String(name),
	}

	out, err := conn.DescribeStreamProcessor(ctx, in)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return nil, err
	}

	if out == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
}

type resourceStreamProcessorData struct {
	ARN      types.String   `tfsdk:"arn"`
	Name     types.String   `tfsdk:"name"`
	Tags     types.Map      `tfsdk:"tags"`
	TagsAll  types.Map      `tfsdk:"tags_all"`
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}
