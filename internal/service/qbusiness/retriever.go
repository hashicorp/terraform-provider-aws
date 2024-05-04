// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package qbusiness

import (
	"context"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/qbusiness"
	awstypes "github.com/aws/aws-sdk-go-v2/service/qbusiness/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/mapvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_qbusiness_retriever", name="Retriever")
// @Tags(identifierAttribute="arn")
func newResourceRetriever(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceRetriever{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameRetriever = "Retriever"
)

type resourceRetriever struct {
	framework.ResourceWithConfigure
	framework.WithImportByID
	framework.WithTimeouts
}

func (r *resourceRetriever) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_qbusiness_retriever"
}

func (r *resourceRetriever) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID:  framework.IDAttribute(),
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			"application_id": schema.StringAttribute{
				Description: "Identifier of the Amazon Q application associated with the retriever",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexache.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9-]{35}$`), "must be a valid application ID"),
				},
			},
			"retriever_id": schema.StringAttribute{
				Computed: true,
			},
			"display_name": schema.StringAttribute{
				Description: "The display name of the Amazon Q retriever.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 1000),
					stringvalidator.RegexMatches(regexache.MustCompile(`^\P{C}*$`), "must not contain control characters"),
				},
			},
			"iam_service_role_arn": schema.StringAttribute{
				CustomType:  fwtypes.ARNType,
				Description: "ARN of an IAM role used by Amazon Q to access the basic authentication credentials stored in a Secrets Manager secret.",
				Optional:    true,
			},
			"type": schema.StringAttribute{
				Required:    true,
				Description: "Type of retriever you are using.",
				Validators: []validator.String{
					enum.FrameworkValidate[awstypes.RetrieverType](),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{

			"kendra_index_configuration": schema.SingleNestedBlock{
				CustomType: fwtypes.NewObjectTypeOf[kendraIndexConfigurationData](ctx),
				Validators: []validator.Object{
					objectvalidator.ConflictsWith(path.Expressions{path.MatchRoot("native_index_configuration")}...),
					objectvalidator.AtLeastOneOf(
						path.MatchRoot("kendra_index_configuration"),
						path.MatchRoot("native_index_configuration"),
					),
				},
				Attributes: map[string]schema.Attribute{
					"index_id": schema.StringAttribute{
						Required:    true,
						Description: "Identifier of the Amazon Kendra index.",
						Validators: []validator.String{
							stringvalidator.RegexMatches(regexache.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9-]{35}$`), "must be a valid index ID"),
						},
					},
				},
			},

			"native_index_configuration": schema.SingleNestedBlock{
				CustomType: fwtypes.NewObjectTypeOf[nativeIndexConfigurationData](ctx),
				Validators: []validator.Object{
					objectvalidator.ConflictsWith(path.Expressions{path.MatchRoot("kendra_index_configuration")}...),
					objectvalidator.AtLeastOneOf(
						path.MatchRoot("kendra_index_configuration"),
						path.MatchRoot("native_index_configuration"),
					),
				},
				Attributes: map[string]schema.Attribute{
					"index_id": schema.StringAttribute{
						Required:    true,
						Description: "Identifier for the Amazon Q index.",
						Validators: []validator.String{
							stringvalidator.RegexMatches(regexache.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9-]{35}$`), "must be a valid index ID"),
						},
					},
				},
				Blocks: map[string]schema.Block{
					"boosting_override": schema.ListNestedBlock{
						CustomType: fwtypes.NewListNestedObjectTypeOf[boostingOverrideData](ctx),
						Validators: []validator.List{
							listvalidator.SizeAtMost(200),
						},
						NestedObject: schema.NestedBlockObject{
							Attributes: map[string]schema.Attribute{
								"boost_key": schema.StringAttribute{
									Description: "Overrides the default boosts applied by Amazon Q Business to supported document attribute data types.",
									Required:    true,
									Validators: []validator.String{
										stringvalidator.LengthBetween(1, 200),
										stringvalidator.RegexMatches(regexache.MustCompile(`^[a-zA-Z0-9_][a-zA-Z0-9_-]*$`), "must begin with a letter or number and contain only alphanumeric, underscore, or hyphen characters"),
									},
								},
							},
							Blocks: map[string]schema.Block{
								"date_configuration": schema.SingleNestedBlock{
									CustomType: fwtypes.NewObjectTypeOf[dateConfigurationData](ctx),
									Validators: []validator.Object{
										objectvalidator.ExactlyOneOf(
											path.MatchRelative().AtParent().AtName("date_configuration"),
											path.MatchRelative().AtParent().AtName("number_configuration"),
											path.MatchRelative().AtParent().AtName("string_configuration"),
											path.MatchRelative().AtParent().AtName("string_list_configuration"),
										),
									},
									Attributes: map[string]schema.Attribute{
										"boosting_level": schema.StringAttribute{
											Description: "Specifies how much a document attribute is boosted.",
											Required:    true,
											Validators: []validator.String{
												enum.FrameworkValidate[awstypes.DocumentAttributeBoostingLevel](),
											},
										},
										"boosting_duration": schema.Int64Attribute{
											Description: "The duration of the boost in seconds.",
											Optional:    true,
											Validators: []validator.Int64{
												int64validator.Between(0, 999999999),
											},
										},
									},
								},
								"number_configuration": schema.SingleNestedBlock{
									CustomType: fwtypes.NewObjectTypeOf[numberConfigurationData](ctx),
									Attributes: map[string]schema.Attribute{
										"boosting_level": schema.StringAttribute{
											Description: "Specifies how much a document attribute is boosted.",
											Required:    true,
											Validators: []validator.String{
												enum.FrameworkValidate[awstypes.DocumentAttributeBoostingLevel](),
											},
										},
										"boosting_type": schema.StringAttribute{
											Description: "Specifies how a document attribute is boosted.",
											Optional:    true,
											Validators: []validator.String{
												enum.FrameworkValidate[awstypes.NumberAttributeBoostingType](),
											},
										},
									},
								},
								"string_configuration": schema.SingleNestedBlock{
									CustomType: fwtypes.NewObjectTypeOf[stringConfigurationData](ctx),
									Attributes: map[string]schema.Attribute{
										"boosting_level": schema.StringAttribute{
											Description: "Specifies how much a document attribute is boosted.",
											Required:    true,
											Validators: []validator.String{
												enum.FrameworkValidate[awstypes.DocumentAttributeBoostingLevel](),
											},
										},
										"attribute_value_boosting": schema.MapAttribute{
											Description: "Specifies specific values of a STRING type document attribute being boosted.",
											ElementType: types.StringType,
											Optional:    true,
											Validators: []validator.Map{
												mapvalidator.SizeAtMost(10),
											},
										},
									},
								},
								"string_list_configuration": schema.SingleNestedBlock{
									CustomType: fwtypes.NewObjectTypeOf[stringListConfigurationData](ctx),
									Attributes: map[string]schema.Attribute{
										"boosting_level": schema.StringAttribute{
											Description: "Specifies how much a document attribute is boosted.",
											Required:    true,
											Validators: []validator.String{
												enum.FrameworkValidate[awstypes.DocumentAttributeBoostingLevel](),
											},
										},
									},
								},
							},
						},
					},
				},
			},
			"timeouts": timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Delete: true,
			}),
		},
	}
}

func (r *resourceRetriever) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data resourceRetrieverData

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().QBusinessClient(ctx)

	input := &qbusiness.CreateRetrieverInput{}
	resp.Diagnostics.Append(fwflex.Expand(ctx, data, input)...)
	if resp.Diagnostics.HasError() {
		return
	}
	input.Tags = getTagsIn(ctx)
	input.ClientToken = aws.String(id.UniqueId())

	// AutoFlEx doesn't handle union types.
	conf, d := data.expandConfiguration(ctx)

	resp.Diagnostics.Append(d...)

	if resp.Diagnostics.HasError() {
		return
	}

	input.Configuration = conf

	out, err := conn.CreateRetriever(ctx, input)

	if err != nil {
		resp.Diagnostics.AddError("failed to create Amazon Q retriever", err.Error())
		return
	}

	data.RetrieverId = fwflex.StringToFramework(ctx, out.RetrieverId)
	data.RetrieverArn = fwflex.StringToFramework(ctx, out.RetrieverArn)

	data.setID()

	if _, err := waitRetrieverCreated(ctx, conn, data.ID.ValueString(), r.CreateTimeout(ctx, data.Timeouts)); err != nil {
		resp.Diagnostics.AddError("failed to wait for Q Business retriever creation", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *resourceRetriever) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
}

func (r *resourceRetriever) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
}

func (r *resourceRetriever) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
}

func (r *resourceRetriever) ModifyPlan(ctx context.Context, request resource.ModifyPlanRequest, response *resource.ModifyPlanResponse) {
	r.SetTagsAll(ctx, request, response)
}

func (data *resourceRetrieverData) setID() {
	data.ID = types.StringValue(errs.Must(flex.FlattenResourceId([]string{data.ApplicationId.ValueString(), data.RetrieverId.ValueString()}, indexResourceIDPartCount, false)))
}

func (data *resourceRetrieverData) expandConfiguration(ctx context.Context) (awstypes.RetrieverConfiguration, diag.Diagnostics) {
	var diags diag.Diagnostics

	if !data.NativeIndexConfiguration.IsNull() {
		nativeConfig, d := data.NativeIndexConfiguration.ToPtr(ctx)

		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		ret := &awstypes.RetrieverConfigurationMemberNativeIndexConfiguration{}
		ret.Value.IndexId = nativeConfig.IndexId.ValueStringPointer()

		boosts := []boostingOverrideData{}

		diags.Append(nativeConfig.BoostingOverride.ElementsAs(ctx, &boosts, false)...)
		if diags.HasError() {
			return nil, diags
		}

		for _, boost := range boosts {
			if !boost.DateConfiguration.IsNull() {

				dateConf, d := boost.DateConfiguration.ToPtr(ctx)

				diags.Append(d...)
				if diags.HasError() {
					return nil, diags
				}

				ret.Value.BoostingOverride[boost.BoostKey.ValueString()] = &awstypes.DocumentAttributeBoostingConfigurationMemberDateConfiguration{
					Value: awstypes.DateAttributeBoostingConfiguration{
						BoostingLevel:             dateConf.BoostingLevel.ValueEnum(),
						BoostingDurationInSeconds: dateConf.BoostingDurationInSeconds.ValueInt64Pointer(),
					},
				}
			} else if !boost.NumberConfiguration.IsNull() {
				numberConf, d := boost.NumberConfiguration.ToPtr(ctx)

				diags.Append(d...)
				if diags.HasError() {
					return nil, diags
				}

				ret.Value.BoostingOverride[boost.BoostKey.ValueString()] = &awstypes.DocumentAttributeBoostingConfigurationMemberNumberConfiguration{
					Value: awstypes.NumberAttributeBoostingConfiguration{
						BoostingLevel: numberConf.BoostingLevel.ValueEnum(),
						BoostingType:  numberConf.BoostingType.ValueEnum(),
					},
				}
			} else if !boost.StringConfiguration.IsNull() {
				stringConf, d := boost.StringConfiguration.ToPtr(ctx)

				diags.Append(d...)
				if diags.HasError() {
					return nil, diags
				}

				m := map[string]awstypes.StringAttributeValueBoostingLevel{}
				diags.Append(stringConf.AttributeValueBoosting.ElementsAs(ctx, &m, false)...)

				if diags.HasError() {
					return nil, diags
				}

				ret.Value.BoostingOverride[boost.BoostKey.ValueString()] = &awstypes.DocumentAttributeBoostingConfigurationMemberStringConfiguration{
					Value: awstypes.StringAttributeBoostingConfiguration{
						BoostingLevel:          stringConf.BoostingLevel.ValueEnum(),
						AttributeValueBoosting: m,
					},
				}
			} else if !boost.StringListConfiguration.IsNull() {
				stringListConf, d := boost.StringListConfiguration.ToPtr(ctx)

				diags.Append(d...)
				if diags.HasError() {
					return nil, diags
				}

				ret.Value.BoostingOverride[boost.BoostKey.ValueString()] = &awstypes.DocumentAttributeBoostingConfigurationMemberStringListConfiguration{
					Value: awstypes.StringListAttributeBoostingConfiguration{
						BoostingLevel: stringListConf.BoostingLevel.ValueEnum(),
					},
				}
			}
		}
		return ret, diags
	}

	if !data.KendraIndexConfiguration.IsNull() {
		kendraConfig, d := data.KendraIndexConfiguration.ToPtr(ctx)

		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		ret := &awstypes.RetrieverConfigurationMemberKendraIndexConfiguration{}
		ret.Value.IndexId = kendraConfig.IndexId.ValueStringPointer()

		return ret, diags
	}

	return nil, diags
}

type resourceRetrieverData struct {
	ApplicationId            types.String                                        `tfsdk:"application_id"`
	ID                       types.String                                        `tfsdk:"id"`
	Tags                     types.Map                                           `tfsdk:"tags"`
	TagsAll                  types.Map                                           `tfsdk:"tags_all"`
	Timeouts                 timeouts.Value                                      `tfsdk:"timeouts"`
	DisplayName              types.String                                        `tfsdk:"display_name"`
	RoleArn                  types.String                                        `tfsdk:"iam_service_role_arn"`
	RetrieverId              types.String                                        `tfsdk:"retriever_id"`
	RetrieverArn             types.String                                        `tfsdk:"arn"`
	Type                     fwtypes.StringEnum[awstypes.RetrieverType]          `tfsdk:"type"`
	KendraIndexConfiguration fwtypes.ObjectValueOf[kendraIndexConfigurationData] `tfsdk:"kendra_index_configuration"`
	NativeIndexConfiguration fwtypes.ObjectValueOf[nativeIndexConfigurationData] `tfsdk:"native_index_configuration"`
}

type kendraIndexConfigurationData struct {
	IndexId types.String `tfsdk:"index_id"`
}

type nativeIndexConfigurationData struct {
	IndexId          types.String                                          `tfsdk:"index_id"`
	BoostingOverride fwtypes.ListNestedObjectValueOf[boostingOverrideData] `tfsdk:"boosting_override"`
}

type boostingOverrideData struct {
	BoostKey                types.String                                       `tfsdk:"boost_key"`
	DateConfiguration       fwtypes.ObjectValueOf[dateConfigurationData]       `tfsdk:"date_configuration"`
	NumberConfiguration     fwtypes.ObjectValueOf[numberConfigurationData]     `tfsdk:"number_configuration"`
	StringListConfiguration fwtypes.ObjectValueOf[stringListConfigurationData] `tfsdk:"string_list_configuration"`
	StringConfiguration     fwtypes.ObjectValueOf[stringConfigurationData]     `tfsdk:"string_configuration"`
}

type dateConfigurationData struct {
	BoostingLevel             fwtypes.StringEnum[awstypes.DocumentAttributeBoostingLevel] `tfsdk:"boosting_level"`
	BoostingDurationInSeconds types.Int64                                                 `tfsdk:"boosting_duration"`
}

type numberConfigurationData struct {
	BoostingLevel fwtypes.StringEnum[awstypes.DocumentAttributeBoostingLevel] `tfsdk:"boosting_level"`
	BoostingType  fwtypes.StringEnum[awstypes.NumberAttributeBoostingType]    `tfsdk:"boosting_type"`
}

type stringConfigurationData struct {
	BoostingLevel          fwtypes.StringEnum[awstypes.DocumentAttributeBoostingLevel] `tfsdk:"boosting_level"`
	AttributeValueBoosting fwtypes.MapValueOf[types.String]                            `tfsdk:"attribute_value_boosting"`
}

type stringListConfigurationData struct {
	BoostingLevel fwtypes.StringEnum[awstypes.DocumentAttributeBoostingLevel] `tfsdk:"boosting_level"`
}

const (
	retrieverResourceIDPartCount = 2
)

func FindRetrieverByID(ctx context.Context, conn *qbusiness.Client, id string) (*qbusiness.GetRetrieverOutput, error) {
	parts, err := flex.ExpandResourceId(id, retrieverResourceIDPartCount, false)

	if err != nil {
		return nil, err
	}

	input := &qbusiness.GetRetrieverInput{
		ApplicationId: aws.String(parts[0]),
		RetrieverId:   aws.String(parts[1]),
	}

	output, err := conn.GetRetriever(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}
