// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3

import (
	"context"
	"fmt"
	"strconv"

	awstypes "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func bucketLifeCycleConfigurationSchemaV0(ctx context.Context) schema.Schema {
	return schema.Schema{
		Version: 0,
		Attributes: map[string]schema.Attribute{
			names.AttrBucket: schema.StringAttribute{
				Required: true,
			},
			names.AttrExpectedBucketOwner: schema.StringAttribute{
				Optional: true,
			},
			names.AttrID: framework.IDAttribute(),
			"transition_default_minimum_object_size": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.TransitionDefaultMinimumObjectSize](),
				Optional:   true,
				Computed:   true,
			},
		},
		Blocks: map[string]schema.Block{
			names.AttrRule: schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[lifecycleRuleModelV0](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrID: schema.StringAttribute{
							Required: true,
						},
						names.AttrPrefix: schema.StringAttribute{
							Optional: true,
						},
						names.AttrStatus: schema.StringAttribute{
							Required: true,
						},
					},
					Blocks: map[string]schema.Block{
						"abort_incomplete_multipart_upload": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[abortIncompleteMultipartUploadModelV0](ctx),
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"days_after_initiation": schema.Int64Attribute{
										Optional: true,
									},
								},
							},
						},
						"expiration": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[lifecycleExpirationModelV0](ctx),
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"date": schema.StringAttribute{
										Optional: true,
									},
									"days": schema.Int64Attribute{
										Optional: true,
										Computed: true,
										Default:  int64default.StaticInt64(0),
									},
									"expired_object_delete_marker": schema.BoolAttribute{
										Optional: true,
										Computed: true,
									},
								},
							},
						},
						names.AttrFilter: schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[lifecycleRuleFilterModelV0](ctx),
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"object_size_greater_than": schema.StringAttribute{
										Optional: true,
									},
									"object_size_less_than": schema.StringAttribute{
										Optional: true,
									},
									names.AttrPrefix: schema.StringAttribute{
										Optional: true,
									},
								},
								Blocks: map[string]schema.Block{
									"and": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[lifecycleRuleAndOperatorModelV0](ctx),
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"object_size_greater_than": schema.Int64Attribute{
													Optional: true,
												},
												"object_size_less_than": schema.Int64Attribute{
													Optional: true,
												},
												names.AttrPrefix: schema.StringAttribute{
													Optional: true,
												},
												names.AttrTags: schema.MapAttribute{
													ElementType: types.StringType,
													Optional:    true,
												},
											},
										},
									},
									"tag": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[tagModel](ctx),
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												names.AttrKey: schema.StringAttribute{
													Required: true,
												},
												names.AttrValue: schema.StringAttribute{
													Required: true,
												},
											},
										},
									},
								},
							},
						},
						"noncurrent_version_expiration": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[noncurrentVersionExpirationModelV0](ctx),
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"newer_noncurrent_versions": schema.StringAttribute{
										Optional: true,
									},
									"noncurrent_days": schema.Int64Attribute{
										Optional: true,
									},
								},
							},
						},
						"noncurrent_version_transition": schema.SetNestedBlock{
							CustomType: fwtypes.NewSetNestedObjectTypeOf[noncurrentVersionTransitionModelV0](ctx),
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"newer_noncurrent_versions": schema.StringAttribute{
										Optional: true,
									},
									"noncurrent_days": schema.Int64Attribute{
										Optional: true,
									},
									names.AttrStorageClass: schema.StringAttribute{
										CustomType: fwtypes.StringEnumType[awstypes.TransitionStorageClass](),
										Required:   true,
									},
								},
							},
						},
						"transition": schema.SetNestedBlock{
							CustomType: fwtypes.NewSetNestedObjectTypeOf[transitionModelV0](ctx),
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"date": schema.StringAttribute{
										Optional: true,
									},
									"days": schema.Int64Attribute{
										Optional: true,
									},
									names.AttrStorageClass: schema.StringAttribute{
										CustomType: fwtypes.StringEnumType[awstypes.TransitionStorageClass](),
										Required:   true,
									},
								},
							},
						},
					},
				},
			},
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
			}),
		},
	}
}

type resourceBucketLifecycleConfigurationModelV0 struct {
	Bucket                             types.String                                                    `tfsdk:"bucket"`
	ExpectedBucketOwner                types.String                                                    `tfsdk:"expected_bucket_owner" autoflex:",legacy"`
	ID                                 types.String                                                    `tfsdk:"id"`
	Rules                              fwtypes.ListNestedObjectValueOf[lifecycleRuleModelV0]           `tfsdk:"rule"`
	TransitionDefaultMinimumObjectSize fwtypes.StringEnum[awstypes.TransitionDefaultMinimumObjectSize] `tfsdk:"transition_default_minimum_object_size" autoflex:",legacy"`
	Timeouts                           timeouts.Value                                                  `tfsdk:"timeouts"`
}

func upgradeBucketLifeCycleConfigurationResourceStateFromV0(ctx context.Context, request resource.UpgradeStateRequest, response *resource.UpgradeStateResponse) {
	var old resourceBucketLifecycleConfigurationModelV0
	response.Diagnostics.Append(request.State.Get(ctx, &old)...)
	if response.Diagnostics.HasError() {
		return
	}

	new := resourceBucketLifecycleConfigurationModel{
		Bucket:                             old.Bucket,
		ExpectedBucketOwner:                old.ExpectedBucketOwner,
		ID:                                 old.ID,
		Rules:                              upgradeLifecycleRuleModelStateFromV0(ctx, old.Rules, &response.Diagnostics),
		TransitionDefaultMinimumObjectSize: old.TransitionDefaultMinimumObjectSize,
		Timeouts:                           old.Timeouts,
	}
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, new)...)
}

type lifecycleRuleModelV0 struct {
	AbortIncompleteMultipartUpload fwtypes.ListNestedObjectValueOf[abortIncompleteMultipartUploadModelV0] `tfsdk:"abort_incomplete_multipart_upload"`
	Expiration                     fwtypes.ListNestedObjectValueOf[lifecycleExpirationModelV0]            `tfsdk:"expiration"`
	Filter                         fwtypes.ListNestedObjectValueOf[lifecycleRuleFilterModelV0]            `tfsdk:"filter"`
	ID                             types.String                                                           `tfsdk:"id"`
	NoncurrentVersionExpirations   fwtypes.ListNestedObjectValueOf[noncurrentVersionExpirationModelV0]    `tfsdk:"noncurrent_version_expiration"`
	NoncurrentVersionTransitions   fwtypes.SetNestedObjectValueOf[noncurrentVersionTransitionModelV0]     `tfsdk:"noncurrent_version_transition"`
	Prefix                         types.String                                                           `tfsdk:"prefix"`
	Status                         fwtypes.StringEnum[awstypes.ExpirationStatus]                          `tfsdk:"status"`
	Transitions                    fwtypes.SetNestedObjectValueOf[transitionModelV0]                      `tfsdk:"transition"`
}

func upgradeLifecycleRuleModelStateFromV0(ctx context.Context, old fwtypes.ListNestedObjectValueOf[lifecycleRuleModelV0], diags *diag.Diagnostics) (result fwtypes.ListNestedObjectValueOf[lifecycleRuleModel]) {
	oldRules, d := old.ToSlice(ctx)
	diags.Append(d...)
	if diags.HasError() {
		return result
	}

	newRules := make([]lifecycleRuleModel, len(oldRules))
	for i, oldRule := range oldRules {
		newRule := lifecycleRuleModel{
			AbortIncompleteMultipartUpload: upgradeAbortIncompleteMultipartStateFromV0(ctx, oldRule.AbortIncompleteMultipartUpload, diags),
			Expiration:                     upgradeLifecycleExpirationModelStateFromV0(ctx, oldRule.Expiration, diags),
			Filter:                         upgradeLifecycleRuleFilterModelStateFromV0(ctx, oldRule.Filter, diags),
			ID:                             oldRule.ID,
			NoncurrentVersionExpirations:   upgradeNoncurrentVersionExpirationModelStateFromV0(ctx, oldRule.NoncurrentVersionExpirations, diags),
			NoncurrentVersionTransitions:   upgradeNoncurrentVersionTransitionModelStateFromV0(ctx, oldRule.NoncurrentVersionTransitions, diags),
			Prefix:                         oldRule.Prefix,
			Status:                         oldRule.Status,
			Transitions:                    upgradeTransitionModelStateFromV0(ctx, oldRule.Transitions, diags),
		}

		if diags.HasError() {
			return result
		}

		newRules[i] = newRule
	}

	result, d = fwtypes.NewListNestedObjectValueOfValueSlice(ctx, newRules)
	diags.Append(d...)
	if diags.HasError() {
		return result
	}

	return result
}

type abortIncompleteMultipartUploadModelV0 struct {
	DaysAfterInitiation types.Int64 `tfsdk:"days_after_initiation"`
}

// Single
func upgradeAbortIncompleteMultipartStateFromV0(ctx context.Context, old fwtypes.ListNestedObjectValueOf[abortIncompleteMultipartUploadModelV0], diags *diag.Diagnostics) (result fwtypes.ListNestedObjectValueOf[abortIncompleteMultipartUploadModel]) {
	oldThings, d := old.ToSlice(ctx)
	diags.Append(d...)
	if diags.HasError() {
		return result
	}

	newThings := make([]abortIncompleteMultipartUploadModel, len(oldThings))
	for i, oldThing := range oldThings {
		newThing := abortIncompleteMultipartUploadModel{
			DaysAfterInitiation: types.Int32Value(fwflex.Int32ValueFromFrameworkInt64(ctx, oldThing.DaysAfterInitiation)),
		}

		newThings[i] = newThing
	}

	result, d = fwtypes.NewListNestedObjectValueOfValueSlice(ctx, newThings)
	diags.Append(d...)
	if diags.HasError() {
		return result
	}

	return result
}

type lifecycleExpirationModelV0 struct {
	Date                      types.String `tfsdk:"date"`
	Days                      types.Int64  `tfsdk:"days"`
	ExpiredObjectDeleteMarker types.Bool   `tfsdk:"expired_object_delete_marker"`
}

// Single
func upgradeLifecycleExpirationModelStateFromV0(ctx context.Context, old fwtypes.ListNestedObjectValueOf[lifecycleExpirationModelV0], diags *diag.Diagnostics) (result fwtypes.ListNestedObjectValueOf[lifecycleExpirationModel]) {
	oldExpirations, d := old.ToSlice(ctx)
	diags.Append(d...)
	if diags.HasError() {
		return result
	}

	newThings := make([]lifecycleExpirationModel, len(oldExpirations))
	for i, oldExpiration := range oldExpirations {
		newExpiration := lifecycleExpirationModel{
			Date:                      migrateDate(ctx, oldExpiration.Date, diags),
			Days:                      int64ToInt32(ctx, oldExpiration.Days),
			ExpiredObjectDeleteMarker: oldExpiration.ExpiredObjectDeleteMarker,
		}
		if diags.HasError() {
			return result
		}

		newThings[i] = newExpiration
	}

	result, d = fwtypes.NewListNestedObjectValueOfValueSlice(ctx, newThings)
	diags.Append(d...)
	if diags.HasError() {
		return result
	}

	return result
}

type lifecycleRuleFilterModelV0 struct {
	And                   fwtypes.ListNestedObjectValueOf[lifecycleRuleAndOperatorModelV0] `tfsdk:"and"`
	ObjectSizeGreaterThan types.String                                                     `tfsdk:"object_size_greater_than"`
	ObjectSizeLessThan    types.String                                                     `tfsdk:"object_size_less_than"`
	Prefix                types.String                                                     `tfsdk:"prefix"`
	Tag                   fwtypes.ListNestedObjectValueOf[tagModel]                        `tfsdk:"tag"`
}

// Single
func upgradeLifecycleRuleFilterModelStateFromV0(ctx context.Context, old fwtypes.ListNestedObjectValueOf[lifecycleRuleFilterModelV0], diags *diag.Diagnostics) (result fwtypes.ListNestedObjectValueOf[lifecycleRuleFilterModel]) {
	oldFilters, d := old.ToSlice(ctx)
	diags.Append(d...)
	if diags.HasError() {
		return result
	}

	newFilters := make([]lifecycleRuleFilterModel, len(oldFilters))
	for i, oldFilter := range oldFilters {
		newFilter := lifecycleRuleFilterModel{
			And:                   upgradeLifecycleRuleAndOperatorModelStateFromV0(ctx, oldFilter.And, diags),
			ObjectSizeGreaterThan: stringToInt64Legacy(ctx, oldFilter.ObjectSizeGreaterThan, diags),
			ObjectSizeLessThan:    stringToInt64Legacy(ctx, oldFilter.ObjectSizeLessThan, diags),
			Prefix:                oldFilter.Prefix,
			Tag:                   oldFilter.Tag,
		}
		if diags.HasError() {
			return result
		}

		newFilters[i] = newFilter
	}

	result, d = fwtypes.NewListNestedObjectValueOfValueSlice(ctx, newFilters)
	diags.Append(d...)
	if diags.HasError() {
		return result
	}

	return result
}

// Identical to lifecycleRuleAndOperatorModel
type lifecycleRuleAndOperatorModelV0 struct {
	ObjectSizeGreaterThan types.Int64  `tfsdk:"object_size_greater_than"`
	ObjectSizeLessThan    types.Int64  `tfsdk:"object_size_less_than"`
	Prefix                types.String `tfsdk:"prefix"`
	Tags                  tftags.Map   `tfsdk:"tags"`
}

// Single
func upgradeLifecycleRuleAndOperatorModelStateFromV0(ctx context.Context, old fwtypes.ListNestedObjectValueOf[lifecycleRuleAndOperatorModelV0], diags *diag.Diagnostics) (result fwtypes.ListNestedObjectValueOf[lifecycleRuleAndOperatorModel]) {
	oldThings, d := old.ToSlice(ctx)
	diags.Append(d...)
	if diags.HasError() {
		return result
	}

	newThings := make([]lifecycleRuleAndOperatorModel, len(oldThings))
	for i, oldThing := range oldThings {
		newThing := lifecycleRuleAndOperatorModel{
			ObjectSizeGreaterThan: oldThing.ObjectSizeGreaterThan,
			ObjectSizeLessThan:    oldThing.ObjectSizeLessThan,
			Prefix:                oldThing.Prefix,
			Tags:                  oldThing.Tags,
		}

		newThings[i] = newThing
	}

	result, d = fwtypes.NewListNestedObjectValueOfValueSlice(ctx, newThings)
	diags.Append(d...)
	if diags.HasError() {
		return result
	}

	return result
}

type noncurrentVersionExpirationModelV0 struct {
	NewerNoncurrentVersions types.String `tfsdk:"newer_noncurrent_versions"`
	NoncurrentDays          types.Int64  `tfsdk:"noncurrent_days"`
}

// Single
func upgradeNoncurrentVersionExpirationModelStateFromV0(ctx context.Context, old fwtypes.ListNestedObjectValueOf[noncurrentVersionExpirationModelV0], diags *diag.Diagnostics) (result fwtypes.ListNestedObjectValueOf[noncurrentVersionExpirationModel]) {
	oldThings, d := old.ToSlice(ctx)
	diags.Append(d...)
	if diags.HasError() {
		return result
	}

	newThings := make([]noncurrentVersionExpirationModel, len(oldThings))
	for i, oldThing := range oldThings {
		newThing := noncurrentVersionExpirationModel{
			NewerNoncurrentVersions: stringToInt32Legacy(ctx, oldThing.NewerNoncurrentVersions, diags),
			NoncurrentDays:          types.Int32Value(fwflex.Int32ValueFromFrameworkInt64(ctx, oldThing.NoncurrentDays)),
		}
		if diags.HasError() {
			return result
		}

		newThings[i] = newThing
	}

	result, d = fwtypes.NewListNestedObjectValueOfValueSlice(ctx, newThings)
	diags.Append(d...)
	if diags.HasError() {
		return result
	}

	return result
}

type noncurrentVersionTransitionModelV0 struct {
	NewerNoncurrentVersions types.String                                        `tfsdk:"newer_noncurrent_versions"`
	NoncurrentDays          types.Int64                                         `tfsdk:"noncurrent_days"`
	StorageClass            fwtypes.StringEnum[awstypes.TransitionStorageClass] `tfsdk:"storage_class"`
}

// Multiple
func upgradeNoncurrentVersionTransitionModelStateFromV0(ctx context.Context, old fwtypes.SetNestedObjectValueOf[noncurrentVersionTransitionModelV0], diags *diag.Diagnostics) (result fwtypes.SetNestedObjectValueOf[noncurrentVersionTransitionModel]) {
	oldThings, d := old.ToSlice(ctx)
	diags.Append(d...)
	if diags.HasError() {
		return result
	}

	newThings := make([]noncurrentVersionTransitionModel, len(oldThings))
	for i, oldThing := range oldThings {
		newThing := noncurrentVersionTransitionModel{
			NewerNoncurrentVersions: stringToInt32Legacy(ctx, oldThing.NewerNoncurrentVersions, diags),
			NoncurrentDays:          types.Int32Value(fwflex.Int32ValueFromFrameworkInt64(ctx, oldThing.NoncurrentDays)),
			StorageClass:            oldThing.StorageClass,
		}

		newThings[i] = newThing
	}

	result, d = fwtypes.NewSetNestedObjectValueOfValueSlice(ctx, newThings)
	diags.Append(d...)
	if diags.HasError() {
		return result
	}

	return result
}

type transitionModelV0 struct {
	Date         types.String                                        `tfsdk:"date"`
	Days         types.Int64                                         `tfsdk:"days"`
	StorageClass fwtypes.StringEnum[awstypes.TransitionStorageClass] `tfsdk:"storage_class"`
}

// Multiple
func upgradeTransitionModelStateFromV0(ctx context.Context, old fwtypes.SetNestedObjectValueOf[transitionModelV0], diags *diag.Diagnostics) (result fwtypes.SetNestedObjectValueOf[transitionModel]) {
	oldThings, d := old.ToSlice(ctx)
	diags.Append(d...)
	if diags.HasError() {
		return result
	}

	newThings := make([]transitionModel, len(oldThings))
	for i, oldThing := range oldThings {
		newThing := transitionModel{
			Date:         migrateDate(ctx, oldThing.Date, diags),
			Days:         int64ToInt32(ctx, oldThing.Days),
			StorageClass: oldThing.StorageClass,
		}
		if diags.HasError() {
			return result
		}

		newThings[i] = newThing
	}

	result, d = fwtypes.NewSetNestedObjectValueOfValueSlice(ctx, newThings)
	diags.Append(d...)
	if diags.HasError() {
		return result
	}

	return result
}

func stringToInt32Legacy(_ context.Context, s types.String, diags *diag.Diagnostics) types.Int32 {
	if s.ValueString() == "" {
		return types.Int32Null()
	}

	v, err := strconv.ParseInt(s.ValueString(), 10, 32)
	if err != nil {
		diags.AddError(
			"Conversion Error",
			fmt.Sprintf("When upgrading state, failed to read a string as an integer value.\n"+
				"Value: %q\nError: %s",
				s.ValueString(),
				err.Error(),
			),
		)
		return types.Int32Unknown()
	}
	return types.Int32Value(int32(v))
}

func stringToInt64Legacy(_ context.Context, s types.String, diags *diag.Diagnostics) types.Int64 {
	if s.ValueString() == "" {
		return types.Int64Null()
	}

	v, err := strconv.ParseInt(s.ValueString(), 10, 64)
	if err != nil {
		diags.AddError(
			"Conversion Error",
			fmt.Sprintf("When upgrading state, failed to read a string as an integer value.\n"+
				"Value: %q\nError: %s",
				s.ValueString(),
				err.Error(),
			),
		)
		return types.Int64Unknown()
	}
	return types.Int64Value(v)
}

func migrateDate(_ context.Context, old types.String, diags *diag.Diagnostics) timetypes.RFC3339 {
	if s := old.ValueString(); s == "" {
		return timetypes.NewRFC3339Null()
	} else {
		v, d := timetypes.NewRFC3339Value(s)
		diags.Append(d...)
		if diags.HasError() {
			return timetypes.NewRFC3339Unknown()
		}
		return v
	}
}

func int64ToInt32(ctx context.Context, i types.Int64) types.Int32 {
	return types.Int32Value(fwflex.Int32ValueFromFrameworkInt64(ctx, i))
}
