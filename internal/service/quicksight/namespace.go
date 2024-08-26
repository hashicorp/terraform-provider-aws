// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package quicksight

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/quicksight"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
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

// @FrameworkResource(name="Namespace")
// @Tags(identifierAttribute="arn")
func newResourceNamespace(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceNamespace{}
	r.SetDefaultCreateTimeout(2 * time.Minute)
	r.SetDefaultDeleteTimeout(2 * time.Minute)

	return r, nil
}

const (
	ResNameNamespace = "Namespace"
)

type resourceNamespace struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
}

func (r *resourceNamespace) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_quicksight_namespace"
}

func (r *resourceNamespace) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrAWSAccountID: schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"capacity_region": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"creation_status": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrID: framework.IDAttribute(),
			"identity_store": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString(quicksight.IdentityStoreQuicksight),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrNamespace: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Delete: true,
			}),
		},
	}
}

func (r *resourceNamespace) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().QuickSightConn(ctx)

	var plan resourceNamespaceData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if plan.AWSAccountID.IsUnknown() || plan.AWSAccountID.IsNull() {
		plan.AWSAccountID = types.StringValue(r.Meta().AccountID)
	}
	plan.ID = types.StringValue(createNamespaceID(plan.AWSAccountID.ValueString(), plan.Namespace.ValueString()))

	in := quicksight.CreateNamespaceInput{
		AwsAccountId:  aws.String(plan.AWSAccountID.ValueString()),
		Namespace:     aws.String(plan.Namespace.ValueString()),
		IdentityStore: aws.String(plan.IdentityStore.ValueString()),
		Tags:          getTagsIn(ctx),
	}

	out, err := conn.CreateNamespaceWithContext(ctx, &in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.QuickSight, create.ErrActionCreating, ResNameNamespace, plan.Namespace.String(), err),
			err.Error(),
		)
		return
	}
	if out == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.QuickSight, create.ErrActionCreating, ResNameNamespace, plan.Namespace.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	waitOut, err := waitNamespaceCreated(ctx, conn, plan.ID.ValueString(), createTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.QuickSight, create.ErrActionWaitingForCreation, ResNameNamespace, plan.Namespace.String(), err),
			err.Error(),
		)
		return
	}
	plan.ARN = flex.StringToFramework(ctx, waitOut.Arn)
	plan.CapacityRegion = flex.StringToFramework(ctx, waitOut.CapacityRegion)
	plan.CreationStatus = flex.StringToFramework(ctx, waitOut.CreationStatus)
	plan.IdentityStore = flex.StringToFramework(ctx, waitOut.IdentityStore)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceNamespace) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().QuickSightConn(ctx)

	var state resourceNamespaceData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := FindNamespaceByID(ctx, conn, state.ID.ValueString())
	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.QuickSight, create.ErrActionSetting, ResNameNamespace, state.ID.String(), nil),
			err.Error(),
		)
		return
	}

	state.ARN = flex.StringToFramework(ctx, out.Arn)
	state.CapacityRegion = flex.StringToFramework(ctx, out.CapacityRegion)
	state.CreationStatus = flex.StringToFramework(ctx, out.CreationStatus)
	state.IdentityStore = flex.StringToFramework(ctx, out.IdentityStore)

	// To support import, parse the ID for the component keys and set
	// individual values in state
	awsAccountID, namespace, err := ParseNamespaceID(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.QuickSight, create.ErrActionSetting, ResNameNamespace, state.ID.String(), nil),
			err.Error(),
		)
		return
	}
	state.AWSAccountID = flex.StringValueToFramework(ctx, awsAccountID)
	state.Namespace = flex.StringValueToFramework(ctx, namespace)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceNamespace) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// There is no update API, and tag updates are handled via a "before"
	// interceptor. Copy the planned tag attributes to state to ensure
	// updates are captured.
	var plan resourceNamespaceData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceNamespace) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().QuickSightConn(ctx)

	var state resourceNamespaceData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := conn.DeleteNamespaceWithContext(ctx, &quicksight.DeleteNamespaceInput{
		AwsAccountId: aws.String(state.AWSAccountID.ValueString()),
		Namespace:    aws.String(state.Namespace.ValueString()),
	})
	if err != nil {
		if tfawserr.ErrCodeEquals(err, quicksight.ErrCodeResourceNotFoundException) {
			return
		}
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.QuickSight, create.ErrActionDeleting, ResNameNamespace, state.ID.String(), nil),
			err.Error(),
		)
	}

	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	_, err = waitNamespaceDeleted(ctx, conn, state.ID.ValueString(), deleteTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.QuickSight, create.ErrActionWaitingForDeletion, ResNameNamespace, state.ID.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceNamespace) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrID), req, resp)
}

func (r *resourceNamespace) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	r.SetTagsAll(ctx, req, resp)
}

func FindNamespaceByID(ctx context.Context, conn *quicksight.QuickSight, id string) (*quicksight.NamespaceInfoV2, error) {
	awsAccountID, namespace, err := ParseNamespaceID(id)
	if err != nil {
		return nil, err
	}

	in := &quicksight.DescribeNamespaceInput{
		AwsAccountId: aws.String(awsAccountID),
		Namespace:    aws.String(namespace),
	}

	out, err := conn.DescribeNamespaceWithContext(ctx, in)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, quicksight.ErrCodeResourceNotFoundException) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return nil, err
	}

	if out == nil || out.Namespace == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out.Namespace, nil
}

func ParseNamespaceID(id string) (string, string, error) {
	parts := strings.SplitN(id, ",", 3)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format of ID (%s), expected AWS_ACCOUNT_ID,NAMESPACE", id)
	}
	return parts[0], parts[1], nil
}

func createNamespaceID(awsAccountID, namespace string) string {
	return fmt.Sprintf("%s,%s", awsAccountID, namespace)
}

type resourceNamespaceData struct {
	ARN            types.String   `tfsdk:"arn"`
	AWSAccountID   types.String   `tfsdk:"aws_account_id"`
	CapacityRegion types.String   `tfsdk:"capacity_region"`
	CreationStatus types.String   `tfsdk:"creation_status"`
	ID             types.String   `tfsdk:"id"`
	IdentityStore  types.String   `tfsdk:"identity_store"`
	Namespace      types.String   `tfsdk:"namespace"`
	Tags           types.Map      `tfsdk:"tags"`
	TagsAll        types.Map      `tfsdk:"tags_all"`
	Timeouts       timeouts.Value `tfsdk:"timeouts"`
}

func waitNamespaceCreated(ctx context.Context, conn *quicksight.QuickSight, id string, timeout time.Duration) (*quicksight.NamespaceInfoV2, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{
			quicksight.NamespaceStatusCreating,
		},
		Target: []string{
			quicksight.NamespaceStatusCreated,
		},
		Refresh:    statusNamespace(ctx, conn, id),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if output, ok := outputRaw.(*quicksight.NamespaceInfoV2); ok {
		return output, err
	}

	return nil, err
}

func waitNamespaceDeleted(ctx context.Context, conn *quicksight.QuickSight, id string, timeout time.Duration) (*quicksight.NamespaceInfoV2, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{
			quicksight.NamespaceStatusDeleting,
		},
		Target:     []string{},
		Refresh:    statusNamespace(ctx, conn, id),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if output, ok := outputRaw.(*quicksight.NamespaceInfoV2); ok {
		return output, err
	}

	return nil, err
}

func statusNamespace(ctx context.Context, conn *quicksight.QuickSight, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindNamespaceByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.CreationStatus), nil
	}
}
