package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/chime"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceAwsChimeVoiceConnectorTermination() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsChimeVoiceConnectorTerminationPut,
		Read:   resourceAwsChimeVoiceConnectorTerminationRead,
		Update: resourceAwsChimeVoiceConnectorTerminationUpdate,
		Delete: resourceAwsChimeVoiceConnectorTerminationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"voice_connector_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"calling_regions": {
				Type:     schema.TypeList,
				Required: true,
				MinItems: 1,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringLenBetween(2, 2),
				},
			},
			"cidr_allow_list": {
				Type:     schema.TypeList,
				Required: true,
				MinItems: 1,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.IsCIDRNetwork(27, 32),
				},
			},
			"disabled": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"cps_limit": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      1,
				ValidateFunc: validation.IntAtMost(1),
			},
			"default_phone_number": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func resourceAwsChimeVoiceConnectorTerminationPut(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).chimeconn

	vcId := d.Get("voice_connector_id").(string)

	input := &chime.PutVoiceConnectorTerminationInput{
		VoiceConnectorId: aws.String(vcId),
	}

	termination := &chime.Termination{
		CidrAllowedList: expandStringList(d.Get("cidr_allow_list").([]interface{})),
		CallingRegions:  expandStringList(d.Get("calling_regions").([]interface{})),
	}

	if v, ok := d.GetOk("disabled"); ok {
		termination.Disabled = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("cps_limit"); ok {
		termination.CpsLimit = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("default_phone_number"); ok {
		termination.DefaultPhoneNumber = aws.String(v.(string))
	}

	input.Termination = termination

	if _, err := conn.PutVoiceConnectorTermination(input); err != nil {
		return fmt.Errorf("Error creating Chime Voice connector termination: %s", vcId)
	}

	d.SetId(fmt.Sprintf("%s-termination", vcId))

	return resourceAwsChimeVoiceConnectorTerminationRead(d, meta)
}

func resourceAwsChimeVoiceConnectorTerminationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).chimeconn

	input := &chime.GetVoiceConnectorTerminationInput{
		VoiceConnectorId: aws.String(d.Get("voice_connector_id").(string)),
	}

	resp, err := conn.GetVoiceConnectorTermination(input)
	if isAWSErr(err, chime.ErrCodeNotFoundException, "") {
		log.Printf("error reading termination for Chime Voice connector")
		d.SetId("")
		return nil
	}

	if err := d.Set("calling_regions", flattenStringList(resp.Termination.CallingRegions)); err != nil {
		return fmt.Errorf("error setting calling regions in Chime Voice connector termination: (%s)", d.Id())
	}

	if err := d.Set("cidr_allow_list", flattenStringList(resp.Termination.CidrAllowedList)); err != nil {
		return fmt.Errorf("error setting cidr allow list in Chime Voice connector termination: (%s)", d.Id())
	}

	d.Set("cps_limit", resp.Termination.CpsLimit)
	d.Set("disabled", resp.Termination.Disabled)
	d.Set("default_phone_number", resp.Termination.DefaultPhoneNumber)

	return nil
}

func resourceAwsChimeVoiceConnectorTerminationUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).chimeconn

	termination := &chime.Termination{
		CallingRegions:  expandStringList(d.Get("calling_regions").([]interface{})),
		CidrAllowedList: expandStringList(d.Get("cidr_allow_list").([]interface{})),
	}

	if v, ok := d.GetOk("disabled"); ok {
		termination.Disabled = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("cps_limit"); ok {
		termination.CpsLimit = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("default_phone_number"); ok {
		termination.DefaultPhoneNumber = aws.String(v.(string))
	}

	input := &chime.PutVoiceConnectorTerminationInput{
		VoiceConnectorId: aws.String(d.Get("voice_connector_id").(string)),
		Termination:      termination,
	}

	if err := input.Validate(); err != nil {
		return fmt.Errorf("error validation Chime Voice connecetor termination input (%s) ", err)
	}

	if _, err := conn.PutVoiceConnectorTermination(input); err != nil {
		return fmt.Errorf("error updating Chime Voice connector termination: (%s), %s, (%v+)", d.Id(), err, input)
	}

	return resourceAwsChimeVoiceConnectorTerminationRead(d, meta)
}

func resourceAwsChimeVoiceConnectorTerminationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).chimeconn

	input := &chime.DeleteVoiceConnectorTerminationInput{
		VoiceConnectorId: aws.String(d.Get("voice_connector_id").(string)),
	}

	if _, err := conn.DeleteVoiceConnectorTermination(input); err != nil {
		return fmt.Errorf("error deleting Chime Voice connector termination (%s)", d.Id())
	}

	return nil
}
