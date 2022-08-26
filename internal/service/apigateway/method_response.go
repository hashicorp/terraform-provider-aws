package apigateway

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

var resourceMethodResponseMutex = &sync.Mutex{}

func ResourceMethodResponse() *schema.Resource {
	return &schema.Resource{
		Create: resourceMethodResponseCreate,
		Read:   resourceMethodResponseRead,
		Update: resourceMethodResponseUpdate,
		Delete: resourceMethodResponseDelete,
		Importer: &schema.ResourceImporter{
			State: func(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				idParts := strings.Split(d.Id(), "/")
				if len(idParts) != 4 || idParts[0] == "" || idParts[1] == "" || idParts[2] == "" || idParts[3] == "" {
					return nil, fmt.Errorf("Unexpected format of ID (%q), expected REST-API-ID/RESOURCE-ID/HTTP-METHOD/STATUS-CODE", d.Id())
				}
				restApiID := idParts[0]
				resourceID := idParts[1]
				httpMethod := idParts[2]
				statusCode := idParts[3]
				d.Set("http_method", httpMethod)
				d.Set("status_code", statusCode)
				d.Set("resource_id", resourceID)
				d.Set("rest_api_id", restApiID)
				d.SetId(fmt.Sprintf("agmr-%s-%s-%s-%s", restApiID, resourceID, httpMethod, statusCode))
				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			"rest_api_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"resource_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"http_method": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validHTTPMethod(),
			},

			"status_code": {
				Type:     schema.TypeString,
				Required: true,
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
		},
	}
}

func resourceMethodResponseCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).APIGatewayConn

	models := make(map[string]string)
	for k, v := range d.Get("response_models").(map[string]interface{}) {
		models[k] = v.(string)
	}

	parameters := make(map[string]bool)
	if kv, ok := d.GetOk("response_parameters"); ok {
		for k, v := range kv.(map[string]interface{}) {
			parameters[k], ok = v.(bool)
			if !ok {
				value, _ := strconv.ParseBool(v.(string))
				parameters[k] = value
			}
		}
	}

	resourceMethodResponseMutex.Lock()
	defer resourceMethodResponseMutex.Unlock()

	_, err := tfresource.RetryWhenAWSErrCodeEquals(2*time.Minute, func() (interface{}, error) {
		return conn.PutMethodResponse(&apigateway.PutMethodResponseInput{
			HttpMethod:         aws.String(d.Get("http_method").(string)),
			ResourceId:         aws.String(d.Get("resource_id").(string)),
			RestApiId:          aws.String(d.Get("rest_api_id").(string)),
			StatusCode:         aws.String(d.Get("status_code").(string)),
			ResponseModels:     aws.StringMap(models),
			ResponseParameters: aws.BoolMap(parameters),
		})
	}, apigateway.ErrCodeConflictException)

	if err != nil {
		return fmt.Errorf("Error creating API Gateway Method Response: %s", err)
	}

	d.SetId(fmt.Sprintf("agmr-%s-%s-%s-%s", d.Get("rest_api_id").(string), d.Get("resource_id").(string), d.Get("http_method").(string), d.Get("status_code").(string)))
	log.Printf("[DEBUG] API Gateway Method ID: %s", d.Id())

	return nil
}

func resourceMethodResponseRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).APIGatewayConn

	log.Printf("[DEBUG] Reading API Gateway Method Response %s", d.Id())
	methodResponse, err := conn.GetMethodResponse(&apigateway.GetMethodResponseInput{
		HttpMethod: aws.String(d.Get("http_method").(string)),
		ResourceId: aws.String(d.Get("resource_id").(string)),
		RestApiId:  aws.String(d.Get("rest_api_id").(string)),
		StatusCode: aws.String(d.Get("status_code").(string)),
	})
	if err != nil {
		if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, apigateway.ErrCodeNotFoundException) {
			log.Printf("[WARN] API Gateway Method Response (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error reading API Gateway Method Response (%s): %w", d.Id(), err)
	}

	log.Printf("[DEBUG] Received API Gateway Method Response: %s", methodResponse)

	if err := d.Set("response_models", aws.StringValueMap(methodResponse.ResponseModels)); err != nil {
		return fmt.Errorf("error setting response_models: %s", err)
	}

	if err := d.Set("response_parameters", aws.BoolValueMap(methodResponse.ResponseParameters)); err != nil {
		return fmt.Errorf("error setting response_parameters: %s", err)
	}

	return nil
}

func resourceMethodResponseUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).APIGatewayConn

	log.Printf("[DEBUG] Updating API Gateway Method Response %s", d.Id())
	operations := make([]*apigateway.PatchOperation, 0)

	if d.HasChange("response_models") {
		operations = append(operations, expandRequestResponseModelOperations(d, "response_models", "responseModels")...)
	}

	if d.HasChange("response_parameters") {
		ops := expandMethodParametersOperations(d, "response_parameters", "responseParameters")
		operations = append(operations, ops...)
	}

	out, err := conn.UpdateMethodResponse(&apigateway.UpdateMethodResponseInput{
		HttpMethod:      aws.String(d.Get("http_method").(string)),
		ResourceId:      aws.String(d.Get("resource_id").(string)),
		RestApiId:       aws.String(d.Get("rest_api_id").(string)),
		StatusCode:      aws.String(d.Get("status_code").(string)),
		PatchOperations: operations,
	})

	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Received API Gateway Method Response: %s", out)

	return resourceMethodResponseRead(d, meta)
}

func resourceMethodResponseDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).APIGatewayConn
	log.Printf("[DEBUG] Deleting API Gateway Method Response: %s", d.Id())

	_, err := conn.DeleteMethodResponse(&apigateway.DeleteMethodResponseInput{
		HttpMethod: aws.String(d.Get("http_method").(string)),
		ResourceId: aws.String(d.Get("resource_id").(string)),
		RestApiId:  aws.String(d.Get("rest_api_id").(string)),
		StatusCode: aws.String(d.Get("status_code").(string)),
	})

	if tfawserr.ErrCodeEquals(err, apigateway.ErrCodeNotFoundException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting API Gateway Method Response (%s): %s", d.Id(), err)
	}

	return nil
}
