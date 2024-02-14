// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elasticache

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/elasticache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/elasticache/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource(name="Serverless Cache")
// @Tags(identifierAttribute="arn")
func newResourceServerlessCache(context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceServerlessCache{}
	r.SetDefaultCreateTimeout(40 * time.Minute)
	r.SetDefaultUpdateTimeout(80 * time.Minute)
	r.SetDefaultDeleteTimeout(40 * time.Minute)

	return r, nil
}

const (
	ResNameServerlessCache = "Serverless Cache"
)

type resourceServerlessCache struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
}

func (r *resourceServerlessCache) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_elasticache_serverless_cache"
}

func (r *resourceServerlessCache) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	s := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"arn": framework.ARNAttributeComputedOnly(),
			"create_time": schema.StringAttribute{
				CustomType: fwtypes.TimestampType,
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"daily_snapshot_time": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"description": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"endpoint": schema.ListAttribute{
				CustomType:  fwtypes.NewListNestedObjectTypeOf[endpoint](ctx),
				ElementType: fwtypes.NewObjectTypeOf[endpoint](ctx),
				Computed:    true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"engine": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"full_engine_version": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"id": framework.IDAttribute(),
			"kms_key_id": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"major_engine_version": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"reader_endpoint": schema.ListAttribute{
				CustomType:  fwtypes.NewListNestedObjectTypeOf[endpoint](ctx),
				ElementType: fwtypes.NewObjectTypeOf[endpoint](ctx),
				Computed:    true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"security_group_ids": schema.SetAttribute{
				CustomType:  fwtypes.SetOfStringType,
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
			},
			"snapshot_arns_to_restore": schema.ListAttribute{
				CustomType:  fwtypes.ListOfARNType,
				ElementType: fwtypes.ARNType,
				Optional:    true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
			},
			"snapshot_retention_limit": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Validators: []validator.Int64{
					int64validator.AtMost(35),
				},
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"status": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"subnet_ids": schema.SetAttribute{
				CustomType:  fwtypes.SetOfStringType,
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
					setplanmodifier.RequiresReplace(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			"user_group_id": schema.StringAttribute{
				Optional: true,
			},
		},
		Blocks: map[string]schema.Block{
			"cache_usage_limits": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[cacheUsageLimits](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"data_storage": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[dataStorage](ctx),
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"maximum": schema.Int64Attribute{
										Required: true,
									},
									"unit": schema.StringAttribute{
										CustomType: fwtypes.StringEnumType[awstypes.DataStorageUnit](),
										Required:   true,
									},
								},
							},
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
						},
						"ecpu_per_second": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[ecpuPerSecond](ctx),
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"maximum": schema.Int64Attribute{
										Required: true,
									},
								},
							},
						},
					},
				},
			},
		},
	}

	if s.Blocks == nil {
		s.Blocks = make(map[string]schema.Block)
	}
	s.Blocks["timeouts"] = timeouts.Block(ctx, timeouts.Opts{
		Create: true,
		Update: true,
		Delete: true,
	})

	response.Schema = s
}

func (r *resourceServerlessCache) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	conn := r.Meta().ElastiCacheClient(ctx)
	var plan resourceServerlessData

	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)

	if response.Diagnostics.HasError() {
		return
	}

	input := &elasticache.CreateServerlessCacheInput{}
	response.Diagnostics.Append(flex.Expand(ctx, plan, input)...)

	if response.Diagnostics.HasError() {
		return
	}

	input.ServerlessCacheName = flex.StringFromFramework(ctx, plan.Name)
	input.Tags = getTagsInV2(ctx)

	output, err := conn.CreateServerlessCache(ctx, input)

	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ElastiCache, create.ErrActionCreating, ResNameServerlessCache, plan.Name.ValueString(), err),
			err.Error(),
		)
		return
	}

	state := plan
	state.ID = flex.StringToFramework(ctx, output.ServerlessCache.ServerlessCacheName)

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	out, err := waitServerlessCacheAvailable(ctx, conn, aws.ToString(output.ServerlessCache.ServerlessCacheName), createTimeout)

	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ElastiCache, create.ErrActionWaitingForCreation, ResNameServerlessCache, plan.Name.ValueString(), err),
			err.Error(),
		)
		return
	}

	response.Diagnostics.Append(flex.Flatten(ctx, out, &state)...)

	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *resourceServerlessCache) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	conn := r.Meta().ElastiCacheClient(ctx)
	var state resourceServerlessData

	response.Diagnostics.Append(request.State.Get(ctx, &state)...)

	if response.Diagnostics.HasError() {
		return
	}

	out, err := FindServerlessCacheByID(ctx, conn, state.ID.ValueString())

	if tfresource.NotFound(err) {
		response.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ElastiCache, create.ErrActionSetting, ResNameServerlessCache, state.Name.ValueString(), err),
			err.Error(),
		)
		return
	}

	response.Diagnostics.Append(flex.Flatten(ctx, out, &state)...)

	if response.Diagnostics.HasError() {
		return
	}

	state.Name = flex.StringToFramework(ctx, out.ServerlessCacheName)

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *resourceServerlessCache) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	conn := r.Meta().ElastiCacheClient(ctx)
	var state, plan resourceServerlessData

	response.Diagnostics.Append(request.State.Get(ctx, &state)...)

	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)

	if response.Diagnostics.HasError() {
		return
	}

	if serverlessCacheHasChanges(ctx, plan, state) {
		input := &elasticache.ModifyServerlessCacheInput{}

		response.Diagnostics.Append(flex.Expand(ctx, plan, input)...)

		if response.Diagnostics.HasError() {
			return
		}

		input.ServerlessCacheName = flex.StringFromFramework(ctx, state.Name)

		_, err := conn.ModifyServerlessCache(ctx, input)

		if err != nil {
			response.Diagnostics.AddError(
				create.ProblemStandardMessage(names.ElastiCache, create.ErrActionUpdating, ResNameServerlessCache, state.Name.ValueString(), err),
				err.Error(),
			)
			return
		}

		updateTimeout := r.UpdateTimeout(ctx, plan.Timeouts)
		_, err = waitServerlessCacheAvailable(ctx, conn, state.Name.ValueString(), updateTimeout)

		if err != nil {
			response.Diagnostics.AddError(
				create.ProblemStandardMessage(names.ElastiCache, create.ErrActionWaitingForUpdate, ResNameServerlessCache, plan.Name.ValueString(), err),
				err.Error(),
			)
			return
		}
	}

	// AWS returns null values for certain values that are available on redis only.
	// always set these values to the state value to avoid unnecessary diff failures on computed values.
	out, err := FindServerlessCacheByID(ctx, conn, state.ID.ValueString())

	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ElastiCache, create.ErrActionUpdating, ResNameServerlessCache, state.Name.ValueString(), err),
			err.Error(),
		)
		return
	}

	response.Diagnostics.Append(flex.Flatten(ctx, out, &plan)...)

	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &plan)...)
}

func (r *resourceServerlessCache) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	conn := r.Meta().ElastiCacheClient(ctx)
	var state resourceServerlessData

	response.Diagnostics.Append(request.State.Get(ctx, &state)...)

	if response.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "deleting ElastiCache Serverless Cache", map[string]interface{}{
		"id": state.ID.ValueString(),
	})

	input := &elasticache.DeleteServerlessCacheInput{
		ServerlessCacheName: flex.StringFromFramework(ctx, state.ID),
		FinalSnapshotName:   nil,
	}

	_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, 5*time.Minute, func() (interface{}, error) {
		return conn.DeleteServerlessCache(ctx, input)
	}, "DependencyViolation")

	if errs.IsA[*awstypes.ServerlessCacheNotFoundFault](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ElastiCache, create.ErrActionDeleting, ResNameServerlessCache, state.ID.ValueString(), err),
			err.Error(),
		)
		return
	}

	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	_, err = waitServerlessCacheDeleted(ctx, conn, state.ID.ValueString(), deleteTimeout)

	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ElastiCache, create.ErrActionWaitingForDeletion, ResNameServerlessCache, state.ID.ValueString(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceServerlessCache) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

func (r *resourceServerlessCache) ModifyPlan(ctx context.Context, request resource.ModifyPlanRequest, response *resource.ModifyPlanResponse) {
	r.SetTagsAll(ctx, request, response)
}

type resourceServerlessData struct {
	ARN                    types.String                                      `tfsdk:"arn"`
	CacheUsageLimits       fwtypes.ListNestedObjectValueOf[cacheUsageLimits] `tfsdk:"cache_usage_limits"`
	CreateTime             fwtypes.Timestamp                                 `tfsdk:"create_time"`
	DailySnapshotTime      types.String                                      `tfsdk:"daily_snapshot_time"`
	Description            types.String                                      `tfsdk:"description"`
	Endpoint               fwtypes.ListNestedObjectValueOf[endpoint]         `tfsdk:"endpoint"`
	Engine                 types.String                                      `tfsdk:"engine"`
	FullEngineVersion      types.String                                      `tfsdk:"full_engine_version"`
	ID                     types.String                                      `tfsdk:"id"`
	KmsKeyId               types.String                                      `tfsdk:"kms_key_id"`
	MajorEngineVersion     types.String                                      `tfsdk:"major_engine_version"`
	Name                   types.String                                      `tfsdk:"name"`
	ReaderEndpoint         fwtypes.ListNestedObjectValueOf[endpoint]         `tfsdk:"reader_endpoint"`
	SecurityGroupIds       fwtypes.SetValueOf[types.String]                  `tfsdk:"security_group_ids"`
	SnapshotARNsToRestore  fwtypes.ListValueOf[fwtypes.ARN]                  `tfsdk:"snapshot_arns_to_restore"`
	SnapshotRetentionLimit types.Int64                                       `tfsdk:"snapshot_retention_limit"`
	Status                 types.String                                      `tfsdk:"status"`
	SubnetIds              fwtypes.SetValueOf[types.String]                  `tfsdk:"subnet_ids"`
	Tags                   types.Map                                         `tfsdk:"tags"`
	TagsAll                types.Map                                         `tfsdk:"tags_all"`
	Timeouts               timeouts.Value                                    `tfsdk:"timeouts"`
	UserGroupID            types.String                                      `tfsdk:"user_group_id"`
}

type cacheUsageLimits struct {
	DataStorage   fwtypes.ListNestedObjectValueOf[dataStorage]   `tfsdk:"data_storage"`
	ECPUPerSecond fwtypes.ListNestedObjectValueOf[ecpuPerSecond] `tfsdk:"ecpu_per_second"`
}

type dataStorage struct {
	Maximum types.Int64                                  `tfsdk:"maximum"`
	Unit    fwtypes.StringEnum[awstypes.DataStorageUnit] `tfsdk:"unit"`
}

type ecpuPerSecond struct {
	Maximum types.Int64 `tfsdk:"maximum"`
}

type endpoint struct {
	Address types.String `tfsdk:"address"`
	Port    types.Int64  `tfsdk:"port"`
}

func serverlessCacheHasChanges(_ context.Context, plan, state resourceServerlessData) bool {
	return !plan.CacheUsageLimits.Equal(state.CacheUsageLimits) ||
		!plan.DailySnapshotTime.Equal(state.DailySnapshotTime) ||
		!plan.Description.Equal(state.Description) ||
		!plan.UserGroupID.Equal(state.UserGroupID) ||
		!plan.SecurityGroupIds.Equal(state.SecurityGroupIds) ||
		!plan.SnapshotRetentionLimit.Equal(state.SnapshotRetentionLimit)
}
