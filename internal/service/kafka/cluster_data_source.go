package kafka

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kafka"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

func DataSourceCluster() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceClusterRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"bootstrap_brokers": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"bootstrap_brokers_public_sasl_iam": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"bootstrap_brokers_public_sasl_scram": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"bootstrap_brokers_public_tls": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"bootstrap_brokers_sasl_iam": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"bootstrap_brokers_sasl_scram": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"bootstrap_brokers_tls": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cluster_name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 64),
			},
			"kafka_version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"number_of_broker_nodes": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"tags": tftags.TagsSchemaComputed(),
			"zookeeper_connect_string": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"zookeeper_connect_string_tls": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceClusterRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).KafkaConn
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	clusterName := d.Get("cluster_name").(string)
	input := &kafka.ListClustersInput{
		ClusterNameFilter: aws.String(clusterName),
	}
	var cluster *kafka.ClusterInfo

	err := conn.ListClustersPages(input, func(page *kafka.ListClustersOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, clusterInfo := range page.ClusterInfoList {
			if aws.StringValue(clusterInfo.ClusterName) == clusterName {
				cluster = clusterInfo

				return false
			}
		}

		return !lastPage
	})

	if err != nil {
		return fmt.Errorf("error listing MSK Clusters: %w", err)
	}

	if cluster == nil {
		return fmt.Errorf("error reading MSK Cluster (%s): no results found", clusterName)
	}

	bootstrapBrokersInput := &kafka.GetBootstrapBrokersInput{
		ClusterArn: cluster.ClusterArn,
	}

	bootstrapBrokersOutput, err := conn.GetBootstrapBrokers(bootstrapBrokersInput)

	if err != nil {
		return fmt.Errorf("error reading MSK Cluster (%s) bootstrap brokers: %w", aws.StringValue(cluster.ClusterArn), err)
	}

	d.Set("arn", cluster.ClusterArn)
	d.Set("bootstrap_brokers", SortEndpointsString(aws.StringValue(bootstrapBrokersOutput.BootstrapBrokerString)))
	d.Set("bootstrap_brokers_public_sasl_iam", SortEndpointsString(aws.StringValue(bootstrapBrokersOutput.BootstrapBrokerStringPublicSaslIam)))
	d.Set("bootstrap_brokers_public_sasl_scram", SortEndpointsString(aws.StringValue(bootstrapBrokersOutput.BootstrapBrokerStringPublicSaslScram)))
	d.Set("bootstrap_brokers_public_tls", SortEndpointsString(aws.StringValue(bootstrapBrokersOutput.BootstrapBrokerStringPublicTls)))
	d.Set("bootstrap_brokers_sasl_iam", SortEndpointsString(aws.StringValue(bootstrapBrokersOutput.BootstrapBrokerStringSaslIam)))
	d.Set("bootstrap_brokers_sasl_scram", SortEndpointsString(aws.StringValue(bootstrapBrokersOutput.BootstrapBrokerStringSaslScram)))
	d.Set("bootstrap_brokers_tls", SortEndpointsString(aws.StringValue(bootstrapBrokersOutput.BootstrapBrokerStringTls)))
	d.Set("cluster_name", cluster.ClusterName)
	d.Set("kafka_version", cluster.CurrentBrokerSoftwareInfo.KafkaVersion)
	d.Set("number_of_broker_nodes", cluster.NumberOfBrokerNodes)

	if err := d.Set("tags", KeyValueTags(cluster.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	d.Set("zookeeper_connect_string", SortEndpointsString(aws.StringValue(cluster.ZookeeperConnectString)))
	d.Set("zookeeper_connect_string_tls", SortEndpointsString(aws.StringValue(cluster.ZookeeperConnectStringTls)))

	d.SetId(aws.StringValue(cluster.ClusterArn))

	return nil
}
