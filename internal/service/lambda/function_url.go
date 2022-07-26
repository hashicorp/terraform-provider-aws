package lambda

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"strings"
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

func ResourceFunctionURL() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceFunctionURLCreate,
		ReadWithoutTimeout:   resourceFunctionURLRead,
		UpdateWithoutTimeout: resourceFunctionURLUpdate,
		DeleteWithoutTimeout: resourceFunctionURLDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
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
			"url_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceFunctionURLCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).LambdaConn

	name := d.Get("function_name").(string)
	qualifier := d.Get("qualifier").(string)
	id := FunctionURLCreateResourceID(name, qualifier)
	input := &lambda.CreateFunctionUrlConfigInput{
		AuthType:     aws.String(d.Get("authorization_type").(string)),
		FunctionName: aws.String(name),
	}

	if qualifier != "" {
		input.Qualifier = aws.String(qualifier)
	}

	if v, ok := d.GetOk("cors"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.Cors = expandCors(v.([]interface{})[0].(map[string]interface{}))
	}

	log.Printf("[DEBUG] Creating Lambda Function URL: %s", input)
	_, err := conn.CreateFunctionUrlConfigWithContext(ctx, input)

	if err != nil {
		return diag.Errorf("error creating Lambda Function URL (%s): %s", id, err)
	}

	d.SetId(id)

	if v := d.Get("authorization_type").(string); v == lambda.FunctionUrlAuthTypeNone {
		input := &lambda.AddPermissionInput{
			Action:              aws.String("lambda:InvokeFunctionUrl"),
			FunctionName:        aws.String(name),
			FunctionUrlAuthType: aws.String(v),
			Principal:           aws.String("*"),
			StatementId:         aws.String("FunctionURLAllowPublicAccess"),
		}

		if qualifier != "" {
			input.Qualifier = aws.String(qualifier)
		}

		log.Printf("[DEBUG] Adding Lambda Permission: %s", input)
		_, err := conn.AddPermissionWithContext(ctx, input)

		if err != nil {
			if tfawserr.ErrMessageContains(err, lambda.ErrCodeResourceConflictException, "The statement id (FunctionURLAllowPublicAccess) provided already exists") {
				log.Printf("[DEBUG] function permission statement 'FunctionURLAllowPublicAccess' already exists.")
			} else {
				return diag.Errorf("error adding Lambda Function URL (%s) permission %s", d.Id(), err)
			}
		}
	}

	return resourceFunctionURLRead(ctx, d, meta)
}

func resourceFunctionURLRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).LambdaConn

	name, qualifier, err := FunctionURLParseResourceID(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	output, err := FindFunctionURLByNameAndQualifier(ctx, conn, name, qualifier)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Lambda Function URL %s not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("error reading Lambda Function URL (%s): %s", d.Id(), err)
	}

	functionURL := aws.StringValue(output.FunctionUrl)

	d.Set("authorization_type", output.AuthType)
	if output.Cors != nil {
		if err := d.Set("cors", []interface{}{flattenCors(output.Cors)}); err != nil {
			return diag.Errorf("error setting cors: %s", err)
		}
	} else {
		d.Set("cors", nil)
	}
	d.Set("function_arn", output.FunctionArn)
	d.Set("function_name", name)
	d.Set("function_url", functionURL)
	d.Set("qualifier", qualifier)

	// Function URL endpoints have the following format:
	// https://<url-id>.lambda-url.<region>.on.aws
	if v, err := url.Parse(functionURL); err != nil {
		return diag.Errorf("error parsing URL (%s): %s", functionURL, err)
	} else if v := strings.Split(v.Host, "."); len(v) > 0 {
		d.Set("url_id", v[0])
	} else {
		d.Set("url_id", nil)
	}

	return nil
}

func resourceFunctionURLUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).LambdaConn

	name, qualifier, err := FunctionURLParseResourceID(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	input := &lambda.UpdateFunctionUrlConfigInput{
		FunctionName: aws.String(name),
	}

	if qualifier != "" {
		input.Qualifier = aws.String(qualifier)
	}

	if d.HasChange("authorization_type") {
		input.AuthType = aws.String(d.Get("authorization_type").(string))
	}

	if d.HasChange("cors") {
		if v, ok := d.GetOk("cors"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			input.Cors = expandCors(v.([]interface{})[0].(map[string]interface{}))
		}
	}

	log.Printf("[DEBUG] Updating Lambda Function URL: %s", input)
	_, err = conn.UpdateFunctionUrlConfigWithContext(ctx, input)

	if err != nil {
		return diag.Errorf("error updating Lambda Function URL (%s): %s", d.Id(), err)
	}

	return resourceFunctionURLRead(ctx, d, meta)
}

func resourceFunctionURLDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).LambdaConn

	name, qualifier, err := FunctionURLParseResourceID(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	input := &lambda.DeleteFunctionUrlConfigInput{
		FunctionName: aws.String(name),
	}

	if qualifier != "" {
		input.Qualifier = aws.String(qualifier)
	}

	log.Printf("[INFO] Deleting Lambda Function URL: %s", d.Id())
	_, err = conn.DeleteFunctionUrlConfigWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, lambda.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return diag.Errorf("error deleting Lambda Function URL (%s): %s", d.Id(), err)
	}

	return nil
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

const functionURLResourceIDSeparator = "/"

func FunctionURLCreateResourceID(functionName, qualifier string) string {
	if qualifier == "" {
		return functionName
	}

	parts := []string{functionName, qualifier}
	id := strings.Join(parts, functionURLResourceIDSeparator)

	return id
}

func FunctionURLParseResourceID(id string) (string, string, error) {
	parts := strings.Split(id, functionURLResourceIDSeparator)

	if len(parts) == 1 && parts[0] != "" {
		return parts[0], "", nil
	}
	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected FUNCTION-NAME%[2]qQUALIFIER or FUNCTION-NAME", id, functionURLResourceIDSeparator)
}

func expandCors(tfMap map[string]interface{}) *lambda.Cors {
	if tfMap == nil {
		return nil
	}

	apiObject := &lambda.Cors{}

	if v, ok := tfMap["allow_credentials"].(bool); ok {
		apiObject.AllowCredentials = aws.Bool(v)
	}

	if v, ok := tfMap["allow_headers"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.AllowHeaders = flex.ExpandStringSet(v)
	}

	if v, ok := tfMap["allow_methods"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.AllowMethods = flex.ExpandStringSet(v)
	}

	if v, ok := tfMap["allow_origins"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.AllowOrigins = flex.ExpandStringSet(v)
	}

	if v, ok := tfMap["expose_headers"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.ExposeHeaders = flex.ExpandStringSet(v)
	}

	if v, ok := tfMap["max_age"].(int); ok && v != 0 {
		apiObject.MaxAge = aws.Int64(int64(v))
	}

	return apiObject
}

func flattenCors(apiObject *lambda.Cors) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.AllowCredentials; v != nil {
		tfMap["allow_credentials"] = aws.BoolValue(v)
	}

	if v := apiObject.AllowHeaders; v != nil {
		tfMap["allow_headers"] = aws.StringValueSlice(v)
	}

	if v := apiObject.AllowMethods; v != nil {
		tfMap["allow_methods"] = aws.StringValueSlice(v)
	}

	if v := apiObject.AllowOrigins; v != nil {
		tfMap["allow_origins"] = aws.StringValueSlice(v)
	}

	if v := apiObject.ExposeHeaders; v != nil {
		tfMap["expose_headers"] = aws.StringValueSlice(v)
	}

	if v := apiObject.MaxAge; v != nil {
		tfMap["max_age"] = aws.Int64Value(v)
	}

	return tfMap
}
