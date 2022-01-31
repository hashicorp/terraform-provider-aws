package kinesis

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceStreamConsumer() *schema.Resource {
	return &schema.Resource{
		Create: resourceStreamConsumerCreate,
		Read:   resourceStreamConsumerRead,
		Delete: resourceStreamConsumerDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"creation_timestamp": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"stream_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
		},
	}
}

func resourceStreamConsumerCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).KinesisConn

	name := d.Get("name").(string)
	input := &kinesis.RegisterStreamConsumerInput{
		ConsumerName: aws.String(name),
		StreamARN:    aws.String(d.Get("stream_arn").(string)),
	}

	log.Printf("[DEBUG] Registering Kinesis Stream Consumer: %s", input)
	output, err := conn.RegisterStreamConsumer(input)

	if err != nil {
		return fmt.Errorf("error creating Kinesis Stream Consumer (%s): %w", name, err)
	}

	d.SetId(aws.StringValue(output.Consumer.ConsumerARN))

	if _, err := waitStreamConsumerCreated(conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for Kinesis Stream Consumer (%s) create: %w", d.Id(), err)
	}

	return resourceStreamConsumerRead(d, meta)
}

func resourceStreamConsumerRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).KinesisConn

	consumer, err := FindStreamConsumerByARN(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Kinesis Stream Consumer (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading Kinesis Stream Consumer (%s): %w", d.Id(), err)
	}

	d.Set("arn", consumer.ConsumerARN)
	d.Set("creation_timestamp", aws.TimeValue(consumer.ConsumerCreationTimestamp).Format(time.RFC3339))
	d.Set("name", consumer.ConsumerName)
	d.Set("stream_arn", consumer.StreamARN)

	return nil
}

func resourceStreamConsumerDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).KinesisConn

	log.Printf("[DEBUG] Deregistering Kinesis Stream Consumer: (%s)", d.Id())
	_, err := conn.DeregisterStreamConsumer(&kinesis.DeregisterStreamConsumerInput{
		ConsumerARN: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, kinesis.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting Kinesis Stream Consumer (%s): %w", d.Id(), err)
	}

	if _, err := waitStreamConsumerDeleted(conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for Kinesis Stream Consumer (%s) delete: %w", d.Id(), err)
	}

	return nil
}

func FindStreamConsumerByARN(conn *kinesis.Kinesis, arn string) (*kinesis.ConsumerDescription, error) {
	input := &kinesis.DescribeStreamConsumerInput{
		ConsumerARN: aws.String(arn),
	}

	output, err := conn.DescribeStreamConsumer(input)

	if tfawserr.ErrCodeEquals(err, kinesis.ErrCodeResourceNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.ConsumerDescription == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.ConsumerDescription, nil
}

func statusStreamConsumer(conn *kinesis.Kinesis, arn string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindStreamConsumerByARN(conn, arn)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.ConsumerStatus), nil
	}
}

const (
	streamConsumerCreatedTimeout = 5 * time.Minute
	streamConsumerDeletedTimeout = 5 * time.Minute
)

func waitStreamConsumerCreated(conn *kinesis.Kinesis, arn string) (*kinesis.ConsumerDescription, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{kinesis.ConsumerStatusCreating},
		Target:  []string{kinesis.ConsumerStatusActive},
		Refresh: statusStreamConsumer(conn, arn),
		Timeout: streamConsumerCreatedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*kinesis.ConsumerDescription); ok {
		return output, err
	}

	return nil, err
}

func waitStreamConsumerDeleted(conn *kinesis.Kinesis, arn string) (*kinesis.ConsumerDescription, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{kinesis.ConsumerStatusDeleting},
		Target:  []string{},
		Refresh: statusStreamConsumer(conn, arn),
		Timeout: streamConsumerDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*kinesis.ConsumerDescription); ok {
		return output, err
	}

	return nil, err
}
