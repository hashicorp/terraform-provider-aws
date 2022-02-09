package grafana

import (
	"fmt"
	"log"
	"time"
)

import (
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/managedgrafana"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
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
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"account_access_type": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(managedgrafana.AccountAccessType_Values(), false),
			},
			"authentication_providers": {
				Type:     schema.TypeList,
				Required: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"organization_role_name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"permission_type": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(managedgrafana.PermissionType_Values(), false),
			},
			"saml_configuration_status": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(managedgrafana.SamlConfigurationStatus_Values(), false),
			},
			"stack_set_name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"data_sources": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"notification_destinations": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"organizational_units": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"status": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(managedgrafana.WorkspaceStatus_Values(), false),
			},
			"role_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: verify.ValidARN,
			},
			"created_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"last_updated_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"endpoint": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"grafana_version": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceWorkspaceCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).GrafanaConn
	input := &managedgrafana.CreateWorkspaceInput{
		ClientToken:             aws.String(resource.UniqueId()),
		AccountAccessType:       aws.String(d.Get("account_access_type").(string)),
		AuthenticationProviders: flex.ExpandStringList(d.Get("authentication_providers").([]interface{})),
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
		input.WorkspaceDataSources = flex.ExpandStringList(v.([]interface{}))
	}

	if v, ok := d.GetOk("description"); ok {
		input.WorkspaceDescription = aws.String(v.(string))
	}

	if v, ok := d.GetOk("name"); ok {
		input.WorkspaceName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("notification_destinations"); ok {
		input.WorkspaceNotificationDestinations = flex.ExpandStringList(v.([]interface{}))
	}

	if v, ok := d.GetOk("organizational_units"); ok {
		input.WorkspaceOrganizationalUnits = flex.ExpandStringList(v.([]interface{}))
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
	d.Set("status", workspace.Status)
	d.Set("role_arn", workspace.WorkspaceRoleArn)
	d.Set("created_date", workspace.Created.Format(time.RFC3339))
	d.Set("last_updated_date", workspace.Modified.Format(time.RFC3339))
	d.Set("endpoint", workspace.Endpoint)
	d.Set("grafana_version", workspace.GrafanaVersion)
	workspaceArn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "grafana",
		Resource:  d.Id(),
	}.String()
	d.Set("arn", workspaceArn)

	return nil
}

func resourceWorkspaceUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).GrafanaConn

	if d.HasChange("authentication_providers") {
		updateWorkspaceAuthenticationInput := &managedgrafana.UpdateWorkspaceAuthenticationInput{
			WorkspaceId:             aws.String(d.Id()),
			AuthenticationProviders: flex.ExpandStringList(d.Get("authentication_providers").([]interface{})),
		}
		_, err := conn.UpdateWorkspaceAuthentication(updateWorkspaceAuthenticationInput)

		if err != nil {
			return fmt.Errorf("error updating Grafana Workspace (%s) authentication provider: %w", d.Id(), err)
		}

		_, err = waitWorkspaceUpdated(conn, d.Id(), d.Timeout(schema.TimeoutUpdate))

		if err != nil {
			return fmt.Errorf("error updating Grafana Workspace (%s) authentication provider: %w", d.Id(), err)
		}
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
		input.WorkspaceDataSources = flex.ExpandStringList(d.Get("data_sources").([]interface{}))
	}

	if d.HasChange("description") {
		input.WorkspaceDescription = aws.String(d.Get("description").(string))
	}

	if d.HasChange("name") {
		input.WorkspaceName = aws.String(d.Get("name").(string))
	}

	if d.HasChange("notification_destinations") {
		input.WorkspaceNotificationDestinations = flex.ExpandStringList(d.Get("notification_destinations").([]interface{}))
	}

	if d.HasChange("organizational_units") {
		input.WorkspaceOrganizationalUnits = flex.ExpandStringList(d.Get("organizational_units").([]interface{}))
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

	if err != nil {
		return fmt.Errorf("error deleting Grafana Workspace (%s): %w", d.Id(), err)
	}

	return nil
}
