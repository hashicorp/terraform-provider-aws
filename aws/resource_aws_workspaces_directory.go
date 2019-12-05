package aws

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/workspaces"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceAwsWorkspacesDirectory() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsWorkspacesDirectoryCreate,
		Read:   resourceAwsWorkspacesDirectoryRead,
		Update: resourceAwsWorkspacesDirectoryUpdate,
		Delete: resourceAwsWorkspacesDirectoryDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"directory_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"subnet_ids": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func resourceAwsWorkspacesDirectoryCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).workspacesconn
	directoryId := d.Get("directory_id").(string)

	input := &workspaces.RegisterWorkspaceDirectoryInput{
		DirectoryId:       aws.String(directoryId),
		EnableSelfService: aws.Bool(false),
		EnableWorkDocs:    aws.Bool(false),
		Tenancy:           aws.String(workspaces.TenancyShared),
	}

	if v, ok := d.GetOk("subnet_ids"); ok {
		subnetIdsSet := v.(*schema.Set)
		for _, id := range subnetIdsSet.List() {
			input.SubnetIds = append(input.SubnetIds, aws.String(id.(string)))
		}
	}

	_, err := conn.RegisterWorkspaceDirectory(input)
	if err != nil {
		return err
	}

	stateConf := &resource.StateChangeConf{
		Pending: []string{
			workspaces.WorkspaceDirectoryStateRegistering,
		},
		Target:  []string{workspaces.WorkspaceDirectoryStateRegistered},
		Refresh: workspacesDirectoryRefreshStateFunc(conn, directoryId),
		Timeout: 10 * time.Minute,
	}

	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("workspace directory was not registered: %s", err)
	}

	d.SetId(directoryId)

	return resourceAwsWorkspacesDirectoryRead(d, meta)
}

func resourceAwsWorkspacesDirectoryRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).workspacesconn

	resp, err := conn.DescribeWorkspaceDirectories(&workspaces.DescribeWorkspaceDirectoriesInput{
		DirectoryIds: []*string{aws.String(d.Id())},
	})
	if err != nil {
		return err
	}
	dir := resp.Directories[0]

	d.Set("directory_id", dir.DirectoryId)
	d.Set("subnet_ids", dir.SubnetIds)

	return nil
}

func resourceAwsWorkspacesDirectoryUpdate(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func resourceAwsWorkspacesDirectoryDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).workspacesconn

	_, err := conn.DeregisterWorkspaceDirectory(&workspaces.DeregisterWorkspaceDirectoryInput{
		DirectoryId: aws.String(d.Id()),
	})
	if err != nil {
		return err
	}

	stateConf := &resource.StateChangeConf{
		Pending: []string{
			workspaces.WorkspaceDirectoryStateRegistering,
			workspaces.WorkspaceDirectoryStateRegistered,
		},
		Target: []string{
			workspaces.WorkspaceDirectoryStateDeregistering,
			workspaces.WorkspaceDirectoryStateDeregistered,
		},
		Refresh: workspacesDirectoryRefreshStateFunc(conn, d.Id()),
		Timeout: 10 * time.Minute,
	}

	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("directory was not deregistered: %s", err)
	}

	return nil
}

func workspacesDirectoryRefreshStateFunc(conn *workspaces.WorkSpaces, directoryID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		resp, err := conn.DescribeWorkspaceDirectories(&workspaces.DescribeWorkspaceDirectoriesInput{
			DirectoryIds: []*string{aws.String(directoryID)},
		})
		if err != nil {
			return nil, workspaces.WorkspaceDirectoryStateError, err
		}
		if len(resp.Directories) == 0 {
			return resp, workspaces.WorkspaceDirectoryStateDeregistered, nil
		}
		return resp, *resp.Directories[0].State, nil
	}
}
