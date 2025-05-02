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
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32planmodifier"
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
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_docdbelastic_cluster", name="Cluster")
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
	framework.WithImportByID
}

const (
	ResNameCluster = "Cluster"
)

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
				CustomType: fwtypes.StringEnumType[awstypes.Auth](),
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"backup_retention_period": schema.Int32Attribute{
				Optional: true,
				Computed: true,
				Validators: []validator.Int32{
					int32validator.Between(1, 35),
				},
				PlanModifiers: []planmodifier.Int32{
					int32planmodifier.UseStateForUnknown(),
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
			"preferred_backup_window": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrPreferredMaintenanceWindow: schema.StringAttribute{
				CustomType: fwtypes.OnceAWeekWindowType,
				Optional:   true,
				Computed:   true,
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
				CustomType: fwtypes.SetOfStringType,
				Optional:   true,
				Computed:   true,
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			names.AttrVPCSecurityGroupIDs: schema.SetAttribute{
				CustomType: fwtypes.SetOfStringType,
				Optional:   true,
				Computed:   true,
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

	optionPrefix := fwflex.WithFieldNamePrefix("Cluster")
	input := docdbelastic.CreateClusterInput{}
	response.Diagnostics.Append(fwflex.Expand(ctx, plan, &input, optionPrefix)...)

	if response.Diagnostics.HasError() {
		return
	}
	input.ClientToken = aws.String(id.UniqueId())
	input.Tags = getTagsIn(ctx)

	createOut, err := conn.CreateCluster(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DocDBElastic, create.ErrActionCreating, ResNameCluster, plan.Name.ValueString(), err),
			err.Error(),
		)
		return
	}

	state := plan
	state.ID = fwflex.StringToFramework(ctx, createOut.Cluster.ClusterArn)

	// set partial state
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root(names.AttrID), createOut.Cluster.ClusterArn)...)

	if response.Diagnostics.HasError() {
		return
	}

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	out, err := waitClusterCreated(ctx, conn, state.ID.ValueString(), createTimeout)

	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DocDBElastic, create.ErrActionCreating, ResNameCluster, plan.Name.ValueString(), err),
			err.Error(),
		)
		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, out, &state, optionPrefix)...)

	if response.Diagnostics.HasError() {
		return
	}

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

	response.Diagnostics.Append(fwflex.Flatten(ctx, out, &state, fwflex.WithFieldNamePrefix("Cluster"))...)

	if response.Diagnostics.HasError() {
		return
	}

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

	diff, d := fwflex.Diff(ctx, plan, state)
	response.Diagnostics.Append(d...)
	if response.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		input := docdbelastic.UpdateClusterInput{}
		response.Diagnostics.Append(fwflex.Expand(ctx, plan, &input, diff.IgnoredFieldNamesOpts()...)...)
		if response.Diagnostics.HasError() {
			return
		}
		input.ClientToken = aws.String(id.UniqueId())
		input.ClusterArn = plan.ID.ValueStringPointer()

		_, err := conn.UpdateCluster(ctx, &input)

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

		response.Diagnostics.Append(fwflex.Flatten(ctx, out, &plan, fwflex.WithFieldNamePrefix("Cluster"))...)

		if response.Diagnostics.HasError() {
			return
		}

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

	tflog.Debug(ctx, "deleting DocDB Elastic Cluster", map[string]any{
		names.AttrID: state.ID.ValueString(),
	})

	input := &docdbelastic.DeleteClusterInput{
		ClusterArn: fwflex.StringFromFramework(ctx, state.ID),
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

type resourceClusterData struct {
	AdminUserName              types.String                      `tfsdk:"admin_user_name"`
	AdminUserPassword          types.String                      `tfsdk:"admin_user_password"`
	ARN                        types.String                      `tfsdk:"arn"`
	AuthType                   fwtypes.StringEnum[awstypes.Auth] `tfsdk:"auth_type"`
	BackupRetentionPeriod      types.Int32                       `tfsdk:"backup_retention_period"`
	Endpoint                   types.String                      `tfsdk:"endpoint"`
	ID                         types.String                      `tfsdk:"id"`
	KmsKeyID                   types.String                      `tfsdk:"kms_key_id"`
	Name                       types.String                      `tfsdk:"name"`
	PreferredBackupWindow      types.String                      `tfsdk:"preferred_backup_window"`
	PreferredMaintenanceWindow fwtypes.OnceAWeekWindow           `tfsdk:"preferred_maintenance_window"`
	ShardCapacity              types.Int64                       `tfsdk:"shard_capacity"`
	ShardCount                 types.Int64                       `tfsdk:"shard_count"`
	SubnetIds                  fwtypes.SetValueOf[types.String]  `tfsdk:"subnet_ids"`
	Tags                       tftags.Map                        `tfsdk:"tags"`
	TagsAll                    tftags.Map                        `tfsdk:"tags_all"`
	Timeouts                   timeouts.Value                    `tfsdk:"timeouts"`
	VpcSecurityGroupIds        fwtypes.SetValueOf[types.String]  `tfsdk:"vpc_security_group_ids"`
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
	return func() (any, string, error) {
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
