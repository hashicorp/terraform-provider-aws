// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package sagemaker

import (
	"context"
	"fmt"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sagemaker"
	awstypes "github.com/aws/aws-sdk-go-v2/service/sagemaker/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/mapvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	fwvalidators "github.com/hashicorp/terraform-provider-aws/internal/framework/validators"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_sagemaker_algorithm", name="Algorithm")
// @Tags(identifierAttribute="arn")
// @IdentityAttribute("algorithm_name")
// @Testing(hasNoPreExistingResource=true)
// @Testing(importStateIdAttribute="algorithm_name")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/sagemaker;sagemaker.DescribeAlgorithmOutput")
func newAlgorithmResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &algorithmResource{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameAlgorithm = "Algorithm"
)

type algorithmResource struct {
	framework.ResourceWithModel[algorithmResourceModel]
	framework.WithNoUpdate
	framework.WithTimeouts
	framework.WithImportByIdentity
}

func (r *algorithmResource) Schema(ctx context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"algorithm_description": schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.LengthAtMost(1024),
					stringvalidator.RegexMatches(regexache.MustCompile(`[\p{L}\p{M}\p{Z}\p{S}\p{N}\p{P}]*`), "algorithm description must contain only letters, marks, spaces, symbols, numbers, and punctuation"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"algorithm_name": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 63),
					stringvalidator.RegexMatches(regexache.MustCompile(`^[a-zA-Z0-9](-*[a-zA-Z0-9])*$`), "algorithm name must start with an alphanumeric character and contain only alphanumeric characters and hyphens"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"algorithm_status": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.AlgorithmStatus](),
				Computed:   true,
			},
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			"certify_for_marketplace": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
					boolplanmodifier.RequiresReplace(),
				},
			},
			"creation_time": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			"product_id": schema.StringAttribute{
				Computed: true,
			},
			names.AttrTags:    tftags.TagsAttributeForceNew(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			"inference_specification": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[inferenceSpecificationModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"supported_content_types": schema.ListAttribute{
							CustomType:  fwtypes.ListOfStringType,
							ElementType: types.StringType,
							Optional:    true,
							Validators: []validator.List{
								listvalidator.ValueStringsAre(
									stringvalidator.LengthAtMost(256),
								),
							},
							PlanModifiers: []planmodifier.List{
								listplanmodifier.RequiresReplace(),
							},
						},
						"supported_realtime_inference_instance_types": schema.ListAttribute{
							CustomType:  fwtypes.ListOfStringEnumType[awstypes.ProductionVariantInstanceType](),
							ElementType: types.StringType,
							Optional:    true,
							PlanModifiers: []planmodifier.List{
								listplanmodifier.RequiresReplace(),
							},
						},
						"supported_response_mime_types": schema.ListAttribute{
							CustomType:  fwtypes.ListOfStringType,
							ElementType: types.StringType,
							Optional:    true,
							Validators: []validator.List{
								listvalidator.ValueStringsAre(
									stringvalidator.LengthAtMost(1024),
									stringvalidator.RegexMatches(regexache.MustCompile(`[-\w]+/.+`), "supported response MIME types must be in type/subtype format"),
								),
							},
							PlanModifiers: []planmodifier.List{
								listplanmodifier.RequiresReplace(),
							},
						},
						"supported_transform_instance_types": schema.ListAttribute{
							CustomType:  fwtypes.ListOfStringEnumType[awstypes.TransformInstanceType](),
							ElementType: types.StringType,
							Optional:    true,
							Validators: []validator.List{
								listvalidator.SizeAtLeast(1),
							},
							PlanModifiers: []planmodifier.List{
								listplanmodifier.RequiresReplace(),
							},
						},
					},
					Blocks: map[string]schema.Block{
						"containers": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[modelPackageContainerDefinitionModel](ctx),
							Validators: []validator.List{
								listvalidator.IsRequired(),
								listvalidator.SizeAtLeast(1),
							},
							PlanModifiers: []planmodifier.List{
								listplanmodifier.RequiresReplace(),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"container_hostname": schema.StringAttribute{
										Optional: true,
										Validators: []validator.String{
											stringvalidator.RegexMatches(regexache.MustCompile(`[a-zA-Z0-9](-*[a-zA-Z0-9]){0,62}`), "container hostname must start with an alphanumeric character and contain only alphanumeric characters and hyphens"),
										},
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.RequiresReplace(),
										},
									},
									"environment": schema.MapAttribute{
										CustomType:  fwtypes.MapOfStringType,
										ElementType: types.StringType,
										Optional:    true,
										Validators: []validator.Map{
											mapvalidator.SizeAtMost(100),
											mapvalidator.KeysAre(
												stringvalidator.LengthAtMost(1024),
												stringvalidator.RegexMatches(regexache.MustCompile(`[a-zA-Z_][a-zA-Z0-9_]*`), "environment keys must start with a letter or underscore and contain only letters, numbers, and underscores"),
											),
											mapvalidator.ValueStringsAre(
												stringvalidator.LengthAtMost(1024),
												stringvalidator.RegexMatches(regexache.MustCompile(`[\S\s]*`), "environment values may contain any characters"),
											),
										},
										PlanModifiers: []planmodifier.Map{
											mapplanmodifier.RequiresReplace(),
										},
									},
									"framework": schema.StringAttribute{
										Optional: true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.RequiresReplace(),
										},
									},
									"framework_version": schema.StringAttribute{
										Optional: true,
										Validators: []validator.String{
											stringvalidator.LengthBetween(3, 10),
											stringvalidator.RegexMatches(regexache.MustCompile(`[0-9]\.[A-Za-z0-9.-]+`), "framework version must start with a digit, followed by a period, and contain only alphanumeric characters, dots, and hyphens"),
										},
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.RequiresReplace(),
										},
									},
									"image": schema.StringAttribute{
										Optional: true,
										Validators: []validator.String{
											stringvalidator.LengthAtMost(255),
											stringvalidator.RegexMatches(regexache.MustCompile(`[\S]+`), "image must not contain whitespace"),
										},
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.RequiresReplace(),
										},
									},
									"image_digest": schema.StringAttribute{
										Optional: true,
										Computed: true,
										Validators: []validator.String{
											stringvalidator.LengthAtMost(72),
											stringvalidator.RegexMatches(regexache.MustCompile(`[Ss][Hh][Aa]256:[0-9a-fA-F]{64}`), "image digest must be a valid sha256 digest"),
										},
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.UseStateForUnknown(),
											stringplanmodifier.RequiresReplace(),
										},
									},
									"is_checkpoint": schema.BoolAttribute{
										Optional: true,
										Computed: true,
										PlanModifiers: []planmodifier.Bool{
											boolplanmodifier.RequiresReplace(),
										},
									},
									"model_data_etag": schema.StringAttribute{
										Optional: true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.RequiresReplace(),
										},
									},
									"model_data_url": schema.StringAttribute{
										Optional: true,
										Validators: []validator.String{
											stringvalidator.LengthAtMost(1024),
											stringvalidator.RegexMatches(httpsOrS3URIRegexp, "model data URL must be HTTPS or Amazon S3 URI"),
										},
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.RequiresReplace(),
										},
									},
									"nearest_model_name": schema.StringAttribute{
										Optional: true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.RequiresReplace(),
										},
									},
									"product_id": schema.StringAttribute{
										Optional: true,
										Validators: []validator.String{
											stringvalidator.LengthAtMost(256),
											stringvalidator.RegexMatches(regexache.MustCompile(`[a-zA-Z0-9](-*[a-zA-Z0-9])*`), "product ID must start with an alphanumeric character and contain only alphanumeric characters and hyphens"),
										},
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.RequiresReplace(),
										},
									},
								},
								Blocks: map[string]schema.Block{
									"additional_s3_data_source": additionalS3DataSourceBlock(ctx),
									"base_model":                baseModelBlock(ctx),
									"model_data_source":         modelDataSourceBlock(ctx),
									"model_input":               modelInputBlock(ctx),
								},
							},
						},
					},
				},
			},
			"training_specification": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[trainingSpecificationModel](ctx),
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
						"supported_training_instance_types": schema.ListAttribute{
							CustomType:  fwtypes.ListOfStringEnumType[awstypes.TrainingInstanceType](),
							ElementType: types.StringType,
							Required:    true,
							PlanModifiers: []planmodifier.List{
								listplanmodifier.RequiresReplace(),
							},
						},
						"supports_distributed_training": schema.BoolAttribute{
							Optional: true,
							Computed: true,
							PlanModifiers: []planmodifier.Bool{
								boolplanmodifier.UseStateForUnknown(),
								boolplanmodifier.RequiresReplace(),
							},
						},
						"training_image": schema.StringAttribute{
							Required: true,
							Validators: []validator.String{
								stringvalidator.LengthAtMost(255),
								stringvalidator.RegexMatches(regexache.MustCompile(`[\S]+`), "training image must not contain whitespace"),
							},
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
						"training_image_digest": schema.StringAttribute{
							Optional: true,
							Computed: true,
							Validators: []validator.String{
								stringvalidator.LengthAtMost(72),
								stringvalidator.RegexMatches(regexache.MustCompile(`[Ss][Hh][Aa]256:[0-9a-fA-F]{64}`), "training image digest must be a valid sha256 digest"),
							},
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
								stringplanmodifier.RequiresReplace(),
							},
						},
					},
					Blocks: map[string]schema.Block{
						"additional_s3_data_source":              additionalS3DataSourceBlock(ctx),
						"metric_definitions":                     metricDefinitionsBlock(ctx),
						"supported_hyper_parameters":             supportedHyperParametersBlock(ctx),
						"supported_tuning_job_objective_metrics": supportedTuningJobObjectiveMetricsBlock(ctx),
						"training_channels":                      trainingChannelsBlock(ctx),
					},
				},
			},
			"validation_specification": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[algorithmValidationSpecificationModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"validation_role": schema.StringAttribute{
							Required: true,
							Validators: []validator.String{
								stringvalidator.LengthBetween(20, 2048),
								stringvalidator.RegexMatches(regexache.MustCompile(`arn:aws[a-z\-]*:iam::\d{12}:role/?[a-zA-Z_0-9+=,.@\-_/]+`), "validation role must be a valid IAM role ARN"),
							},
							PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
						},
					},
					Blocks: map[string]schema.Block{
						"validation_profiles": validationProfilesBlock(ctx),
					},
				},
			},
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{Create: true, Delete: true}),
		},
	}
}

func additionalS3DataSourceBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[additionalS3DataSourceModel](ctx),
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		PlanModifiers: []planmodifier.List{
			listplanmodifier.RequiresReplace(),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"compression_type": schema.StringAttribute{
					CustomType: fwtypes.StringEnumType[awstypes.CompressionType](),
					Optional:   true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
				"etag": schema.StringAttribute{
					Optional: true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
				"s3_data_type": schema.StringAttribute{
					CustomType: fwtypes.StringEnumType[awstypes.AdditionalS3DataSourceDataType](),
					Required:   true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
				"s3_uri": schema.StringAttribute{
					Required: true,
					Validators: []validator.String{
						stringvalidator.LengthAtMost(1024),
						stringvalidator.RegexMatches(httpsOrS3URIRegexp, "S3 URI must be HTTPS or Amazon S3 URI"),
					},
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
			},
		},
	}
}

func baseModelBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[baseModelModel](ctx),
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		PlanModifiers: []planmodifier.List{
			listplanmodifier.RequiresReplace(),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"hub_content_name": schema.StringAttribute{
					Optional: true,
					Validators: []validator.String{
						stringvalidator.LengthAtMost(63),
						stringvalidator.RegexMatches(regexache.MustCompile(`[a-zA-Z0-9](-*[a-zA-Z0-9]){0,62}`), "hub content name must start with an alphanumeric character and contain only alphanumeric characters and hyphens"),
					},
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
				"hub_content_version": schema.StringAttribute{
					Optional: true,
					Validators: []validator.String{
						stringvalidator.LengthBetween(5, 14),
						stringvalidator.RegexMatches(regexache.MustCompile(`\d{1,4}\.\d{1,4}\.\d{1,4}`), "hub content version must be in major.minor.patch format with 1 to 4 digits per segment"),
					},
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
				"recipe_name": schema.StringAttribute{
					Optional: true,
					Validators: []validator.String{
						stringvalidator.LengthAtMost(255),
					},
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
			},
		},
	}
}

func modelDataSourceBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[modelDataSourceModel](ctx),
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		PlanModifiers: []planmodifier.List{
			listplanmodifier.RequiresReplace(),
		},
		NestedObject: schema.NestedBlockObject{
			Blocks: map[string]schema.Block{
				"s3_data_source": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[modelDataSourceS3DataSourceModel](ctx),
					Validators: []validator.List{
						listvalidator.SizeAtMost(1),
					},
					PlanModifiers: []planmodifier.List{
						listplanmodifier.RequiresReplace(),
					},
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"compression_type": schema.StringAttribute{
								CustomType: fwtypes.StringEnumType[awstypes.ModelCompressionType](),
								Required:   true,
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.RequiresReplace(),
								},
							},
							"etag": schema.StringAttribute{
								Optional: true,
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.RequiresReplace(),
								},
							},
							"manifest_etag": schema.StringAttribute{
								Optional: true,
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.RequiresReplace(),
								},
							},
							"manifest_s3_uri": schema.StringAttribute{
								Optional: true,
								Validators: []validator.String{
									stringvalidator.LengthAtMost(1024),
									stringvalidator.RegexMatches(httpsOrS3URIRegexp, "manifest S3 URI must be HTTPS or Amazon S3 URI"),
								},
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.RequiresReplace(),
								},
							},
							"s3_data_type": schema.StringAttribute{
								CustomType: fwtypes.StringEnumType[awstypes.S3ModelDataType](),
								Required:   true,
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.RequiresReplace(),
								},
							},
							"s3_uri": schema.StringAttribute{
								Required: true,
								Validators: []validator.String{
									stringvalidator.LengthAtMost(1024),
									stringvalidator.RegexMatches(httpsOrS3URIRegexp, "S3 URI must be HTTPS or Amazon S3 URI"),
								},
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.RequiresReplace(),
								},
							},
						},
						Blocks: map[string]schema.Block{
							"hub_access_config":   hubAccessConfigBlock(ctx),
							"model_access_config": modelAccessConfigBlock(ctx),
						},
					},
				},
			},
		},
	}
}

func hubAccessConfigBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[hubAccessConfigModel](ctx),
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		PlanModifiers: []planmodifier.List{
			listplanmodifier.RequiresReplace(),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"hub_content_arn": schema.StringAttribute{
					Optional: true,
					Validators: []validator.String{
						fwvalidators.ARN(),
					},
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
			},
		},
	}
}

func modelAccessConfigBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[modelAccessConfigModel](ctx),
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		PlanModifiers: []planmodifier.List{
			listplanmodifier.RequiresReplace(),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"accept_eula": schema.BoolAttribute{
					Optional: true,
					PlanModifiers: []planmodifier.Bool{
						boolplanmodifier.RequiresReplace(),
					},
				},
			},
		},
	}
}

func modelInputBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[modelInputModel](ctx),
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		PlanModifiers: []planmodifier.List{
			listplanmodifier.RequiresReplace(),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"data_input_config": schema.StringAttribute{
					Optional: true,
					Validators: []validator.String{
						stringvalidator.LengthBetween(1, 63),
					},
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
			},
		},
	}
}

func metricDefinitionsBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[metricDefinitionModel](ctx),
		Validators: []validator.List{
			listvalidator.SizeAtMost(40),
		},
		PlanModifiers: []planmodifier.List{listplanmodifier.RequiresReplace()},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"name": schema.StringAttribute{
					Required: true,
					Validators: []validator.String{
						stringvalidator.LengthBetween(1, 255),
						stringvalidator.RegexMatches(regexache.MustCompile(`.+`), "metric name must not be empty"),
					},
					PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
				},
				"regex": schema.StringAttribute{
					Required: true,
					Validators: []validator.String{
						stringvalidator.LengthBetween(1, 500),
						stringvalidator.RegexMatches(regexache.MustCompile(`.+`), "metric regex must not be empty"),
					},
					PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
				},
			},
		},
	}
}

func supportedHyperParametersBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[hyperParameterSpecificationModel](ctx),
		Validators: []validator.List{
			listvalidator.SizeAtMost(100),
		},
		PlanModifiers: []planmodifier.List{listplanmodifier.RequiresReplace()},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"default_value": schema.StringAttribute{
					Optional: true,
					Validators: []validator.String{
						stringvalidator.LengthAtMost(2500),
					},
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
				"description": schema.StringAttribute{
					Optional: true,
					Validators: []validator.String{
						stringvalidator.LengthAtMost(1024),
						stringvalidator.RegexMatches(regexache.MustCompile(`[\p{L}\p{M}\p{Z}\p{S}\p{N}\p{P}]*`), "description must contain only letters, marks, spaces, symbols, numbers, and punctuation"),
					},
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
				"is_required": schema.BoolAttribute{
					Optional: true,
					Computed: true,
					PlanModifiers: []planmodifier.Bool{
						boolplanmodifier.UseStateForUnknown(),
						boolplanmodifier.RequiresReplace(),
					},
				},
				"is_tunable": schema.BoolAttribute{
					Optional: true,
					Computed: true,
					PlanModifiers: []planmodifier.Bool{
						boolplanmodifier.UseStateForUnknown(),
						boolplanmodifier.RequiresReplace(),
					},
				},
				"name": schema.StringAttribute{
					Required: true,
					Validators: []validator.String{
						stringvalidator.LengthAtMost(256),
						stringvalidator.RegexMatches(regexache.MustCompile(`[\p{L}\p{M}\p{Z}\p{S}\p{N}\p{P}]*`), "name must contain only letters, marks, spaces, symbols, numbers, and punctuation"),
					},
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
				"type": schema.StringAttribute{
					CustomType: fwtypes.StringEnumType[awstypes.ParameterType](),
					Required:   true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
			},
			Blocks: map[string]schema.Block{
				"range": parameterRangeBlock(ctx),
			},
		},
	}
}

func parameterRangeBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[parameterRangeModel](ctx),
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		PlanModifiers: []planmodifier.List{
			listplanmodifier.RequiresReplace(),
		},
		NestedObject: schema.NestedBlockObject{
			Blocks: map[string]schema.Block{
				"categorical_parameter_range_specification": categoricalParameterRangeSpecificationBlock(ctx),
				"continuous_parameter_range_specification":  continuousParameterRangeSpecificationBlock(ctx),
				"integer_parameter_range_specification":     integerParameterRangeSpecificationBlock(ctx),
			},
		},
	}
}

func categoricalParameterRangeSpecificationBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[categoricalParameterRangeSpecificationModel](ctx),
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		PlanModifiers: []planmodifier.List{
			listplanmodifier.RequiresReplace(),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"values": schema.ListAttribute{
					CustomType:  fwtypes.ListOfStringType,
					ElementType: types.StringType,
					Required:    true,
					Validators: []validator.List{
						listvalidator.SizeAtLeast(1),
						listvalidator.SizeAtMost(30),
						listvalidator.ValueStringsAre(
							stringvalidator.LengthAtMost(256),
						),
					},
					PlanModifiers: []planmodifier.List{
						listplanmodifier.RequiresReplace(),
					},
				},
			},
		},
	}
}

func continuousParameterRangeSpecificationBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[continuousParameterRangeSpecificationModel](ctx),
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		PlanModifiers: []planmodifier.List{
			listplanmodifier.RequiresReplace(),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"max_value": schema.StringAttribute{
					Required: true,
					Validators: []validator.String{
						stringvalidator.LengthAtMost(256),
					},
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
				"min_value": schema.StringAttribute{
					Required: true,
					Validators: []validator.String{
						stringvalidator.LengthAtMost(256),
					},
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
			},
		},
	}
}

func integerParameterRangeSpecificationBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[integerParameterRangeSpecificationModel](ctx),
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		PlanModifiers: []planmodifier.List{
			listplanmodifier.RequiresReplace(),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"max_value": schema.StringAttribute{
					Required: true,
					Validators: []validator.String{
						stringvalidator.LengthAtMost(256),
					},
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
				"min_value": schema.StringAttribute{
					Required: true,
					Validators: []validator.String{
						stringvalidator.LengthAtMost(256),
					},
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
			},
		},
	}
}

func supportedTuningJobObjectiveMetricsBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[hyperParameterTuningJobObjectiveModel](ctx),
		PlanModifiers: []planmodifier.List{
			listplanmodifier.RequiresReplace(),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"metric_name": schema.StringAttribute{
					Required: true,
					Validators: []validator.String{
						stringvalidator.LengthBetween(1, 255),
					},
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
				"type": schema.StringAttribute{
					CustomType: fwtypes.StringEnumType[awstypes.HyperParameterTuningJobObjectiveType](),
					Required:   true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
			},
		},
	}
}

func trainingChannelsBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[channelSpecificationModel](ctx),
		Validators: []validator.List{
			listvalidator.IsRequired(),
			listvalidator.SizeAtLeast(1),
			listvalidator.SizeAtMost(8),
		},
		PlanModifiers: []planmodifier.List{
			listplanmodifier.RequiresReplace(),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"description": schema.StringAttribute{
					Optional: true,
					Validators: []validator.String{
						stringvalidator.LengthAtMost(1024),
						stringvalidator.RegexMatches(regexache.MustCompile(`[\p{L}\p{M}\p{Z}\p{S}\p{N}\p{P}]*`), "description must contain only letters, marks, spaces, symbols, numbers, and punctuation"),
					},
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
				"is_required": schema.BoolAttribute{
					Optional: true,
					Computed: true,
					PlanModifiers: []planmodifier.Bool{
						boolplanmodifier.UseStateForUnknown(),
						boolplanmodifier.RequiresReplace(),
					},
				},
				"name": schema.StringAttribute{
					Required: true,
					Validators: []validator.String{
						stringvalidator.LengthBetween(1, 64),
						stringvalidator.RegexMatches(regexache.MustCompile(`[A-Za-z0-9\.\-_]+`), "name must contain only letters, numbers, periods, hyphens, and underscores"),
					},
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
				"supported_compression_types": schema.ListAttribute{
					CustomType:  fwtypes.ListOfStringEnumType[awstypes.CompressionType](),
					ElementType: types.StringType,
					Optional:    true,
					PlanModifiers: []planmodifier.List{
						listplanmodifier.RequiresReplace(),
					},
				},
				"supported_content_types": schema.ListAttribute{
					CustomType:  fwtypes.ListOfStringType,
					ElementType: types.StringType,
					Required:    true,
					Validators: []validator.List{
						listvalidator.ValueStringsAre(
							stringvalidator.LengthAtMost(256),
						),
					},
					PlanModifiers: []planmodifier.List{
						listplanmodifier.RequiresReplace(),
					},
				},
				"supported_input_modes": schema.ListAttribute{
					CustomType:  fwtypes.ListOfStringEnumType[awstypes.TrainingInputMode](),
					ElementType: types.StringType,
					Required:    true,
					Validators: []validator.List{
						listvalidator.SizeAtLeast(1),
					},
					PlanModifiers: []planmodifier.List{
						listplanmodifier.RequiresReplace(),
					},
				},
			},
		},
	}
}

func validationProfilesBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[algorithmValidationProfileModel](ctx),
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
				"profile_name": schema.StringAttribute{
					Required: true,
					Validators: []validator.String{
						stringvalidator.LengthBetween(1, 63),
						stringvalidator.RegexMatches(regexache.MustCompile(`[a-zA-Z0-9](-*[a-zA-Z0-9]){0,62}`), "profile name must start with an alphanumeric character and contain only alphanumeric characters and hyphens"),
					},
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
			},
			Blocks: map[string]schema.Block{
				"training_job_definition":  trainingJobDefinitionBlock(ctx),
				"transform_job_definition": transformJobDefinitionBlock(ctx),
			},
		},
	}
}

func trainingJobDefinitionBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[trainingJobDefinitionModel](ctx),
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
				"hyper_parameters": schema.MapAttribute{
					CustomType:  fwtypes.MapOfStringType,
					ElementType: types.StringType,
					Optional:    true,
					Validators: []validator.Map{
						mapvalidator.SizeAtMost(100),
						mapvalidator.KeysAre(
							stringvalidator.LengthAtMost(256),
						),
						mapvalidator.ValueStringsAre(
							stringvalidator.LengthAtMost(2500),
						),
					},
					PlanModifiers: []planmodifier.Map{
						mapplanmodifier.RequiresReplace(),
					},
				},
				"training_input_mode": schema.StringAttribute{
					CustomType: fwtypes.StringEnumType[awstypes.TrainingInputMode](),
					Required:   true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
			},
			Blocks: map[string]schema.Block{
				"input_data_config":  inputDataConfigBlock(ctx),
				"output_data_config": outputDataConfigBlock(ctx),
				"resource_config":    resourceConfigBlock(ctx),
				"stopping_condition": stoppingConditionBlock(ctx),
			},
		},
	}
}

func inputDataConfigBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[channelModel](ctx),
		Validators: []validator.List{
			listvalidator.IsRequired(),
			listvalidator.SizeAtLeast(1),
		},
		PlanModifiers: []planmodifier.List{
			listplanmodifier.RequiresReplace(),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"channel_name": schema.StringAttribute{
					Required: true,
					Validators: []validator.String{
						stringvalidator.LengthBetween(1, 64),
						stringvalidator.RegexMatches(regexache.MustCompile(`[A-Za-z0-9\.\-_]+`), "channel name must contain only letters, numbers, periods, hyphens, and underscores"),
					},
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
				"compression_type": schema.StringAttribute{
					CustomType: fwtypes.StringEnumType[awstypes.CompressionType](),
					Optional:   true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
				"content_type": schema.StringAttribute{
					Optional: true,
					Validators: []validator.String{
						stringvalidator.LengthAtMost(256),
					},
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
				"input_mode": schema.StringAttribute{
					CustomType: fwtypes.StringEnumType[awstypes.TrainingInputMode](),
					Optional:   true,
					Computed:   true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
				"record_wrapper_type": schema.StringAttribute{
					CustomType: fwtypes.StringEnumType[awstypes.RecordWrapper](),
					Optional:   true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
			},
			Blocks: map[string]schema.Block{
				"data_source":    dataSourceBlock(ctx),
				"shuffle_config": shuffleConfigBlock(ctx),
			},
		},
	}
}

func dataSourceBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[dataSourceModel](ctx),
		Validators: []validator.List{
			listvalidator.IsRequired(),
			listvalidator.SizeAtMost(1),
			listvalidator.SizeAtLeast(1),
		},
		PlanModifiers: []planmodifier.List{
			listplanmodifier.RequiresReplace(),
		},
		NestedObject: schema.NestedBlockObject{
			Blocks: map[string]schema.Block{
				"file_system_data_source": fileSystemDataSourceBlock(ctx),
				"s3_data_source":          trainingS3DataSourceBlock(ctx),
			},
		},
	}
}

func fileSystemDataSourceBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[fileSystemDataSourceModel](ctx),
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		PlanModifiers: []planmodifier.List{
			listplanmodifier.RequiresReplace(),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"directory_path": schema.StringAttribute{
					Required: true,
					Validators: []validator.String{
						stringvalidator.LengthAtMost(4096),
					},
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
				"file_system_access_mode": schema.StringAttribute{
					CustomType: fwtypes.StringEnumType[awstypes.FileSystemAccessMode](),
					Required:   true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
				"file_system_id": schema.StringAttribute{
					Required: true,
					Validators: []validator.String{
						stringvalidator.LengthBetween(11, 21),
						stringvalidator.RegexMatches(regexache.MustCompile(`^(fs-[0-9a-f]{8,})$`), "file system ID must match the required pattern"),
					},
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
				"file_system_type": schema.StringAttribute{
					CustomType: fwtypes.StringEnumType[awstypes.FileSystemType](),
					Required:   true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
			},
		},
	}
}

func trainingS3DataSourceBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[trainingS3DataSourceModel](ctx),
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		PlanModifiers: []planmodifier.List{
			listplanmodifier.RequiresReplace(),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"attribute_names": schema.ListAttribute{
					CustomType:  fwtypes.ListOfStringType,
					ElementType: types.StringType,
					Optional:    true,
					Validators: []validator.List{
						listvalidator.SizeAtMost(16),
						listvalidator.ValueStringsAre(
							stringvalidator.LengthBetween(1, 256),
							stringvalidator.RegexMatches(regexache.MustCompile(`.+`), "attribute names must not be empty"),
						),
					},
					PlanModifiers: []planmodifier.List{
						listplanmodifier.RequiresReplace(),
					},
				},
				"instance_group_names": schema.ListAttribute{
					CustomType:  fwtypes.ListOfStringType,
					ElementType: types.StringType,
					Optional:    true,
					Validators: []validator.List{
						listvalidator.SizeAtMost(5),
						listvalidator.ValueStringsAre(
							stringvalidator.LengthBetween(1, 64),
						),
					},
					PlanModifiers: []planmodifier.List{
						listplanmodifier.RequiresReplace(),
					},
				},
				"s3_data_distribution_type": schema.StringAttribute{
					CustomType: fwtypes.StringEnumType[awstypes.S3DataDistribution](),
					Optional:   true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
				"s3_data_type": schema.StringAttribute{
					CustomType: fwtypes.StringEnumType[awstypes.S3DataType](),
					Required:   true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
				"s3_uri": schema.StringAttribute{
					Required: true,
					Validators: []validator.String{
						stringvalidator.LengthAtMost(1024),
						stringvalidator.RegexMatches(httpsOrS3URIRegexp, "S3 URI must be HTTPS or Amazon S3 URI"),
					},
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
			},
			Blocks: map[string]schema.Block{
				"hub_access_config":   hubAccessConfigBlock(ctx),
				"model_access_config": modelAccessConfigBlock(ctx),
			},
		},
	}
}

func shuffleConfigBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[shuffleConfigModel](ctx),
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		PlanModifiers: []planmodifier.List{
			listplanmodifier.RequiresReplace(),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"seed": schema.Int64Attribute{
					Required: true,
					Validators: []validator.Int64{
						int64validator.AtLeast(0),
					},
					PlanModifiers: []planmodifier.Int64{
						int64planmodifier.RequiresReplace(),
					},
				},
			},
		},
	}
}

func outputDataConfigBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[outputDataConfigModel](ctx),
		Validators: []validator.List{
			listvalidator.IsRequired(),
			listvalidator.SizeAtMost(1),
			listvalidator.SizeAtLeast(1),
		},
		PlanModifiers: []planmodifier.List{
			listplanmodifier.RequiresReplace(),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"compression_type": schema.StringAttribute{
					CustomType: fwtypes.StringEnumType[awstypes.OutputCompressionType](),
					Optional:   true,
					Computed:   true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
				"kms_key_id": schema.StringAttribute{
					Optional: true,
					Validators: []validator.String{
						stringvalidator.LengthAtMost(2048),
						stringvalidator.RegexMatches(regexache.MustCompile(`[a-zA-Z0-9:/_-]*`), "KMS key ID must contain only letters, numbers, colons, slashes, underscores, and hyphens"),
					},
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
				"s3_output_path": schema.StringAttribute{
					Required: true,
					Validators: []validator.String{
						stringvalidator.LengthAtMost(1024),
						stringvalidator.RegexMatches(httpsOrS3URIRegexp, "S3 output path must be HTTPS or Amazon S3 URI"),
					},
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
			},
		},
	}
}

func resourceConfigBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[resourceConfigModel](ctx),
		Validators: []validator.List{
			listvalidator.IsRequired(),
			listvalidator.SizeAtMost(1),
			listvalidator.SizeAtLeast(1),
		},
		PlanModifiers: []planmodifier.List{
			listplanmodifier.RequiresReplace(),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"instance_count": schema.Int32Attribute{
					Optional: true,
					Validators: []validator.Int32{
						int32validator.AtLeast(0),
					},
					PlanModifiers: []planmodifier.Int32{
						int32planmodifier.RequiresReplace(),
					},
				},
				"instance_type": schema.StringAttribute{
					CustomType: fwtypes.StringEnumType[awstypes.TrainingInstanceType](),
					Optional:   true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
				"keep_alive_period_in_seconds": schema.Int32Attribute{
					Optional: true,
					Computed: true,
					Validators: []validator.Int32{
						int32validator.Between(0, 3600),
					},
					PlanModifiers: []planmodifier.Int32{
						int32planmodifier.RequiresReplace(),
					},
				},
				"training_plan_arn": schema.StringAttribute{
					Optional: true,
					Validators: []validator.String{
						stringvalidator.LengthBetween(50, 2048),
						stringvalidator.RegexMatches(regexache.MustCompile(`arn:aws[a-z\-]*:sagemaker:[a-z0-9\-]*:[0-9]{12}:training-plan/.*`), "training plan ARN must match the required SageMaker training plan ARN pattern"),
					},
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
				"volume_kms_key_id": schema.StringAttribute{
					Optional: true,
					Validators: []validator.String{
						stringvalidator.LengthAtMost(2048),
						stringvalidator.RegexMatches(regexache.MustCompile(`[a-zA-Z0-9:/_-]*`), "volume KMS key ID must contain only letters, numbers, colons, slashes, underscores, and hyphens"),
					},
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
				"volume_size_in_gb": schema.Int32Attribute{
					Optional: true,
					Validators: []validator.Int32{
						int32validator.AtLeast(0),
					},
					PlanModifiers: []planmodifier.Int32{
						int32planmodifier.RequiresReplace(),
					},
				},
			},
			Blocks: map[string]schema.Block{
				"instance_groups":           instanceGroupsBlock(ctx),
				"instance_placement_config": instancePlacementConfigBlock(ctx),
			},
		},
	}
}

func instanceGroupsBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[instanceGroupModel](ctx),
		Validators: []validator.List{
			listvalidator.SizeAtMost(5),
		},
		PlanModifiers: []planmodifier.List{
			listplanmodifier.RequiresReplace(),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"instance_count": schema.Int32Attribute{
					Required: true,
					Validators: []validator.Int32{
						int32validator.AtLeast(0),
					},
					PlanModifiers: []planmodifier.Int32{
						int32planmodifier.RequiresReplace(),
					},
				},
				"instance_group_name": schema.StringAttribute{
					Required: true,
					Validators: []validator.String{
						stringvalidator.LengthBetween(1, 64),
					},
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
				"instance_type": schema.StringAttribute{
					CustomType: fwtypes.StringEnumType[awstypes.TrainingInstanceType](),
					Required:   true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
			},
		},
	}
}

func instancePlacementConfigBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[instancePlacementConfigModel](ctx),
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		PlanModifiers: []planmodifier.List{
			listplanmodifier.RequiresReplace(),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"enable_multiple_jobs": schema.BoolAttribute{
					Optional: true,
					PlanModifiers: []planmodifier.Bool{
						boolplanmodifier.RequiresReplace(),
					},
				},
			},
			Blocks: map[string]schema.Block{
				"placement_specifications": placementSpecificationsBlock(ctx),
			},
		},
	}
}

func placementSpecificationsBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[placementSpecificationModel](ctx),
		Validators: []validator.List{
			listvalidator.SizeAtMost(10),
		},
		PlanModifiers: []planmodifier.List{listplanmodifier.RequiresReplace()},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"instance_count": schema.Int32Attribute{
					Required: true,
					Validators: []validator.Int32{
						int32validator.AtLeast(0),
					},
					PlanModifiers: []planmodifier.Int32{int32planmodifier.RequiresReplace()},
				},
				"ultra_server_id": schema.StringAttribute{
					Optional: true,
					Validators: []validator.String{
						stringvalidator.LengthAtMost(256),
					},
					PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
				},
			},
		},
	}
}

func stoppingConditionBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[stoppingConditionModel](ctx),
		Validators: []validator.List{
			listvalidator.IsRequired(),
			listvalidator.SizeAtMost(1),
			listvalidator.SizeAtLeast(1),
		},
		PlanModifiers: []planmodifier.List{
			listplanmodifier.RequiresReplace()},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"max_pending_time_in_seconds": schema.Int32Attribute{
					Optional: true,
					Computed: true,
					Validators: []validator.Int32{
						int32validator.Between(7200, 2419200),
					},
					PlanModifiers: []planmodifier.Int32{
						int32planmodifier.RequiresReplace(),
					},
				},
				"max_runtime_in_seconds": schema.Int32Attribute{
					Optional: true,
					Validators: []validator.Int32{
						int32validator.AtLeast(1),
					},
					PlanModifiers: []planmodifier.Int32{
						int32planmodifier.RequiresReplace(),
					},
				},
				"max_wait_time_in_seconds": schema.Int32Attribute{
					Optional: true,
					Computed: true,
					Validators: []validator.Int32{
						int32validator.AtLeast(1),
					},
					PlanModifiers: []planmodifier.Int32{
						int32planmodifier.RequiresReplace(),
					},
				},
			},
		},
	}
}

func transformJobDefinitionBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[transformJobDefinitionModel](ctx),
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		PlanModifiers: []planmodifier.List{
			listplanmodifier.RequiresReplace(),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"batch_strategy": schema.StringAttribute{
					CustomType: fwtypes.StringEnumType[awstypes.BatchStrategy](),
					Optional:   true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
				"environment": schema.MapAttribute{
					CustomType:  fwtypes.MapOfStringType,
					ElementType: types.StringType,
					Optional:    true,
					Validators: []validator.Map{
						mapvalidator.SizeAtMost(16),
						mapvalidator.KeysAre(
							stringvalidator.LengthAtMost(1024),
							stringvalidator.RegexMatches(regexache.MustCompile(`[a-zA-Z_][a-zA-Z0-9_]`), "environment keys must start with a letter or underscore and contain only letters, numbers, and underscores"),
						),
						mapvalidator.ValueStringsAre(
							stringvalidator.LengthAtMost(10240),
						),
					},
					PlanModifiers: []planmodifier.Map{
						mapplanmodifier.RequiresReplace(),
					},
				},
				"max_concurrent_transforms": schema.Int32Attribute{
					Optional: true,
					Validators: []validator.Int32{
						int32validator.AtLeast(0),
					},
					PlanModifiers: []planmodifier.Int32{
						int32planmodifier.RequiresReplace(),
					},
				},
				"max_payload_in_mb": schema.Int32Attribute{
					Optional: true,
					Validators: []validator.Int32{
						int32validator.AtLeast(0),
					},
					PlanModifiers: []planmodifier.Int32{
						int32planmodifier.RequiresReplace(),
					},
				},
			},
			Blocks: map[string]schema.Block{
				"transform_input":     transformInputBlock(ctx),
				"transform_output":    transformOutputBlock(ctx),
				"transform_resources": transformResourcesBlock(ctx),
			},
		},
	}
}

func transformInputBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[transformInputModel](ctx),
		Validators: []validator.List{
			listvalidator.IsRequired(),
			listvalidator.SizeAtMost(1),
			listvalidator.SizeAtLeast(1),
		},
		PlanModifiers: []planmodifier.List{
			listplanmodifier.RequiresReplace(),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"compression_type": schema.StringAttribute{
					CustomType: fwtypes.StringEnumType[awstypes.CompressionType](),
					Optional:   true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
				"content_type": schema.StringAttribute{
					Optional: true,
					Validators: []validator.String{
						stringvalidator.LengthAtMost(256),
					},
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
				"split_type": schema.StringAttribute{
					CustomType: fwtypes.StringEnumType[awstypes.SplitType](),
					Optional:   true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
			},
			Blocks: map[string]schema.Block{
				"data_source": transformJobDataSourceBlock(ctx),
			},
		},
	}
}

func transformJobDataSourceBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[transformJobDataSourceModel](ctx),
		Validators: []validator.List{
			listvalidator.IsRequired(),
			listvalidator.SizeAtMost(1),
			listvalidator.SizeAtLeast(1),
		},
		PlanModifiers: []planmodifier.List{
			listplanmodifier.RequiresReplace(),
		},
		NestedObject: schema.NestedBlockObject{
			Blocks: map[string]schema.Block{
				"s3_data_source": transformS3DataSourceBlock(ctx),
			},
		},
	}
}

func transformS3DataSourceBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[transformS3DataSourceModel](ctx),
		Validators: []validator.List{
			listvalidator.IsRequired(),
			listvalidator.SizeAtMost(1),
			listvalidator.SizeAtLeast(1),
		},
		PlanModifiers: []planmodifier.List{
			listplanmodifier.RequiresReplace(),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"s3_data_type": schema.StringAttribute{
					CustomType: fwtypes.StringEnumType[awstypes.S3DataType](),
					Required:   true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
				"s3_uri": schema.StringAttribute{
					Required: true,
					Validators: []validator.String{
						stringvalidator.LengthAtMost(1024),
						stringvalidator.RegexMatches(httpsOrS3URIRegexp, "S3 URI must be HTTPS or Amazon S3 URI"),
					},
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
			},
		},
	}
}

func transformOutputBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[transformOutputModel](ctx),
		Validators: []validator.List{
			listvalidator.IsRequired(),
			listvalidator.SizeAtMost(1),
			listvalidator.SizeAtLeast(1),
		},
		PlanModifiers: []planmodifier.List{
			listplanmodifier.RequiresReplace(),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"accept": schema.StringAttribute{
					Optional: true,
					Validators: []validator.String{
						stringvalidator.LengthAtMost(256),
					},
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
				"assemble_with": schema.StringAttribute{
					CustomType: fwtypes.StringEnumType[awstypes.AssemblyType](),
					Optional:   true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
				"kms_key_id": schema.StringAttribute{
					Optional: true,
					Validators: []validator.String{
						stringvalidator.LengthAtMost(2048),
						stringvalidator.RegexMatches(regexache.MustCompile(`[a-zA-Z0-9:/_-]*`), "KMS key ID must contain only letters, numbers, colons, slashes, underscores, and hyphens"),
					},
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
				"s3_output_path": schema.StringAttribute{
					Required: true,
					Validators: []validator.String{
						stringvalidator.LengthAtMost(1024),
						stringvalidator.RegexMatches(httpsOrS3URIRegexp, "S3 output path must be HTTPS or Amazon S3 URI"),
					},
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
			},
		},
	}
}

func transformResourcesBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[transformResourcesModel](ctx),
		Validators: []validator.List{
			listvalidator.IsRequired(),
			listvalidator.SizeAtMost(1),
			listvalidator.SizeAtLeast(1),
		},
		PlanModifiers: []planmodifier.List{
			listplanmodifier.RequiresReplace(),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"instance_count": schema.Int32Attribute{
					Required: true,
					Validators: []validator.Int32{
						int32validator.Between(1, 100),
					},
					PlanModifiers: []planmodifier.Int32{
						int32planmodifier.RequiresReplace(),
					},
				},
				"instance_type": schema.StringAttribute{
					CustomType: fwtypes.StringEnumType[awstypes.TransformInstanceType](),
					Required:   true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
				"transform_ami_version": schema.StringAttribute{
					Optional: true,
					Validators: []validator.String{
						stringvalidator.LengthBetween(1, 63),
						stringvalidator.RegexMatches(regexache.MustCompile(`[a-zA-Z0-9]+(-[a-zA-Z0-9]+)*`), "transform AMI version must contain only alphanumeric characters and single hyphen-separated segments"),
					},
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
				"volume_kms_key_id": schema.StringAttribute{
					Optional: true,
					Validators: []validator.String{
						stringvalidator.LengthAtMost(2048),
						stringvalidator.RegexMatches(regexache.MustCompile(`[a-zA-Z0-9:/_-]*`), "volume KMS key ID must contain only letters, numbers, colons, slashes, underscores, and hyphens"),
					},
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
			},
		},
	}
}

func (r *algorithmResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data algorithmResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}
	plan := data

	conn := r.Meta().SageMakerClient(ctx)

	var input sagemaker.CreateAlgorithmInput
	response.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)
	if response.Diagnostics.HasError() {
		return
	}
	input.Tags = getTagsIn(ctx)

	_, err := conn.CreateAlgorithm(ctx, &input)
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating SageMaker Algorithm (%s)", data.AlgorithmName.ValueString()), err.Error())
		return
	}

	outputWait, err := waitAlgorithmCreated(ctx, conn, data.AlgorithmName.ValueString(), r.CreateTimeout(ctx, data.Timeouts))
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for SageMaker Algorithm (%s) create", data.AlgorithmName.ValueString()), err.Error())
		return
	}

	response.Diagnostics.Append(r.flatten(ctx, outputWait, &data, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *algorithmResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data algorithmResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}
	state := data

	conn := r.Meta().SageMakerClient(ctx)

	output, err := findAlgorithmByName(ctx, conn, data.AlgorithmName.ValueString())
	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading SageMaker Algorithm (%s)", data.AlgorithmName.ValueString()), err.Error())
		return
	}

	response.Diagnostics.Append(r.flatten(ctx, output, &data, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *algorithmResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data algorithmResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().SageMakerClient(ctx)
	deleteTimeout := r.DeleteTimeout(ctx, data.Timeouts)
	input := sagemaker.DeleteAlgorithmInput{
		AlgorithmName: data.AlgorithmName.ValueStringPointer(),
	}

	_, err := tfresource.RetryWhenAWSErrMessageContains(ctx, deleteTimeout,
		func(ctx context.Context) (any, error) {
			return conn.DeleteAlgorithm(ctx, &input)
		},
		ErrCodeValidationException, "Cannot delete Algorithm in InProgress status")

	if errs.IsA[*awstypes.ResourceNotFound](err) {
		return
	}
	if tfawserr.ErrMessageContains(err, ErrCodeValidationException, "does not exist") {
		return
	}
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting SageMaker Algorithm (%s)", data.AlgorithmName.ValueString()), err.Error())
		return
	}

	if err := waitAlgorithmDeleted(ctx, conn, data.AlgorithmName.ValueString(), deleteTimeout); err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for SageMaker Algorithm (%s) delete", data.AlgorithmName.ValueString()), err.Error())
	}
}

func (r *algorithmResource) flatten(ctx context.Context, apiObject any, data, prior *algorithmResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	diags.Append(fwflex.Flatten(ctx, apiObject, data, fwflex.WithFieldNamePrefix("Algorithm"))...)
	if diags.HasError() {
		return diags
	}

	diags.Append(preserveAlgorithmValidationSpecification(ctx, data, prior)...)

	return diags
}

// AWS does not seem to return :
// 1. validation_specification.validation_profiles.training_job_definition.resource_config.keep_alive_period_in_seconds
// 2. validation_specification.validation_profiles.training_job_definition.stopping_condition.max_pending_time_in_seconds
// 3. validation_specification.validation_profiles.training_job_definition.stopping_condition.max_wait_time_in_seconds
// 4. validation_specification.validation_profiles.training_job_definition.input_data_config.shuffle_config (block)
// 5. validation_specification.validation_profiles.training_job_definition.output_data_config.compression_type
func preserveAlgorithmValidationSpecification(ctx context.Context, current, prior *algorithmResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	if current == nil || prior == nil {
		return diags
	}

	currentValidation, d := current.ValidationSpecification.ToPtr(ctx)
	diags.Append(d...)
	priorValidation, d := prior.ValidationSpecification.ToPtr(ctx)
	diags.Append(d...)
	if diags.HasError() || currentValidation == nil || priorValidation == nil {
		return diags
	}

	currentProfile, d := currentValidation.ValidationProfiles.ToPtr(ctx)
	diags.Append(d...)
	priorProfile, d := priorValidation.ValidationProfiles.ToPtr(ctx)
	diags.Append(d...)
	if diags.HasError() || currentProfile == nil || priorProfile == nil {
		return diags
	}

	currentTraining, d := currentProfile.TrainingJobDefinition.ToPtr(ctx)
	diags.Append(d...)
	priorTraining, d := priorProfile.TrainingJobDefinition.ToPtr(ctx)
	diags.Append(d...)
	if diags.HasError() || currentTraining == nil || priorTraining == nil {
		return diags
	}

	currentInputs, d := currentTraining.InputDataConfig.ToSlice(ctx)
	diags.Append(d...)
	priorInputs, d := priorTraining.InputDataConfig.ToSlice(ctx)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}

	if len(currentInputs) > 0 && len(priorInputs) > 0 && currentInputs[0] != nil && priorInputs[0] != nil {
		currentInputs[0].InputMode = priorInputs[0].InputMode
		currentInputs[0].ShuffleConfig = priorInputs[0].ShuffleConfig

		currentTraining.InputDataConfig, d = fwtypes.NewListNestedObjectValueOfSlice(ctx, currentInputs, nil)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}
	}

	currentOutput, d := currentTraining.OutputDataConfig.ToPtr(ctx)
	diags.Append(d...)
	priorOutput, d := priorTraining.OutputDataConfig.ToPtr(ctx)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}
	if currentOutput != nil && priorOutput != nil {
		currentOutput.CompressionType = priorOutput.CompressionType
		currentTraining.OutputDataConfig, d = fwtypes.NewListNestedObjectValueOfPtr(ctx, currentOutput)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}
	}

	currentResource, d := currentTraining.ResourceConfig.ToPtr(ctx)
	diags.Append(d...)
	priorResource, d := priorTraining.ResourceConfig.ToPtr(ctx)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}
	if currentResource != nil && priorResource != nil {
		currentResource.KeepAlivePeriodInSeconds = priorResource.KeepAlivePeriodInSeconds
		currentTraining.ResourceConfig, d = fwtypes.NewListNestedObjectValueOfPtr(ctx, currentResource)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}
	}

	currentStopping, d := currentTraining.StoppingCondition.ToPtr(ctx)
	diags.Append(d...)
	priorStopping, d := priorTraining.StoppingCondition.ToPtr(ctx)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}
	if currentStopping != nil && priorStopping != nil {
		currentStopping.MaxPendingTimeInSeconds = priorStopping.MaxPendingTimeInSeconds
		currentStopping.MaxWaitTimeInSeconds = priorStopping.MaxWaitTimeInSeconds

		currentTraining.StoppingCondition, d = fwtypes.NewListNestedObjectValueOfPtr(ctx, currentStopping)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}
	}

	currentProfile.TrainingJobDefinition, d = fwtypes.NewListNestedObjectValueOfPtr(ctx, currentTraining)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}

	currentValidation.ValidationProfiles, d = fwtypes.NewListNestedObjectValueOfPtr(ctx, currentProfile)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}

	current.ValidationSpecification, d = fwtypes.NewListNestedObjectValueOfPtr(ctx, currentValidation)
	diags.Append(d...)

	return diags
}

func findAlgorithmByName(ctx context.Context, conn *sagemaker.Client, name string) (*sagemaker.DescribeAlgorithmOutput, error) {
	input := sagemaker.DescribeAlgorithmInput{
		AlgorithmName: aws.String(name),
	}

	output, err := conn.DescribeAlgorithm(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFound](err) || tfawserr.ErrMessageContains(err, "ValidationException", "does not exist") {
		return nil, &retry.NotFoundError{LastError: err}
	}
	if err != nil {
		return nil, err
	}
	if output == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output, nil
}

type algorithmResourceModel struct {
	framework.WithRegionModel
	AlgorithmDescription    types.String                                                           `tfsdk:"algorithm_description"`
	AlgorithmName           types.String                                                           `tfsdk:"algorithm_name"`
	AlgorithmStatus         fwtypes.StringEnum[awstypes.AlgorithmStatus]                           `tfsdk:"algorithm_status"`
	ARN                     types.String                                                           `tfsdk:"arn"`
	CertifyForMarketplace   types.Bool                                                             `tfsdk:"certify_for_marketplace"`
	CreationTime            timetypes.RFC3339                                                      `tfsdk:"creation_time"`
	InferenceSpecification  fwtypes.ListNestedObjectValueOf[inferenceSpecificationModel]           `tfsdk:"inference_specification"`
	ProductID               types.String                                                           `tfsdk:"product_id"`
	Tags                    tftags.Map                                                             `tfsdk:"tags"`
	TagsAll                 tftags.Map                                                             `tfsdk:"tags_all"`
	Timeouts                timeouts.Value                                                         `tfsdk:"timeouts"`
	TrainingSpecification   fwtypes.ListNestedObjectValueOf[trainingSpecificationModel]            `tfsdk:"training_specification"`
	ValidationSpecification fwtypes.ListNestedObjectValueOf[algorithmValidationSpecificationModel] `tfsdk:"validation_specification"`
}

type inferenceSpecificationModel struct {
	Containers                              fwtypes.ListNestedObjectValueOf[modelPackageContainerDefinitionModel] `tfsdk:"containers"`
	SupportedContentTypes                   fwtypes.ListOfString                                                  `tfsdk:"supported_content_types"`
	SupportedRealtimeInferenceInstanceTypes fwtypes.ListOfStringEnum[awstypes.ProductionVariantInstanceType]      `tfsdk:"supported_realtime_inference_instance_types"`
	SupportedResponseMIMETypes              fwtypes.ListOfString                                                  `tfsdk:"supported_response_mime_types"`
	SupportedTransformInstanceTypes         fwtypes.ListOfStringEnum[awstypes.TransformInstanceType]              `tfsdk:"supported_transform_instance_types"`
}

type modelPackageContainerDefinitionModel struct {
	AdditionalS3DataSource fwtypes.ListNestedObjectValueOf[additionalS3DataSourceModel] `tfsdk:"additional_s3_data_source"`
	BaseModel              fwtypes.ListNestedObjectValueOf[baseModelModel]              `tfsdk:"base_model"`
	ContainerHostname      types.String                                                 `tfsdk:"container_hostname"`
	Environment            fwtypes.MapOfString                                          `tfsdk:"environment"`
	Framework              types.String                                                 `tfsdk:"framework"`
	FrameworkVersion       types.String                                                 `tfsdk:"framework_version"`
	Image                  types.String                                                 `tfsdk:"image"`
	ImageDigest            types.String                                                 `tfsdk:"image_digest"`
	IsCheckpoint           types.Bool                                                   `tfsdk:"is_checkpoint"`
	ModelDataETag          types.String                                                 `tfsdk:"model_data_etag"`
	ModelDataSource        fwtypes.ListNestedObjectValueOf[modelDataSourceModel]        `tfsdk:"model_data_source"`
	ModelDataURL           types.String                                                 `tfsdk:"model_data_url"`
	ModelInput             fwtypes.ListNestedObjectValueOf[modelInputModel]             `tfsdk:"model_input"`
	NearestModelName       types.String                                                 `tfsdk:"nearest_model_name"`
	ProductID              types.String                                                 `tfsdk:"product_id"`
}

type additionalS3DataSourceModel struct {
	CompressionType fwtypes.StringEnum[awstypes.CompressionType]                `tfsdk:"compression_type"`
	ETag            types.String                                                `tfsdk:"etag"`
	S3DataType      fwtypes.StringEnum[awstypes.AdditionalS3DataSourceDataType] `tfsdk:"s3_data_type"`
	S3URI           types.String                                                `tfsdk:"s3_uri"`
}

type baseModelModel struct {
	HubContentName    types.String `tfsdk:"hub_content_name"`
	HubContentVersion types.String `tfsdk:"hub_content_version"`
	RecipeName        types.String `tfsdk:"recipe_name"`
}

type modelDataSourceModel struct {
	S3DataSource fwtypes.ListNestedObjectValueOf[modelDataSourceS3DataSourceModel] `tfsdk:"s3_data_source"`
}

type modelDataSourceS3DataSourceModel struct {
	CompressionType   fwtypes.StringEnum[awstypes.ModelCompressionType]       `tfsdk:"compression_type"`
	ETag              types.String                                            `tfsdk:"etag"`
	HubAccessConfig   fwtypes.ListNestedObjectValueOf[hubAccessConfigModel]   `tfsdk:"hub_access_config"`
	ManifestETag      types.String                                            `tfsdk:"manifest_etag"`
	ManifestS3URI     types.String                                            `tfsdk:"manifest_s3_uri"`
	ModelAccessConfig fwtypes.ListNestedObjectValueOf[modelAccessConfigModel] `tfsdk:"model_access_config"`
	S3DataType        fwtypes.StringEnum[awstypes.S3ModelDataType]            `tfsdk:"s3_data_type"`
	S3URI             types.String                                            `tfsdk:"s3_uri"`
}

type hubAccessConfigModel struct {
	HubContentARN types.String `tfsdk:"hub_content_arn"`
}

type modelAccessConfigModel struct {
	AcceptEula types.Bool `tfsdk:"accept_eula"`
}

type modelInputModel struct {
	DataInputConfig types.String `tfsdk:"data_input_config"`
}

type trainingSpecificationModel struct {
	AdditionalS3DataSource             fwtypes.ListNestedObjectValueOf[additionalS3DataSourceModel]           `tfsdk:"additional_s3_data_source"`
	MetricDefinitions                  fwtypes.ListNestedObjectValueOf[metricDefinitionModel]                 `tfsdk:"metric_definitions"`
	SupportedHyperParameters           fwtypes.ListNestedObjectValueOf[hyperParameterSpecificationModel]      `tfsdk:"supported_hyper_parameters"`
	SupportedTrainingInstanceTypes     fwtypes.ListOfStringEnum[awstypes.TrainingInstanceType]                `tfsdk:"supported_training_instance_types"`
	SupportedTuningJobObjectiveMetrics fwtypes.ListNestedObjectValueOf[hyperParameterTuningJobObjectiveModel] `tfsdk:"supported_tuning_job_objective_metrics"`
	SupportsDistributedTraining        types.Bool                                                             `tfsdk:"supports_distributed_training"`
	TrainingChannels                   fwtypes.ListNestedObjectValueOf[channelSpecificationModel]             `tfsdk:"training_channels"`
	TrainingImage                      types.String                                                           `tfsdk:"training_image"`
	TrainingImageDigest                types.String                                                           `tfsdk:"training_image_digest"`
}

type metricDefinitionModel struct {
	Name  types.String `tfsdk:"name"`
	Regex types.String `tfsdk:"regex"`
}

type hyperParameterSpecificationModel struct {
	DefaultValue types.String                                         `tfsdk:"default_value"`
	Description  types.String                                         `tfsdk:"description"`
	IsRequired   types.Bool                                           `tfsdk:"is_required"`
	IsTunable    types.Bool                                           `tfsdk:"is_tunable"`
	Name         types.String                                         `tfsdk:"name"`
	Range        fwtypes.ListNestedObjectValueOf[parameterRangeModel] `tfsdk:"range"`
	Type         fwtypes.StringEnum[awstypes.ParameterType]           `tfsdk:"type"`
}

type parameterRangeModel struct {
	CategoricalParameterRangeSpecification fwtypes.ListNestedObjectValueOf[categoricalParameterRangeSpecificationModel] `tfsdk:"categorical_parameter_range_specification"`
	ContinuousParameterRangeSpecification  fwtypes.ListNestedObjectValueOf[continuousParameterRangeSpecificationModel]  `tfsdk:"continuous_parameter_range_specification"`
	IntegerParameterRangeSpecification     fwtypes.ListNestedObjectValueOf[integerParameterRangeSpecificationModel]     `tfsdk:"integer_parameter_range_specification"`
}

type categoricalParameterRangeSpecificationModel struct {
	Values fwtypes.ListOfString `tfsdk:"values"`
}

type continuousParameterRangeSpecificationModel struct {
	MaxValue types.String `tfsdk:"max_value"`
	MinValue types.String `tfsdk:"min_value"`
}

type integerParameterRangeSpecificationModel struct {
	MaxValue types.String `tfsdk:"max_value"`
	MinValue types.String `tfsdk:"min_value"`
}

type hyperParameterTuningJobObjectiveModel struct {
	MetricName types.String                                                      `tfsdk:"metric_name"`
	Type       fwtypes.StringEnum[awstypes.HyperParameterTuningJobObjectiveType] `tfsdk:"type"`
}

type channelSpecificationModel struct {
	Description               types.String                                         `tfsdk:"description"`
	IsRequired                types.Bool                                           `tfsdk:"is_required"`
	Name                      types.String                                         `tfsdk:"name"`
	SupportedCompressionTypes fwtypes.ListOfStringEnum[awstypes.CompressionType]   `tfsdk:"supported_compression_types"`
	SupportedContentTypes     fwtypes.ListOfString                                 `tfsdk:"supported_content_types"`
	SupportedInputModes       fwtypes.ListOfStringEnum[awstypes.TrainingInputMode] `tfsdk:"supported_input_modes"`
}

type algorithmValidationSpecificationModel struct {
	ValidationProfiles fwtypes.ListNestedObjectValueOf[algorithmValidationProfileModel] `tfsdk:"validation_profiles"`
	ValidationRole     types.String                                                     `tfsdk:"validation_role"`
}

type algorithmValidationProfileModel struct {
	ProfileName            types.String                                                 `tfsdk:"profile_name"`
	TrainingJobDefinition  fwtypes.ListNestedObjectValueOf[trainingJobDefinitionModel]  `tfsdk:"training_job_definition"`
	TransformJobDefinition fwtypes.ListNestedObjectValueOf[transformJobDefinitionModel] `tfsdk:"transform_job_definition"`
}

type trainingJobDefinitionModel struct {
	HyperParameters   fwtypes.MapOfString                                     `tfsdk:"hyper_parameters"`
	InputDataConfig   fwtypes.ListNestedObjectValueOf[channelModel]           `tfsdk:"input_data_config"`
	OutputDataConfig  fwtypes.ListNestedObjectValueOf[outputDataConfigModel]  `tfsdk:"output_data_config"`
	ResourceConfig    fwtypes.ListNestedObjectValueOf[resourceConfigModel]    `tfsdk:"resource_config"`
	StoppingCondition fwtypes.ListNestedObjectValueOf[stoppingConditionModel] `tfsdk:"stopping_condition"`
	TrainingInputMode fwtypes.StringEnum[awstypes.TrainingInputMode]          `tfsdk:"training_input_mode"`
}

type channelModel struct {
	ChannelName       types.String                                        `tfsdk:"channel_name"`
	CompressionType   fwtypes.StringEnum[awstypes.CompressionType]        `tfsdk:"compression_type"`
	ContentType       types.String                                        `tfsdk:"content_type"`
	DataSource        fwtypes.ListNestedObjectValueOf[dataSourceModel]    `tfsdk:"data_source"`
	InputMode         fwtypes.StringEnum[awstypes.TrainingInputMode]      `tfsdk:"input_mode"`
	RecordWrapperType fwtypes.StringEnum[awstypes.RecordWrapper]          `tfsdk:"record_wrapper_type"`
	ShuffleConfig     fwtypes.ListNestedObjectValueOf[shuffleConfigModel] `tfsdk:"shuffle_config"`
}

type dataSourceModel struct {
	FileSystemDataSource fwtypes.ListNestedObjectValueOf[fileSystemDataSourceModel] `tfsdk:"file_system_data_source"`
	S3DataSource         fwtypes.ListNestedObjectValueOf[trainingS3DataSourceModel] `tfsdk:"s3_data_source"`
}

type fileSystemDataSourceModel struct {
	DirectoryPath        types.String                                      `tfsdk:"directory_path"`
	FileSystemAccessMode fwtypes.StringEnum[awstypes.FileSystemAccessMode] `tfsdk:"file_system_access_mode"`
	FileSystemID         types.String                                      `tfsdk:"file_system_id"`
	FileSystemType       fwtypes.StringEnum[awstypes.FileSystemType]       `tfsdk:"file_system_type"`
}

type trainingS3DataSourceModel struct {
	AttributeNames         fwtypes.ListOfString                                    `tfsdk:"attribute_names"`
	HubAccessConfig        fwtypes.ListNestedObjectValueOf[hubAccessConfigModel]   `tfsdk:"hub_access_config"`
	InstanceGroupNames     fwtypes.ListOfString                                    `tfsdk:"instance_group_names"`
	ModelAccessConfig      fwtypes.ListNestedObjectValueOf[modelAccessConfigModel] `tfsdk:"model_access_config"`
	S3DataDistributionType fwtypes.StringEnum[awstypes.S3DataDistribution]         `tfsdk:"s3_data_distribution_type"`
	S3DataType             fwtypes.StringEnum[awstypes.S3DataType]                 `tfsdk:"s3_data_type"`
	S3URI                  types.String                                            `tfsdk:"s3_uri"`
}

type shuffleConfigModel struct {
	Seed types.Int64 `tfsdk:"seed"`
}

type outputDataConfigModel struct {
	CompressionType fwtypes.StringEnum[awstypes.OutputCompressionType] `tfsdk:"compression_type"`
	KMSKeyID        types.String                                       `tfsdk:"kms_key_id"`
	S3OutputPath    types.String                                       `tfsdk:"s3_output_path"`
}

type resourceConfigModel struct {
	InstanceCount            types.Int32                                                   `tfsdk:"instance_count"`
	InstanceGroups           fwtypes.ListNestedObjectValueOf[instanceGroupModel]           `tfsdk:"instance_groups"`
	InstancePlacementConfig  fwtypes.ListNestedObjectValueOf[instancePlacementConfigModel] `tfsdk:"instance_placement_config"`
	InstanceType             fwtypes.StringEnum[awstypes.TrainingInstanceType]             `tfsdk:"instance_type"`
	KeepAlivePeriodInSeconds types.Int32                                                   `tfsdk:"keep_alive_period_in_seconds"`
	TrainingPlanARN          types.String                                                  `tfsdk:"training_plan_arn"`
	VolumeKMSKeyID           types.String                                                  `tfsdk:"volume_kms_key_id"`
	VolumeSizeInGB           types.Int32                                                   `tfsdk:"volume_size_in_gb"`
}

type instanceGroupModel struct {
	InstanceCount     types.Int32                                       `tfsdk:"instance_count"`
	InstanceGroupName types.String                                      `tfsdk:"instance_group_name"`
	InstanceType      fwtypes.StringEnum[awstypes.TrainingInstanceType] `tfsdk:"instance_type"`
}

type instancePlacementConfigModel struct {
	EnableMultipleJobs      types.Bool                                                   `tfsdk:"enable_multiple_jobs"`
	PlacementSpecifications fwtypes.ListNestedObjectValueOf[placementSpecificationModel] `tfsdk:"placement_specifications"`
}

type placementSpecificationModel struct {
	InstanceCount types.Int32  `tfsdk:"instance_count"`
	UltraServerID types.String `tfsdk:"ultra_server_id"`
}

type stoppingConditionModel struct {
	MaxPendingTimeInSeconds types.Int32 `tfsdk:"max_pending_time_in_seconds"`
	MaxRuntimeInSeconds     types.Int32 `tfsdk:"max_runtime_in_seconds"`
	MaxWaitTimeInSeconds    types.Int32 `tfsdk:"max_wait_time_in_seconds"`
}

type transformJobDefinitionModel struct {
	BatchStrategy           fwtypes.StringEnum[awstypes.BatchStrategy]               `tfsdk:"batch_strategy"`
	Environment             fwtypes.MapOfString                                      `tfsdk:"environment"`
	MaxConcurrentTransforms types.Int32                                              `tfsdk:"max_concurrent_transforms"`
	MaxPayloadInMB          types.Int32                                              `tfsdk:"max_payload_in_mb"`
	TransformInput          fwtypes.ListNestedObjectValueOf[transformInputModel]     `tfsdk:"transform_input"`
	TransformOutput         fwtypes.ListNestedObjectValueOf[transformOutputModel]    `tfsdk:"transform_output"`
	TransformResources      fwtypes.ListNestedObjectValueOf[transformResourcesModel] `tfsdk:"transform_resources"`
}

type transformInputModel struct {
	CompressionType fwtypes.StringEnum[awstypes.CompressionType]                 `tfsdk:"compression_type"`
	ContentType     types.String                                                 `tfsdk:"content_type"`
	DataSource      fwtypes.ListNestedObjectValueOf[transformJobDataSourceModel] `tfsdk:"data_source"`
	SplitType       fwtypes.StringEnum[awstypes.SplitType]                       `tfsdk:"split_type"`
}

type transformJobDataSourceModel struct {
	S3DataSource fwtypes.ListNestedObjectValueOf[transformS3DataSourceModel] `tfsdk:"s3_data_source"`
}

type transformS3DataSourceModel struct {
	S3DataType fwtypes.StringEnum[awstypes.S3DataType] `tfsdk:"s3_data_type"`
	S3URI      types.String                            `tfsdk:"s3_uri"`
}

type transformOutputModel struct {
	Accept       types.String                              `tfsdk:"accept"`
	AssembleWith fwtypes.StringEnum[awstypes.AssemblyType] `tfsdk:"assemble_with"`
	KMSKeyID     types.String                              `tfsdk:"kms_key_id"`
	S3OutputPath types.String                              `tfsdk:"s3_output_path"`
}

type transformResourcesModel struct {
	InstanceCount       types.Int32                                        `tfsdk:"instance_count"`
	InstanceType        fwtypes.StringEnum[awstypes.TransformInstanceType] `tfsdk:"instance_type"`
	TransformAmiVersion types.String                                       `tfsdk:"transform_ami_version"`
	VolumeKMSKeyID      types.String                                       `tfsdk:"volume_kms_key_id"`
}
