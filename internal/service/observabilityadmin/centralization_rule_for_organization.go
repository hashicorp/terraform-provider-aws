// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package observabilityadmin

import (
	"context"
	"errors"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/observabilityadmin"
	awstypes "github.com/aws/aws-sdk-go-v2/service/observabilityadmin/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdkretry "github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	fwvalidators "github.com/hashicorp/terraform-provider-aws/internal/framework/validators"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_observabilityadmin_centralization_rule_for_organization", name="Centralization Rule For Organization")
// @Tags(identifierAttribute="rule_arn")
func newCentralizationRuleForOrganizationResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &centralizationRuleForOrganizationResource{}

	r.SetDefaultCreateTimeout(5 * time.Minute)
	r.SetDefaultUpdateTimeout(5 * time.Minute)

	return r, nil
}

type centralizationRuleForOrganizationResource struct {
	framework.ResourceWithModel[centralizationRuleForOrganizationResourceModel]
	framework.WithTimeouts
}

func (r *centralizationRuleForOrganizationResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"rule_arn": framework.ARNAttributeComputedOnly(),
			"rule_name": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 100),
					stringvalidator.RegexMatches(regexache.MustCompile(`[0-9A-Za-z-_.#/]+`), ""),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			names.AttrRule: schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[centralizationRuleModel](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtLeast(1),
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						names.AttrDestination: schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[centralizationRuleDestinationModel](ctx),
							Validators: []validator.List{
								listvalidator.IsRequired(),
								listvalidator.SizeAtLeast(1),
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"account": schema.StringAttribute{
										Required: true,
										Validators: []validator.String{
											fwvalidators.AWSAccountID(),
										},
									},
									names.AttrRegion: schema.StringAttribute{
										Required: true,
										Validators: []validator.String{
											fwvalidators.AWSRegion(),
										},
									},
								},
								Blocks: map[string]schema.Block{
									"destination_logs_configuration": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[destinationLogsConfigurationModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Blocks: map[string]schema.Block{
												"backup_configuration": schema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[logsBackupConfigurationModel](ctx),
													Validators: []validator.List{
														listvalidator.SizeAtMost(1),
													},
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															names.AttrKMSKeyARN: schema.StringAttribute{
																CustomType: fwtypes.ARNType,
																Optional:   true,
															},
															names.AttrRegion: schema.StringAttribute{
																Optional: true,
																Validators: []validator.String{
																	fwvalidators.AWSRegion(),
																},
															},
														},
													},
												},
												"logs_encryption_configuration": schema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[logsEncryptionConfigurationModel](ctx),
													Validators: []validator.List{
														listvalidator.SizeAtMost(1),
													},
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															"encryption_conflict_resolution_strategy": schema.StringAttribute{
																Optional:   true,
																CustomType: fwtypes.StringEnumType[awstypes.EncryptionConflictResolutionStrategy](),
															},
															"encryption_strategy": schema.StringAttribute{
																Required:   true,
																CustomType: fwtypes.StringEnumType[awstypes.EncryptionStrategy](),
															},
															names.AttrKMSKeyARN: schema.StringAttribute{
																CustomType: fwtypes.ARNType,
																Optional:   true,
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
						names.AttrSource: schema.ListNestedBlock{
							Validators: []validator.List{
								listvalidator.IsRequired(),
								listvalidator.SizeAtMost(1),
								listvalidator.SizeAtLeast(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"regions": schema.SetAttribute{
										CustomType:  fwtypes.SetOfStringType,
										ElementType: types.StringType,
										Validators: []validator.Set{
											setvalidator.SizeAtLeast(1),
											setvalidator.ValueStringsAre(fwvalidators.AWSRegion()),
										},
										Required: true,
									},
									names.AttrScope: schema.StringAttribute{
										Required: true,
										Validators: []validator.String{
											stringvalidator.LengthAtLeast(1),
											stringvalidator.LengthAtMost(2000),
										},
									},
								},
								Blocks: map[string]schema.Block{
									"source_logs_configuration": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[sourceLogsConfigurationModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"encrypted_log_group_strategy": schema.StringAttribute{
													CustomType: fwtypes.StringEnumType[awstypes.EncryptedLogGroupStrategy](),
													Required:   true,
												},
												"log_group_selection_criteria": schema.StringAttribute{
													Required: true,
													Validators: []validator.String{
														stringvalidator.LengthAtLeast(1),
														stringvalidator.LengthAtMost(2000),
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
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
			}),
		},
	}
}

func (r *centralizationRuleForOrganizationResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data centralizationRuleForOrganizationResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ObservabilityAdminClient(ctx)

	ruleName := fwflex.StringValueFromFramework(ctx, data.RuleName)
	var input observabilityadmin.CreateCentralizationRuleForOrganizationInput
	response.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	input.Tags = getTagsIn(ctx)

	output, err := conn.CreateCentralizationRuleForOrganization(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, ruleName)
		return
	}

	// Set values for unknowns.
	data.RuleARN = fwflex.StringToFramework(ctx, output.RuleArn)

	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Flatten(ctx, output, &data))
	if response.Diagnostics.HasError() {
		return
	}

	if _, err := waitCentralizationRuleForOrganizationHealthy(ctx, conn, ruleName, r.CreateTimeout(ctx, data.Timeouts)); err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, ruleName)
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, data))
}

func (r *centralizationRuleForOrganizationResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data centralizationRuleForOrganizationResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ObservabilityAdminClient(ctx)

	ruleName := fwflex.StringValueFromFramework(ctx, data.RuleName)
	out, err := findCentralizationRuleForOrganizationByID(ctx, conn, ruleName)
	if retry.NotFound(err) {
		smerr.AddOne(ctx, &response.Diagnostics, fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, ruleName)
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Flatten(ctx, out, &data, fwflex.WithFieldNamePrefix("Centralization")))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &data))
}

func (r *centralizationRuleForOrganizationResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var new, old centralizationRuleForOrganizationResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &new))
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &old))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ObservabilityAdminClient(ctx)

	diff, d := fwflex.Diff(ctx, new, old)
	smerr.AddEnrich(ctx, &response.Diagnostics, d)
	if response.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		ruleName := fwflex.StringValueFromFramework(ctx, new.RuleName)
		var input observabilityadmin.UpdateCentralizationRuleForOrganizationInput
		smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Expand(ctx, new, &input))
		if response.Diagnostics.HasError() {
			return
		}

		// Additional fields.
		input.RuleIdentifier = aws.String(ruleName)

		_, err := conn.UpdateCentralizationRuleForOrganization(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, ruleName)
			return
		}

		if _, err := waitCentralizationRuleForOrganizationHealthy(ctx, conn, ruleName, r.UpdateTimeout(ctx, new.Timeouts)); err != nil {
			smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, ruleName)
			return
		}
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &new))
}

func (r *centralizationRuleForOrganizationResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data centralizationRuleForOrganizationResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ObservabilityAdminClient(ctx)

	ruleName := fwflex.StringValueFromFramework(ctx, data.RuleName)
	input := observabilityadmin.DeleteCentralizationRuleForOrganizationInput{
		RuleIdentifier: aws.String(ruleName),
	}

	_, err := conn.DeleteCentralizationRuleForOrganization(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, ruleName)
		return
	}
}

func (r *centralizationRuleForOrganizationResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("rule_name"), request, response)
}

func findCentralizationRuleForOrganizationByID(ctx context.Context, conn *observabilityadmin.Client, id string) (*observabilityadmin.GetCentralizationRuleForOrganizationOutput, error) {
	input := observabilityadmin.GetCentralizationRuleForOrganizationInput{
		RuleIdentifier: aws.String(id),
	}

	return findCentralizationRuleForOrganization(ctx, conn, &input)
}

func findCentralizationRuleForOrganization(ctx context.Context, conn *observabilityadmin.Client, input *observabilityadmin.GetCentralizationRuleForOrganizationInput) (*observabilityadmin.GetCentralizationRuleForOrganizationOutput, error) {
	output, err := conn.GetCentralizationRuleForOrganization(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &sdkretry.NotFoundError{
			LastError:   err,
			LastRequest: &input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output, nil
}

func statusCentralizationRuleForOrganization(conn *observabilityadmin.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findCentralizationRuleForOrganizationByID(ctx, conn, id)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", smarterr.NewError(err)
		}

		return output, string(output.RuleHealth), nil
	}
}

func waitCentralizationRuleForOrganizationHealthy(ctx context.Context, conn *observabilityadmin.Client, id string, timeout time.Duration) (*observabilityadmin.GetCentralizationRuleForOrganizationOutput, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.RuleHealthProvisioning),
		Target:                    enum.Slice(awstypes.RuleHealthHealthy),
		Refresh:                   statusCentralizationRuleForOrganization(conn, id),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*observabilityadmin.GetCentralizationRuleForOrganizationOutput); ok {
		retry.SetLastError(err, errors.New(string(out.FailureReason)))
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

type centralizationRuleForOrganizationResourceModel struct {
	framework.WithRegionModel
	Rule     fwtypes.ListNestedObjectValueOf[centralizationRuleModel] `tfsdk:"rule"`
	RuleARN  types.String                                             `tfsdk:"rule_arn"`
	RuleName types.String                                             `tfsdk:"rule_name"`
	Tags     tftags.Map                                               `tfsdk:"tags"`
	TagsAll  tftags.Map                                               `tfsdk:"tags_all"`
	Timeouts timeouts.Value                                           `tfsdk:"timeouts"`
}

type centralizationRuleModel struct {
	Destination fwtypes.ListNestedObjectValueOf[centralizationRuleDestinationModel] `tfsdk:"destination"`
	Source      fwtypes.ListNestedObjectValueOf[centralizationRuleSourceModel]      `tfsdk:"source"`
}

type centralizationRuleDestinationModel struct {
	Account                      types.String                                                       `tfsdk:"account"`
	DestinationLogsConfiguration fwtypes.ListNestedObjectValueOf[destinationLogsConfigurationModel] `tfsdk:"destination_logs_configuration"`
	Region                       types.String                                                       `tfsdk:"region"`
}

type centralizationRuleSourceModel struct {
	Regions                 fwtypes.SetOfString                                           `tfsdk:"regions"`
	Scope                   types.String                                                  `tfsdk:"scope"`
	SourceLogsConfiguration fwtypes.ListNestedObjectValueOf[sourceLogsConfigurationModel] `tfsdk:"source_logs_configuration"`
}

type destinationLogsConfigurationModel struct {
	BackupConfiguration         fwtypes.ListNestedObjectValueOf[logsBackupConfigurationModel]     `tfsdk:"backup_configuration"`
	LogsEncryptionConfiguration fwtypes.ListNestedObjectValueOf[logsEncryptionConfigurationModel] `tfsdk:"logs_encryption_configuration"`
}

type sourceLogsConfigurationModel struct {
	EncryptedLogGroupStrategy fwtypes.StringEnum[awstypes.EncryptedLogGroupStrategy] `tfsdk:"encrypted_log_group_strategy"`
	LogGroupSelectionCriteria types.String                                           `tfsdk:"log_group_selection_criteria"`
}

type logsBackupConfigurationModel struct {
	KMSKeyARN fwtypes.ARN  `tfsdk:"kms_key_arn"`
	Region    types.String `tfsdk:"region"`
}

type logsEncryptionConfigurationModel struct {
	EncryptionConflictResolutionStrategy fwtypes.StringEnum[awstypes.EncryptionConflictResolutionStrategy] `tfsdk:"encryption_conflict_resolution_strategy"`
	EncryptionStrategy                   fwtypes.StringEnum[awstypes.EncryptionStrategy]                   `tfsdk:"encryption_strategy"`
	KMSKeyARN                            fwtypes.ARN                                                       `tfsdk:"kms_key_arn"`
}
