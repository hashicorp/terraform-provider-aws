package sagemaker

import (
	"fmt"
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sagemaker"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func ResourceWorkforce() *schema.Resource {
	return &schema.Resource{
		Create: resourceWorkforceCreate,
		Read:   resourceWorkforceRead,
		Update: resourceWorkforceUpdate,
		Delete: resourceWorkforceDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cognito_config": {
				Type:         schema.TypeList,
				Optional:     true,
				ForceNew:     true,
				MaxItems:     1,
				ExactlyOneOf: []string{"oidc_config", "cognito_config"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"client_id": {
							Type:     schema.TypeString,
							Required: true,
						},
						"user_pool": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"oidc_config": {
				Type:         schema.TypeList,
				Optional:     true,
				MaxItems:     1,
				ExactlyOneOf: []string{"oidc_config", "cognito_config"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"authorization_endpoint": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.All(
								validation.StringLenBetween(1, 500),
								validation.IsURLWithHTTPS,
							),
						},
						"client_id": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(1, 1024),
						},
						"client_secret": {
							Type:         schema.TypeString,
							Required:     true,
							Sensitive:    true,
							ValidateFunc: validation.StringLenBetween(1, 1024),
						},
						"issuer": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.All(
								validation.StringLenBetween(1, 500),
								validation.IsURLWithHTTPS,
							)},
						"jwks_uri": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.All(
								validation.StringLenBetween(1, 500),
								validation.IsURLWithHTTPS,
							)},
						"logout_endpoint": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.All(
								validation.StringLenBetween(1, 500),
								validation.IsURLWithHTTPS,
							)},
						"token_endpoint": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.All(
								validation.StringLenBetween(1, 500),
								validation.IsURLWithHTTPS,
							)},
						"user_info_endpoint": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.All(
								validation.StringLenBetween(1, 500),
								validation.IsURLWithHTTPS,
							),
						},
					},
				},
			},
			"source_ip_config": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cidrs": {
							Type:     schema.TypeSet,
							Required: true,
							MaxItems: 10,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validation.IsCIDR,
							},
						},
					},
				},
			},
			"subdomain": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"workforce_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 63),
					validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9\-])*$`), "Valid characters are a-z, A-Z, 0-9, and - (hyphen)."),
				),
			},
		},
	}
}

func resourceWorkforceCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SageMakerConn

	name := d.Get("workforce_name").(string)
	input := &sagemaker.CreateWorkforceInput{
		WorkforceName: aws.String(name),
	}

	if v, ok := d.GetOk("cognito_config"); ok {
		input.CognitoConfig = expandWorkforceCognitoConfig(v.([]interface{}))
	}

	if v, ok := d.GetOk("oidc_config"); ok {
		input.OidcConfig = expandWorkforceOidcConfig(v.([]interface{}))
	}

	if v, ok := d.GetOk("source_ip_config"); ok {
		input.SourceIpConfig = expandWorkforceSourceIpConfig(v.([]interface{}))
	}

	log.Printf("[DEBUG] Creating SageMaker Workforce: %s", input)
	_, err := conn.CreateWorkforce(input)

	if err != nil {
		return fmt.Errorf("error creating SageMaker Workforce (%s): %w", name, err)
	}

	d.SetId(name)

	return resourceWorkforceRead(d, meta)
}

func resourceWorkforceRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SageMakerConn

	workforce, err := FindWorkforceByName(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] SageMaker Workforce (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading SageMaker Workforce (%s): %w", d.Id(), err)
	}

	d.Set("arn", workforce.WorkforceArn)
	d.Set("subdomain", workforce.SubDomain)
	d.Set("workforce_name", workforce.WorkforceName)

	if err := d.Set("cognito_config", flattenWorkforceCognitoConfig(workforce.CognitoConfig)); err != nil {
		return fmt.Errorf("error setting cognito_config : %w", err)
	}

	if workforce.OidcConfig != nil {
		if err := d.Set("oidc_config", flattenWorkforceOidcConfig(workforce.OidcConfig, d.Get("oidc_config.0.client_secret").(string))); err != nil {
			return fmt.Errorf("error setting oidc_config: %w", err)
		}
	}

	if err := d.Set("source_ip_config", flattenWorkforceSourceIpConfig(workforce.SourceIpConfig)); err != nil {
		return fmt.Errorf("error setting source_ip_config: %w", err)
	}

	return nil
}

func resourceWorkforceUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SageMakerConn

	input := &sagemaker.UpdateWorkforceInput{
		WorkforceName: aws.String(d.Id()),
	}

	if d.HasChange("source_ip_config") {
		input.SourceIpConfig = expandWorkforceSourceIpConfig(d.Get("source_ip_config").([]interface{}))
	}

	if d.HasChange("oidc_config") {
		input.OidcConfig = expandWorkforceOidcConfig(d.Get("oidc_config").([]interface{}))
	}

	log.Printf("[DEBUG] Updating SageMaker Workforce: %s", input)
	_, err := conn.UpdateWorkforce(input)

	if err != nil {
		return fmt.Errorf("error updating SageMaker Workforce (%s): %w", d.Id(), err)
	}

	return resourceWorkforceRead(d, meta)
}

func resourceWorkforceDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SageMakerConn

	log.Printf("[DEBUG] Deleting SageMaker Workforce: %s", d.Id())
	_, err := conn.DeleteWorkforce(&sagemaker.DeleteWorkforceInput{
		WorkforceName: aws.String(d.Id()),
	})

	if tfawserr.ErrMessageContains(err, "ValidationException", "No workforce") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting SageMaker Workforce (%s): %w", d.Id(), err)
	}

	return nil
}

func expandWorkforceSourceIpConfig(l []interface{}) *sagemaker.SourceIpConfig {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	config := &sagemaker.SourceIpConfig{
		Cidrs: flex.ExpandStringSet(m["cidrs"].(*schema.Set)),
	}

	return config
}

func flattenWorkforceSourceIpConfig(config *sagemaker.SourceIpConfig) []map[string]interface{} {
	if config == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"cidrs": flex.FlattenStringSet(config.Cidrs),
	}

	return []map[string]interface{}{m}
}

func expandWorkforceCognitoConfig(l []interface{}) *sagemaker.CognitoConfig {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	config := &sagemaker.CognitoConfig{
		ClientId: aws.String(m["client_id"].(string)),
		UserPool: aws.String(m["user_pool"].(string)),
	}

	return config
}

func flattenWorkforceCognitoConfig(config *sagemaker.CognitoConfig) []map[string]interface{} {
	if config == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"client_id": aws.StringValue(config.ClientId),
		"user_pool": aws.StringValue(config.UserPool),
	}

	return []map[string]interface{}{m}
}

func expandWorkforceOidcConfig(l []interface{}) *sagemaker.OidcConfig {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	config := &sagemaker.OidcConfig{
		AuthorizationEndpoint: aws.String(m["authorization_endpoint"].(string)),
		ClientId:              aws.String(m["client_id"].(string)),
		ClientSecret:          aws.String(m["client_secret"].(string)),
		Issuer:                aws.String(m["issuer"].(string)),
		JwksUri:               aws.String(m["jwks_uri"].(string)),
		LogoutEndpoint:        aws.String(m["logout_endpoint"].(string)),
		TokenEndpoint:         aws.String(m["token_endpoint"].(string)),
		UserInfoEndpoint:      aws.String(m["user_info_endpoint"].(string)),
	}

	return config
}

func flattenWorkforceOidcConfig(config *sagemaker.OidcConfigForResponse, clientSecret string) []map[string]interface{} {
	if config == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"authorization_endpoint": aws.StringValue(config.AuthorizationEndpoint),
		"client_id":              aws.StringValue(config.ClientId),
		"client_secret":          clientSecret,
		"issuer":                 aws.StringValue(config.Issuer),
		"jwks_uri":               aws.StringValue(config.JwksUri),
		"logout_endpoint":        aws.StringValue(config.LogoutEndpoint),
		"token_endpoint":         aws.StringValue(config.TokenEndpoint),
		"user_info_endpoint":     aws.StringValue(config.UserInfoEndpoint),
	}

	return []map[string]interface{}{m}
}
