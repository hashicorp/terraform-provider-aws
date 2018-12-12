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
	"time"
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
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(120 * time.Minute),
			Delete: schema.DefaultTimeout(120 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"client_subnets": {
				Type:     schema.TypeList,
				Required: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"broker_instance_type": {
				Type:     schema.TypeString,
				Required: true,
			},
			"broker_security_groups": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"broker_volume_size": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"encrypt_rest_key": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
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
			},
			"kafka_version": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "1.1.1",
			},
			"broker_count": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"encrypt_rest_arn": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"zookeeper_connect": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"bootstrap_brokers": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
		},
	}
}
func resourceAwsMskClusterCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).kafkaconn
	cn := d.Get("name").(string)
	encryptRestKey := d.Get("encrypt_rest_key").(string)
	clientSubnets, _ := d.GetOk("client_subnets")
	securityGroups, _ := d.GetOk("broker_security_groups")

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

	if encryptRestKey != "" {
		createOpts.EncryptionInfo = &kafka.EncryptionInfo{}
		createOpts.EncryptionInfo.EncryptionAtRest = &kafka.EncryptionAtRest{
			DataVolumeKMSKeyId: aws.String(encryptRestKey),
		}
	}

	cluster, err := conn.CreateCluster(createOpts)
	if err != nil {
		return fmt.Errorf("Unable to create cluster: %s", err)
	}

	arn := aws.StringValue(cluster.ClusterArn)

	// No error, wait for ACTIVE state
	stateConf := &resource.StateChangeConf{
		Pending:    []string{"CREATING"},
		Target:     []string{"ACTIVE"},
		Refresh:    clusterStateRefreshFunc(conn, arn),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	clusterRaw, err := stateConf.WaitForState()

	if err != nil {
		return fmt.Errorf(
			"Error waiting for MSK cluster (%s) to become active: %s",
			cn, err)
	}

	c := clusterRaw.(*mskClusterState)
	d.SetId(c.arn)
	d.Set("arn", c.arn)

	return resourceAwsMskClusterUpdate(d, meta)
}
func resourceAwsMskClusterRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).kafkaconn
	arn := d.Get("arn").(string)

	state, err := readMskClusterState(conn, arn)
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			if awsErr.Code() == "ResourceNotFoundException" {
				d.SetId("")
				return nil
			}
			return fmt.Errorf("Error reading MSK cluster: \"%s\", code: \"%s\"", awsErr.Message(), awsErr.Code())
		}
		return err

	}
	d.SetId(state.arn)
	d.Set("arn", state.arn)
	d.Set("encrypt_rest_arn", state.encryptRestArn)
	d.Set("zookeeper_connect", state.zookeeper_connect)
	d.Set("bootstrap_brokers", state.bootstrap_brokers)

	return nil
}

func resourceAwsMskClusterUpdate(d *schema.ResourceData, meta interface{}) error {
	// TODO: Figure out update as API calls not yet implemented
	return nil
}

func resourceAwsMskClusterDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).kafkaconn
	arn := d.Get("arn").(string)

	_, err := conn.DeleteCluster(&kafka.DeleteClusterInput{
		ClusterArn: aws.String(arn),
	})
	if err != nil {
		return err
	}

	stateConf := &resource.StateChangeConf{
		Pending:    []string{"DELETING"},
		Target:     []string{"DESTROYED"},
		Refresh:    clusterStateRefreshFunc(conn, arn),
		Timeout:    d.Timeout(schema.TimeoutDelete),
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf(
			"Error waiting for MSK Cluster (%s) to be destroyed: %s",
			arn, err)
	}

	return nil
}

type mskClusterState struct {
	arn               string
	creationTimestamp int64
	status            string
	encryptRestArn    string
	zookeeper_connect string
	bootstrap_brokers string
}

func readMskClusterState(conn *kafka.Kafka, arn string) (*mskClusterState, error) {
	describeOpts := &kafka.DescribeClusterInput{
		ClusterArn: aws.String(arn),
	}

	state := &mskClusterState{}
	cluster, err := conn.DescribeCluster(describeOpts)

	if cluster.ClusterInfo != nil {
		state.arn = aws.StringValue(cluster.ClusterInfo.ClusterArn)
		state.creationTimestamp = aws.TimeValue(cluster.ClusterInfo.CreationTime).Unix()
		state.status = aws.StringValue(cluster.ClusterInfo.State)
		state.encryptRestArn = aws.StringValue(cluster.ClusterInfo.EncryptionInfo.EncryptionAtRest.DataVolumeKMSKeyId)
		state.zookeeper_connect = aws.StringValue(cluster.ClusterInfo.ZookeeperConnectString)

		bb, bb_err := conn.GetBootstrapBrokers(
			&kafka.GetBootstrapBrokersInput{
				ClusterArn: cluster.ClusterInfo.ClusterArn,
			})

		if bb_err == nil {
			state.bootstrap_brokers = aws.StringValue(bb.BootstrapBrokerString)
		}
	}

	return state, err
}

func clusterStateRefreshFunc(conn *kafka.Kafka, arn string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		state, err := readMskClusterState(conn, arn)
		if err != nil {
			if awsErr, ok := err.(awserr.Error); ok {
				if awsErr.Code() == "NotFoundException" {
					return 42, "DESTROYED", nil
				}
				return nil, awsErr.Code(), err
			}
			return nil, "failed", err
		}

		if state.status == "FAILED" {
			return nil, "failed", errors.New("MSK Cluster in FAILED state")
		}

		return state, state.status, nil
	}
}
