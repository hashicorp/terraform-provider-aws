// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package opsworks

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/opsworks"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_opsworks_application")
func ResourceApplication() *schema.Resource {
	return &schema.Resource{

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
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(opsworks.AppType_Values(), false),
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
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(append(opsworks.SourceType_Values(), "other"), false),
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
							StateFunc: func(v interface{}) string {
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
							StateFunc: func(v interface{}) string {
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
							StateFunc: func(v interface{}) string {
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

	if d.Get(names.AttrType) == opsworks.AppTypeNodejs || d.Get(names.AttrType) == opsworks.AppTypeJava {
		// allowed attributes: none
		if d.Get("document_root").(string) != "" || d.Get("rails_env").(string) != "" || d.Get("auto_bundle_on_deploy").(string) != "" || d.Get("aws_flow_ruby_settings").(string) != "" {
			return fmt.Errorf("No additional attributes are allowed for app type '%s'.", d.Get(names.AttrType).(string))
		}
	} else if d.Get(names.AttrType) == opsworks.AppTypeRails {
		// allowed attributes: document_root, rails_env, auto_bundle_on_deploy
		if d.Get("aws_flow_ruby_settings").(string) != "" {
			return fmt.Errorf("Only 'document_root, rails_env, auto_bundle_on_deploy' are allowed for app type '%s'.", opsworks.AppTypeRails)
		}
		// rails_env is required
		if _, ok := d.GetOk("rails_env"); !ok {
			return fmt.Errorf("Set rails_env must be set if type is set to rails.")
		}
	} else if d.Get(names.AttrType) == opsworks.AppTypePhp || d.Get(names.AttrType) == opsworks.AppTypeStatic || d.Get(names.AttrType) == opsworks.AppTypeOther {
		log.Printf("[DEBUG] the app type is : %s", d.Get(names.AttrType).(string))
		log.Printf("[DEBUG] the attributes are: document_root '%s', rails_env '%s', auto_bundle_on_deploy '%s', aws_flow_ruby_settings '%s'", d.Get("document_root").(string), d.Get("rails_env").(string), d.Get("auto_bundle_on_deploy").(string), d.Get("aws_flow_ruby_settings").(string))
		// allowed attributes: document_root
		if d.Get("rails_env").(string) != "" || d.Get("auto_bundle_on_deploy").(string) != "" || d.Get("aws_flow_ruby_settings").(string) != "" {
			return fmt.Errorf("Only 'document_root' is allowed for app type '%s'.", d.Get(names.AttrType).(string))
		}
	} else if d.Get(names.AttrType) == opsworks.AppTypeAwsFlowRuby {
		// allowed attributes: aws_flow_ruby_settings
		if d.Get("document_root").(string) != "" || d.Get("rails_env").(string) != "" || d.Get("auto_bundle_on_deploy").(string) != "" {
			return fmt.Errorf("Only 'aws_flow_ruby_settings' is allowed for app type '%s'.", d.Get(names.AttrType).(string))
		}
	}

	return nil
}

func resourceApplicationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OpsWorksConn(ctx)

	req := &opsworks.DescribeAppsInput{
		AppIds: []*string{
			aws.String(d.Id()),
		},
	}

	log.Printf("[DEBUG] Reading OpsWorks Application: %s", d.Id())

	resp, err := conn.DescribeAppsWithContext(ctx, req)

	if tfawserr.ErrCodeEquals(err, opsworks.ErrCodeResourceNotFoundException) {
		log.Printf("[DEBUG] OpsWorks Application (%s) not found", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading OpsWorks Application (%s): %s", d.Id(), err)
	}

	app := resp.Apps[0]
	log.Printf("[DEBUG] Opsworks Application: %#v", app)

	d.Set(names.AttrName, app.Name)
	d.Set("stack_id", app.StackId)
	d.Set(names.AttrType, app.Type)
	d.Set(names.AttrDescription, app.Description)
	d.Set("domains", flex.FlattenStringList(app.Domains))
	d.Set("enable_ssl", app.EnableSsl)
	err = resourceSetApplicationSSL(d, app.SslConfiguration)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading OpsWorks Application (%s): setting ssl_configuration: %s", d.Id(), err)
	}
	err = resourceSetApplicationSource(d, app.AppSource)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading OpsWorks Application (%s): setting app_source: %s", d.Id(), err)
	}
	resourceSetApplicationsDataSource(d, app.DataSources)
	resourceSetApplicationEnvironmentVariable(d, app.Environment)
	resourceSetApplicationAttributes(d, app.Attributes)

	return diags
}

func resourceApplicationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OpsWorksConn(ctx)

	err := resourceApplicationValidate(d)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating OpsWorks Application: %s", err)
	}

	req := &opsworks.CreateAppInput{
		Name:             aws.String(d.Get(names.AttrName).(string)),
		Shortname:        aws.String(d.Get("short_name").(string)),
		StackId:          aws.String(d.Get("stack_id").(string)),
		Type:             aws.String(d.Get(names.AttrType).(string)),
		Description:      aws.String(d.Get(names.AttrDescription).(string)),
		Domains:          flex.ExpandStringList(d.Get("domains").([]interface{})),
		EnableSsl:        aws.Bool(d.Get("enable_ssl").(bool)),
		SslConfiguration: resourceApplicationSSL(d),
		AppSource:        resourceApplicationSource(d),
		DataSources:      resourceApplicationsDataSource(d),
		Environment:      resourceApplicationEnvironmentVariable(d),
		Attributes:       resourceApplicationAttributes(d),
	}

	resp, err := conn.CreateAppWithContext(ctx, req)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating OpsWorks Application: %s", err)
	}

	d.SetId(aws.StringValue(resp.AppId))

	return append(diags, resourceApplicationRead(ctx, d, meta)...)
}

func resourceApplicationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OpsWorksConn(ctx)

	err := resourceApplicationValidate(d)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating OpsWorks Application (%s): %s", d.Id(), err)
	}

	req := &opsworks.UpdateAppInput{
		AppId:            aws.String(d.Id()),
		Name:             aws.String(d.Get(names.AttrName).(string)),
		Type:             aws.String(d.Get(names.AttrType).(string)),
		Description:      aws.String(d.Get(names.AttrDescription).(string)),
		Domains:          flex.ExpandStringList(d.Get("domains").([]interface{})),
		EnableSsl:        aws.Bool(d.Get("enable_ssl").(bool)),
		SslConfiguration: resourceApplicationSSL(d),
		AppSource:        resourceApplicationSource(d),
		DataSources:      resourceApplicationsDataSource(d),
		Environment:      resourceApplicationEnvironmentVariable(d),
		Attributes:       resourceApplicationAttributes(d),
	}

	log.Printf("[DEBUG] Updating OpsWorks Application: %s", d.Id())

	_, err = conn.UpdateAppWithContext(ctx, req)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating OpsWorks Application (%s): %s", d.Id(), err)
	}

	return append(diags, resourceApplicationRead(ctx, d, meta)...)
}

func resourceApplicationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OpsWorksConn(ctx)

	log.Printf("[DEBUG] Deleting OpsWorks Application: %s", d.Id())
	_, err := conn.DeleteAppWithContext(ctx, &opsworks.DeleteAppInput{
		AppId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, opsworks.ErrCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting OpsWorks Application (%s): %s", d.Id(), err)
	}

	return diags
}

func resourceFindEnvironmentVariable(key string, vs []*opsworks.EnvironmentVariable) *opsworks.EnvironmentVariable {
	for _, v := range vs {
		if aws.StringValue(v.Key) == key {
			return v
		}
	}
	return nil
}

func resourceSetApplicationEnvironmentVariable(d *schema.ResourceData, vs []*opsworks.EnvironmentVariable) {
	if len(vs) == 0 {
		d.Set(names.AttrEnvironment, nil)
		return
	}

	// sensitive variables are returned obfuscated from the API, this creates a
	// permadiff between the obfuscated API response and the config value. We
	// start with the existing state so it can passthrough when the key is secure
	values := d.Get(names.AttrEnvironment).(*schema.Set).List()

	for i := 0; i < len(values); i++ {
		value := values[i].(map[string]interface{})
		if v := resourceFindEnvironmentVariable(value[names.AttrKey].(string), vs); v != nil {
			if !aws.BoolValue(v.Secure) {
				value["secure"] = aws.BoolValue(v.Secure)
				value[names.AttrKey] = aws.StringValue(v.Key)
				value[names.AttrValue] = aws.StringValue(v.Value)
				values[i] = value
			}
		} else {
			// delete if not found in API response
			values = append(values[:i], values[i+1:]...)
		}
	}

	d.Set(names.AttrEnvironment, values)
}

func resourceApplicationEnvironmentVariable(d *schema.ResourceData) []*opsworks.EnvironmentVariable {
	environmentVariables := d.Get(names.AttrEnvironment).(*schema.Set).List()
	result := make([]*opsworks.EnvironmentVariable, len(environmentVariables))

	for i := 0; i < len(environmentVariables); i++ {
		env := environmentVariables[i].(map[string]interface{})

		result[i] = &opsworks.EnvironmentVariable{
			Key:    aws.String(env[names.AttrKey].(string)),
			Value:  aws.String(env[names.AttrValue].(string)),
			Secure: aws.Bool(env["secure"].(bool)),
		}
	}
	return result
}

func resourceApplicationSource(d *schema.ResourceData) *opsworks.Source {
	count := d.Get("app_source.#").(int)
	if count == 0 {
		return nil
	}

	return &opsworks.Source{
		Type:     aws.String(d.Get("app_source.0.type").(string)),
		Url:      aws.String(d.Get("app_source.0.url").(string)),
		Username: aws.String(d.Get("app_source.0.username").(string)),
		Password: aws.String(d.Get("app_source.0.password").(string)),
		Revision: aws.String(d.Get("app_source.0.revision").(string)),
		SshKey:   aws.String(d.Get("app_source.0.ssh_key").(string)),
	}
}

func resourceSetApplicationSource(d *schema.ResourceData, v *opsworks.Source) error {
	nv := make([]interface{}, 0, 1)
	if v != nil {
		m := make(map[string]interface{})
		if v.Type != nil {
			m[names.AttrType] = aws.StringValue(v.Type)
		}
		if v.Url != nil {
			m[names.AttrURL] = aws.StringValue(v.Url)
		}
		if v.Username != nil {
			m[names.AttrUsername] = aws.StringValue(v.Username)
		}
		if v.Revision != nil {
			m["revision"] = aws.StringValue(v.Revision)
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

func resourceApplicationsDataSource(d *schema.ResourceData) []*opsworks.DataSource {
	arn := d.Get("data_source_arn").(string)
	databaseName := d.Get("data_source_database_name").(string)
	databaseType := d.Get("data_source_type").(string)

	result := make([]*opsworks.DataSource, 1)

	if len(arn) > 0 || len(databaseName) > 0 || len(databaseType) > 0 {
		result[0] = &opsworks.DataSource{
			Arn:          aws.String(arn),
			DatabaseName: aws.String(databaseName),
			Type:         aws.String(databaseType),
		}
	}
	return result
}

func resourceSetApplicationsDataSource(d *schema.ResourceData, v []*opsworks.DataSource) {
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

func resourceApplicationSSL(d *schema.ResourceData) *opsworks.SslConfiguration {
	count := d.Get("ssl_configuration.#").(int)
	if count == 0 {
		return nil
	}

	return &opsworks.SslConfiguration{
		PrivateKey:  aws.String(d.Get("ssl_configuration.0.private_key").(string)),
		Certificate: aws.String(d.Get("ssl_configuration.0.certificate").(string)),
		Chain:       aws.String(d.Get("ssl_configuration.0.chain").(string)),
	}
}

func resourceSetApplicationSSL(d *schema.ResourceData, v *opsworks.SslConfiguration) error {
	nv := make([]interface{}, 0, 1)
	set := false
	if v != nil {
		m := make(map[string]interface{})
		if v.PrivateKey != nil {
			m[names.AttrPrivateKey] = aws.StringValue(v.PrivateKey)
			set = true
		}
		if v.Certificate != nil {
			m[names.AttrCertificate] = aws.StringValue(v.Certificate)
			set = true
		}
		if v.Chain != nil {
			m["chain"] = aws.StringValue(v.Chain)
			set = true
		}
		if set {
			nv = append(nv, m)
		}
	}

	return d.Set("ssl_configuration", nv)
}

func resourceApplicationAttributes(d *schema.ResourceData) map[string]*string {
	attributes := make(map[string]*string)

	if val := d.Get("document_root").(string); len(val) > 0 {
		attributes[opsworks.AppAttributesKeysDocumentRoot] = aws.String(val)
	}
	if val := d.Get("aws_flow_ruby_settings").(string); len(val) > 0 {
		attributes[opsworks.AppAttributesKeysAwsFlowRubySettings] = aws.String(val)
	}
	if val := d.Get("rails_env").(string); len(val) > 0 {
		attributes[opsworks.AppAttributesKeysRailsEnv] = aws.String(val)
	}
	if val := d.Get("auto_bundle_on_deploy").(string); len(val) > 0 {
		if val == "1" {
			val = "true"
		} else if val == "0" {
			val = "false"
		}
		attributes[opsworks.AppAttributesKeysAutoBundleOnDeploy] = aws.String(val)
	}

	return attributes
}

func resourceSetApplicationAttributes(d *schema.ResourceData, v map[string]*string) {
	d.Set("document_root", nil)
	d.Set("rails_env", nil)
	d.Set("aws_flow_ruby_settings", nil)
	d.Set("auto_bundle_on_deploy", nil)

	if d.Get(names.AttrType) == opsworks.AppTypeNodejs || d.Get(names.AttrType) == opsworks.AppTypeJava {
		return
	} else if d.Get(names.AttrType) == opsworks.AppTypeRails {
		if val, ok := v[opsworks.AppAttributesKeysDocumentRoot]; ok {
			d.Set("document_root", val)
		}
		if val, ok := v[opsworks.AppAttributesKeysRailsEnv]; ok {
			d.Set("rails_env", val)
		}
		if val, ok := v[opsworks.AppAttributesKeysAutoBundleOnDeploy]; ok {
			d.Set("auto_bundle_on_deploy", val)
		}
		return
	} else if d.Get(names.AttrType) == opsworks.AppTypePhp || d.Get(names.AttrType) == opsworks.AppTypeStatic || d.Get(names.AttrType) == opsworks.AppTypeOther {
		if val, ok := v[opsworks.AppAttributesKeysDocumentRoot]; ok {
			d.Set("document_root", val)
		}
		return
	} else if d.Get(names.AttrType) == opsworks.AppTypeAwsFlowRuby {
		if val, ok := v[opsworks.AppAttributesKeysAwsFlowRubySettings]; ok {
			d.Set("aws_flow_ruby_settings", val)
		}
		return
	}
}
