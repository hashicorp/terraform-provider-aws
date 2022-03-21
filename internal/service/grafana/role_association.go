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
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"role": {
				Type:     schema.TypeString,
				Required: true,
			},
			"user_ids": {
				Type:     schema.TypeList,
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

	updateInstructions := make([]*managedgrafana.UpdateInstruction, 0)
	var typeSsoUser string
	if v, ok := d.GetOk("user_ids"); ok {
		typeSsoUser = managedgrafana.UserTypeSsoUser
		updateInstructions = populateUpdateInstructions(d, v, managedgrafana.UpdateActionAdd, typeSsoUser, updateInstructions)
	}

	if v, ok := d.GetOk("group_ids"); ok {
		typeSsoUser = managedgrafana.UserTypeSsoGroup
		updateInstructions = populateUpdateInstructions(d, v, managedgrafana.UpdateActionAdd, typeSsoUser, updateInstructions)
	}

	input := &managedgrafana.UpdatePermissionsInput{
		UpdateInstructionBatch: updateInstructions,
		WorkspaceId:            aws.String(d.Get("workspace_id").(string)),
	}

	log.Printf("[DEBUG] Creating Grafana Workspace Role Association: %s", input)
	response, err := conn.UpdatePermissions(input)

	for _, updateError := range response.Errors {
		return fmt.Errorf("error creating Grafana Workspace Role Association: %s", aws.StringValue(updateError.Message))
	}

	if err != nil {
		return fmt.Errorf("error creating Grafana Workspace Role Association: %w", err)
	}

	return resourceRoleAssociationRead(d, meta)
}

func populateUpdateInstructions(d *schema.ResourceData, v interface{}, action string, typeSsoUser string, updateInstructions []*managedgrafana.UpdateInstruction) []*managedgrafana.UpdateInstruction {
	list := flex.ExpandStringList(v.([]interface{}))
	users := make([]*managedgrafana.User, len(list))
	for i := 0; i < len(users); i++ {
		users[i] = &managedgrafana.User{
			Id:   list[i],
			Type: aws.String(typeSsoUser),
		}
	}
	updateInstructions = append(updateInstructions, &managedgrafana.UpdateInstruction{
		Action: aws.String(action),
		Role:   aws.String(d.Get("role").(string)),
		Users:  users,
	})

	return updateInstructions
}

func resourceRoleAssociationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).GrafanaConn

	roleAssociation, err := FindRoleAssociationByRoleAndWorkspaceID(conn, d.Get("role").(string), d.Get("workspace_id").(string))

	if err != nil {
		return fmt.Errorf("error reading Grafana Workspace (%s-%s): %w", d.Get("workspace_id").(string), d.Get("role").(string), err)
	}

	users := roleAssociation[managedgrafana.UserTypeSsoUser]
	groups := roleAssociation[managedgrafana.UserTypeSsoGroup]

	usersLength := len(users)
	groupsLength := len(groups)
	if usersLength == 0 && groupsLength == 0 {
		return fmt.Errorf("role association not found %s-%s", d.Get("workspace_id").(string), d.Get("role").(string))
	}

	if usersLength > 0 {
		userIds := make([]*string, usersLength)
		for i := 0; i < len(userIds); i++ {
			userIds[i] = users[i].Id
		}
		d.Set("user_ids", userIds)
	}

	if groupsLength > 0 {
		groupIds := make([]*string, groupsLength)
		for i := 0; i < len(groupIds); i++ {
			groupIds[i] = groups[i].Id
		}
		d.Set("group_ids", groupIds)
	}

	return nil
}

func resourceRoleAssociationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).GrafanaConn

	updateInstructions := make([]*managedgrafana.UpdateInstruction, 0)
	var typeSsoUser string
	if v, ok := d.GetOk("user_ids"); ok {
		typeSsoUser = managedgrafana.UserTypeSsoUser
		updateInstructions = populateUpdateInstructions(d, v, managedgrafana.UpdateActionRevoke, typeSsoUser, updateInstructions)
	}

	if v, ok := d.GetOk("group_ids"); ok {
		typeSsoUser = managedgrafana.UserTypeSsoGroup
		updateInstructions = populateUpdateInstructions(d, v, managedgrafana.UpdateActionRevoke, typeSsoUser, updateInstructions)
	}

	input := &managedgrafana.UpdatePermissionsInput{
		UpdateInstructionBatch: updateInstructions,
		WorkspaceId:            aws.String(d.Get("workspace_id").(string)),
	}

	log.Printf("[DEBUG] Creating Grafana Workspace Role Association: %s", input)
	_, err := conn.UpdatePermissions(input)

	if err != nil {
		return fmt.Errorf("error creating Grafana Workspace Role Association: %w", err)
	}

	return nil
}
