// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package observabilityadmin

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/observabilityadmin"
	awstypes "github.com/aws/aws-sdk-go-v2/service/observabilityadmin/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdkretry "github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	sweepfw "github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_observabilityadmin_centralization_rule_for_organization", name="Centralization Rule For Organization")
// @Tags(identifierAttribute="rule_arn")
func newResourceCentralizationRuleForOrganization(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceCentralizationRuleForOrganization{}

	r.SetDefaultCreateTimeout(5 * time.Minute)
	r.SetDefaultUpdateTimeout(5 * time.Minute)
	r.SetDefaultDeleteTimeout(5 * time.Minute)

	return r, nil
}

const (
	ResNameCentralizationRuleForOrganization = "Centralization Rule For Organization"
)

type resourceCentralizationRuleForOrganization struct {
	framework.ResourceWithModel[centralizationRuleForOrganizationModel]
	framework.WithTimeouts

	flexOpt fwflex.AutoFlexOptionsFunc
}

func (r *resourceCentralizationRuleForOrganization) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"rule_arn": framework.ARNAttributeComputedOnly(),
			// names.AttrID: framework.IDAttribute(),
			"rule_name": schema.StringAttribute{
				Required: true,
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
										Optional: true,
									},
									names.AttrRegion: schema.StringAttribute{
										Required: true,
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
															names.AttrRegion: schema.StringAttribute{
																Optional: true,
															},
															names.AttrKMSKeyARN: schema.StringAttribute{
																CustomType: fwtypes.ARNType,
																Optional:   true,
															},
														},
													},
												},
												"logs_encryption_configuration": schema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[logsEncryptionConfigurationModel](ctx),
													Validators: []validator.List{
														listvalidator.SizeAtMost(1),
													},
													PlanModifiers: []planmodifier.List{
														listplanmodifier.RequiresReplace(),
													},
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															"encryption_strategy": schema.StringAttribute{
																Required:   true,
																CustomType: fwtypes.StringEnumType[awstypes.EncryptionStrategy](),
																Validators: []validator.String{
																	stringvalidator.OneOf(
																		string(awstypes.EncryptionStrategyAwsOwned),
																		string(awstypes.EncryptionStrategyCustomerManaged),
																	),
																},
															},
															"encryption_conflict_resolution_strategy": schema.StringAttribute{
																Optional:   true,
																CustomType: fwtypes.StringEnumType[awstypes.EncryptionConflictResolutionStrategy](),
																Validators: []validator.String{
																	stringvalidator.OneOf(
																		string(awstypes.EncryptionConflictResolutionStrategyAllow),
																		string(awstypes.EncryptionConflictResolutionStrategySkip),
																	),
																},
															},
															"kms_key_arn": schema.StringAttribute{
																CustomType: fwtypes.ARNType,
																Optional:   true,
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
									"regions": schema.ListAttribute{
										CustomType:  fwtypes.ListOfStringType,
										ElementType: types.StringType,
										Validators: []validator.List{
											listvalidator.SizeAtLeast(1),
										},
										Required: true,
									},
									names.AttrScope: schema.StringAttribute{
										Optional: true,
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
													Validators: []validator.String{
														stringvalidator.OneOf(
															string(awstypes.EncryptedLogGroupStrategyAllow),
															string(awstypes.EncryptedLogGroupStrategySkip),
														),
													},
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
				Delete: true,
			}),
		},
	}
}

func (r *resourceCentralizationRuleForOrganization) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().ObservabilityAdminClient(ctx)

	var plan centralizationRuleForOrganizationModel
	smerr.EnrichAppend(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	var input observabilityadmin.CreateCentralizationRuleForOrganizationInput
	smerr.EnrichAppend(ctx, &resp.Diagnostics, flex.Expand(ctx, plan, &input))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := conn.CreateCentralizationRuleForOrganization(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.RuleName.String())
		return
	}
	if out == nil || out.RuleArn == nil {
		smerr.AddError(ctx, &resp.Diagnostics, errors.New("empty output"), smerr.ID, plan.RuleName.String())
		return
	}

	// Set the ID to the rule name for identification
	plan.ARN = flex.StringToFramework(ctx, out.RuleArn)

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	_, err = waitCentralizationRuleForOrganizationCreated(ctx, conn, plan.RuleName.ValueString(), createTimeout)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.RuleName.String())
		return
	}

	smerr.EnrichAppend(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *resourceCentralizationRuleForOrganization) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().ObservabilityAdminClient(ctx)

	var state centralizationRuleForOrganizationModel
	smerr.EnrichAppend(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findCentralizationRuleForOrganizationByID(ctx, conn, state.RuleName.ValueString())
	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.RuleName.String())
		return
	}

	smerr.EnrichAppend(ctx, &resp.Diagnostics, flex.Flatten(ctx, out, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.EnrichAppend(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *resourceCentralizationRuleForOrganization) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().ObservabilityAdminClient(ctx)

	var plan, state centralizationRuleForOrganizationModel
	smerr.EnrichAppend(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	smerr.EnrichAppend(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	diff, d := flex.Diff(ctx, plan, state)
	smerr.EnrichAppend(ctx, &resp.Diagnostics, d)
	if resp.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		var input observabilityadmin.UpdateCentralizationRuleForOrganizationInput
		smerr.EnrichAppend(ctx, &resp.Diagnostics, flex.Expand(ctx, plan, &input))
		if resp.Diagnostics.HasError() {
			return
		}

		out, err := conn.UpdateCentralizationRuleForOrganization(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.RuleName.String())
			return
		}
		if out == nil || out.RuleArn == nil {
			smerr.AddError(ctx, &resp.Diagnostics, errors.New("empty output"), smerr.ID, plan.RuleName.String())
			return
		}

		plan.ARN = flex.StringToFramework(ctx, out.RuleArn)
	}

	updateTimeout := r.UpdateTimeout(ctx, plan.Timeouts)
	_, err := waitCentralizationRuleForOrganizationUpdated(ctx, conn, plan.RuleName.ValueString(), updateTimeout)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.RuleName.String())
		return
	}

	smerr.EnrichAppend(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *resourceCentralizationRuleForOrganization) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().ObservabilityAdminClient(ctx)

	var state centralizationRuleForOrganizationModel
	smerr.EnrichAppend(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	input := observabilityadmin.DeleteCentralizationRuleForOrganizationInput{
		RuleIdentifier: state.RuleName.ValueStringPointer(),
	}

	_, err := conn.DeleteCentralizationRuleForOrganization(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}

		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.RuleName)
		return
	}

	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	_, err = waitCentralizationRuleForOrganizationDeleted(ctx, conn, state.RuleName.String(), deleteTimeout)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.RuleName)
		return
	}
}

func (r *resourceCentralizationRuleForOrganization) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrID), req, resp)
}

const (
	statusChangePending = "Pending"
	statusDeleting      = "Deleting"
	statusNormal        = "Normal"
	statusUpdated       = "Updated"
)

func waitCentralizationRuleForOrganizationCreated(ctx context.Context, conn *observabilityadmin.Client, id string, timeout time.Duration) (*observabilityadmin.GetCentralizationRuleForOrganizationOutput, error) {
	stateConf := &sdkretry.StateChangeConf{
		Pending:                   []string{},
		Target:                    []string{statusNormal},
		Refresh:                   statusCentralizationRuleForOrganization(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*observabilityadmin.GetCentralizationRuleForOrganizationOutput); ok {
		return out, err
	}

	return nil, err
}

func waitCentralizationRuleForOrganizationUpdated(ctx context.Context, conn *observabilityadmin.Client, id string, timeout time.Duration) (*observabilityadmin.GetCentralizationRuleForOrganizationOutput, error) {
	stateConf := &sdkretry.StateChangeConf{
		Pending:                   []string{statusChangePending},
		Target:                    []string{statusUpdated},
		Refresh:                   statusCentralizationRuleForOrganization(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*observabilityadmin.GetCentralizationRuleForOrganizationOutput); ok {
		return out, err
	}

	return nil, err
}

func waitCentralizationRuleForOrganizationDeleted(ctx context.Context, conn *observabilityadmin.Client, id string, timeout time.Duration) (*observabilityadmin.GetCentralizationRuleForOrganizationOutput, error) {
	stateConf := &sdkretry.StateChangeConf{
		Pending: []string{statusDeleting, statusNormal},
		Target:  []string{},
		Refresh: statusCentralizationRuleForOrganization(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*observabilityadmin.GetCentralizationRuleForOrganizationOutput); ok {
		return out, err
	}

	return nil, err
}

func statusCentralizationRuleForOrganization(ctx context.Context, conn *observabilityadmin.Client, id string) sdkretry.StateRefreshFunc {
	return func() (any, string, error) {
		out, err := findCentralizationRuleForOrganizationByID(ctx, conn, id)
		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, string(out.RuleHealth), nil
	}
}

func findCentralizationRuleForOrganizationByID(ctx context.Context, conn *observabilityadmin.Client, id string) (*observabilityadmin.GetCentralizationRuleForOrganizationOutput, error) {
	input := observabilityadmin.GetCentralizationRuleForOrganizationInput{
		RuleIdentifier: aws.String(id),
	}

	out, err := conn.GetCentralizationRuleForOrganization(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &sdkretry.NotFoundError{
				LastError:   err,
				LastRequest: &input,
			}
		}

		return nil, err
	}

	if out == nil {
		return nil, &sdkretry.NotFoundError{
			LastError:   errors.New("empty result"),
			LastRequest: &input,
		}
	}

	return out, nil
}

type centralizationRuleForOrganizationModel struct {
	framework.WithRegionModel
	ARN types.String `tfsdk:"rule_arn"`
	// ID       types.String   `tfsdk:"id"`
	Rule     fwtypes.ListNestedObjectValueOf[centralizationRuleModel] `tfsdk:"rule"`
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
	Regions                 fwtypes.ListOfString                                          `tfsdk:"regions"`
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
	Region    types.String `tfsdk:"region"`
	KMSKeyARN fwtypes.ARN  `tfsdk:"kms_key_arn"`
}

type logsEncryptionConfigurationModel struct {
	EncryptionStrategy                   fwtypes.StringEnum[awstypes.EncryptionStrategy]                   `tfsdk:"encryption_strategy"`
	EncryptionConflictResolutionStrategy fwtypes.StringEnum[awstypes.EncryptionConflictResolutionStrategy] `tfsdk:"encryption_conflict_resolution_strategy"`
	KMSKeyARN                            fwtypes.ARN                                                       `tfsdk:"kms_key_arn"`
}

func sweepCentralizationRuleForOrganizations(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	input := observabilityadmin.ListCentralizationRulesForOrganizationInput{}
	conn := client.ObservabilityAdminClient(ctx)
	var sweepResources []sweep.Sweepable

	pages := observabilityadmin.NewListCentralizationRulesForOrganizationPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, v := range page.CentralizationRuleSummaries {
			sweepResources = append(sweepResources, sweepfw.NewSweepResource(newResourceCentralizationRuleForOrganization, client,
				sweepfw.NewAttribute(names.AttrID, aws.ToString(v.RuleName))),
			)
		}
	}

	return sweepResources, nil
}
