// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appfabric

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/appfabric"
	awstypes "github.com/aws/aws-sdk-go-v2/service/appfabric/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for resource registration to the Provider. DO NOT EDIT.
// @FrameworkResource(name="Ingestion Destination")
func newResourceIngestionDestination(context.Context) (resource.ResourceWithConfigure, error) {
	r := &ingestionDestinationResource{}

	r.SetDefaultCreateTimeout(5 * time.Minute)
	r.SetDefaultUpdateTimeout(5 * time.Minute)
	r.SetDefaultDeleteTimeout(5 * time.Minute)

	return r, nil
}

const (
	ResNameIngestionDestination = "Ingestion Destination"
)

type ingestionDestinationResource struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
}

func (r *ingestionDestinationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_appfabric_ingestion_destination"
}

func (r *ingestionDestinationResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"app_bundle_identifier": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrARN:   framework.IDAttribute(),
			names.AttrID:    framework.IDAttribute(),
			"ingestion_arn": framework.ARNAttributeComputedOnly(),
			"ingestion_identifier": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"destination_configuration": schema.ListNestedBlock{
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"audit_log": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[destinationConfigurationAuditLogModel](ctx),
							Validators: []validator.List{
								listvalidator.IsRequired(),
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Blocks: map[string]schema.Block{
									"destination": schema.ListNestedBlock{
										Validators: []validator.List{
											listvalidator.IsRequired(),
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"firehose_stream": schema.ListAttribute{
													Optional:   true,
													CustomType: fwtypes.NewListNestedObjectTypeOf[firehoseStreamModel](ctx),
													ElementType: types.ObjectType{
														AttrTypes: map[string]attr.Type{
															"stream_name": types.StringType,
														},
													},
													PlanModifiers: []planmodifier.List{
														listplanmodifier.UseStateForUnknown(),
													},
												},
												"s3_bucket": schema.ListAttribute{
													Optional:   true,
													CustomType: fwtypes.NewListNestedObjectTypeOf[s3BucketModel](ctx),
													ElementType: types.ObjectType{
														AttrTypes: map[string]attr.Type{
															"bucket_name": types.StringType,
															"prefix":      types.StringType,
														},
													},
													PlanModifiers: []planmodifier.List{
														listplanmodifier.UseStateForUnknown(),
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
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"audit_log": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[processingConfigurationAuditLogModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"format": schema.StringAttribute{
										Required: true,
										Validators: []validator.String{
											stringvalidator.LengthBetween(1, 255),
											stringvalidator.RegexMatches(
												regexache.MustCompile(`^json$|^parquet$`),
												"Valid values one of JSON | PARQUET",
											),
										},
									},
									"schema": schema.StringAttribute{
										Required: true,
										Validators: []validator.String{
											stringvalidator.LengthBetween(1, 255),
											stringvalidator.RegexMatches(
												regexache.MustCompile(`^ocsf$|^raw$`),
												"Valid values one of OCSF | RAW",
											),
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

func (r *ingestionDestinationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan resourceIngestionDestinationModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().AppFabricClient(ctx)

	var destinationConfigurationData []destinationConfigurationModel
	resp.Diagnostics.Append(plan.DestinationConfiguration.ElementsAs(ctx, &destinationConfigurationData, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	destinationConfiguration, d := expandDestinationConfiguration(ctx, destinationConfigurationData)
	resp.Diagnostics.Append(d...)

	var processingConfigurationData []processingConfigurationModel
	resp.Diagnostics.Append(plan.ProcessingConfiguration.ElementsAs(ctx, &processingConfigurationData, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	processingConfiguration, d := expandProcessingConfiguration(ctx, processingConfigurationData)
	resp.Diagnostics.Append(d...)

	in := &appfabric.CreateIngestionDestinationInput{}
	resp.Diagnostics.Append(fwflex.Expand(ctx, plan, in)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in.DestinationConfiguration = destinationConfiguration
	in.ProcessingConfiguration = processingConfiguration

	out, err := conn.CreateIngestionDestination(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.AppFabric, create.ErrActionWaitingForCreation, ResNameIngestionDestination, plan.ID.String(), err),
			err.Error(),
		)
		return
	}

	// Set values for unknowns.
	ingestionDestination := out.IngestionDestination
	plan.IngestionDestinationArn = fwflex.StringToFramework(ctx, ingestionDestination.Arn)
	plan.setID()

	iDest, err := waitIngestionDestinationCreated(ctx, conn, plan.IngestionDestinationArn.ValueString(), plan.AppBundleIdentifier.ValueString(), plan.IngestionIdentifier.ValueString(), r.CreateTimeout(ctx, plan.Timeouts))
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("Waiting for Ingestion Destination (%s) to be created", plan.IngestionDestinationArn.ValueString()), err.Error())

		return
	}

	// Set values for unknowns after creation is complete.
	plan.IngestionDestinationArn = fwflex.StringToFramework(ctx, iDest.Arn)
	plan.IngestionArn = fwflex.StringToFramework(ctx, iDest.IngestionArn)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *ingestionDestinationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state resourceIngestionDestinationModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	conn := r.Meta().AppFabricClient(ctx)

	if err := state.InitFromID(); err != nil {
		resp.Diagnostics.AddError("parsing resource ID", err.Error())

		return
	}

	out, err := findIngestionDestinationByID(ctx, conn, state.IngestionDestinationArn.ValueString(), state.AppBundleIdentifier.ValueString(), state.IngestionIdentifier.ValueString())

	if tfresource.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("reading Ingestion Destination ID (%s)", state.IngestionDestinationArn.ValueString()), err.Error())

		return
	}

	resp.Diagnostics.Append(fwflex.Flatten(ctx, out, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	destinationConfigurationOutput, d := flattenDestinationConfiguration(ctx, out.DestinationConfiguration)
	resp.Diagnostics.Append(d...)
	state.DestinationConfiguration = destinationConfigurationOutput

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ingestionDestinationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var old, new resourceIngestionDestinationModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &new)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(req.State.Get(ctx, &old)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().AppFabricClient(ctx)

	// Check if updates are necessary based on the changed attributes
	if !old.DestinationConfiguration.Equal(new.DestinationConfiguration) {
		var destinationConfigurationData []destinationConfigurationModel
		resp.Diagnostics.Append(new.DestinationConfiguration.ElementsAs(ctx, &destinationConfigurationData, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		destinationConfiguration, diags := expandDestinationConfiguration(ctx, destinationConfigurationData)
		resp.Diagnostics.Append(diags...)
		if diags.HasError() {
			return
		}

		input := &appfabric.UpdateIngestionDestinationInput{
			IngestionDestinationIdentifier: aws.String(new.IngestionDestinationArn.ValueString()),
		}
		resp.Diagnostics.Append(fwflex.Expand(ctx, new, input)...)
		if resp.Diagnostics.HasError() {
			return
		}

		input.DestinationConfiguration = destinationConfiguration

		_, err := conn.UpdateIngestionDestination(ctx, input)
		if err != nil {
			resp.Diagnostics.AddError(
				"Failed to update Ingestion Destination",
				fmt.Sprintf("Error updating Ingestion Destination with ID %s: %s", new.IngestionDestinationArn.String(), err.Error()),
			)
			return
		}

		if _, err = waitIngestionDestinationUpdated(ctx, conn, new.IngestionDestinationArn.ValueString(), new.AppBundleIdentifier.ValueString(), new.IngestionIdentifier.ValueString(), r.CreateTimeout(ctx, new.Timeouts)); err != nil {
			resp.Diagnostics.AddError(fmt.Sprintf("Waiting for Ingestion Destination (%s) to be updated", new.IngestionDestinationArn.ValueString()), err.Error())

			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &new)...)
}

func (r *ingestionDestinationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state resourceIngestionDestinationModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().AppFabricClient(ctx)

	_, err := conn.DeleteIngestionDestination(ctx, &appfabric.DeleteIngestionDestinationInput{
		IngestionDestinationIdentifier: aws.String(state.IngestionDestinationArn.ValueString()),
		AppBundleIdentifier:            aws.String(state.AppBundleIdentifier.ValueString()),
		IngestionIdentifier:            aws.String(state.IngestionIdentifier.ValueString()),
	})

	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.AppFabric, create.ErrActionDeleting, ResNameIngestionDestination, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	if _, err = waitIngestionDestinationDeleted(ctx, conn, state.IngestionDestinationArn.ValueString(), state.AppBundleIdentifier.ValueString(), state.IngestionIdentifier.ValueString(), r.CreateTimeout(ctx, state.Timeouts)); err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("Waiting for Ingestion Destination (%s) to be deleted", state.IngestionDestinationArn.ValueString()), err.Error())

		return
	}

}

func (r *ingestionDestinationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrID), req, resp)
}

func waitIngestionDestinationCreated(ctx context.Context, conn *appfabric.Client, ingestionDestinationArn, appBundleIdentifier, ingestionIdentifier string, timeout time.Duration) (*awstypes.IngestionDestination, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{},
		Target:                    enum.Slice(awstypes.IngestionDestinationStatusActive, awstypes.IngestionDestinationStatusActive),
		Refresh:                   statusIngestionDestination(ctx, conn, ingestionDestinationArn, appBundleIdentifier, ingestionIdentifier),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.IngestionDestination); ok {
		return out, err
	}

	return nil, err
}

func waitIngestionDestinationUpdated(ctx context.Context, conn *appfabric.Client, ingestionDestinationArn, appBundleIdentifier, ingestionIdentifier string, timeout time.Duration) (*awstypes.IngestionDestination, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{},
		Target:                    enum.Slice(awstypes.IngestionDestinationStatusActive, awstypes.IngestionDestinationStatusActive),
		Refresh:                   statusIngestionDestination(ctx, conn, ingestionDestinationArn, appBundleIdentifier, ingestionIdentifier),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.IngestionDestination); ok {
		return out, err
	}

	return nil, err
}

func waitIngestionDestinationDeleted(ctx context.Context, conn *appfabric.Client, ingestionDestinationArn, appBundleIdentifier, ingestionIdentifier string, timeout time.Duration) (*awstypes.IngestionDestination, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.IngestionDestinationStatusActive, awstypes.IngestionDestinationStatusActive),
		Target:  []string{},
		Refresh: statusIngestionDestination(ctx, conn, ingestionDestinationArn, appBundleIdentifier, ingestionIdentifier),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.IngestionDestination); ok {
		return out, err
	}

	return nil, err
}

func statusIngestionDestination(ctx context.Context, conn *appfabric.Client, ingestionDestinationArn, appBundleIdentifier, ingestionIdentifier string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := findIngestionDestinationByID(ctx, conn, ingestionDestinationArn, appBundleIdentifier, ingestionIdentifier)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, string(out.Status), nil
	}
}

func findIngestionDestinationByID(ctx context.Context, conn *appfabric.Client, ingestionDestinationArn, appBundleIdentifier, ingestionIdentifier string) (*awstypes.IngestionDestination, error) {
	in := &appfabric.GetIngestionDestinationInput{
		IngestionDestinationIdentifier: aws.String(ingestionDestinationArn),
		AppBundleIdentifier:            aws.String(appBundleIdentifier),
		IngestionIdentifier:            aws.String(ingestionIdentifier),
	}

	out, err := conn.GetIngestionDestination(ctx, in)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return nil, err
	}

	if out == nil || out.IngestionDestination == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out.IngestionDestination, nil
}

func expandDestinationConfiguration(ctx context.Context, tfList []destinationConfigurationModel) (awstypes.DestinationConfiguration, diag.Diagnostics) {
	auditLog := []awstypes.DestinationConfiguration{}
	var diags diag.Diagnostics

	tfObj := tfList[0]

	if !tfObj.DestinationConfigurationAuditLog.IsNull() {
		var destinationConfigurationAuditLog []destinationConfigurationAuditLogModel
		diags.Append(tfObj.DestinationConfigurationAuditLog.ElementsAs(ctx, &destinationConfigurationAuditLog, false)...)
		apiObject := expandDestinationConfigurationAuditLog(ctx, destinationConfigurationAuditLog)
		auditLog = append(auditLog, apiObject)
	}
	return auditLog[0], diags
}

func expandDestinationConfigurationAuditLog(ctx context.Context, tfList []destinationConfigurationAuditLogModel) *awstypes.DestinationConfigurationMemberAuditLog {
	var diags diag.Diagnostics

	if len(tfList) == 0 {
		return nil
	}

	tfObj := tfList[0]

	var destinationData []destinationModel
	diags.Append(tfObj.Destination.ElementsAs(ctx, &destinationData, false)...)

	destination := &awstypes.DestinationConfigurationMemberAuditLog{
		Value: awstypes.AuditLogDestinationConfiguration{
			Destination: expandDestination(ctx, destinationData),
		},
	}

	return destination
}

func expandDestination(ctx context.Context, tfList []destinationModel) awstypes.Destination {
	destination := []awstypes.Destination{}
	var diags diag.Diagnostics

	if len(tfList) == 0 {
		return nil
	}

	tfObj := tfList[0]

	for _, item := range tfList {
		if !item.FirehoseStream.IsNull() && (len(item.FirehoseStream.Elements()) > 0) {
			var firehoseStream []firehoseStreamModel
			diags.Append(tfObj.FirehoseStream.ElementsAs(ctx, &firehoseStream, false)...)
			streamName := expandFirehoseStream(ctx, firehoseStream)
			destination = append(destination, streamName)
		}
		if !item.S3Bucket.IsNull() && (len(item.S3Bucket.Elements()) > 0) {
			var s3Bucket []s3BucketModel
			diags.Append(tfObj.S3Bucket.ElementsAs(ctx, &s3Bucket, false)...)
			bucketName := expandS3Bucket(ctx, s3Bucket)
			destination = append(destination, bucketName)
		}
	}

	return destination[0]
}

func expandFirehoseStream(ctx context.Context, tfList []firehoseStreamModel) *awstypes.DestinationMemberFirehoseStream {
	if len(tfList) == 0 {
		return nil
	}

	return &awstypes.DestinationMemberFirehoseStream{
		Value: awstypes.FirehoseStream{
			StreamName: fwflex.StringFromFramework(ctx, tfList[0].StreamName),
		},
	}
}

func expandS3Bucket(ctx context.Context, tfList []s3BucketModel) *awstypes.DestinationMemberS3Bucket {
	if len(tfList) == 0 {
		return nil
	}

	return &awstypes.DestinationMemberS3Bucket{
		Value: awstypes.S3Bucket{
			BucketName: fwflex.StringFromFramework(ctx, tfList[0].BucketName),
			Prefix:     fwflex.StringFromFramework(ctx, tfList[0].Prefix),
		},
	}
}

func expandProcessingConfiguration(ctx context.Context, tfList []processingConfigurationModel) (awstypes.ProcessingConfiguration, diag.Diagnostics) {
	auditLog := []awstypes.ProcessingConfiguration{}
	var diags diag.Diagnostics

	tfObj := tfList[0]

	if !tfObj.ProcessingConfigurationAuditLog.IsNull() {
		var processingConfigurationAuditLog []processingConfigurationAuditLogModel
		diags.Append(tfObj.ProcessingConfigurationAuditLog.ElementsAs(ctx, &processingConfigurationAuditLog, false)...)
		apiObject := expandProcessingConfigurationAuditLog(processingConfigurationAuditLog)
		auditLog = append(auditLog, apiObject)
	}
	return auditLog[0], diags
}

func expandProcessingConfigurationAuditLog(auditLog []processingConfigurationAuditLogModel) *awstypes.ProcessingConfigurationMemberAuditLog {
	if len(auditLog) == 0 {
		return nil
	}

	return &awstypes.ProcessingConfigurationMemberAuditLog{
		Value: awstypes.AuditLogProcessingConfiguration{
			Format: auditLog[0].Format.ValueEnum(),
			Schema: auditLog[0].Schema.ValueEnum(),
		},
	}
}

func flattenDestinationConfiguration(ctx context.Context, apiObject awstypes.DestinationConfiguration) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	elemType := types.ObjectType{AttrTypes: destinationConfigurationModelAttrTypes}

	if apiObject == nil {
		return types.ListNull(elemType), diags
	}

	obj := map[string]attr.Value{}

	switch v := apiObject.(type) {
	case *awstypes.DestinationConfigurationMemberAuditLog:
		auditLog, d := flattenDestinationConfigurationAuditLog(ctx, &v.Value)
		diags.Append(d...)
		obj = map[string]attr.Value{
			"audit_log": auditLog,
		}
	default:
		log.Println("union is nil or unknown type")
	}

	objVal, d := types.ObjectValue(destinationConfigurationModelAttrTypes, obj)
	diags.Append(d...)

	listVal, d := types.ListValue(elemType, []attr.Value{objVal})
	diags.Append(d...)

	return listVal, diags
}

func flattenDestinationConfigurationAuditLog(ctx context.Context, apiObject *awstypes.AuditLogDestinationConfiguration) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	elemType := types.ObjectType{AttrTypes: destinationConfigurationMemberAuditLogModelAttrTypes}

	if apiObject == nil {
		return types.ListNull(elemType), diags
	}

	obj := map[string]attr.Value{}

	destination, d := flattenDestination(ctx, apiObject.Destination)
	diags.Append(d...)
	obj["destination"] = destination

	objVal, d := types.ObjectValue(destinationConfigurationMemberAuditLogModelAttrTypes, obj)
	diags.Append(d...)

	listVal, d := types.ListValue(elemType, []attr.Value{objVal})
	diags.Append(d...)

	return listVal, diags
}

func flattenDestination(ctx context.Context, apiObject awstypes.Destination) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	elemType := types.ObjectType{AttrTypes: destinationModelAttrTypes}

	obj := map[string]attr.Value{}

	switch v := apiObject.(type) {
	case *awstypes.DestinationMemberFirehoseStream:
		destination, d := flattenDestinationModel(ctx, &v.Value, nil, "firehose")
		diags.Append(d...)
		obj = map[string]attr.Value{
			"firehose_stream": destination,
			"s3_bucket":       types.ListNull(s3BucketAttrTypes),
		}
	case *awstypes.DestinationMemberS3Bucket:
		destination, d := flattenDestinationModel(ctx, nil, &v.Value, "s3")
		diags.Append(d...)
		obj = map[string]attr.Value{
			"firehose_stream": types.ListNull(firehoseStreamAttrTypes),
			"s3_bucket":       destination,
		}
	}

	objVal, d := types.ObjectValue(destinationModelAttrTypes, obj)
	diags.Append(d...)

	listVal, d := types.ListValue(elemType, []attr.Value{objVal})
	diags.Append(d...)

	return listVal, diags
}

func flattenDestinationModel(ctx context.Context, firehoseApiObject *awstypes.FirehoseStream, s3ApiObject *awstypes.S3Bucket, destinationType string) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	var elemType types.ObjectType
	var obj map[string]attr.Value
	var objVal basetypes.ObjectValue

	if destinationType == "firehose" {
		var d diag.Diagnostics
		elemType = types.ObjectType{AttrTypes: firehoseStreamModelAttrTypes}
		obj = map[string]attr.Value{
			"stream_name": fwflex.StringToFramework(ctx, firehoseApiObject.StreamName),
		}
		objVal, d = types.ObjectValue(firehoseStreamModelAttrTypes, obj)
		diags.Append(d...)
	} else if destinationType == "s3" {
		var d diag.Diagnostics
		elemType = types.ObjectType{AttrTypes: s3BucketModelAttrTypes}
		obj = map[string]attr.Value{
			"bucket_name": fwflex.StringToFramework(ctx, s3ApiObject.BucketName),
			"prefix":      fwflex.StringToFramework(ctx, s3ApiObject.Prefix),
		}
		objVal, d = types.ObjectValue(s3BucketModelAttrTypes, obj)
		diags.Append(d...)
	}

	listVal, d := types.ListValue(elemType, []attr.Value{objVal})
	diags.Append(d...)

	return listVal, diags
}

var (
	destinationConfigurationModelAttrTypes = map[string]attr.Type{
		"audit_log": types.ListType{ElemType: types.ObjectType{AttrTypes: destinationConfigurationMemberAuditLogModelAttrTypes}},
	}

	destinationConfigurationMemberAuditLogModelAttrTypes = map[string]attr.Type{
		"destination": types.ListType{ElemType: types.ObjectType{AttrTypes: destinationModelAttrTypes}},
	}

	destinationModelAttrTypes = map[string]attr.Type{
		"firehose_stream": types.ListType{ElemType: firehoseStreamAttrTypes},
		"s3_bucket":       types.ListType{ElemType: s3BucketAttrTypes},
	}

	firehoseStreamAttrTypes = types.ObjectType{AttrTypes: firehoseStreamModelAttrTypes}
	s3BucketAttrTypes       = types.ObjectType{AttrTypes: s3BucketModelAttrTypes}

	firehoseStreamModelAttrTypes = map[string]attr.Type{
		"stream_name": types.StringType,
	}

	s3BucketModelAttrTypes = map[string]attr.Type{
		"bucket_name": types.StringType,
		"prefix":      types.StringType,
	}

	processingConfigurationModelAttrTypes = map[string]attr.Type{
		"audit_log": types.ListType{ElemType: types.ObjectType{AttrTypes: processingConfigurationMemberAuditLogModelAttrTypes}},
	}

	processingConfigurationMemberAuditLogModelAttrTypes = map[string]attr.Type{
		"format": types.StringType,
		"schema": types.StringType,
	}
)

const (
	ingestionDestinationResourceIDPartCount = 3
)

func (m *resourceIngestionDestinationModel) InitFromID() error {
	parts, err := flex.ExpandResourceId(m.ID.ValueString(), ingestionDestinationResourceIDPartCount, false)
	if err != nil {
		return err
	}

	m.IngestionDestinationArn = types.StringValue(parts[0])
	m.AppBundleIdentifier = types.StringValue(parts[1])
	m.IngestionIdentifier = types.StringValue(parts[2])

	return nil
}

func (m *resourceIngestionDestinationModel) setID() {
	m.ID = types.StringValue(errs.Must(flex.FlattenResourceId([]string{m.IngestionDestinationArn.ValueString(), m.AppBundleIdentifier.ValueString(), m.IngestionIdentifier.ValueString()}, ingestionDestinationResourceIDPartCount, false)))
}

type resourceIngestionDestinationModel struct {
	AppBundleIdentifier      types.String                                                  `tfsdk:"app_bundle_identifier"`
	DestinationConfiguration types.List                                                    `tfsdk:"destination_configuration"`
	ID                       types.String                                                  `tfsdk:"id"`
	IngestionArn             types.String                                                  `tfsdk:"ingestion_arn"`
	IngestionDestinationArn  types.String                                                  `tfsdk:"arn"`
	IngestionIdentifier      types.String                                                  `tfsdk:"ingestion_identifier"`
	ProcessingConfiguration  fwtypes.ListNestedObjectValueOf[processingConfigurationModel] `tfsdk:"processing_configuration"`
	Timeouts                 timeouts.Value                                                `tfsdk:"timeouts"`
}

type destinationConfigurationModel struct {
	DestinationConfigurationAuditLog fwtypes.ListNestedObjectValueOf[destinationConfigurationAuditLogModel] `tfsdk:"audit_log"`
}

type destinationConfigurationAuditLogModel struct {
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
	ProcessingConfigurationAuditLog fwtypes.ListNestedObjectValueOf[processingConfigurationAuditLogModel] `tfsdk:"audit_log"`
}

type processingConfigurationAuditLogModel struct {
	Format fwtypes.StringEnum[awstypes.Format] `tfsdk:"format"`
	Schema fwtypes.StringEnum[awstypes.Schema] `tfsdk:"schema"`
}
