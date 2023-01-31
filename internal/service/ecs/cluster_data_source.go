package ecs

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func DataSourceCluster() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceClusterRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cluster_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"pending_tasks_count": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"registered_container_instances_count": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"running_tasks_count": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"service_connect_defaults": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"namespace": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"setting": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"value": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceClusterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ECSConn()

	clusterName := d.Get("cluster_name").(string)
	cluster, err := FindClusterByNameOrARN(ctx, conn, d.Get("cluster_name").(string))

	if err != nil {
		return diag.Errorf("reading ECS Cluster (%s): %s", clusterName, err)
	}

	d.SetId(aws.StringValue(cluster.ClusterArn))
	d.Set("arn", cluster.ClusterArn)
	d.Set("pending_tasks_count", cluster.PendingTasksCount)
	d.Set("running_tasks_count", cluster.RunningTasksCount)
	d.Set("registered_container_instances_count", cluster.RegisteredContainerInstancesCount)
	d.Set("status", cluster.Status)

	if cluster.ServiceConnectDefaults != nil {
		if err := d.Set("service_connect_defaults", []interface{}{flattenClusterServiceConnectDefaults(cluster.ServiceConnectDefaults)}); err != nil {
			return diag.Errorf("setting service_connect_defaults: %s", err)
		}
	} else {
		d.Set("service_connect_defaults", nil)
	}

	if err := d.Set("setting", flattenClusterSettings(cluster.Settings)); err != nil {
		return diag.Errorf("setting setting: %s", err)
	}

	return nil
}
