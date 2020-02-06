package aws

import (
	"encoding/base64"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/amplify"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsAmplifyApp() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsAmplifyAppCreate,
		Read:   resourceAwsAmplifyAppRead,
		Update: resourceAwsAmplifyAppUpdate,
		Delete: resourceAwsAmplifyAppDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"auto_branch_creation_config": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"auto_branch_creation_patterns": {
							Type:     schema.TypeList,
							Optional: true,
							MinItems: 1,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"basic_auth_config": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"enable_basic_auth": {
										Type:     schema.TypeBool,
										Optional: true,
									},
									"password": {
										Type:         schema.TypeString,
										Optional:     true,
										Sensitive:    true,
										ValidateFunc: validation.StringLenBetween(1, 255),
									},
									"username": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringLenBetween(1, 255),
									},
								},
							},
						},
						"build_spec": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"enable_auto_branch_creation": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"enable_auto_build": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"enable_pull_request_preview": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"environment_variables": {
							Type:     schema.TypeMap,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"framework": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"pull_request_environment_name": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"stage": {
							Type:     schema.TypeString,
							Optional: true,
							DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
								// stage is "NONE" by default
								if old == "NONE" && new == "" {
									return true
								}
								return false
							},
							ValidateFunc: validation.StringInSlice([]string{
								amplify.StageProduction,
								amplify.StageBeta,
								amplify.StageDevelopment,
								amplify.StageExperimental,
								amplify.StagePullRequest,
							}, false),
						},
					},
				},
			},
			"basic_auth_config": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enable_basic_auth": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"password": {
							Type:         schema.TypeString,
							Optional:     true,
							Sensitive:    true,
							ValidateFunc: validation.StringLenBetween(1, 255),
						},
						"username": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(1, 255),
						},
					},
				},
			},
			"build_spec": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"custom_rules": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"condition": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"source": {
							Type:     schema.TypeString,
							Required: true,
						},
						"status": {
							Type:     schema.TypeString,
							Optional: true,
							ValidateFunc: validation.StringInSlice([]string{
								"200",
								"301",
								"302",
								"404",
							}, false),
						},
						"target": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"default_domain": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"enable_branch_auto_build": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"environment_variables": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"iam_service_role_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateArn,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(3, 1024),
					validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9_-]+$`), "should only contains letters, numbers, _ and -"),
				),
			},
			"platform": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  amplify.PlatformWeb,
				ValidateFunc: validation.StringInSlice([]string{
					amplify.PlatformWeb,
				}, false),
			},
			"repository": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"oauth_token": {
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
			},
			"access_token": {
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
			},
			"tags": tagsSchema(),
		},
	}
}

func resourceAwsAmplifyAppCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).amplifyconn
	log.Print("[DEBUG] Creating Amplify App")

	params := &amplify.CreateAppInput{
		Name: aws.String(d.Get("name").(string)),
	}

	if v, ok := d.GetOk("auto_branch_creation_config"); ok {
		config, patterns, enable := expandAmplifyAutoBranchCreationConfig(v.([]interface{}))
		params.AutoBranchCreationConfig = config
		params.AutoBranchCreationPatterns = patterns
		params.EnableAutoBranchCreation = enable
	}

	if v, ok := d.GetOk("basic_auth_config"); ok {
		enable, credentials := expandAmplifyBasicAuthConfig(v.([]interface{}))
		params.EnableBasicAuth = enable
		params.BasicAuthCredentials = credentials
	}

	if v, ok := d.GetOk("build_spec"); ok {
		params.BuildSpec = aws.String(v.(string))
	}

	if v, ok := d.GetOk("custom_rules"); ok {
		params.CustomRules = expandAmplifyCustomRules(v.([]interface{}))
	}

	if v, ok := d.GetOk("description"); ok {
		params.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("enable_branch_auto_build"); ok {
		params.EnableBranchAutoBuild = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("environment_variables"); ok {
		params.EnvironmentVariables = stringMapToPointers(v.(map[string]interface{}))
	}

	if v, ok := d.GetOk("iam_service_role_arn"); ok {
		params.IamServiceRoleArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("platform"); ok {
		params.Platform = aws.String(v.(string))
	}

	if v, ok := d.GetOk("repository"); ok {
		params.Repository = aws.String(v.(string))
	}

	if v, ok := d.GetOk("access_token"); ok {
		params.AccessToken = aws.String(v.(string))
	}

	if v, ok := d.GetOk("oauth_token"); ok {
		params.OauthToken = aws.String(v.(string))
	}

	if v := d.Get("tags").(map[string]interface{}); len(v) > 0 {
		params.Tags = keyvaluetags.New(v).IgnoreAws().AmplifyTags()
	}

	resp, err := conn.CreateApp(params)
	if err != nil {
		return fmt.Errorf("Error creating Amplify App: %s", err)
	}

	d.SetId(*resp.App.AppId)

	return resourceAwsAmplifyAppRead(d, meta)
}

func resourceAwsAmplifyAppRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).amplifyconn
	log.Printf("[DEBUG] Reading Amplify App: %s", d.Id())

	resp, err := conn.GetApp(&amplify.GetAppInput{
		AppId: aws.String(d.Id()),
	})
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok && awsErr.Code() == amplify.ErrCodeNotFoundException {
			log.Printf("[WARN] Amplify App (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return err
	}

	d.Set("arn", resp.App.AppArn)
	if err := d.Set("auto_branch_creation_config", flattenAmplifyAutoBranchCreationConfig(resp.App.AutoBranchCreationConfig, resp.App.AutoBranchCreationPatterns, resp.App.EnableAutoBranchCreation)); err != nil {
		return fmt.Errorf("error setting auto_branch_creation_config: %s", err)
	}
	if err := d.Set("basic_auth_config", flattenAmplifyBasicAuthConfig(resp.App.EnableBasicAuth, resp.App.BasicAuthCredentials)); err != nil {
		return fmt.Errorf("error setting basic_auth_config: %s", err)
	}
	d.Set("build_spec", resp.App.BuildSpec)
	if err := d.Set("custom_rules", flattenAmplifyCustomRules(resp.App.CustomRules)); err != nil {
		return fmt.Errorf("error setting custom_rules: %s", err)
	}
	d.Set("default_domain", resp.App.DefaultDomain)
	d.Set("description", resp.App.Description)
	d.Set("enable_branch_auto_build", resp.App.EnableBranchAutoBuild)
	if err := d.Set("environment_variables", aws.StringValueMap(resp.App.EnvironmentVariables)); err != nil {
		return fmt.Errorf("error setting environment_variables: %s", err)
	}
	d.Set("iam_service_role_arn", resp.App.IamServiceRoleArn)
	d.Set("name", resp.App.Name)
	d.Set("platform", resp.App.Platform)
	d.Set("repository", resp.App.Repository)
	if err := d.Set("tags", keyvaluetags.AmplifyKeyValueTags(resp.App.Tags).IgnoreAws().Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	return nil
}

func resourceAwsAmplifyAppUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).amplifyconn
	log.Printf("[DEBUG] Updating Amplify App: %s", d.Id())

	params := &amplify.UpdateAppInput{
		AppId: aws.String(d.Id()),
	}

	if d.HasChange("auto_branch_creation_config") {
		v := d.Get("auto_branch_creation_config")
		config, patterns, enable := expandAmplifyAutoBranchCreationConfig(v.([]interface{}))
		params.AutoBranchCreationConfig = config
		params.AutoBranchCreationPatterns = patterns
		params.EnableAutoBranchCreation = enable
	}

	if d.HasChange("basic_auth_config") {
		enable, credentials := expandAmplifyBasicAuthConfig(d.Get("basic_auth_config").([]interface{}))
		params.EnableBasicAuth = enable
		params.BasicAuthCredentials = credentials
	}

	if d.HasChange("build_spec") {
		params.BuildSpec = aws.String(d.Get("build_spec").(string))
	}

	if d.HasChange("custom_rules") {
		params.CustomRules = expandAmplifyCustomRules(d.Get("custom_rules").([]interface{}))
	}

	if d.HasChange("description") {
		params.Description = aws.String(d.Get("description").(string))
	}

	if d.HasChange("enable_branch_auto_build") {
		params.EnableBranchAutoBuild = aws.Bool(d.Get("enable_branch_auto_build").(bool))
	}

	if d.HasChange("environment_variables") {
		v := d.Get("environment_variables")
		params.EnvironmentVariables = expandAmplifyEnvironmentVariables(v.(map[string]interface{}))
	}

	if d.HasChange("iam_service_role_arn") {
		params.IamServiceRoleArn = aws.String(d.Get("iam_service_role_arn").(string))
	}

	if d.HasChange("name") {
		params.Name = aws.String(d.Get("name").(string))
	}

	if d.HasChange("platform") {
		params.Platform = aws.String(d.Get("platform").(string))
	}

	if d.HasChange("repository") {
		params.Repository = aws.String(d.Get("repository").(string))
	}

	if v, ok := d.GetOk("access_token"); ok {
		params.AccessToken = aws.String(v.(string))
	}

	if v, ok := d.GetOk("oauth_token"); ok {
		params.OauthToken = aws.String(v.(string))
	}

	_, err := conn.UpdateApp(params)
	if err != nil {
		return fmt.Errorf("Error updating Amplify App: %s", err)
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")
		if err := keyvaluetags.AmplifyUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating tags: %s", err)
		}
	}

	return resourceAwsAmplifyAppRead(d, meta)
}

func resourceAwsAmplifyAppDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).amplifyconn
	log.Printf("[DEBUG] Deleting Amplify App: %s", d.Id())

	err := deleteAmplifyApp(conn, d.Id())
	if err != nil {
		return fmt.Errorf("Error deleting Amplify App: %s", err)
	}

	return nil
}

func deleteAmplifyApp(conn *amplify.Amplify, appId string) error {
	params := &amplify.DeleteAppInput{
		AppId: aws.String(appId),
	}

	_, err := conn.DeleteApp(params)
	return err
}

func expandAmplifyEnvironmentVariables(envs map[string]interface{}) map[string]*string {
	if len(envs) == 0 {
		empty := ""
		return map[string]*string{"": &empty}
	} else {
		return stringMapToPointers(envs)
	}
}

func expandAmplifyAutoBranchCreationConfig(v []interface{}) (*amplify.AutoBranchCreationConfig, []*string, *bool) {
	config := &amplify.AutoBranchCreationConfig{}
	patterns := make([]*string, 0)
	enable := aws.Bool(false)

	if len(v) == 0 {
		return config, patterns, enable
	}

	e := v[0].(map[string]interface{})

	if ev, ok := e["auto_branch_creation_patterns"]; ok && len(ev.([]interface{})) > 0 {
		patterns = expandStringList(ev.([]interface{}))
	}

	if ev, ok := e["basic_auth_config"]; ok {
		enable, credentials := expandAmplifyBasicAuthConfig(ev.([]interface{}))
		config.EnableBasicAuth = enable
		config.BasicAuthCredentials = credentials
	}

	if ev, ok := e["build_spec"].(string); ok && ev != "" {
		config.BuildSpec = aws.String(ev)
	}

	if ev, ok := e["enable_auto_branch_creation"].(bool); ok {
		enable = aws.Bool(ev)
	}

	if ev, ok := e["enable_auto_build"].(bool); ok {
		config.EnableAutoBuild = aws.Bool(ev)
	}

	if ev, ok := e["enable_pull_request_preview"].(bool); ok {
		config.EnablePullRequestPreview = aws.Bool(ev)
	}

	if ev, ok := e["environment_variables"].(map[string]interface{}); ok {
		config.EnvironmentVariables = expandAmplifyEnvironmentVariables(ev)
	}

	if ev, ok := e["framework"].(string); ok {
		config.Framework = aws.String(ev)
	}

	if ev, ok := e["pull_request_environment_name"].(string); ok {
		config.PullRequestEnvironmentName = aws.String(ev)
	}

	if ev, ok := e["stage"].(string); ok {
		config.Stage = aws.String(ev)
	}

	return config, patterns, enable
}

func flattenAmplifyAutoBranchCreationConfig(config *amplify.AutoBranchCreationConfig, patterns []*string, enable *bool) []map[string]interface{} {
	value := make(map[string]interface{})

	if !aws.BoolValue(enable) {
		return nil
	}

	value["enable_auto_branch_creation"] = aws.BoolValue(enable)
	value["auto_branch_creation_patterns"] = patterns

	if config != nil {
		value["basic_auth_config"] = flattenAmplifyBasicAuthConfig(config.EnableBasicAuth, config.BasicAuthCredentials)
		value["build_spec"] = aws.StringValue(config.BuildSpec)
		value["enable_auto_build"] = aws.BoolValue(config.EnableAutoBuild)
		value["enable_pull_request_preview"] = aws.BoolValue(config.EnablePullRequestPreview)
		value["environment_variables"] = aws.StringValueMap(config.EnvironmentVariables)
		value["framework"] = aws.StringValue(config.Framework)
		value["pull_request_environment_name"] = aws.StringValue(config.PullRequestEnvironmentName)
		value["stage"] = aws.StringValue(config.Stage)
	}

	return []map[string]interface{}{value}
}

func expandAmplifyBasicAuthConfig(v []interface{}) (*bool, *string) {
	enable := false
	credentials := ""

	if len(v) == 0 {
		return aws.Bool(enable), aws.String(credentials)
	}

	config := v[0].(map[string]interface{})

	if ev, ok := config["enable_basic_auth"].(bool); ok {
		enable = ev
	}

	// build basic_auth_credentials from raw username and password
	username, ok1 := config["username"].(string)
	password, ok2 := config["password"].(string)
	if ok1 && ok2 {
		credentials = encodeAmplifyBasicAuthCredentials(username, password)
	}

	return aws.Bool(enable), aws.String(credentials)
}

func flattenAmplifyBasicAuthConfig(enableBasicAuth *bool, basicAuthCredentials *string) []map[string]interface{} {
	value := make(map[string]interface{})

	if !aws.BoolValue(enableBasicAuth) {
		return nil
	}

	value["enable_basic_auth"] = aws.BoolValue(enableBasicAuth)

	if basicAuthCredentials != nil {
		// Decode BasicAuthCredentials to username and password
		username, password, _ := decodeAmplifyBasicAuthCredentials(aws.StringValue(basicAuthCredentials))
		value["username"] = username
		value["password"] = password
	}

	return []map[string]interface{}{value}
}

func encodeAmplifyBasicAuthCredentials(username string, password string) string {
	data := fmt.Sprintf("%s:%s", username, password)
	return base64.StdEncoding.EncodeToString([]byte(data))
}

func decodeAmplifyBasicAuthCredentials(encoded string) (string, string, error) {
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", "", err
	}
	s := strings.SplitN(string(data), ":", 2)
	return s[0], s[1], nil
}

func expandAmplifyCustomRules(l []interface{}) []*amplify.CustomRule {
	rules := make([]*amplify.CustomRule, 0)

	for _, v := range l {
		e := v.(map[string]interface{})

		rule := &amplify.CustomRule{}

		if ev, ok := e["condition"].(string); ok && ev != "" {
			rule.Condition = aws.String(ev)
		}

		if ev, ok := e["source"].(string); ok {
			rule.Source = aws.String(ev)
		}

		if ev, ok := e["status"].(string); ok && ev != "" {
			rule.Status = aws.String(ev)
		}

		if ev, ok := e["target"].(string); ok {
			rule.Target = aws.String(ev)
		}

		rules = append(rules, rule)
	}

	return rules
}

func flattenAmplifyCustomRules(rules []*amplify.CustomRule) []map[string]interface{} {
	values := make([]map[string]interface{}, 0)

	for _, rule := range rules {
		value := make(map[string]interface{})
		value["condition"] = aws.StringValue(rule.Condition)
		value["source"] = aws.StringValue(rule.Source)
		value["status"] = aws.StringValue(rule.Status)
		value["target"] = aws.StringValue(rule.Target)
		values = append(values, value)
	}

	return values
}
