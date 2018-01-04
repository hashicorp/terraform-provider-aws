package aws

import (
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsApiGatewayDocumentationPart() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsApiGatewayDocumentationPartCreate,
		Read:   resourceAwsApiGatewayDocumentationPartRead,
		Update: resourceAwsApiGatewayDocumentationPartUpdate,
		Delete: resourceAwsApiGatewayDocumentationPartDelete,

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

func resourceAwsApiGatewayDocumentationPartCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigateway

	out, err := conn.CreateDocumentationPart(&apigateway.CreateDocumentationPartInput{
		Location:   expandApiGatewayDocumentationPartLocation(d.Get("location").([]interface{})),
		Properties: aws.String(d.Get("properties").(string)),
		RestApiId:  aws.String(d.Get("rest_api_id").(string)),
	})
	if err != nil {
		return err
	}
	d.SetId(*out.Id)

	return nil
}

func resourceAwsApiGatewayDocumentationPartRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigateway

	log.Printf("[INFO] Reading API Gateway Documentation Part %s", d.Id())
	docPart, err := conn.GetDocumentationPart(&apigateway.GetDocumentationPartInput{
		DocumentationPartId: aws.String(d.Id()),
		RestApiId:           aws.String(d.Get("rest_api_id").(string)),
	})
	if err != nil {
		if isAWSErr(err, "NotFoundException", "") {
			log.Printf("[WARN] API Gateway Documentation Part (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return err
	}

	log.Printf("[DEBUG] Received API Gateway Documentation Part: %s", docPart)

	d.Set("location", flattenApiGatewayDocumentationPartLocation(docPart.Location))
	d.Set("properties", docPart.Properties)

	return nil
}

func resourceAwsApiGatewayDocumentationPartUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigateway

	input := apigateway.UpdateDocumentationPartInput{
		DocumentationPartId: aws.String(d.Id()),
		RestApiId:           aws.String(d.Get("rest_api_id").(string)),
	}
	operations := make([]*apigateway.PatchOperation, 0)

	if d.HasChange("properties") {
		properties := d.Get("properties").(string)
		operations = append(operations, &apigateway.PatchOperation{
			Op:    aws.String("replace"),
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

	return resourceAwsApiGatewayDocumentationPartRead(d, meta)
}

func resourceAwsApiGatewayDocumentationPartDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigateway

	_, err := conn.DeleteDocumentationPart(&apigateway.DeleteDocumentationPartInput{
		DocumentationPartId: aws.String(d.Id()),
		RestApiId:           aws.String(d.Get("rest_api_id").(string)),
	})
	if err != nil {
		return err
	}

	return nil
}

func expandApiGatewayDocumentationPartLocation(l []interface{}) *apigateway.DocumentationPartLocation {
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

func flattenApiGatewayDocumentationPartLocation(l *apigateway.DocumentationPartLocation) []interface{} {
	if l == nil {
		return []interface{}{}
	}

	m := make(map[string]interface{}, 0)
	m["type"] = *l.Type
	if l.Method != nil {
		m["method"] = *l.Method
	}
	if l.Name != nil {
		m["name"] = *l.Name
	}
	if l.Path != nil {
		m["path"] = *l.Path
	}
	if l.StatusCode != nil {
		m["status_code"] = *l.StatusCode
	}

	return []interface{}{m}
}
