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
	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
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

func authConfigurationSchema(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[authConfigurationData](ctx),
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"role_arn": schema.StringAttribute{
					Description: "ARN of an IAM role used by Amazon Q to access the basic authentication credentials stored in a Secrets Manager secret.",
					Required:    true,
					Validators: []validator.String{
						stringvalidator.RegexMatches(regexache.MustCompile(`^arn:[a-z0-9-\.]{1,63}:[a-z0-9-\.]{0,63}:[a-z0-9-\.]{0,63}:[a-z0-9-\.]{0,63}:[^/].{0,1023}$`), "must be valid ARN"),
					},
				},
				"secret_arn": schema.StringAttribute{
					Description: "ARN of the Secrets Manager secret that stores the basic authentication credentials used for plugin configuration.",
					Required:    true,
					Validators: []validator.String{
						stringvalidator.RegexMatches(regexache.MustCompile(`^arn:[a-z0-9-\.]{1,63}:[a-z0-9-\.]{0,63}:[a-z0-9-\.]{0,63}:[a-z0-9-\.]{0,63}:[^/].{0,1023}$`), "must be valid ARN"),
					},
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
			"application_id": schema.StringAttribute{
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
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 1000),
					stringvalidator.RegexMatches(regexache.MustCompile(`^\P{C}*$`), "must not contain control characters"),
				},
			},
			names.AttrState: schema.StringAttribute{
				CustomType:  fwtypes.StringEnumType[awstypes.PluginState](),
				Optional:    true,
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
			"basic_auth_configuration":               authConfigurationSchema(ctx),
			"oauth2_client_credential_configuration": authConfigurationSchema(ctx),
			"no_auth_configuration": schema.SingleNestedBlock{
				CustomType: fwtypes.NewObjectTypeOf[noAuthConfigurationData](ctx),
			},
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
								stringvalidator.LengthBetween(0, 200),
								stringvalidator.RegexMatches(regexache.MustCompile(`^\P{C}*$`), "must not contain control characters"),
							},
						},
						"api_schema_type": schema.StringAttribute{
							CustomType:  fwtypes.StringEnumType[awstypes.APISchemaType](),
							Required:    true,
							Description: "The type of OpenAPI schema to use.",
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
									"bucket": schema.StringAttribute{
										Description: "The name of the S3 bucket where the OpenAPI schema is stored.",
										Required:    true,
										Validators: []validator.String{
											stringvalidator.LengthBetween(1, 63),
											stringvalidator.RegexMatches(regexache.MustCompile(`^[a-z0-9][\.\-a-z0-9]{1,61}[a-z0-9]$`), "must be a valid bucket name"),
										},
									},
									"key": schema.StringAttribute{
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
	resp.Diagnostics.Append(req.Plan.Get(ctx, data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().QBusinessClient(ctx)
	input := &qbusiness.CreatePluginInput{}

	resp.Diagnostics.Append(fwflex.Expand(ctx, data, input)...)
	if resp.Diagnostics.HasError() {
		return
	}

	authConf, d := data.expandAuthConfiguration(ctx)
	if d.HasError() {
		resp.Diagnostics.Append(d...)
		return
	}
	input.AuthConfiguration = authConf

	if input.CustomPluginConfiguration != nil {
		apiSchema, d := data.expandApiSchema(ctx)
		if d.HasError() {
			resp.Diagnostics.Append(d...)
			return
		}
		input.CustomPluginConfiguration.ApiSchema = apiSchema
	}

	output, err := conn.CreatePlugin(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError("failed to create Q Business plugin", err.Error())
		return
	}

	data.PluginId = fwflex.StringToFramework(ctx, output.PluginId)
	data.PluginArn = fwflex.StringToFramework(ctx, output.PluginArn)

	data.setID()

	if _, err := waitPluginCreated(ctx, conn, data.PluginId.ValueString(), r.CreateTimeout(ctx, data.Timeouts)); err != nil {
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
	if out.CustomPluginConfiguration != nil {
		data.flattenApiSchema(ctx, out.CustomPluginConfiguration.ApiSchema)
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *resourcePlugin) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var old, new resourcePluginData
	resp.Diagnostics.Append(req.Plan.Get(ctx, new)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(req.State.Get(ctx, &old)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !old.BasicAuthConfiguration.Equal(new.BasicAuthConfiguration) ||
		!old.OAuth2ClientCredentialConfiguration.Equal(new.OAuth2ClientCredentialConfiguration) ||
		!old.NoAuthConfiguration.Equal(new.NoAuthConfiguration) ||
		!old.CustomPluginConfiguration.Equal(new.CustomPluginConfiguration) ||
		!old.DisplayName.Equal(new.DisplayName) ||
		!old.ServerURL.Equal(new.ServerURL) ||
		!old.State.Equal(new.State) {
		conn := r.Meta().QBusinessClient(ctx)

		input := &qbusiness.UpdatePluginInput{}

		authConf, d := new.expandAuthConfiguration(ctx)
		if d.HasError() {
			resp.Diagnostics.Append(d...)
			return
		}
		input.AuthConfiguration = authConf

		if input.CustomPluginConfiguration != nil {
			apiSchema, d := new.expandApiSchema(ctx)
			if d.HasError() {
				resp.Diagnostics.Append(d...)
				return
			}
			input.CustomPluginConfiguration.ApiSchema = apiSchema
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

func (r *resourcePlugin) ConfigValidators(_ context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcevalidator.ExactlyOneOf(
			path.MatchRoot("basic_auth_configuration"),
			path.MatchRoot("oauth2_client_credential_configuration"),
			path.MatchRoot("no_auth_configuration"),
		),
	}
}

type resourcePluginData struct {
	ApplicationId                       types.String                                                   `tfsdk:"application_id"`
	PluginArn                           types.String                                                   `tfsdk:"arn"`
	ID                                  types.String                                                   `tfsdk:"id"`
	Tags                                types.Map                                                      `tfsdk:"tags"`
	TagsAll                             types.Map                                                      `tfsdk:"tags_all"`
	Timeouts                            timeouts.Value                                                 `tfsdk:"timeouts"`
	DisplayName                         types.String                                                   `tfsdk:"display_name"`
	PluginId                            types.String                                                   `tfsdk:"plugin_id"`
	Type                                fwtypes.StringEnum[awstypes.PluginType]                        `tfsdk:"type"`
	BasicAuthConfiguration              fwtypes.ListNestedObjectValueOf[authConfigurationData]         `tfsdk:"basic_auth_configuration"`
	OAuth2ClientCredentialConfiguration fwtypes.ListNestedObjectValueOf[authConfigurationData]         `tfsdk:"oauth2_client_credential_configuration"`
	NoAuthConfiguration                 fwtypes.ObjectValueOf[noAuthConfigurationData]                 `tfsdk:"no_auth_configuration"`
	CustomPluginConfiguration           fwtypes.ListNestedObjectValueOf[customPluginConfigurationData] `tfsdk:"custom_plugin_configuration"`
	ServerURL                           types.String                                                   `tfsdk:"server_url"`
	State                               fwtypes.StringEnum[awstypes.PluginState]                       `tfsdk:"state"`
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

type noAuthConfigurationData struct {
	Description types.String `tfsdk:"description"`
}

type s3Data struct {
	Bucket types.String `tfsdk:"bucket"`
	Key    types.String `tfsdk:"key"`
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
	case *awstypes.PluginAuthConfigurationMemberNoAuthConfiguration:
		r.NoAuthConfiguration = fwtypes.NewObjectValueOfMust(ctx, &noAuthConfigurationData{})
	}
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

	if !r.NoAuthConfiguration.IsNull() {
		return &awstypes.PluginAuthConfigurationMemberNoAuthConfiguration{}, diags
	}

	return nil, diags
}

func (r *resourcePluginData) flattenApiSchema(ctx context.Context, apiSchema awstypes.APISchema) {
	switch v := apiSchema.(type) {
	case *awstypes.APISchemaMemberPayload:
		r.CustomPluginConfiguration = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &customPluginConfigurationData{
			Payload: fwflex.StringValueToFramework(ctx, v.Value),
		})
	case *awstypes.APISchemaMemberS3:
		r.CustomPluginConfiguration = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &customPluginConfigurationData{
			S3: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &s3Data{
				Bucket: fwflex.StringToFramework(ctx, v.Value.Bucket),
				Key:    fwflex.StringToFramework(ctx, v.Value.Key),
			}),
		})
	}
}

func (r *resourcePluginData) expandApiSchema(ctx context.Context) (awstypes.APISchema, diag.Diagnostics) {
	var diags diag.Diagnostics

	if !r.CustomPluginConfiguration.IsNull() {
		pluginConf, d := r.CustomPluginConfiguration.ToPtr(ctx)
		if d.HasError() {
			return nil, d
		}

		if !pluginConf.Payload.IsNull() {
			return &awstypes.APISchemaMemberPayload{
				Value: pluginConf.Payload.ValueString(),
			}, diags
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
			}, diags
		}
	}
	return nil, diags
}

const (
	pluginResourceIDPartCount = 2
)

func (data *resourcePluginData) setID() {
	data.ID = types.StringValue(errs.Must(flex.FlattenResourceId([]string{data.ApplicationId.ValueString(), data.PluginId.ValueString()}, pluginResourceIDPartCount, false)))
}

func (data *resourcePluginData) initFromID() error {
	parts, err := flex.ExpandResourceId(data.ID.ValueString(), pluginResourceIDPartCount, false)
	if err != nil {
		return err
	}

	data.ApplicationId = types.StringValue(parts[0])
	data.PluginId = types.StringValue(parts[1])
	return nil
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
