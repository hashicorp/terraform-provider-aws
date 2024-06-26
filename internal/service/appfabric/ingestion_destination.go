// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appfabric

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/appfabric"
	awstypes "github.com/aws/aws-sdk-go-v2/service/appfabric/types"
	uuid "github.com/hashicorp/go-uuid"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource(name="Ingestion Destination")
// @Tags(identifierAttribute="arn")
func newIngestionDestinationResource(context.Context) (resource.ResourceWithConfigure, error) {
	r := &ingestionDestinationResource{}

	r.SetDefaultCreateTimeout(5 * time.Minute)
	r.SetDefaultUpdateTimeout(5 * time.Minute)
	r.SetDefaultDeleteTimeout(5 * time.Minute)

	return r, nil
}

type ingestionDestinationResource struct {
	framework.ResourceWithConfigure
	framework.WithImportByID
	framework.WithTimeouts
}

func (*ingestionDestinationResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_appfabric_ingestion_destination"
}

func (r *ingestionDestinationResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"app_bundle_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrID:  framework.IDAttribute(),
			"ingestion_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			"destination_configuration": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[destinationConfigurationModel](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"audit_log": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[auditLogDestinationConfigurationModel](ctx),
							Validators: []validator.List{
								listvalidator.IsRequired(),
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Blocks: map[string]schema.Block{
									names.AttrDestination: schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[destinationModel](ctx),
										Validators: []validator.List{
											listvalidator.IsRequired(),
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Blocks: map[string]schema.Block{
												"firehose_stream": schema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[firehoseStreamModel](ctx),
													Validators: []validator.List{
														listvalidator.SizeAtMost(1),
													},
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															"stream_name": schema.StringAttribute{
																Required: true,
																Validators: []validator.String{
																	stringvalidator.LengthBetween(3, 64),
																},
															},
														},
													},
												},
												names.AttrS3Bucket: schema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[s3BucketModel](ctx),
													Validators: []validator.List{
														listvalidator.SizeAtMost(1),
													},
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															names.AttrBucketName: schema.StringAttribute{
																Required: true,
																Validators: []validator.String{
																	stringvalidator.LengthBetween(3, 63),
																},
															},
															names.AttrPrefix: schema.StringAttribute{
																Optional: true,
																Validators: []validator.String{
																	stringvalidator.LengthBetween(1, 120),
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
						},
					},
				},
			},
			"processing_configuration": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[processingConfigurationModel](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtMost(1),
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"audit_log": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[auditLogProcessingConfigurationModel](ctx),
							Validators: []validator.List{
								listvalidator.IsRequired(),
								listvalidator.SizeAtMost(1),
							},
							PlanModifiers: []planmodifier.List{
								listplanmodifier.RequiresReplace(),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrFormat: schema.StringAttribute{
										CustomType: fwtypes.StringEnumType[awstypes.Format](),
										Required:   true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.RequiresReplace(),
										},
									},
									names.AttrSchema: schema.StringAttribute{
										CustomType: fwtypes.StringEnumType[awstypes.Schema](),
										Required:   true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.RequiresReplace(),
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
				Update: true,
				Delete: true,
			}),
		},
	}
}

func (r *ingestionDestinationResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data ingestionDestinationResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().AppFabricClient(ctx)

	input := &appfabric.CreateIngestionDestinationInput{}
	response.Diagnostics.Append(fwflex.Expand(ctx, data, input)...)
	if response.Diagnostics.HasError() {
		return
	}

	// AutoFlEx doesn't yet handle union types.
	if !data.DestinationConfiguration.IsNull() {
		destinationConfigurationData, diags := data.DestinationConfiguration.ToPtr(ctx)
		response.Diagnostics.Append(diags...)
		if response.Diagnostics.HasError() {
			return
		}

		destinationConfiguration, diags := expandDestinationConfiguration(ctx, destinationConfigurationData)
		response.Diagnostics.Append(diags...)
		if response.Diagnostics.HasError() {
			return
		}

		input.DestinationConfiguration = destinationConfiguration
	}

	if !data.ProcessingConfiguration.IsNull() {
		processingConfigurationData, diags := data.ProcessingConfiguration.ToPtr(ctx)
		response.Diagnostics.Append(diags...)
		if response.Diagnostics.HasError() {
			return
		}

		processingConfiguration, diags := expandProcessingConfiguration(ctx, processingConfigurationData)
		response.Diagnostics.Append(diags...)
		if response.Diagnostics.HasError() {
			return
		}

		input.ProcessingConfiguration = processingConfiguration
	}

	// Additional fields.
	input.AppBundleIdentifier = aws.String(data.AppBundleARN.ValueString())
	input.ClientToken = aws.String(errs.Must(uuid.GenerateUUID()))
	input.IngestionIdentifier = aws.String(data.IngestionARN.ValueString())
	input.Tags = getTagsIn(ctx)

	output, err := conn.CreateIngestionDestination(ctx, input)

	if err != nil {
		response.Diagnostics.AddError("creating AppFabric Ingestion Destination", err.Error())

		return
	}

	// Set values for unknowns.
	data.ARN = fwflex.StringToFramework(ctx, output.IngestionDestination.Arn)
	data.setID()

	if _, err := waitIngestionDestinationActive(ctx, conn, data.AppBundleARN.ValueString(), data.IngestionARN.ValueString(), data.ARN.ValueString(), r.CreateTimeout(ctx, data.Timeouts)); err != nil {
		response.State.SetAttribute(ctx, path.Root(names.AttrID), data.ID) // Set 'id' so as to taint the resource.
		response.Diagnostics.AddError(fmt.Sprintf("waiting for AppFabric Ingestion Destination (%s) create", data.ID.ValueString()), err.Error())

		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *ingestionDestinationResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data ingestionDestinationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	if err := data.InitFromID(); err != nil {
		response.Diagnostics.AddError("parsing resource ID", err.Error())

		return
	}

	conn := r.Meta().AppFabricClient(ctx)

	output, err := findIngestionDestinationByThreePartKey(ctx, conn, data.AppBundleARN.ValueString(), data.IngestionARN.ValueString(), data.ARN.ValueString())

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading AppFabric Ingestion Destination (%s)", data.ID.ValueString()), err.Error())

		return
	}

	// Set attributes for import.
	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	// AutoFlEx doesn't yet handle union types.
	if output.DestinationConfiguration != nil {
		destinationConfigurationData, diags := flattenDestinationConfiguration(ctx, output.DestinationConfiguration)
		response.Diagnostics.Append(diags...)
		if response.Diagnostics.HasError() {
			return
		}

		data.DestinationConfiguration = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, destinationConfigurationData)
	}

	if output.ProcessingConfiguration != nil {
		processingConfigurationData, diags := flattenProcessingConfiguration(ctx, output.ProcessingConfiguration)
		response.Diagnostics.Append(diags...)
		if response.Diagnostics.HasError() {
			return
		}

		data.ProcessingConfiguration = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, processingConfigurationData)
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *ingestionDestinationResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var old, new ingestionDestinationResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.State.Get(ctx, &old)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().AppFabricClient(ctx)

	if !old.DestinationConfiguration.Equal(new.DestinationConfiguration) {
		input := &appfabric.UpdateIngestionDestinationInput{}
		response.Diagnostics.Append(fwflex.Expand(ctx, new, input)...)
		if response.Diagnostics.HasError() {
			return
		}

		// AutoFlEx doesn't yet handle union types.
		if !new.DestinationConfiguration.IsNull() {
			destinationConfigurationData, diags := new.DestinationConfiguration.ToPtr(ctx)
			response.Diagnostics.Append(diags...)
			if response.Diagnostics.HasError() {
				return
			}

			destinationConfiguration, diags := expandDestinationConfiguration(ctx, destinationConfigurationData)
			response.Diagnostics.Append(diags...)
			if response.Diagnostics.HasError() {
				return
			}

			input.DestinationConfiguration = destinationConfiguration
		}

		// Additional fields.
		input.AppBundleIdentifier = aws.String(new.AppBundleARN.ValueString())
		input.IngestionDestinationIdentifier = aws.String(new.ARN.ValueString())
		input.IngestionIdentifier = aws.String(new.IngestionARN.ValueString())

		_, err := conn.UpdateIngestionDestination(ctx, input)

		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("updating AppFabric Ingestion Destination (%s)", new.ID.ValueString()), err.Error())

			return
		}

		if _, err := waitIngestionDestinationActive(ctx, conn, new.AppBundleARN.ValueString(), new.IngestionARN.ValueString(), new.ARN.ValueString(), r.UpdateTimeout(ctx, new.Timeouts)); err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("waiting for AppFabric Ingestion Destination (%s) update", new.ID.ValueString()), err.Error())

			return
		}
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *ingestionDestinationResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data ingestionDestinationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().AppFabricClient(ctx)

	_, err := conn.DeleteIngestionDestination(ctx, &appfabric.DeleteIngestionDestinationInput{
		AppBundleIdentifier:            aws.String(data.AppBundleARN.ValueString()),
		IngestionDestinationIdentifier: aws.String(data.ARN.ValueString()),
		IngestionIdentifier:            aws.String(data.IngestionARN.ValueString()),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting AppFabric Ingestion Destination (%s)", data.ID.ValueString()), err.Error())

		return
	}

	if _, err = waitIngestionDestinationDeleted(ctx, conn, data.AppBundleARN.ValueString(), data.IngestionARN.ValueString(), data.ARN.ValueString(), r.DeleteTimeout(ctx, data.Timeouts)); err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for AppFabric Ingestion Destination (%s) delete", data.ID.ValueString()), err.Error())

		return
	}
}

func (r *ingestionDestinationResource) ConfigValidators(context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcevalidator.AtLeastOneOf(
			path.MatchRoot("destination_configuration").AtListIndex(0).AtName("audit_log").AtListIndex(0).AtName(names.AttrDestination).AtListIndex(0).AtName("firehose_stream"),
			path.MatchRoot("destination_configuration").AtListIndex(0).AtName("audit_log").AtListIndex(0).AtName(names.AttrDestination).AtListIndex(0).AtName(names.AttrS3Bucket),
		),
	}
}

func (r *ingestionDestinationResource) ModifyPlan(ctx context.Context, request resource.ModifyPlanRequest, response *resource.ModifyPlanResponse) {
	r.SetTagsAll(ctx, request, response)
}

func findIngestionDestinationByThreePartKey(ctx context.Context, conn *appfabric.Client, appBundleARN, ingestionARN, arn string) (*awstypes.IngestionDestination, error) {
	in := &appfabric.GetIngestionDestinationInput{
		AppBundleIdentifier:            aws.String(appBundleARN),
		IngestionDestinationIdentifier: aws.String(arn),
		IngestionIdentifier:            aws.String(ingestionARN),
	}

	output, err := conn.GetIngestionDestination(ctx, in)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.IngestionDestination == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return output.IngestionDestination, nil
}

func statusIngestionDestination(ctx context.Context, conn *appfabric.Client, appBundleARN, ingestionARN, arn string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := findIngestionDestinationByThreePartKey(ctx, conn, appBundleARN, ingestionARN, arn)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, string(out.Status), nil
	}
}

func waitIngestionDestinationActive(ctx context.Context, conn *appfabric.Client, appBundleARN, ingestionARN, arn string, timeout time.Duration) (*awstypes.IngestionDestination, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending: []string{},
		Target:  enum.Slice(awstypes.IngestionDestinationStatusActive),
		Refresh: statusIngestionDestination(ctx, conn, appBundleARN, ingestionARN, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.IngestionDestination); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.StatusReason)))

		return output, err
	}

	return nil, err
}

func waitIngestionDestinationDeleted(ctx context.Context, conn *appfabric.Client, appBundleARN, ingestionARN, arn string, timeout time.Duration) (*awstypes.IngestionDestination, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.IngestionDestinationStatusActive),
		Target:  []string{},
		Refresh: statusIngestionDestination(ctx, conn, appBundleARN, ingestionARN, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.IngestionDestination); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.StatusReason)))

		return output, err
	}

	return nil, err
}

type ingestionDestinationResourceModel struct {
	AppBundleARN             fwtypes.ARN                                                    `tfsdk:"app_bundle_arn"`
	ARN                      types.String                                                   `tfsdk:"arn"`
	DestinationConfiguration fwtypes.ListNestedObjectValueOf[destinationConfigurationModel] `tfsdk:"destination_configuration"`
	ID                       types.String                                                   `tfsdk:"id"`
	IngestionARN             fwtypes.ARN                                                    `tfsdk:"ingestion_arn"`
	ProcessingConfiguration  fwtypes.ListNestedObjectValueOf[processingConfigurationModel]  `tfsdk:"processing_configuration"`
	Tags                     types.Map                                                      `tfsdk:"tags"`
	TagsAll                  types.Map                                                      `tfsdk:"tags_all"`
	Timeouts                 timeouts.Value                                                 `tfsdk:"timeouts"`
}

const (
	ingestionDestinationResourceIDPartCount = 3
)

func (m *ingestionDestinationResourceModel) InitFromID() error {
	parts, err := flex.ExpandResourceId(m.ID.ValueString(), ingestionDestinationResourceIDPartCount, false)
	if err != nil {
		return err
	}

	m.AppBundleARN = fwtypes.ARNValue(parts[0])
	m.IngestionARN = fwtypes.ARNValue(parts[1])
	m.ARN = types.StringValue(parts[2])

	return nil
}

func (m *ingestionDestinationResourceModel) setID() {
	m.ID = types.StringValue(errs.Must(flex.FlattenResourceId([]string{m.AppBundleARN.ValueString(), m.IngestionARN.ValueString(), m.ARN.ValueString()}, ingestionDestinationResourceIDPartCount, false)))
}

type destinationConfigurationModel struct {
	AuditLog fwtypes.ListNestedObjectValueOf[auditLogDestinationConfigurationModel] `tfsdk:"audit_log"`
}

type auditLogDestinationConfigurationModel struct {
	Destination fwtypes.ListNestedObjectValueOf[destinationModel] `tfsdk:"destination"`
}

type destinationModel struct {
	FirehoseStream fwtypes.ListNestedObjectValueOf[firehoseStreamModel] `tfsdk:"firehose_stream"`
	S3Bucket       fwtypes.ListNestedObjectValueOf[s3BucketModel]       `tfsdk:"s3_bucket"`
}

type firehoseStreamModel struct {
	StreamName types.String `tfsdk:"stream_name"`
}

type s3BucketModel struct {
	BucketName types.String `tfsdk:"bucket_name"`
	Prefix     types.String `tfsdk:"prefix"`
}

type processingConfigurationModel struct {
	AuditLog fwtypes.ListNestedObjectValueOf[auditLogProcessingConfigurationModel] `tfsdk:"audit_log"`
}

type auditLogProcessingConfigurationModel struct {
	Format fwtypes.StringEnum[awstypes.Format] `tfsdk:"format"`
	Schema fwtypes.StringEnum[awstypes.Schema] `tfsdk:"schema"`
}

func expandDestinationConfiguration(ctx context.Context, destinationConfigurationData *destinationConfigurationModel) (awstypes.DestinationConfiguration, diag.Diagnostics) {
	var diags diag.Diagnostics

	if !destinationConfigurationData.AuditLog.IsNull() {
		auditLogDestinationConfigurationData, d := destinationConfigurationData.AuditLog.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		destinationData, d := auditLogDestinationConfigurationData.Destination.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		destination, d := expandDestination(ctx, destinationData)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		apiObject := &awstypes.DestinationConfigurationMemberAuditLog{
			Value: awstypes.AuditLogDestinationConfiguration{
				Destination: destination,
			},
		}

		return apiObject, diags
	}

	return nil, diags
}

func expandDestination(ctx context.Context, destinationData *destinationModel) (awstypes.Destination, diag.Diagnostics) {
	var diags diag.Diagnostics

	if !destinationData.FirehoseStream.IsNull() {
		firehoseStreamData, d := destinationData.FirehoseStream.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		apiObject := &awstypes.DestinationMemberFirehoseStream{}
		diags.Append(fwflex.Expand(ctx, firehoseStreamData, &apiObject.Value)...)
		if diags.HasError() {
			return nil, diags
		}

		return apiObject, diags
	}
	if !destinationData.S3Bucket.IsNull() {
		s3BucketData, d := destinationData.S3Bucket.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		apiObject := &awstypes.DestinationMemberS3Bucket{}
		diags.Append(fwflex.Expand(ctx, s3BucketData, &apiObject.Value)...)
		if diags.HasError() {
			return nil, diags
		}

		return apiObject, diags
	}

	return nil, diags
}

func expandProcessingConfiguration(ctx context.Context, processingConfigurationData *processingConfigurationModel) (awstypes.ProcessingConfiguration, diag.Diagnostics) {
	var diags diag.Diagnostics

	if !processingConfigurationData.AuditLog.IsNull() {
		auditLogProcessingConfigurationData, d := processingConfigurationData.AuditLog.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		apiObject := &awstypes.ProcessingConfigurationMemberAuditLog{}
		diags.Append(fwflex.Expand(ctx, auditLogProcessingConfigurationData, &apiObject.Value)...)
		if diags.HasError() {
			return nil, diags
		}

		return apiObject, diags
	}

	return nil, diags
}

func flattenDestinationConfiguration(ctx context.Context, apiObject awstypes.DestinationConfiguration) (*destinationConfigurationModel, diag.Diagnostics) {
	var diags diag.Diagnostics
	var destinationConfigurationData *destinationConfigurationModel

	switch v := apiObject.(type) {
	case *awstypes.DestinationConfigurationMemberAuditLog:
		destinationData, d := flattenDestination(ctx, v.Value.Destination)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		auditLogDestinationConfigurationData := &auditLogDestinationConfigurationModel{
			Destination: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, destinationData),
		}
		destinationConfigurationData = &destinationConfigurationModel{
			AuditLog: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, auditLogDestinationConfigurationData),
		}
	}

	return destinationConfigurationData, diags
}

func flattenDestination(ctx context.Context, apiObject awstypes.Destination) (*destinationModel, diag.Diagnostics) {
	var diags diag.Diagnostics
	var destinationData *destinationModel

	switch v := apiObject.(type) {
	case *awstypes.DestinationMemberFirehoseStream:
		var firehoseStreamData firehoseStreamModel
		d := fwflex.Flatten(ctx, v.Value, &firehoseStreamData)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		destinationData = &destinationModel{
			FirehoseStream: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &firehoseStreamData),
			S3Bucket:       fwtypes.NewListNestedObjectValueOfNull[s3BucketModel](ctx),
		}

	case *awstypes.DestinationMemberS3Bucket:
		var s3BucketData s3BucketModel
		d := fwflex.Flatten(ctx, v.Value, &s3BucketData)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		destinationData = &destinationModel{
			FirehoseStream: fwtypes.NewListNestedObjectValueOfNull[firehoseStreamModel](ctx),
			S3Bucket:       fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &s3BucketData),
		}
	}

	return destinationData, diags
}

func flattenProcessingConfiguration(ctx context.Context, apiObject awstypes.ProcessingConfiguration) (*processingConfigurationModel, diag.Diagnostics) {
	var diags diag.Diagnostics
	var processingConfigurationData *processingConfigurationModel

	switch v := apiObject.(type) {
	case *awstypes.ProcessingConfigurationMemberAuditLog:
		var auditLogProcessingConfigurationData auditLogProcessingConfigurationModel
		d := fwflex.Flatten(ctx, v.Value, &auditLogProcessingConfigurationData)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		processingConfigurationData = &processingConfigurationModel{
			AuditLog: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &auditLogProcessingConfigurationData),
		}
	}

	return processingConfigurationData, diags
}
