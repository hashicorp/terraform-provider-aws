// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package quicksight

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/quicksight"
	awstypes "github.com/aws/aws-sdk-go-v2/service/quicksight/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_quicksight_namespace", name="Namespace")
// @Tags(identifierAttribute="arn")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/quicksight/types;awstypes;awstypes.NamespaceInfoV2")
// @Testing(skipEmptyTags=true, skipNullTags=true)
func newNamespaceResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &namespaceResource{}
	r.SetDefaultCreateTimeout(2 * time.Minute)
	r.SetDefaultDeleteTimeout(2 * time.Minute)

	return r, nil
}

const (
	resNameNamespace = "Namespace"
)

type namespaceResource struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
	framework.WithNoOpUpdate[resourceNamespaceData]
	framework.WithImportByID
}

func (r *namespaceResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
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
				Default:  stringdefault.StaticString(string(awstypes.IdentityStoreQuicksight)),
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

func (r *namespaceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().QuickSightClient(ctx)

	var plan resourceNamespaceData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if plan.AWSAccountID.IsUnknown() || plan.AWSAccountID.IsNull() {
		plan.AWSAccountID = types.StringValue(r.Meta().AccountID(ctx))
	}
	awsAccountID, namespace := flex.StringValueFromFramework(ctx, plan.AWSAccountID), flex.StringValueFromFramework(ctx, plan.Namespace)
	in := quicksight.CreateNamespaceInput{
		AwsAccountId:  aws.String(awsAccountID),
		IdentityStore: awstypes.IdentityStore(plan.IdentityStore.ValueString()),
		Namespace:     aws.String(namespace),
		Tags:          getTagsIn(ctx),
	}

	out, err := conn.CreateNamespace(ctx, &in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.QuickSight, create.ErrActionCreating, resNameNamespace, plan.Namespace.String(), err),
			err.Error(),
		)
		return
	}
	if out == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.QuickSight, create.ErrActionCreating, resNameNamespace, plan.Namespace.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	plan.ID = flex.StringValueToFramework(ctx, namespaceCreateResourceID(awsAccountID, namespace))

	waitOut, err := waitNamespaceCreated(ctx, conn, awsAccountID, namespace, r.CreateTimeout(ctx, plan.Timeouts))
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.QuickSight, create.ErrActionWaitingForCreation, resNameNamespace, plan.Namespace.String(), err),
			err.Error(),
		)
		return
	}
	plan.ARN = flex.StringToFramework(ctx, waitOut.Arn)
	plan.CapacityRegion = flex.StringToFramework(ctx, waitOut.CapacityRegion)
	plan.CreationStatus = flex.StringValueToFramework(ctx, waitOut.CreationStatus)
	plan.IdentityStore = flex.StringValueToFramework(ctx, waitOut.IdentityStore)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *namespaceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().QuickSightClient(ctx)

	var state resourceNamespaceData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	awsAccountID, namespace, err := namespaceParseResourceID(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.QuickSight, create.ErrActionReading, resNameNamespace, state.ID.String(), nil),
			err.Error(),
		)
		return
	}

	out, err := findNamespaceByTwoPartKey(ctx, conn, awsAccountID, namespace)
	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.QuickSight, create.ErrActionSetting, resNameNamespace, state.ID.String(), nil),
			err.Error(),
		)
		return
	}

	state.ARN = flex.StringToFramework(ctx, out.Arn)
	state.CapacityRegion = flex.StringToFramework(ctx, out.CapacityRegion)
	state.CreationStatus = flex.StringValueToFramework(ctx, out.CreationStatus)
	state.IdentityStore = flex.StringValueToFramework(ctx, out.IdentityStore)
	state.AWSAccountID = flex.StringValueToFramework(ctx, awsAccountID)
	state.Namespace = flex.StringValueToFramework(ctx, namespace)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *namespaceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().QuickSightClient(ctx)

	var state resourceNamespaceData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	awsAccountID, namespace, err := namespaceParseResourceID(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.QuickSight, create.ErrActionDeleting, resNameNamespace, state.ID.String(), nil),
			err.Error(),
		)
		return
	}

	_, err = conn.DeleteNamespace(ctx, &quicksight.DeleteNamespaceInput{
		AwsAccountId: aws.String(awsAccountID),
		Namespace:    aws.String(namespace),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.QuickSight, create.ErrActionDeleting, resNameNamespace, state.ID.String(), nil),
			err.Error(),
		)
	}

	_, err = waitNamespaceDeleted(ctx, conn, awsAccountID, namespace, r.DeleteTimeout(ctx, state.Timeouts))
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.QuickSight, create.ErrActionWaitingForDeletion, resNameNamespace, state.ID.String(), err),
			err.Error(),
		)
		return
	}
}

func findNamespaceByTwoPartKey(ctx context.Context, conn *quicksight.Client, awsAccountID, namespace string) (*awstypes.NamespaceInfoV2, error) {
	input := &quicksight.DescribeNamespaceInput{
		AwsAccountId: aws.String(awsAccountID),
		Namespace:    aws.String(namespace),
	}

	return findNamespace(ctx, conn, input)
}

func findNamespace(ctx context.Context, conn *quicksight.Client, input *quicksight.DescribeNamespaceInput) (*awstypes.NamespaceInfoV2, error) {
	output, err := conn.DescribeNamespace(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Namespace == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Namespace, nil
}

const namespaceResourceIDSeparator = ","

func namespaceCreateResourceID(awsAccountID, namespace string) string {
	parts := []string{awsAccountID, namespace}
	id := strings.Join(parts, namespaceResourceIDSeparator)

	return id
}

func namespaceParseResourceID(id string) (string, string, error) {
	parts := strings.SplitN(id, namespaceResourceIDSeparator, 3)

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format of ID (%[1]s), expected AWS_ACCOUNT_ID%[2]sNAMESPACE", id, namespaceResourceIDSeparator)
	}

	return parts[0], parts[1], nil
}

type resourceNamespaceData struct {
	ARN            types.String   `tfsdk:"arn"`
	AWSAccountID   types.String   `tfsdk:"aws_account_id"`
	CapacityRegion types.String   `tfsdk:"capacity_region"`
	CreationStatus types.String   `tfsdk:"creation_status"`
	ID             types.String   `tfsdk:"id"`
	IdentityStore  types.String   `tfsdk:"identity_store"`
	Namespace      types.String   `tfsdk:"namespace"`
	Tags           tftags.Map     `tfsdk:"tags"`
	TagsAll        tftags.Map     `tfsdk:"tags_all"`
	Timeouts       timeouts.Value `tfsdk:"timeouts"`
}

func waitNamespaceCreated(ctx context.Context, conn *quicksight.Client, awsAccountID, namespace string, timeout time.Duration) (*awstypes.NamespaceInfoV2, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(awstypes.NamespaceStatusCreating),
		Target:     enum.Slice(awstypes.NamespaceStatusCreated),
		Refresh:    statusNamespace(ctx, conn, awsAccountID, namespace),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.NamespaceInfoV2); ok {
		return output, err
	}

	return nil, err
}

func waitNamespaceDeleted(ctx context.Context, conn *quicksight.Client, awsAccountID, namespace string, timeout time.Duration) (*awstypes.NamespaceInfoV2, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(awstypes.NamespaceStatusDeleting),
		Target:     []string{},
		Refresh:    statusNamespace(ctx, conn, awsAccountID, namespace),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.NamespaceInfoV2); ok {
		return output, err
	}

	return nil, err
}

func statusNamespace(ctx context.Context, conn *quicksight.Client, awsAccountID, namespace string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findNamespaceByTwoPartKey(ctx, conn, awsAccountID, namespace)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.CreationStatus), nil
	}
}
