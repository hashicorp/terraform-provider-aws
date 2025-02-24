// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3

import (
	"context"
	"fmt"
	"reflect"
	"slices"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	awstypes "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	fwvalidators "github.com/hashicorp/terraform-provider-aws/internal/framework/validators"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_s3_bucket_lifecycle_configuration", name="Bucket Lifecycle Configuration")
func newResourceBucketLifecycleConfiguration(context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceBucketLifecycleConfiguration{}
	r.SetDefaultCreateTimeout(3 * time.Minute)
	r.SetDefaultUpdateTimeout(3 * time.Minute)

	return r, nil
}

var (
	_ resource.ResourceWithUpgradeState = &resourceBucketLifecycleConfiguration{}
)

type resourceBucketLifecycleConfiguration struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
}

// Schema returns the schema for this resource.
func (r *resourceBucketLifecycleConfiguration) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Version: 1,
		Attributes: map[string]schema.Attribute{
			names.AttrBucket: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 63),
				},
			},
			names.AttrExpectedBucketOwner: schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					fwvalidators.AWSAccountID(),
				},
			},
			names.AttrID: framework.IDAttributeDeprecatedNoReplacement(),
			"transition_default_minimum_object_size": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.TransitionDefaultMinimumObjectSize](),
				Optional:   true,
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					transitionDefaultMinimumObjectSizeDefault(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			names.AttrRule: schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[lifecycleRuleModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrID: schema.StringAttribute{
							Required: true,
							Validators: []validator.String{
								stringvalidator.LengthBetween(1, 255),
							},
						},
						names.AttrPrefix: schema.StringAttribute{
							Optional: true,
							Computed: true, // Because of Legacy value handling
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
							DeprecationMessage: "Use filter instead",
						},
						names.AttrStatus: schema.StringAttribute{
							Required: true,
							Validators: []validator.String{
								stringvalidator.OneOf(lifecycleRuleStatus_Values()...),
							},
						},
					},
					Blocks: map[string]schema.Block{
						"abort_incomplete_multipart_upload": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[abortIncompleteMultipartUploadModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"days_after_initiation": schema.Int32Attribute{
										Optional: true,
									},
								},
							},
						},
						"expiration": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[lifecycleExpirationModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"date": schema.StringAttribute{
										CustomType: timetypes.RFC3339Type{},
										Optional:   true,
									},
									"days": schema.Int32Attribute{
										Optional: true,
										Computed: true, // Because of Legacy value handling
										PlanModifiers: []planmodifier.Int32{
											int32planmodifier.UseStateForUnknown(),
										},
									},
									"expired_object_delete_marker": schema.BoolAttribute{
										Optional: true,
										Computed: true,
										Validators: []validator.Bool{
											warnIfSetWith(
												path.MatchRelative().AtParent().AtName("date"),
												path.MatchRelative().AtParent().AtName("days"),
											),
										},
									},
								},
							},
						},
						names.AttrFilter: schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[lifecycleRuleFilterModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"object_size_greater_than": schema.Int64Attribute{
										Optional: true,
										Computed: true, // Because of Legacy value handling
										PlanModifiers: []planmodifier.Int64{
											int64planmodifier.UseStateForUnknown(),
										},
									},
									"object_size_less_than": schema.Int64Attribute{
										Optional: true,
										Computed: true, // Because of Legacy value handling
										PlanModifiers: []planmodifier.Int64{
											int64planmodifier.UseStateForUnknown(),
										},
									},
									names.AttrPrefix: schema.StringAttribute{
										Optional: true,
										Computed: true, // Because of Legacy value handling
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.UseStateForUnknown(),
										},
									},
								},
								Blocks: map[string]schema.Block{
									"and": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[lifecycleRuleAndOperatorModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"object_size_greater_than": schema.Int64Attribute{
													Optional: true,
													Computed: true, // Because of Legacy value handling
													// PlanModifiers: []planmodifier.Int64{
													// 	int64planmodifier.UseStateForUnknown(),
													// },
													Validators: []validator.Int64{
														int64validator.AtLeast(0),
													},
												},
												"object_size_less_than": schema.Int64Attribute{
													Optional: true,
													Computed: true, // Because of Legacy value handling
													// PlanModifiers: []planmodifier.Int64{
													// 	int64planmodifier.UseStateForUnknown(),
													// },
													Validators: []validator.Int64{
														int64validator.AtLeast(1),
													},
												},
												names.AttrPrefix: schema.StringAttribute{
													Optional: true,
													Computed: true, // Because of Legacy value handling
													// PlanModifiers: []planmodifier.String{
													// 	stringplanmodifier.UseStateForUnknown(),
													// },
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
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
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
							CustomType: fwtypes.NewListNestedObjectTypeOf[noncurrentVersionExpirationModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"newer_noncurrent_versions": schema.Int32Attribute{
										Optional: true,
										Computed: true, // Because of schema change
										PlanModifiers: []planmodifier.Int32{
											int32planmodifier.UseStateForUnknown(),
										},
										Validators: []validator.Int32{
											int32validator.AtLeast(1),
										},
									},
									"noncurrent_days": schema.Int32Attribute{
										Optional: true,
										Computed: true, // Because of schema change
										PlanModifiers: []planmodifier.Int32{
											int32planmodifier.UseStateForUnknown(),
										},
										Validators: []validator.Int32{
											int32validator.AtLeast(1),
										},
									},
								},
							},
						},
						"noncurrent_version_transition": schema.SetNestedBlock{
							CustomType: fwtypes.NewSetNestedObjectTypeOf[noncurrentVersionTransitionModel](ctx),
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"newer_noncurrent_versions": schema.Int32Attribute{
										Optional: true,
										Computed: true, // Because of schema change
										PlanModifiers: []planmodifier.Int32{
											int32planmodifier.UseStateForUnknown(),
										},
										Validators: []validator.Int32{
											int32validator.AtLeast(1),
										},
									},
									"noncurrent_days": schema.Int32Attribute{
										Optional: true,
										Computed: true, // Because of schema change
										PlanModifiers: []planmodifier.Int32{
											int32planmodifier.UseStateForUnknown(),
										},
										Validators: []validator.Int32{
											int32validator.AtLeast(0),
										},
									},
									names.AttrStorageClass: schema.StringAttribute{
										CustomType: fwtypes.StringEnumType[awstypes.TransitionStorageClass](),
										Required:   true,
									},
								},
							},
						},
						"transition": schema.SetNestedBlock{
							CustomType: fwtypes.NewSetNestedObjectTypeOf[transitionModel](ctx),
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"date": schema.StringAttribute{
										CustomType: timetypes.RFC3339Type{},
										Optional:   true,
									},
									"days": schema.Int32Attribute{
										Optional: true,
										Computed: true,
										Validators: []validator.Int32{
											int32validator.AtLeast(0),
										},
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

func (r *resourceBucketLifecycleConfiguration) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data resourceBucketLifecycleConfigurationModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().S3Client(ctx)
	bucket := data.Bucket.ValueString()
	if isDirectoryBucket(bucket) {
		conn = r.Meta().S3ExpressClient(ctx)
	}

	var input s3.PutBucketLifecycleConfigurationInput
	response.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	var rules []awstypes.LifecycleRule
	response.Diagnostics.Append(fwflex.Expand(ctx, data.Rules, &rules)...)
	if response.Diagnostics.HasError() {
		return
	}

	lifecycleConfiguraton := awstypes.BucketLifecycleConfiguration{
		Rules: rules,
	}

	input.LifecycleConfiguration = &lifecycleConfiguraton

	_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, bucketPropagationTimeout, func() (any, error) {
		return conn.PutBucketLifecycleConfiguration(ctx, &input)
	}, errCodeNoSuchBucket)
	if tfawserr.ErrMessageContains(err, errCodeInvalidArgument, "LifecycleConfiguration is not valid, expected CreateBucketConfiguration") {
		err = errDirectoryBucket(err)
	}
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating S3 Bucket (%s) Lifecycle Configuration", bucket), err.Error())
		return
	}

	expectedBucketOwner := data.ExpectedBucketOwner.ValueString()
	createTimeout := r.CreateTimeout(ctx, data.Timeouts)
	rules, err = waitLifecycleRulesEquals(ctx, conn, bucket, expectedBucketOwner, input.LifecycleConfiguration.Rules, createTimeout)
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating S3 Bucket (%s) Lifecycle Configuration", bucket), fmt.Sprintf("While waiting: %s", err.Error()))
		return
	}

	output, err := findBucketLifecycleConfiguration(ctx, conn, bucket, expectedBucketOwner)
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating S3 Bucket (%s) Lifecycle Configuration", bucket), err.Error())
		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)

	data.ID = types.StringValue(createResourceID(bucket, expectedBucketOwner))
	data.ExpectedBucketOwner = types.StringValue(expectedBucketOwner)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *resourceBucketLifecycleConfiguration) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data resourceBucketLifecycleConfigurationModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().S3Client(ctx)
	bucket := data.Bucket.ValueString()
	if isDirectoryBucket(bucket) {
		conn = r.Meta().S3ExpressClient(ctx)
	}

	expectedBucketOwner := data.ExpectedBucketOwner.ValueString()

	const (
		lifecycleConfigurationRulesSteadyTimeout = 2 * time.Minute
	)
	var lastOutput, output *s3.GetBucketLifecycleConfigurationOutput
	err := retry.RetryContext(ctx, lifecycleConfigurationRulesSteadyTimeout, func() *retry.RetryError {
		var err error

		output, err = findBucketLifecycleConfiguration(ctx, conn, bucket, expectedBucketOwner)
		if err != nil {
			return retry.NonRetryableError(err)
		}

		if lastOutput == nil || !lifecycleRulesEqual(lastOutput.Rules, output.Rules) {
			lastOutput = output
			return retry.RetryableError(fmt.Errorf("S3 Bucket Lifecycle Configuration (%s) has not stablized; retrying", bucket))
		}

		return nil
	})
	if tfresource.TimedOut(err) {
		output, err = findBucketLifecycleConfiguration(ctx, conn, bucket, expectedBucketOwner)
	}
	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading S3 Bucket Lifecycle Configuration (%s)", data.Bucket.ValueString()), err.Error())
		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *resourceBucketLifecycleConfiguration) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var old, new resourceBucketLifecycleConfigurationModel

	response.Diagnostics.Append(request.State.Get(ctx, &old)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().S3Client(ctx)
	bucket := new.Bucket.ValueString()
	if isDirectoryBucket(bucket) {
		conn = r.Meta().S3ExpressClient(ctx)
	}

	var input s3.PutBucketLifecycleConfigurationInput
	response.Diagnostics.Append(fwflex.Expand(ctx, new, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	var rules []awstypes.LifecycleRule
	response.Diagnostics.Append(fwflex.Expand(ctx, new.Rules, &rules)...)
	if response.Diagnostics.HasError() {
		return
	}

	lifecycleConfiguraton := awstypes.BucketLifecycleConfiguration{
		Rules: rules,
	}

	input.LifecycleConfiguration = &lifecycleConfiguraton

	_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, bucketPropagationTimeout, func() (any, error) {
		return conn.PutBucketLifecycleConfiguration(ctx, &input)
	}, errCodeNoSuchBucket)
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("updating S3 Bucket (%s) Lifecycle Configuration", bucket), err.Error())
		return
	}

	expectedBucketOwner := new.ExpectedBucketOwner.ValueString()
	updateTimeout := r.UpdateTimeout(ctx, new.Timeouts)
	rules, err = waitLifecycleRulesEquals(ctx, conn, bucket, expectedBucketOwner, input.LifecycleConfiguration.Rules, updateTimeout)
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("updating S3 Bucket (%s) Lifecycle Configuration", bucket), fmt.Sprintf("While waiting: %s", err.Error()))
		return
	}

	output, err := findBucketLifecycleConfiguration(ctx, conn, bucket, expectedBucketOwner)
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("updating S3 Bucket (%s) Lifecycle Configuration", bucket), err.Error())
		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &new)...)

	new.ID = types.StringValue(createResourceID(bucket, expectedBucketOwner))
	new.ExpectedBucketOwner = types.StringValue(expectedBucketOwner)

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *resourceBucketLifecycleConfiguration) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data resourceBucketLifecycleConfigurationModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().S3Client(ctx)
	bucket := data.Bucket.ValueString()
	if isDirectoryBucket(bucket) {
		conn = r.Meta().S3ExpressClient(ctx)
	}

	input := s3.DeleteBucketLifecycleInput{
		Bucket: aws.String(bucket),
	}
	expectedBucketOwner := data.ExpectedBucketOwner.ValueString()
	if expectedBucketOwner != "" {
		input.ExpectedBucketOwner = aws.String(expectedBucketOwner)
	}

	_, err := conn.DeleteBucketLifecycle(ctx, &input)
	if tfawserr.ErrCodeEquals(err, errCodeNoSuchBucket, errCodeNoSuchLifecycleConfiguration) {
		return
	}
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting S3 Bucket Lifecycle Configuration (%s)", data.Bucket.ValueString()), err.Error())
		return
	}

	_, err = tfresource.RetryUntilNotFound(ctx, bucketPropagationTimeout, func() (any, error) {
		return findBucketLifecycleConfiguration(ctx, conn, bucket, expectedBucketOwner)
	})
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting S3 Bucket Lifecycle Configuration (%s)", data.Bucket.ValueString()), fmt.Sprintf("While waiting: %s", err.Error()))
		return
	}
}

func (r *resourceBucketLifecycleConfiguration) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	bucket, expectedBucketOwner, err := parseResourceID(request.ID)
	if err != nil {
		response.Diagnostics.AddError("Resource Import Invalid ID", err.Error())
		return
	}

	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root(names.AttrBucket), bucket)...)
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root(names.AttrExpectedBucketOwner), expectedBucketOwner)...)

	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root(names.AttrID), request.ID)...)
}

func (r *resourceBucketLifecycleConfiguration) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	schemaV0 := bucketLifeCycleConfigurationSchemaV0(ctx)

	return map[int64]resource.StateUpgrader{
		0: {
			PriorSchema:   &schemaV0,
			StateUpgrader: upgradeBucketLifeCycleConfigurationResourceStateFromV0,
		},
	}
}

func findBucketLifecycleConfiguration(ctx context.Context, conn *s3.Client, bucket, expectedBucketOwner string) (*s3.GetBucketLifecycleConfigurationOutput, error) {
	input := s3.GetBucketLifecycleConfigurationInput{
		Bucket: aws.String(bucket),
	}
	if expectedBucketOwner != "" {
		input.ExpectedBucketOwner = aws.String(expectedBucketOwner)
	}

	output, err := conn.GetBucketLifecycleConfiguration(ctx, &input)

	if tfawserr.ErrCodeEquals(err, errCodeNoSuchBucket, errCodeNoSuchLifecycleConfiguration) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.Rules) == 0 {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func lifecycleRulesEqual(rules1, rules2 []awstypes.LifecycleRule) bool {
	if len(rules1) != len(rules2) {
		return false
	}

	for _, rule1 := range rules1 {
		// We consider 2 LifecycleRules equal if their IDs and Statuses are equal.
		if !slices.ContainsFunc(rules2, func(rule2 awstypes.LifecycleRule) bool {
			return aws.ToString(rule1.ID) == aws.ToString(rule2.ID) && rule1.Status == rule2.Status
		}) {
			return false
		}
	}

	return true
}

func statusLifecycleRulesEquals(ctx context.Context, conn *s3.Client, bucket, expectedBucketOwner string, rules []awstypes.LifecycleRule) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findBucketLifecycleConfiguration(ctx, conn, bucket, expectedBucketOwner)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, strconv.FormatBool(lifecycleRulesEqual(output.Rules, rules)), nil
	}
}

func waitLifecycleRulesEquals(ctx context.Context, conn *s3.Client, bucket, expectedBucketOwner string, rules []awstypes.LifecycleRule, timeout time.Duration) ([]awstypes.LifecycleRule, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Target:                    []string{strconv.FormatBool(true)},
		Refresh:                   statusLifecycleRulesEquals(ctx, conn, bucket, expectedBucketOwner, rules),
		Timeout:                   timeout,
		MinTimeout:                10 * time.Second,
		ContinuousTargetOccurence: 3,
		NotFoundChecks:            20,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.([]awstypes.LifecycleRule); ok {
		return output, err
	}

	return nil, err
}

const (
	lifecycleRuleStatusDisabled = "Disabled"
	lifecycleRuleStatusEnabled  = "Enabled"
)

func lifecycleRuleStatus_Values() []string {
	return []string{
		lifecycleRuleStatusDisabled,
		lifecycleRuleStatusEnabled,
	}
}

type resourceBucketLifecycleConfigurationModel struct {
	Bucket                             types.String                                                    `tfsdk:"bucket"`
	ExpectedBucketOwner                types.String                                                    `tfsdk:"expected_bucket_owner" autoflex:",legacy"`
	ID                                 types.String                                                    `tfsdk:"id"`
	Rules                              fwtypes.ListNestedObjectValueOf[lifecycleRuleModel]             `tfsdk:"rule"`
	TransitionDefaultMinimumObjectSize fwtypes.StringEnum[awstypes.TransitionDefaultMinimumObjectSize] `tfsdk:"transition_default_minimum_object_size" autoflex:",legacy"`
	Timeouts                           timeouts.Value                                                  `tfsdk:"timeouts"`
}

var (
	_ fwflex.Expander  = lifecycleRuleModel{}
	_ fwflex.Flattener = &lifecycleRuleModel{}
)

type lifecycleRuleModel struct {
	AbortIncompleteMultipartUpload fwtypes.ListNestedObjectValueOf[abortIncompleteMultipartUploadModel] `tfsdk:"abort_incomplete_multipart_upload"`
	Expiration                     fwtypes.ListNestedObjectValueOf[lifecycleExpirationModel]            `tfsdk:"expiration"`
	Filter                         fwtypes.ListNestedObjectValueOf[lifecycleRuleFilterModel]            `tfsdk:"filter"`
	ID                             types.String                                                         `tfsdk:"id"`
	NoncurrentVersionExpirations   fwtypes.ListNestedObjectValueOf[noncurrentVersionExpirationModel]    `tfsdk:"noncurrent_version_expiration"`
	NoncurrentVersionTransitions   fwtypes.SetNestedObjectValueOf[noncurrentVersionTransitionModel]     `tfsdk:"noncurrent_version_transition"`
	Prefix                         types.String                                                         `tfsdk:"prefix"`
	Status                         fwtypes.StringEnum[awstypes.ExpirationStatus]                        `tfsdk:"status"`
	Transitions                    fwtypes.SetNestedObjectValueOf[transitionModel]                      `tfsdk:"transition"`
}

func (m lifecycleRuleModel) Expand(ctx context.Context) (result any, diags diag.Diagnostics) {
	var r awstypes.LifecycleRule

	d := fwflex.Expand(ctx, m.AbortIncompleteMultipartUpload, &r.AbortIncompleteMultipartUpload)
	diags.Append(d...)
	if d.HasError() {
		return nil, diags
	}

	d = fwflex.Expand(ctx, m.Expiration, &r.Expiration)
	diags.Append(d...)
	if d.HasError() {
		return nil, diags
	}

	// For legacy-mode reasons, `prefix` may be empty, but should be treated as `nil`
	prefix := fwflex.EmptyStringAsNull(m.Prefix)

	// The AWS API requires a value for `filter` unless `prefix` is set. If `filter` is set, one and only one of
	// `and`, `object_size_greater_than`, `object_size_less_than`, `prefix`, or `tags` must be set.
	// However, the provider historically has allowed `filter` to be null, empty, or have one child value set.
	// (Setting multiple elements would result in a run-time error)
	// For null `filter`, send an empty LifecycleRuleFilter
	if m.Filter.IsUnknown() || m.Filter.IsNull() {
		if prefix.IsUnknown() || prefix.IsNull() {
			filter := awstypes.LifecycleRuleFilter{}
			r.Filter = &filter
		}
	} else {
		filter, d := m.Filter.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		if filter == nil {
			diags.AddError(
				"Unexpected Error",
				"An unexpected error occurred while preparing request. "+
					"This is always an error in the provider. "+
					"Please report the following to the provider developer:\n\n"+
					`Expanding "lifecycleRuleModel": "Filter" has value, but returned nil`,
			)
			return nil, diags
		}
		if isFilterModelZero(filter) {
			filter := awstypes.LifecycleRuleFilter{
				Prefix: aws.String(""),
			}
			r.Filter = &filter
		} else {
			d = fwflex.Expand(ctx, m.Filter, &r.Filter)
			diags.Append(d...)
			if d.HasError() {
				return nil, diags
			}
		}
	}

	r.ID = fwflex.StringFromFramework(ctx, m.ID)

	d = fwflex.Expand(ctx, m.NoncurrentVersionExpirations, &r.NoncurrentVersionExpiration)
	diags.Append(d...)
	if d.HasError() {
		return nil, diags
	}

	d = fwflex.Expand(ctx, m.NoncurrentVersionTransitions, &r.NoncurrentVersionTransitions)
	diags.Append(d...)
	if d.HasError() {
		return nil, diags
	}

	r.Prefix = fwflex.StringFromFramework(ctx, prefix)

	r.Status = m.Status.ValueEnum()

	d = fwflex.Expand(ctx, m.Transitions, &r.Transitions)
	diags.Append(d...)
	if d.HasError() {
		return nil, diags
	}

	return &r, diags
}

func isFilterModelZero(v *lifecycleRuleFilterModel) bool {
	if !v.And.IsNull() {
		return false
	}

	if !v.ObjectSizeGreaterThan.IsUnknown() {
		return false
	}

	if !v.ObjectSizeLessThan.IsUnknown() {
		return false
	}

	if !v.Prefix.IsUnknown() {
		return false
	}

	if !v.Tag.IsNull() {
		return false
	}

	return true
}

func (m *lifecycleRuleModel) Flatten(ctx context.Context, v any) (diags diag.Diagnostics) {
	rule, ok := v.(awstypes.LifecycleRule)
	if !ok {
		diags.Append(fwflex.DiagFlatteningIncompatibleTypes(reflect.TypeOf(v), reflect.TypeFor[lifecycleRuleModel]()))
		return diags
	}

	d := fwflex.Flatten(ctx, rule.AbortIncompleteMultipartUpload, &m.AbortIncompleteMultipartUpload)
	diags.Append(d...)
	if d.HasError() {
		return diags
	}

	d = fwflex.Flatten(ctx, rule.Expiration, &m.Expiration)
	diags.Append(d...)
	if d.HasError() {
		return diags
	}

	// If Filter has no values set, the value in the configuration was null
	if isLifecycleRuleFilterZero(rule.Filter) {
		m.Filter = fwtypes.NewListNestedObjectValueOfNull[lifecycleRuleFilterModel](ctx)
	} else {
		d = fwflex.Flatten(ctx, rule.Filter, &m.Filter)
		diags.Append(d...)
		if d.HasError() {
			return diags
		}
	}

	m.ID = fwflex.StringToFramework(ctx, rule.ID)

	d = fwflex.Flatten(ctx, rule.NoncurrentVersionExpiration, &m.NoncurrentVersionExpirations)
	diags.Append(d...)
	if d.HasError() {
		return diags
	}

	d = fwflex.Flatten(ctx, rule.NoncurrentVersionTransitions, &m.NoncurrentVersionTransitions)
	diags.Append(d...)
	if d.HasError() {
		return diags
	}

	m.Prefix = fwflex.StringToFrameworkLegacy(ctx, rule.Prefix)

	m.Status = fwtypes.StringEnumValue(rule.Status)

	d = fwflex.Flatten(ctx, rule.Transitions, &m.Transitions)
	diags.Append(d...)
	if d.HasError() {
		return diags
	}

	return diags
}

func isLifecycleRuleFilterZero(v *awstypes.LifecycleRuleFilter) bool {
	return v == nil ||
		(v.And == nil &&
			v.ObjectSizeGreaterThan == nil &&
			v.ObjectSizeLessThan == nil &&
			v.Prefix == nil &&
			v.Tag == nil)
}

type abortIncompleteMultipartUploadModel struct {
	DaysAfterInitiation types.Int32 `tfsdk:"days_after_initiation"`
}

var (
	_ fwflex.Expander = lifecycleExpirationModel{}
)

type lifecycleExpirationModel struct {
	Date                      timetypes.RFC3339 `tfsdk:"date" autoflex:",legacy"`
	Days                      types.Int32       `tfsdk:"days" autoflex:",legacy"`
	ExpiredObjectDeleteMarker types.Bool        `tfsdk:"expired_object_delete_marker" autoflex:",legacy"`
}

func (m lifecycleExpirationModel) Expand(ctx context.Context) (result any, diags diag.Diagnostics) {
	var r awstypes.LifecycleExpiration

	r.Date = fwflex.TimeFromFramework(ctx, m.Date)

	// For legacy-mode reasons, `days` may be zero, but should be treated as `nil`
	days := fwflex.ZeroInt32AsNull(m.Days)

	r.Days = fwflex.Int32FromFrameworkInt32(ctx, days)

	if m.ExpiredObjectDeleteMarker.IsUnknown() || m.ExpiredObjectDeleteMarker.IsNull() {
		if (m.Date.IsUnknown() || m.Date.IsNull()) && (days.IsUnknown() || days.IsNull()) {
			r.ExpiredObjectDeleteMarker = aws.Bool(false)
		}
	} else if (m.Date.IsUnknown() || m.Date.IsNull()) && (days.IsUnknown() || days.IsNull()) {
		r.ExpiredObjectDeleteMarker = fwflex.BoolFromFramework(ctx, m.ExpiredObjectDeleteMarker)
	} else {
		r.ExpiredObjectDeleteMarker = nil
	}

	return &r, diags
}

type lifecycleRuleFilterModel struct {
	And                   fwtypes.ListNestedObjectValueOf[lifecycleRuleAndOperatorModel] `tfsdk:"and"`
	ObjectSizeGreaterThan types.Int64                                                    `tfsdk:"object_size_greater_than"`
	ObjectSizeLessThan    types.Int64                                                    `tfsdk:"object_size_less_than"`
	Prefix                types.String                                                   `tfsdk:"prefix"`
	Tag                   fwtypes.ListNestedObjectValueOf[tagModel]                      `tfsdk:"tag"`
}

type noncurrentVersionExpirationModel struct {
	NewerNoncurrentVersions types.Int32 `tfsdk:"newer_noncurrent_versions"`
	NoncurrentDays          types.Int32 `tfsdk:"noncurrent_days"`
}

type noncurrentVersionTransitionModel struct {
	NewerNoncurrentVersions types.Int32                                         `tfsdk:"newer_noncurrent_versions"`
	NoncurrentDays          types.Int32                                         `tfsdk:"noncurrent_days"`
	StorageClass            fwtypes.StringEnum[awstypes.TransitionStorageClass] `tfsdk:"storage_class"`
}

var (
	_ fwflex.Expander = transitionModel{}
)

type transitionModel struct {
	Date         timetypes.RFC3339                                   `tfsdk:"date"`
	Days         types.Int32                                         `tfsdk:"days"`
	StorageClass fwtypes.StringEnum[awstypes.TransitionStorageClass] `tfsdk:"storage_class"`
}

func (m transitionModel) Expand(ctx context.Context) (result any, diags diag.Diagnostics) {
	var r awstypes.Transition

	r.Date = fwflex.TimeFromFramework(ctx, m.Date)

	if m.Days.IsUnknown() || m.Days.IsNull() {
		if m.Date.IsUnknown() || m.Date.IsNull() {
			r.Days = aws.Int32(0)
		}
	} else {
		r.Days = fwflex.Int32FromFrameworkInt32(ctx, m.Days)
	}

	r.StorageClass = m.StorageClass.ValueEnum()

	return &r, diags
}

type tagModel struct {
	Key   types.String `tfsdk:"key"`
	Value types.String `tfsdk:"value"`
}

var (
	_ fwflex.Expander  = lifecycleRuleAndOperatorModel{}
	_ fwflex.Flattener = &lifecycleRuleAndOperatorModel{}
)

type lifecycleRuleAndOperatorModel struct {
	ObjectSizeGreaterThan types.Int64  `tfsdk:"object_size_greater_than" autoflex:",legacy"`
	ObjectSizeLessThan    types.Int64  `tfsdk:"object_size_less_than" autoflex:",legacy"`
	Prefix                types.String `tfsdk:"prefix" autoflex:",legacy"`
	Tags                  tftags.Map   `tfsdk:"tags"`
}

func (m lifecycleRuleAndOperatorModel) Expand(ctx context.Context) (result any, diags diag.Diagnostics) {
	var r awstypes.LifecycleRuleAndOperator

	r.ObjectSizeGreaterThan = fwflex.Int64FromFramework(ctx, m.ObjectSizeGreaterThan)

	r.ObjectSizeLessThan = fwflex.Int64FromFramework(ctx, m.ObjectSizeLessThan)

	r.Prefix = fwflex.StringFromFramework(ctx, m.Prefix)

	if tags := Tags(tftags.New(ctx, m.Tags).IgnoreAWS()); len(tags) > 0 {
		r.Tags = tags
	}

	return &r, diags
}

func (m *lifecycleRuleAndOperatorModel) Flatten(ctx context.Context, v any) (diags diag.Diagnostics) {
	and, ok := v.(awstypes.LifecycleRuleAndOperator)
	if !ok {
		diags.Append(fwflex.DiagFlatteningIncompatibleTypes(reflect.TypeOf(v), reflect.TypeFor[lifecycleRuleAndOperatorModel]()))
		return diags
	}

	m.ObjectSizeGreaterThan = fwflex.Int64ToFrameworkLegacy(ctx, and.ObjectSizeGreaterThan)

	m.ObjectSizeLessThan = fwflex.Int64ToFrameworkLegacy(ctx, and.ObjectSizeLessThan)

	m.Prefix = fwflex.StringToFrameworkLegacy(ctx, and.Prefix)

	m.Tags = tftags.NewMapFromMapValue(fwflex.FlattenFrameworkStringValueMap(ctx, keyValueTags(ctx, and.Tags).Map()))

	return diags
}

func transitionDefaultMinimumObjectSizeDefault() planmodifier.String {
	return transitionDefaultMinimumObjectSizeDefaultModifier{}
}

type transitionDefaultMinimumObjectSizeDefaultModifier struct{}

func (m transitionDefaultMinimumObjectSizeDefaultModifier) Description(_ context.Context) string {
	return "Defaults to '" + string(awstypes.TransitionDefaultMinimumObjectSizeAllStorageClasses128k) + "' for general purpose buckets, nothing otherwise."
}

func (m transitionDefaultMinimumObjectSizeDefaultModifier) MarkdownDescription(_ context.Context) string {
	return "Defaults to `" + string(awstypes.TransitionDefaultMinimumObjectSizeAllStorageClasses128k) + "` for general purpose buckets, nothing otherwise."
}

func (m transitionDefaultMinimumObjectSizeDefaultModifier) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	var bucket types.String
	resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, path.Root(names.AttrBucket), &bucket)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if isDirectoryBucket(bucket.ValueString()) {
		return
	}

	// There's already a value configured, so do nothing
	if !req.ConfigValue.IsNull() {
		return
	}

	v, d := fwtypes.StringEnumValue(awstypes.TransitionDefaultMinimumObjectSizeAllStorageClasses128k).ToStringValue(ctx)
	resp.Diagnostics.Append(d...)
	if d.HasError() {
		return
	}
	resp.PlanValue = v
}

var (
	_ validator.Bool = warnIfSetWithValidator{}
)

func warnIfSetWith(expressions ...path.Expression) validator.Bool {
	return warnIfSetWithValidator{
		PathExpressions: expressions,
	}
}

type warnIfSetWithValidator struct {
	PathExpressions path.Expressions
}

func (v warnIfSetWithValidator) Description(ctx context.Context) string {
	return v.MarkdownDescription(ctx)
}

func (v warnIfSetWithValidator) MarkdownDescription(_ context.Context) string {
	return fmt.Sprintf("Ensure that if an attribute is set, a warning is emitted if any of these are also set: %q", v.PathExpressions)
}

// Validation logic is adapted from the standard ConflictsWith validator
// available for all types
func (v warnIfSetWithValidator) ValidateBool(ctx context.Context, req validator.BoolRequest, resp *validator.BoolResponse) {
	// If attribute configuration is null, it cannot conflict with others
	// If attribute configuration is unknown, delay the validation until it is known.
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	expressions := req.PathExpression.MergeExpressions(v.PathExpressions...)

	for _, expression := range expressions {
		matchedPaths, diags := req.Config.PathMatches(ctx, expression)

		resp.Diagnostics.Append(diags...)

		// Collect all errors
		if diags.HasError() {
			continue
		}

		for _, mp := range matchedPaths {
			// If the user specifies the same attribute this validator is applied to,
			// also as part of the input, skip it
			if mp.Equal(req.Path) {
				continue
			}

			var mpVal attr.Value
			diags := req.Config.GetAttribute(ctx, mp, &mpVal)
			resp.Diagnostics.Append(diags...)

			// Collect all errors
			if diags.HasError() {
				continue
			}

			// Delay validation until all involved attribute have a known value
			if mpVal.IsUnknown() {
				return
			}

			if !mpVal.IsNull() {
				resp.Diagnostics.Append(diag.NewAttributeWarningDiagnostic(
					req.Path,
					"Invalid Attribute Combination",
					fmt.Sprintf("Attribute %q should not be specified when %q is also specified", req.Path, mp),
				))
			}
		}
	}
}
