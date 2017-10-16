package aws

import "github.com/hashicorp/terraform/helper/schema"

func resourceAwsWorkspacesWorkspace() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsWorkspacesWorkspaceCreate,
		Read:   resourceAwsWorkspacesWorkspaceRead,
		Update: resourceAwsWorkspacesWorkspaceUpdate,
		Delete: resourceAwsWorkspacesWorkspaceDelete,

		Schema: map[string]*schema.Schema{},
	}
}

func resourceAwsWorkspacesWorkspaceCreate(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func resourceAwsWorkspacesWorkspaceRead(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func resourceAwsWorkspacesWorkspaceUpdate(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func resourceAwsWorkspacesWorkspaceDelete(d *schema.ResourceData, meta interface{}) error {
	return nil
}
