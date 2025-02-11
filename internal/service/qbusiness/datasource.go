// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package qbusiness

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/qbusiness"
	"github.com/aws/aws-sdk-go-v2/service/qbusiness/document"
	awstypes "github.com/aws/aws-sdk-go-v2/service/qbusiness/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
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

// @FrameworkResource("aws_qbusiness_datasource", name="Datasource")
// @Tags(identifierAttribute="arn")
func newResourceDatasource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceDatasource{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameDatasource = "Datasource"
)

type resourceDatasource struct {
	framework.ResourceWithConfigure
	framework.WithImportByID
	framework.WithTimeouts
}

func documentAttributeValueSchema(ctx context.Context) schema.SingleNestedBlock {
	return schema.SingleNestedBlock{
		CustomType: fwtypes.NewObjectTypeOf[resourceDocumentAttributeValueData](ctx),
		Attributes: map[string]schema.Attribute{
			"date_value": schema.StringAttribute{
				CustomType:  timetypes.RFC3339Type{},
				Description: "A date expressed as an ISO 8601 string.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.AtLeastOneOf(
						path.MatchRelative().AtParent().AtName("date_value"),
						path.MatchRelative().AtParent().AtName("long_value"),
						path.MatchRelative().AtParent().AtName("string_list_value"),
						path.MatchRelative().AtParent().AtName("string_value"),
					),
					stringvalidator.RegexMatches(regexache.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z$`), "must be an ISO 8601 date/time string"),
				},
			},
			"long_value": schema.Int64Attribute{
				Description: "A long integer value.",
				Optional:    true,
			},
			"string_list_value": schema.ListAttribute{
				CustomType:  fwtypes.ListOfStringType,
				ElementType: types.StringType,
				Description: "A list of string values.",
				Optional:    true,
				Validators: []validator.List{
					listvalidator.SizeAtMost(2048),
					listvalidator.SizeAtLeast(1),
				},
			},
			"string_value": schema.StringAttribute{
				Description: "A string value.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(0, 2048),
				},
			},
		},
	}
}

func conditionSchema(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[resourceConditionData](ctx),
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				names.AttrKey: schema.StringAttribute{
					Description: "The identifier of the document attribute used for the condition.",
					Required:    true,
					Validators: []validator.String{
						stringvalidator.LengthBetween(1, 200),
						stringvalidator.RegexMatches(regexache.MustCompile(`^[a-zA-Z0-9_][a-zA-Z0-9_-]*$`), "must be a valid document attribute"),
					},
				},
				"operator": schema.StringAttribute{
					CustomType:  fwtypes.StringEnumType[awstypes.DocumentEnrichmentConditionOperator](),
					Required:    true,
					Description: "Operator of the document attribute used for the condition.",
					Validators: []validator.String{
						enum.FrameworkValidate[awstypes.DocumentEnrichmentConditionOperator](),
					},
				},
			},
			Blocks: map[string]schema.Block{
				names.AttrValue: documentAttributeValueSchema(ctx),
			},
		},
	}
}

func hookConfigurationSchema(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[resourceHookConfigurationData](ctx),
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"lambda_arn": schema.StringAttribute{
					CustomType:  fwtypes.ARNType,
					Description: "ARN of a role with permission to run a Lambda function during ingestion.",
					Optional:    true,
					Validators: []validator.String{
						stringvalidator.LengthBetween(0, 2048),
						stringvalidator.RegexMatches(regexache.MustCompile(`^arn:aws[a-zA-Z-]*:lambda:[a-z-]*-[0-9]:[0-9]{12}:function:[a-zA-Z0-9-_]+(/[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12})?(:[a-zA-Z0-9-_]+)?$`), "must be a valid Lambda ARN"),
					},
				},
				names.AttrRoleARN: schema.StringAttribute{
					CustomType:  fwtypes.ARNType,
					Description: "ARN of a role with permission to run PreExtractionHookConfiguration and PostExtractionHookConfiguration for altering document metadata and content during the document ingestion process.",
					Optional:    true,
				},
				names.AttrS3Bucket: schema.StringAttribute{
					Description: "Stores the original, raw documents or the structured, parsed documents before and after altering them.",
					Optional:    true,
					Validators: []validator.String{
						stringvalidator.LengthBetween(1, 63),
						stringvalidator.RegexMatches(regexache.MustCompile(`^[a-z0-9][\.\-a-z0-9]{1,61}[a-z0-9]$`), "must be a valid bucket name"),
					},
				},
			},
			Blocks: map[string]schema.Block{
				"invocation_condition": conditionSchema(ctx),
			},
		},
	}
}

func documentAttributeTargetSchema(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[resourceDocumentAttributeTargetData](ctx),
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				names.AttrKey: schema.StringAttribute{
					Description: "The identifier of the document attribute used for the condition.",
					Required:    true,
					Validators: []validator.String{
						stringvalidator.LengthBetween(1, 200),
						stringvalidator.RegexMatches(regexache.MustCompile(`^[a-zA-Z0-9_][a-zA-Z0-9_-]*$`), "must be a valid document attribute"),
					},
				},
				"attribute_value_operator": schema.StringAttribute{
					CustomType:  fwtypes.StringEnumType[awstypes.AttributeValueOperator](),
					Required:    true,
					Description: "Operator of the document attribute used for the condition.",
					Validators: []validator.String{
						enum.FrameworkValidate[awstypes.AttributeValueOperator](),
					},
				},
			},
			Blocks: map[string]schema.Block{
				names.AttrValue: documentAttributeValueSchema(ctx),
			},
		},
	}
}

func (r *resourceDatasource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_qbusiness_datasource"
}

func (r *resourceDatasource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID:  framework.IDAttribute(),
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrDescription: schema.StringAttribute{
				Description: "A description of the Amazon Q datasource.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 1000),
					stringvalidator.RegexMatches(regexache.MustCompile(`^\P{C}*$`), "must not contain control characters"),
				},
			},
			names.AttrConfiguration: schema.StringAttribute{
				Description: "Configuration information (JSON) to connect to your data source repository.",
				Required:    true,
			},
			names.AttrApplicationID: schema.StringAttribute{
				Description: "Identifier of the Amazon Q application associated with the datasource",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexache.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9-]{35}$`), "must be a valid application ID"),
				},
			},
			"datasource_id": framework.IDAttribute(),
			names.AttrDisplayName: schema.StringAttribute{
				Description: "The display name of the datasource.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 100),
					stringvalidator.RegexMatches(regexache.MustCompile(`^\P{C}*$`), "must not contain control characters"),
				},
			},
			"index_id": schema.StringAttribute{
				Description: "The identifier of the index that you want to use with the data source connector.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexache.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9-]{35}$`), "must be a valid index ID"),
				},
			},
			"iam_service_role_arn": schema.StringAttribute{
				CustomType:  fwtypes.ARNType,
				Description: "ARN of an IAM role with permission to access the data source and required resources.",
				Optional:    true,
			},
			"sync_schedule": schema.StringAttribute{
				Description: "Frequency for Amazon Q to check the documents in your data source repository and update your index.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 998),
					stringvalidator.RegexMatches(regexache.MustCompile(`^\P{C}*$`), "must not contain control characters"),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			names.AttrVPCConfig: schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[resourceVPCConfigurationData](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrVPCSecurityGroupIDs: schema.SetAttribute{
							PlanModifiers: []planmodifier.Set{
								setplanmodifier.RequiresReplace(),
							},
							Required:    true,
							CustomType:  fwtypes.SetOfStringType,
							ElementType: types.StringType,
							Description: "A list of security group IDs to allow access to the data source.",
						},
						names.AttrSubnetIDs: schema.SetAttribute{
							PlanModifiers: []planmodifier.Set{
								setplanmodifier.RequiresReplace(),
							},
							Required:    true,
							CustomType:  fwtypes.SetOfStringType,
							ElementType: types.StringType,
							Description: "A list of subnet IDs to allow access to the data source.",
						},
					},
				},
			},
			"document_enrichment_configuration": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[resourceDocumentEnrichmentConfigurationData](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				Description: "Configuration information for altering document metadata and content during the document ingestion process.",
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"inline_configuration": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[resourceInlineDocumentEnrichmentConfigurationData](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(100),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"document_content_operator": schema.StringAttribute{
										CustomType:  fwtypes.StringEnumType[awstypes.DocumentContentOperator](),
										Description: "'DELETE' to delete content if the condition used for the target attribute is met",
										Optional:    true,
										Validators: []validator.String{
											enum.FrameworkValidate[awstypes.DocumentContentOperator](),
										},
									},
								},
								Blocks: map[string]schema.Block{
									names.AttrCondition: conditionSchema(ctx),
									names.AttrTarget:    documentAttributeTargetSchema(ctx),
								},
							},
						},
						"post_extraction_hook_configuration": hookConfigurationSchema(ctx),
						"pre_extraction_hook_configuration":  hookConfigurationSchema(ctx),
					},
				},
			},
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Delete: true,
				Update: true,
			}),
		},
	}
}

func (r *resourceDatasource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data resourceDatasourceData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input, d := data.expandToCreateDataSourceInput(ctx)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}
	input.Tags = getTagsIn(ctx)
	input.ClientToken = aws.String(id.UniqueId())

	conn := r.Meta().QBusinessClient(ctx)
	out, err := conn.CreateDataSource(ctx, input)

	if err != nil {
		resp.Diagnostics.AddError("failed to create Q Business datasource", err.Error())
		return
	}

	data.DatasourceId = fwflex.StringToFramework(ctx, out.DataSourceId)
	data.DatasourceArn = fwflex.StringToFramework(ctx, out.DataSourceArn)
	resp.Diagnostics.Append(data.setID()...)
	if resp.Diagnostics.HasError() {
		return
	}

	if _, err := waitDatasourceCreated(ctx, conn, data.ID.ValueString(), r.CreateTimeout(ctx, data.Timeouts)); err != nil {
		resp.Diagnostics.AddError("failed to wait for datasource to be created", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *resourceDatasource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	data := &resourceDatasourceData{}

	resp.Diagnostics.Append(req.State.Get(ctx, data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().QBusinessClient(ctx)

	input := &qbusiness.DeleteDataSourceInput{
		ApplicationId: data.ApplicationId.ValueStringPointer(),
		IndexId:       data.IndexId.ValueStringPointer(),
		DataSourceId:  data.DatasourceId.ValueStringPointer(),
	}

	_, err := conn.DeleteDataSource(ctx, input)

	if err != nil {
		var nfe *awstypes.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return
		}
		resp.Diagnostics.AddError("failed to delete Q Business datasource", err.Error())
		return
	}

	if _, err := waitDatasourceDeleted(ctx, conn, data.ID.ValueString(), r.DeleteTimeout(ctx, data.Timeouts)); err != nil {
		resp.Diagnostics.AddError("failed to wait for datasource to be deleted", err.Error())
		return
	}
}

func (r *resourceDatasource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data resourceDatasourceData
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().QBusinessClient(ctx)
	out, err := FindDatasourceByID(ctx, conn, data.ID.ValueString())

	if tfresource.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("failed to retrieve Q Business datasource (%s)", data.ID.ValueString()), err.Error())
		return
	}

	resp.Diagnostics.Append(data.flattenFromGetDataSourceOutput(ctx, out)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *resourceDatasource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var old, new resourceDatasourceData

	resp.Diagnostics.Append(req.Plan.Get(ctx, &new)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(req.State.Get(ctx, &old)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !old.DisplayName.Equal(new.DisplayName) ||
		!old.Description.Equal(new.Description) ||
		!old.IndexId.Equal(new.IndexId) ||
		!old.RoleArn.Equal(new.RoleArn) ||
		!old.SyncSchedule.Equal(new.SyncSchedule) ||
		!old.Configuration.Equal(new.Configuration) ||
		!old.DocumentEnrichmentConfiguration.Equal(new.DocumentEnrichmentConfiguration) {
		conn := r.Meta().QBusinessClient(ctx)

		input, d := new.expandToUpdateDataSourceInput(ctx)
		resp.Diagnostics.Append(d...)
		if resp.Diagnostics.HasError() {
			return
		}

		_, err := conn.UpdateDataSource(ctx, input)

		if err != nil {
			resp.Diagnostics.AddError("failed to update Q Business datasource", err.Error())
			return
		}

		if _, err := waitDatasourceUpdated(ctx, conn, new.ID.ValueString(), r.UpdateTimeout(ctx, new.Timeouts)); err != nil {
			resp.Diagnostics.AddError("failed to wait for datasource to be updated", err.Error())
			return
		}
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &new)...)
}

func (r *resourceDatasource) ModifyPlan(ctx context.Context, request resource.ModifyPlanRequest, response *resource.ModifyPlanResponse) {
	r.SetTagsAll(ctx, request, response)
}

type resourceDatasourceData struct {
	ApplicationId                   types.String                                                                 `tfsdk:"application_id"`
	Configuration                   types.String                                                                 `tfsdk:"configuration"`
	DatasourceArn                   types.String                                                                 `tfsdk:"arn"`
	DatasourceId                    types.String                                                                 `tfsdk:"datasource_id"`
	Description                     types.String                                                                 `tfsdk:"description"`
	DisplayName                     types.String                                                                 `tfsdk:"display_name"`
	DocumentEnrichmentConfiguration fwtypes.ListNestedObjectValueOf[resourceDocumentEnrichmentConfigurationData] `tfsdk:"document_enrichment_configuration"`
	ID                              types.String                                                                 `tfsdk:"id"`
	IndexId                         types.String                                                                 `tfsdk:"index_id"`
	RoleArn                         fwtypes.ARN                                                                  `tfsdk:"iam_service_role_arn"`
	SyncSchedule                    types.String                                                                 `tfsdk:"sync_schedule"`
	Tags                            tftags.Map                                                                   `tfsdk:"tags"`
	TagsAll                         tftags.Map                                                                   `tfsdk:"tags_all"`
	Timeouts                        timeouts.Value                                                               `tfsdk:"timeouts"`
	VpcConfiguration                fwtypes.ListNestedObjectValueOf[resourceVPCConfigurationData]                `tfsdk:"vpc_config"`
}

type resourceDocumentAttributeValueData struct {
	DateValue       timetypes.RFC3339                 `tfsdk:"date_value"`
	LongValue       types.Int64                       `tfsdk:"long_value"`
	StringListValue fwtypes.ListValueOf[types.String] `tfsdk:"string_list_value"`
	StringValue     types.String                      `tfsdk:"string_value"`
}

type resourceConditionData struct {
	Key      types.String                                                     `tfsdk:"key"`
	Operator fwtypes.StringEnum[awstypes.DocumentEnrichmentConditionOperator] `tfsdk:"operator"`
	Value    fwtypes.ObjectValueOf[resourceDocumentAttributeValueData]        `tfsdk:"value"`
}

type resourceHookConfigurationData struct {
	LambdaARN    types.String                                           `tfsdk:"lambda_arn"`
	RoleARN      types.String                                           `tfsdk:"role_arn"`
	S3BucketName types.String                                           `tfsdk:"s3_bucket"`
	Condition    fwtypes.ListNestedObjectValueOf[resourceConditionData] `tfsdk:"invocation_condition"`
}

type resourceDocumentAttributeTargetData struct {
	Key      types.String                                              `tfsdk:"key"`
	Operator fwtypes.StringEnum[awstypes.AttributeValueOperator]       `tfsdk:"attribute_value_operator"`
	Value    fwtypes.ObjectValueOf[resourceDocumentAttributeValueData] `tfsdk:"value"`
}

type resourceInlineDocumentEnrichmentConfigurationData struct {
	Condition fwtypes.ListNestedObjectValueOf[resourceConditionData]               `tfsdk:"condition"`
	Target    fwtypes.ListNestedObjectValueOf[resourceDocumentAttributeTargetData] `tfsdk:"target"`
	Operator  fwtypes.StringEnum[awstypes.DocumentContentOperator]                 `tfsdk:"document_content_operator"`
}

type resourceDocumentEnrichmentConfigurationData struct {
	InlineConfigurations            fwtypes.ListNestedObjectValueOf[resourceInlineDocumentEnrichmentConfigurationData] `tfsdk:"inline_configuration"`
	PreExreactionHookConfiguration  fwtypes.ListNestedObjectValueOf[resourceHookConfigurationData]                     `tfsdk:"pre_extraction_hook_configuration"`
	PostExtractionHookConfiguration fwtypes.ListNestedObjectValueOf[resourceHookConfigurationData]                     `tfsdk:"post_extraction_hook_configuration"`
}

type resourceVPCConfigurationData struct {
	SecurityGroupIds fwtypes.SetValueOf[types.String] `tfsdk:"vpc_security_group_ids"`
	SubnetIds        fwtypes.SetValueOf[types.String] `tfsdk:"subnet_ids"`
}

func (r *resourceDatasourceData) expandConfiguration() (document.Interface, diag.Diagnostics) {
	var diags diag.Diagnostics

	var c map[string]interface{}
	err := json.Unmarshal([]byte(r.Configuration.ValueString()), &c)

	if err != nil {
		diags.AddError("failed to unmarshal configuration", err.Error())
		return nil, diags
	}
	return document.NewLazyDocument(c), nil
}

func (r *resourceDatasourceData) flattenFromGetDataSourceOutput(ctx context.Context, out *qbusiness.GetDataSourceOutput) diag.Diagnostics {
	var diags diag.Diagnostics
	r.ApplicationId = fwflex.StringValueToFramework(ctx, aws.ToString(out.ApplicationId))
	r.DatasourceArn = fwflex.StringValueToFramework(ctx, aws.ToString(out.DataSourceArn))
	r.DatasourceId = fwflex.StringValueToFramework(ctx, aws.ToString(out.DataSourceId))
	r.Description = fwflex.StringValueToFramework(ctx, aws.ToString(out.Description))
	r.DisplayName = fwflex.StringValueToFramework(ctx, aws.ToString(out.DisplayName))
	r.IndexId = fwflex.StringValueToFramework(ctx, aws.ToString(out.IndexId))
	if len(aws.ToString(out.RoleArn)) > 0 {
		r.RoleArn = fwflex.StringToFrameworkARN(ctx, out.RoleArn)
	}
	r.SyncSchedule = fwflex.StringValueToFramework(ctx, aws.ToString(out.SyncSchedule))

	if d := r.flattenConfiguration(out.Configuration); d.HasError() {
		return d
	}
	if d := r.flattenDocumentEnrichmentConfiguration(ctx, out.DocumentEnrichmentConfiguration); d.HasError() {
		return d
	}
	diags.Append(r.setID()...)
	return diags
}

func flatInlineConfiguration(ctx context.Context, conf []awstypes.InlineDocumentEnrichmentConfiguration) ([]*resourceInlineDocumentEnrichmentConfigurationData, diag.Diagnostics) {
	var diags diag.Diagnostics
	var idec []*resourceInlineDocumentEnrichmentConfigurationData

	for _, c := range conf {
		var ic resourceInlineDocumentEnrichmentConfigurationData
		ic.Operator = fwtypes.StringEnumValue(c.DocumentContentOperator)
		var cond resourceConditionData
		d := cond.flattenDocumentAttributeCondition(ctx, c.Condition)
		if d.HasError() {
			return nil, d
		}
		if ic.Condition, diags = fwtypes.NewListNestedObjectValueOfPtr[resourceConditionData](ctx, &cond); diags.HasError() {
			return nil, diags
		}
		var target resourceDocumentAttributeTargetData
		d = target.flattenResourceDocumentAttributeTargetData(ctx, c.Target)
		if d.HasError() {
			return nil, d
		}
		if ic.Target, diags = fwtypes.NewListNestedObjectValueOfPtr[resourceDocumentAttributeTargetData](ctx, &target); diags.HasError() {
			return nil, diags
		}
		idec = append(idec, &ic)
	}

	return idec, diags
}

func (data *resourceDocumentAttributeTargetData) flattenResourceDocumentAttributeTargetData(ctx context.Context, conf *awstypes.DocumentAttributeTarget) diag.Diagnostics {
	var diags diag.Diagnostics

	data.Key = fwflex.StringToFramework(ctx, conf.Key)
	data.Operator = fwtypes.StringEnumValue(conf.AttributeValueOperator)
	data.Value, diags = fwtypes.NewObjectValueOf[resourceDocumentAttributeValueData](ctx, flatValue(ctx, conf.Value))
	return diags
}

func (data *resourceConditionData) flattenDocumentAttributeCondition(ctx context.Context, conf *awstypes.DocumentAttributeCondition) diag.Diagnostics {
	var diags diag.Diagnostics

	data.Key = fwflex.StringToFramework(ctx, conf.Key)
	data.Operator = fwtypes.StringEnumValue(conf.Operator)
	data.Value, diags = fwtypes.NewObjectValueOf[resourceDocumentAttributeValueData](ctx, flatValue(ctx, conf.Value))
	return diags
}

func (hc *resourceHookConfigurationData) flattenHookConfiguration(ctx context.Context, conf *awstypes.HookConfiguration) diag.Diagnostics {
	var diags diag.Diagnostics

	hc.LambdaARN = fwflex.StringToFramework(ctx, conf.LambdaArn)
	hc.RoleARN = fwflex.StringToFramework(ctx, conf.RoleArn)
	hc.S3BucketName = fwflex.StringToFramework(ctx, conf.S3BucketName)

	if conf.InvocationCondition != nil {
		var c resourceConditionData
		d := c.flattenDocumentAttributeCondition(ctx, conf.InvocationCondition)
		if d.HasError() {
			return d
		}
		ic, d := fwtypes.NewListNestedObjectValueOfPtr[resourceConditionData](ctx, &c)
		if d.HasError() {
			return d
		}
		hc.Condition = ic
	}
	return diags
}

func (r *resourceDatasourceData) flattenDocumentEnrichmentConfiguration(ctx context.Context, conf *awstypes.DocumentEnrichmentConfiguration) diag.Diagnostics {
	var dec resourceDocumentEnrichmentConfigurationData
	var diags diag.Diagnostics

	if conf.InlineConfigurations != nil {
		ic, d := flatInlineConfiguration(ctx, conf.InlineConfigurations)
		if d.HasError() {
			return d
		}
		dec.InlineConfigurations = fwtypes.NewListNestedObjectValueOfSliceMust[resourceInlineDocumentEnrichmentConfigurationData](ctx, ic)
	} else {
		dec.InlineConfigurations = fwtypes.NewListNestedObjectValueOfNull[resourceInlineDocumentEnrichmentConfigurationData](ctx)
	}

	if conf.PreExtractionHookConfiguration != nil {
		var pre resourceHookConfigurationData
		d := pre.flattenHookConfiguration(ctx, conf.PreExtractionHookConfiguration)
		if d.HasError() {
			return d
		}
		if dec.PreExreactionHookConfiguration, diags = fwtypes.NewListNestedObjectValueOfPtr[resourceHookConfigurationData](ctx, &pre); diags.HasError() {
			return diags
		}
	} else {
		dec.PreExreactionHookConfiguration = fwtypes.NewListNestedObjectValueOfNull[resourceHookConfigurationData](ctx)
	}

	if conf.PostExtractionHookConfiguration != nil {
		var post resourceHookConfigurationData
		d := post.flattenHookConfiguration(ctx, conf.PostExtractionHookConfiguration)
		if d.HasError() {
			return d
		}
		if dec.PostExtractionHookConfiguration, diags = fwtypes.NewListNestedObjectValueOfPtr[resourceHookConfigurationData](ctx, &post); diags.HasError() {
			return diags
		}
	} else {
		dec.PostExtractionHookConfiguration = fwtypes.NewListNestedObjectValueOfNull[resourceHookConfigurationData](ctx)
	}

	if dec.InlineConfigurations.IsNull() && dec.PreExreactionHookConfiguration.IsNull() && dec.PostExtractionHookConfiguration.IsNull() {
		r.DocumentEnrichmentConfiguration = fwtypes.NewListNestedObjectValueOfNull[resourceDocumentEnrichmentConfigurationData](ctx)
		return diags
	}

	l, d := fwtypes.NewListNestedObjectValueOfPtr[resourceDocumentEnrichmentConfigurationData](ctx, &dec)
	if d.HasError() {
		return d
	}
	r.DocumentEnrichmentConfiguration = l
	return diags
}

func (r *resourceDatasourceData) flattenConfiguration(conf document.Interface) diag.Diagnostics {
	var diags diag.Diagnostics
	b, err := conf.MarshalSmithyDocument()
	if err != nil {
		diags.AddError("failed to marshal configuration", err.Error())
		return diags
	}
	var kv map[string]interface{}
	if err := json.Unmarshal(b, &kv); err != nil {
		diags.AddError("failed to unmarshal configuration", err.Error())
		return diags
	}
	delete(kv, "IngestionMode")
	b, err = json.Marshal(kv)
	if err != nil {
		diags.AddError("failed to marshal configuration", err.Error())
		return diags
	}
	r.Configuration = types.StringValue(string(b))
	return diags
}

func (r *resourceDocumentAttributeValueData) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics

	r.DateValue = timetypes.NewRFC3339Null()
	r.LongValue = types.Int64Null()
	r.StringValue = types.StringNull()
	r.StringListValue = fwtypes.NewListValueOfNull[types.String](ctx)

	switch t := v.(type) {
	case awstypes.DocumentAttributeValueMemberDateValue:
		r.DateValue = timetypes.NewRFC3339TimeValue(t.Value)
	case awstypes.DocumentAttributeValueMemberLongValue:
		r.LongValue = types.Int64Value(t.Value)
	case awstypes.DocumentAttributeValueMemberStringListValue:
		r.StringListValue = fwflex.FlattenFrameworkStringValueListOfString(ctx, t.Value)
	case awstypes.DocumentAttributeValueMemberStringValue:
		r.StringValue = types.StringValue(t.Value)
	}

	return diags
}

func flatValue(ctx context.Context, av awstypes.DocumentAttributeValue) *resourceDocumentAttributeValueData {
	rvd := resourceDocumentAttributeValueData{
		DateValue:       timetypes.NewRFC3339Null(),
		LongValue:       types.Int64Null(),
		StringValue:     types.StringNull(),
		StringListValue: fwtypes.NewListValueOfNull[types.String](ctx),
	}

	switch v := av.(type) {
	case *awstypes.DocumentAttributeValueMemberDateValue:
		rvd.DateValue = timetypes.NewRFC3339TimeValue(v.Value)
	case *awstypes.DocumentAttributeValueMemberLongValue:
		rvd.LongValue = types.Int64Value(v.Value)
	case *awstypes.DocumentAttributeValueMemberStringListValue:
		rvd.StringListValue = fwflex.FlattenFrameworkStringValueListOfString(ctx, v.Value)
	case *awstypes.DocumentAttributeValueMemberStringValue:
		rvd.StringValue = types.StringValue(v.Value)
	}
	return &rvd
}

func (r *resourceDatasourceData) expandToUpdateDataSourceInput(ctx context.Context) (*qbusiness.UpdateDataSourceInput, diag.Diagnostics) {
	var diags diag.Diagnostics
	input := &qbusiness.UpdateDataSourceInput{}

	input.ApplicationId = r.ApplicationId.ValueStringPointer()
	input.DataSourceId = r.DatasourceId.ValueStringPointer()
	input.DisplayName = r.DisplayName.ValueStringPointer()
	input.Description = r.Description.ValueStringPointer()
	input.DisplayName = r.DisplayName.ValueStringPointer()
	input.IndexId = r.IndexId.ValueStringPointer()
	input.RoleArn = r.RoleArn.ValueStringPointer()
	input.SyncSchedule = r.SyncSchedule.ValueStringPointer()

	if input.Configuration, diags = r.expandConfiguration(); diags.HasError() {
		return nil, diags
	}
	if input.DocumentEnrichmentConfiguration, diags = r.expandDocumentEnrichmentConfiguration(ctx); diags.HasError() {
		return nil, diags
	}
	if input.VpcConfiguration, diags = r.expandVPCConfiguration(ctx); diags.HasError() {
		return nil, diags
	}
	return input, nil
}

func (r *resourceDatasourceData) expandToCreateDataSourceInput(ctx context.Context) (*qbusiness.CreateDataSourceInput, diag.Diagnostics) {
	var diags diag.Diagnostics
	input := &qbusiness.CreateDataSourceInput{}

	input.ApplicationId = r.ApplicationId.ValueStringPointer()
	input.DisplayName = r.DisplayName.ValueStringPointer()
	input.Description = r.Description.ValueStringPointer()
	input.DisplayName = r.DisplayName.ValueStringPointer()
	input.IndexId = r.IndexId.ValueStringPointer()
	input.RoleArn = r.RoleArn.ValueStringPointer()
	input.SyncSchedule = r.SyncSchedule.ValueStringPointer()

	if input.Configuration, diags = r.expandConfiguration(); diags.HasError() {
		return nil, diags
	}
	if input.DocumentEnrichmentConfiguration, diags = r.expandDocumentEnrichmentConfiguration(ctx); diags.HasError() {
		return nil, diags
	}
	if input.VpcConfiguration, diags = r.expandVPCConfiguration(ctx); diags.HasError() {
		return nil, diags
	}
	return input, nil
}

func (r *resourceDatasourceData) expandVPCConfiguration(ctx context.Context) (*awstypes.DataSourceVpcConfiguration, diag.Diagnostics) {
	vpcConf := awstypes.DataSourceVpcConfiguration{}
	if r.VpcConfiguration.IsNull() {
		return nil, nil
	}
	conf, d := r.VpcConfiguration.ToPtr(ctx)
	if d.HasError() {
		return nil, d
	}
	if d := conf.SecurityGroupIds.ElementsAs(ctx, &vpcConf.SecurityGroupIds, false); d.HasError() {
		return nil, d
	}
	if d := conf.SubnetIds.ElementsAs(ctx, &vpcConf.SubnetIds, false); d.HasError() {
		return nil, d
	}
	return &vpcConf, nil
}

func (r *resourceDatasourceData) expandDocumentEnrichmentConfiguration(
	ctx context.Context) (*awstypes.DocumentEnrichmentConfiguration, diag.Diagnostics) {
	if r.DocumentEnrichmentConfiguration.IsNull() {
		return &awstypes.DocumentEnrichmentConfiguration{}, nil
	}

	dec := awstypes.DocumentEnrichmentConfiguration{}

	ic, diags := r.DocumentEnrichmentConfiguration.ToPtr(ctx)
	if diags.HasError() {
		return nil, diags
	}
	inlineConfs, diags := ic.InlineConfigurations.ToSlice(ctx)
	if diags.HasError() {
		return nil, diags
	}

	if dec.InlineConfigurations, diags = expInlineConfiguration(ctx, inlineConfs); diags.HasError() {
		return nil, diags
	}

	postHook, diags := ic.PostExtractionHookConfiguration.ToPtr(ctx)
	if diags.HasError() {
		return nil, diags
	}

	if dec.PostExtractionHookConfiguration, diags = expHookConfiguration(ctx, postHook); diags.HasError() {
		return nil, diags
	}

	preHook, diags := ic.PreExreactionHookConfiguration.ToPtr(ctx)
	if diags.HasError() {
		return nil, diags
	}
	if dec.PreExtractionHookConfiguration, diags = expHookConfiguration(ctx, preHook); diags.HasError() {
		return nil, diags
	}
	return &dec, diags
}

func expInlineConfiguration(ctx context.Context, conf []*resourceInlineDocumentEnrichmentConfigurationData) ([]awstypes.InlineDocumentEnrichmentConfiguration, diag.Diagnostics) {
	inlineDocConf := []awstypes.InlineDocumentEnrichmentConfiguration{}

	for _, c := range conf {
		var ic awstypes.InlineDocumentEnrichmentConfiguration

		ic.DocumentContentOperator = awstypes.DocumentContentOperator(c.Operator.ValueString())

		rcs, diags := c.Condition.ToPtr(ctx)
		if diags.HasError() {
			return nil, diags
		}
		if ic.Condition, diags = expDocumentAttributeCondition(ctx, rcs); diags.HasError() {
			return nil, diags
		}

		dat, diags := c.Target.ToPtr(ctx)
		if diags.HasError() {
			return nil, diags
		}
		if ic.Target, diags = expDocumentAttributeTarget(ctx, dat); diags.HasError() {
			return nil, diags
		}
		inlineDocConf = append(inlineDocConf, ic)
	}

	return inlineDocConf, nil
}

func expDocumentAttributeTarget(ctx context.Context, conf *resourceDocumentAttributeTargetData) (*awstypes.DocumentAttributeTarget, diag.Diagnostics) {
	if conf == nil {
		return nil, nil
	}

	var diags diag.Diagnostics
	var dat awstypes.DocumentAttributeTarget

	dat.Key = conf.Key.ValueStringPointer()
	dat.AttributeValueOperator = awstypes.AttributeValueOperator(conf.Operator.ValueString())

	val, diags := conf.Value.ToPtr(ctx)
	if diags.HasError() {
		return nil, diags
	}

	dat.Value, diags = expValue(ctx, val)
	if diags.HasError() {
		return nil, diags
	}
	return &dat, nil
}

func expHookConfiguration(ctx context.Context, conf *resourceHookConfigurationData) (*awstypes.HookConfiguration, diag.Diagnostics) {
	if conf == nil {
		return &awstypes.HookConfiguration{}, nil
	}

	var hookConf awstypes.HookConfiguration
	if !conf.Condition.IsNull() {
		c, d := conf.Condition.ToPtr(ctx)
		if d.HasError() {
			return nil, d
		}
		if hookConf.InvocationCondition, d = expDocumentAttributeCondition(ctx, c); d.HasError() {
			return nil, d
		}
	}
	hookConf.LambdaArn = conf.LambdaARN.ValueStringPointer()
	hookConf.RoleArn = conf.RoleARN.ValueStringPointer()
	hookConf.S3BucketName = conf.S3BucketName.ValueStringPointer()
	return &hookConf, nil
}

func expDocumentAttributeCondition(ctx context.Context, conf *resourceConditionData) (*awstypes.DocumentAttributeCondition, diag.Diagnostics) {
	if conf == nil {
		return nil, nil
	}

	c, d := conf.Value.ToPtr(ctx)
	if d.HasError() {
		return nil, d
	}
	v, d := expValue(ctx, c)
	if d.HasError() {
		return nil, d
	}
	cond := awstypes.DocumentAttributeCondition{
		Key:      conf.Key.ValueStringPointer(),
		Operator: awstypes.DocumentEnrichmentConditionOperator(conf.Operator.ValueString()),
		Value:    v,
	}
	return &cond, nil
}

func (r resourceDocumentAttributeValueData) Expand(ctx context.Context) (any, diag.Diagnostics) {
	return expValue(ctx, &r)
}

func expValue(ctx context.Context, rvd *resourceDocumentAttributeValueData) (awstypes.DocumentAttributeValue, diag.Diagnostics) {
	if rvd == nil {
		return nil, nil
	}
	var diag diag.Diagnostics
	if !rvd.DateValue.IsNull() {
		tv, diag := rvd.DateValue.ValueRFC3339Time()
		if diag.HasError() {
			return nil, diag
		}
		return &awstypes.DocumentAttributeValueMemberDateValue{
			Value: tv,
		}, nil
	}
	if !rvd.LongValue.IsNull() {
		return &awstypes.DocumentAttributeValueMemberLongValue{
			Value: rvd.LongValue.ValueInt64(),
		}, nil
	}
	if !rvd.StringListValue.IsNull() {
		var l []string
		if diag = rvd.StringListValue.ElementsAs(ctx, &l, false); diag.HasError() {
			return nil, diag
		}
		return &awstypes.DocumentAttributeValueMemberStringListValue{
			Value: l,
		}, nil
	}
	if !rvd.StringValue.IsNull() {
		return &awstypes.DocumentAttributeValueMemberStringValue{
			Value: rvd.StringValue.ValueString(),
		}, nil
	}
	return nil, nil
}

const (
	datasourceResourceIDPartCount = 3
)

func (r *resourceDatasourceData) setID() diag.Diagnostics {
	var diags diag.Diagnostics

	id, err := flex.FlattenResourceId([]string{
		r.ApplicationId.ValueString(),
		r.IndexId.ValueString(),
		r.DatasourceId.ValueString(),
	}, datasourceResourceIDPartCount, false)
	if err != nil {
		diags.AddError(
			create.ProblemStandardMessage(names.QBusiness, create.ErrActionFlatteningResourceId, ResNameDatasource, id, err),
			err.Error())
		return diags
	}
	r.ID = types.StringValue(id)
	return diags
}

func FindDatasourceByID(ctx context.Context, conn *qbusiness.Client, id string) (*qbusiness.GetDataSourceOutput, error) {
	parts, err := flex.ExpandResourceId(id, datasourceResourceIDPartCount, false)

	if err != nil {
		return nil, err
	}

	input := qbusiness.GetDataSourceInput{
		ApplicationId: aws.String(parts[0]),
		IndexId:       aws.String(parts[1]),
		DataSourceId:  aws.String(parts[2]),
	}

	output, err := conn.GetDataSource(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}
	return output, nil
}
