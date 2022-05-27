package grafana

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/managedgrafana"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

func DataSourceWorkspace() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceWorkspaceRead,

		Schema: map[string]*schema.Schema{
			"account_access_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"authentication_providers": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"created_date": {
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
			"endpoint": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"grafana_version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"last_updated_date": {
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
			"organization_role_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"organizational_units": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"permission_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"role_arn": {
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
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tftags.TagsSchemaComputed(),
			"workspace_id": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func dataSourceWorkspaceRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).GrafanaConn
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	workspaceID := d.Get("workspace_id").(string)
	workspace, err := FindWorkspaceByID(conn, workspaceID)

	if err != nil {
		return fmt.Errorf("error reading Grafana Workspace (%s): %w", workspaceID, err)
	}

	d.SetId(workspaceID)
	d.Set("account_access_type", workspace.AccountAccessType)
	// https://docs.aws.amazon.com/service-authorization/latest/reference/list_amazonmanagedgrafana.html#amazonmanagedgrafana-resources-for-iam-policies.
	workspaceARN := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   managedgrafana.ServiceName,
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("/workspaces/%s", d.Id()),
	}.String()
	d.Set("arn", workspaceARN)
	d.Set("authentication_providers", workspace.Authentication.Providers)
	d.Set("created_date", workspace.Created.Format(time.RFC3339))
	d.Set("data_sources", workspace.DataSources)
	d.Set("description", workspace.Description)
	d.Set("endpoint", workspace.Endpoint)
	d.Set("grafana_version", workspace.GrafanaVersion)
	d.Set("last_updated_date", workspace.Modified.Format(time.RFC3339))
	d.Set("name", workspace.Name)
	d.Set("notification_destinations", workspace.NotificationDestinations)
	d.Set("organization_role_name", workspace.OrganizationRoleName)
	d.Set("organizational_units", workspace.OrganizationalUnits)
	d.Set("permission_type", workspace.PermissionType)
	d.Set("role_arn", workspace.WorkspaceRoleArn)
	d.Set("saml_configuration_status", workspace.Authentication.SamlConfigurationStatus)
	d.Set("stack_set_name", workspace.StackSetName)
	d.Set("status", workspace.Status)

	if err := d.Set("tags", KeyValueTags(workspace.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	return nil
}
