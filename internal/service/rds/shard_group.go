// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds

import (
	"context"
	"fmt"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	awstypes "github.com/aws/aws-sdk-go-v2/service/rds/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/float64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_rds_shard_group", name="Shard Group")
// @Tags(identifierAttribute="arn")
// @Testing(tagsTest=false)
func newShardGroupResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &shardGroupResource{}

	r.SetDefaultCreateTimeout(45 * time.Minute)
	r.SetDefaultUpdateTimeout(45 * time.Minute)
	r.SetDefaultDeleteTimeout(45 * time.Minute)

	return r, nil
}

type shardGroupResource struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
}

func (r *shardGroupResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			"compute_redundancy": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(0),
				Validators: []validator.Int64{
					int64validator.Between(0, 2),
				},
			},
			"db_cluster_identifier": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"db_shard_group_identifier": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 63),
					stringvalidator.RegexMatches(regexache.MustCompile(`[a-zA-Z][a-zA-Z0-9]*(-[a-zA-Z0-9]+)*`), ""),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"db_shard_group_resource_id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrEndpoint: schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"max_acu": schema.Float64Attribute{
				Required: true,
			},
			"min_acu": schema.Float64Attribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Float64{
					float64planmodifier.UseStateForUnknown(),
				},
			},
			names.AttrPubliclyAccessible: schema.BoolAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

func (r *shardGroupResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data shardGroupResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().RDSClient(ctx)

	input := rds.CreateDBShardGroupInput{}
	response.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	input.Tags = getTagsIn(ctx)

	_, err := conn.CreateDBShardGroup(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating RDS Shard Group (%s)", data.DBShardGroupIdentifier.ValueString()), err.Error())

		return
	}

	deadline := tfresource.NewDeadline(r.CreateTimeout(ctx, data.Timeouts))
	shardGroup, err := waitShardGroupCreated(ctx, conn, data.DBShardGroupIdentifier.ValueString(), deadline.Remaining())

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for RDS Shard Group (%s) create", data.DBShardGroupIdentifier.ValueString()), err.Error())

		return
	}

	// Set values for unknowns.
	response.Diagnostics.Append(fwflex.Flatten(ctx, shardGroup, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	if _, err := waitDBClusterUpdated(ctx, conn, data.DBClusterIdentifier.ValueString(), false, deadline.Remaining()); err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for RDS Cluster (%s) update", data.DBClusterIdentifier.ValueString()), err.Error())

		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *shardGroupResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data shardGroupResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().RDSClient(ctx)

	output, err := findDBShardGroupByID(ctx, conn, data.DBShardGroupIdentifier.ValueString())

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading RDS Shard Group (%s)", data.DBShardGroupIdentifier.ValueString()), err.Error())

		return
	}

	// Set attributes for import.
	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	setTagsOut(ctx, output.TagList)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *shardGroupResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var old, new shardGroupResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.State.Get(ctx, &old)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().RDSClient(ctx)

	if !new.ComputeRedundancy.Equal(old.ComputeRedundancy) || !new.MaxACU.Equal(old.MaxACU) || !new.MinACU.Equal(old.MinACU) {
		input := rds.ModifyDBShardGroupInput{}
		response.Diagnostics.Append(fwflex.Expand(ctx, new, &input)...)
		if response.Diagnostics.HasError() {
			return
		}

		_, err := conn.ModifyDBShardGroup(ctx, &input)

		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("updating RDS Shard Group (%s)", new.DBShardGroupIdentifier.ValueString()), err.Error())

			return
		}

		deadline := tfresource.NewDeadline(r.UpdateTimeout(ctx, new.Timeouts))
		if _, err := waitShardGroupUpdated(ctx, conn, new.DBShardGroupIdentifier.ValueString(), deadline.Remaining()); err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("waiting for RDS Shard Group (%s) update", new.DBShardGroupIdentifier.ValueString()), err.Error())

			return
		}

		if _, err := waitDBClusterUpdated(ctx, conn, new.DBClusterIdentifier.ValueString(), false, deadline.Remaining()); err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("waiting for RDS Cluster (%s) update", new.DBClusterIdentifier.ValueString()), err.Error())

			return
		}
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *shardGroupResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data shardGroupResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().RDSClient(ctx)

	_, err := conn.DeleteDBShardGroup(ctx, &rds.DeleteDBShardGroupInput{
		DBShardGroupIdentifier: fwflex.StringFromFramework(ctx, data.DBShardGroupIdentifier),
	})

	if errs.IsA[*awstypes.DBShardGroupNotFoundFault](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting RDS Shard Group (%s)", data.DBShardGroupIdentifier.ValueString()), err.Error())

		return
	}

	deadline := tfresource.NewDeadline(r.DeleteTimeout(ctx, data.Timeouts))
	if _, err := waitShardGroupDeleted(ctx, conn, data.DBShardGroupIdentifier.ValueString(), deadline.Remaining()); err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for RDS Shard Group (%s) delete", data.DBShardGroupIdentifier.ValueString()), err.Error())

		return
	}

	if _, err := waitDBClusterUpdated(ctx, conn, data.DBClusterIdentifier.ValueString(), false, deadline.Remaining()); err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for RDS Cluster (%s) update", data.DBClusterIdentifier.ValueString()), err.Error())

		return
	}
}

func (r *shardGroupResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("db_shard_group_identifier"), request, response)
}

func findDBShardGroupByID(ctx context.Context, conn *rds.Client, id string) (*awstypes.DBShardGroup, error) {
	input := rds.DescribeDBShardGroupsInput{
		DBShardGroupIdentifier: aws.String(id),
	}

	return findDBShardGroup(ctx, conn, &input, tfslices.PredicateTrue[*awstypes.DBShardGroup]())
}

func findDBShardGroup(ctx context.Context, conn *rds.Client, input *rds.DescribeDBShardGroupsInput, filter tfslices.Predicate[*awstypes.DBShardGroup]) (*awstypes.DBShardGroup, error) {
	output, err := findDBShardGroups(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findDBShardGroups(ctx context.Context, conn *rds.Client, input *rds.DescribeDBShardGroupsInput, filter tfslices.Predicate[*awstypes.DBShardGroup]) ([]awstypes.DBShardGroup, error) {
	var output []awstypes.DBShardGroup

	err := describeDBShardGroupsPages(ctx, conn, input, func(page *rds.DescribeDBShardGroupsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.DBShardGroups {
			if filter(&v) {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if errs.IsA[*awstypes.DBShardGroupNotFoundFault](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func statusShardGroup(ctx context.Context, conn *rds.Client, id string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findDBShardGroupByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.ToString(output.Status), nil
	}
}

const (
	shardGroupStatusAvailable = "available"
	shardGroupStatusCreating  = "creating"
	shardGroupStatusDeleting  = "deleting"
	shardGroupStatusModifying = "modifying"
)

func waitShardGroupCreated(ctx context.Context, conn *rds.Client, id string, timeout time.Duration) (*awstypes.DBShardGroup, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{shardGroupStatusCreating},
		Target:  []string{shardGroupStatusAvailable},
		Refresh: statusShardGroup(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.DBShardGroup); ok {
		return output, err
	}

	return nil, err
}

func waitShardGroupUpdated(ctx context.Context, conn *rds.Client, id string, timeout time.Duration) (*awstypes.DBShardGroup, error) {
	const (
		delay = 1 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending: []string{shardGroupStatusModifying},
		Target:  []string{shardGroupStatusAvailable},
		Refresh: statusShardGroup(ctx, conn, id),
		Delay:   delay,
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.DBShardGroup); ok {
		return output, err
	}

	return nil, err
}

func waitShardGroupDeleted(ctx context.Context, conn *rds.Client, id string, timeout time.Duration) (*awstypes.DBShardGroup, error) {
	const (
		delay = 1 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending: []string{shardGroupStatusDeleting},
		Target:  []string{},
		Refresh: statusShardGroup(ctx, conn, id),
		Delay:   delay,
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.DBShardGroup); ok {
		return output, err
	}

	return nil, err
}

type shardGroupResourceModel struct {
	ComputeRedundancy      types.Int64    `tfsdk:"compute_redundancy"`
	DBClusterIdentifier    types.String   `tfsdk:"db_cluster_identifier"`
	DBShardGroupARN        types.String   `tfsdk:"arn"`
	DBShardGroupIdentifier types.String   `tfsdk:"db_shard_group_identifier"`
	DBShardGroupResourceID types.String   `tfsdk:"db_shard_group_resource_id"`
	Endpoint               types.String   `tfsdk:"endpoint"`
	MaxACU                 types.Float64  `tfsdk:"max_acu"`
	MinACU                 types.Float64  `tfsdk:"min_acu"`
	PubliclyAccessible     types.Bool     `tfsdk:"publicly_accessible"`
	Tags                   tftags.Map     `tfsdk:"tags"`
	TagsAll                tftags.Map     `tfsdk:"tags_all"`
	Timeouts               timeouts.Value `tfsdk:"timeouts"`
}
