package grafana

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/managedgrafana"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func ResourceWorkspaceAPIKey() *schema.Resource {
	return &schema.Resource{
		Create: resourceWorkspaceAPIKeyCreate,
		Read:   schema.Noop,
		Update: schema.Noop,
		Delete: resourceWorkspaceAPIKeyDelete,

		Schema: map[string]*schema.Schema{
			"key": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"key_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 100),
			},
			"key_role": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(managedgrafana.Role_Values(), false),
			},
			"seconds_to_live": {
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validation.IntBetween(1, 2592000),
			},
			"workspace_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^g-[0-9a-f]{10}$`), "must be a valid workspace id"),
			},
		},
	}
}

func resourceWorkspaceAPIKeyCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).GrafanaConn
	wrkspc_id := d.Get("workspace_id").(string)
	_, err := FindWorkspaceByID(conn, wrkspc_id)

	if err != nil {
		return fmt.Errorf("error reading Grafana Workspace (%s): %w", d.Id(), err)
	}

	input := &managedgrafana.CreateWorkspaceApiKeyInput{
		KeyName:       aws.String(d.Get("key_name").(string)),
		KeyRole:       aws.String(d.Get("key_role").(string)),
		SecondsToLive: aws.Int64(int64(d.Get("seconds_to_live").(int))),
		WorkspaceId:   aws.String(wrkspc_id),
	}

	log.Printf("[DEBUG] Creating Grafana API Key: %s", input)
	output, err := conn.CreateWorkspaceApiKey(input)

	if err != nil {
		return fmt.Errorf("error creating Grafana API Key : %w", err)
	}

	d.SetId(fmt.Sprintf("%s/%s", wrkspc_id, d.Get("key_name").(string)))
	d.Set("key", output.Key)
	return nil

}

func resourceWorkspaceAPIKeyDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).GrafanaConn

	log.Printf("[DEBUG] Deleting Grafana Workspace API Key: %s", d.Id())

	wrkspacID, keyName, error := WorkspaceKeyParseID(d.Id())
	if error != nil {
		return error
	}

	input := &managedgrafana.DeleteWorkspaceApiKeyInput{
		KeyName:     aws.String(keyName),
		WorkspaceId: aws.String(wrkspacID),
	}
	_, err := conn.DeleteWorkspaceApiKey(input)

	if err != nil {
		if tfawserr.ErrCodeEquals(err, managedgrafana.ErrCodeResourceNotFoundException) {
			return nil
		}
		return err
	}

	return nil

}

func WorkspaceKeyParseID(id string) (string, string, error) {
	parts := strings.SplitN(id, "/", 2)

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format of ID (%s), expected wrkspcID:keyName", id)
	}

	return parts[0], parts[1], nil
}
