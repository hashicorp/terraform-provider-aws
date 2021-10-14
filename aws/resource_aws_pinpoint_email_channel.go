package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/pinpoint"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

func ResourceEmailChannel() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsPinpointEmailChannelUpsert,
		Read:   resourceEmailChannelRead,
		Update: resourceAwsPinpointEmailChannelUpsert,
		Delete: resourceEmailChannelDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"application_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"configuration_set": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"from_address": {
				Type:     schema.TypeString,
				Required: true,
			},
			"identity": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateArn,
			},
			"role_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateArn,
			},
			"messages_per_second": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func resourceAwsPinpointEmailChannelUpsert(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).PinpointConn

	applicationId := d.Get("application_id").(string)

	params := &pinpoint.EmailChannelRequest{}

	params.Enabled = aws.Bool(d.Get("enabled").(bool))
	params.FromAddress = aws.String(d.Get("from_address").(string))
	params.Identity = aws.String(d.Get("identity").(string))

	if v, ok := d.GetOk("role_arn"); ok {
		params.RoleArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("configuration_set"); ok {
		params.ConfigurationSet = aws.String(v.(string))
	}

	req := pinpoint.UpdateEmailChannelInput{
		ApplicationId:       aws.String(applicationId),
		EmailChannelRequest: params,
	}

	_, err := conn.UpdateEmailChannel(&req)
	if err != nil {
		return fmt.Errorf("error updating Pinpoint Email Channel for application %s: %w", applicationId, err)
	}

	d.SetId(applicationId)

	return resourceEmailChannelRead(d, meta)
}

func resourceEmailChannelRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).PinpointConn

	log.Printf("[INFO] Reading Pinpoint Email Channel for application %s", d.Id())

	output, err := conn.GetEmailChannel(&pinpoint.GetEmailChannelInput{
		ApplicationId: aws.String(d.Id()),
	})
	if err != nil {
		if tfawserr.ErrMessageContains(err, pinpoint.ErrCodeNotFoundException, "") {
			log.Printf("[WARN] Pinpoint Email Channel for application %s not found, error code (404)", d.Id())
			d.SetId("")
			return nil
		}

		return fmt.Errorf("error getting Pinpoint Email Channel for application %s: %w", d.Id(), err)
	}

	res := output.EmailChannelResponse
	d.Set("application_id", res.ApplicationId)
	d.Set("enabled", res.Enabled)
	d.Set("from_address", res.FromAddress)
	d.Set("identity", res.Identity)
	d.Set("role_arn", res.RoleArn)
	d.Set("configuration_set", res.ConfigurationSet)
	d.Set("messages_per_second", res.MessagesPerSecond)

	return nil
}

func resourceEmailChannelDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).PinpointConn

	log.Printf("[DEBUG] Deleting Pinpoint Email Channel for application %s", d.Id())
	_, err := conn.DeleteEmailChannel(&pinpoint.DeleteEmailChannelInput{
		ApplicationId: aws.String(d.Id()),
	})

	if tfawserr.ErrMessageContains(err, pinpoint.ErrCodeNotFoundException, "") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting Pinpoint Email Channel for application %s: %w", d.Id(), err)
	}
	return nil
}
