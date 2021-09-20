package pinpoint

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/pinpoint"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func ResourceBaiduChannel() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsPinpointBaiduChannelUpsert,
		Read:   resourceBaiduChannelRead,
		Update: resourceAwsPinpointBaiduChannelUpsert,
		Delete: resourceBaiduChannelDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"application_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"api_key": {
				Type:      schema.TypeString,
				Required:  true,
				Sensitive: true,
			},
			"secret_key": {
				Type:      schema.TypeString,
				Required:  true,
				Sensitive: true,
			},
		},
	}
}

func resourceAwsPinpointBaiduChannelUpsert(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).PinpointConn

	applicationId := d.Get("application_id").(string)

	params := &pinpoint.BaiduChannelRequest{}

	params.Enabled = aws.Bool(d.Get("enabled").(bool))
	params.ApiKey = aws.String(d.Get("api_key").(string))
	params.SecretKey = aws.String(d.Get("secret_key").(string))

	req := pinpoint.UpdateBaiduChannelInput{
		ApplicationId:       aws.String(applicationId),
		BaiduChannelRequest: params,
	}

	_, err := conn.UpdateBaiduChannel(&req)
	if err != nil {
		return fmt.Errorf("error updating Pinpoint Baidu Channel for application %s: %s", applicationId, err)
	}

	d.SetId(applicationId)

	return resourceBaiduChannelRead(d, meta)
}

func resourceBaiduChannelRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).PinpointConn

	log.Printf("[INFO] Reading Pinpoint Baidu Channel for application %s", d.Id())

	output, err := conn.GetBaiduChannel(&pinpoint.GetBaiduChannelInput{
		ApplicationId: aws.String(d.Id()),
	})
	if err != nil {
		if tfawserr.ErrMessageContains(err, pinpoint.ErrCodeNotFoundException, "") {
			log.Printf("[WARN] Pinpoint Baidu Channel for application %s not found, error code (404)", d.Id())
			d.SetId("")
			return nil
		}

		return fmt.Errorf("error getting Pinpoint Baidu Channel for application %s: %s", d.Id(), err)
	}

	d.Set("application_id", output.BaiduChannelResponse.ApplicationId)
	d.Set("enabled", output.BaiduChannelResponse.Enabled)
	// ApiKey and SecretKey are never returned

	return nil
}

func resourceBaiduChannelDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).PinpointConn

	log.Printf("[DEBUG] Deleting Pinpoint Baidu Channel for application %s", d.Id())
	_, err := conn.DeleteBaiduChannel(&pinpoint.DeleteBaiduChannelInput{
		ApplicationId: aws.String(d.Id()),
	})

	if tfawserr.ErrMessageContains(err, pinpoint.ErrCodeNotFoundException, "") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting Pinpoint Baidu Channel for application %s: %s", d.Id(), err)
	}
	return nil
}
