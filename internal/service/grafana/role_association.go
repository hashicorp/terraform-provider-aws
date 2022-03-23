package grafana

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/managedgrafana"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func ResourceRoleAssociation() *schema.Resource {
	return &schema.Resource{
		Create: resourceRoleAssociationUpsert,
		Read:   resourceRoleAssociationRead,
		Update: resourceRoleAssociationUpsert,
		Delete: resourceRoleAssociationDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"group_ids": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"role": {
				Type:     schema.TypeString,
				Required: true,
			},
			"user_ids": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"workspace_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceRoleAssociationUpsert(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).GrafanaConn

	role := d.Get("role").(string)
	workspaceID := d.Get("workspace_id").(string)

	updateInstructions := make([]*managedgrafana.UpdateInstruction, 0)
	if v, ok := d.GetOk("user_ids"); ok && v.(*schema.Set).Len() > 0 {
		typeSsoUser := managedgrafana.UserTypeSsoUser
		updateInstructions = populateUpdateInstructions(role, flex.ExpandStringSet(v.(*schema.Set)), managedgrafana.UpdateActionAdd, typeSsoUser, updateInstructions)
	}

	if v, ok := d.GetOk("group_ids"); ok && v.(*schema.Set).Len() > 0 {
		typeSsoUser := managedgrafana.UserTypeSsoGroup
		updateInstructions = populateUpdateInstructions(role, flex.ExpandStringSet(v.(*schema.Set)), managedgrafana.UpdateActionAdd, typeSsoUser, updateInstructions)
	}

	input := &managedgrafana.UpdatePermissionsInput{
		UpdateInstructionBatch: updateInstructions,
		WorkspaceId:            aws.String(workspaceID),
	}

	log.Printf("[DEBUG] Creating Grafana Workspace Role Association: %s", input)
	response, err := conn.UpdatePermissions(input)

	for _, updateError := range response.Errors {
		return fmt.Errorf("error creating Grafana Workspace Role Association: %s", aws.StringValue(updateError.Message))
	}

	if err != nil {
		return fmt.Errorf("error creating Grafana Workspace Role Association: %w", err)
	}

	if d.Id() == "" {
		d.SetId(fmt.Sprintf("%s/%s", workspaceID, role))
	}

	return resourceRoleAssociationRead(d, meta)
}

func populateUpdateInstructions(role string, list []*string, action string, typeSsoUser string, updateInstructions []*managedgrafana.UpdateInstruction) []*managedgrafana.UpdateInstruction {
	users := make([]*managedgrafana.User, len(list))
	for i := 0; i < len(users); i++ {
		users[i] = &managedgrafana.User{
			Id:   list[i],
			Type: aws.String(typeSsoUser),
		}
	}
	updateInstructions = append(updateInstructions, &managedgrafana.UpdateInstruction{
		Action: aws.String(action),
		Role:   aws.String(role),
		Users:  users,
	})

	return updateInstructions
}

func resourceRoleAssociationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).GrafanaConn

	roleAssociations, err := FindRoleAssociationsByRoleAndWorkspaceID(conn, d.Get("role").(string), d.Get("workspace_id").(string))

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Grafana Workspace Role Association %s not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading Grafana Workspace Role Association (%s): %w", d.Id(), err)
	}

	d.Set("group_ids", roleAssociations[managedgrafana.UserTypeSsoGroup])
	d.Set("user_ids", roleAssociations[managedgrafana.UserTypeSsoUser])

	return nil
}

func resourceRoleAssociationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).GrafanaConn

	updateInstructions := make([]*managedgrafana.UpdateInstruction, 0)
	if v, ok := d.GetOk("user_ids"); ok && v.(*schema.Set).Len() > 0 {
		typeSsoUser := managedgrafana.UserTypeSsoUser
		updateInstructions = populateUpdateInstructions(d.Get("role").(string), flex.ExpandStringSet(v.(*schema.Set)), managedgrafana.UpdateActionRevoke, typeSsoUser, updateInstructions)
	}

	if v, ok := d.GetOk("group_ids"); ok && v.(*schema.Set).Len() > 0 {
		typeSsoUser := managedgrafana.UserTypeSsoGroup
		updateInstructions = populateUpdateInstructions(d.Get("role").(string), flex.ExpandStringSet(v.(*schema.Set)), managedgrafana.UpdateActionRevoke, typeSsoUser, updateInstructions)
	}

	input := &managedgrafana.UpdatePermissionsInput{
		UpdateInstructionBatch: updateInstructions,
		WorkspaceId:            aws.String(d.Get("workspace_id").(string)),
	}

	log.Printf("[DEBUG] Deleting Grafana Workspace Role Association: %s", input)
	_, err := conn.UpdatePermissions(input)

	if err != nil {
		return fmt.Errorf("error deleting Grafana Workspace Role Association: %w", err)
	}

	return nil
}
