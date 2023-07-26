package kafka

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
    "github.com/aws/aws-sdk-go-v2/service/kafka"
	awstypes "github.com/aws/aws-sdk-go-v2/service/kafka/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for resource registration to the Provider. DO NOT EDIT.
// @FrameworkResource(name="Vpc Connection")
// @Tags(identifierAttribute="arn")
func newResourceVpcConnection(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceVpcConnection{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameVpcConnection = "Vpc Connection"
)

type resourceVpcConnection struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
}

func (r *resourceVpcConnection) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_msk_vpc_connection"
}

func (r *resourceVpcConnection) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"arn": framework.ARNAttributeComputedOnly(),
			"id": framework.IDAttribute(),
			
			"authentication": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"client_subnets": schema.SetAttribute{
				Required:    true,
				ElementType: types.StringType,
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.RequiresReplace(),
				},
			},
			"security_groups": schema.SetAttribute{
				Optional:    true,
				ElementType: types.StringType,
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.RequiresReplace(),
				},
			},
			"vpc_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"target_cluster_arn": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},

			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
	}
}

func (r *resourceVpcConnection) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().KafkaClient(ctx)
	
	var data resourceVpcConnectionData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	
	in := &kafka.CreateVpcConnectionInput{
		Authentication: aws.String(data.Authentication.ValueString()),
		ClientSubnets: flex.ExpandFrameworkStringValueSet(ctx, data.ClientSubnets),
		SecurityGroups: flex.ExpandFrameworkStringValueSet(ctx, data.SecurityGroups),
		TargetClusterArn: aws.String(data.TargetClusterArn.ValueString()),
		VpcId: aws.String(data.VpcId.ValueString()),
	}

	out, err := conn.CreateVpcConnection(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Kafka, create.ErrActionCreating, ResNameVpcConnection, data.ID.String(), err),
			err.Error(),
		)
		return
	}
	if out == nil || out.VpcConnectionArn == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Kafka, create.ErrActionCreating, ResNameVpcConnection, data.ARN.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}
	
	// TIP: -- 5. Using the output from the create function, set the minimum attributes
	data.ARN = flex.StringToFramework(ctx, out.VpcConnection.Arn)
	data.ID = flex.StringToFramework(ctx, out.VpcConnection.VpcConnectionId)
	
	// TIP: -- 6. Use a waiter to wait for create to complete
	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	_, err = waitVpcConnectionCreated(ctx, conn, plan.ID.ValueString(), createTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Kafka, create.ErrActionWaitingForCreation, ResNameVpcConnection, plan.Name.String(), err),
			err.Error(),
		)
		return
	}
	
	// TIP: -- 7. Save the request plan to response state
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceVpcConnection) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
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
	conn := r.Meta().KafkaClient(ctx)
	
	// TIP: -- 2. Fetch the state
	var state resourceVpcConnectionData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	
	// TIP: -- 3. Get the resource from AWS using an API Get, List, or Describe-
	// type function, or, better yet, using a finder.
	out, err := findVpcConnectionByID(ctx, conn, state.ID.ValueString())
	// TIP: -- 4. Remove resource from state if it is not found
	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Kafka, create.ErrActionSetting, ResNameVpcConnection, state.ID.String(), err),
			err.Error(),
		)
		return
	}
	
	// TIP: -- 5. Set the arguments and attributes
	//
	// For simple data types (i.e., schema.StringAttribute, schema.BoolAttribute,
	// schema.Int64Attribute, and schema.Float64Attribue), simply setting the  
	// appropriate data struct field is sufficient. The flex package implements
	// helpers for converting between Go and Plugin-Framework types seamlessly. No 
	// error or nil checking is necessary.
	//
	// However, there are some situations where more handling is needed such as
	// complex data types (e.g., schema.ListAttribute, schema.SetAttribute). In 
	// these cases the flatten function may have a diagnostics return value, which
	// should be appended to resp.Diagnostics.
	state.ARN = flex.StringToFramework(ctx, out.Arn)
	state.ID = flex.StringToFramework(ctx, out.VpcConnectionId)
	state.Name = flex.StringToFramework(ctx, out.VpcConnectionName)
	state.Type = flex.StringToFramework(ctx, out.VpcConnectionType)
	
	// TIP: Setting a complex type.
	complexArgument, d := flattenComplexArgument(ctx, out.ComplexArgument)
	resp.Diagnostics.Append(d...)
	state.ComplexArgument = complexArgument
	
	// TIP: -- 6. Set the state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceVpcConnection) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
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
	conn := r.Meta().KafkaClient(ctx)
	
	// TIP: -- 2. Fetch the plan
	var plan, state resourceVpcConnectionData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	
	// TIP: -- 3. Populate a modify input structure and check for changes
	if !plan.Name.Equal(state.Name) ||
		!plan.Description.Equal(state.Description) ||
		!plan.ComplexArgument.Equal(state.ComplexArgument) ||
		!plan.Type.Equal(state.Type) {

		in := &kafka.UpdateVpcConnectionInput{
			// TIP: Mandatory or fields that will always be present can be set when
			// you create the Input structure. (Replace these with real fields.)
			VpcConnectionId:   aws.String(plan.ID.ValueString()),
			VpcConnectionName: aws.String(plan.Name.ValueString()),
			VpcConnectionType: aws.String(plan.Type.ValueString()),
		}

		if !plan.Description.IsNull() {
			// TIP: Optional fields should be set based on whether or not they are
			// used.
			in.Description = aws.String(plan.Description.ValueString())
		}
		if !plan.ComplexArgument.IsNull() {
			// TIP: Use an expander to assign a complex argument. The elements must be
			// deserialized into the appropriate struct before being passed to the expander.
			var tfList []complexArgumentData
			resp.Diagnostics.Append(plan.ComplexArgument.ElementsAs(ctx, &tfList, false)...)
			if resp.Diagnostics.HasError() {
				return
			}

			in.ComplexArgument = expandComplexArgument(tfList)
		}
		
		// TIP: -- 4. Call the AWS modify/update function
		out, err := conn.UpdateVpcConnection(ctx, in)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.Kafka, create.ErrActionUpdating, ResNameVpcConnection, plan.ID.String(), err),
				err.Error(),
			)
			return
		}
		if out == nil || out.VpcConnection == nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.Kafka, create.ErrActionUpdating, ResNameVpcConnection, plan.ID.String(), nil),
				errors.New("empty output").Error(),
			)
			return
		}
		
		// TIP: Using the output from the update function, re-set any computed attributes
		plan.ARN = flex.StringToFramework(ctx, out.VpcConnection.Arn)
		plan.ID = flex.StringToFramework(ctx, out.VpcConnection.VpcConnectionId)
	}

	
	// TIP: -- 5. Use a waiter to wait for update to complete
	updateTimeout := r.UpdateTimeout(ctx, plan.Timeouts)
	_, err := waitVpcConnectionUpdated(ctx, conn, plan.ID.ValueString(), updateTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Kafka, create.ErrActionWaitingForUpdate, ResNameVpcConnection, plan.ID.String(), err),
			err.Error(),
		)
		return
	}

	
	// TIP: -- 6. Save the request plan to response state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceVpcConnection) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
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
	conn := r.Meta().KafkaClient(ctx)
	
	// TIP: -- 2. Fetch the state
	var state resourceVpcConnectionData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	
	// TIP: -- 3. Populate a delete input structure
	in := &kafka.DeleteVpcConnectionInput{
		VpcConnectionId: aws.String(state.ID.ValueString()),
	}
	
	// TIP: -- 4. Call the AWS delete function
	_, err := conn.DeleteVpcConnection(ctx, in)
	// TIP: On rare occassions, the API returns a not found error after deleting a
	// resource. If that happens, we don't want it to show up as an error.
	if err != nil {
		var nfe *awstypes.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return
		}
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Kafka, create.ErrActionDeleting, ResNameVpcConnection, state.ID.String(), err),
			err.Error(),
		)
		return
	}
	
	// TIP: -- 5. Use a waiter to wait for delete to complete
	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	_, err = waitVpcConnectionDeleted(ctx, conn, state.ID.ValueString(), deleteTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Kafka, create.ErrActionWaitingForDeletion, ResNameVpcConnection, state.ID.String(), err),
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
func (r *resourceVpcConnection) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// TIP: ==== STATUS CONSTANTS ====
// Create constants for states and statuses if the service does not
// already have suitable constants. We prefer that you use the constants
// provided in the service if available (e.g., amp.WorkspaceStatusCodeActive).
const (
	statusChangePending = "Pending"
	statusDeleting      = "Deleting"
	statusNormal        = "Normal"
	statusUpdated       = "Updated"
)

// TIP: ==== WAITERS ====
// Some resources of some services have waiters provided by the AWS API.
// Unless they do not work properly, use them rather than defining new ones
// here.
//
// Sometimes we define the wait, status, and find functions in separate
// files, wait.go, status.go, and find.go. Follow the pattern set out in the
// service and define these where it makes the most sense.
//
// If these functions are used in the _test.go file, they will need to be
// exported (i.e., capitalized).
//
// You will need to adjust the parameters and names to fit the service.
func waitVpcConnectionCreated(ctx context.Context, conn *kafka.Client, id string, timeout time.Duration) (*kafka.VpcConnection, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{},
		Target:                    []string{statusNormal},
		Refresh:                   statusVpcConnection(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*kafka.CreateVpcConnectionInput); ok {
		return out, err
	}

	return nil, err
}

// TIP: It is easier to determine whether a resource is updated for some
// resources than others. The best case is a status flag that tells you when
// the update has been fully realized. Other times, you can check to see if a
// key resource argument is updated to a new value or not.
func waitVpcConnectionUpdated(ctx context.Context, conn *kafka.Client, id string, timeout time.Duration) (*kafka.VpcConnection, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{statusChangePending},
		Target:                    []string{statusUpdated},
		Refresh:                   statusVpcConnection(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*kafka.VpcConnection); ok {
		return out, err
	}

	return nil, err
}

// TIP: A deleted waiter is almost like a backwards created waiter. There may
// be additional pending states, however.
func waitVpcConnectionDeleted(ctx context.Context, conn *kafka.Client, id string, timeout time.Duration) (*kafka.VpcConnection, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{statusDeleting, statusNormal},
		Target:                    []string{},
		Refresh:                   statusVpcConnection(ctx, conn, id),
		Timeout:                   timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*kafka.VpcConnection); ok {
		return out, err
	}

	return nil, err
}

// TIP: ==== STATUS ====
// The status function can return an actual status when that field is
// available from the API (e.g., out.Status). Otherwise, you can use custom
// statuses to communicate the states of the resource.
//
// Waiters consume the values returned by status functions. Design status so
// that it can be reused by a create, update, and delete waiter, if possible.
func statusVpcConnection(ctx context.Context, conn *kafka.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := findVpcConnectionByID(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, aws.ToString(out.Status), nil
	}
}

// TIP: ==== FINDERS ====
// The find function is not strictly necessary. You could do the API
// request from the status function. However, we have found that find often
// comes in handy in other places besides the status function. As a result, it
// is good practice to define it separately.
func findVpcConnectionByID(ctx context.Context, conn *kafka.Client, id string) (*kafka.VpcConnection, error) {
	in := &kafka.GetVpcConnectionInput{
		Id: aws.String(id),
	}
	
	out, err := conn.GetVpcConnection(ctx, in)
	if err != nil {
		var nfe *awstypes.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return nil, err
	}

	if out == nil || out.VpcConnection == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out.VpcConnection, nil
}

// TIP: ==== FLEX ====
// Flatteners and expanders ("flex" functions) help handle complex data
// types. Flatteners take an API data type and return the equivalent Plugin-Framework 
// type. In other words, flatteners translate from AWS -> Terraform.
//
// On the other hand, expanders take a Terraform data structure and return
// something that you can send to the AWS API. In other words, expanders
// translate from Terraform -> AWS.
//
// See more:
// https://hashicorp.github.io/terraform-provider-aws/data-handling-and-conversion/
func flattenComplexArgument(ctx context.Context, apiObject *kafka.ComplexArgument) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	elemType := types.ObjectType{AttrTypes: complexArgumentAttrTypes}

	if apiObject == nil {
		return types.ListNull(elemType), diags
	}

	obj := map[string]attr.Value{
		"nested_required": flex.StringValueToFramework(ctx, apiObject.NestedRequired),
		"nested_optional": flex.StringValueToFramework(ctx, apiObject.NestedOptional),
	}
	objVal, d := types.ObjectValue(complexArgumentAttrTypes, obj)
	diags.Append(d...)

	listVal, d := types.ListValue(elemType, []attr.Value{objVal})
	diags.Append(d...)

	return listVal, diags
}

type resourceVpcConnectionData struct {
	ARN             types.String   `tfsdk:"arn"`
	ID              types.String   `tfsdk:"id"`
	Authentication            types.String   `tfsdk:"authentication"`
	ClientSubnets           types.Set      `tfsdk:"client_subnets"`
	SecurityGroups           types.Set      `tfsdk:"security_groups"`	
	TargetClusterArn	types.String	`tfsdk:"target_cluster_arn"`
	Tags            types.Map      `tfsdk:"tags"`
	TagsAll         types.Map      `tfsdk:"tags_all"`
	Timeouts        timeouts.Value `tfsdk:"timeouts"`
	VpcId                      types.String   `tfsdk:"vpc_id"`
}
