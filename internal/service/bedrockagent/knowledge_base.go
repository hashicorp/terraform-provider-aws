// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrockagent

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagent"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrockagent/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_bedrockagent_knowledge_base", name="Knowledge Base")
// @Tags(identifierAttribute="arn")
func newKnowledgeBaseResource(context.Context) (resource.ResourceWithConfigure, error) {
	r := &knowledgeBaseResource{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

type knowledgeBaseResource struct {
	framework.ResourceWithModel[knowledgeBaseResourceModel]
	framework.WithImportByID
	framework.WithTimeouts
}

func (r *knowledgeBaseResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrCreatedAt: schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
			},
			"failure_reasons": schema.ListAttribute{
				CustomType:  fwtypes.ListOfStringType,
				ElementType: types.StringType,
				Computed:    true,
			},
			names.AttrID: framework.IDAttribute(),
			names.AttrName: schema.StringAttribute{
				Required: true,
			},
			names.AttrRoleARN: schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			"updated_at": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
		},
		Blocks: map[string]schema.Block{
			"knowledge_base_configuration": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[knowledgeBaseConfigurationModel](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtLeast(1),
					listvalidator.SizeAtMost(1),
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrType: schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.KnowledgeBaseType](),
							Required:   true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
					},
					Blocks: map[string]schema.Block{
						"kendra_knowledge_base_configuration": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[kendraKnowledgeBaseConfigurationModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
								listvalidator.ExactlyOneOf(
									path.MatchRelative().AtParent().AtName("kendra_knowledge_base_configuration"),
									path.MatchRelative().AtParent().AtName("vector_knowledge_base_configuration"),
								),
							},
							PlanModifiers: []planmodifier.List{
								listplanmodifier.RequiresReplace(),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"kendra_index_arn": schema.StringAttribute{
										CustomType: fwtypes.ARNType,
										Required:   true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.RequiresReplace(),
										},
									},
								},
							},
						},
						"vector_knowledge_base_configuration": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[vectorKnowledgeBaseConfigurationModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							PlanModifiers: []planmodifier.List{
								listplanmodifier.RequiresReplace(),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"embedding_model_arn": schema.StringAttribute{
										CustomType: fwtypes.ARNType,
										Required:   true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.RequiresReplace(),
										},
									},
								},
								Blocks: map[string]schema.Block{
									"embedding_model_configuration": schema.ListNestedBlock{
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Blocks: map[string]schema.Block{
												"bedrock_embedding_model_configuration": schema.ListNestedBlock{
													Validators: []validator.List{
														listvalidator.SizeAtMost(1),
													},
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															"dimensions": schema.Int64Attribute{
																Optional: true,
															},
															"embedding_data_type": schema.StringAttribute{
																CustomType: fwtypes.StringEnumType[awstypes.EmbeddingDataType](),
																Optional:   true,
															},
														},
													},
												},
											},
										},
									},
									"supplemental_data_storage_configuration": schema.ListNestedBlock{
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Blocks: map[string]schema.Block{
												"storage_location": schema.ListNestedBlock{
													Validators: []validator.List{
														listvalidator.IsRequired(),
														listvalidator.SizeAtLeast(1),
													},
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															names.AttrType: schema.StringAttribute{
																CustomType: fwtypes.StringEnumType[awstypes.SupplementalDataStorageLocationType](),
																Required:   true,
															},
														},
														Blocks: map[string]schema.Block{
															"s3_location": schema.ListNestedBlock{
																Validators: []validator.List{
																	listvalidator.SizeAtMost(1),
																},
																NestedObject: schema.NestedBlockObject{
																	Attributes: map[string]schema.Attribute{
																		names.AttrURI: schema.StringAttribute{
																			Required: true,
																			Validators: []validator.String{
																				stringvalidator.RegexMatches(
																					regexache.MustCompile(`^s3://[a-z0-9.-]+(/.*)?$`),
																					"must be a valid S3 URI",
																				),
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
				},
			},
			"storage_configuration": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[storageConfigurationModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrType: schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.KnowledgeBaseStorageType](),
							Required:   true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
					},
					Blocks: map[string]schema.Block{
						"opensearch_managed_cluster_configuration": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[openSearchManagedClusterConfigurationModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
								listvalidator.ExactlyOneOf(
									path.MatchRelative().AtParent().AtName("opensearch_managed_cluster_configuration"),
									path.MatchRelative().AtParent().AtName("opensearch_serverless_configuration"),
									path.MatchRelative().AtParent().AtName("pinecone_configuration"),
									path.MatchRelative().AtParent().AtName("rds_configuration"),
									path.MatchRelative().AtParent().AtName("redis_enterprise_cloud_configuration"),
								),
							},
							PlanModifiers: []planmodifier.List{
								listplanmodifier.RequiresReplace(),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"domain_arn": schema.StringAttribute{
										CustomType: fwtypes.ARNType,
										Required:   true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.RequiresReplace(),
										},
									},
									"domain_endpoint": schema.StringAttribute{
										Required: true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.RequiresReplace(),
										},
										Validators: []validator.String{
											stringvalidator.RegexMatches(
												regexache.MustCompile(`^https://.*$`),
												"must be a valid HTTPS URL",
											),
										},
									},
									"vector_index_name": schema.StringAttribute{
										Required: true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.RequiresReplace(),
										},
									},
								},
								Blocks: map[string]schema.Block{
									"field_mapping": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[opensearchManagedClusterFieldMappingModel](ctx),
										Validators: []validator.List{
											listvalidator.IsRequired(),
											listvalidator.SizeAtLeast(1),
											listvalidator.SizeAtMost(1),
										},
										PlanModifiers: []planmodifier.List{
											listplanmodifier.RequiresReplace(),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"metadata_field": schema.StringAttribute{
													Required: true,
													PlanModifiers: []planmodifier.String{
														stringplanmodifier.RequiresReplace(),
													},
												},
												"text_field": schema.StringAttribute{
													Required: true,
													PlanModifiers: []planmodifier.String{
														stringplanmodifier.RequiresReplace(),
													},
												},
												"vector_field": schema.StringAttribute{
													Required: true,
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
						"opensearch_serverless_configuration": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[openSearchServerlessConfigurationModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							PlanModifiers: []planmodifier.List{
								listplanmodifier.RequiresReplace(),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"collection_arn": schema.StringAttribute{
										CustomType: fwtypes.ARNType,
										Required:   true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.RequiresReplace(),
										},
									},
									"vector_index_name": schema.StringAttribute{
										Required: true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.RequiresReplace(),
										},
									},
								},
								Blocks: map[string]schema.Block{
									"field_mapping": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[openSearchServerlessFieldMappingModel](ctx),
										Validators: []validator.List{
											listvalidator.IsRequired(),
											listvalidator.SizeAtLeast(1),
											listvalidator.SizeAtMost(1),
										},
										PlanModifiers: []planmodifier.List{
											listplanmodifier.RequiresReplace(),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"metadata_field": schema.StringAttribute{
													Required: true,
													PlanModifiers: []planmodifier.String{
														stringplanmodifier.RequiresReplace(),
													},
												},
												"text_field": schema.StringAttribute{
													Required: true,
													PlanModifiers: []planmodifier.String{
														stringplanmodifier.RequiresReplace(),
													},
												},
												"vector_field": schema.StringAttribute{
													Required: true,
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
						"pinecone_configuration": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[pineconeConfigurationModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							PlanModifiers: []planmodifier.List{
								listplanmodifier.RequiresReplace(),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"connection_string": schema.StringAttribute{
										Required: true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.RequiresReplace(),
										},
									},
									"credentials_secret_arn": schema.StringAttribute{
										CustomType: fwtypes.ARNType,
										Required:   true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.RequiresReplace(),
										},
									},
									names.AttrNamespace: schema.StringAttribute{
										Optional: true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.RequiresReplace(),
										},
									},
								},
								Blocks: map[string]schema.Block{
									"field_mapping": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[pineconeFieldMappingModel](ctx),
										Validators: []validator.List{
											listvalidator.IsRequired(),
											listvalidator.SizeAtLeast(1),
											listvalidator.SizeAtMost(1),
										},
										PlanModifiers: []planmodifier.List{
											listplanmodifier.RequiresReplace(),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"metadata_field": schema.StringAttribute{
													Required: true,
													PlanModifiers: []planmodifier.String{
														stringplanmodifier.RequiresReplace(),
													},
												},
												"text_field": schema.StringAttribute{
													Required: true,
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
						"rds_configuration": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[rdsConfigurationModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							PlanModifiers: []planmodifier.List{
								listplanmodifier.RequiresReplace(),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"credentials_secret_arn": schema.StringAttribute{
										CustomType: fwtypes.ARNType,
										Required:   true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.RequiresReplace(),
										},
									},
									names.AttrDatabaseName: schema.StringAttribute{
										Required: true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.RequiresReplace(),
										},
									},
									names.AttrResourceARN: schema.StringAttribute{
										CustomType: fwtypes.ARNType,
										Required:   true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.RequiresReplace(),
										},
									},
									names.AttrTableName: schema.StringAttribute{
										Required: true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.RequiresReplace(),
										},
									},
								},
								Blocks: map[string]schema.Block{
									"field_mapping": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[rdsFieldMappingModel](ctx),
										Validators: []validator.List{
											listvalidator.IsRequired(),
											listvalidator.SizeAtLeast(1),
											listvalidator.SizeAtMost(1),
										},
										PlanModifiers: []planmodifier.List{
											listplanmodifier.RequiresReplace(),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"custom_metadata_field": schema.StringAttribute{
													Optional: true,
													PlanModifiers: []planmodifier.String{
														stringplanmodifier.RequiresReplace(),
													},
												},
												"metadata_field": schema.StringAttribute{
													Required: true,
													PlanModifiers: []planmodifier.String{
														stringplanmodifier.RequiresReplace(),
													},
												},
												"primary_key_field": schema.StringAttribute{
													Required: true,
													PlanModifiers: []planmodifier.String{
														stringplanmodifier.RequiresReplace(),
													},
												},
												"text_field": schema.StringAttribute{
													Required: true,
													PlanModifiers: []planmodifier.String{
														stringplanmodifier.RequiresReplace(),
													},
												},
												"vector_field": schema.StringAttribute{
													Required: true,
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
						"redis_enterprise_cloud_configuration": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[redisEnterpriseCloudConfigurationModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							PlanModifiers: []planmodifier.List{
								listplanmodifier.RequiresReplace(),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"credentials_secret_arn": schema.StringAttribute{
										CustomType: fwtypes.ARNType,
										Required:   true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.RequiresReplace(),
										},
									},
									names.AttrEndpoint: schema.StringAttribute{
										Required: true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.RequiresReplace(),
										},
									},
									"vector_index_name": schema.StringAttribute{
										Required: true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.RequiresReplace(),
										},
									},
								},
								Blocks: map[string]schema.Block{
									"field_mapping": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[redisEnterpriseCloudFieldMappingModel](ctx),
										Validators: []validator.List{
											listvalidator.IsRequired(),
											listvalidator.SizeAtLeast(1),
											listvalidator.SizeAtMost(1),
										},
										PlanModifiers: []planmodifier.List{
											listplanmodifier.RequiresReplace(),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"metadata_field": schema.StringAttribute{
													Optional: true,
													PlanModifiers: []planmodifier.String{
														stringplanmodifier.RequiresReplace(),
													},
												},
												"text_field": schema.StringAttribute{
													Optional: true,
													PlanModifiers: []planmodifier.String{
														stringplanmodifier.RequiresReplace(),
													},
												},
												"vector_field": schema.StringAttribute{
													Optional: true,
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
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

func (r *knowledgeBaseResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data knowledgeBaseResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentClient(ctx)

	var input bedrockagent.CreateKnowledgeBaseInput
	response.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	input.ClientToken = aws.String(id.UniqueId())
	input.Tags = getTagsIn(ctx)

	var output *bedrockagent.CreateKnowledgeBaseOutput
	var err error
	err = tfresource.Retry(ctx, propagationTimeout, func(ctx context.Context) *tfresource.RetryError {
		output, err = conn.CreateKnowledgeBase(ctx, &input)

		// IAM propagation
		if tfawserr.ErrMessageContains(err, errCodeValidationException, "cannot assume role") {
			return tfresource.RetryableError(err)
		}
		if tfawserr.ErrMessageContains(err, errCodeValidationException, "unable to assume the given role") {
			return tfresource.RetryableError(err)
		}

		// Kendra access propagation
		if tfawserr.ErrMessageContains(err, errCodeValidationException, "Encountered AccessDeniedException from Kendra") {
			return tfresource.RetryableError(err)
		}

		// OpenSearch data access propagation
		if tfawserr.ErrMessageContains(err, errCodeValidationException, "storage configuration provided is invalid") {
			return tfresource.RetryableError(err)
		}

		if err != nil {
			return tfresource.NonRetryableError(err)
		}

		return nil
	})

	if err != nil {
		response.Diagnostics.AddError("creating Bedrock Agent Knowledge Base", err.Error())
		return
	}

	kb := output.KnowledgeBase
	knowledgeBaseID := aws.ToString(kb.KnowledgeBaseId)
	data.KnowledgeBaseARN = fwflex.StringToFramework(ctx, kb.KnowledgeBaseArn)
	data.KnowledgeBaseID = fwflex.StringValueToFramework(ctx, knowledgeBaseID)

	kb, err = waitKnowledgeBaseCreated(ctx, conn, knowledgeBaseID, r.CreateTimeout(ctx, data.Timeouts))

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for Bedrock Agent Knowledge Base (%s) create", knowledgeBaseID), err.Error())
		return
	}

	// Set values for unknowns after creation is complete.
	data.CreatedAt = fwflex.TimeToFramework(ctx, kb.CreatedAt)
	data.FailureReasons = fwflex.FlattenFrameworkStringValueListOfString(ctx, kb.FailureReasons)
	data.UpdatedAt = fwflex.TimeToFramework(ctx, kb.UpdatedAt)

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *knowledgeBaseResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data knowledgeBaseResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentClient(ctx)

	knowledgeBaseID := fwflex.StringValueFromFramework(ctx, data.KnowledgeBaseID)
	kb, err := findKnowledgeBaseByID(ctx, conn, knowledgeBaseID)

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Bedrock Agent Knowledge Base (%s)", knowledgeBaseID), err.Error())

		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, kb, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *knowledgeBaseResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var old, new knowledgeBaseResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.State.Get(ctx, &old)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentClient(ctx)

	if !new.Description.Equal(old.Description) ||
		!new.Name.Equal(old.Name) ||
		!new.RoleARN.Equal(old.RoleARN) {
		knowledgeBaseID := fwflex.StringValueFromFramework(ctx, new.KnowledgeBaseID)
		var input bedrockagent.UpdateKnowledgeBaseInput
		response.Diagnostics.Append(fwflex.Expand(ctx, new, &input)...)
		if response.Diagnostics.HasError() {
			return
		}

		_, err := tfresource.RetryWhenAWSErrMessageContains(ctx, propagationTimeout, func(ctx context.Context) (any, error) {
			return conn.UpdateKnowledgeBase(ctx, &input)
		}, errCodeValidationException, "cannot assume role")

		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("updating Bedrock Agent Knowledge Base (%s)", knowledgeBaseID), err.Error())

			return
		}

		kb, err := waitKnowledgeBaseUpdated(ctx, conn, knowledgeBaseID, r.UpdateTimeout(ctx, new.Timeouts))

		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("waiting for Bedrock Agent Knowledge Base (%s) update", knowledgeBaseID), err.Error())

			return
		}

		new.FailureReasons = fwflex.FlattenFrameworkStringValueListOfString(ctx, kb.FailureReasons)
		new.UpdatedAt = fwflex.TimeToFramework(ctx, kb.UpdatedAt)
	} else {
		new.FailureReasons = old.FailureReasons
		new.UpdatedAt = old.UpdatedAt
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *knowledgeBaseResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data knowledgeBaseResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentClient(ctx)

	knowledgeBaseID := fwflex.StringValueFromFramework(ctx, data.KnowledgeBaseID)
	input := bedrockagent.DeleteKnowledgeBaseInput{
		KnowledgeBaseId: aws.String(knowledgeBaseID),
	}
	_, err := conn.DeleteKnowledgeBase(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting Bedrock Agent Knowledge Base (%s)", knowledgeBaseID), err.Error())

		return
	}

	if _, err := waitKnowledgeBaseDeleted(ctx, conn, knowledgeBaseID, r.DeleteTimeout(ctx, data.Timeouts)); err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for Bedrock Agent Knowledge Base (%s) delete", knowledgeBaseID), err.Error())

		return
	}
}

func waitKnowledgeBaseCreated(ctx context.Context, conn *bedrockagent.Client, id string, timeout time.Duration) (*awstypes.KnowledgeBase, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.KnowledgeBaseStatusCreating),
		Target:  enum.Slice(awstypes.KnowledgeBaseStatusActive),
		Refresh: statusKnowledgeBase(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.KnowledgeBase); ok {
		tfresource.SetLastError(err, errors.Join(tfslices.ApplyToAll(output.FailureReasons, errors.New)...))

		return output, err
	}

	return nil, err
}

func waitKnowledgeBaseUpdated(ctx context.Context, conn *bedrockagent.Client, id string, timeout time.Duration) (*awstypes.KnowledgeBase, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.KnowledgeBaseStatusUpdating),
		Target:  enum.Slice(awstypes.KnowledgeBaseStatusActive),
		Refresh: statusKnowledgeBase(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.KnowledgeBase); ok {
		tfresource.SetLastError(err, errors.Join(tfslices.ApplyToAll(output.FailureReasons, errors.New)...))

		return output, err
	}

	return nil, err
}

func waitKnowledgeBaseDeleted(ctx context.Context, conn *bedrockagent.Client, id string, timeout time.Duration) (*awstypes.KnowledgeBase, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.KnowledgeBaseStatusActive, awstypes.KnowledgeBaseStatusDeleting),
		Target:  []string{},
		Refresh: statusKnowledgeBase(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.KnowledgeBase); ok {
		tfresource.SetLastError(err, errors.Join(tfslices.ApplyToAll(output.FailureReasons, errors.New)...))

		return output, err
	}

	return nil, err
}

func statusKnowledgeBase(ctx context.Context, conn *bedrockagent.Client, id string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findKnowledgeBaseByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func findKnowledgeBaseByID(ctx context.Context, conn *bedrockagent.Client, id string) (*awstypes.KnowledgeBase, error) {
	input := bedrockagent.GetKnowledgeBaseInput{
		KnowledgeBaseId: aws.String(id),
	}

	return findKnowledgeBase(ctx, conn, &input)
}

func findKnowledgeBase(ctx context.Context, conn *bedrockagent.Client, input *bedrockagent.GetKnowledgeBaseInput) (*awstypes.KnowledgeBase, error) {
	output, err := conn.GetKnowledgeBase(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.KnowledgeBase == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.KnowledgeBase, nil
}

type knowledgeBaseResourceModel struct {
	framework.WithRegionModel
	CreatedAt                  timetypes.RFC3339                                                `tfsdk:"created_at"`
	Description                types.String                                                     `tfsdk:"description"`
	FailureReasons             fwtypes.ListValueOf[types.String]                                `tfsdk:"failure_reasons"`
	KnowledgeBaseARN           types.String                                                     `tfsdk:"arn"`
	KnowledgeBaseConfiguration fwtypes.ListNestedObjectValueOf[knowledgeBaseConfigurationModel] `tfsdk:"knowledge_base_configuration"`
	KnowledgeBaseID            types.String                                                     `tfsdk:"id"`
	Name                       types.String                                                     `tfsdk:"name"`
	RoleARN                    fwtypes.ARN                                                      `tfsdk:"role_arn"`
	StorageConfiguration       fwtypes.ListNestedObjectValueOf[storageConfigurationModel]       `tfsdk:"storage_configuration"`
	Tags                       tftags.Map                                                       `tfsdk:"tags"`
	TagsAll                    tftags.Map                                                       `tfsdk:"tags_all"`
	Timeouts                   timeouts.Value                                                   `tfsdk:"timeouts"`
	UpdatedAt                  timetypes.RFC3339                                                `tfsdk:"updated_at"`
}

type knowledgeBaseConfigurationModel struct {
	KendraKnowledgeBaseConfiguration fwtypes.ListNestedObjectValueOf[kendraKnowledgeBaseConfigurationModel] `tfsdk:"kendra_knowledge_base_configuration"`
	Type                             fwtypes.StringEnum[awstypes.KnowledgeBaseType]                         `tfsdk:"type"`
	VectorKnowledgeBaseConfiguration fwtypes.ListNestedObjectValueOf[vectorKnowledgeBaseConfigurationModel] `tfsdk:"vector_knowledge_base_configuration"`
}

type vectorKnowledgeBaseConfigurationModel struct {
	EmbeddingModelARN                    fwtypes.ARN                                                                `tfsdk:"embedding_model_arn"`
	EmbeddingModelConfiguration          fwtypes.ListNestedObjectValueOf[embeddingModelConfigurationModel]          `tfsdk:"embedding_model_configuration"`
	SupplementalDataStorageConfiguration fwtypes.ListNestedObjectValueOf[supplementalDataStorageConfigurationModel] `tfsdk:"supplemental_data_storage_configuration"`
}

type embeddingModelConfigurationModel struct {
	BedrockEmbeddingModelConfiguration fwtypes.ListNestedObjectValueOf[bedrockEmbeddingModelConfigurationModel] `tfsdk:"bedrock_embedding_model_configuration"`
}

type bedrockEmbeddingModelConfigurationModel struct {
	Dimensions        types.Int64                                    `tfsdk:"dimensions"`
	EmbeddingDataType fwtypes.StringEnum[awstypes.EmbeddingDataType] `tfsdk:"embedding_data_type"`
}

type supplementalDataStorageConfigurationModel struct {
	StorageLocations fwtypes.ListNestedObjectValueOf[storageLocationModel] `tfsdk:"storage_location"`
}

type storageLocationModel struct {
	S3Location fwtypes.ListNestedObjectValueOf[s3LocationModel]                 `tfsdk:"s3_location"`
	Type       fwtypes.StringEnum[awstypes.SupplementalDataStorageLocationType] `tfsdk:"type"`
}

type kendraKnowledgeBaseConfigurationModel struct {
	KendraIndexARN fwtypes.ARN `tfsdk:"kendra_index_arn"`
}

type storageConfigurationModel struct {
	OpensearchManagedClusterConfiguration fwtypes.ListNestedObjectValueOf[openSearchManagedClusterConfigurationModel] `tfsdk:"opensearch_managed_cluster_configuration"`
	OpensearchServerlessConfiguration     fwtypes.ListNestedObjectValueOf[openSearchServerlessConfigurationModel]     `tfsdk:"opensearch_serverless_configuration"`
	PineconeConfiguration                 fwtypes.ListNestedObjectValueOf[pineconeConfigurationModel]                 `tfsdk:"pinecone_configuration"`
	RDSConfiguration                      fwtypes.ListNestedObjectValueOf[rdsConfigurationModel]                      `tfsdk:"rds_configuration"`
	RedisEnterpriseCloudConfiguration     fwtypes.ListNestedObjectValueOf[redisEnterpriseCloudConfigurationModel]     `tfsdk:"redis_enterprise_cloud_configuration"`
	Type                                  fwtypes.StringEnum[awstypes.KnowledgeBaseStorageType]                       `tfsdk:"type"`
}

type openSearchManagedClusterConfigurationModel struct {
	DomainARN       fwtypes.ARN                                                                `tfsdk:"domain_arn"`
	DomainEndpoint  types.String                                                               `tfsdk:"domain_endpoint"`
	FieldMapping    fwtypes.ListNestedObjectValueOf[opensearchManagedClusterFieldMappingModel] `tfsdk:"field_mapping"`
	VectorIndexName types.String                                                               `tfsdk:"vector_index_name"`
}

type opensearchManagedClusterFieldMappingModel struct {
	MetadataField types.String `tfsdk:"metadata_field"`
	TextField     types.String `tfsdk:"text_field"`
	VectorField   types.String `tfsdk:"vector_field"`
}

type openSearchServerlessConfigurationModel struct {
	CollectionARN   fwtypes.ARN                                                            `tfsdk:"collection_arn"`
	FieldMapping    fwtypes.ListNestedObjectValueOf[openSearchServerlessFieldMappingModel] `tfsdk:"field_mapping"`
	VectorIndexName types.String                                                           `tfsdk:"vector_index_name"`
}

type openSearchServerlessFieldMappingModel struct {
	MetadataField types.String `tfsdk:"metadata_field"`
	TextField     types.String `tfsdk:"text_field"`
	VectorField   types.String `tfsdk:"vector_field"`
}

type pineconeConfigurationModel struct {
	ConnectionString     types.String                                               `tfsdk:"connection_string"`
	CredentialsSecretARN fwtypes.ARN                                                `tfsdk:"credentials_secret_arn"`
	FieldMapping         fwtypes.ListNestedObjectValueOf[pineconeFieldMappingModel] `tfsdk:"field_mapping"`
	Namespace            types.String                                               `tfsdk:"namespace"`
}

type pineconeFieldMappingModel struct {
	MetadataField types.String `tfsdk:"metadata_field"`
	TextField     types.String `tfsdk:"text_field"`
}

type rdsConfigurationModel struct {
	CredentialsSecretARN fwtypes.ARN                                           `tfsdk:"credentials_secret_arn"`
	DatabaseName         types.String                                          `tfsdk:"database_name"`
	FieldMapping         fwtypes.ListNestedObjectValueOf[rdsFieldMappingModel] `tfsdk:"field_mapping"`
	ResourceARN          fwtypes.ARN                                           `tfsdk:"resource_arn"`
	TableName            types.String                                          `tfsdk:"table_name"`
}

type rdsFieldMappingModel struct {
	CustomMetadataField types.String `tfsdk:"custom_metadata_field"`
	MetadataField       types.String `tfsdk:"metadata_field"`
	PrimaryKeyField     types.String `tfsdk:"primary_key_field"`
	TextField           types.String `tfsdk:"text_field"`
	VectorField         types.String `tfsdk:"vector_field"`
}

type redisEnterpriseCloudConfigurationModel struct {
	CredentialsSecretARN fwtypes.ARN                                                            `tfsdk:"credentials_secret_arn"`
	Endpoint             types.String                                                           `tfsdk:"endpoint"`
	FieldMapping         fwtypes.ListNestedObjectValueOf[redisEnterpriseCloudFieldMappingModel] `tfsdk:"field_mapping"`
	VectorIndexName      types.String                                                           `tfsdk:"vector_index_name"`
}

type redisEnterpriseCloudFieldMappingModel struct {
	MetadataField types.String `tfsdk:"metadata_field"`
	TextField     types.String `tfsdk:"text_field"`
	VectorField   types.String `tfsdk:"vector_field"`
}
