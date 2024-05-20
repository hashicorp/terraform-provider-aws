// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package docdbelastic

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/docdbelastic"
	awstypes "github.com/aws/aws-sdk-go-v2/service/docdbelastic/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource(name="Cluster")
// @Tags(identifierAttribute="arn")
func newResourceCluster(context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceCluster{}
	r.SetDefaultCreateTimeout(45 * time.Minute)
	r.SetDefaultUpdateTimeout(45 * time.Minute)
	r.SetDefaultDeleteTimeout(45 * time.Minute)

	return r, nil
}

type resourceCluster struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
}

const (
	ResNameCluster = "Cluster"
)

func (r *resourceCluster) Metadata(_ context.Context, _ resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_docdbelastic_cluster"
}

func (r *resourceCluster) Schema(ctx context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	s := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"admin_user_name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"admin_user_password": schema.StringAttribute{
				Required:  true,
				Sensitive: true,
			},
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			"auth_type": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					enum.FrameworkValidate[awstypes.Auth](),
				},
			},
			names.AttrEndpoint: schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrID: framework.IDAttribute(),
			names.AttrKMSKeyID: schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrName: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrPreferredMaintenanceWindow: schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"shard_capacity": schema.Int64Attribute{
				Required: true,
			},
			"shard_count": schema.Int64Attribute{
				Required: true,
				Validators: []validator.Int64{
					int64validator.Between(1, 32),
				},
			},
			names.AttrSubnetIDs: schema.SetAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			names.AttrVPCSecurityGroupIDs: schema.SetAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}

	if s.Blocks == nil {
		s.Blocks = make(map[string]schema.Block)
	}
	s.Blocks[names.AttrTimeouts] = timeouts.Block(ctx, timeouts.Opts{
		Create: true,
		Update: true,
		Delete: true,
	})

	response.Schema = s
}

func (r *resourceCluster) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	conn := r.Meta().DocDBElasticClient(ctx)
	var plan resourceClusterData

	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)

	if response.Diagnostics.HasError() {
		return
	}

	input := &docdbelastic.CreateClusterInput{
		ClientToken:       aws.String(id.UniqueId()),
		AdminUserName:     flex.StringFromFramework(ctx, plan.AdminUserName),
		AdminUserPassword: flex.StringFromFramework(ctx, plan.AdminUserPassword),
		AuthType:          awstypes.Auth(plan.AuthType.ValueString()),
		ClusterName:       flex.StringFromFramework(ctx, plan.Name),
		ShardCapacity:     flex.Int32FromFramework(ctx, plan.ShardCapacity),
		ShardCount:        flex.Int32FromFramework(ctx, plan.ShardCount),
		Tags:              getTagsIn(ctx),
	}

	if !plan.KmsKeyID.IsNull() || !plan.KmsKeyID.IsUnknown() {
		input.KmsKeyId = flex.StringFromFramework(ctx, plan.KmsKeyID)
	}

	if !plan.PreferredMaintenanceWindow.IsNull() || !plan.PreferredMaintenanceWindow.IsUnknown() {
		input.PreferredMaintenanceWindow = flex.StringFromFramework(ctx, plan.PreferredMaintenanceWindow)
	}

	if !plan.SubnetIds.IsNull() || !plan.SubnetIds.IsUnknown() {
		input.SubnetIds = flex.ExpandFrameworkStringValueSet(ctx, plan.SubnetIds)
	}

	if !plan.VpcSecurityGroupIds.IsNull() || !plan.VpcSecurityGroupIds.IsUnknown() {
		input.VpcSecurityGroupIds = flex.ExpandFrameworkStringValueSet(ctx, plan.VpcSecurityGroupIds)
	}

	createOut, err := conn.CreateCluster(ctx, input)

	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DocDBElastic, create.ErrActionCreating, ResNameCluster, plan.Name.ValueString(), err),
			err.Error(),
		)
		return
	}

	state := plan
	state.ID = flex.StringToFramework(ctx, createOut.Cluster.ClusterArn)

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	out, err := waitClusterCreated(ctx, conn, state.ID.ValueString(), createTimeout)

	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DocDBElastic, create.ErrActionCreating, ResNameCluster, plan.Name.ValueString(), err),
			err.Error(),
		)
		return
	}

	state.refreshFromOutput(ctx, out)
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *resourceCluster) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	conn := r.Meta().DocDBElasticClient(ctx)
	var state resourceClusterData

	response.Diagnostics.Append(request.State.Get(ctx, &state)...)

	if response.Diagnostics.HasError() {
		return
	}

	out, err := findClusterByID(ctx, conn, state.ID.ValueString())

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DocDBElastic, create.ErrActionReading, ResNameCluster, state.ID.ValueString(), err),
			err.Error(),
		)
		return
	}

	state.refreshFromOutput(ctx, out)
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *resourceCluster) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	conn := r.Meta().DocDBElasticClient(ctx)
	var state, plan resourceClusterData

	response.Diagnostics.Append(request.State.Get(ctx, &state)...)

	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)

	if response.Diagnostics.HasError() {
		return
	}

	if clusterHasChanges(ctx, plan, state) {
		input := &docdbelastic.UpdateClusterInput{
			ClientToken: aws.String(id.UniqueId()),
			ClusterArn:  flex.StringFromFramework(ctx, state.ID),
		}

		if !plan.AdminUserPassword.Equal(state.AdminUserPassword) {
			input.AdminUserPassword = flex.StringFromFramework(ctx, plan.AdminUserPassword)
		}

		if !plan.AuthType.Equal(state.AuthType) {
			input.AuthType = awstypes.Auth(plan.AuthType.ValueString())
		}

		if !plan.PreferredMaintenanceWindow.Equal(state.PreferredMaintenanceWindow) {
			input.PreferredMaintenanceWindow = flex.StringFromFramework(ctx, plan.PreferredMaintenanceWindow)
		}

		if !plan.ShardCapacity.Equal(state.ShardCapacity) {
			input.ShardCapacity = flex.Int32FromFramework(ctx, plan.ShardCapacity)
		}

		if !plan.ShardCount.Equal(state.ShardCount) {
			input.ShardCount = flex.Int32FromFramework(ctx, plan.ShardCount)
		}

		if !plan.SubnetIds.Equal(state.SubnetIds) {
			input.SubnetIds = flex.ExpandFrameworkStringValueSet(ctx, plan.SubnetIds)
		}

		if !plan.VpcSecurityGroupIds.Equal(state.VpcSecurityGroupIds) {
			input.VpcSecurityGroupIds = flex.ExpandFrameworkStringValueSet(ctx, plan.VpcSecurityGroupIds)
		}

		_, err := conn.UpdateCluster(ctx, input)

		if err != nil {
			response.Diagnostics.AddError(
				create.ProblemStandardMessage(names.DocDBElastic, create.ErrActionUpdating, ResNameCluster, state.ID.ValueString(), err),
				err.Error(),
			)
			return
		}

		updateTimeout := r.UpdateTimeout(ctx, plan.Timeouts)
		out, err := waitClusterUpdated(ctx, conn, state.ID.ValueString(), updateTimeout)

		if err != nil {
			response.Diagnostics.AddError(
				create.ProblemStandardMessage(names.DocDBElastic, create.ErrActionWaitingForUpdate, ResNameCluster, state.ID.ValueString(), err),
				err.Error(),
			)
			return
		}

		plan.refreshFromOutput(ctx, out)
		response.Diagnostics.Append(response.State.Set(ctx, &plan)...)
	}

	response.Diagnostics.Append(response.State.Set(ctx, &plan)...)
}

func (r *resourceCluster) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	conn := r.Meta().DocDBElasticClient(ctx)
	var state resourceClusterData

	response.Diagnostics.Append(request.State.Get(ctx, &state)...)

	if response.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "deleting DocDB Elastic Cluster", map[string]interface{}{
		names.AttrID: state.ID.ValueString(),
	})

	input := &docdbelastic.DeleteClusterInput{
		ClusterArn: flex.StringFromFramework(ctx, state.ID),
	}

	_, err := conn.DeleteCluster(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DocDBElastic, create.ErrActionDeleting, ResNameCluster, state.ID.ValueString(), err),
			err.Error(),
		)
		return
	}

	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	_, err = waitClusterDeleted(ctx, conn, state.ID.ValueString(), deleteTimeout)

	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DocDBElastic, create.ErrActionWaitingForDeletion, ResNameCluster, state.ID.ValueString(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceCluster) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrID), request, response)
}

func (r *resourceCluster) ModifyPlan(ctx context.Context, request resource.ModifyPlanRequest, response *resource.ModifyPlanResponse) {
	r.SetTagsAll(ctx, request, response)
}

type resourceClusterData struct {
	AdminUserName              types.String   `tfsdk:"admin_user_name"`
	AdminUserPassword          types.String   `tfsdk:"admin_user_password"`
	ARN                        types.String   `tfsdk:"arn"`
	AuthType                   types.String   `tfsdk:"auth_type"`
	Endpoint                   types.String   `tfsdk:"endpoint"`
	ID                         types.String   `tfsdk:"id"`
	KmsKeyID                   types.String   `tfsdk:"kms_key_id"`
	Name                       types.String   `tfsdk:"name"`
	PreferredMaintenanceWindow types.String   `tfsdk:"preferred_maintenance_window"`
	ShardCapacity              types.Int64    `tfsdk:"shard_capacity"`
	ShardCount                 types.Int64    `tfsdk:"shard_count"`
	SubnetIds                  types.Set      `tfsdk:"subnet_ids"`
	Tags                       types.Map      `tfsdk:"tags"`
	TagsAll                    types.Map      `tfsdk:"tags_all"`
	Timeouts                   timeouts.Value `tfsdk:"timeouts"`
	VpcSecurityGroupIds        types.Set      `tfsdk:"vpc_security_group_ids"`
}

func waitClusterCreated(ctx context.Context, conn *docdbelastic.Client, id string, timeout time.Duration) (*awstypes.Cluster, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.StatusCreating),
		Target:                    enum.Slice(awstypes.StatusActive),
		Refresh:                   statusCluster(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.Cluster); ok {
		return out, err
	}

	return nil, err
}

func waitClusterUpdated(ctx context.Context, conn *docdbelastic.Client, id string, timeout time.Duration) (*awstypes.Cluster, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.StatusUpdating),
		Target:                    enum.Slice(awstypes.StatusActive),
		Refresh:                   statusCluster(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.Cluster); ok {
		return out, err
	}

	return nil, err
}

func waitClusterDeleted(ctx context.Context, conn *docdbelastic.Client, id string, timeout time.Duration) (*awstypes.Cluster, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.StatusActive, awstypes.StatusDeleting),
		Target:  []string{},
		Refresh: statusCluster(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.Cluster); ok {
		return out, err
	}

	return nil, err
}

func statusCluster(ctx context.Context, conn *docdbelastic.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := findClusterByID(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, string(out.Status), nil
	}
}

func findClusterByID(ctx context.Context, conn *docdbelastic.Client, id string) (*awstypes.Cluster, error) {
	in := &docdbelastic.GetClusterInput{
		ClusterArn: aws.String(id),
	}
	out, err := conn.GetCluster(ctx, in)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}

	if err != nil {
		return nil, err
	}

	if out == nil || out.Cluster == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out.Cluster, nil
}

func (r *resourceClusterData) refreshFromOutput(ctx context.Context, output *awstypes.Cluster) {
	r.AdminUserName = flex.StringToFrameworkLegacy(ctx, output.AdminUserName)
	r.AuthType = flex.StringValueToFramework(ctx, string(output.AuthType))
	r.ARN = flex.StringToFramework(ctx, output.ClusterArn)
	r.Endpoint = flex.StringToFramework(ctx, output.ClusterEndpoint)
	r.KmsKeyID = flex.StringToFramework(ctx, output.KmsKeyId)
	r.Name = flex.StringToFramework(ctx, output.ClusterName)
	r.PreferredMaintenanceWindow = flex.StringToFramework(ctx, output.PreferredMaintenanceWindow)
	r.ShardCapacity = flex.Int32ToFramework(ctx, output.ShardCapacity)
	r.ShardCount = flex.Int32ToFramework(ctx, output.ShardCount)
	r.SubnetIds = flex.FlattenFrameworkStringValueSet(ctx, output.SubnetIds)
	r.VpcSecurityGroupIds = flex.FlattenFrameworkStringValueSet(ctx, output.VpcSecurityGroupIds)
}

func clusterHasChanges(_ context.Context, plan, state resourceClusterData) bool {
	return !plan.Name.Equal(state.Name) ||
		!plan.AdminUserPassword.Equal(state.AdminUserPassword) ||
		!plan.AuthType.Equal(state.AuthType) ||
		!plan.PreferredMaintenanceWindow.Equal(state.PreferredMaintenanceWindow) ||
		!plan.ShardCapacity.Equal(state.ShardCapacity) ||
		!plan.ShardCount.Equal(state.ShardCount) ||
		!plan.SubnetIds.Equal(state.SubnetIds) ||
		!plan.VpcSecurityGroupIds.Equal(state.VpcSecurityGroupIds)
}
