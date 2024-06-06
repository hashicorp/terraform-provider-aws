// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lambda

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	awstypes "github.com/aws/aws-sdk-go-v2/service/lambda/types"
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
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_lambda_function_url", name="Function URL")
func resourceFunctionURL() *schema.Resource {
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
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: enum.Validate[awstypes.FunctionUrlAuthType](),
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
			names.AttrFunctionARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"function_name": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				DiffSuppressFunc: suppressEquivalentFunctionNameOrARN,
			},
			"function_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"invoke_mode": {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          awstypes.InvokeModeBuffered,
				ValidateDiagFunc: enum.Validate[awstypes.InvokeMode](),
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
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LambdaClient(ctx)

	name := d.Get("function_name").(string)
	qualifier := d.Get("qualifier").(string)
	id := functionURLCreateResourceID(name, qualifier)
	authorizationType := awstypes.FunctionUrlAuthType(d.Get("authorization_type").(string))
	input := &lambda.CreateFunctionUrlConfigInput{
		AuthType:     authorizationType,
		FunctionName: aws.String(name),
		InvokeMode:   awstypes.InvokeMode(d.Get("invoke_mode").(string)),
	}

	if qualifier != "" {
		input.Qualifier = aws.String(qualifier)
	}

	if v, ok := d.GetOk("cors"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.Cors = expandCors(v.([]interface{})[0].(map[string]interface{}))
	}

	_, err := conn.CreateFunctionUrlConfig(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Lambda Function URL (%s): %s", id, err)
	}

	d.SetId(id)

	if authorizationType == awstypes.FunctionUrlAuthTypeNone {
		input := &lambda.AddPermissionInput{
			Action:              aws.String("lambda:InvokeFunctionUrl"),
			FunctionName:        aws.String(name),
			FunctionUrlAuthType: authorizationType,
			Principal:           aws.String("*"),
			StatementId:         aws.String("FunctionURLAllowPublicAccess"),
		}

		if qualifier != "" {
			input.Qualifier = aws.String(qualifier)
		}

		_, err := conn.AddPermission(ctx, input)

		if err != nil {
			if errs.IsAErrorMessageContains[*awstypes.ResourceConflictException](err, "The statement id (FunctionURLAllowPublicAccess) provided already exists") {
				log.Printf("[DEBUG] function permission statement 'FunctionURLAllowPublicAccess' already exists.")
			} else {
				return sdkdiag.AppendErrorf(diags, "adding Lambda Function URL (%s) permission %s", d.Id(), err)
			}
		}
	}

	return append(diags, resourceFunctionURLRead(ctx, d, meta)...)
}

func resourceFunctionURLRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LambdaClient(ctx)

	name, qualifier, err := functionURLParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	output, err := findFunctionURLByTwoPartKey(ctx, conn, name, qualifier)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Lambda Function URL %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Lambda Function URL (%s): %s", d.Id(), err)
	}

	functionURL := aws.ToString(output.FunctionUrl)
	d.Set("authorization_type", output.AuthType)
	if output.Cors != nil {
		if err := d.Set("cors", []interface{}{flattenCors(output.Cors)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting cors: %s", err)
		}
	} else {
		d.Set("cors", nil)
	}
	d.Set(names.AttrFunctionARN, output.FunctionArn)
	d.Set("function_name", name)
	d.Set("function_url", functionURL)
	d.Set("invoke_mode", output.InvokeMode)
	d.Set("qualifier", qualifier)

	// Function URL endpoints have the following format:
	// https://<url-id>.lambda-url.<region>.on.aws/
	if v, err := url.Parse(functionURL); err != nil {
		return sdkdiag.AppendErrorf(diags, "parsing URL (%s): %s", functionURL, err)
	} else if v := strings.Split(v.Host, "."); len(v) > 0 {
		d.Set("url_id", v[0])
	} else {
		d.Set("url_id", nil)
	}

	return diags
}

func resourceFunctionURLUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LambdaClient(ctx)

	name, qualifier, err := functionURLParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	input := &lambda.UpdateFunctionUrlConfigInput{
		FunctionName: aws.String(name),
	}

	if qualifier != "" {
		input.Qualifier = aws.String(qualifier)
	}

	if d.HasChange("authorization_type") {
		input.AuthType = awstypes.FunctionUrlAuthType(d.Get("authorization_type").(string))
	}

	if d.HasChange("cors") {
		if v, ok := d.GetOk("cors"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			input.Cors = expandCors(v.([]interface{})[0].(map[string]interface{}))
		} else {
			input.Cors = &awstypes.Cors{}
		}
	}

	if d.HasChange("invoke_mode") {
		input.InvokeMode = awstypes.InvokeMode(d.Get("invoke_mode").(string))
	}

	_, err = conn.UpdateFunctionUrlConfig(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Lambda Function URL (%s): %s", d.Id(), err)
	}

	return append(diags, resourceFunctionURLRead(ctx, d, meta)...)
}

func resourceFunctionURLDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LambdaClient(ctx)

	name, qualifier, err := functionURLParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	input := &lambda.DeleteFunctionUrlConfigInput{
		FunctionName: aws.String(name),
	}

	if qualifier != "" {
		input.Qualifier = aws.String(qualifier)
	}

	log.Printf("[INFO] Deleting Lambda Function URL: %s", d.Id())
	_, err = conn.DeleteFunctionUrlConfig(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Lambda Function URL (%s): %s", d.Id(), err)
	}

	return diags
}

func findFunctionURLByTwoPartKey(ctx context.Context, conn *lambda.Client, name, qualifier string) (*lambda.GetFunctionUrlConfigOutput, error) {
	input := &lambda.GetFunctionUrlConfigInput{
		FunctionName: aws.String(name),
	}
	if qualifier != "" {
		input.Qualifier = aws.String(qualifier)
	}

	return findFunctionURL(ctx, conn, input)
}

func findFunctionURL(ctx context.Context, conn *lambda.Client, input *lambda.GetFunctionUrlConfigInput) (*lambda.GetFunctionUrlConfigOutput, error) {
	output, err := conn.GetFunctionUrlConfig(ctx, input)

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

const functionURLResourceIDSeparator = "/"

func functionURLCreateResourceID(functionName, qualifier string) string {
	if qualifier == "" {
		return functionName
	}

	parts := []string{functionName, qualifier}
	id := strings.Join(parts, functionURLResourceIDSeparator)

	return id
}

func functionURLParseResourceID(id string) (string, string, error) {
	parts := strings.Split(id, functionURLResourceIDSeparator)

	if len(parts) == 1 && parts[0] != "" {
		return parts[0], "", nil
	}
	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected FUNCTION-NAME%[2]sQUALIFIER or FUNCTION-NAME", id, functionURLResourceIDSeparator)
}

func expandCors(tfMap map[string]interface{}) *awstypes.Cors {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.Cors{}

	if v, ok := tfMap["allow_credentials"].(bool); ok {
		apiObject.AllowCredentials = aws.Bool(v)
	}

	if v, ok := tfMap["allow_headers"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.AllowHeaders = flex.ExpandStringValueSet(v)
	}

	if v, ok := tfMap["allow_methods"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.AllowMethods = flex.ExpandStringValueSet(v)
	}

	if v, ok := tfMap["allow_origins"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.AllowOrigins = flex.ExpandStringValueSet(v)
	}

	if v, ok := tfMap["expose_headers"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.ExposeHeaders = flex.ExpandStringValueSet(v)
	}

	if v, ok := tfMap["max_age"].(int); ok && v != 0 {
		apiObject.MaxAge = aws.Int32(int32(v))
	}

	return apiObject
}

func flattenCors(apiObject *awstypes.Cors) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.AllowCredentials; v != nil {
		tfMap["allow_credentials"] = aws.ToBool(v)
	}

	if v := apiObject.AllowHeaders; v != nil {
		tfMap["allow_headers"] = v
	}

	if v := apiObject.AllowMethods; v != nil {
		tfMap["allow_methods"] = v
	}

	if v := apiObject.AllowOrigins; v != nil {
		tfMap["allow_origins"] = v
	}

	if v := apiObject.ExposeHeaders; v != nil {
		tfMap["expose_headers"] = v
	}

	if v := apiObject.MaxAge; v != nil {
		tfMap["max_age"] = aws.ToInt32(v)
	}

	return tfMap
}
