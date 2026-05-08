// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package apigateway

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/apigateway"
	"github.com/aws/aws-sdk-go-v2/service/apigateway/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_api_gateway_method_response", name="Method Response")
// @IdAttrFormat("agmr-{rest_api_id}-{resource_id}-{http_method}-{status_code}")
// @IdentityAttribute("rest_api_id")
// @IdentityAttribute("resource_id")
// @IdentityAttribute("http_method")
// @IdentityAttribute("status_code")
// @ImportIDHandler("methodResponseImportID")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/apigateway;apigateway.GetMethodResponseOutput")
// @Testing(preIdentityVersion="v6.40.0")
// @Testing(importStateIdFunc="testAccMethodResponseImportStateIdFunc")
// @Testing(plannableImportAction="NoOp")
func resourceMethodResponse() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceMethodResponseCreate,
		ReadWithoutTimeout:   resourceMethodResponseRead,
		UpdateWithoutTimeout: resourceMethodResponseUpdate,
		DeleteWithoutTimeout: resourceMethodResponseDelete,

		Schema: map[string]*schema.Schema{
			"http_method": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validHTTPMethod(),
			},
			names.AttrResourceID: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"response_models": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"response_parameters": {
				Type:     schema.TypeMap,
				Elem:     &schema.Schema{Type: schema.TypeBool},
				Optional: true,
			},
			"rest_api_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrStatusCode: {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceMethodResponseCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayClient(ctx)

	input := apigateway.PutMethodResponseInput{
		HttpMethod: aws.String(d.Get("http_method").(string)),
		ResourceId: aws.String(d.Get(names.AttrResourceID).(string)),
		RestApiId:  aws.String(d.Get("rest_api_id").(string)),
		StatusCode: aws.String(d.Get(names.AttrStatusCode).(string)),
	}

	if v, ok := d.GetOk("response_models"); ok && len(v.(map[string]any)) > 0 {
		input.ResponseModels = flex.ExpandStringValueMap(v.(map[string]any))
	}

	if v, ok := d.GetOk("response_parameters"); ok && len(v.(map[string]any)) > 0 {
		input.ResponseParameters = flex.ExpandBoolValueMap(v.(map[string]any))
	}

	mutexKey := "api-gateway-method-response"
	conns.GlobalMutexKV.Lock(mutexKey)
	defer conns.GlobalMutexKV.Unlock(mutexKey)

	const (
		timeout = 2 * time.Minute
	)
	_, err := tfresource.RetryWhenIsA[any, *types.ConflictException](ctx, timeout, func(ctx context.Context) (any, error) {
		return conn.PutMethodResponse(ctx, &input)
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating API Gateway Method Response: %s", err)
	}

	d.SetId(resourceMethodResponseIDAttr(d.Get("rest_api_id").(string), d.Get(names.AttrResourceID).(string), d.Get("http_method").(string), d.Get(names.AttrStatusCode).(string)))

	return diags
}

func resourceMethodResponseRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayClient(ctx)

	methodResponse, err := findMethodResponseByFourPartKey(ctx, conn, d.Get("http_method").(string), d.Get(names.AttrResourceID).(string), d.Get("rest_api_id").(string), d.Get(names.AttrStatusCode).(string))

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] API Gateway Method Response (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading API Gateway Method Response (%s): %s", d.Id(), err)
	}

	resourceMethodResponseFlatten(d, methodResponse)

	return diags
}

func resourceMethodResponseFlatten(d *schema.ResourceData, methodResponse *apigateway.GetMethodResponseOutput) {
	d.Set("response_models", methodResponse.ResponseModels)
	d.Set("response_parameters", methodResponse.ResponseParameters)
}

func resourceMethodResponseUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayClient(ctx)

	operations := make([]types.PatchOperation, 0)

	if d.HasChange("response_models") {
		operations = append(operations, expandRequestResponseModelOperations(d, "response_models", "responseModels")...)
	}

	if d.HasChange("response_parameters") {
		operations = append(operations, expandMethodParametersOperations(d, "response_parameters", "responseParameters")...)
	}

	input := apigateway.UpdateMethodResponseInput{
		HttpMethod:      aws.String(d.Get("http_method").(string)),
		PatchOperations: operations,
		ResourceId:      aws.String(d.Get(names.AttrResourceID).(string)),
		RestApiId:       aws.String(d.Get("rest_api_id").(string)),
		StatusCode:      aws.String(d.Get(names.AttrStatusCode).(string)),
	}

	_, err := conn.UpdateMethodResponse(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating API Gateway Method Response (%s): %s", d.Id(), err)
	}

	return append(diags, resourceMethodResponseRead(ctx, d, meta)...)
}

func resourceMethodResponseDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayClient(ctx)

	log.Printf("[DEBUG] Deleting API Gateway Method Response: %s", d.Id())
	input := apigateway.DeleteMethodResponseInput{
		HttpMethod: aws.String(d.Get("http_method").(string)),
		ResourceId: aws.String(d.Get(names.AttrResourceID).(string)),
		RestApiId:  aws.String(d.Get("rest_api_id").(string)),
		StatusCode: aws.String(d.Get(names.AttrStatusCode).(string)),
	}
	_, err := conn.DeleteMethodResponse(ctx, &input)

	if errs.IsA[*types.NotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting API Gateway Method Response (%s): %s", d.Id(), err)
	}

	return diags
}

func findMethodResponseByFourPartKey(ctx context.Context, conn *apigateway.Client, httpMethod, resourceID, apiID, statusCode string) (*apigateway.GetMethodResponseOutput, error) {
	input := apigateway.GetMethodResponseInput{
		HttpMethod: aws.String(httpMethod),
		ResourceId: aws.String(resourceID),
		RestApiId:  aws.String(apiID),
		StatusCode: aws.String(statusCode),
	}

	output, err := conn.GetMethodResponse(ctx, &input)

	if errs.IsA[*types.NotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output, nil
}

var _ inttypes.SDKv2ImportID = methodResponseImportID{}

type methodResponseImportID struct{}

func methodResponseCreateImportID(restApiID, resourceID, httpMethod, statusCode string) string {
	return restApiID + "/" + resourceID + "/" + httpMethod + "/" + statusCode
}

func resourceMethodResponseIDAttr(restApiID, resourceID, httpMethod, statusCode string) string {
	return fmt.Sprintf("agmr-%s-%s-%s-%s", restApiID, resourceID, httpMethod, statusCode)
}

func (methodResponseImportID) Create(d *schema.ResourceData) string {
	return methodResponseCreateImportID(d.Get("rest_api_id").(string), d.Get(names.AttrResourceID).(string), d.Get("http_method").(string), d.Get(names.AttrStatusCode).(string))
}

func (methodResponseImportID) Parse(id string) (string, map[string]any, error) {
	parts := strings.Split(id, "/")
	if len(parts) != 4 || parts[0] == "" || parts[1] == "" || parts[2] == "" || parts[3] == "" {
		return "", nil, fmt.Errorf("id %q should be in the format <rest-api-id>/<resource-id>/<http-method>/<status-code>", id)
	}

	result := map[string]any{
		"rest_api_id":        parts[0],
		names.AttrResourceID: parts[1],
		"http_method":        parts[2],
		names.AttrStatusCode: parts[3],
	}

	return resourceMethodResponseIDAttr(parts[0], parts[1], parts[2], parts[3]), result, nil
}
