package aws

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kafka"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
)

func resourceAwsMskCluster() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsMskClusterCreate,
		Read:   resourceAwsMskClusterRead,
		Update: resourceAwsMskClusterUpdate,
		Delete: resourceAwsMskClusterDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(45 * time.Minute),
			Delete: schema.DefaultTimeout(45 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"bootstrap_brokers": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"broker_node_group_info": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"az_distribution": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
							Default:  kafka.BrokerAZDistributionDefault,
							ValidateFunc: validation.StringInSlice([]string{
								kafka.BrokerAZDistributionDefault,
							}, true),
						},
						"client_subnets": {
							Type:     schema.TypeList,
							Required: true,
							ForceNew: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"instance_type": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"security_groups": {
							Type:     schema.TypeList,
							Required: true,
							ForceNew: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"ebs_volume_size": {
							Type:         schema.TypeInt,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validation.IntBetween(1, 16384),
						},
					},
				},
			},
			"cluster_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 64),
			},
			"encryption_info": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"encryption_at_rest_kms_id": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
							ForceNew: true,
							DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
								// MSK api accepts either KMS short id or arn, but always returns arn.
								// treat them as equivalent
								return strings.Contains(old, new)
							},
						},
					},
				},
			},
			"enhanced_monitoring": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  kafka.EnhancedMonitoringDefault,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					kafka.EnhancedMonitoringDefault,
					kafka.EnhancedMonitoringPerBroker,
					kafka.EnhancedMonitoringPerTopicPerBroker,
				}, true),
			},
			"kafka_version": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 64),
			},
			"number_of_broker_nodes": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},
			"tags": tagsSchema(),
			"zookeeper_connect_string": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsMskClusterCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).kafkaconn

	nodeInfo := d.Get("broker_node_group_info").([]interface{})[0].(map[string]interface{})

	input := &kafka.CreateClusterInput{
		ClusterName:         aws.String(d.Get("cluster_name").(string)),
		EnhancedMonitoring:  aws.String(d.Get("enhanced_monitoring").(string)),
		NumberOfBrokerNodes: aws.Int64(int64(d.Get("number_of_broker_nodes").(int))),
		BrokerNodeGroupInfo: &kafka.BrokerNodeGroupInfo{
			BrokerAZDistribution: aws.String(nodeInfo["az_distribution"].(string)),
			InstanceType:         aws.String(nodeInfo["instance_type"].(string)),
			StorageInfo: &kafka.StorageInfo{
				EbsStorageInfo: &kafka.EBSStorageInfo{
					VolumeSize: aws.Int64(int64(nodeInfo["ebs_volume_size"].(int))),
				},
			},
			ClientSubnets:  expandStringList(nodeInfo["client_subnets"].([]interface{})),
			SecurityGroups: expandStringList(nodeInfo["security_groups"].([]interface{})),
		},
		KafkaVersion: aws.String(d.Get("kafka_version").(string)),
	}

	if v, ok := d.GetOk("encryption_info"); ok {
		info := v.([]interface{})
		if len(info) == 1 {
			if info[0] == nil {
				return fmt.Errorf("At least one item is expected inside encryption_info")

			}

			i := info[0].(map[string]interface{})

			input.EncryptionInfo = &kafka.EncryptionInfo{
				EncryptionAtRest: &kafka.EncryptionAtRest{
					DataVolumeKMSKeyId: aws.String(i["encryption_at_rest_kms_id"].(string)),
				},
			}
		}
	}

	var out *kafka.CreateClusterOutput
	err := resource.Retry(30*time.Second, func() *resource.RetryError {
		var err error
		out, err = conn.CreateCluster(input)

		if err != nil {
			if isAWSErr(err, kafka.ErrCodeTooManyRequestsException, "Too Many Requests") {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("error creating MSK cluster (%s): %s", d.Id(), err)
	}

	d.SetId(aws.StringValue(out.ClusterArn))

	// set tags while cluster is being created
	tags := tagsFromMapMskCluster(d.Get("tags").(map[string]interface{}))

	if err := setTagsMskCluster(conn, d, aws.StringValue(out.ClusterArn)); err != nil {
		return err
	}

	d.Set("tags", tagsToMapMskCluster(tags))
	d.SetPartial("tags")

	log.Printf("[DEBUG] Waiting for MSK cluster %q to be created", d.Id())
	err = waitForClusterCreation(conn, d.Id())
	if err != nil {
		return fmt.Errorf("error waiting for MSK cluster creation (%s): %s", d.Id(), err)
	}
	d.Partial(false)

	log.Printf("[DEBUG] MSK cluster %q created", d.Id())

	return resourceAwsMskClusterRead(d, meta)
}

func waitForClusterCreation(conn *kafka.Kafka, arn string) error {
	return resource.Retry(60*time.Minute, func() *resource.RetryError {
		out, err := conn.DescribeCluster(&kafka.DescribeClusterInput{
			ClusterArn: aws.String(arn),
		})
		if err != nil {
			return resource.NonRetryableError(err)
		}
		if out.ClusterInfo != nil {
			if aws.StringValue(out.ClusterInfo.State) == kafka.ClusterStateFailed {
				return resource.NonRetryableError(fmt.Errorf("Cluster creation failed with cluster state %q", kafka.ClusterStateFailed))
			}
			if aws.StringValue(out.ClusterInfo.State) == kafka.ClusterStateActive {
				return nil
			}
		}
		return resource.RetryableError(fmt.Errorf("%q: cluster still creating", arn))
	})
}

func resourceAwsMskClusterRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).kafkaconn

	out, err := conn.DescribeCluster(&kafka.DescribeClusterInput{
		ClusterArn: aws.String(d.Id()),
	})
	if err != nil {
		if isAWSErr(err, kafka.ErrCodeNotFoundException, "") {
			log.Printf("[WARN] MSK Cluster (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("failed lookup cluster %s: %s", d.Id(), err)
	}

	brokerOut, err := conn.GetBootstrapBrokers(&kafka.GetBootstrapBrokersInput{
		ClusterArn: aws.String(d.Id()),
	})
	if err != nil {
		return fmt.Errorf("failed requesting bootstrap broker info for %q : %s", d.Id(), err)
	}

	cluster := out.ClusterInfo

	d.SetId(aws.StringValue(cluster.ClusterArn))
	d.Set("arn", aws.StringValue(cluster.ClusterArn))
	d.Set("bootstrap_brokers", brokerOut.BootstrapBrokerString)

	d.Set("broker_node_group_info", flattenMskBrokerNodeGroupInfo(cluster.BrokerNodeGroupInfo))

	d.Set("cluster_name", aws.StringValue(cluster.ClusterName))
	d.Set("enhanced_monitoring", aws.StringValue(cluster.EnhancedMonitoring))
	d.Set("encryption_info", flattenMskEncryptionInfo(cluster.EncryptionInfo))
	d.Set("kafka_version", aws.StringValue(cluster.CurrentBrokerSoftwareInfo.KafkaVersion))
	d.Set("number_of_broker_nodes", aws.Int64Value(cluster.NumberOfBrokerNodes))
	d.Set("zookeeper_connect_string", cluster.ZookeeperConnectString)

	listTagsOut, err := conn.ListTagsForResource(&kafka.ListTagsForResourceInput{
		ResourceArn: cluster.ClusterArn,
	})
	if err != nil {
		return fmt.Errorf("failed listing tags for msk cluster %q: %s", d.Id(), err)
	}

	d.Set("tags", tagsToMapMskCluster(listTagsOut.Tags))

	return nil
}

func resourceAwsMskClusterUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).kafkaconn

	// currently tags are the only thing that are updatable..
	if err := setTagsMskCluster(conn, d, d.Id()); err != nil {
		return fmt.Errorf("failed updating tags for msk cluster %q: %s", d.Id(), err)
	}

	return resourceAwsMskClusterRead(d, meta)

}

func flattenMskBrokerNodeGroupInfo(b *kafka.BrokerNodeGroupInfo) []map[string]interface{} {

	if b == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"az_distribution": aws.StringValue(b.BrokerAZDistribution),
		"client_subnets":  flattenStringList(b.ClientSubnets),
		"instance_type":   aws.StringValue(b.InstanceType),
		"security_groups": flattenStringList(b.SecurityGroups),
	}
	if b.StorageInfo != nil {
		if b.StorageInfo.EbsStorageInfo != nil {
			m["ebs_volume_size"] = int(aws.Int64Value(b.StorageInfo.EbsStorageInfo.VolumeSize))
		}
	}
	return []map[string]interface{}{m}
}

func flattenMskEncryptionInfo(e *kafka.EncryptionInfo) []map[string]interface{} {
	if e == nil || e.EncryptionAtRest == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"encryption_at_rest_kms_id": aws.StringValue(e.EncryptionAtRest.DataVolumeKMSKeyId),
	}

	return []map[string]interface{}{m}
}

func resourceAwsMskClusterDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).kafkaconn

	log.Printf("[DEBUG] Deleting MSK cluster: %q", d.Id())
	_, err := conn.DeleteCluster(&kafka.DeleteClusterInput{
		ClusterArn: aws.String(d.Id()),
	})
	if err != nil {
		if isAWSErr(err, kafka.ErrCodeNotFoundException, "") {
			return nil
		}
		return fmt.Errorf("failed deleting MSK cluster %q: %s", d.Id(), err)
	}

	log.Printf("[DEBUG] Waiting for MSK cluster %q to be deleted", d.Id())

	return resourceAwsMskClusterDeleteWaiter(conn, d.Id())
}

func resourceAwsMskClusterDeleteWaiter(conn *kafka.Kafka, arn string) error {
	return resource.Retry(60*time.Minute, func() *resource.RetryError {
		_, err := conn.DescribeCluster(&kafka.DescribeClusterInput{
			ClusterArn: aws.String(arn),
		})

		if err != nil {
			if isAWSErr(err, kafka.ErrCodeNotFoundException, "") {
				return nil
			}
			return resource.NonRetryableError(err)
		}

		return resource.RetryableError(fmt.Errorf("timeout while waiting for the cluster %q to be deleted", arn))
	})
}
