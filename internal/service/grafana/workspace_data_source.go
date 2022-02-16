package grafana

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func DataSourceWorkspace() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceWorkspaceRead,

		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"account_access_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"authentication_providers": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"organization_role_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"permission_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"saml_configuration_status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"stack_set_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"data_sources": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"notification_destinations": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"organizational_units": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"role_arn": {
				Type:     schema.TypeString,
				Computed: true,
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

func dataSourceWorkspaceRead(d *schema.ResourceData, meta interface{}) error {
	d.SetId(d.Get("id").(string))
	id := d.Id()
	conn := meta.(*conns.AWSClient).GrafanaConn
	workspace, err := FindWorkspaceByID(conn, id)

	if tfresource.NotFound(err) && !d.IsNewResource() {
		log.Printf("[WARN] Grafana Workspace (%s) not found, removing from state", id)
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading Grafana Workspace (%s): %w", id, err)
	}
	d.Set("id", d.Id())
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
		Resource:  id,
	}.String()
	d.Set("arn", workspaceArn)

	return nil
}
