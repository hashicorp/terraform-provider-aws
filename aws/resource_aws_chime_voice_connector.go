package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/chime"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceAwsChimeVoiceConnector() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsChimeVoiceConnectorCreate,
		Read:   resourceAwsChimeVoiceConnectorRead,
		Update: resourceAwsChimeVoiceConnectorUpdate,
		Delete: resourceAwsChimeVoiceConnectorDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"outbound_host_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			"aws_region": {
				Type:         schema.TypeString,
				ForceNew:     true,
				Optional:     true,
				Default:      chime.VoiceConnectorAwsRegionUsEast1,
				ValidateFunc: validation.StringInSlice([]string{chime.VoiceConnectorAwsRegionUsEast1, chime.VoiceConnectorAwsRegionUsWest2}, false),
			},
			"require_encryption": {
				Type:     schema.TypeBool,
				Required: true,
			},
		},
	}
}

func resourceAwsChimeVoiceConnectorCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).chimeconn

	createInput := &chime.CreateVoiceConnectorInput{}

	if v, ok := d.GetOk("name"); ok {
		createInput.Name = aws.String(v.(string))
	}

	if v, ok := d.GetOk("aws_region"); ok {
		createInput.AwsRegion = aws.String(v.(string))
	}

	if v, ok := d.GetOk("require_encryption"); ok {
		createInput.RequireEncryption = aws.Bool(v.(bool))
	}

	resp, err := conn.CreateVoiceConnector(createInput)
	if err != nil {
		return fmt.Errorf("Error creating Chime Voice connector: %s", err)
	}

	d.SetId(aws.StringValue(resp.VoiceConnector.VoiceConnectorId))

	return resourceAwsChimeVoiceConnectorRead(d, meta)
}

func resourceAwsChimeVoiceConnectorRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).chimeconn

	getInput := &chime.GetVoiceConnectorInput{
		VoiceConnectorId: aws.String(d.Id()),
	}

	resp, err := conn.GetVoiceConnector(getInput)
	if isAWSErr(err, chime.ErrCodeNotFoundException, "") {
		log.Printf("[WARN] Chime Voice connector %s not found", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("Error getting Voice connector (%s): %s", d.Id(), err)
	}

	d.Set("aws_region", resp.VoiceConnector.AwsRegion)
	d.Set("outbound_host_name", resp.VoiceConnector.OutboundHostName)
	d.Set("require_encryption", resp.VoiceConnector.RequireEncryption)
	d.Set("name", resp.VoiceConnector.Name)

	return nil
}

func resourceAwsChimeVoiceConnectorUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).chimeconn

	updateInput := &chime.UpdateVoiceConnectorInput{
		VoiceConnectorId: aws.String(d.Id()),
	}

	if v, ok := d.GetOk("name"); ok {
		updateInput.Name = aws.String(v.(string))
	}

	if v, ok := d.GetOk("require_encryption"); ok {
		updateInput.RequireEncryption = aws.Bool(v.(bool))
	}

	if _, err := conn.UpdateVoiceConnector(updateInput); err != nil {
		if isAWSErr(err, chime.ErrCodeNotFoundException, "") {
			log.Printf("[WARN] Chime Voice connector %s not found", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error updating Voice connector (%s): %s", d.Id(), err)
	}

	return resourceAwsChimeVoiceConnectorRead(d, meta)
}

func resourceAwsChimeVoiceConnectorDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).chimeconn

	input := &chime.DeleteVoiceConnectorInput{
		VoiceConnectorId: aws.String(d.Id()),
	}

	if _, err := conn.DeleteVoiceConnector(input); err != nil {
		return fmt.Errorf("Error deleting Voice connector (%s)", d.Id())
	}

	return nil
}
