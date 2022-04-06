package lambda

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func ResourceFunctionUrl() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceFunctionURLCreate,
		ReadWithoutTimeout:   resourceFunctionURLRead,
		UpdateWithoutTimeout: resourceFunctionURLUpdate,
		DeleteWithoutTimeout: resourceFunctionURLDelete,

		Importer: &schema.ResourceImporter{
			StateContext: func(_ context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				d.Set("function_name", d.Id())

				return []*schema.ResourceData{d}, nil
			},
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
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
							Type:     schema.TypeSet,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"allow_methods": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"allow_origins": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"expose_headers": {
							Type:     schema.TypeSet,
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
			"function_arn": {
				Type:     schema.TypeString,
				Computed: true,
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
			"function_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"qualifier": {
				Type:     schema.TypeString,
				ForceNew: true,
				Optional: true,
			},
		},
	}
}

func resourceFunctionURLCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).LambdaConn

	name := d.Get("function_name").(string)
	input := &lambda.CreateFunctionUrlConfigInput{
		AuthType:     aws.String(d.Get("authorization_type").(string)),
		FunctionName: aws.String(name),
	}

	if v, ok := d.GetOk("cors"); ok && len(v.([]interface{})) > 0 {
		input.Cors = expandFunctionUrlCorsConfigs(v.([]interface{}))
	}

	if v, ok := d.GetOk("qualifier"); ok {
		input.Qualifier = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating Lambda Function URL: %s", input)
	_, err := conn.CreateFunctionUrlConfigWithContext(ctx, input)

	if err != nil {
		return diag.Errorf("error creating Lambda Function URL (%s): %s", name, err)
	}

	d.SetId(name)

	if v := d.Get("authorization_type").(string); v == lambda.FunctionUrlAuthTypeNone {
		input := &lambda.AddPermissionInput{
			Action:              aws.String("lambda:InvokeFunctionUrl"),
			FunctionName:        aws.String(d.Id()),
			FunctionUrlAuthType: aws.String(v),
			Principal:           aws.String("*"),
			StatementId:         aws.String("FunctionURLAllowPublicAccess"),
		}

		log.Printf("[DEBUG] Adding Lambda Permission: %s", input)
		_, err := conn.AddPermissionWithContext(ctx, input)

		if err != nil {
			return diag.Errorf("error adding Lambda Function URL (%s) permission %s", d.Id(), err)
		}
	}

	return resourceFunctionURLRead(ctx, d, meta)
}

func resourceFunctionURLRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).LambdaConn

	output, err := FindFunctionURLByNameAndQualifier(ctx, conn, d.Id(), d.Get("qualifier").(string))

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Lambda Function URL %s not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("error reading Lambda Function URL (%s): %w", d.Id(), err)
	}

	d.Set("authorization_type", output.AuthType)
	d.Set("cors", flattenFunctionUrlCorsConfigs(output.Cors))
	d.Set("function_arn", output.FunctionArn)
	d.Set("function_url", output.FunctionUrl)

	return nil
}

func resourceFunctionURLUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).LambdaConn

	input := &lambda.UpdateFunctionUrlConfigInput{
		FunctionName: aws.String(d.Get("function_name").(string)),
	}

	if d.HasChange("authorization_type") {
		input.AuthType = aws.String(d.Get("authorization_type").(string))
	}

	if d.HasChange("cors") {
		input.Cors = expandFunctionUrlCorsConfigs(d.Get("cors").([]interface{}))
	}

	if v, ok := d.GetOk("qualifier"); ok {
		input.Qualifier = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Updating Lambda Function URL: %s", input)
	_, err := conn.UpdateFunctionUrlConfigWithContext(ctx, input)

	if err != nil {
		return diag.Errorf("error updating Lambda Function URL (%s): %s", d.Id(), err)
	}

	return resourceFunctionURLRead(ctx, d, meta)
}

func resourceFunctionURLDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).LambdaConn

	input := &lambda.DeleteFunctionUrlConfigInput{
		FunctionName: aws.String(d.Id()),
	}

	if v, ok := d.GetOk("qualifier"); ok {
		input.Qualifier = aws.String(v.(string))
	}

	log.Printf("[INFO] Deleting Lambda Function URL: %s", d.Id())
	_, err := conn.DeleteFunctionUrlConfigWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, lambda.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return diag.Errorf("error deleting Lambda Function URL (%s): %s", d.Id(), err)
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

func FindFunctionURLByNameAndQualifier(ctx context.Context, conn *lambda.Lambda, name, qualifier string) (*lambda.GetFunctionUrlConfigOutput, error) {
	input := &lambda.GetFunctionUrlConfigInput{
		FunctionName: aws.String(name),
	}

	if qualifier != "" {
		input.Qualifier = aws.String(qualifier)
	}

	output, err := conn.GetFunctionUrlConfigWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, lambda.ErrCodeResourceNotFoundException) {
		return nil, &resource.NotFoundError{
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
