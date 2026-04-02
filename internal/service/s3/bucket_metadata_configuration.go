// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package s3

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	awstypes "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	fwvalidators "github.com/hashicorp/terraform-provider-aws/internal/framework/validators"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_s3_bucket_metadata_configuration", name="Bucket Metadata Configuration")
// @IdentityAttribute("bucket")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/s3/types;awstypes;awstypes.MetadataConfigurationResult")
// @Testing(importStateIdAttribute="bucket")
// @Testing(preIdentityVersion="6.32.0")
func newBucketMetadataConfigurationResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &bucketMetadataConfigurationResource{}

	r.SetDefaultCreateTimeout(30 * time.Minute)

	return r, nil
}

type bucketMetadataConfigurationResource struct {
	framework.ResourceWithModel[bucketMetadataConfigurationResourceModel]
	framework.WithImportByIdentity
	framework.WithTimeouts
}

func (r *bucketMetadataConfigurationResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrBucket: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrExpectedBucketOwner: schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					fwvalidators.AWSAccountID(),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				DeprecationMessage: "This attribute will be removed in a future verion of the provider.",
			},
		},
		Blocks: map[string]schema.Block{
			"metadata_configuration": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[metadataConfigurationModel](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtLeast(1),
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrDestination: framework.ResourceComputedListOfObjectsAttribute[destinationResultModel](ctx, listplanmodifier.UseStateForUnknown()),
					},
					Blocks: map[string]schema.Block{
						"inventory_table_configuration": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[inventoryTableConfigurationModel](ctx),
							Validators: []validator.List{
								listvalidator.IsRequired(),
								listvalidator.SizeAtLeast(1),
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"configuration_state": schema.StringAttribute{
										CustomType: fwtypes.StringEnumType[awstypes.InventoryConfigurationState](),
										Required:   true,
									},
									"table_arn": schema.StringAttribute{
										Computed: true,
									},
									names.AttrTableName: schema.StringAttribute{
										Computed: true,
									},
								},
								Blocks: map[string]schema.Block{
									names.AttrEncryptionConfiguration: schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[metadataTableEncryptionConfigurationModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												names.AttrKMSKeyARN: schema.StringAttribute{
													CustomType: fwtypes.ARNType,
													Optional:   true,
												},
												"sse_algorithm": schema.StringAttribute{
													CustomType: fwtypes.StringEnumType[awstypes.TableSseAlgorithm](),
													Required:   true,
												},
											},
										},
									},
								},
							},
						},
						"journal_table_configuration": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[journalTableConfigurationModel](ctx),
							Validators: []validator.List{
								listvalidator.IsRequired(),
								listvalidator.SizeAtLeast(1),
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"table_arn": schema.StringAttribute{
										Computed: true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.UseStateForUnknown(),
										},
									},
									names.AttrTableName: schema.StringAttribute{
										Computed: true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.UseStateForUnknown(),
										},
									},
								},
								Blocks: map[string]schema.Block{
									names.AttrEncryptionConfiguration: schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[metadataTableEncryptionConfigurationModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												names.AttrKMSKeyARN: schema.StringAttribute{
													CustomType: fwtypes.ARNType,
													Optional:   true,
												},
												"sse_algorithm": schema.StringAttribute{
													CustomType: fwtypes.StringEnumType[awstypes.TableSseAlgorithm](),
													Required:   true,
												},
											},
										},
									},
									"record_expiration": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[recordExpirationModel](ctx),
										Validators: []validator.List{
											listvalidator.IsRequired(),
											listvalidator.SizeAtLeast(1),
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"days": schema.Int32Attribute{
													Optional: true,
													Validators: []validator.Int32{
														int32validator.Between(7, 2147483647),
													},
												},
												"expiration": schema.StringAttribute{
													CustomType: fwtypes.StringEnumType[awstypes.ExpirationState](),
													Required:   true,
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
			}),
		},
	}
}

func (r *bucketMetadataConfigurationResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data bucketMetadataConfigurationResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().S3Client(ctx)

	bucket := fwflex.StringValueFromFramework(ctx, data.Bucket)
	if isDirectoryBucket(bucket) {
		conn = r.Meta().S3ExpressClient(ctx)
	}
	expectedBucketOwner := fwflex.StringValueFromFramework(ctx, data.ExpectedBucketOwner)
	var input s3.CreateBucketMetadataConfigurationInput
	response.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	_, err := conn.CreateBucketMetadataConfiguration(ctx, &input)

	if tfawserr.ErrHTTPStatusCodeEquals(err, http.StatusMethodNotAllowed) {
		err = errDirectoryBucket(err)
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating S3 Bucket Metadata Configuration (%s)", bucket), err.Error())

		return
	}

	if _, err := waitBucketMetadataJournalTableConfigurationCreated(ctx, conn, bucket, expectedBucketOwner, r.CreateTimeout(ctx, data.Timeouts)); err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for S3 Bucket Metadata journal table configuration (%s) create", bucket), err.Error())

		return
	}

	if input.MetadataConfiguration.InventoryTableConfiguration.ConfigurationState == awstypes.InventoryConfigurationStateEnabled {
		if _, err := waitBucketMetadataInventoryTableConfigurationCreated(ctx, conn, bucket, expectedBucketOwner, r.CreateTimeout(ctx, data.Timeouts)); err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("waiting for S3 Bucket Metadata inventory table configuration (%s) create", bucket), err.Error())

			return
		}
	}

	// Set values for unknowns.
	output, err := findBucketMetadataConfigurationByTwoPartKey(ctx, conn, bucket, expectedBucketOwner)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading S3 Bucket Metadata Configuration (%s)", bucket), err.Error())

		return
	}

	// Encryption configurations are not returned via the API.
	// Propagate from Plan.
	inventoryEncryptionConfiguration, journalEncryptionConfiguration, diags := getMetadataTableEncryptionConfigurationModels(ctx, &data)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data.MetadataConfiguration, fwflex.WithFieldNameSuffix("Result"))...)
	if response.Diagnostics.HasError() {
		return
	}

	diags = setMetadataTableEncryptionConfigurationModels(ctx, &data, inventoryEncryptionConfiguration, journalEncryptionConfiguration)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *bucketMetadataConfigurationResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data bucketMetadataConfigurationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().S3Client(ctx)

	bucket, expectedBucketOwner := fwflex.StringValueFromFramework(ctx, data.Bucket), fwflex.StringValueFromFramework(ctx, data.ExpectedBucketOwner)
	output, err := findBucketMetadataConfigurationByTwoPartKey(ctx, conn, bucket, expectedBucketOwner)

	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading S3 Bucket Metadata Configuration (%s)", bucket), err.Error())

		return
	}

	// Encryption configurations are not returned via the API.
	// Propagate from State.
	inventoryEncryptionConfiguration, journalEncryptionConfiguration, diags := getMetadataTableEncryptionConfigurationModels(ctx, &data)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	// Set attributes for import.
	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data.MetadataConfiguration, fwflex.WithFieldNameSuffix("Result"))...)
	if response.Diagnostics.HasError() {
		return
	}

	diags = setMetadataTableEncryptionConfigurationModels(ctx, &data, inventoryEncryptionConfiguration, journalEncryptionConfiguration)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *bucketMetadataConfigurationResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var old, new bucketMetadataConfigurationResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.State.Get(ctx, &old)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().S3Client(ctx)

	bucket, expectedBucketOwner := fwflex.StringValueFromFramework(ctx, new.Bucket), fwflex.StringValueFromFramework(ctx, new.ExpectedBucketOwner)

	newMetadataConfigurationModel, diags := new.MetadataConfiguration.ToPtr(ctx)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}
	oldMetadataConfigurationModel, diags := old.MetadataConfiguration.ToPtr(ctx)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	if !newMetadataConfigurationModel.InventoryTableConfiguration.Equal(oldMetadataConfigurationModel.InventoryTableConfiguration) {
		var input s3.UpdateBucketMetadataInventoryTableConfigurationInput
		response.Diagnostics.Append(fwflex.Expand(ctx, new.MetadataConfiguration, &input)...)
		if response.Diagnostics.HasError() {
			return
		}
		input.Bucket = aws.String(bucket)
		if expectedBucketOwner != "" {
			input.ExpectedBucketOwner = aws.String(expectedBucketOwner)
		}

		_, err := conn.UpdateBucketMetadataInventoryTableConfiguration(ctx, &input)

		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("updating S3 Bucket Metadata inventory table configuration (%s)", bucket), err.Error())

			return
		}
	}

	if !newMetadataConfigurationModel.JournalTableConfiguration.Equal(oldMetadataConfigurationModel.JournalTableConfiguration) {
		var input s3.UpdateBucketMetadataJournalTableConfigurationInput
		response.Diagnostics.Append(fwflex.Expand(ctx, new.MetadataConfiguration, &input)...)
		if response.Diagnostics.HasError() {
			return
		}
		input.Bucket = aws.String(bucket)
		if expectedBucketOwner != "" {
			input.ExpectedBucketOwner = aws.String(expectedBucketOwner)
		}

		_, err := conn.UpdateBucketMetadataJournalTableConfiguration(ctx, &input)

		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("updating S3 Bucket Metadata journal table configuration (%s)", bucket), err.Error())

			return
		}
	}

	// Set values for unknowns.
	output, err := findBucketMetadataConfigurationByTwoPartKey(ctx, conn, bucket, expectedBucketOwner)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading S3 Bucket Metadata Configuration (%s)", bucket), err.Error())

		return
	}

	// Encryption configurations are not returned via the API.
	// Propagate from Plan.
	inventoryEncryptionConfiguration, journalEncryptionConfiguration, diags := getMetadataTableEncryptionConfigurationModels(ctx, &new)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &new.MetadataConfiguration, fwflex.WithFieldNameSuffix("Result"))...)
	if response.Diagnostics.HasError() {
		return
	}

	diags = setMetadataTableEncryptionConfigurationModels(ctx, &new, inventoryEncryptionConfiguration, journalEncryptionConfiguration)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, new)...)
}

func (r *bucketMetadataConfigurationResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data bucketMetadataConfigurationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().S3Client(ctx)

	bucket, expectedBucketOwner := fwflex.StringValueFromFramework(ctx, data.Bucket), fwflex.StringValueFromFramework(ctx, data.ExpectedBucketOwner)
	input := s3.DeleteBucketMetadataConfigurationInput{
		Bucket: aws.String(bucket),
	}
	if expectedBucketOwner != "" {
		input.ExpectedBucketOwner = aws.String(expectedBucketOwner)
	}
	_, err := conn.DeleteBucketMetadataConfiguration(ctx, &input)

	if tfawserr.ErrCodeEquals(err, errCodeNoSuchBucket, errCodeMetadataConfigurationNotFound) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting S3 Bucket Metadata Configuration (%s)", bucket), err.Error())

		return
	}

	_, err = tfresource.RetryUntilNotFound(ctx, bucketPropagationTimeout, func(ctx context.Context) (any, error) {
		return findBucketMetadataConfigurationByTwoPartKey(ctx, conn, bucket, expectedBucketOwner)
	})

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for S3 Bucket Metadata Configuration (%s) delete", bucket), err.Error())

		return
	}
}

func waitBucketMetadataInventoryTableConfigurationCreated(ctx context.Context, conn *s3.Client, bucket, expectedBucketOwner string, timeout time.Duration) (*awstypes.InventoryTableConfigurationResult, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{inventoryTableConfigurationStatusCreating},
		Target:                    []string{inventoryTableConfigurationStatusActive, inventoryTableConfigurationStatusBackfilling},
		Refresh:                   statusBucketMetadataInventoryTableConfiguration(conn, bucket, expectedBucketOwner),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.InventoryTableConfigurationResult); ok {
		if v := output.Error; v != nil {
			retry.SetLastError(err, fmt.Errorf("%s: %s", aws.ToString(v.ErrorCode), aws.ToString(v.ErrorMessage)))
		}

		return output, err
	}

	return nil, err
}

func waitBucketMetadataJournalTableConfigurationCreated(ctx context.Context, conn *s3.Client, bucket, expectedBucketOwner string, timeout time.Duration) (*awstypes.JournalTableConfigurationResult, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{journalTableConfigurationStatusCreating},
		Target:                    []string{journalTableConfigurationStatusActive},
		Refresh:                   statusBucketMetadataJournalTableConfiguration(conn, bucket, expectedBucketOwner),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.JournalTableConfigurationResult); ok {
		if v := output.Error; v != nil {
			retry.SetLastError(err, fmt.Errorf("%s: %s", aws.ToString(v.ErrorCode), aws.ToString(v.ErrorMessage)))
		}

		return output, err
	}

	return nil, err
}

func statusBucketMetadataInventoryTableConfiguration(conn *s3.Client, bucket, expectedBucketOwner string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		mcr, err := findBucketMetadataConfigurationByTwoPartKey(ctx, conn, bucket, expectedBucketOwner)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		output := mcr.InventoryTableConfigurationResult
		if output == nil {
			return nil, "", nil
		}

		return output, aws.ToString(output.TableStatus), nil
	}
}

func statusBucketMetadataJournalTableConfiguration(conn *s3.Client, bucket, expectedBucketOwner string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		mcr, err := findBucketMetadataConfigurationByTwoPartKey(ctx, conn, bucket, expectedBucketOwner)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		output := mcr.JournalTableConfigurationResult
		if output == nil {
			return nil, "", nil
		}

		return output, aws.ToString(output.TableStatus), nil
	}
}

func findBucketMetadataConfigurationByTwoPartKey(ctx context.Context, conn *s3.Client, bucket, expectedBucketOwner string) (*awstypes.MetadataConfigurationResult, error) {
	input := s3.GetBucketMetadataConfigurationInput{
		Bucket: aws.String(bucket),
	}
	if expectedBucketOwner != "" {
		input.ExpectedBucketOwner = aws.String(expectedBucketOwner)
	}

	return findBucketMetadataConfiguration(ctx, conn, &input)
}

func findBucketMetadataConfiguration(ctx context.Context, conn *s3.Client, input *s3.GetBucketMetadataConfigurationInput) (*awstypes.MetadataConfigurationResult, error) {
	output, err := conn.GetBucketMetadataConfiguration(ctx, input)

	if tfawserr.ErrCodeEquals(err, errCodeNoSuchBucket, errCodeMetadataConfigurationNotFound) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.GetBucketMetadataConfigurationResult == nil || output.GetBucketMetadataConfigurationResult.MetadataConfigurationResult == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output.GetBucketMetadataConfigurationResult.MetadataConfigurationResult, nil
}

func getMetadataTableEncryptionConfigurationModels(ctx context.Context, data *bucketMetadataConfigurationResourceModel) (fwtypes.ListNestedObjectValueOf[metadataTableEncryptionConfigurationModel], fwtypes.ListNestedObjectValueOf[metadataTableEncryptionConfigurationModel], diag.Diagnostics) {
	var diags diag.Diagnostics
	nullMetadataTableEncryptionConfigurationModel := fwtypes.NewListNestedObjectValueOfNull[metadataTableEncryptionConfigurationModel](ctx)

	metadataConfigurationModel, d := data.MetadataConfiguration.ToPtr(ctx)
	diags.Append(d...)
	if diags.HasError() {
		return nullMetadataTableEncryptionConfigurationModel, nullMetadataTableEncryptionConfigurationModel, diags
	}

	if metadataConfigurationModel == nil {
		return nullMetadataTableEncryptionConfigurationModel, nullMetadataTableEncryptionConfigurationModel, diags
	}

	inventoryTableConfigurationModel, d := metadataConfigurationModel.InventoryTableConfiguration.ToPtr(ctx)
	diags.Append(d...)
	if diags.HasError() {
		return nullMetadataTableEncryptionConfigurationModel, nullMetadataTableEncryptionConfigurationModel, diags
	}

	if inventoryTableConfigurationModel == nil {
		return nullMetadataTableEncryptionConfigurationModel, nullMetadataTableEncryptionConfigurationModel, diags
	}

	journalTableConfigurationModel, d := metadataConfigurationModel.JournalTableConfiguration.ToPtr(ctx)
	diags.Append(d...)
	if diags.HasError() {
		return nullMetadataTableEncryptionConfigurationModel, nullMetadataTableEncryptionConfigurationModel, diags
	}

	if journalTableConfigurationModel == nil {
		return inventoryTableConfigurationModel.EncryptionConfiguration, nullMetadataTableEncryptionConfigurationModel, diags
	}

	return inventoryTableConfigurationModel.EncryptionConfiguration, journalTableConfigurationModel.EncryptionConfiguration, diags
}

func setMetadataTableEncryptionConfigurationModels(ctx context.Context, data *bucketMetadataConfigurationResourceModel, inventoryEncryptionConfiguration fwtypes.ListNestedObjectValueOf[metadataTableEncryptionConfigurationModel], journalEncryptionConfiguration fwtypes.ListNestedObjectValueOf[metadataTableEncryptionConfigurationModel]) diag.Diagnostics {
	var diags diag.Diagnostics

	metadataConfigurationModel, d := data.MetadataConfiguration.ToPtr(ctx)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}

	inventoryTableConfigurationModel, d := metadataConfigurationModel.InventoryTableConfiguration.ToPtr(ctx)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}

	journalTableConfigurationModel, d := metadataConfigurationModel.JournalTableConfiguration.ToPtr(ctx)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}

	inventoryTableConfigurationModel.EncryptionConfiguration = inventoryEncryptionConfiguration
	journalTableConfigurationModel.EncryptionConfiguration = journalEncryptionConfiguration

	metadataConfigurationModel.InventoryTableConfiguration, d = fwtypes.NewListNestedObjectValueOfPtr(ctx, inventoryTableConfigurationModel)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}

	metadataConfigurationModel.JournalTableConfiguration, d = fwtypes.NewListNestedObjectValueOfPtr(ctx, journalTableConfigurationModel)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}

	data.MetadataConfiguration, d = fwtypes.NewListNestedObjectValueOfPtr(ctx, metadataConfigurationModel)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}

	return diags
}

type bucketMetadataConfigurationResourceModel struct {
	framework.WithRegionModel
	Bucket                types.String                                                `tfsdk:"bucket"`
	ExpectedBucketOwner   types.String                                                `tfsdk:"expected_bucket_owner"`
	MetadataConfiguration fwtypes.ListNestedObjectValueOf[metadataConfigurationModel] `tfsdk:"metadata_configuration"`
	Timeouts              timeouts.Value                                              `tfsdk:"timeouts"`
}

type metadataConfigurationModel struct {
	Destination                 fwtypes.ListNestedObjectValueOf[destinationResultModel]           `tfsdk:"destination"`
	InventoryTableConfiguration fwtypes.ListNestedObjectValueOf[inventoryTableConfigurationModel] `tfsdk:"inventory_table_configuration"`
	JournalTableConfiguration   fwtypes.ListNestedObjectValueOf[journalTableConfigurationModel]   `tfsdk:"journal_table_configuration"`
}

type destinationResultModel struct {
	TableBucketARN  fwtypes.ARN                                     `tfsdk:"table_bucket_arn"`
	TableBucketType fwtypes.StringEnum[awstypes.S3TablesBucketType] `tfsdk:"table_bucket_type"`
	TableNamespace  types.String                                    `tfsdk:"table_namespace"`
}

type inventoryTableConfigurationModel struct {
	ConfigurationState      fwtypes.StringEnum[awstypes.InventoryConfigurationState]                   `tfsdk:"configuration_state"`
	EncryptionConfiguration fwtypes.ListNestedObjectValueOf[metadataTableEncryptionConfigurationModel] `tfsdk:"encryption_configuration"`
	TableARN                fwtypes.ARN                                                                `tfsdk:"table_arn"`
	TableName               types.String                                                               `tfsdk:"table_name"`
}

type journalTableConfigurationModel struct {
	EncryptionConfiguration fwtypes.ListNestedObjectValueOf[metadataTableEncryptionConfigurationModel] `tfsdk:"encryption_configuration"`
	RecordExpiration        fwtypes.ListNestedObjectValueOf[recordExpirationModel]                     `tfsdk:"record_expiration"`
	TableARN                fwtypes.ARN                                                                `tfsdk:"table_arn"`
	TableName               types.String                                                               `tfsdk:"table_name"`
}

type metadataTableEncryptionConfigurationModel struct {
	KMSKeyARN    fwtypes.ARN                                    `tfsdk:"kms_key_arn"`
	SSEAlgorithm fwtypes.StringEnum[awstypes.TableSseAlgorithm] `tfsdk:"sse_algorithm"`
}

type recordExpirationModel struct {
	Days       types.Int32                                  `tfsdk:"days"`
	Expiration fwtypes.StringEnum[awstypes.ExpirationState] `tfsdk:"expiration"`
}
