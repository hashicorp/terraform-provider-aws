package grafana

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/managedgrafana"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func ResourceWorkspaceApiKey() *schema.Resource {
	return &schema.Resource{
		Create: resourceWorkspaceApiKeyInsert,
		Read:   schema.Noop,
		Update: schema.Noop,
		Delete: resourceWorkspaceApiKeyDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"key_name": {
				Type:     schema.TypeString,
				Required: true,
				//Computed: true,
				ForceNew: true,
			},
			"key_role": {
				Type:     schema.TypeString,
				Required: true,
				//Computed:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(managedgrafana.Role_Values(), false),
			},
			"seconds_to_live": {
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validation.IntBetween(1, 2592000),
			},
			"workspace_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceWorkspaceApiKeyInsert(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).GrafanaConn

	d.SetId(d.Get("workspace_id").(string))

	_, err := FindWorkspaceByID(conn, d.Id())

	if err != nil {
		return fmt.Errorf("error reading Grafana Workspace (%s): %w", d.Id(), err)
	}

	input := &managedgrafana.CreateWorkspaceApiKeyInput{
		KeyName:       aws.String(d.Get("key_name").(string)),
		KeyRole:       aws.String(d.Get("key_role").(string)),
		SecondsToLive: aws.Int64(int64(d.Get("seconds_to_live").(int))),
		WorkspaceId:   aws.String(d.Id()),
	}

	_, err = conn.CreateWorkspaceApiKey(input)

	if err != nil {
		return fmt.Errorf("error creating Grafana Workspace API Configuration: %w", err)
	}

	return nil

}

func resourceWorkspaceApiKeyDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).GrafanaConn

	log.Printf("[DEBUG] Deleting Grafana Workspace API Key: %s", d.Id())
	input := &managedgrafana.DeleteWorkspaceApiKeyInput{
		KeyName:     aws.String(d.Get("key_name").(string)),
		WorkspaceId: aws.String(d.Id()),
	}
	_, err := conn.DeleteWorkspaceApiKey(input)

	if err != nil {
		return fmt.Errorf("error Deleting Grafana Workspace API Configuration: %w", err)
	}

	return nil
}
