package appsync

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appsync"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func ResourceType() *schema.Resource {
	return &schema.Resource{
		Create: resourceTypeCreate,
		Read:   resourceTypeRead,
		Update: resourceTypeUpdate,
		Delete: resourceTypeDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"api_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"definition": {
				Type:     schema.TypeString,
				Required: true,
			},
			"format": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(appsync.TypeDefinitionFormat_Values(), false),
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceTypeCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).AppSyncConn()

	apiID := d.Get("api_id").(string)

	params := &appsync.CreateTypeInput{
		ApiId:      aws.String(apiID),
		Definition: aws.String(d.Get("definition").(string)),
		Format:     aws.String(d.Get("format").(string)),
	}

	out, err := conn.CreateType(params)
	if err != nil {
		return fmt.Errorf("error creating Appsync Type: %w", err)
	}

	d.SetId(fmt.Sprintf("%s:%s:%s", apiID, aws.StringValue(out.Type.Format), aws.StringValue(out.Type.Name)))

	return resourceTypeRead(d, meta)
}

func resourceTypeRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).AppSyncConn()

	apiID, format, name, err := DecodeTypeID(d.Id())
	if err != nil {
		return err
	}

	resp, err := FindTypeByID(conn, apiID, format, name)
	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] AppSync Type (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error getting Appsync Type %q: %s", d.Id(), err)
	}

	d.Set("api_id", apiID)
	d.Set("arn", resp.Arn)
	d.Set("name", resp.Name)
	d.Set("format", resp.Format)
	d.Set("definition", resp.Definition)
	d.Set("description", resp.Description)

	return nil
}

func resourceTypeUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).AppSyncConn()

	params := &appsync.UpdateTypeInput{
		ApiId:      aws.String(d.Get("api_id").(string)),
		Format:     aws.String(d.Get("format").(string)),
		TypeName:   aws.String(d.Get("name").(string)),
		Definition: aws.String(d.Get("definition").(string)),
	}

	_, err := conn.UpdateType(params)
	if err != nil {
		return fmt.Errorf("error updating Appsync Type %q: %w", d.Id(), err)
	}

	return resourceTypeRead(d, meta)
}

func resourceTypeDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).AppSyncConn()

	input := &appsync.DeleteTypeInput{
		ApiId:    aws.String(d.Get("api_id").(string)),
		TypeName: aws.String(d.Get("name").(string)),
	}
	_, err := conn.DeleteType(input)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, appsync.ErrCodeNotFoundException) {
			return nil
		}
		return fmt.Errorf("error deleting Appsync Type: %w", err)
	}

	return nil
}

func DecodeTypeID(id string) (string, string, string, error) {
	parts := strings.Split(id, ":")
	if len(parts) != 3 {
		return "", "", "", fmt.Errorf("Unexpected format of ID (%q), expected API-ID:FORMAT:TYPE-NAME", id)
	}
	return parts[0], parts[1], parts[2], nil
}
