package apigateway

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func ResourceDocumentationPart() *schema.Resource {
	return &schema.Resource{
		Create: resourceDocumentationPartCreate,
		Read:   resourceDocumentationPartRead,
		Update: resourceDocumentationPartUpdate,
		Delete: resourceDocumentationPartDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"location": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"method": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"name": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"path": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"status_code": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"type": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
					},
				},
			},
			"properties": {
				Type:     schema.TypeString,
				Required: true,
			},
			"rest_api_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceDocumentationPartCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).APIGatewayConn

	apiId := d.Get("rest_api_id").(string)
	out, err := conn.CreateDocumentationPart(&apigateway.CreateDocumentationPartInput{
		Location:   expandDocumentationPartLocation(d.Get("location").([]interface{})),
		Properties: aws.String(d.Get("properties").(string)),
		RestApiId:  aws.String(apiId),
	})
	if err != nil {
		return err
	}
	d.SetId(apiId + "/" + *out.Id)

	return nil
}

func resourceDocumentationPartRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).APIGatewayConn

	log.Printf("[INFO] Reading API Gateway Documentation Part %s", d.Id())

	apiId, id, err := DecodeDocumentationPartID(d.Id())
	if err != nil {
		return err
	}

	docPart, err := conn.GetDocumentationPart(&apigateway.GetDocumentationPartInput{
		DocumentationPartId: aws.String(id),
		RestApiId:           aws.String(apiId),
	})
	if err != nil {
		if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, apigateway.ErrCodeNotFoundException) {
			log.Printf("[WARN] API Gateway Documentation Part (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error reading API Gateway Documentation Part (%s): %w", d.Id(), err)
	}

	d.Set("rest_api_id", apiId)
	d.Set("location", flattenDocumentationPartLocation(docPart.Location))
	d.Set("properties", docPart.Properties)

	return nil
}

func resourceDocumentationPartUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).APIGatewayConn

	apiId, id, err := DecodeDocumentationPartID(d.Id())
	if err != nil {
		return err
	}

	input := apigateway.UpdateDocumentationPartInput{
		DocumentationPartId: aws.String(id),
		RestApiId:           aws.String(apiId),
	}
	operations := make([]*apigateway.PatchOperation, 0)

	if d.HasChange("properties") {
		properties := d.Get("properties").(string)
		operations = append(operations, &apigateway.PatchOperation{
			Op:    aws.String(apigateway.OpReplace),
			Path:  aws.String("/properties"),
			Value: aws.String(properties),
		})
	}

	input.PatchOperations = operations

	log.Printf("[INFO] Updating API Gateway Documentation Part: %s", input)

	out, err := conn.UpdateDocumentationPart(&input)
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] API Gateway Documentation Part updated: %s", out)

	return resourceDocumentationPartRead(d, meta)
}

func resourceDocumentationPartDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).APIGatewayConn

	apiId, id, err := DecodeDocumentationPartID(d.Id())
	if err != nil {
		return err
	}

	_, err = conn.DeleteDocumentationPart(&apigateway.DeleteDocumentationPartInput{
		DocumentationPartId: aws.String(id),
		RestApiId:           aws.String(apiId),
	})
	return err
}

func expandDocumentationPartLocation(l []interface{}) *apigateway.DocumentationPartLocation {
	if len(l) == 0 {
		return nil
	}
	loc := l[0].(map[string]interface{})
	out := &apigateway.DocumentationPartLocation{
		Type: aws.String(loc["type"].(string)),
	}
	if v, ok := loc["method"]; ok {
		out.Method = aws.String(v.(string))
	}
	if v, ok := loc["name"]; ok {
		out.Name = aws.String(v.(string))
	}
	if v, ok := loc["path"]; ok {
		out.Path = aws.String(v.(string))
	}
	if v, ok := loc["status_code"]; ok {
		out.StatusCode = aws.String(v.(string))
	}
	return out
}

func flattenDocumentationPartLocation(l *apigateway.DocumentationPartLocation) []interface{} {
	if l == nil {
		return []interface{}{}
	}

	m := make(map[string]interface{})

	if v := l.Method; v != nil {
		m["method"] = aws.StringValue(v)
	}

	if v := l.Name; v != nil {
		m["name"] = aws.StringValue(v)
	}

	if v := l.Path; v != nil {
		m["path"] = aws.StringValue(v)
	}

	if v := l.StatusCode; v != nil {
		m["status_code"] = aws.StringValue(v)
	}

	if v := l.Type; v != nil {
		m["type"] = aws.StringValue(v)
	}

	return []interface{}{m}
}

func DecodeDocumentationPartID(id string) (string, string, error) {
	parts := strings.Split(id, "/")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("Unexpected format of ID (%q), expected REST-API-ID/ID", id)
	}
	return parts[0], parts[1], nil
}
