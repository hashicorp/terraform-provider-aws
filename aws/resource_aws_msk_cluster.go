package aws

import (
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/kafka"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
	"log"
	"time"
)

func resourceAwsMskCluster() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsMskClusterCreate,
		Read:   resourceAwsMskClusterRead,
		Delete: resourceAwsMskClusterDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(120 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 64),
			},
			"client_subnets": {
				Type:     schema.TypeList,
				Required: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				ForceNew: true,
			},
			"broker_count": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},
			"broker_instance_type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"broker_security_groups": {
				Type:     schema.TypeList,
				Required: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				ForceNew: true,
			},
			"broker_volume_size": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},
			"encryption_key": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
				ForceNew: true,
			},
			"enhanced_monitoring": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  kafka.EnhancedMonitoringDefault,
				ValidateFunc: validation.StringInSlice([]string{
					kafka.EnhancedMonitoringDefault,
					kafka.EnhancedMonitoringPerBroker,
					kafka.EnhancedMonitoringPerTopicPerBroker,
				}, true),
				ForceNew: true,
			},
			"kafka_version": {
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
			"zookeeper_connect": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"bootstrap_brokers": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsMskClusterCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).kafkaconn
	cn := d.Get("name").(string)
	clientSubnets := d.Get("client_subnets")
	securityGroups := d.Get("broker_security_groups")

	createOpts := &kafka.CreateClusterInput{
		ClusterName:         aws.String(cn),
		EnhancedMonitoring:  aws.String(d.Get("enhanced_monitoring").(string)),
		NumberOfBrokerNodes: aws.Int64(int64(d.Get("broker_count").(int))),
		BrokerNodeGroupInfo: &kafka.BrokerNodeGroupInfo{
			BrokerAZDistribution: aws.String(kafka.BrokerAZDistributionDefault),
			ClientSubnets:        expandStringList(clientSubnets.([]interface{})),
			InstanceType:         aws.String(d.Get("broker_instance_type").(string)),
			SecurityGroups:       expandStringList(securityGroups.([]interface{})),
			StorageInfo: &kafka.StorageInfo{
				EbsStorageInfo: &kafka.EBSStorageInfo{
					VolumeSize: aws.Int64(int64(d.Get("broker_volume_size").(int))),
				},
			},
		},
		KafkaVersion: aws.String(d.Get("kafka_version").(string)),
	}

	if v, ok := d.GetOk("encryption_key"); ok {
		createOpts.EncryptionInfo = &kafka.EncryptionInfo{
			EncryptionAtRest: &kafka.EncryptionAtRest{
				DataVolumeKMSKeyId: aws.String(v.(string)),
			},
		}
	}

	clusterOutput, err := conn.CreateCluster(createOpts)
	if err != nil {
		return fmt.Errorf("Unable to create cluster: %s", err)
	}

	d.SetId(*clusterOutput.ClusterArn)

	// No error, wait for ACTIVE state
	stateConf := &resource.StateChangeConf{
		Pending:    []string{kafka.ClusterStateCreating},
		Target:     []string{kafka.ClusterStateActive},
		Refresh:    clusterStateRefreshFunc(conn, d.Id()),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("Error waiting for MSK cluster (%s) to become active: %s", cn, err)
	}

	return resourceAwsMskClusterRead(d, meta)
}

func resourceAwsMskClusterRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).kafkaconn
	id := d.Id()

	state, err := readMskClusterState(conn, id)
	if isAWSErr(err, kafka.ErrCodeNotFoundException, "") {
		log.Printf("[WARN] MSK Cluster (%s) not found, removing from state", id)
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading MSK Cluster (%s): %s", id, err)
	}

	d.SetId(*state.ClusterArn)
	d.Set("arn", *state.ClusterArn)
	d.Set("status", *state.State)
	d.Set("encryption_key", *state.EncryptionInfo.EncryptionAtRest.DataVolumeKMSKeyId)
	d.Set("zookeeper_connect", *state.ZookeeperConnectString)

	if *state.State == kafka.ClusterStateActive {
		bb, err := conn.GetBootstrapBrokers(&kafka.GetBootstrapBrokersInput{ClusterArn: state.ClusterArn})
		if err != nil {
			return err
		}

		d.Set("bootstrap_brokers", aws.StringValue(bb.BootstrapBrokerString))
	}

	return nil
}

func resourceAwsMskClusterDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).kafkaconn
	id := d.Id()

	_, err := conn.DeleteCluster(&kafka.DeleteClusterInput{ClusterArn: aws.String(id)})
	if err != nil {
		return fmt.Errorf("error deleting MSK Cluster (%s): %s", id, err)
	}

	stateConf := &resource.StateChangeConf{
		Pending:    []string{kafka.ClusterStateDeleting},
		Target:     []string{"DESTROYED"},
		Refresh:    clusterStateRefreshFunc(conn, id),
		Timeout:    d.Timeout(schema.TimeoutDelete),
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("error waiting for MSK Cluster (%s) to be destroyed: %s", id, err)
	}

	return nil
}

func readMskClusterState(conn *kafka.Kafka, id string) (*kafka.ClusterInfo, error) {
	clusterOutput, err := conn.DescribeCluster(&kafka.DescribeClusterInput{ClusterArn: &id})
	if err != nil {
		return &kafka.ClusterInfo{}, err
	}

	return clusterOutput.ClusterInfo, err
}

func clusterStateRefreshFunc(conn *kafka.Kafka, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		state, err := readMskClusterState(conn, id)
		if err != nil {
			if awsErr, ok := err.(awserr.Error); ok {
				if awsErr.Code() == "NotFoundException" {
					return 42, "DESTROYED", nil
				}
				return nil, awsErr.Code(), err
			}
			return nil, "failed", err
		}

		if *state.State == kafka.ClusterStateFailed {
			return nil, "failed", errors.New("MSK Cluster in FAILED state")
		}

		return state, *state.State, nil
	}
}
