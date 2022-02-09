package grafana

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DataSourceWorkspace() *schema.Resource {
	return &schema.Resource{
		Read: resourceWorkspaceRead,

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
