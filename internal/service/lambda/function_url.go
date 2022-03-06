package lambda

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
)

func ResourceFunctionUrl() *schema.Resource {
	return &schema.Resource{
		Create: resourceFunctionUrlCreate,
		Read:   resourceFunctionUrlRead,
		Update: resourceFunctionUrlUpdate,
		Delete: resourceFunctionUrlDelete,
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
		},

		Importer: &schema.ResourceImporter{
			State: resourceFunctionUrlImport,
		},

		Schema: map[string]*schema.Schema{
			"authorization_type": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(lambda.FunctionUrlAuthType_Values(), false),
			},
			"cors": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"allow_credentials": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"allow_headers": {
							Type:     schema.TypeList,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"allow_methods": {
							Type:     schema.TypeList,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"allow_origins": {
							Type:     schema.TypeList,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"expose_headers": {
							Type:     schema.TypeList,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"max_age": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntAtMost(86400),
						},
					},
				},
			},
			"function_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					// Using function name or ARN should not be shown as a diff.
					// Try to convert the old and new values from ARN to function name
					oldFunctionName, oldFunctionNameErr := GetFunctionNameFromARN(old)
					newFunctionName, newFunctionNameErr := GetFunctionNameFromARN(new)
					return (oldFunctionName == new && oldFunctionNameErr == nil) || (newFunctionName == old && newFunctionNameErr == nil)
				},
			},
			"qualifier": {
				Type:     schema.TypeString,
				ForceNew: true,
				Optional: true,
			},
			"creation_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"function_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"function_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"last_modified_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceFunctionUrlCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).LambdaConn

	params := &lambda.CreateFunctionUrlConfigInput{
		FunctionName: aws.String(d.Get("function_name").(string)),
		AuthType:     aws.String(d.Get("authorization_type").(string)),
	}

	if v, ok := d.GetOk("qualifier"); ok {
		params.Qualifier = aws.String(v.(string))
	}

	if v, ok := d.GetOk("cors"); ok && len(v.([]interface{})) > 0 {
		params.Cors = expandFunctionUrlCorsConfigs(v.([]interface{}))
	}

	output, err := conn.CreateFunctionUrlConfig(params)

	if err != nil {
		return fmt.Errorf("Error creating Lambda function url: %s", err)
	}
	log.Printf("[DEBUG] Creating Lambda Function Url Config Output: %s", output)

	if d.Get("authorization_type").(string) == lambda.FunctionUrlAuthTypeNone {
		permissionParams := &lambda.AddPermissionInput{
			Action:              aws.String("lambda:InvokeFunctionUrl"),
			FunctionName:        aws.String(d.Get("function_name").(string)),
			Principal:           aws.String("*"),
			FunctionUrlAuthType: aws.String(lambda.FunctionUrlAuthTypeNone),
			StatementId:         aws.String("FunctionURLAllowPublicAccess"),
		}
		permissionOutput, permissionErr := conn.AddPermission(permissionParams)

		if permissionErr != nil {
			return fmt.Errorf("Error adding permission for Lambda function url: %s", permissionErr)
		}
		log.Printf("[DEBUG] Add permission for Lambda Function Url Output: %s", permissionOutput)
	}

	d.SetId(aws.StringValue(output.FunctionArn))

	return resourceFunctionUrlRead(d, meta)
}

func resourceFunctionUrlRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).LambdaConn

	input := &lambda.GetFunctionUrlConfigInput{
		FunctionName: aws.String(d.Get("function_name").(string)),
	}

	if v, ok := d.GetOk("qualifier"); ok {
		input.Qualifier = aws.String(v.(string))
	}

	output, err := conn.GetFunctionUrlConfig(input)
	log.Printf("[DEBUG] Getting Lambda Function Url Config Output: %s", output)

	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok && awsErr.Code() == lambda.ErrCodeResourceNotFoundException && !d.IsNewResource() {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error getting Lambda Function Url Config (%s): %w", d.Id(), err)
	}

	if err = d.Set("authorization_type", output.AuthType); err != nil {
		return err
	}
	if err = d.Set("cors", flattenFunctionUrlCorsConfigs(output.Cors)); err != nil {
		return err
	}
	if err = d.Set("creation_time", output.CreationTime); err != nil {
		return err
	}
	if err = d.Set("function_arn", output.FunctionArn); err != nil {
		return err
	}
	if err = d.Set("function_url", output.FunctionUrl); err != nil {
		return err
	}
	if err = d.Set("last_modified_time", output.LastModifiedTime); err != nil {
		return err
	}

	return nil
}

func resourceFunctionUrlUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).LambdaConn

	log.Printf("[DEBUG] Updating Lambda Function Url: %s", d.Id())

	params := &lambda.UpdateFunctionUrlConfigInput{
		FunctionName: aws.String(d.Get("function_name").(string)),
	}

	if v, ok := d.GetOk("qualifier"); ok {
		params.Qualifier = aws.String(v.(string))
	}

	if d.HasChange("authorization_type") {
		params.AuthType = aws.String(d.Get("authorization_type").(string))
	}

	if d.HasChange("cors") {
		params.Cors = expandFunctionUrlCorsConfigs(d.Get("cors").([]interface{}))
	}

	_, err := conn.UpdateFunctionUrlConfig(params)

	if err != nil {
		return fmt.Errorf("error updating Lambda Function Url (%s): %w", d.Id(), err)
	}

	return resourceFunctionUrlRead(d, meta)
}

func resourceFunctionUrlDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).LambdaConn

	log.Printf("[INFO] Deleting Lambda Function Url: %s", d.Id())

	params := &lambda.DeleteFunctionUrlConfigInput{
		FunctionName: aws.String(d.Get("function_name").(string)),
	}

	if v, ok := d.GetOk("qualifier"); ok {
		params.Qualifier = aws.String(v.(string))
	}

	_, err := conn.DeleteFunctionUrlConfig(params)

	if tfawserr.ErrCodeEquals(err, lambda.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting Lambda Function Url (%s): %w", d.Id(), err)
	}

	return nil
}

func expandFunctionUrlCorsConfigs(urlConfigMap []interface{}) *lambda.Cors {
	cors := &lambda.Cors{}
	if len(urlConfigMap) == 1 && urlConfigMap[0] != nil {
		config := urlConfigMap[0].(map[string]interface{})
		cors.AllowCredentials = aws.Bool(config["allow_credentials"].(bool))
		if len(config["allow_headers"].([]interface{})) > 0 {
			cors.AllowHeaders = flex.ExpandStringList(config["allow_headers"].([]interface{}))
		}
		if len(config["allow_methods"].([]interface{})) > 0 {
			cors.AllowMethods = flex.ExpandStringList(config["allow_methods"].([]interface{}))
		}
		if len(config["allow_origins"].([]interface{})) > 0 {
			cors.AllowOrigins = flex.ExpandStringList(config["allow_origins"].([]interface{}))
		}
		if len(config["expose_headers"].([]interface{})) > 0 {
			cors.ExposeHeaders = flex.ExpandStringList(config["expose_headers"].([]interface{}))
		}
		if config["max_age"].(int) > 0 {
			cors.MaxAge = aws.Int64(int64(config["max_age"].(int)))
		}
	}
	return cors
}

func flattenFunctionUrlCorsConfigs(cors *lambda.Cors) []map[string]interface{} {
	settings := make(map[string]interface{})

	if cors == nil {
		return nil
	}

	settings["allow_credentials"] = cors.AllowCredentials
	settings["allow_headers"] = cors.AllowHeaders
	settings["allow_methods"] = cors.AllowMethods
	settings["allow_origins"] = cors.AllowOrigins
	settings["expose_headers"] = cors.ExposeHeaders
	settings["max_age"] = cors.MaxAge

	return []map[string]interface{}{settings}
}

func resourceFunctionUrlImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {

	idSplit := strings.Split(d.Id(), ":")

	functionName := idSplit[len(idSplit)-2]
	qualifier := idSplit[len(idSplit)-1]

	d.Set("function_name", functionName)
	d.Set("qualifier", qualifier)

	return []*schema.ResourceData{d}, nil
}
