// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package opsworks

import (
	"context"
	"fmt"
	"log"
	"slices"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/opsworks"
	awstypes "github.com/aws/aws-sdk-go-v2/service/opsworks/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_opsworks_application", name="Application")
func resourceApplication() *schema.Resource {
	return &schema.Resource{
		DeprecationMessage: "This resource is deprecated and will be removed in the next major version of the AWS Provider. Consider the AWS Systems Manager service for managing applications.",

		CreateWithoutTimeout: resourceApplicationCreate,
		ReadWithoutTimeout:   resourceApplicationRead,
		UpdateWithoutTimeout: resourceApplicationUpdate,
		DeleteWithoutTimeout: resourceApplicationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
			},
			"short_name": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
				ForceNew: true,
			},
			names.AttrType: {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: enum.Validate[awstypes.AppType](),
			},
			"stack_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			// TODO: the following 4 vals are really part of the Attributes array. We should validate that only ones relevant to the chosen type are set, perhaps. (what is the default type? how do they map?)
			"document_root": {
				Type:     schema.TypeString,
				Optional: true,
				//Default:  "public",
			},
			"rails_env": {
				Type:     schema.TypeString,
				Optional: true,
				//Default:  "production",
			},
			"auto_bundle_on_deploy": {
				Type:     schema.TypeString,
				Optional: true,
				//Default:  true,
			},
			"aws_flow_ruby_settings": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"app_source": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrType: {
							Type:     schema.TypeString,
							Required: true,
							// Because the SDK only accepts typed arguments from SourceType, we cannot add `other` even though the API will accept it.
							// This service has been deprecated and will only for completeness sake be migrated, will remove the validation as to not cause a client validation exception but rather let AWS API handle it.
							//ValidateFunc: validation.StringInSlice(append(opsworks.SourceType_Values(), "other"), false),
						},

						names.AttrURL: {
							Type:     schema.TypeString,
							Optional: true,
						},

						names.AttrUsername: {
							Type:     schema.TypeString,
							Optional: true,
						},

						names.AttrPassword: {
							Type:      schema.TypeString,
							Optional:  true,
							Sensitive: true,
						},

						"revision": {
							Type:     schema.TypeString,
							Optional: true,
						},

						"ssh_key": {
							Type:      schema.TypeString,
							Optional:  true,
							Sensitive: true,
						},
					},
				},
			},
			"data_source_type": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.StringInSlice([]string{
					"AutoSelectOpsworksMysqlInstance",
					"OpsworksMysqlInstance",
					"RdsDbInstance",
					"None",
				}, false),
			},
			"data_source_database_name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"data_source_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: verify.ValidARN,
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Optional: true,
			},
			"domains": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrEnvironment: {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrKey: {
							Type:     schema.TypeString,
							Required: true,
						},
						names.AttrValue: {
							Type:     schema.TypeString,
							Required: true,
						},
						"secure": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
						},
					},
				},
			},
			"enable_ssl": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"ssl_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				//Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrCertificate: {
							Type:     schema.TypeString,
							Required: true,
							StateFunc: func(v any) string {
								switch v := v.(type) {
								case string:
									return strings.TrimSpace(v)
								default:
									return ""
								}
							},
						},
						names.AttrPrivateKey: {
							Type:      schema.TypeString,
							Required:  true,
							Sensitive: true,
							StateFunc: func(v any) string {
								switch v := v.(type) {
								case string:
									return strings.TrimSpace(v)
								default:
									return ""
								}
							},
						},
						"chain": {
							Type:     schema.TypeString,
							Optional: true,
							StateFunc: func(v any) string {
								switch v := v.(type) {
								case string:
									return strings.TrimSpace(v)
								default:
									return ""
								}
							},
						},
					},
				},
			},
		},
	}
}

func resourceApplicationValidate(d *schema.ResourceData) error {
	appSourceCount := d.Get("app_source.#").(int)
	if appSourceCount > 1 {
		return fmt.Errorf("Only one app_source is permitted.")
	}

	sslCount := d.Get("ssl_configuration.#").(int)
	if sslCount > 1 {
		return fmt.Errorf("Only one ssl_configuration is permitted.")
	}

	attrType := awstypes.AppType(d.Get(names.AttrType).(string))
	if attrType == awstypes.AppTypeNodejs || attrType == awstypes.AppTypeJava {
		// allowed attributes: none
		if d.Get("document_root").(string) != "" || d.Get("rails_env").(string) != "" || d.Get("auto_bundle_on_deploy").(string) != "" || d.Get("aws_flow_ruby_settings").(string) != "" {
			return fmt.Errorf("No additional attributes are allowed for app type '%s'.", attrType)
		}
	} else if attrType == awstypes.AppTypeRails {
		// allowed attributes: document_root, rails_env, auto_bundle_on_deploy
		if d.Get("aws_flow_ruby_settings").(string) != "" {
			return fmt.Errorf("Only 'document_root, rails_env, auto_bundle_on_deploy' are allowed for app type '%s'.", awstypes.AppTypeRails)
		}
		// rails_env is required
		if _, ok := d.GetOk("rails_env"); !ok {
			return fmt.Errorf("Set rails_env must be set if type is set to rails.")
		}
	} else if attrType == awstypes.AppTypePhp || attrType == awstypes.AppTypeStatic || attrType == awstypes.AppTypeOther {
		log.Printf("[DEBUG] the app type is : %s", attrType)
		log.Printf("[DEBUG] the attributes are: document_root '%s', rails_env '%s', auto_bundle_on_deploy '%s', aws_flow_ruby_settings '%s'", d.Get("document_root").(string), d.Get("rails_env").(string), d.Get("auto_bundle_on_deploy").(string), d.Get("aws_flow_ruby_settings").(string))
		// allowed attributes: document_root
		if d.Get("rails_env").(string) != "" || d.Get("auto_bundle_on_deploy").(string) != "" || d.Get("aws_flow_ruby_settings").(string) != "" {
			return fmt.Errorf("Only 'document_root' is allowed for app type '%s'.", attrType)
		}
	} else if attrType == awstypes.AppTypeAwsFlowRuby {
		// allowed attributes: aws_flow_ruby_settings
		if d.Get("document_root").(string) != "" || d.Get("rails_env").(string) != "" || d.Get("auto_bundle_on_deploy").(string) != "" {
			return fmt.Errorf("Only 'aws_flow_ruby_settings' is allowed for app type '%s'.", attrType)
		}
	}

	return nil
}

func resourceApplicationRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OpsWorksClient(ctx)

	log.Printf("[DEBUG] Reading OpsWorks Application: %s", d.Id())

	output, err := findAppByID(ctx, conn, d.Id())

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		log.Printf("[DEBUG] OpsWorks Application (%s) not found", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading OpsWorks Application (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrName, output.Name)
	d.Set("stack_id", output.StackId)
	d.Set(names.AttrType, output.Type)
	d.Set(names.AttrDescription, output.Description)
	d.Set("domains", output.Domains)
	d.Set("enable_ssl", output.EnableSsl)
	err = resourceSetApplicationSSL(d, output.SslConfiguration)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading OpsWorks Application (%s): setting ssl_configuration: %s", d.Id(), err)
	}
	err = resourceSetApplicationSource(d, output.AppSource)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading OpsWorks Application (%s): setting app_source: %s", d.Id(), err)
	}
	resourceSetApplicationsDataSource(d, output.DataSources)
	resourceSetApplicationEnvironmentVariable(d, output.Environment)
	resourceSetApplicationAttributes(d, output.Attributes)

	return diags
}

func resourceApplicationCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OpsWorksClient(ctx)

	err := resourceApplicationValidate(d)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating OpsWorks Application: %s", err)
	}

	req := &opsworks.CreateAppInput{
		Name:             aws.String(d.Get(names.AttrName).(string)),
		Shortname:        aws.String(d.Get("short_name").(string)),
		StackId:          aws.String(d.Get("stack_id").(string)),
		Type:             awstypes.AppType(d.Get(names.AttrType).(string)),
		Description:      aws.String(d.Get(names.AttrDescription).(string)),
		Domains:          flex.ExpandStringValueList(d.Get("domains").([]any)),
		EnableSsl:        aws.Bool(d.Get("enable_ssl").(bool)),
		SslConfiguration: resourceApplicationSSL(d),
		AppSource:        resourceApplicationSource(d),
		DataSources:      resourceApplicationsDataSource(d),
		Environment:      resourceApplicationEnvironmentVariable(d),
		Attributes:       resourceApplicationAttributes(d),
	}

	resp, err := conn.CreateApp(ctx, req)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating OpsWorks Application: %s", err)
	}

	d.SetId(aws.ToString(resp.AppId))

	return append(diags, resourceApplicationRead(ctx, d, meta)...)
}

func resourceApplicationUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OpsWorksClient(ctx)

	err := resourceApplicationValidate(d)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating OpsWorks Application (%s): %s", d.Id(), err)
	}

	req := &opsworks.UpdateAppInput{
		AppId:            aws.String(d.Id()),
		Name:             aws.String(d.Get(names.AttrName).(string)),
		Type:             awstypes.AppType(d.Get(names.AttrType).(string)),
		Description:      aws.String(d.Get(names.AttrDescription).(string)),
		Domains:          flex.ExpandStringValueList(d.Get("domains").([]any)),
		EnableSsl:        aws.Bool(d.Get("enable_ssl").(bool)),
		SslConfiguration: resourceApplicationSSL(d),
		AppSource:        resourceApplicationSource(d),
		DataSources:      resourceApplicationsDataSource(d),
		Environment:      resourceApplicationEnvironmentVariable(d),
		Attributes:       resourceApplicationAttributes(d),
	}

	log.Printf("[DEBUG] Updating OpsWorks Application: %s", d.Id())

	_, err = conn.UpdateApp(ctx, req)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating OpsWorks Application (%s): %s", d.Id(), err)
	}

	return append(diags, resourceApplicationRead(ctx, d, meta)...)
}

func resourceApplicationDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OpsWorksClient(ctx)

	log.Printf("[DEBUG] Deleting OpsWorks Application: %s", d.Id())
	_, err := conn.DeleteApp(ctx, &opsworks.DeleteAppInput{
		AppId: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting OpsWorks Application (%s): %s", d.Id(), err)
	}

	return diags
}

func findAppByID(ctx context.Context, conn *opsworks.Client, id string) (*awstypes.App, error) {
	input := &opsworks.DescribeAppsInput{
		AppIds: []string{id},
	}

	output, err := conn.DescribeApps(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}
	if err != nil {
		return nil, err
	}

	if output == nil || output.Apps == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return tfresource.AssertSingleValueResult(output.Apps)
}

func resourceFindEnvironmentVariable(key string, vs []awstypes.EnvironmentVariable) *awstypes.EnvironmentVariable {
	for _, v := range vs {
		if aws.ToString(v.Key) == key {
			return &v
		}
	}
	return nil
}

func resourceSetApplicationEnvironmentVariable(d *schema.ResourceData, vs []awstypes.EnvironmentVariable) {
	if len(vs) == 0 {
		d.Set(names.AttrEnvironment, nil)
		return
	}

	// sensitive variables are returned obfuscated from the API, this creates a
	// permadiff between the obfuscated API response and the config value. We
	// start with the existing state so it can passthrough when the key is secure
	values := d.Get(names.AttrEnvironment).(*schema.Set).List()

	for i := range values {
		value := values[i].(map[string]any)
		if v := resourceFindEnvironmentVariable(value[names.AttrKey].(string), vs); v != nil {
			if !aws.ToBool(v.Secure) {
				value["secure"] = aws.ToBool(v.Secure)
				value[names.AttrKey] = aws.ToString(v.Key)
				value[names.AttrValue] = aws.ToString(v.Value)
				values[i] = value
			}
		} else {
			// delete if not found in API response
			values = slices.Delete(values, i, i+1)
		}
	}

	d.Set(names.AttrEnvironment, values)
}

func resourceApplicationEnvironmentVariable(d *schema.ResourceData) []awstypes.EnvironmentVariable {
	environmentVariables := d.Get(names.AttrEnvironment).(*schema.Set).List()
	result := make([]awstypes.EnvironmentVariable, len(environmentVariables))

	for i := range environmentVariables {
		env := environmentVariables[i].(map[string]any)

		result[i] = awstypes.EnvironmentVariable{
			Key:    aws.String(env[names.AttrKey].(string)),
			Value:  aws.String(env[names.AttrValue].(string)),
			Secure: aws.Bool(env["secure"].(bool)),
		}
	}
	return result
}

func resourceApplicationSource(d *schema.ResourceData) *awstypes.Source {
	count := d.Get("app_source.#").(int)
	if count == 0 {
		return nil
	}

	return &awstypes.Source{
		Type:     awstypes.SourceType(d.Get("app_source.0.type").(string)),
		Url:      aws.String(d.Get("app_source.0.url").(string)),
		Username: aws.String(d.Get("app_source.0.username").(string)),
		Password: aws.String(d.Get("app_source.0.password").(string)),
		Revision: aws.String(d.Get("app_source.0.revision").(string)),
		SshKey:   aws.String(d.Get("app_source.0.ssh_key").(string)),
	}
}

func resourceSetApplicationSource(d *schema.ResourceData, v *awstypes.Source) error {
	nv := make([]any, 0, 1)
	if v != nil {
		m := make(map[string]any)

		m[names.AttrType] = v.Type

		if v.Url != nil {
			m[names.AttrURL] = aws.ToString(v.Url)
		}
		if v.Username != nil {
			m[names.AttrUsername] = aws.ToString(v.Username)
		}
		if v.Revision != nil {
			m["revision"] = aws.ToString(v.Revision)
		}

		// v.Password and v.SshKey will, on read, contain the placeholder string
		// "*****FILTERED*****", so we ignore it on read and let persist
		// the value already in the state.
		m[names.AttrPassword] = d.Get("app_source.0.password").(string)
		m["ssh_key"] = d.Get("app_source.0.ssh_key").(string)

		nv = append(nv, m)
	}

	return d.Set("app_source", nv)
}

func resourceApplicationsDataSource(d *schema.ResourceData) []awstypes.DataSource {
	arn := d.Get("data_source_arn").(string)
	databaseName := d.Get("data_source_database_name").(string)
	databaseType := d.Get("data_source_type").(string)

	result := make([]awstypes.DataSource, 1)

	if len(arn) > 0 || len(databaseName) > 0 || len(databaseType) > 0 {
		result[0] = awstypes.DataSource{
			Arn:          aws.String(arn),
			DatabaseName: aws.String(databaseName),
			Type:         aws.String(databaseType),
		}
	}
	return result
}

func resourceSetApplicationsDataSource(d *schema.ResourceData, v []awstypes.DataSource) {
	d.Set("data_source_arn", nil)
	d.Set("data_source_database_name", nil)
	d.Set("data_source_type", nil)

	if len(v) == 0 {
		return
	}

	d.Set("data_source_arn", v[0].Arn)
	d.Set("data_source_database_name", v[0].DatabaseName)
	d.Set("data_source_type", v[0].Type)
}

func resourceApplicationSSL(d *schema.ResourceData) *awstypes.SslConfiguration {
	count := d.Get("ssl_configuration.#").(int)
	if count == 0 {
		return nil
	}

	return &awstypes.SslConfiguration{
		PrivateKey:  aws.String(d.Get("ssl_configuration.0.private_key").(string)),
		Certificate: aws.String(d.Get("ssl_configuration.0.certificate").(string)),
		Chain:       aws.String(d.Get("ssl_configuration.0.chain").(string)),
	}
}

func resourceSetApplicationSSL(d *schema.ResourceData, v *awstypes.SslConfiguration) error {
	nv := make([]any, 0, 1)
	set := false
	if v != nil {
		m := make(map[string]any)
		if v.PrivateKey != nil {
			m[names.AttrPrivateKey] = aws.ToString(v.PrivateKey)
			set = true
		}
		if v.Certificate != nil {
			m[names.AttrCertificate] = aws.ToString(v.Certificate)
			set = true
		}
		if v.Chain != nil {
			m["chain"] = aws.ToString(v.Chain)
			set = true
		}
		if set {
			nv = append(nv, m)
		}
	}

	return d.Set("ssl_configuration", nv)
}

func resourceApplicationAttributes(d *schema.ResourceData) map[string]string {
	attributes := make(map[string]string)

	if val := d.Get("document_root").(string); len(val) > 0 {
		attributes[string(awstypes.AppAttributesKeysDocumentRoot)] = val
	}
	if val := d.Get("aws_flow_ruby_settings").(string); len(val) > 0 {
		attributes[string(awstypes.AppAttributesKeysAwsFlowRubySettings)] = val
	}
	if val := d.Get("rails_env").(string); len(val) > 0 {
		attributes[string(awstypes.AppAttributesKeysRailsEnv)] = val
	}
	if val := d.Get("auto_bundle_on_deploy").(string); len(val) > 0 {
		if val == "1" {
			val = "true"
		} else if val == "0" {
			val = "false"
		}
		attributes[string(awstypes.AppAttributesKeysAutoBundleOnDeploy)] = val
	}

	return attributes
}

func resourceSetApplicationAttributes(d *schema.ResourceData, v map[string]string) {
	d.Set("document_root", nil)
	d.Set("rails_env", nil)
	d.Set("aws_flow_ruby_settings", nil)
	d.Set("auto_bundle_on_deploy", nil)

	attrType := d.Get(names.AttrType)
	if attrType == awstypes.AppTypeNodejs || attrType == awstypes.AppTypeJava {
		return
	} else if attrType == awstypes.AppTypeRails {
		if val, ok := v[string(awstypes.AppAttributesKeysDocumentRoot)]; ok {
			d.Set("document_root", val)
		}
		if val, ok := v[string(awstypes.AppAttributesKeysRailsEnv)]; ok {
			d.Set("rails_env", val)
		}
		if val, ok := v[string(awstypes.AppAttributesKeysAutoBundleOnDeploy)]; ok {
			d.Set("auto_bundle_on_deploy", val)
		}
		return
	} else if attrType == awstypes.AppTypePhp || attrType == awstypes.AppTypeStatic || attrType == awstypes.AppTypeOther {
		if val, ok := v[string(awstypes.AppAttributesKeysDocumentRoot)]; ok {
			d.Set("document_root", val)
		}
		return
	} else if attrType == awstypes.AppTypeAwsFlowRuby {
		if val, ok := v[string(awstypes.AppAttributesKeysAwsFlowRubySettings)]; ok {
			d.Set("aws_flow_ruby_settings", val)
		}
		return
	}
}
