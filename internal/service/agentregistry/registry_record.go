// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package agentregistry

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/agentregistrycontrol"
	awstypes "github.com/aws/aws-sdk-go-v2/service/agentregistrycontrol/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tfobjectvalidator "github.com/hashicorp/terraform-provider-aws/internal/framework/validators/objectvalidator"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_agentregistry_registry_record", name="Registry Record")
// @Tags(identifierAttribute="record_arn")
// @IdentityAttribute("registry_id")
// @IdentityAttribute("record_id")
// @ImportIDHandler("registryRecordImportID")
// @Testing(hasNoPreExistingResource=true)
// @Testing(importStateIdFunc=testAccRegistryRecordImportStateIDFunc)
// @Testing(importStateIdAttribute="record_id")
func newRegistryRecordResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &registryRecordResource{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

type registryRecordResource struct {
	framework.ResourceWithModel[registryRecordResourceModel]
	framework.WithTimeouts
	framework.WithImportByIdentity
}

func (r *registryRecordResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
			},
			names.AttrDisplayName: schema.StringAttribute{
				Optional: true,
			},
			names.AttrName: schema.StringAttribute{
				Required: true,
			},
			"record_arn": framework.ARNAttributeComputedOnly(),
			"record_id":  framework.IDAttribute(),
			"record_type": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.RecordType](),
				Required:   true,
			},
			"record_version": schema.StringAttribute{
				Optional: true,
			},
			"registry_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			"descriptors": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[descriptorsModel](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtLeast(1),
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Validators: []validator.Object{
						tfobjectvalidator.ExactlyOneOfChildren(
							path.MatchRelative().AtName("a2a_agent_card"),
							path.MatchRelative().AtName("agent_skills_definition"),
							path.MatchRelative().AtName("custom"),
							path.MatchRelative().AtName("mcp_server"),
						),
					},
					Blocks: map[string]schema.Block{
						"a2a_agent_card": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[a2aAgentCardDescriptorModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: descriptorPayloadAttributes(),
								Blocks: map[string]schema.Block{
									names.AttrSource: descriptorSourceBlock(ctx),
								},
							},
						},
						"agent_skills_definition": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[agentSkillsDefinitionDescriptorModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: descriptorPayloadAttributes(),
								Blocks: map[string]schema.Block{
									"additional_data": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[agentSkillsAdditionalDataModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Blocks: map[string]schema.Block{
												"skill_md": schema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[agentSkillsMDDescriptorModel](ctx),
													Validators: []validator.List{
														listvalidator.SizeAtMost(1),
													},
													NestedObject: schema.NestedBlockObject{
														Attributes: descriptorPayloadAttributes(),
														Blocks: map[string]schema.Block{
															names.AttrSource: descriptorSourceBlock(ctx),
														},
													},
												},
											},
										},
									},
								},
							},
						},
						"custom": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[customDescriptorModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"data": schema.StringAttribute{
										Optional: true,
									},
								},
							},
						},
						"mcp_server": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[mcpServerDescriptorModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: descriptorPayloadAttributes(),
								Blocks: map[string]schema.Block{
									"additional_data": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[mcpServerAdditionalDataModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Blocks: map[string]schema.Block{
												"tools": schema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[mcpToolsDescriptorModel](ctx),
													Validators: []validator.List{
														listvalidator.SizeAtMost(1),
													},
													NestedObject: schema.NestedBlockObject{
														Attributes: descriptorPayloadAttributes(),
													},
												},
											},
										},
									},
									names.AttrSource: descriptorSourceBlock(ctx),
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

// descriptorPayloadAttributes returns the data attributes shared by all typed descriptor payloads.
func descriptorPayloadAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"data": schema.StringAttribute{
			Optional: true,
		},
		"data_schema_version": schema.StringAttribute{
			Optional: true,
		},
	}
}

// descriptorSourceBlock returns the source block used to synchronize descriptor content from a URL.
func descriptorSourceBlock(ctx context.Context) schema.Block {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[descriptorSourceModel](ctx),
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		NestedObject: schema.NestedBlockObject{
			Blocks: map[string]schema.Block{
				"from_url": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[descriptorSourceFromURLModel](ctx),
					Validators: []validator.List{
						listvalidator.IsRequired(),
						listvalidator.SizeAtLeast(1),
						listvalidator.SizeAtMost(1),
					},
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							names.AttrURL: schema.StringAttribute{
								Required: true,
							},
						},
						Blocks: map[string]schema.Block{
							"credential_provider_configuration": schema.ListNestedBlock{
								CustomType: fwtypes.NewListNestedObjectTypeOf[credentialProviderConfigurationModel](ctx),
								NestedObject: schema.NestedBlockObject{
									Attributes: map[string]schema.Attribute{
										"credential_provider_type": schema.StringAttribute{
											CustomType: fwtypes.StringEnumType[awstypes.RegistryRecordCredentialProviderType](),
											Required:   true,
										},
									},
									Blocks: map[string]schema.Block{
										"credential_provider": schema.ListNestedBlock{
											CustomType: fwtypes.NewListNestedObjectTypeOf[credentialProviderModel](ctx),
											Validators: []validator.List{
												listvalidator.IsRequired(),
												listvalidator.SizeAtLeast(1),
												listvalidator.SizeAtMost(1),
											},
											NestedObject: schema.NestedBlockObject{
												Validators: []validator.Object{
													tfobjectvalidator.ExactlyOneOfChildren(
														path.MatchRelative().AtName("iam_credential_provider"),
														path.MatchRelative().AtName("oauth_credential_provider"),
													),
												},
												Blocks: map[string]schema.Block{
													"iam_credential_provider": schema.ListNestedBlock{
														CustomType: fwtypes.NewListNestedObjectTypeOf[iamCredentialProviderModel](ctx),
														Validators: []validator.List{
															listvalidator.SizeAtMost(1),
														},
														NestedObject: schema.NestedBlockObject{
															Attributes: map[string]schema.Attribute{
																names.AttrRegion: schema.StringAttribute{
																	Optional: true,
																},
																names.AttrRoleARN: schema.StringAttribute{
																	CustomType: fwtypes.ARNType,
																	Optional:   true,
																},
																"service": schema.StringAttribute{
																	Optional: true,
																},
															},
														},
													},
													"oauth_credential_provider": schema.ListNestedBlock{
														CustomType: fwtypes.NewListNestedObjectTypeOf[oauthCredentialProviderModel](ctx),
														Validators: []validator.List{
															listvalidator.SizeAtMost(1),
														},
														NestedObject: schema.NestedBlockObject{
															Attributes: map[string]schema.Attribute{
																"custom_parameters": schema.MapAttribute{
																	CustomType:  fwtypes.MapOfStringType,
																	ElementType: types.StringType,
																	Optional:    true,
																},
																"grant_type": schema.StringAttribute{
																	CustomType: fwtypes.StringEnumType[awstypes.RegistryRecordOAuthGrantType](),
																	Optional:   true,
																},
																"provider_arn": schema.StringAttribute{
																	CustomType: fwtypes.ARNType,
																	Required:   true,
																},
																"scopes": schema.ListAttribute{
																	CustomType:  fwtypes.ListOfStringType,
																	ElementType: types.StringType,
																	Optional:    true,
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
		},
	}
}

func (r *registryRecordResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan registryRecordResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().AgentRegistryClient(ctx)

	name := fwflex.StringValueFromFramework(ctx, plan.Name)
	var input agentregistrycontrol.CreateRegistryRecordInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &input))
	if resp.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	input.ClientToken = aws.String(create.UniqueId(ctx))
	input.Tags = getTagsIn(ctx)

	out, err := conn.CreateRegistryRecord(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.Name, name)
		return
	}

	// CreateRegistryRecord only returns the record ARN. GetRegistryRecord accepts
	// the ARN as the record identifier, so wait using it and read the record ID
	// from the API response.
	recordARN := aws.ToString(out.RecordArn)
	registryID := fwflex.StringValueFromFramework(ctx, plan.RegistryID)

	created, err := waitRegistryRecordCreated(ctx, conn, registryID, recordARN, r.CreateTimeout(ctx, plan.Timeouts))
	if err != nil {
		// Taint the resource.
		resp.State.SetAttribute(ctx, path.Root("registry_id"), registryID)
		resp.State.SetAttribute(ctx, path.Root("record_id"), recordARN)
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.Name, name)
		return
	}

	// Set values for unknowns.
	plan.RecordARN = fwflex.StringToFramework(ctx, created.RecordArn)
	plan.RecordID = fwflex.StringToFramework(ctx, created.RecordId)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *registryRecordResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state registryRecordResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().AgentRegistryClient(ctx)

	registryID := fwflex.StringValueFromFramework(ctx, state.RegistryID)
	recordID := fwflex.StringValueFromFramework(ctx, state.RecordID)
	out, err := findRegistryRecordByTwoPartKey(ctx, conn, registryID, recordID)
	if retry.NotFound(err) {
		smerr.AddOne(ctx, &resp.Diagnostics, fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, recordID)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, out, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *registryRecordResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state registryRecordResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().AgentRegistryClient(ctx)

	diff, d := fwflex.Diff(ctx, plan, state)
	smerr.AddEnrich(ctx, &resp.Diagnostics, d)
	if resp.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		recordID := fwflex.StringValueFromFramework(ctx, state.RecordID)
		// Description, DisplayName and Descriptors are wrapped in "Updated*"
		// optional-value types which AutoFlex cannot map; they are built by hand below.
		optFns := []fwflex.AutoFlexOptionsFunc{
			fwflex.WithIgnoredFieldNamesAppend("Description"),
			fwflex.WithIgnoredFieldNamesAppend("Descriptors"),
			fwflex.WithIgnoredFieldNamesAppend("DisplayName"),
		}
		var input agentregistrycontrol.UpdateRegistryRecordInput
		smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &input, optFns...))
		if resp.Diagnostics.HasError() {
			return
		}
		input.RecordId = aws.String(recordID)

		if !plan.Description.Equal(state.Description) {
			input.Description = &awstypes.UpdatedDescription{
				OptionalValue: fwflex.StringFromFramework(ctx, plan.Description),
			}
		}

		if !plan.DisplayName.Equal(state.DisplayName) {
			input.DisplayName = &awstypes.UpdatedDisplayName{
				OptionalValue: fwflex.StringFromFramework(ctx, plan.DisplayName),
			}
		}

		if !plan.Descriptors.Equal(state.Descriptors) {
			var descriptors awstypes.Descriptors
			smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan.Descriptors, &descriptors))
			if resp.Diagnostics.HasError() {
				return
			}
			input.Descriptors = expandUpdatedDescriptors(&descriptors)
		}

		_, err := conn.UpdateRegistryRecord(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, recordID)
			return
		}

		registryID := fwflex.StringValueFromFramework(ctx, state.RegistryID)
		if _, err := waitRegistryRecordUpdated(ctx, conn, registryID, recordID, r.UpdateTimeout(ctx, plan.Timeouts)); err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, recordID)
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *registryRecordResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state registryRecordResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().AgentRegistryClient(ctx)

	registryID := fwflex.StringValueFromFramework(ctx, state.RegistryID)
	recordID := fwflex.StringValueFromFramework(ctx, state.RecordID)
	input := agentregistrycontrol.DeleteRegistryRecordInput{
		RegistryId: aws.String(registryID),
		RecordId:   aws.String(recordID),
	}
	_, err := conn.DeleteRegistryRecord(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, recordID)
		return
	}

	if _, err := waitRegistryRecordDeleted(ctx, conn, registryID, recordID, r.DeleteTimeout(ctx, state.Timeouts)); err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, recordID)
		return
	}
}

// expandUpdatedDescriptors converts descriptors to the full-desired-state patch
// expected by UpdateRegistryRecord. All patch wrappers are always present so that
// any descriptor or field removed from configuration is unset in the record.
func expandUpdatedDescriptors(d *awstypes.Descriptors) *awstypes.UpdatedDescriptors {
	fields := &awstypes.UpdatedDescriptorsFields{
		A2aAgentCard:          &awstypes.UpdatedA2aAgentCardDescriptor{},
		AgentSkillsDefinition: &awstypes.UpdatedAgentSkillsDefinitionDescriptor{},
		Custom:                &awstypes.UpdatedCustomDescriptor{},
		McpServer:             &awstypes.UpdatedMcpServerDescriptor{},
	}

	if v := d.A2aAgentCard; v != nil {
		fields.A2aAgentCard.OptionalValue = &awstypes.UpdatedA2aAgentCardDescriptorFields{
			Data:              &awstypes.UpdatedDescriptorData{OptionalValue: v.Data},
			DataSchemaVersion: &awstypes.UpdatedDataSchemaVersion{OptionalValue: v.DataSchemaVersion},
			Source:            &awstypes.UpdatedDescriptorSource{OptionalValue: v.Source},
		}
	}

	if v := d.AgentSkillsDefinition; v != nil {
		additionalData := &awstypes.UpdatedAgentSkillsAdditionalData{}
		if ad := v.AdditionalData; ad != nil {
			skillMD := &awstypes.UpdatedAgentSkillsMdDescriptor{}
			if md := ad.SkillMd; md != nil {
				skillMD.OptionalValue = &awstypes.UpdatedAgentSkillsMdDescriptorFields{
					Data:              &awstypes.UpdatedDescriptorData{OptionalValue: md.Data},
					DataSchemaVersion: &awstypes.UpdatedDataSchemaVersion{OptionalValue: md.DataSchemaVersion},
					Source:            &awstypes.UpdatedDescriptorSource{OptionalValue: md.Source},
				}
			}
			additionalData.OptionalValue = &awstypes.UpdatedAgentSkillsAdditionalDataFields{
				SkillMd: skillMD,
			}
		}
		fields.AgentSkillsDefinition.OptionalValue = &awstypes.UpdatedAgentSkillsDefinitionDescriptorFields{
			AdditionalData:    additionalData,
			Data:              &awstypes.UpdatedDescriptorData{OptionalValue: v.Data},
			DataSchemaVersion: &awstypes.UpdatedDataSchemaVersion{OptionalValue: v.DataSchemaVersion},
		}
	}

	if v := d.Custom; v != nil {
		fields.Custom.OptionalValue = &awstypes.UpdatedCustomDescriptorFields{
			Data: &awstypes.UpdatedDescriptorData{OptionalValue: v.Data},
		}
	}

	if v := d.McpServer; v != nil {
		additionalData := &awstypes.UpdatedMcpServerAdditionalData{}
		if ad := v.AdditionalData; ad != nil {
			tools := &awstypes.UpdatedMcpToolsDescriptor{}
			if t := ad.Tools; t != nil {
				tools.OptionalValue = &awstypes.UpdatedMcpToolsDescriptorFields{
					Data:              &awstypes.UpdatedDescriptorData{OptionalValue: t.Data},
					DataSchemaVersion: &awstypes.UpdatedDataSchemaVersion{OptionalValue: t.DataSchemaVersion},
				}
			}
			additionalData.OptionalValue = &awstypes.UpdatedMcpServerAdditionalDataFields{
				Tools: tools,
			}
		}
		fields.McpServer.OptionalValue = &awstypes.UpdatedMcpServerDescriptorFields{
			AdditionalData:    additionalData,
			Data:              &awstypes.UpdatedDescriptorData{OptionalValue: v.Data},
			DataSchemaVersion: &awstypes.UpdatedDataSchemaVersion{OptionalValue: v.DataSchemaVersion},
			Source:            &awstypes.UpdatedDescriptorSource{OptionalValue: v.Source},
		}
	}

	return &awstypes.UpdatedDescriptors{OptionalValue: fields}
}

func findRegistryRecordByTwoPartKey(ctx context.Context, conn *agentregistrycontrol.Client, registryID, recordID string) (*agentregistrycontrol.GetRegistryRecordOutput, error) {
	input := agentregistrycontrol.GetRegistryRecordInput{
		RegistryId: aws.String(registryID),
		RecordId:   aws.String(recordID),
	}
	return findRegistryRecord(ctx, conn, &input)
}

func findRegistryRecord(ctx context.Context, conn *agentregistrycontrol.Client, input *agentregistrycontrol.GetRegistryRecordInput) (*agentregistrycontrol.GetRegistryRecordOutput, error) {
	out, err := conn.GetRegistryRecord(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, smarterr.NewError(err)
	}

	if out == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return out, nil
}

func waitRegistryRecordCreated(ctx context.Context, conn *agentregistrycontrol.Client, registryID, recordID string, timeout time.Duration) (*agentregistrycontrol.GetRegistryRecordOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.RegistryRecordStatusCreating),
		Target:                    enum.Slice(awstypes.RegistryRecordStatusDraft, awstypes.RegistryRecordStatusPendingApproval, awstypes.RegistryRecordStatusApproved),
		Refresh:                   statusRegistryRecord(conn, registryID, recordID),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*agentregistrycontrol.GetRegistryRecordOutput); ok {
		if out.Status == awstypes.RegistryRecordStatusCreateFailed {
			retry.SetLastError(err, errors.New(aws.ToString(out.StatusReason)))
		}
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func waitRegistryRecordUpdated(ctx context.Context, conn *agentregistrycontrol.Client, registryID, recordID string, timeout time.Duration) (*agentregistrycontrol.GetRegistryRecordOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.RegistryRecordStatusUpdating),
		Target:                    enum.Slice(awstypes.RegistryRecordStatusDraft, awstypes.RegistryRecordStatusPendingApproval, awstypes.RegistryRecordStatusApproved, awstypes.RegistryRecordStatusRejected, awstypes.RegistryRecordStatusDeprecated),
		Refresh:                   statusRegistryRecord(conn, registryID, recordID),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*agentregistrycontrol.GetRegistryRecordOutput); ok {
		if out.Status == awstypes.RegistryRecordStatusUpdateFailed {
			retry.SetLastError(err, errors.New(aws.ToString(out.StatusReason)))
		}
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func waitRegistryRecordDeleted(ctx context.Context, conn *agentregistrycontrol.Client, registryID, recordID string, timeout time.Duration) (*agentregistrycontrol.GetRegistryRecordOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Values[awstypes.RegistryRecordStatus](),
		Target:  []string{},
		Refresh: statusRegistryRecord(conn, registryID, recordID),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*agentregistrycontrol.GetRegistryRecordOutput); ok {
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func statusRegistryRecord(conn *agentregistrycontrol.Client, registryID, recordID string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		out, err := findRegistryRecordByTwoPartKey(ctx, conn, registryID, recordID)
		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", smarterr.NewError(err)
		}

		return out, string(out.Status), nil
	}
}

const registryRecordImportIDSeparator = intflex.ResourceIdSeparator

var (
	_ inttypes.ImportIDParser = registryRecordImportID{}
)

type registryRecordImportID struct{}

func (registryRecordImportID) Parse(id string) (string, map[string]any, error) {
	parts := strings.Split(id, registryRecordImportIDSeparator)

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", nil, fmt.Errorf("unexpected format for ID (%[1]s), expected registry-id%[2]srecord-id", id, registryRecordImportIDSeparator)
	}

	result := map[string]any{
		"registry_id": parts[0],
		"record_id":   parts[1],
	}

	return id, result, nil
}

type registryRecordResourceModel struct {
	framework.WithRegionModel
	Description   types.String                                      `tfsdk:"description"`
	Descriptors   fwtypes.ListNestedObjectValueOf[descriptorsModel] `tfsdk:"descriptors"`
	DisplayName   types.String                                      `tfsdk:"display_name"`
	Name          types.String                                      `tfsdk:"name"`
	RecordARN     types.String                                      `tfsdk:"record_arn"`
	RecordID      types.String                                      `tfsdk:"record_id"`
	RecordType    fwtypes.StringEnum[awstypes.RecordType]           `tfsdk:"record_type"`
	RecordVersion types.String                                      `tfsdk:"record_version"`
	RegistryID    types.String                                      `tfsdk:"registry_id"`
	Tags          tftags.Map                                        `tfsdk:"tags"`
	TagsAll       tftags.Map                                        `tfsdk:"tags_all"`
	Timeouts      timeouts.Value                                    `tfsdk:"timeouts"`
}

type descriptorsModel struct {
	A2AAgentCard          fwtypes.ListNestedObjectValueOf[a2aAgentCardDescriptorModel]          `tfsdk:"a2a_agent_card"`
	AgentSkillsDefinition fwtypes.ListNestedObjectValueOf[agentSkillsDefinitionDescriptorModel] `tfsdk:"agent_skills_definition"`
	Custom                fwtypes.ListNestedObjectValueOf[customDescriptorModel]                `tfsdk:"custom"`
	MCPServer             fwtypes.ListNestedObjectValueOf[mcpServerDescriptorModel]             `tfsdk:"mcp_server"`
}

type a2aAgentCardDescriptorModel struct {
	Data              types.String                                           `tfsdk:"data"`
	DataSchemaVersion types.String                                           `tfsdk:"data_schema_version"`
	Source            fwtypes.ListNestedObjectValueOf[descriptorSourceModel] `tfsdk:"source"`
}

type agentSkillsDefinitionDescriptorModel struct {
	AdditionalData    fwtypes.ListNestedObjectValueOf[agentSkillsAdditionalDataModel] `tfsdk:"additional_data"`
	Data              types.String                                                    `tfsdk:"data"`
	DataSchemaVersion types.String                                                    `tfsdk:"data_schema_version"`
}

type agentSkillsAdditionalDataModel struct {
	SkillMD fwtypes.ListNestedObjectValueOf[agentSkillsMDDescriptorModel] `tfsdk:"skill_md"`
}

type agentSkillsMDDescriptorModel struct {
	Data              types.String                                           `tfsdk:"data"`
	DataSchemaVersion types.String                                           `tfsdk:"data_schema_version"`
	Source            fwtypes.ListNestedObjectValueOf[descriptorSourceModel] `tfsdk:"source"`
}

type customDescriptorModel struct {
	Data types.String `tfsdk:"data"`
}

type mcpServerDescriptorModel struct {
	AdditionalData    fwtypes.ListNestedObjectValueOf[mcpServerAdditionalDataModel] `tfsdk:"additional_data"`
	Data              types.String                                                  `tfsdk:"data"`
	DataSchemaVersion types.String                                                  `tfsdk:"data_schema_version"`
	Source            fwtypes.ListNestedObjectValueOf[descriptorSourceModel]        `tfsdk:"source"`
}

type mcpServerAdditionalDataModel struct {
	Tools fwtypes.ListNestedObjectValueOf[mcpToolsDescriptorModel] `tfsdk:"tools"`
}

type mcpToolsDescriptorModel struct {
	Data              types.String `tfsdk:"data"`
	DataSchemaVersion types.String `tfsdk:"data_schema_version"`
}

type descriptorSourceModel struct {
	FromURL fwtypes.ListNestedObjectValueOf[descriptorSourceFromURLModel] `tfsdk:"from_url"`
}

type descriptorSourceFromURLModel struct {
	CredentialProviderConfigurations fwtypes.ListNestedObjectValueOf[credentialProviderConfigurationModel] `tfsdk:"credential_provider_configuration"`
	URL                              types.String                                                          `tfsdk:"url"`
}

type credentialProviderConfigurationModel struct {
	CredentialProvider     fwtypes.ListNestedObjectValueOf[credentialProviderModel]          `tfsdk:"credential_provider"`
	CredentialProviderType fwtypes.StringEnum[awstypes.RegistryRecordCredentialProviderType] `tfsdk:"credential_provider_type"`
}

type credentialProviderModel struct {
	IAMCredentialProvider   fwtypes.ListNestedObjectValueOf[iamCredentialProviderModel]   `tfsdk:"iam_credential_provider"`
	OAuthCredentialProvider fwtypes.ListNestedObjectValueOf[oauthCredentialProviderModel] `tfsdk:"oauth_credential_provider"`
}

var (
	_ fwflex.Expander  = credentialProviderModel{}
	_ fwflex.Flattener = &credentialProviderModel{}
)

func (m *credentialProviderModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics
	switch t := v.(type) {
	case awstypes.RegistryRecordCredentialProviderUnionMemberIamCredentialProvider:
		var model iamCredentialProviderModel
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t.Value, &model))
		if diags.HasError() {
			return diags
		}
		var d diag.Diagnostics
		m.IAMCredentialProvider, d = fwtypes.NewListNestedObjectValueOfPtr(ctx, &model)
		smerr.AddEnrich(ctx, &diags, d)

	case awstypes.RegistryRecordCredentialProviderUnionMemberOauthCredentialProvider:
		var model oauthCredentialProviderModel
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t.Value, &model))
		if diags.HasError() {
			return diags
		}
		var d diag.Diagnostics
		m.OAuthCredentialProvider, d = fwtypes.NewListNestedObjectValueOfPtr(ctx, &model)
		smerr.AddEnrich(ctx, &diags, d)

	default:
		diags.AddError(
			"Unsupported Type",
			fmt.Sprintf("credentialProviderModel.Flatten: %T", v),
		)
	}

	return diags
}

func (m credentialProviderModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	switch {
	case !m.IAMCredentialProvider.IsNull():
		model, d := m.IAMCredentialProvider.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
		var r awstypes.RegistryRecordCredentialProviderUnionMemberIamCredentialProvider
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, model, &r.Value))
		return &r, diags

	case !m.OAuthCredentialProvider.IsNull():
		model, d := m.OAuthCredentialProvider.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
		var r awstypes.RegistryRecordCredentialProviderUnionMemberOauthCredentialProvider
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, model, &r.Value))
		return &r, diags
	}

	return nil, diags
}

type iamCredentialProviderModel struct {
	Region  types.String `tfsdk:"region"`
	RoleARN fwtypes.ARN  `tfsdk:"role_arn"`
	Service types.String `tfsdk:"service"`
}

type oauthCredentialProviderModel struct {
	CustomParameters fwtypes.MapOfString                                       `tfsdk:"custom_parameters"`
	GrantType        fwtypes.StringEnum[awstypes.RegistryRecordOAuthGrantType] `tfsdk:"grant_type"`
	ProviderARN      fwtypes.ARN                                               `tfsdk:"provider_arn"`
	Scopes           fwtypes.ListOfString                                      `tfsdk:"scopes"`
}
