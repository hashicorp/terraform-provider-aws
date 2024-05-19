// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package qbusiness

import (
	"context"
	"fmt"
	"log"
	"strings"
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
	sdkschema "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
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
}

func (r *resourcePlugin) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {

}

func (r *resourcePlugin) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
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

func (data *resourcePluginData) setID() {
	data.ID = types.StringValue(errs.Must(flex.FlattenResourceId([]string{data.ApplicationId.ValueString(), data.PluginId.ValueString()}, pluginResourceIDPartCount, false)))
}

func authConfigurationSchema1() *sdkschema.Schema {
	return &sdkschema.Schema{
		Type:         sdkschema.TypeList,
		Optional:     true,
		MaxItems:     1,
		ExactlyOneOf: []string{"basic_auth_configuration", "oauth2_client_credential_configuration"},
		Elem: &sdkschema.Resource{
			Schema: map[string]*sdkschema.Schema{
				"role_arn": {
					Type:         sdkschema.TypeString,
					Required:     true,
					Description:  "ARN of an IAM role used by Amazon Q to access the basic authentication credentials stored in a Secrets Manager secret.",
					ValidateFunc: verify.ValidARN,
				},
				"secret_arn": {
					Type:         sdkschema.TypeString,
					Required:     true,
					Description:  "ARN of the Secrets Manager secret that stores the basic authentication credentials used for plugin configuration.",
					ValidateFunc: verify.ValidARN,
				},
			},
		},
	}
}

func ResourcePlugin() *sdkschema.Resource {
	return &sdkschema.Resource{

		Importer: &sdkschema.ResourceImporter{
			StateContext: sdkschema.ImportStatePassthroughContext,
		},

		CustomizeDiff: verify.SetTagsDiff,

		Schema: map[string]*sdkschema.Schema{
			"application_id": {
				Type:        sdkschema.TypeString,
				Required:    true,
				Description: "Identifier of the Amazon Q application associated with the plugin.",
				ValidateFunc: validation.All(
					validation.StringMatch(regexache.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9-]{35}$`), "must be a valid application ID"),
				),
			},
			"arn": {
				Type:        sdkschema.TypeString,
				Computed:    true,
				Description: "ARN of the Amazon Q plugin.",
			},
			"basic_auth_configuration":               authConfigurationSchema1(),
			"oauth2_client_credential_configuration": authConfigurationSchema1(),
			"display_name": {
				Type:        sdkschema.TypeString,
				Required:    true,
				Description: "The name of the Amazon Q plugin.",
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 100),
					validation.StringMatch(regexache.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]*$`), "must begin with a letter or number and contain only alphanumeric, underscore, or hyphen characters"),
				),
			},
			"plugin_id": {
				Type:        sdkschema.TypeString,
				Computed:    true,
				Description: "The identifier of the Amazon Q plugin.",
			},
			"server_url": {
				Type:        sdkschema.TypeString,
				Required:    true,
				Description: "Source URL used for plugin configuration.",
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 2048),
					validation.StringMatch(regexache.MustCompile(`^(https?|ftp|file)://([^\s]*)$`), "must be a valid URL"),
				),
			},
			"state": {
				Type:             sdkschema.TypeString,
				Required:         true,
				Description:      "State of plugin. Valid value are `ENABLED` and `DISABLED`",
				ValidateDiagFunc: enum.Validate[awstypes.PluginState](),
			},
			"type": {
				Type:             sdkschema.TypeString,
				Required:         true,
				ForceNew:         true,
				Description:      "Type of plugin. Valid value are `SERVICE_NOW`, `SALESFORCE`, `JIRA`, and `ZENDESK`",
				ValidateDiagFunc: enum.Validate[awstypes.PluginType](),
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func resourcePluginCreate(ctx context.Context, d *sdkschema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).QBusinessClient(ctx)

	application_id := d.Get("application_id").(string)

	input := &qbusiness.CreatePluginInput{
		ApplicationId: aws.String(application_id),
		DisplayName:   aws.String(d.Get("display_name").(string)),
		ServerUrl:     aws.String(d.Get("server_url").(string)),
		Type:          awstypes.PluginType(d.Get("type").(string)),
		Tags:          getTagsIn(ctx),
	}

	if v, ok := d.GetOk("basic_auth_configuration"); ok {
		input.AuthConfiguration = expandBasicAuthConfiguration(v.([]interface{}))
	}

	if v, ok := d.GetOk("oauth2_client_credential_configuration"); ok {
		input.AuthConfiguration = expandOAuth2ClientCredentialConfiguration(v.([]interface{}))
	}

	output, err := conn.CreatePlugin(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating qbusiness plugin: %s", err)
	}

	d.SetId(application_id + "/" + aws.ToString(output.PluginId))

	updateInput := &qbusiness.UpdatePluginInput{
		ApplicationId: aws.String(application_id),
		PluginId:      output.PluginId,
		State:         awstypes.PluginState(d.Get("state").(string)),
	}

	_, err = conn.UpdatePlugin(ctx, updateInput)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating qbusiness plugin: %s", err)
	}

	return append(diags, resourcePluginRead(ctx, d, meta)...)
}

func resourcePluginRead(ctx context.Context, d *sdkschema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).QBusinessClient(ctx)

	output, err := FindPluginByID(ctx, conn, d.Id())

	if !d.IsNewResource() && errs.IsA[*awstypes.ResourceNotFoundException](err) {
		log.Printf("[WARN] qbusiness plugin (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading qbusiness plugin (%s): %s", d.Id(), err)
	}

	d.Set("application_id", output.ApplicationId)
	d.Set("arn", output.PluginArn)
	d.Set("display_name", output.DisplayName)

	switch v := output.AuthConfiguration.(type) {
	case *awstypes.PluginAuthConfigurationMemberBasicAuthConfiguration:
		if err := d.Set("basic_auth_configuration", flattenBasicAuthConfiguration(&v.Value)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting qbusiness plugin basic_auth_configuration: %s", err)
		}
	case *awstypes.PluginAuthConfigurationMemberOAuth2ClientCredentialConfiguration:
		if err := d.Set("oauth2_client_credential_configuration", flattenOAuth2ClientCredentialConfiguration(&v.Value)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting qbusiness plugin oauth2_client_credential_configuration: %s", err)
		}
	}

	d.Set("plugin_id", output.PluginId)
	d.Set("server_url", output.ServerUrl)
	d.Set("state", output.State)
	d.Set("type", output.Type)

	return diags
}

func resourcePluginUpdate(ctx context.Context, d *sdkschema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).QBusinessClient(ctx)

	application_id, plugin_id, err := parsePluginID(d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "parsing qbusiness plugin ID (%s): %s", d.Id(), err)
	}

	input := &qbusiness.UpdatePluginInput{
		ApplicationId: aws.String(application_id),
		PluginId:      aws.String(plugin_id),
	}

	if d.HasChange("display_name") {
		input.DisplayName = aws.String(d.Get("display_name").(string))
	}
	if d.HasChange("server_url") {
		input.ServerUrl = aws.String(d.Get("server_url").(string))
	}
	if d.HasChange("state") {
		input.State = awstypes.PluginState(d.Get("state").(string))
	}

	_, err = conn.UpdatePlugin(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating qbusiness plugin: %s", err)
	}

	return append(diags, resourcePluginRead(ctx, d, meta)...)
}

func resourcePluginDelete(ctx context.Context, d *sdkschema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).QBusinessClient(ctx)

	application_id, plugin_id, err := parsePluginID(d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "parsing qbusiness plugin ID (%s): %s", d.Id(), err)
	}

	_, err = conn.DeletePlugin(ctx, &qbusiness.DeletePluginInput{
		ApplicationId: aws.String(application_id),
		PluginId:      aws.String(plugin_id),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting qbusiness plugin (%s): %s", d.Id(), err)
	}

	return diags
}

func parsePluginID(id string) (string, string, error) {
	parts := strings.Split(id, "/")

	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid plugin ID: %s", id)
	}

	return parts[0], parts[1], nil
}

const (
	pluginResourceIDPartCount = 2
)

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

func flattenBasicAuthConfiguration(basicAuthConfiguration *awstypes.BasicAuthConfiguration) []interface{} {
	if basicAuthConfiguration == nil {
		return []interface{}{}
	}
	return []interface{}{
		map[string]interface{}{
			"role_arn":   aws.ToString(basicAuthConfiguration.RoleArn),
			"secret_arn": aws.ToString(basicAuthConfiguration.SecretArn),
		},
	}
}

func flattenOAuth2ClientCredentialConfiguration(oauth2ClientCredentialConfiguration *awstypes.OAuth2ClientCredentialConfiguration) []interface{} {
	if oauth2ClientCredentialConfiguration == nil {
		return []interface{}{}
	}
	return []interface{}{
		map[string]interface{}{
			"role_arn":   aws.ToString(oauth2ClientCredentialConfiguration.RoleArn),
			"secret_arn": aws.ToString(oauth2ClientCredentialConfiguration.SecretArn),
		},
	}
}

func expandBasicAuthConfiguration(basicAuthConfiguration []interface{}) *awstypes.PluginAuthConfigurationMemberBasicAuthConfiguration {
	if len(basicAuthConfiguration) == 0 {
		return nil
	}

	basicAuth := basicAuthConfiguration[0].(map[string]interface{})

	return &awstypes.PluginAuthConfigurationMemberBasicAuthConfiguration{
		Value: awstypes.BasicAuthConfiguration{
			RoleArn:   aws.String(basicAuth["role_arn"].(string)),
			SecretArn: aws.String(basicAuth["secret_arn"].(string)),
		},
	}
}

func expandOAuth2ClientCredentialConfiguration(oauth2ClientCredentialConfiguration []interface{}) *awstypes.PluginAuthConfigurationMemberOAuth2ClientCredentialConfiguration {
	if len(oauth2ClientCredentialConfiguration) == 0 {
		return nil
	}

	oAuth2 := oauth2ClientCredentialConfiguration[0].(map[string]interface{})

	return &awstypes.PluginAuthConfigurationMemberOAuth2ClientCredentialConfiguration{
		Value: awstypes.OAuth2ClientCredentialConfiguration{
			RoleArn:   aws.String(oAuth2["role_arn"].(string)),
			SecretArn: aws.String(oAuth2["secret_arn"].(string)),
		},
	}
}
