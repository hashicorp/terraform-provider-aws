package amp

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/prometheusservice"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

func DataSourceWorkspace() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceWorkspaceRead,

		Schema: map[string]*schema.Schema{
			"alias": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"workspace_id"},
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
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"alias"},
			},
		},
	}
}

func dataSourceWorkspaceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AMPConn
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	workspaceID := d.Get("workspace_id").(string)
	workspaceAlias := d.Get("alias").(string)

	var workspace *prometheusservice.WorkspaceDescription
	var err error

	if workspaceID != "" {
		workspace, err = FindWorkspaceByID(ctx, conn, workspaceID)
	} else if workspaceAlias != "" {
		workspace, err = FindWorkspaceByAlias(ctx, conn, workspaceAlias)
	}

	if err != nil {
		if workspaceID != "" {
			return diag.Errorf("reading AMP Workspace by ID (%s): %s", workspaceID, err)
		} else {
			return diag.Errorf("reading AMP Workspace by alias (%s): %s", workspaceAlias, err)
		}
	}

	d.SetId(aws.StringValue(workspace.WorkspaceId))

	d.Set("alias", workspace.Alias)
	d.Set("arn", workspace.Arn)
	d.Set("created_date", workspace.CreatedAt.Format(time.RFC3339))
	d.Set("prometheus_endpoint", workspace.PrometheusEndpoint)
	d.Set("status", workspace.Status.StatusCode)

	if err := d.Set("tags", KeyValueTags(workspace.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return diag.Errorf("setting tags: %s", err)
	}

	return nil
}
