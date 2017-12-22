package aws

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appsync"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsAppsyncGraphqlApi() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsAppsyncGraphqlApiCreate,
		Read:   resourceAwsAppsyncGraphqlApiRead,
		Update: resourceAwsAppsyncGraphqlApiUpdate,
		Delete: resourceAwsAppsyncGraphqlApiDelete,

		Schema: map[string]*schema.Schema{
			"authentication_type": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
					value := strings.ToUpper(v.(string))
					validTypes := []string{"API_KEY", "AWS_IAM", "AMAZON_COGNITO_USER_POOLS"}
					for _, str := range validTypes {
						if value == str {
							return
						}
					}
					errors = append(errors, fmt.Errorf("expected %s to be one of %v, got %s", k, validTypes, value))
					return
				},
				StateFunc: func(v interface{}) string {
					return strings.ToUpper(v.(string))
				},
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
					value := v.(string)
					if !regexp.MustCompile(`[_A-Za-z][_0-9A-Za-z]*`).MatchString(value) {
						errors = append(errors, fmt.Errorf("%q must match [_A-Za-z][_0-9A-Za-z]*", k))
					}
					return
				},
			},
			"user_pool_config": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"app_id_client_regex": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"aws_region": {
							Type:     schema.TypeString,
							Required: true,
						},
						"default_action": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
								value := strings.ToUpper(v.(string))
								validTypes := []string{"ALLOW", "DENY"}
								for _, str := range validTypes {
									if value == str {
										return
									}
								}
								errors = append(errors, fmt.Errorf("expected %s to be one of %v, got %s", k, validTypes, value))
								return
							},
							StateFunc: func(v interface{}) string {
								return strings.ToUpper(v.(string))
							},
						},
						"user_pool_id": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsAppsyncGraphqlApiCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).appsyncconn

	input := &appsync.CreateGraphqlApiInput{
		AuthenticationType: aws.String(d.Get("authentication_type").(string)),
		Name:               aws.String(d.Get("name").(string)),
		UserPoolConfig:     expandAppsyncGraphqlApiUserPoolConfig(d.Get("user_pool_config").([]interface{})),
	}

	resp, err := conn.CreateGraphqlApi(input)
	if err != nil {
		return err
	}

	d.SetId(*resp.GraphqlApi.ApiId)
	d.Set("arn", resp.GraphqlApi.Arn)
	return nil
}

func resourceAwsAppsyncGraphqlApiRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).appsyncconn

	input := &appsync.GetGraphqlApiInput{
		ApiId: aws.String(d.Id()),
	}

	resp, err := conn.GetGraphqlApi(input)
	if err != nil {
		if isAWSErr(err, appsync.ErrCodeNotFoundException, "") {
			d.SetId("")
			return nil
		}
		return err
	}

	d.Set("authentication_type", resp.GraphqlApi.AuthenticationType)
	d.Set("name", resp.GraphqlApi.Name)
	d.Set("user_pool_config", flattenAppsyncGraphqlApiUserPoolConfig(resp.GraphqlApi.UserPoolConfig))
	d.Set("arn", resp.GraphqlApi.Arn)
	return nil
}

func resourceAwsAppsyncGraphqlApiUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).appsyncconn

	input := &appsync.UpdateGraphqlApiInput{
		ApiId: aws.String(d.Id()),
		Name:  aws.String(d.Get("name").(string)),
	}

	if d.HasChange("authentication_type") {
		input.AuthenticationType = aws.String(d.Get("authentication_type").(string))
	}
	if d.HasChange("user_pool_config") {
		input.UserPoolConfig = expandAppsyncGraphqlApiUserPoolConfig(d.Get("user_pool_config").([]interface{}))
	}

	_, err := conn.UpdateGraphqlApi(input)
	if err != nil {
		if isAWSErr(err, appsync.ErrCodeNotFoundException, "") {
			d.SetId("")
			return nil
		}
		return err
	}

	return resourceAwsAppsyncGraphqlApiRead(d, meta)
}

func resourceAwsAppsyncGraphqlApiDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).appsyncconn

	input := &appsync.DeleteGraphqlApiInput{
		ApiId: aws.String(d.Id()),
	}
	_, err := conn.DeleteGraphqlApi(input)
	if err != nil {
		if isAWSErr(err, appsync.ErrCodeNotFoundException, "") {
			d.SetId("")
			return nil
		}
		return err
	}

	d.SetId("")
	return nil
}

func expandAppsyncGraphqlApiUserPoolConfig(config []interface{}) *appsync.UserPoolConfig {
	if len(config) < 1 {
		return nil
	}
	cg := config[0].(map[string]interface{})
	upc := &appsync.UserPoolConfig{
		AwsRegion:     aws.String(cg["aws_region"].(string)),
		DefaultAction: aws.String(cg["default_action"].(string)),
		UserPoolId:    aws.String(cg["user_pool_id"].(string)),
	}
	if v, ok := cg["app_id_client_regex"].(string); ok && v != "" {
		upc.AppIdClientRegex = aws.String(v)
	}
	return upc
}

func flattenAppsyncGraphqlApiUserPoolConfig(upc *appsync.UserPoolConfig) []interface{} {
	if upc == nil {
		return []interface{}{}
	}
	m := make(map[string]interface{}, 1)

	m["aws_region"] = *upc.AwsRegion
	m["default_action"] = *upc.DefaultAction
	m["user_pool_id"] = *upc.UserPoolId
	if upc.AppIdClientRegex != nil {
		m["app_id_client_regex"] = *upc.AppIdClientRegex
	}

	return []interface{}{m}
}
