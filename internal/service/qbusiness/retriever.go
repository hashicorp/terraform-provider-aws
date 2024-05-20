// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package qbusiness

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/qbusiness"
	awstypes "github.com/aws/aws-sdk-go-v2/service/qbusiness/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
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
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
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
			"retriever_id": framework.IDAttribute(),
			names.AttrDisplayName: schema.StringAttribute{
				Description: "The display name of the Amazon Q retriever.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 1000),
					stringvalidator.RegexMatches(regexache.MustCompile(`^\P{C}*$`), "must not contain control characters"),
				},
			},
			"iam_service_role_arn": schema.StringAttribute{
				Description: "ARN of an IAM role used by Amazon Q to access the basic authentication credentials stored in a Secrets Manager secret.",
				Optional:    true,
			},
			names.AttrType: schema.StringAttribute{
				CustomType:  fwtypes.StringEnumType[awstypes.RetrieverType](),
				Required:    true,
				Description: "Type of retriever you are using.",
				Validators: []validator.String{
					enum.FrameworkValidate[awstypes.RetrieverType](),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
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
						Optional:    true,
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
						Optional:    true,
						Description: "Identifier for the Amazon Q index.",
						Validators: []validator.String{
							stringvalidator.RegexMatches(regexache.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9-]{35}$`), "must be a valid index ID"),
						},
					},
				},
				Blocks: map[string]schema.Block{
					"date_boost_override": schema.SetNestedBlock{
						CustomType: fwtypes.NewSetNestedObjectTypeOf[dateBoostConfigurationData](ctx),
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
								"boosting_level": schema.StringAttribute{
									Description: "Specifies how much a document attribute is boosted.",
									Optional:    true,
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
					},
					"number_boost_override": schema.SetNestedBlock{
						CustomType: fwtypes.NewSetNestedObjectTypeOf[numberBoostConfigurationData](ctx),
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
								"boosting_level": schema.StringAttribute{
									Description: "Specifies how much a document attribute is boosted.",
									Optional:    true,
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
					},
					"string_boost_override": schema.SetNestedBlock{
						CustomType: fwtypes.NewSetNestedObjectTypeOf[stringBoostConfigurationData](ctx),
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
								"boosting_level": schema.StringAttribute{
									Description: "Specifies how much a document attribute is boosted.",
									Optional:    true,
									Validators: []validator.String{
										enum.FrameworkValidate[awstypes.DocumentAttributeBoostingLevel](),
									},
								},
								"attribute_value_boosting": schema.MapAttribute{
									Description: "Specifies specific values of a STRING type document attribute being boosted.",
									ElementType: types.StringType,
									Optional:    true,
									Validators: []validator.Map{
										mapvalidator.SizeAtMost(10), //nolint:mnd // max number of attributes
									},
								},
							},
						},
					},
					"string_list_boost_override": schema.SetNestedBlock{
						CustomType: fwtypes.NewSetNestedObjectTypeOf[stringListBoostConfigurationData](ctx),
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
								"boosting_level": schema.StringAttribute{
									Description: "Specifies how much a document attribute is boosted.",
									Optional:    true,
									Validators: []validator.String{
										enum.FrameworkValidate[awstypes.DocumentAttributeBoostingLevel](),
									},
								},
							},
						},
					},
				},
			},
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
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

	input := &qbusiness.CreateRetrieverInput{}
	resp.Diagnostics.Append(fwflex.Expand(ctx, data, input)...)
	if resp.Diagnostics.HasError() {
		return
	}
	input.Tags = getTagsIn(ctx)
	input.ClientToken = aws.String(id.UniqueId())

	conf, d := data.expandConfiguration(ctx, true)

	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	input.Configuration = conf

	conn := r.Meta().QBusinessClient(ctx)
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

	conf, d = data.expandConfiguration(ctx, false)

	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	if c, ok := conf.(*awstypes.RetrieverConfigurationMemberNativeIndexConfiguration); ok {
		if len(c.Value.BoostingOverride) > 0 {
			update := &qbusiness.UpdateRetrieverInput{
				ApplicationId: input.ApplicationId,
				RetrieverId:   out.RetrieverId,
				Configuration: conf,
			}

			if _, err := conn.UpdateRetriever(ctx, update); err != nil {
				resp.Diagnostics.AddError("failed to update Amazon Q retriever", err.Error())
				return
			}
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *resourceRetriever) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	data := &resourceRetrieverData{}

	resp.Diagnostics.Append(req.State.Get(ctx, data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().QBusinessClient(ctx)

	input := &qbusiness.DeleteRetrieverInput{
		ApplicationId: aws.String(data.ApplicationId.ValueString()),
		RetrieverId:   aws.String(data.RetrieverId.ValueString()),
	}

	_, err := conn.DeleteRetriever(ctx, input)

	if err != nil {
		var nfe *awstypes.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return
		}
		resp.Diagnostics.AddError(fmt.Sprintf("failed to delete Q Business retriever (%s)", data.RetrieverId.ValueString()), err.Error())
		return
	}

	if _, err := waitRetrieverDeleted(ctx, conn, data.ID.ValueString(), r.DeleteTimeout(ctx, data.Timeouts)); err != nil {
		resp.Diagnostics.AddError("failed to wait for Q Business retriever deletion", err.Error())
		return
	}
}

func (r *resourceRetriever) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data resourceRetrieverData
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().QBusinessClient(ctx)

	out, err := FindRetrieverByID(ctx, conn, data.ID.ValueString())

	if tfresource.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("failed to retrieve Q Business retriever (%s)", data.ID.ValueString()), err.Error())
		return
	}

	resp.Diagnostics.Append(fwflex.Flatten(ctx, out, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(data.flattenConfiguration(ctx, out.Configuration)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *resourceRetriever) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var old, new resourceRetrieverData
	resp.Diagnostics.Append(req.State.Get(ctx, &old)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(req.Plan.Get(ctx, &new)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !old.KendraIndexConfiguration.Equal(new.KendraIndexConfiguration) ||
		!old.NativeIndexConfiguration.Equal(new.NativeIndexConfiguration) ||
		!old.RoleArn.Equal(new.RoleArn) ||
		!old.DisplayName.Equal(new.DisplayName) {
		input := &qbusiness.UpdateRetrieverInput{
			ApplicationId: aws.String(old.ApplicationId.ValueString()),
			RetrieverId:   aws.String(old.RetrieverId.ValueString()),
		}

		resp.Diagnostics.Append(fwflex.Expand(ctx, new, input)...)
		if resp.Diagnostics.HasError() {
			return
		}
		conf, d := new.expandConfiguration(ctx, false)

		resp.Diagnostics.Append(d...)
		if resp.Diagnostics.HasError() {
			return
		}

		input.Configuration = conf
		conn := r.Meta().QBusinessClient(ctx)

		if _, err := conn.UpdateRetriever(ctx, input); err != nil {
			resp.Diagnostics.AddError("failed to update Amazon Q retriever", err.Error())
			return
		}
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &new)...)
}

func (r *resourceRetriever) ModifyPlan(ctx context.Context, request resource.ModifyPlanRequest, response *resource.ModifyPlanResponse) {
	r.SetTagsAll(ctx, request, response)
}

func (r *resourceRetriever) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data resourceRetrieverData
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !data.NativeIndexConfiguration.IsNull() && data.Type.ValueEnum() == awstypes.RetrieverTypeKendraIndex {
		resp.Diagnostics.AddAttributeError(
			path.Root("native_index_configuration"),
			"Invalid Attribute Configuration",
			"'native_index_configuration' is not allowed when 'type' is 'KENDRA_INDEX'",
		)
	}

	if !data.KendraIndexConfiguration.IsNull() && data.Type.ValueEnum() == awstypes.RetrieverTypeNativeIndex {
		resp.Diagnostics.AddAttributeError(
			path.Root("kendra_index_configuration"),
			"Invalid Attribute Configuration",
			"'kendra_index_configuration' is not allowed when 'type' is 'NATIVE_INDEX'",
		)
	}
}

const (
	retrieverResourceIDPartCount = 2
)

func (data *resourceRetrieverData) setID() {
	data.ID = types.StringValue(errs.Must(flex.FlattenResourceId([]string{data.ApplicationId.ValueString(), data.RetrieverId.ValueString()}, retrieverResourceIDPartCount, false)))
}

func (data *resourceRetrieverData) flattenConfiguration(ctx context.Context, retrieverConf awstypes.RetrieverConfiguration) diag.Diagnostics {
	var diags diag.Diagnostics

	if conf, ok := retrieverConf.(*awstypes.RetrieverConfigurationMemberKendraIndexConfiguration); ok {
		c, d := fwtypes.NewObjectValueOf[kendraIndexConfigurationData](ctx, &kendraIndexConfigurationData{
			IndexId: fwflex.StringToFramework(ctx, conf.Value.IndexId),
		})
		diags.Append(d...)
		data.KendraIndexConfiguration = c
	}

	if conf, ok := retrieverConf.(*awstypes.RetrieverConfigurationMemberNativeIndexConfiguration); ok {
		stringBoostOverrides := []*stringBoostConfigurationData{}
		numberBoostOverrides := []*numberBoostConfigurationData{}
		dateBoostOverrides := []*dateBoostConfigurationData{}
		stringListBoostOverrides := []*stringListBoostConfigurationData{}

		for k, v := range conf.Value.BoostingOverride {
			if stringConf, d := v.(*awstypes.DocumentAttributeBoostingConfigurationMemberStringConfiguration); d {
				var sb stringBoostConfigurationData
				diags.Append(fwflex.Flatten(ctx, struct {
					BoostingLevel          awstypes.DocumentAttributeBoostingLevel
					AttributeValueBoosting map[string]string
				}{
					BoostingLevel:          stringConf.Value.BoostingLevel,
					AttributeValueBoosting: convertStringAttributeValueBoostingLevelMap(stringConf.Value.AttributeValueBoosting),
				}, &sb)...)
				sb.BoostKey = fwflex.StringValueToFramework(ctx, k)
				stringBoostOverrides = append(stringBoostOverrides, &sb)
			} else if numberConf, d := v.(*awstypes.DocumentAttributeBoostingConfigurationMemberNumberConfiguration); d {
				var nb numberBoostConfigurationData
				diags.Append(fwflex.Flatten(ctx, numberConf.Value, &nb)...)
				nb.BoostKey = fwflex.StringValueToFramework(ctx, k)
				numberBoostOverrides = append(numberBoostOverrides, &nb)
			} else if dateConf, d := v.(*awstypes.DocumentAttributeBoostingConfigurationMemberDateConfiguration); d {
				var db dateBoostConfigurationData
				diags.Append(fwflex.Flatten(ctx, dateConf.Value, &db)...)
				db.BoostKey = fwflex.StringValueToFramework(ctx, k)
				dateBoostOverrides = append(dateBoostOverrides, &db)
			} else if stringListConf, d := v.(*awstypes.DocumentAttributeBoostingConfigurationMemberStringListConfiguration); d {
				var slb stringListBoostConfigurationData
				diags.Append(fwflex.Flatten(ctx, stringListConf.Value, &slb)...)
				slb.BoostKey = fwflex.StringValueToFramework(ctx, k)
				stringListBoostOverrides = append(stringListBoostOverrides, &slb)
			}
		}

		c, d := fwtypes.NewObjectValueOf[nativeIndexConfigurationData](ctx, &nativeIndexConfigurationData{
			IndexId:                    fwflex.StringToFramework(ctx, conf.Value.IndexId),
			StringBoostingOverride:     fwtypes.NewSetNestedObjectValueOfSliceMust(ctx, stringBoostOverrides),
			NumberBoostingOverride:     fwtypes.NewSetNestedObjectValueOfSliceMust(ctx, numberBoostOverrides),
			StringListBoostingOverride: fwtypes.NewSetNestedObjectValueOfSliceMust(ctx, stringListBoostOverrides),
			DateBoostingOverride:       fwtypes.NewSetNestedObjectValueOfSliceMust(ctx, dateBoostOverrides),
		})
		diags.Append(d...)
		data.NativeIndexConfiguration = c
	}

	return diags
}

func (data *resourceRetrieverData) expandConfiguration(ctx context.Context, omitBoostingOverrideData bool) (awstypes.RetrieverConfiguration, diag.Diagnostics) {
	var diags diag.Diagnostics

	if !data.NativeIndexConfiguration.IsNull() {
		nativeConfig, d := data.NativeIndexConfiguration.ToPtr(ctx)
		diags.Append(d...)

		ret := &awstypes.RetrieverConfigurationMemberNativeIndexConfiguration{}
		ret.Value.IndexId = nativeConfig.IndexId.ValueStringPointer()

		if omitBoostingOverrideData {
			return ret, diags
		}

		ret.Value.BoostingOverride = make(map[string]awstypes.DocumentAttributeBoostingConfiguration)

		dateBoosts := []dateBoostConfigurationData{}
		diags.Append(nativeConfig.DateBoostingOverride.ElementsAs(ctx, &dateBoosts, false)...)
		for _, dateBoost := range dateBoosts {
			ret.Value.BoostingOverride[dateBoost.BoostKey.ValueString()] = &awstypes.DocumentAttributeBoostingConfigurationMemberDateConfiguration{
				Value: awstypes.DateAttributeBoostingConfiguration{
					BoostingLevel:             dateBoost.BoostingLevel.ValueEnum(),
					BoostingDurationInSeconds: dateBoost.BoostingDurationInSeconds.ValueInt64Pointer(),
				},
			}
		}

		numberBoosts := []numberBoostConfigurationData{}
		diags.Append(nativeConfig.NumberBoostingOverride.ElementsAs(ctx, &numberBoosts, false)...)
		for _, numberBoost := range numberBoosts {
			ret.Value.BoostingOverride[numberBoost.BoostKey.ValueString()] = &awstypes.DocumentAttributeBoostingConfigurationMemberNumberConfiguration{
				Value: awstypes.NumberAttributeBoostingConfiguration{
					BoostingLevel: numberBoost.BoostingLevel.ValueEnum(),
					BoostingType:  numberBoost.BoostingType.ValueEnum(),
				},
			}
		}

		stringBoosts := []stringBoostConfigurationData{}
		diags.Append(nativeConfig.StringBoostingOverride.ElementsAs(ctx, &stringBoosts, false)...)
		for _, stringBoost := range stringBoosts {
			m := map[string]awstypes.StringAttributeValueBoostingLevel{}
			diags.Append(stringBoost.AttributeValueBoosting.ElementsAs(ctx, &m, false)...)
			ret.Value.BoostingOverride[stringBoost.BoostKey.ValueString()] = &awstypes.DocumentAttributeBoostingConfigurationMemberStringConfiguration{
				Value: awstypes.StringAttributeBoostingConfiguration{
					BoostingLevel:          stringBoost.BoostingLevel.ValueEnum(),
					AttributeValueBoosting: m,
				},
			}
		}

		stringListBoosts := []stringListBoostConfigurationData{}
		diags.Append(nativeConfig.StringListBoostingOverride.ElementsAs(ctx, &stringListBoosts, false)...)
		for _, stringListBoost := range stringListBoosts {
			ret.Value.BoostingOverride[stringListBoost.BoostKey.ValueString()] = &awstypes.DocumentAttributeBoostingConfigurationMemberStringListConfiguration{
				Value: awstypes.StringListAttributeBoostingConfiguration{
					BoostingLevel: stringListBoost.BoostingLevel.ValueEnum(),
				},
			}
		}
		return ret, diags
	}

	if !data.KendraIndexConfiguration.IsNull() {
		kendraConfig, d := data.KendraIndexConfiguration.ToPtr(ctx)
		diags.Append(d...)

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
	IndexId                    types.String                                                     `tfsdk:"index_id"`
	DateBoostingOverride       fwtypes.SetNestedObjectValueOf[dateBoostConfigurationData]       `tfsdk:"date_boost_override"`
	NumberBoostingOverride     fwtypes.SetNestedObjectValueOf[numberBoostConfigurationData]     `tfsdk:"number_boost_override"`
	StringBoostingOverride     fwtypes.SetNestedObjectValueOf[stringBoostConfigurationData]     `tfsdk:"string_boost_override"`
	StringListBoostingOverride fwtypes.SetNestedObjectValueOf[stringListBoostConfigurationData] `tfsdk:"string_list_boost_override"`
}

type dateBoostConfigurationData struct {
	BoostKey                  types.String                                                `tfsdk:"boost_key"`
	BoostingLevel             fwtypes.StringEnum[awstypes.DocumentAttributeBoostingLevel] `tfsdk:"boosting_level"`
	BoostingDurationInSeconds types.Int64                                                 `tfsdk:"boosting_duration"`
}

type numberBoostConfigurationData struct {
	BoostKey      types.String                                                `tfsdk:"boost_key"`
	BoostingLevel fwtypes.StringEnum[awstypes.DocumentAttributeBoostingLevel] `tfsdk:"boosting_level"`
	BoostingType  fwtypes.StringEnum[awstypes.NumberAttributeBoostingType]    `tfsdk:"boosting_type"`
}

type stringBoostConfigurationData struct {
	BoostKey               types.String                                                `tfsdk:"boost_key"`
	BoostingLevel          fwtypes.StringEnum[awstypes.DocumentAttributeBoostingLevel] `tfsdk:"boosting_level"`
	AttributeValueBoosting fwtypes.MapValueOf[types.String]                            `tfsdk:"attribute_value_boosting"`
}

type stringListBoostConfigurationData struct {
	BoostKey      types.String                                                `tfsdk:"boost_key"`
	BoostingLevel fwtypes.StringEnum[awstypes.DocumentAttributeBoostingLevel] `tfsdk:"boosting_level"`
}

func convertStringAttributeValueBoostingLevelMap(from map[string]awstypes.StringAttributeValueBoostingLevel) map[string]string {
	ret := make(map[string]string)
	for k, v := range from {
		ret[k] = string(v)
	}
	return ret
}

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
