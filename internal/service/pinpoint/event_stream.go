package pinpoint

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/pinpoint"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceEventStream() *schema.Resource {
	return &schema.Resource{
		Create: resourceEventStreamUpsert,
		Read:   resourceEventStreamRead,
		Update: resourceEventStreamUpsert,
		Delete: resourceEventStreamDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"application_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"destination_stream_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
			"role_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
		},
	}
}

func resourceEventStreamUpsert(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).PinpointConn

	applicationId := d.Get("application_id").(string)

	params := &pinpoint.WriteEventStream{
		DestinationStreamArn: aws.String(d.Get("destination_stream_arn").(string)),
		RoleArn:              aws.String(d.Get("role_arn").(string)),
	}

	req := pinpoint.PutEventStreamInput{
		ApplicationId:    aws.String(applicationId),
		WriteEventStream: params,
	}

	// Retry for IAM eventual consistency
	err := resource.Retry(propagationTimeout, func() *resource.RetryError {
		_, err := conn.PutEventStream(&req)

		if tfawserr.ErrMessageContains(err, pinpoint.ErrCodeBadRequestException, "make sure the IAM Role is configured correctly") {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		_, err = conn.PutEventStream(&req)
	}

	if err != nil {
		return fmt.Errorf("error putting Pinpoint Event Stream for application %s: %w", applicationId, err)
	}

	d.SetId(applicationId)

	return resourceEventStreamRead(d, meta)
}

func resourceEventStreamRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).PinpointConn

	log.Printf("[INFO] Reading Pinpoint Event Stream for application %s", d.Id())

	output, err := conn.GetEventStream(&pinpoint.GetEventStreamInput{
		ApplicationId: aws.String(d.Id()),
	})
	if err != nil {
		if tfawserr.ErrCodeEquals(err, pinpoint.ErrCodeNotFoundException) {
			log.Printf("[WARN] Pinpoint Event Stream for application %s not found, error code (404)", d.Id())
			d.SetId("")
			return nil
		}

		return fmt.Errorf("error getting Pinpoint Event Stream for application %s: %w", d.Id(), err)
	}

	res := output.EventStream
	d.Set("application_id", res.ApplicationId)
	d.Set("destination_stream_arn", res.DestinationStreamArn)
	d.Set("role_arn", res.RoleArn)

	return nil
}

func resourceEventStreamDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).PinpointConn

	log.Printf("[DEBUG] Pinpoint Delete Event Stream: %s", d.Id())
	_, err := conn.DeleteEventStream(&pinpoint.DeleteEventStreamInput{
		ApplicationId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, pinpoint.ErrCodeNotFoundException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting Pinpoint Event Stream for application %s: %w", d.Id(), err)
	}
	return nil
}
