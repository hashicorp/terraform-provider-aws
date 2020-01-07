package aws

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceAwsKinesisStreamConsumer() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsKinesisStreamConsumerCreate,
		Read:   resourceAwsKinesisStreamConsumerRead,
		Delete: resourceAwsKinesisStreamConsumerDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(120 * time.Minute),
		},

		Importer: &schema.ResourceImporter{
			State: func(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				importParts, err := validateResourceAwsKinesisStreamConsumerImportString(d.Id())
				if err != nil {
					return nil, err
				}
				d.Set("name", importParts[0])
				d.Set("stream_arn", importParts[1])
				return []*schema.ResourceData{d}, nil
			},
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
				ForceNew: true,
			},

			"arn": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
		},
	}
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
		Refresh:    streamConsumerStateRefreshFunc(conn, cn, sa),
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
	sa := d.Get("stream_arn").(string)

	state, err := readKinesisStreamConsumerState(conn, cn, sa)
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
	sa := d.Get("stream_arn").(string)

	log.Printf("[DEBUG] Deregister Stream Consumer: %s", cn)
	_, err := conn.DeregisterStreamConsumer(&kinesis.DeregisterStreamConsumerInput{
		ConsumerName: aws.String(cn),
		StreamARN:    aws.String(sa),
	})
	if err != nil {
		// Missing Stream Consumer or Stream (API error)
		if awsErr, ok := err.(awserr.Error); ok {
			if awsErr.Code() == "ResourceNotFoundException" {
				log.Printf("[WARN] No Stream Consumer found: %v", cn)
				return nil
			}
		}
		return err
	}

	stateConf := &resource.StateChangeConf{
		Pending:    []string{"DELETING"},
		Target:     []string{"DESTROYED"},
		Refresh:    streamConsumerStateRefreshFunc(conn, cn, sa),
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

func validateResourceAwsKinesisStreamConsumerImportString(importStr string) ([]string, error) {
	// example: my_consumer@arn:aws:kinesis:us-west-2:123456789012:stream/my-stream
	importParts := strings.Split(strings.ToLower(importStr), "_")
	errStr := "unexpected format of import string (%q), expected <consumer name>@<stream arn>: %s"
	if len(importParts) != 2 {
		return nil, fmt.Errorf(errStr, importStr, "invalid no. of parts")
	}

	consumerName := importParts[0]
	streamArn := importParts[1]

	consumerNameRe := regexp.MustCompile(`(^[a-zA-Z0-9_.-]+$)`)
	streamArnRe := regexp.MustCompile(`arn:aws.*:kinesis:.*:\d{12}:stream/.+`)

	if !consumerNameRe.MatchString(consumerName) {
		return nil, fmt.Errorf(errStr, importStr, "invalid consumer name")
	}

	if !streamArnRe.MatchString(streamArn) {
		return nil, fmt.Errorf(errStr, importStr, "invalid stream arn")
	}

	return importParts, nil
}

func readKinesisStreamConsumerState(conn *kinesis.Kinesis, cn string, sa string) (*kinesisStreamConsumerState, error) {
	input := &kinesis.DescribeStreamConsumerInput{
		ConsumerName: aws.String(cn),
		StreamARN:    aws.String(sa),
	}

	state := &kinesisStreamConsumerState{}
	response, err := conn.DescribeStreamConsumer(input)
	if err == nil {
		state.arn = aws.StringValue(response.ConsumerDescription.ConsumerARN)
		state.streamArn = aws.StringValue(response.ConsumerDescription.StreamARN)
		state.creationTimestamp = aws.TimeValue(response.ConsumerDescription.ConsumerCreationTimestamp).Unix()
		state.status = aws.StringValue(response.ConsumerDescription.ConsumerStatus)
	}

	return state, err
}

func streamConsumerStateRefreshFunc(conn *kinesis.Kinesis, cn string, sa string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		state, err := readKinesisStreamConsumerState(conn, cn, sa)
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
