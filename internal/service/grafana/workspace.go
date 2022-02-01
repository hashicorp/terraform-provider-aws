package grafana

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/managedgrafana"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"log"
	"time"
)

func ResourceWorkspace() *schema.Resource {
	return &schema.Resource{
		Create: resourceWorkspaceCreate,
		Read:   resourceWorkspaceRead,
		Update: resourceWorkspaceUpdate,
		Delete: resourceWorkspaceDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"account_access_type": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(managedgrafana.AccountAccessType_Values(), false),
			},
			"authentication_providers": {
				Type:         schema.TypeList,
				Required:     true,
				ValidateFunc: validation.StringInSlice(managedgrafana.AuthenticationProviderTypes_Values(), false),
			},
			"saml_configuration_status": {
				Type:         schema.TypeString,
				Required:     false,
				ValidateFunc: validation.StringInSlice(managedgrafana.SamlConfigurationStatus_Values(), false),
			},
			"organization_role_name": {
				Type:     schema.TypeString,
				Required: false,
			},
			"permission_type": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(managedgrafana.PermissionType_Values(), false),
			},
			"stack_set_name": {
				Type:     schema.TypeString,
				Required: false,
			},
			"data_sources": {
				Type:     schema.TypeList,
				Required: false,
			},
			"description": {
				Type:     schema.TypeString,
				Required: false,
			},
			"name": {
				Type:     schema.TypeString,
				Required: false,
			},
			"notification_destinations": {
				Type:     schema.TypeList,
				Required: false,
			},
			"organizational_units": {
				Type:     schema.TypeString,
				Required: false,
			},
			"role_arn": {
				Type:     schema.TypeString,
				Required: false,
			},
		},
	}
}

func resourceWorkspaceCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).GrafanaConn
	input := &managedgrafana.CreateWorkspaceInput{
		AccountAccessType:       aws.String(d.Get("account_access_type").(string)),
		AuthenticationProviders: aws.StringSlice(d.Get("authentication_providers").([]string)),
	}

	if v, ok := d.GetOk("organization_role_name"); ok {
		input.OrganizationRoleName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("permission_type"); ok {
		input.PermissionType = aws.String(v.(string))
	}

	if v, ok := d.GetOk("stack_set_name"); ok {
		input.StackSetName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("data_sources"); ok {
		input.WorkspaceDataSources = aws.StringSlice(v.([]string))
	}

	if v, ok := d.GetOk("description"); ok {
		input.WorkspaceDescription = aws.String(v.(string))
	}

	if v, ok := d.GetOk("name"); ok {
		input.WorkspaceName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("notification_destinations"); ok {
		input.WorkspaceNotificationDestinations = aws.StringSlice(v.([]string))
	}

	if v, ok := d.GetOk("organizational_units"); ok {
		input.WorkspaceOrganizationalUnits = aws.StringSlice(v.([]string))
	}

	if v, ok := d.GetOk("role_arn"); ok {
		input.WorkspaceRoleArn = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating Grafana Workspace: %s", input)
	output, err := conn.CreateWorkspace(input)

	if err != nil {
		return fmt.Errorf("error creating Grafana Workspace: %w", err)
	}

	d.SetId(aws.StringValue(output.Workspace.Id))

	_, err = waitWorkspaceCreated(conn, d.Id(), d.Timeout(schema.TimeoutCreate))

	if err != nil {
		return fmt.Errorf("error waiting for Grafana Workspace (%s) create: %w", d.Id(), err)
	}
	return resourceWorkspaceRead(d, meta)
}

func resourceWorkspaceRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).GrafanaConn

	workspace, err := FindWorkspaceById(conn, d.Id())

	if tfresource.NotFound(err) && !d.IsNewResource() {
		log.Printf("[WARN] Grafana Workspace (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading Grafana Workspace (%s): %w", d.Id(), err)
	}

	d.Set("account_access_type", workspace.AccountAccessType)
	d.Set("authentication_providers", workspace.Authentication.Providers)
	d.Set("saml_configuration_status", workspace.Authentication.SamlConfigurationStatus)
	d.Set("organization_role_name", workspace.OrganizationRoleName)
	d.Set("permission_type", workspace.PermissionType)
	d.Set("stack_set_name", workspace.StackSetName)
	d.Set("data_sources", workspace.DataSources)
	d.Set("description", workspace.Description)
	d.Set("name", workspace.Name)
	d.Set("notification_destinations", workspace.NotificationDestinations)
	d.Set("organizational_units", workspace.OrganizationalUnits)
	d.Set("role_arn", workspace.WorkspaceRoleArn)

	return nil
}

func resourceWorkspaceUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).GrafanaConn

	if d.HasChange("authentication_providers") {
		updateWorkspaceAuthenticationInput := &managedgrafana.UpdateWorkspaceAuthenticationInput{
			WorkspaceId:             aws.String(d.Id()),
			AuthenticationProviders: aws.StringSlice(d.Get("authentication_providers").([]string)),
		}
		_, err := conn.UpdateWorkspaceAuthentication(updateWorkspaceAuthenticationInput)

		if err != nil {
			return fmt.Errorf("error updating Grafana Workspace (%s) authentication provider: %w", d.Id(), err)
		}

		_, err = waitWorkspaceUpdated(conn, d.Id(), d.Timeout(schema.TimeoutUpdate))
	}

	input := &managedgrafana.UpdateWorkspaceInput{
		WorkspaceId: aws.String(d.Id()),
	}

	if d.HasChange("account_access_type") {
		input.AccountAccessType = aws.String(d.Get("account_access_type").(string))
	}

	if d.HasChange("organization_role_name") {
		input.OrganizationRoleName = aws.String(d.Get("organization_role_name").(string))
	}

	if d.HasChange("permission_type") {
		input.PermissionType = aws.String(d.Get("permission_type").(string))
	}

	if d.HasChange("stack_set_name") {
		input.StackSetName = aws.String(d.Get("stack_set_name").(string))
	}

	if d.HasChange("data_sources") {
		input.WorkspaceDataSources = aws.StringSlice(d.Get("data_sources").([]string))
	}

	if d.HasChange("description") {
		input.WorkspaceDescription = aws.String(d.Get("description").(string))
	}

	if d.HasChange("name") {
		input.WorkspaceName = aws.String(d.Get("name").(string))
	}

	if d.HasChange("notification_destinations") {
		input.WorkspaceNotificationDestinations = aws.StringSlice(d.Get("notification_destinations").([]string))
	}

	if d.HasChange("organizational_units") {
		input.WorkspaceOrganizationalUnits = aws.StringSlice(d.Get("organizational_units").([]string))
	}

	if d.HasChange("role_arn") {
		input.WorkspaceRoleArn = aws.String(d.Get("role_arn").(string))
	}

	_, err := conn.UpdateWorkspace(input)

	if err != nil {
		return fmt.Errorf("error updating Grafana Workspace (%s): %w", d.Id(), err)
	}

	return resourceWorkspaceRead(d, meta)
}

func resourceWorkspaceDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).GrafanaConn

	id := d.Get("id").(string)
	input := &managedgrafana.DeleteWorkspaceInput{
		WorkspaceId: aws.String(id),
	}

	_, err := conn.DeleteWorkspace(input)
	if tfawserr.ErrCodeEquals(err, managedgrafana.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting Grafana Workspace (%s): %w", d.Id(), err)
	}

	_, err = waitWorkspaceDeleted(conn, d.Id(), d.Timeout(schema.TimeoutDelete))

	return nil
}
