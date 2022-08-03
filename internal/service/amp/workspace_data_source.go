package amp

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/prometheusservice"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

func DataSourceWorkspace() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceWorkspaceRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"workspace_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"prometheus_endpoint": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"alias": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourceWorkspaceRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).AMPConn
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	workspaceID := d.Get("workspace_id").(string)
	workspace, err := FindWorkspaceByID(conn, workspaceID)

	if err != nil {
		return fmt.Errorf("error reading AMP Workspace (%s): %w", workspaceID, err)
	}

	d.SetId(workspaceID)

	workspaceARN := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   prometheusservice.ServiceName,
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("/workspaces/%s", d.Id()),
	}.String()
	d.Set("arn", workspaceARN)
	d.Set("prometheus_endpoint", workspace.PrometheusEndpoint)
	d.Set("alias", workspace.Alias)
	d.Set("created_date", workspace.CreatedAt.Format(time.RFC3339))
	d.Set("status", workspace.Status.StatusCode)

	if err := d.Set("tags", KeyValueTags(workspace.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	return nil
}
