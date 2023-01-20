package apigateway

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

func ResourceRequestValidator() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceRequestValidatorCreate,
		ReadWithoutTimeout:   resourceRequestValidatorRead,
		UpdateWithoutTimeout: resourceRequestValidatorUpdate,
		DeleteWithoutTimeout: resourceRequestValidatorDelete,
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				idParts := strings.Split(d.Id(), "/")
				if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
					return nil, fmt.Errorf("Unexpected format of ID (%q), expected REST-API-ID/REQUEST-VALIDATOR-ID", d.Id())
				}
				restApiID := idParts[0]
				requestValidatorID := idParts[1]
				d.Set("rest_api_id", restApiID)
				d.SetId(requestValidatorID)
				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			"rest_api_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"name": {
				Type:     schema.TypeString,
				Required: true,
			},

			"validate_request_body": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			"validate_request_parameters": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
		},
	}
}

func resourceRequestValidatorCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayConn()

	input := apigateway.CreateRequestValidatorInput{
		Name:                      aws.String(d.Get("name").(string)),
		RestApiId:                 aws.String(d.Get("rest_api_id").(string)),
		ValidateRequestBody:       aws.Bool(d.Get("validate_request_body").(bool)),
		ValidateRequestParameters: aws.Bool(d.Get("validate_request_parameters").(bool)),
	}

	out, err := conn.CreateRequestValidatorWithContext(ctx, &input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "Error creating Request Validator: %s", err)
	}

	d.SetId(aws.StringValue(out.Id))

	return diags
}

func resourceRequestValidatorRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayConn()

	input := apigateway.GetRequestValidatorInput{
		RequestValidatorId: aws.String(d.Id()),
		RestApiId:          aws.String(d.Get("rest_api_id").(string)),
	}

	out, err := conn.GetRequestValidatorWithContext(ctx, &input)
	if err != nil {
		if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, apigateway.ErrCodeNotFoundException) {
			log.Printf("[WARN] API Gateway Request Validator (%s) not found, removing from state", d.Id())
			d.SetId("")
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "reading API Gateway Request Validator (%s): %s", d.Id(), err)
	}

	d.Set("name", out.Name)
	d.Set("validate_request_body", out.ValidateRequestBody)
	d.Set("validate_request_parameters", out.ValidateRequestParameters)

	return diags
}

func resourceRequestValidatorUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayConn()
	log.Printf("[DEBUG] Updating Request Validator %s", d.Id())

	operations := make([]*apigateway.PatchOperation, 0)

	if d.HasChange("name") {
		operations = append(operations, &apigateway.PatchOperation{
			Op:    aws.String(apigateway.OpReplace),
			Path:  aws.String("/name"),
			Value: aws.String(d.Get("name").(string)),
		})
	}

	if d.HasChange("validate_request_body") {
		operations = append(operations, &apigateway.PatchOperation{
			Op:    aws.String(apigateway.OpReplace),
			Path:  aws.String("/validateRequestBody"),
			Value: aws.String(fmt.Sprintf("%t", d.Get("validate_request_body").(bool))),
		})
	}

	if d.HasChange("validate_request_parameters") {
		operations = append(operations, &apigateway.PatchOperation{
			Op:    aws.String(apigateway.OpReplace),
			Path:  aws.String("/validateRequestParameters"),
			Value: aws.String(fmt.Sprintf("%t", d.Get("validate_request_parameters").(bool))),
		})
	}

	input := apigateway.UpdateRequestValidatorInput{
		RequestValidatorId: aws.String(d.Id()),
		RestApiId:          aws.String(d.Get("rest_api_id").(string)),
		PatchOperations:    operations,
	}

	_, err := conn.UpdateRequestValidatorWithContext(ctx, &input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating API Gateway Request Validator (%s): %s", d.Id(), err)
	}

	log.Printf("[DEBUG] Updated Request Validator %s", d.Id())

	return append(diags, resourceRequestValidatorRead(ctx, d, meta)...)
}

func resourceRequestValidatorDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayConn()
	log.Printf("[DEBUG] Deleting Request Validator %s", d.Id())

	_, err := conn.DeleteRequestValidatorWithContext(ctx, &apigateway.DeleteRequestValidatorInput{
		RequestValidatorId: aws.String(d.Id()),
		RestApiId:          aws.String(d.Get("rest_api_id").(string)),
	})
	if err != nil {
		// XXX: Figure out a way to delete the method that depends on the request validator first
		// otherwise the validator will be dangling until the API is deleted
		if !strings.Contains(err.Error(), apigateway.ErrCodeConflictException) {
			return sdkdiag.AppendErrorf(diags, "deleting API Gateway Request Validator (%s): %s", d.Id(), err)
		}
	}

	return diags
}
