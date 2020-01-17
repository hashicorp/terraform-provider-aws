package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func dataSourceAwsEcsCluster() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsEcsClusterRead,

		Schema: map[string]*schema.Schema{
			"cluster_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"pending_tasks_count": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			"running_tasks_count": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			"registered_container_instances_count": {
				Type:     schema.TypeInt,
				Computed: true,
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
			"capacity_providers": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"default_capacity_provider_strategy": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"base": {
							Type:     schema.TypeInt,
							Computed: true,
						},

						"capacity_provider": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"weight": {
							Type:     schema.TypeInt,
							Computed: true,
						},
					},
				},
			},
			"tags": tagsSchemaComputed(),
		},
	}
}

func dataSourceAwsEcsClusterRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ecsconn

	params := &ecs.DescribeClustersInput{
		Clusters: []*string{aws.String(d.Get("cluster_name").(string))},
	}
	log.Printf("[DEBUG] Reading ECS Cluster: %s", params)
	desc, err := conn.DescribeClusters(params)

	if err != nil {
		return err
	}

	if len(desc.Clusters) == 0 {
		return fmt.Errorf("no matches found for name: %s", d.Get("cluster_name").(string))
	}

	if len(desc.Clusters) > 1 {
		return fmt.Errorf("multiple matches found for name: %s", d.Get("cluster_name").(string))
	}

	cluster := desc.Clusters[0]
	d.SetId(aws.StringValue(cluster.ClusterArn))
	d.Set("arn", cluster.ClusterArn)
	d.Set("status", cluster.Status)
	d.Set("pending_tasks_count", cluster.PendingTasksCount)
	d.Set("running_tasks_count", cluster.RunningTasksCount)
	d.Set("registered_container_instances_count", cluster.RegisteredContainerInstancesCount)

	if err := d.Set("setting", flattenEcsSettings(cluster.Settings)); err != nil {
		return fmt.Errorf("error setting setting: %s", err)
	}
	if err := d.Set("capacity_providers", aws.StringValueSlice(cluster.CapacityProviders)); err != nil {
		return fmt.Errorf("error setting capacity_providers: %s", err)
	}
	if err := d.Set("default_capacity_provider_strategy", flattenEcsCapacityProviderStrategy(cluster.DefaultCapacityProviderStrategy)); err != nil {
		return fmt.Errorf("error setting default_capacity_provider_strategy: %s", err)
	}
	if err := d.Set("tags", keyvaluetags.EcsKeyValueTags(cluster.Tags).IgnoreAws().Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	return nil
}
