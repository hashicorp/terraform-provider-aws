package aws

import (
	"context"
	"fmt"
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/chime"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceAwsChimeVoiceConnectorTermination() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceAwsChimeVoiceConnectorTerminationPut,
		ReadContext:   resourceAwsChimeVoiceConnectorTerminationRead,
		UpdateContext: resourceAwsChimeVoiceConnectorTerminationUpdate,
		DeleteContext: resourceAwsChimeVoiceConnectorTerminationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
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
			"cps_limit": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      1,
				ValidateFunc: validation.IntAtMost(1),
			},
			"default_phone_number": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^\+?[1-9]\d{1,14}$`), "must match ^\\+?[1-9]\\d{1,14}$"),
			},
			"disabled": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"voice_connector_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceAwsChimeVoiceConnectorTerminationPut(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

	if _, err := conn.PutVoiceConnectorTerminationWithContext(ctx, input); err != nil {
		return diag.Errorf("error creating Chime Voice Connector (%s) termination: %s", vcId, err)
	}

	d.SetId(fmt.Sprintf("termination-%s", vcId))

	return resourceAwsChimeVoiceConnectorTerminationRead(ctx, d, meta)
}

func resourceAwsChimeVoiceConnectorTerminationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).chimeconn

	vcId := d.Get("voice_connector_id").(string)
	input := &chime.GetVoiceConnectorTerminationInput{
		VoiceConnectorId: aws.String(vcId),
	}

	resp, err := conn.GetVoiceConnectorTerminationWithContext(ctx, input)
	if !d.IsNewResource() && isAWSErr(err, chime.ErrCodeNotFoundException, "") {
		log.Printf("[WARN] error getting Chime Voice Connector (%s) termination: %s", vcId, err)
		d.SetId("")
		return nil
	}

	if err != nil || resp.Termination == nil {
		return diag.Errorf("error getting Chime Voice Connector (%s) termination: %s", vcId, err)
	}

	d.Set("cps_limit", resp.Termination.CpsLimit)
	d.Set("disabled", resp.Termination.Disabled)
	d.Set("default_phone_number", resp.Termination.DefaultPhoneNumber)

	if err := d.Set("calling_regions", flattenStringList(resp.Termination.CallingRegions)); err != nil {
		return diag.Errorf("error setting termination calling regions (%s): %s", vcId, err)
	}
	if err := d.Set("cidr_allow_list", flattenStringList(resp.Termination.CidrAllowedList)); err != nil {
		return diag.Errorf("error setting termination cidr allow list (%s): %s", vcId, err)
	}

	return nil
}

func resourceAwsChimeVoiceConnectorTerminationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).chimeconn

	if d.HasChanges("calling_regions", "cidr_allow_list", "disabled", "cps_limit", "default_phone_number") {
		vcId := d.Get("voice_connector_id").(string)
		termination := &chime.Termination{
			CallingRegions:  expandStringList(d.Get("calling_regions").([]interface{})),
			CidrAllowedList: expandStringList(d.Get("cidr_allow_list").([]interface{})),
			CpsLimit:        aws.Int64(int64(d.Get("cps_limit").(int))),
		}

		if v, ok := d.GetOk("default_phone_number"); ok {
			termination.DefaultPhoneNumber = aws.String(v.(string))
		}

		if v, ok := d.GetOk("disabled"); ok {
			termination.Disabled = aws.Bool(v.(bool))
		}

		input := &chime.PutVoiceConnectorTerminationInput{
			VoiceConnectorId: aws.String(vcId),
			Termination:      termination,
		}

		if _, err := conn.PutVoiceConnectorTerminationWithContext(ctx, input); err != nil {
			if isAWSErr(err, chime.ErrCodeNotFoundException, "") {
				log.Printf("[WARN] error getting Chime Voice Connector (%s) termination: %s", vcId, err)
				d.SetId("")
				return nil
			}

			return diag.Errorf("error updating Chime Voice Connector (%s) termination: %s", vcId, err)
		}
	}

	return resourceAwsChimeVoiceConnectorTerminationRead(ctx, d, meta)
}

func resourceAwsChimeVoiceConnectorTerminationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).chimeconn

	vcId := d.Get("voice_connector_id").(string)
	input := &chime.DeleteVoiceConnectorTerminationInput{
		VoiceConnectorId: aws.String(vcId),
	}

	if _, err := conn.DeleteVoiceConnectorTerminationWithContext(ctx, input); err != nil {
		return diag.Errorf("error deleting Chime Voice Connector (%s) termination (%s): %s", vcId, d.Id(), err)
	}

	return nil
}
