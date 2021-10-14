package aws

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

func ResourceDocumentationVersion() *schema.Resource {
	return &schema.Resource{
		Create: resourceDocumentationVersionCreate,
		Read:   resourceDocumentationVersionRead,
		Update: resourceDocumentationVersionUpdate,
		Delete: resourceDocumentationVersionDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"version": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"rest_api_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func resourceDocumentationVersionCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).APIGatewayConn

	restApiId := d.Get("rest_api_id").(string)

	params := &apigateway.CreateDocumentationVersionInput{
		DocumentationVersion: aws.String(d.Get("version").(string)),
		RestApiId:            aws.String(restApiId),
	}
	if v, ok := d.GetOk("description"); ok {
		params.Description = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating API Gateway Documentation Version: %s", params)

	version, err := conn.CreateDocumentationVersion(params)
	if err != nil {
		return fmt.Errorf("Error creating API Gateway Documentation Version: %s", err)
	}

	d.SetId(restApiId + "/" + *version.Version)

	return resourceDocumentationVersionRead(d, meta)
}

func resourceDocumentationVersionRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).APIGatewayConn
	log.Printf("[DEBUG] Reading API Gateway Documentation Version %s", d.Id())

	apiId, docVersion, err := decodeApiGatewayDocumentationVersionId(d.Id())
	if err != nil {
		return err
	}

	version, err := conn.GetDocumentationVersion(&apigateway.GetDocumentationVersionInput{
		DocumentationVersion: aws.String(docVersion),
		RestApiId:            aws.String(apiId),
	})
	if err != nil {
		if tfawserr.ErrMessageContains(err, apigateway.ErrCodeNotFoundException, "") {
			log.Printf("[WARN] API Gateway Documentation Version (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return err
	}

	d.Set("rest_api_id", apiId)
	d.Set("description", version.Description)
	d.Set("version", version.Version)

	return nil
}

func resourceDocumentationVersionUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).APIGatewayConn
	log.Printf("[DEBUG] Updating API Gateway Documentation Version %s", d.Id())

	_, err := conn.UpdateDocumentationVersion(&apigateway.UpdateDocumentationVersionInput{
		DocumentationVersion: aws.String(d.Get("version").(string)),
		RestApiId:            aws.String(d.Get("rest_api_id").(string)),
		PatchOperations: []*apigateway.PatchOperation{
			{
				Op:    aws.String(apigateway.OpReplace),
				Path:  aws.String("/description"),
				Value: aws.String(d.Get("description").(string)),
			},
		},
	})
	if err != nil {
		return err
	}
	log.Printf("[DEBUG] Updated API Gateway Documentation Version %s", d.Id())

	return resourceDocumentationVersionRead(d, meta)
}

func resourceDocumentationVersionDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).APIGatewayConn
	log.Printf("[DEBUG] Deleting API Gateway Documentation Version: %s", d.Id())

	_, err := conn.DeleteDocumentationVersion(&apigateway.DeleteDocumentationVersionInput{
		DocumentationVersion: aws.String(d.Get("version").(string)),
		RestApiId:            aws.String(d.Get("rest_api_id").(string)),
	})
	return err
}

func decodeApiGatewayDocumentationVersionId(id string) (string, string, error) {
	parts := strings.Split(id, "/")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("Expected ID in the form of REST-API-ID/VERSION, given: %q", id)
	}
	return parts[0], parts[1], nil
}
