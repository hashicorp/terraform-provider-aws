package aws

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsKinesisStreamConsumer() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsKinesisStreamConsumerCreate,
		Read:   resourceAwsKinesisStreamConsumerRead,
		Delete: resourceAwsKinesisStreamConsumerDelete,
		Importer: &schema.ResourceImporter{
			State: resourceAwsKinesisStreamConsumerImport,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(120 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"stream_arn": {
				Type:     schema.TypeString,
				Required: true,
			},

			"arn": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func resourceAwsKinesisStreamConsumerImport(
	d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	d.Set("name", d.Id())
	return []*schema.ResourceData{d}, nil
}

func resourceAwsKinesisStreamConsumerCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).kinesisconn
	cn := d.Get("name").(string)
	sa := d.Get("stream_arn").(string)
	createOpts := &kinesis.RegisterStreamConsumerInput{
		ConsumerName: aws.String(cn),
		StreamARN:    aws.String(sa),
	}

	_, err := conn.RegisterStreamConsumer(createOpts)
	if err != nil {
		return fmt.Errorf("Unable to create stream consumer: %s", err)
	}

	// No error, wait for ACTIVE state
	stateConf := &resource.StateChangeConf{
		Pending:    []string{"CREATING"},
		Target:     []string{"ACTIVE"},
		Refresh:    streamConsumerStateRefreshFunc(conn, cn),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	streamRaw, err := stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf(
			"Error waiting for Kinesis Stream Consumer (%s) to become active: %s",
			cn, err)
	}

	s := streamRaw.(*kinesisStreamConsumerState)

	d.SetId(s.arn)
	d.Set("arn", s.arn)

	return resourceAwsKinesisStreamConsumerRead(d, meta)
}

func resourceAwsKinesisStreamConsumerRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).kinesisconn
	cn := d.Get("name").(string)

	state, err := readKinesisStreamConsumerState(conn, cn)
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			if awsErr.Code() == "ResourceNotFoundException" {
				d.SetId("")
				return nil
			}
			return fmt.Errorf("Error reading Kinesis Stream Consumer: \"%s\", code: \"%s\"", awsErr.Message(), awsErr.Code())
		}
		return err

	}
	d.SetId(state.arn)
	d.Set("arn", state.arn)

	return nil
}

func resourceAwsKinesisStreamConsumerDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).kinesisconn
	cn := d.Get("name").(string)

	_, err := conn.DeregisterStreamConsumer(&kinesis.DeregisterStreamConsumerInput{
		ConsumerName: aws.String(cn),
	})
	if err != nil {
		return err
	}

	stateConf := &resource.StateChangeConf{
		Pending:    []string{"DELETING"},
		Target:     []string{"DESTROYED"},
		Refresh:    streamConsumerStateRefreshFunc(conn, cn),
		Timeout:    d.Timeout(schema.TimeoutDelete),
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf(
			"Error waiting for Stream Consumer (%s) to be destroyed: %s",
			cn, err)
	}

	return nil
}

type kinesisStreamConsumerState struct {
	arn               string
	streamArn         string
	creationTimestamp int64
	status            string
}

func readKinesisStreamConsumerState(conn *kinesis.Kinesis, cn string) (*kinesisStreamConsumerState, error) {
	describeOpts := &kinesis.DescribeStreamConsumerInput{
		ConsumerName: aws.String(cn),
	}

	state := &kinesisStreamConsumerState{}
	respDescribe, err := conn.DescribeStreamConsumer(describeOpts)

	state.arn = aws.StringValue(respDescribe.ConsumerDescription.ConsumerARN)
	state.streamArn = aws.StringValue(respDescribe.ConsumerDescription.StreamARN)
	state.creationTimestamp = aws.TimeValue(respDescribe.ConsumerDescription.ConsumerCreationTimestamp).Unix()
	state.status = aws.StringValue(respDescribe.ConsumerDescription.ConsumerStatus)

	return state, err
}

func streamConsumerStateRefreshFunc(conn *kinesis.Kinesis, cn string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		state, err := readKinesisStreamConsumerState(conn, cn)
		if err != nil {
			if awsErr, ok := err.(awserr.Error); ok {
				if awsErr.Code() == "ResourceNotFoundException" {
					return 42, "DESTROYED", nil
				}
				return nil, awsErr.Code(), err
			}
			return nil, "failed", err
		}

		return state, state.status, nil
	}
}
