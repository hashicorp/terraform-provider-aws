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

// @FrameworkResource("aws_qbusiness_plugin", name="Plugin")
// @Tags(identifierAttribute="arn")
func newResourcePlugin(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourcePlugin{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNamePlugin = "Plugin"
)

type resourcePlugin struct {
	framework.ResourceWithConfigure
	framework.WithImportByID
	framework.WithTimeouts
}

func (r *resourcePlugin) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_qbusiness_plugin"
}

func authConfigurationSchema(ctx context.Context, conflictsWith string) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[authConfigurationData](ctx),
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
			listvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName(conflictsWith)),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				names.AttrRoleARN: schema.StringAttribute{
					CustomType:  fwtypes.ARNType,
					Description: "ARN of an IAM role used by Amazon Q to access the basic authentication credentials stored in a Secrets Manager secret.",
					Required:    true,
				},
				"secret_arn": schema.StringAttribute{
					CustomType:  fwtypes.ARNType,
					Description: "ARN of the Secrets Manager secret that stores the basic authentication credentials used for plugin configuration.",
					Required:    true,
				},
			},
		},
	}
}

func (r *resourcePlugin) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID:  framework.IDAttribute(),
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrApplicationID: schema.StringAttribute{
				Description: "Identifier of the Amazon Q application associated with the plugin.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexache.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9-]{35}$`), "must be a valid application ID"),
				},
			},
			"plugin_id": framework.IDAttribute(),
			names.AttrDisplayName: schema.StringAttribute{
				Description: "The display name of the Amazon Q plugin.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 1000),
					stringvalidator.RegexMatches(regexache.MustCompile(`^\P{C}*$`), "must not contain control characters"),
				},
			},
			names.AttrState: schema.StringAttribute{
				CustomType:  fwtypes.StringEnumType[awstypes.PluginState](),
				Required:    true,
				Description: "The state of the Amazon Q plugin.",
				Validators: []validator.String{
					enum.FrameworkValidate[awstypes.PluginState](),
				},
			},
			"server_url": schema.StringAttribute{
				Description: "Source URL used for plugin configuration.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 2048),
					stringvalidator.RegexMatches(regexache.MustCompile(`^(https?|ftp|file)://([^\s]*)$`), "must be a valid server URL"),
				},
			},
			names.AttrType: schema.StringAttribute{
				CustomType:  fwtypes.StringEnumType[awstypes.PluginType](),
				Required:    true,
				Description: "Type of plugin. Valid value are `SERVICE_NOW`, `SALESFORCE`, `JIRA`, `ZENDESK`, `CUSTOM`",
				Validators: []validator.String{
					enum.FrameworkValidate[awstypes.PluginType](),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			"basic_auth_configuration":               authConfigurationSchema(ctx, "oauth2_client_credential_configuration"),
			"oauth2_client_credential_configuration": authConfigurationSchema(ctx, "basic_auth_configuration"),
			"custom_plugin_configuration": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[customPluginConfigurationData](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrDescription: schema.StringAttribute{
							Description: "A description for your custom plugin configuration.",
							Required:    true,
							Validators: []validator.String{
								stringvalidator.LengthBetween(1, 200),
								stringvalidator.RegexMatches(regexache.MustCompile(`^\P{C}*$`), "must not contain control characters"),
							},
						},
						"api_schema_type": schema.StringAttribute{
							CustomType:  fwtypes.StringEnumType[awstypes.APISchemaType](),
							Required:    true,
							Description: "The type of OpenAPI schema to use. Valid value is `OPEN_API_V3`.",
							Validators: []validator.String{
								enum.FrameworkValidate[awstypes.APISchemaType](),
							},
						},
						"payload": schema.StringAttribute{
							Description: "The JSON or YAML-formatted payload defining the OpenAPI schema for a custom plugin.",
							Optional:    true,
							Validators: []validator.String{
								stringvalidator.AtLeastOneOf(
									path.MatchRelative().AtParent().AtName("s3"),
									path.MatchRelative().AtParent().AtName("payload"),
								),
								stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("s3")),
								stringvalidator.LengthAtLeast(1),
							},
						},
					},
					Blocks: map[string]schema.Block{
						"s3": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[s3Data](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
								listvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("payload")),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrBucket: schema.StringAttribute{
										Description: "The name of the S3 bucket where the OpenAPI schema is stored.",
										Required:    true,
										Validators: []validator.String{
											stringvalidator.LengthBetween(1, 63),
											stringvalidator.RegexMatches(regexache.MustCompile(`^[a-z0-9][\.\-a-z0-9]{1,61}[a-z0-9]$`), "must be a valid bucket name"),
										},
									},
									names.AttrKey: schema.StringAttribute{
										Description: "The key of the OpenAPI schema object in the S3 bucket.",
										Required:    true,
										Validators: []validator.String{
											stringvalidator.LengthBetween(1, 1024),
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
				Delete: true,
				Update: true,
			}),
		},
	}
}

func (r *resourcePlugin) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data resourcePluginData
	var diags diag.Diagnostics
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().QBusinessClient(ctx)
	input := &qbusiness.CreatePluginInput{}

	resp.Diagnostics.Append(fwflex.Expand(ctx, data, input)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if input.AuthConfiguration, diags = data.expandAuthConfiguration(ctx); diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	if input.CustomPluginConfiguration, diags = data.expandPluginConfiguration(ctx); diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	input.Tags = getTagsIn(ctx)
	input.ClientToken = aws.String(id.UniqueId())

	output, err := conn.CreatePlugin(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError("failed to create Q Business plugin", err.Error())
		return
	}

	data.PluginId = fwflex.StringToFramework(ctx, output.PluginId)
	data.PluginArn = fwflex.StringToFramework(ctx, output.PluginArn)

	data.setID()

	if _, err := waitPluginCreated(ctx, conn, data.ID.ValueString(), r.CreateTimeout(ctx, data.Timeouts)); err != nil {
		resp.Diagnostics.AddError("failed to wait for Q Business plugin to be created", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *resourcePlugin) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data resourcePluginData
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().QBusinessClient(ctx)
	out, err := FindPluginByID(ctx, conn, data.ID.ValueString())

	if tfresource.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("failed to retrieve Q Business plugin (%s)", data.ID.ValueString()), err.Error())
		return
	}

	resp.Diagnostics.Append(fwflex.Flatten(ctx, out, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.flattenAuthConfiguration(ctx, out.AuthConfiguration)
	data.flattenPluginConfiguration(ctx, out.CustomPluginConfiguration)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *resourcePlugin) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var old, new resourcePluginData
	var diags diag.Diagnostics
	resp.Diagnostics.Append(req.Plan.Get(ctx, &new)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(req.State.Get(ctx, &old)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !old.BasicAuthConfiguration.Equal(new.BasicAuthConfiguration) ||
		!old.OAuth2ClientCredentialConfiguration.Equal(new.OAuth2ClientCredentialConfiguration) ||
		!old.PluginConfiguration.Equal(new.PluginConfiguration) ||
		!old.DisplayName.Equal(new.DisplayName) ||
		!old.ServerURL.Equal(new.ServerURL) ||
		!old.State.Equal(new.State) {
		conn := r.Meta().QBusinessClient(ctx)

		input := &qbusiness.UpdatePluginInput{
			ApplicationId: new.ApplicationId.ValueStringPointer(),
			PluginId:      new.PluginId.ValueStringPointer(),
			DisplayName:   new.DisplayName.ValueStringPointer(),
			ServerUrl:     new.ServerURL.ValueStringPointer(),
			State:         awstypes.PluginState(new.State.ValueString()),
		}

		if input.AuthConfiguration, diags = new.expandAuthConfiguration(ctx); diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}

		if input.CustomPluginConfiguration, diags = new.expandPluginConfiguration(ctx); diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}

		_, err := conn.UpdatePlugin(ctx, input)

		if err != nil {
			resp.Diagnostics.AddError("failed to update Q Business plugin", err.Error())
			return
		}

		if _, err := waitPluginUpdated(ctx, conn, new.ID.ValueString(), r.UpdateTimeout(ctx, new.Timeouts)); err != nil {
			resp.Diagnostics.AddError("failed to wait for Q Business plugin to be updated", err.Error())
			return
		}
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &new)...)
}

func (r *resourcePlugin) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data resourcePluginData
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().QBusinessClient(ctx)

	input := &qbusiness.DeletePluginInput{
		ApplicationId: aws.String(data.ApplicationId.ValueString()),
		PluginId:      aws.String(data.PluginId.ValueString()),
	}

	_, err := conn.DeletePlugin(ctx, input)

	if err != nil {
		var nfe *awstypes.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return
		}
		resp.Diagnostics.AddError(fmt.Sprintf("failed to delete Q Business plugin (%s)", data.PluginId.ValueString()), err.Error())
		return
	}

	if _, err := waitPluginDeleted(ctx, conn, data.ID.ValueString(), r.DeleteTimeout(ctx, data.Timeouts)); err != nil {
		resp.Diagnostics.AddError("failed to wait for Q Business plugin to be deleted", err.Error())
		return
	}
}

func (r *resourcePlugin) ModifyPlan(ctx context.Context, request resource.ModifyPlanRequest, response *resource.ModifyPlanResponse) {
	r.SetTagsAll(ctx, request, response)
}

type resourcePluginData struct {
	ApplicationId                       types.String                                                   `tfsdk:"application_id"`
	BasicAuthConfiguration              fwtypes.ListNestedObjectValueOf[authConfigurationData]         `tfsdk:"basic_auth_configuration"`
	DisplayName                         types.String                                                   `tfsdk:"display_name"`
	ID                                  types.String                                                   `tfsdk:"id"`
	OAuth2ClientCredentialConfiguration fwtypes.ListNestedObjectValueOf[authConfigurationData]         `tfsdk:"oauth2_client_credential_configuration"`
	PluginArn                           types.String                                                   `tfsdk:"arn"`
	PluginConfiguration                 fwtypes.ListNestedObjectValueOf[customPluginConfigurationData] `tfsdk:"custom_plugin_configuration"`
	PluginId                            types.String                                                   `tfsdk:"plugin_id"`
	Type                                fwtypes.StringEnum[awstypes.PluginType]                        `tfsdk:"type"`
	ServerURL                           types.String                                                   `tfsdk:"server_url"`
	State                               fwtypes.StringEnum[awstypes.PluginState]                       `tfsdk:"state"`
	Tags                                types.Map                                                      `tfsdk:"tags"`
	TagsAll                             types.Map                                                      `tfsdk:"tags_all"`
	Timeouts                            timeouts.Value                                                 `tfsdk:"timeouts"`
}

type authConfigurationData struct {
	RoleArn   types.String `tfsdk:"role_arn"`
	SecretArn types.String `tfsdk:"secret_arn"`
}

type customPluginConfigurationData struct {
	Description   types.String                               `tfsdk:"description"`
	ApiSchemaType fwtypes.StringEnum[awstypes.APISchemaType] `tfsdk:"api_schema_type"`
	S3            fwtypes.ListNestedObjectValueOf[s3Data]    `tfsdk:"s3"`
	Payload       types.String                               `tfsdk:"payload"`
}

type s3Data struct {
	Bucket types.String `tfsdk:"bucket"`
	Key    types.String `tfsdk:"key"`
}

func (r *resourcePluginData) expandAuthConfiguration(ctx context.Context) (awstypes.PluginAuthConfiguration, diag.Diagnostics) {
	var diags diag.Diagnostics

	if !r.BasicAuthConfiguration.IsNull() {
		basicAuthConf, d := r.BasicAuthConfiguration.ToPtr(ctx)
		if d.HasError() {
			return nil, d
		}

		return &awstypes.PluginAuthConfigurationMemberBasicAuthConfiguration{
			Value: awstypes.BasicAuthConfiguration{
				RoleArn:   basicAuthConf.RoleArn.ValueStringPointer(),
				SecretArn: basicAuthConf.SecretArn.ValueStringPointer(),
			},
		}, diags
	}

	if !r.OAuth2ClientCredentialConfiguration.IsNull() {
		oauth2Conf, d := r.OAuth2ClientCredentialConfiguration.ToPtr(ctx)
		if d.HasError() {
			return nil, d
		}

		return &awstypes.PluginAuthConfigurationMemberOAuth2ClientCredentialConfiguration{
			Value: awstypes.OAuth2ClientCredentialConfiguration{
				RoleArn:   oauth2Conf.RoleArn.ValueStringPointer(),
				SecretArn: oauth2Conf.SecretArn.ValueStringPointer(),
			},
		}, diags
	}

	return &awstypes.PluginAuthConfigurationMemberNoAuthConfiguration{}, diags
}

func (r *resourcePluginData) flattenAuthConfiguration(ctx context.Context, authConf awstypes.PluginAuthConfiguration) {
	switch v := authConf.(type) {
	case *awstypes.PluginAuthConfigurationMemberBasicAuthConfiguration:
		r.BasicAuthConfiguration = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &authConfigurationData{
			RoleArn:   fwflex.StringToFramework(ctx, v.Value.RoleArn),
			SecretArn: fwflex.StringToFramework(ctx, v.Value.SecretArn),
		})
	case *awstypes.PluginAuthConfigurationMemberOAuth2ClientCredentialConfiguration:
		r.OAuth2ClientCredentialConfiguration = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &authConfigurationData{
			RoleArn:   fwflex.StringToFramework(ctx, v.Value.RoleArn),
			SecretArn: fwflex.StringToFramework(ctx, v.Value.SecretArn),
		})
	}
}

func (r *resourcePluginData) expandPluginConfiguration(ctx context.Context) (*awstypes.CustomPluginConfiguration, diag.Diagnostics) {
	if r.PluginConfiguration.IsNull() {
		return nil, nil
	}
	pluginConf, d := r.PluginConfiguration.ToPtr(ctx)
	if d.HasError() {
		return nil, d
	}

	schema, d := r.expandAPISchema(ctx, pluginConf)
	if d.HasError() {
		return nil, d
	}

	return &awstypes.CustomPluginConfiguration{
		Description:   pluginConf.Description.ValueStringPointer(),
		ApiSchemaType: awstypes.APISchemaType(pluginConf.ApiSchemaType.ValueString()),
		ApiSchema:     schema,
	}, nil
}

func (r *resourcePluginData) flattenPluginConfiguration(ctx context.Context, pluginConf *awstypes.CustomPluginConfiguration) {
	if pluginConf == nil {
		return
	}

	pc := customPluginConfigurationData{
		Description:   fwflex.StringToFramework(ctx, pluginConf.Description),
		ApiSchemaType: fwtypes.StringEnumValue[awstypes.APISchemaType](pluginConf.ApiSchemaType),
	}
	switch v := pluginConf.ApiSchema.(type) {
	case *awstypes.APISchemaMemberPayload:
		pc.Payload = fwflex.StringValueToFramework(ctx, v.Value)
		pc.S3 = fwtypes.NewListNestedObjectValueOfNull[s3Data](ctx)
	case *awstypes.APISchemaMemberS3:
		pc.S3 = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &s3Data{
			Bucket: fwflex.StringToFramework(ctx, v.Value.Bucket),
			Key:    fwflex.StringToFramework(ctx, v.Value.Key),
		})
	}
	r.PluginConfiguration = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &pc)
}

func (r *resourcePluginData) expandAPISchema(ctx context.Context, pluginConf *customPluginConfigurationData) (awstypes.APISchema, diag.Diagnostics) {
	if !pluginConf.Payload.IsNull() {
		return &awstypes.APISchemaMemberPayload{
			Value: pluginConf.Payload.ValueString(),
		}, nil
	}
	if !pluginConf.S3.IsNull() {
		s3Conf, d := pluginConf.S3.ToPtr(ctx)
		if d.HasError() {
			return nil, d
		}
		return &awstypes.APISchemaMemberS3{
			Value: awstypes.S3{
				Bucket: s3Conf.Bucket.ValueStringPointer(),
				Key:    s3Conf.Key.ValueStringPointer(),
			},
		}, nil
	}
	return nil, nil
}

const (
	pluginResourceIDPartCount = 2
)

func (r *resourcePluginData) setID() {
	r.ID = types.StringValue(errs.Must(flex.FlattenResourceId([]string{r.ApplicationId.ValueString(), r.PluginId.ValueString()}, pluginResourceIDPartCount, false)))
}

func FindPluginByID(ctx context.Context, conn *qbusiness.Client, id string) (*qbusiness.GetPluginOutput, error) {
	parts, err := flex.ExpandResourceId(id, pluginResourceIDPartCount, false)

	if err != nil {
		return nil, err
	}

	input := &qbusiness.GetPluginInput{
		ApplicationId: aws.String(parts[0]),
		PluginId:      aws.String(parts[1]),
	}

	output, err := conn.GetPlugin(ctx, input)

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
