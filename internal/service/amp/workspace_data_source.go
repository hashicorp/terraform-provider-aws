package amp

import (
	"context"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/prometheusservice"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

// @SDKDataSource("aws_prometheus_workspace")
func DataSourceWorkspace() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceWorkspaceRead,

		Schema: map[string]*schema.Schema{
			"alias": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{"alias", "workspace_id"},
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"prometheus_endpoint": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tftags.TagsSchemaComputed(),
			"workspace_id": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{"alias", "workspace_id"},
			},
		},
	}
}

func dataSourceWorkspaceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AMPConn()
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	var err error
	workspace := &prometheusservice.WorkspaceDescription{}

	if v, ok := d.GetOk("workspace_id"); ok {
		workspaceID := v.(string)
		workspace, err = FindWorkspaceByID(ctx, conn, workspaceID)

		if err != nil {
			return diag.Errorf("reading AMP Workspace (%s): %s", workspaceID, err)
		}
	} else if v, ok := d.GetOk("alias"); ok {
		workspaceAlias := v.(string)
		workspace, err = FindWorkspaceByAlias(ctx, conn, workspaceAlias)

		if err != nil {
			return diag.Errorf("reading AMP Workspace (%s): %s", workspaceAlias, err)
		}
	}

	d.SetId(aws.StringValue(workspace.WorkspaceId))

	d.Set("alias", workspace.Alias)
	d.Set("arn", workspace.Arn)
	d.Set("created_date", workspace.CreatedAt.Format(time.RFC3339))
	d.Set("prometheus_endpoint", workspace.PrometheusEndpoint)
	d.Set("status", workspace.Status.StatusCode)

	if err := d.Set("tags", KeyValueTags(ctx, workspace.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return diag.Errorf("setting tags: %s", err)
	}

	return nil
}
