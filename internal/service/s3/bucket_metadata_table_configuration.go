// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	awstypes "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/aws-sdk-go-v2/service/s3tables"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/validators"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for resource registration to the Provider. DO NOT EDIT.
// @FrameworkResource("aws_s3_bucket_metadata_table_configuration", name="Bucket Metadata Table Configuration")
func newResourceBucketMetadataTableConfiguration(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceBucketMetadataTableConfiguration{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameBucketMetadataTableConfiguration = "Bucket Metadata Table Configuration"
)

type resourceBucketMetadataTableConfiguration struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
}

func (r *resourceBucketMetadataTableConfiguration) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_s3_bucket_metadata_table_configuration"
}

func (r *resourceBucketMetadataTableConfiguration) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrBucket: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"checksum_algorithm": schema.StringAttribute{
				Optional:   true,
				CustomType: fwtypes.StringEnumType[awstypes.ChecksumAlgorithm](),
			},
			"content_md5": schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					validators.Hash("md5"),
				},
			},
			"expected_bucket_owner": schema.StringAttribute{
				Optional: true,
			},
			names.AttrStatus: schema.StringAttribute{
				Computed: true,
			},
		},
		Blocks: map[string]schema.Block{
			"metadata_table_configuration": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[metadataTableConfiguration](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"s3_tables_destination": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[s3TablesDestination](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"table_bucket_arn": schema.StringAttribute{
										Required: true,
										Validators: []validator.String{
											validators.ARN(),
										},
									},
									"table_name": schema.StringAttribute{
										Required: true,
									},
									"table_namespace": schema.StringAttribute{
										Computed: true,
									},
									"table_arn": framework.ARNAttributeComputedOnly(),
								},
							},
						},
					},
				},
			},
			"timeouts": timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: false,
				Delete: true,
			}),
		},
	}
}

func (r *resourceBucketMetadataTableConfiguration) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().S3Client(ctx)

	var plan resourceBucketMetadataTableConfigurationModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := new(s3.CreateBucketMetadataTableConfigurationInput)

	resp.Diagnostics.Append(flex.Expand(ctx, plan, input)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := conn.CreateBucketMetadataTableConfiguration(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.S3, create.ErrActionCreating, ResNameBucketMetadataTableConfiguration, plan.Bucket.String(), err),
			err.Error(),
		)
		return
	}

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	out, err := waitBucketMetadataTableConfigurationCreated(ctx, conn, plan.Bucket.ValueString(), createTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.S3, create.ErrActionWaitingForCreation, ResNameBucketMetadataTableConfiguration, plan.Bucket.String(), err),
			err.Error(),
		)
		return
	}

	//metadataTableConfigPtr, diags := plan.MetadataTableConfiguration.ToPtr(ctx)
	//resp.Diagnostics.Append(diags...)
	//
	//s3TablesDestPtr, diags := metadataTableConfigPtr.S3TablesDestination.ToPtr(ctx)
	//resp.Diagnostics.Append(diags...)

	plan.Status = flex.StringToFramework(ctx, out.Status)
	resp.Diagnostics.Append(flex.Flatten(ctx, out, &plan, flex.WithFieldNameSuffix("Result"))...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceBucketMetadataTableConfiguration) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().S3Client(ctx)

	var state resourceBucketMetadataTableConfigurationModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findBucketMetadataTableConfigurationByBucket(ctx, conn, state.Bucket.ValueString())
	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.S3, create.ErrActionSetting, ResNameBucketMetadataTableConfiguration, state.Bucket.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &state, flex.WithFieldNameSuffix("Result"))...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceBucketMetadataTableConfiguration) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	return
}

func (r *resourceBucketMetadataTableConfiguration) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().S3Client(ctx)

	s3TablesConn := r.Meta().S3TablesClient(ctx)

	var state resourceBucketMetadataTableConfigurationModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	metadataTableConfig, diags := state.MetadataTableConfiguration.ToPtr(ctx)
	resp.Diagnostics.Append(diags...)

	s3TablesDest, diags := metadataTableConfig.S3TablesDestination.ToPtr(ctx)
	resp.Diagnostics.Append(diags...)

	s3TableDeleteInput := s3tables.DeleteTableInput{
		TableBucketARN: s3TablesDest.TableBucketArn.ValueStringPointer(),
		Name:           s3TablesDest.TableName.ValueStringPointer(),
		Namespace:      s3TablesDest.TableNamespace.ValueStringPointer(),
	}

	_, err := s3TablesConn.DeleteTable(ctx, &s3TableDeleteInput)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.S3Tables, create.ErrActionDeleting, "Table", s3TablesDest.TableName.String(), err),
			err.Error(),
		)
		return
	}

	input := s3.DeleteBucketMetadataTableConfigurationInput{
		Bucket: state.Bucket.ValueStringPointer(),
	}

	if state.ExpectedBucketOwner.ValueString() != "" {
		input.ExpectedBucketOwner = state.ExpectedBucketOwner.ValueStringPointer()
	}

	_, err = conn.DeleteBucketMetadataTableConfiguration(ctx, &input)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.S3, create.ErrActionDeleting, ResNameBucketMetadataTableConfiguration, state.Bucket.String(), err),
			err.Error(),
		)
		return
	}

	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	_, err = waitBucketMetadataTableConfigurationDeleted(ctx, conn, state.Bucket.ValueString(), deleteTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.S3, create.ErrActionWaitingForDeletion, ResNameBucketMetadataTableConfiguration, state.Bucket.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceBucketMetadataTableConfiguration) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("bucket"), req, resp)
}

// TIP: ==== STATUS CONSTANTS ====
// Create constants for states and statuses if the service does not
// already have suitable constants. We prefer that you use the constants
// provided in the service if available (e.g., awstypes.StatusInProgress).
const (
	statusDeleting = "Deleting"
	statusFailed   = "FAILED"
	statusNormal   = "ACTIVE"
	statusCreating = "CREATING"
)

func waitBucketMetadataTableConfigurationCreated(ctx context.Context, conn *s3.Client, bucket string, timeout time.Duration) (*awstypes.GetBucketMetadataTableConfigurationResult, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{},
		Target:                    []string{statusNormal},
		Refresh:                   statusBucketMetadataTableConfiguration(ctx, conn, bucket),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.GetBucketMetadataTableConfigurationResult); ok {
		return out, err
	}

	return nil, err
}

func waitBucketMetadataTableConfigurationDeleted(ctx context.Context, conn *s3.Client, bucket string, timeout time.Duration) (*awstypes.GetBucketMetadataTableConfigurationResult, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{statusDeleting},
		Target:  []string{},
		Refresh: statusBucketMetadataTableConfiguration(ctx, conn, bucket),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.GetBucketMetadataTableConfigurationResult); ok {
		return out, err
	}

	return nil, err
}

func statusBucketMetadataTableConfiguration(ctx context.Context, conn *s3.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := findBucketMetadataTableConfigurationByBucket(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, aws.ToString(out.Status), nil
	}
}

func findBucketMetadataTableConfigurationByBucket(ctx context.Context, conn *s3.Client, bucket string) (*awstypes.GetBucketMetadataTableConfigurationResult, error) {
	in := &s3.GetBucketMetadataTableConfigurationInput{
		Bucket: aws.String(bucket),
	}

	out, err := conn.GetBucketMetadataTableConfiguration(ctx, in)
	if err != nil {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}

	if out == nil || out.GetBucketMetadataTableConfigurationResult == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out.GetBucketMetadataTableConfigurationResult, nil
}

type resourceBucketMetadataTableConfigurationModel struct {
	Bucket                     types.String                                                `tfsdk:"bucket"`
	MetadataTableConfiguration fwtypes.ListNestedObjectValueOf[metadataTableConfiguration] `tfsdk:"metadata_table_configuration"`
	ChecksumAlgorithm          fwtypes.StringEnum[awstypes.ChecksumAlgorithm]              `tfsdk:"checksum_algorithm"`
	Timeouts                   timeouts.Value                                              `tfsdk:"timeouts"`
	ExpectedBucketOwner        types.String                                                `tfsdk:"expected_bucket_owner"`
	ContentMD5                 types.String                                                `tfsdk:"content_md5"`
	Status                     types.String                                                `tfsdk:"status"`
}

type metadataTableConfiguration struct {
	S3TablesDestination fwtypes.ListNestedObjectValueOf[s3TablesDestination] `tfsdk:"s3_tables_destination"`
}

type s3TablesDestination struct {
	TableBucketArn types.String `tfsdk:"table_bucket_arn"`
	TableName      types.String `tfsdk:"table_name"`
	TableArn       types.String `tfsdk:"table_arn"`
	TableNamespace types.String `tfsdk:"table_namespace"`
}
