package grafana

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/managedgrafana"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func ResourceWorkspaceApiKey() *schema.Resource {
	return &schema.Resource{
		Create: resourceWorkspaceApiKeyCreate,
		Read:   resourceWorkspaceApiKeyRead,
		Delete: resourceWorkspaceApiKeyDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"key_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"key_role": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(managedgrafana.Role_Values(), false),
				ForceNew:     true,
			},
			"seconds_to_live": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},
			"workspace_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceWorkspaceApiKeyCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).GrafanaConn

	input := &managedgrafana.CreateWorkspaceApiKeyInput{
		KeyName:       aws.String(d.Get("key_name").(string)),
		KeyRole:       aws.String(d.Get("key_role").(string)),
		SecondsToLive: aws.Int64(int64(d.Get("seconds_to_live").(int))),
		WorkspaceId:   aws.String(d.Get("workspace_id").(string)),
	}

	log.Printf("[DEBUG] Creating Grafana Workspace Api Key: %s", input)
	output, err := conn.CreateWorkspaceApiKey(input)

	if err != nil {
		return fmt.Errorf("error creating Grafana Workspace Api Key: %w", err)
	}

	d.SetId(aws.StringValue(output.KeyName))

	return resourceWorkspaceApiKeyRead(d, meta)
}

func resourceWorkspaceApiKeyRead(d *schema.ResourceData, meta interface{}) error {
	d.Set("key_name", d.Get("key_name").(string))
	d.Set("key_role", d.Get("key_role").(string))
	d.Set("seconds_to_live", d.Get("seconds_to_live").(int))
	d.Set("workspace_id", d.Get("workspace_id").(string))

	return nil
}

func resourceWorkspaceApiKeyDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).GrafanaConn

	log.Printf("[DEBUG] Deleting Grafana Workspace Api Key: %s", d.Id())
	_, err := conn.DeleteWorkspaceApiKey(&managedgrafana.DeleteWorkspaceApiKeyInput{
		KeyName:     aws.String(d.Get("key_name").(string)),
		WorkspaceId: aws.String(d.Get("workspace_id").(string)),
	})

	if err != nil {
		return fmt.Errorf("error deleting Grafana Workspace Api Key (%s): %w", d.Id(), err)
	}

	return nil
}
