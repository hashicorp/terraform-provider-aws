// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package bedrockagentcore

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol/types"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
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
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
	// fwvalidators "github.com/hashicorp/terraform-provider-aws/internal/framework/validators"
)

// @FrameworkResource("aws_bedrockagentcore_registry_record", name="Registry Record")
// @IdentityAttribute("registry_id")
// @IdentityAttribute("record_id")
// @ImportIDHandler(registryRecordImportID)
// @Testing(hasNoPreExistingResource=true)
// @Testing(importStateIdAttribute="record_id")
// @Testing(preCheck="testAccPreCheckRegistries")
// @Testing(importStateIdFunc="testAccRegistryRecordImportStateIDFunc")
func newRegistryRecordResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &registryRecordResource{}

	r.SetDefaultCreateTimeout(5 * time.Minute)
	r.SetDefaultUpdateTimeout(5 * time.Minute)

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
			"descriptor_type": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.DescriptorType](),
				Required:   true,
			},
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 4096),
				},
			},
			names.AttrName: schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 255),
				},
			},
			"record_arn": framework.ARNAttributeComputedOnly(),
			"record_id":  framework.IDAttribute(),
			"record_version": schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 255),
				},
			},
			"registry_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrStatus: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.RegistryRecordStatus](),
				Computed:   true,
			},
			// "synchronization_type": schema.StringAttribute{
			// 	CustomType: fwtypes.StringEnumType[awstypes.SynchronizationType](),
			// 	Optional:   true,
			// },
		},
		Blocks: map[string]schema.Block{
			"descriptors": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[descriptorsModel](ctx),
				Validators: []validator.List{
					// listvalidator.AtLeastOneOf(path.MatchRoot("descriptors"), path.MatchRoot("synchronization_configuration")),
					listvalidator.IsRequired(),
					listvalidator.SizeAtLeast(1),
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Validators: []validator.Object{
						tfobjectvalidator.ExactlyOneOfChildren(
							path.MatchRelative().AtName("a2a"),
							path.MatchRelative().AtName("agent_skills"),
							path.MatchRelative().AtName("custom"),
							path.MatchRelative().AtName("mcp"),
						),
					},
					Blocks: map[string]schema.Block{
						"a2a": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[a2aDescriptorModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Blocks: map[string]schema.Block{
									"agent_card": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[agentCardDefinitionModel](ctx),
										Validators: []validator.List{
											listvalidator.IsRequired(),
											listvalidator.SizeAtLeast(1),
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"inline_content": schema.StringAttribute{
													CustomType: jsontypes.NormalizedType{},
													Required:   true,
												},
												"schema_version": schema.StringAttribute{
													Optional: true,
													Computed: true,
													Validators: []validator.String{
														stringvalidator.LengthBetween(1, 255),
													},
												},
											},
										},
									},
								},
							},
						},
						"agent_skills": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[agentSkillsDescriptorModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Validators: []validator.Object{
									tfobjectvalidator.ExactlyOneOfChildren(
										path.MatchRelative().AtName("skill_definition"),
										path.MatchRelative().AtName("skill_md"),
									),
								},
								Blocks: map[string]schema.Block{
									"skill_definition": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[skillDefinitionModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"inline_content": schema.StringAttribute{
													CustomType: jsontypes.NormalizedType{},
													Required:   true,
												},
												"schema_version": schema.StringAttribute{
													Optional: true,
													Validators: []validator.String{
														stringvalidator.LengthBetween(1, 255),
													},
												},
											},
										},
									},
									"skill_md": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[skillMdDefinitionModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"inline_content": schema.StringAttribute{
													Required: true,
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
									"inline_content": schema.StringAttribute{
										CustomType: jsontypes.NormalizedType{},
										Required:   true,
									},
								},
							},
						},
						"mcp": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[mcpDescriptorModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Blocks: map[string]schema.Block{
									"server": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[serverDefinitionModel](ctx),
										Validators: []validator.List{
											listvalidator.IsRequired(),
											listvalidator.SizeAtLeast(1),
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"inline_content": schema.StringAttribute{
													CustomType: jsontypes.NormalizedType{},
													Required:   true,
												},
												"schema_version": schema.StringAttribute{
													Optional: true,
													Computed: true,
													Validators: []validator.String{
														stringvalidator.LengthBetween(1, 255),
													},
												},
											},
										},
									},
									"tools": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[toolsDefinitionModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"inline_content": schema.StringAttribute{
													CustomType: jsontypes.NormalizedType{},
													Required:   true,
												},
												"schema_version": schema.StringAttribute{
													Optional: true,
													Computed: true,
													Validators: []validator.String{
														stringvalidator.LengthBetween(1, 255),
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
			/*
				"synchronization_configuration": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[synchronizationConfigurationModel](ctx),
					Validators: []validator.List{
						listvalidator.SizeAtMost(1),
						listvalidator.AlsoRequires(path.MatchRoot("synchronization_type")),
					},
					NestedObject: schema.NestedBlockObject{
						Blocks: map[string]schema.Block{
							"from_url": schema.ListNestedBlock{
								CustomType: fwtypes.NewListNestedObjectTypeOf[fromURLSynchronizationConfigurationModel](ctx),
								Validators: []validator.List{
									listvalidator.SizeAtMost(1),
								},
								NestedObject: schema.NestedBlockObject{
									Attributes: map[string]schema.Attribute{
										"url": schema.StringAttribute{
											Required: true,
										},
									},
									Blocks: map[string]schema.Block{
										"credential_provider_configuration": schema.ListNestedBlock{
											CustomType: fwtypes.NewListNestedObjectTypeOf[registryRecordCredentialProviderConfigurationModel](ctx),
											Validators: []validator.List{
												listvalidator.SizeAtMost(1),
											},
											NestedObject: schema.NestedBlockObject{
												Validators: []validator.Object{
													tfobjectvalidator.ExactlyOneOfChildren(
														path.MatchRelative().AtName("iam"),
														path.MatchRelative().AtName("oauth"),
													),
												},
												Attributes: map[string]schema.Attribute{
													"credential_provider_type": schema.StringAttribute{
														CustomType: fwtypes.StringEnumType[awstypes.RegistryRecordCredentialProviderType](),
														Required:   true,
													},
												},
												Blocks: map[string]schema.Block{
													"iam": schema.ListNestedBlock{
														CustomType: fwtypes.NewListNestedObjectTypeOf[registryRecordIAMCredentialProviderModel](ctx),
														Validators: []validator.List{
															listvalidator.SizeAtMost(1),
														},
														NestedObject: schema.NestedBlockObject{
															Attributes: map[string]schema.Attribute{
																"region": schema.StringAttribute{
																	Optional: true,
																	Validators: []validator.String{
																		fwvalidators.AWSRegion(),
																	},
																},
																"role_arn": schema.StringAttribute{
																	CustomType: fwtypes.ARNType,
																	Optional:   true,
																},
																"service": schema.StringAttribute{
																	Optional: true,
																},
															},
														},
													},
													"oauth": schema.ListNestedBlock{
														CustomType: fwtypes.NewListNestedObjectTypeOf[registryRecordOAuthCredentialProviderModel](ctx),
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
																"scopes": schema.SetAttribute{
																	CustomType:  fwtypes.SetOfStringType,
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
			*/
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
			}),
		},
	}
}

func (r *registryRecordResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan registryRecordResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentCoreClient(ctx)

	name := fwflex.StringValueFromFramework(ctx, plan.Name)
	var input bedrockagentcorecontrol.CreateRegistryRecordInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &input))
	if resp.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	input.ClientToken = aws.String(create.UniqueId(ctx))

	out, err := conn.CreateRegistryRecord(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, name)
		return
	}

	// e.g. arn:aws:bedrock-agentcore:us-west-2:1234567890:registry/N0QqP8RXK3FJkJMg/record/dmzgjnmXO3z7
	recordARN := aws.ToString(out.RecordArn)
	registryID, recordID := fwflex.StringValueFromFramework(ctx, plan.RegistryID), recordARN[strings.LastIndex(recordARN, "/")+1:]
	created, err := waitRegistryRecordCreated(ctx, conn, registryID, recordID, r.CreateTimeout(ctx, plan.Timeouts))
	if err != nil {
		// Taint the resource.
		resp.State.SetAttribute(ctx, path.Root("registry_id"), registryID)
		resp.State.SetAttribute(ctx, path.Root("record_id"), recordID)
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, recordARN)
		return
	}

	// Set values for unknowns.
	smerr.AddEnrich(ctx, &resp.Diagnostics, r.flatten(ctx, created, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *registryRecordResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state registryRecordResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentCoreClient(ctx)

	registryID, recordID := fwflex.StringValueFromFramework(ctx, state.RegistryID), fwflex.StringValueFromFramework(ctx, state.RecordID)
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

	smerr.AddEnrich(ctx, &resp.Diagnostics, r.flatten(ctx, out, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *registryRecordResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state registryRecordResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentCoreClient(ctx)

	diff, d := fwflex.Diff(ctx, plan, state)
	smerr.AddEnrich(ctx, &resp.Diagnostics, d)
	if resp.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		registryID, recordID := fwflex.StringValueFromFramework(ctx, plan.RegistryID), fwflex.StringValueFromFramework(ctx, plan.RecordID)
		var input bedrockagentcorecontrol.UpdateRegistryRecordInput
		smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &input, fwflex.WithIgnoredFieldNamesAppend("Description"), fwflex.WithIgnoredFieldNamesAppend("Descriptors")))
		if resp.Diagnostics.HasError() {
			return
		}

		// Additional fields.
		input.Description = updatedDescription(ctx, plan.Description, state.Description)
		if !plan.Descriptors.Equal(state.Descriptors) {
			if plan.Descriptors.IsNull() {
				input.Descriptors = &awstypes.UpdatedDescriptors{}
			} else {
				smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan.Descriptors, &input.Descriptors))
				if resp.Diagnostics.HasError() {
					return
				}
			}
		}

		_, err := conn.UpdateRegistryRecord(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, recordID)
			return
		}

		updated, err := waitRegistryRecordUpdated(ctx, conn, registryID, recordID, r.UpdateTimeout(ctx, plan.Timeouts))
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, recordID)
			return
		}

		// Set values for unknowns.
		smerr.AddEnrich(ctx, &resp.Diagnostics, r.flatten(ctx, updated, &plan))
		if resp.Diagnostics.HasError() {
			return
		}
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *registryRecordResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state registryRecordResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentCoreClient(ctx)

	registryID, recordID := fwflex.StringValueFromFramework(ctx, state.RegistryID), fwflex.StringValueFromFramework(ctx, state.RecordID)
	input := bedrockagentcorecontrol.DeleteRegistryRecordInput{
		RecordId:   aws.String(recordID),
		RegistryId: aws.String(registryID),
	}

	_, err := conn.DeleteRegistryRecord(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, recordID)
		return
	}
}

func (r *registryRecordResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data registryRecordResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Config.Get(ctx, &data))
	if resp.Diagnostics.HasError() {
		return
	}

	if !data.Descriptors.IsUnknown() {
		descriptors, diags := data.Descriptors.ToPtr(ctx)
		smerr.AddEnrich(ctx, &resp.Diagnostics, diags)
		if resp.Diagnostics.HasError() {
			return
		}

		switch descriptorType := data.DescriptorType.ValueEnum(); descriptorType {
		case awstypes.DescriptorTypeA2a:
			if descriptors != nil && descriptors.A2A.IsNull() {
				resp.Diagnostics.AddAttributeError(
					path.Root("descriptors"),
					"Missing Configuration",
					fmt.Sprintf("descriptors.a2a is required when descriptor_type is %q", descriptorType),
				)
			}
		case awstypes.DescriptorTypeAgentSkills:
			if descriptors != nil && descriptors.AgentSkills.IsNull() {
				resp.Diagnostics.AddAttributeError(
					path.Root("descriptors"),
					"Missing Configuration",
					fmt.Sprintf("descriptors.agent_skills is required when descriptor_type is %q", descriptorType),
				)
			}
		case awstypes.DescriptorTypeCustom:
			if descriptors != nil && descriptors.Custom.IsNull() {
				resp.Diagnostics.AddAttributeError(
					path.Root("descriptors"),
					"Missing Configuration",
					fmt.Sprintf("descriptors.custom is required when descriptor_type is %q", descriptorType),
				)
			}
		case awstypes.DescriptorTypeMcp:
			if descriptors != nil && descriptors.MCP.IsNull() {
				resp.Diagnostics.AddAttributeError(
					path.Root("descriptors"),
					"Missing Configuration",
					fmt.Sprintf("descriptors.mcp is required when descriptor_type is %q", descriptorType),
				)
			}
		}

		if resp.Diagnostics.HasError() {
			return
		}
	}

	/*
		if !data.SynchronizationConfiguration.IsUnknown() {
			synchronizationConfiguration, diags := data.SynchronizationConfiguration.ToPtr(ctx)
			smerr.AddEnrich(ctx, &resp.Diagnostics, diags)
			if resp.Diagnostics.HasError() {
				return
			}

			switch synchronizationType := data.SynchronizationType.ValueEnum(); synchronizationType {
			case awstypes.SynchronizationTypeUrl:
				if synchronizationConfiguration != nil && synchronizationConfiguration.FromURL.IsNull() {
					resp.Diagnostics.AddAttributeError(
						path.Root("synchronization_configuration"),
						"Missing Configuration",
						fmt.Sprintf("synchronization_configuration.from_url is required when synchronization_type is %q", synchronizationType),
					)
				}
			}
		}
	*/
}

func (r *registryRecordResource) flatten(ctx context.Context, registryRecord *bedrockagentcorecontrol.GetRegistryRecordOutput, data *registryRecordResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics
	diags.Append(fwflex.Flatten(ctx, registryRecord, data)...)
	return diags
}

func findRegistryRecordByTwoPartKey(ctx context.Context, conn *bedrockagentcorecontrol.Client, registryID, recordID string) (*bedrockagentcorecontrol.GetRegistryRecordOutput, error) {
	input := bedrockagentcorecontrol.GetRegistryRecordInput{
		RecordId:   aws.String(recordID),
		RegistryId: aws.String(registryID),
	}

	return findRegistryRecord(ctx, conn, &input)
}

func findRegistryRecord(ctx context.Context, conn *bedrockagentcorecontrol.Client, input *bedrockagentcorecontrol.GetRegistryRecordInput) (*bedrockagentcorecontrol.GetRegistryRecordOutput, error) {
	out, err := conn.GetRegistryRecord(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, smarterr.NewError(&retry.NotFoundError{
			LastError: err,
		})
	}

	if err != nil {
		return nil, smarterr.NewError(err)
	}

	if out == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}

	return out, nil
}

func statusRegistryRecord(conn *bedrockagentcorecontrol.Client, registryID, recordID string) retry.StateRefreshFunc {
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

// https://docs.aws.amazon.com/bedrock-agentcore/latest/devguide/registry-record-lifecycle.html.

func waitRegistryRecordCreated(ctx context.Context, conn *bedrockagentcorecontrol.Client, registryID, recordID string, timeout time.Duration) (*bedrockagentcorecontrol.GetRegistryRecordOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.RegistryRecordStatusCreating),
		Target:  enum.Slice(awstypes.RegistryRecordStatusDraft),
		Refresh: statusRegistryRecord(conn, registryID, recordID),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*bedrockagentcorecontrol.GetRegistryRecordOutput); ok {
		retry.SetLastError(err, errors.New(aws.ToString(out.StatusReason)))
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func waitRegistryRecordUpdated(ctx context.Context, conn *bedrockagentcorecontrol.Client, registryID, recordID string, timeout time.Duration) (*bedrockagentcorecontrol.GetRegistryRecordOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.RegistryRecordStatusUpdating),
		Target:                    enum.Slice(awstypes.RegistryRecordStatusDraft, awstypes.RegistryRecordStatusPendingApproval, awstypes.RegistryRecordStatusApproved, awstypes.RegistryRecordStatusRejected),
		Refresh:                   statusRegistryRecord(conn, registryID, recordID),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*bedrockagentcorecontrol.GetRegistryRecordOutput); ok {
		retry.SetLastError(err, errors.New(aws.ToString(out.StatusReason)))
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

var (
	_ inttypes.ImportIDParser = registryRecordImportID{}
)

type registryRecordImportID struct{}

func (registryRecordImportID) Parse(id string) (string, map[string]any, error) {
	const (
		registryRecordIDParts = 2
	)
	parts, err := intflex.ExpandResourceId(id, registryRecordIDParts, true)

	if err != nil {
		return "", nil, err
	}

	result := map[string]any{
		"registry_id": parts[0],
		"record_id":   parts[1],
	}

	return id, result, nil
}

type registryRecordResourceModel struct {
	framework.WithRegionModel
	Description    types.String                                      `tfsdk:"description"`
	DescriptorType fwtypes.StringEnum[awstypes.DescriptorType]       `tfsdk:"descriptor_type"`
	Descriptors    fwtypes.ListNestedObjectValueOf[descriptorsModel] `tfsdk:"descriptors"`
	Name           types.String                                      `tfsdk:"name"`
	RecordARN      types.String                                      `tfsdk:"record_arn"`
	RecordID       types.String                                      `tfsdk:"record_id"`
	RecordVersion  types.String                                      `tfsdk:"record_version"`
	RegistryID     types.String                                      `tfsdk:"registry_id"`
	Status         fwtypes.StringEnum[awstypes.RegistryRecordStatus] `tfsdk:"status"`
	// SynchronizationConfiguration fwtypes.ListNestedObjectValueOf[synchronizationConfigurationModel] `tfsdk:"synchronization_configuration"`
	// SynchronizationType          fwtypes.StringEnum[awstypes.SynchronizationType]                   `tfsdk:"synchronization_type"`
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

type descriptorsModel struct {
	A2A         fwtypes.ListNestedObjectValueOf[a2aDescriptorModel]         `tfsdk:"a2a"`
	AgentSkills fwtypes.ListNestedObjectValueOf[agentSkillsDescriptorModel] `tfsdk:"agent_skills"`
	Custom      fwtypes.ListNestedObjectValueOf[customDescriptorModel]      `tfsdk:"custom"`
	MCP         fwtypes.ListNestedObjectValueOf[mcpDescriptorModel]         `tfsdk:"mcp"`
}

var (
	_ fwflex.TypedExpander = descriptorsModel{}
)

func (m descriptorsModel) ExpandTo(ctx context.Context, targetType reflect.Type) (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	switch targetType {
	case reflect.TypeFor[awstypes.Descriptors]():
		return m.expandToDescriptors(ctx)
	case reflect.TypeFor[awstypes.UpdatedDescriptors]():
		return m.expandToUpdatedDescriptors(ctx)
	default:
		diags.AddError("Unsupported Type", fmt.Sprintf("descriptorsModel ExpandTo: %q", targetType))
	}
	return nil, diags
}

func (m descriptorsModel) expandToDescriptors(ctx context.Context) (awstypes.Descriptors, diag.Diagnostics) {
	var diags diag.Diagnostics
	var r awstypes.Descriptors
	if !m.A2A.IsNull() {
		data, d := m.A2A.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return inttypes.Zero[awstypes.Descriptors](), diags
		}
		var v awstypes.A2aDescriptor
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, data, &v))
		if diags.HasError() {
			return inttypes.Zero[awstypes.Descriptors](), diags
		}
		r.A2a = &v
	}
	if !m.AgentSkills.IsNull() {
		data, d := m.AgentSkills.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return inttypes.Zero[awstypes.Descriptors](), diags
		}
		var v awstypes.AgentSkillsDescriptor
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, data, &v))
		if diags.HasError() {
			return inttypes.Zero[awstypes.Descriptors](), diags
		}
		r.AgentSkills = &v
	}
	if !m.Custom.IsNull() {
		data, d := m.Custom.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return inttypes.Zero[awstypes.Descriptors](), diags
		}
		var v awstypes.CustomDescriptor
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, data, &v))
		if diags.HasError() {
			return inttypes.Zero[awstypes.Descriptors](), diags
		}
		r.Custom = &v
	}
	if !m.MCP.IsNull() {
		data, d := m.MCP.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return inttypes.Zero[awstypes.Descriptors](), diags
		}
		var v awstypes.McpDescriptor
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, data, &v))
		if diags.HasError() {
			return inttypes.Zero[awstypes.Descriptors](), diags
		}
		r.Mcp = &v
	}
	return r, diags
}

func (m descriptorsModel) expandToUpdatedDescriptors(ctx context.Context) (awstypes.UpdatedDescriptors, diag.Diagnostics) {
	var diags diag.Diagnostics
	r := awstypes.UpdatedDescriptorsUnion{
		A2a:         &awstypes.UpdatedA2aDescriptor{},
		AgentSkills: &awstypes.UpdatedAgentSkillsDescriptor{},
		Custom:      &awstypes.UpdatedCustomDescriptor{},
		Mcp:         &awstypes.UpdatedMcpDescriptor{},
	}
	if !m.A2A.IsNull() {
		data, d := m.A2A.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return inttypes.Zero[awstypes.UpdatedDescriptors](), diags
		}
		var v awstypes.A2aDescriptor
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, data, &v))
		if diags.HasError() {
			return inttypes.Zero[awstypes.UpdatedDescriptors](), diags
		}
		r.A2a.OptionalValue = &v
	}
	if !m.AgentSkills.IsNull() {
		data, d := m.AgentSkills.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return inttypes.Zero[awstypes.UpdatedDescriptors](), diags
		}
		var v awstypes.UpdatedAgentSkillsDescriptorFields
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, data, &v))
		if diags.HasError() {
			return inttypes.Zero[awstypes.UpdatedDescriptors](), diags
		}
		r.AgentSkills.OptionalValue = &v
	}
	if !m.Custom.IsNull() {
		data, d := m.Custom.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return inttypes.Zero[awstypes.UpdatedDescriptors](), diags
		}
		var v awstypes.CustomDescriptor
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, data, &v))
		if diags.HasError() {
			return inttypes.Zero[awstypes.UpdatedDescriptors](), diags
		}
		r.Custom.OptionalValue = &v
	}
	if !m.MCP.IsNull() {
		data, d := m.MCP.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return inttypes.Zero[awstypes.UpdatedDescriptors](), diags
		}
		var v awstypes.UpdatedMcpDescriptorFields
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, data, &v))
		if diags.HasError() {
			return inttypes.Zero[awstypes.UpdatedDescriptors](), diags
		}
		r.Mcp.OptionalValue = &v
	}
	return awstypes.UpdatedDescriptors{OptionalValue: &r}, diags
}

type a2aDescriptorModel struct {
	AgentCard fwtypes.ListNestedObjectValueOf[agentCardDefinitionModel] `tfsdk:"agent_card"`
}

type agentCardDefinitionModel struct {
	InlineContent jsontypes.Normalized `tfsdk:"inline_content"`
	SchemaVersion types.String         `tfsdk:"schema_version"`
}

type agentSkillsDescriptorModel struct {
	SkillDefinition fwtypes.ListNestedObjectValueOf[skillDefinitionModel]   `tfsdk:"skill_definition"`
	SkillMd         fwtypes.ListNestedObjectValueOf[skillMdDefinitionModel] `tfsdk:"skill_md"`
}

var (
	_ fwflex.TypedExpander = agentSkillsDescriptorModel{}
)

func (m agentSkillsDescriptorModel) ExpandTo(ctx context.Context, targetType reflect.Type) (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	switch targetType {
	case reflect.TypeFor[awstypes.AgentSkillsDescriptor]():
		return m.expandToAgentSkillsDescriptor(ctx)
	case reflect.TypeFor[awstypes.UpdatedAgentSkillsDescriptorFields]():
		return m.expandToUpdatedAgentSkillsDescriptorFields(ctx)
	default:
		diags.AddError("Unsupported Type", fmt.Sprintf("agentSkillsDescriptorModel ExpandTo: %q", targetType))
	}
	return nil, diags
}

func (m agentSkillsDescriptorModel) expandToAgentSkillsDescriptor(ctx context.Context) (awstypes.AgentSkillsDescriptor, diag.Diagnostics) {
	var diags diag.Diagnostics
	var r awstypes.AgentSkillsDescriptor
	if !m.SkillDefinition.IsNull() {
		data, d := m.SkillDefinition.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return inttypes.Zero[awstypes.AgentSkillsDescriptor](), diags
		}
		var v awstypes.SkillDefinition
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, data, &v))
		if diags.HasError() {
			return inttypes.Zero[awstypes.AgentSkillsDescriptor](), diags
		}
		r.SkillDefinition = &v
	}
	if !m.SkillMd.IsNull() {
		data, d := m.SkillMd.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return inttypes.Zero[awstypes.AgentSkillsDescriptor](), diags
		}
		var v awstypes.SkillMdDefinition
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, data, &v))
		if diags.HasError() {
			return inttypes.Zero[awstypes.AgentSkillsDescriptor](), diags
		}
		r.SkillMd = &v
	}
	return r, diags
}

func (m agentSkillsDescriptorModel) expandToUpdatedAgentSkillsDescriptorFields(ctx context.Context) (awstypes.UpdatedAgentSkillsDescriptorFields, diag.Diagnostics) {
	var diags diag.Diagnostics
	r := awstypes.UpdatedAgentSkillsDescriptorFields{
		SkillDefinition: &awstypes.UpdatedSkillDefinition{},
		SkillMd:         &awstypes.UpdatedSkillMdDefinition{},
	}
	if !m.SkillDefinition.IsNull() {
		data, d := m.SkillDefinition.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return inttypes.Zero[awstypes.UpdatedAgentSkillsDescriptorFields](), diags
		}
		var v awstypes.SkillDefinition
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, data, &v))
		if diags.HasError() {
			return inttypes.Zero[awstypes.UpdatedAgentSkillsDescriptorFields](), diags
		}
		r.SkillDefinition.OptionalValue = &v
	}
	if !m.SkillMd.IsNull() {
		data, d := m.SkillMd.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return inttypes.Zero[awstypes.UpdatedAgentSkillsDescriptorFields](), diags
		}
		var v awstypes.SkillMdDefinition
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, data, &v))
		if diags.HasError() {
			return inttypes.Zero[awstypes.UpdatedAgentSkillsDescriptorFields](), diags
		}
		r.SkillMd.OptionalValue = &v
	}
	return r, diags
}

type skillDefinitionModel struct {
	InlineContent jsontypes.Normalized `tfsdk:"inline_content"`
	SchemaVersion types.String         `tfsdk:"schema_version"`
}

type skillMdDefinitionModel struct {
	InlineContent types.String `tfsdk:"inline_content"`
}

type customDescriptorModel struct {
	InlineContent jsontypes.Normalized `tfsdk:"inline_content"`
}

type mcpDescriptorModel struct {
	Server fwtypes.ListNestedObjectValueOf[serverDefinitionModel] `tfsdk:"server"`
	Tools  fwtypes.ListNestedObjectValueOf[toolsDefinitionModel]  `tfsdk:"tools"`
}

type serverDefinitionModel struct {
	InlineContent jsontypes.Normalized `tfsdk:"inline_content"`
	SchemaVersion types.String         `tfsdk:"schema_version"`
}

type toolsDefinitionModel struct {
	InlineContent jsontypes.Normalized `tfsdk:"inline_content"`
	SchemaVersion types.String         `tfsdk:"schema_version"`
}

/*
type synchronizationConfigurationModel struct {
	FromURL fwtypes.ListNestedObjectValueOf[fromURLSynchronizationConfigurationModel] `tfsdk:"from_url"`
}

type fromURLSynchronizationConfigurationModel struct {
	CredentialProviderConfigurations fwtypes.ListNestedObjectValueOf[registryRecordCredentialProviderConfigurationModel] `tfsdk:"credential_provider_configuration"`
	URL                              types.String                                                                        `tfsdk:"url"`
}

type registryRecordCredentialProviderConfigurationModel struct {
	CredentialProviderType fwtypes.StringEnum[awstypes.RegistryRecordCredentialProviderType]      `tfsdk:"credential_provider_type"`
	CredentialProvider     fwtypes.ListNestedObjectValueOf[registryRecordCredentialProviderModel] `tfsdk:"credential_provider"`
}

type registryRecordCredentialProviderModel struct {
	IAM   fwtypes.ListNestedObjectValueOf[registryRecordIAMCredentialProviderModel]   `tfsdk:"iam"`
	OAuth fwtypes.ListNestedObjectValueOf[registryRecordOAuthCredentialProviderModel] `tfsdk:"oauth"`
}

var (
	_ fwflex.Expander  = registryRecordCredentialProviderModel{}
	_ fwflex.Flattener = &registryRecordCredentialProviderModel{}
)

func (m *registryRecordCredentialProviderModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics
	switch t := v.(type) {
	case awstypes.RegistryRecordCredentialProviderUnionMemberIamCredentialProvider:
		var data registryRecordIAMCredentialProviderModel
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t.Value, &data))
		if diags.HasError() {
			return diags
		}
		m.IAM = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &data)
	case awstypes.RegistryRecordCredentialProviderUnionMemberOauthCredentialProvider:
		var data registryRecordOAuthCredentialProviderModel
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t.Value, &data))
		if diags.HasError() {
			return diags
		}
		m.OAuth = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &data)
	default:
		diags.AddError("Unsupported Type", fmt.Sprintf("model registryRecordCredentialProvider flatten: %T", v))
	}
	return diags
}

func (m registryRecordCredentialProviderModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	switch {
	case !m.IAM.IsNull():
		data, d := m.IAM.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
		var r awstypes.RegistryRecordCredentialProviderUnionMemberIamCredentialProvider
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, data, &r.Value))
		return &r, diags
	case !m.OAuth.IsNull():
		data, d := m.OAuth.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
		var r awstypes.RegistryRecordCredentialProviderUnionMemberOauthCredentialProvider
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, data, &r.Value))
		return &r, diags
	}
	return nil, diags
}

type registryRecordIAMCredentialProviderModel struct {
	Region  types.String `tfsdk:"region"`
	RoleARN fwtypes.ARN  `tfsdk:"role_arn"`
	Service types.String `tfsdk:"service"`
}

type registryRecordOAuthCredentialProviderModel struct {
	CustomParameters fwtypes.MapOfString                                       `tfsdk:"custom_parameters"`
	GrantType        fwtypes.StringEnum[awstypes.RegistryRecordOAuthGrantType] `tfsdk:"grant_type"`
	ProviderARN      fwtypes.ARN                                               `tfsdk:"provider_arn"`
	Scopes           fwtypes.SetOfString                                       `tfsdk:"scopes"`
}
*/
