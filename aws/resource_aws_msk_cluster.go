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
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				ForceNew: true,
			},
			"broker_volume_size": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},
			"encrypt_rest_arn": {
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
				Optional: true,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"encrypt_rest_key": {
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
	encryptRestArn := d.Get("encrypt_rest_arn").(string)
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

	if encryptRestArn != "" {
		createOpts.EncryptionInfo = &kafka.EncryptionInfo{}
		createOpts.EncryptionInfo.EncryptionAtRest = &kafka.EncryptionAtRest{
			DataVolumeKMSKeyId: aws.String(encryptRestArn),
		}
	}

	_, err := conn.CreateCluster(createOpts)
	if err != nil {
		return fmt.Errorf("Unable to create cluster: %s", err)
	}

	// No error, wait for ACTIVE state
	stateConf := &resource.StateChangeConf{
		Pending:    []string{kafka.ClusterStateCreating},
		Target:     []string{kafka.ClusterStateActive},
		Refresh:    clusterStateRefreshFunc(conn, cn),
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
	d.Set("status", c.status)
	d.Set("zookeeper_connect", c.zookeeperConnect)
	d.Set("bootstrap_brokers", c.bootstrapBrokers)

	return nil
}
func resourceAwsMskClusterRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).kafkaconn
	name := d.Get("name").(string)

	state, err := readMskClusterState(conn, name)
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
	d.Set("status", state.status)
	d.Set("encrypt_rest_key", state.encryptRestKey)
	d.Set("zookeeper_connect", state.zookeeperConnect)
	d.Set("bootstrap_brokers", state.bootstrapBrokers)

	return nil
}

func resourceAwsMskClusterDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).kafkaconn
	arn := d.Get("arn").(string)
	name := d.Get("name").(string)

	_, err := conn.DeleteCluster(&kafka.DeleteClusterInput{
		ClusterArn: aws.String(arn),
	})
	if err != nil {
		return err
	}

	stateConf := &resource.StateChangeConf{
		Pending:    []string{kafka.ClusterStateDeleting},
		Target:     []string{"DESTROYED"},
		Refresh:    clusterStateRefreshFunc(conn, name),
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
	encryptRestKey    string
	zookeeperConnect  string
	bootstrapBrokers  string
	brokerCount       int64
	kafkaVersion      string
}

func readMskClusterState(conn *kafka.Kafka, name string) (*mskClusterState, error) {
	listOpts := &kafka.ListClustersInput{
		ClusterNameFilter: aws.String(name),
	}

	cluster_list, err := conn.ListClusters(listOpts)

	if len(cluster_list.ClusterInfoList) == 0 {
		return nil, awserr.New("NotFoundException", fmt.Sprintf("MSK Cluster (%s) not found", name), nil)
	}

	if len(cluster_list.ClusterInfoList) > 1 {
		return nil, fmt.Errorf("Ambiguous MSK Cluster name (%s)", name)
	}

	cluster := cluster_list.ClusterInfoList[0]
	state := &mskClusterState{}

	if cluster != nil {
		state.arn = aws.StringValue(cluster.ClusterArn)
		state.creationTimestamp = aws.TimeValue(cluster.CreationTime).Unix()
		state.status = aws.StringValue(cluster.State)
		state.encryptRestKey = aws.StringValue(cluster.EncryptionInfo.EncryptionAtRest.DataVolumeKMSKeyId)
		state.brokerCount = aws.Int64Value(cluster.NumberOfBrokerNodes)
		state.zookeeperConnect = aws.StringValue(cluster.ZookeeperConnectString)
		state.kafkaVersion = aws.StringValue(cluster.CurrentBrokerSoftwareInfo.KafkaVersion)

		if state.status == kafka.ClusterStateActive {
			bb, bb_err := conn.GetBootstrapBrokers(
				&kafka.GetBootstrapBrokersInput{
					ClusterArn: cluster.ClusterArn,
				})

			if bb_err == nil {
				state.bootstrapBrokers = aws.StringValue(bb.BootstrapBrokerString)
			} else {
				log.Println(bb_err)
			}
		}
	}

	return state, err
}

func clusterStateRefreshFunc(conn *kafka.Kafka, name string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		state, err := readMskClusterState(conn, name)
		if err != nil {
			if awsErr, ok := err.(awserr.Error); ok {
				if awsErr.Code() == "NotFoundException" {
					return 42, "DESTROYED", nil
				}
				return nil, awsErr.Code(), err
			}
			return nil, "failed", err
		}

		if state.status == kafka.ClusterStateFailed {
			return nil, "failed", errors.New("MSK Cluster in FAILED state")
		}

		return state, state.status, nil
	}
}
