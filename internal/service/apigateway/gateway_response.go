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

func ResourceGatewayResponse() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceGatewayResponsePut,
		ReadWithoutTimeout:   resourceGatewayResponseRead,
		UpdateWithoutTimeout: resourceGatewayResponsePut,
		DeleteWithoutTimeout: resourceGatewayResponseDelete,
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				idParts := strings.Split(d.Id(), "/")
				if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
					return nil, fmt.Errorf("Unexpected format of ID (%q), expected REST-API-ID/RESPONSE-TYPE", d.Id())
				}
				restApiID := idParts[0]
				responseType := idParts[1]
				d.Set("response_type", responseType)
				d.Set("rest_api_id", restApiID)
				d.SetId(fmt.Sprintf("aggr-%s-%s", restApiID, responseType))
				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			"rest_api_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"response_type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"status_code": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"response_templates": {
				Type:     schema.TypeMap,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Optional: true,
			},

			"response_parameters": {
				Type:     schema.TypeMap,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Optional: true,
			},
		},
	}
}

func resourceGatewayResponsePut(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayConn()

	templates := make(map[string]string)
	if kv, ok := d.GetOk("response_templates"); ok {
		for k, v := range kv.(map[string]interface{}) {
			templates[k] = v.(string)
		}
	}

	parameters := make(map[string]string)
	if kv, ok := d.GetOk("response_parameters"); ok {
		for k, v := range kv.(map[string]interface{}) {
			parameters[k] = v.(string)
		}
	}

	input := apigateway.PutGatewayResponseInput{
		RestApiId:          aws.String(d.Get("rest_api_id").(string)),
		ResponseType:       aws.String(d.Get("response_type").(string)),
		ResponseTemplates:  aws.StringMap(templates),
		ResponseParameters: aws.StringMap(parameters),
	}

	if v, ok := d.GetOk("status_code"); ok {
		input.StatusCode = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Putting API Gateway Gateway Response: %s", input)

	_, err := conn.PutGatewayResponseWithContext(ctx, &input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "Error putting API Gateway Gateway Response: %s", err)
	}

	d.SetId(fmt.Sprintf("aggr-%s-%s", d.Get("rest_api_id").(string), d.Get("response_type").(string)))
	log.Printf("[DEBUG] API Gateway Gateway Response put (%q)", d.Id())

	return append(diags, resourceGatewayResponseRead(ctx, d, meta)...)
}

func resourceGatewayResponseRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayConn()

	log.Printf("[DEBUG] Reading API Gateway Gateway Response %s", d.Id())
	gatewayResponse, err := conn.GetGatewayResponseWithContext(ctx, &apigateway.GetGatewayResponseInput{
		RestApiId:    aws.String(d.Get("rest_api_id").(string)),
		ResponseType: aws.String(d.Get("response_type").(string)),
	})
	if err != nil {
		if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, apigateway.ErrCodeNotFoundException) {
			log.Printf("[WARN] API Gateway Gateway Response (%s) not found, removing from state", d.Id())
			d.SetId("")
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "reading API Gateway Response (%s): %s", d.Id(), err)
	}

	log.Printf("[DEBUG] Received API Gateway Gateway Response: %s", gatewayResponse)

	d.Set("response_type", gatewayResponse.ResponseType)
	d.Set("status_code", gatewayResponse.StatusCode)
	d.Set("response_templates", aws.StringValueMap(gatewayResponse.ResponseTemplates))
	d.Set("response_parameters", aws.StringValueMap(gatewayResponse.ResponseParameters))

	return diags
}

func resourceGatewayResponseDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayConn()
	log.Printf("[DEBUG] Deleting API Gateway Gateway Response: %s", d.Id())

	_, err := conn.DeleteGatewayResponseWithContext(ctx, &apigateway.DeleteGatewayResponseInput{
		RestApiId:    aws.String(d.Get("rest_api_id").(string)),
		ResponseType: aws.String(d.Get("response_type").(string)),
	})

	if tfawserr.ErrCodeEquals(err, apigateway.ErrCodeNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "Error deleting API Gateway gateway response: %s", err)
	}
	return diags
}
