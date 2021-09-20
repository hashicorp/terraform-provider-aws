package chime

import (
	"context"
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/chime"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
)

func ResourceVoiceConnectorTermination() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceVoiceConnectorTerminationCreate,
		ReadWithoutTimeout:   resourceVoiceConnectorTerminationRead,
		UpdateWithoutTimeout: resourceVoiceConnectorTerminationUpdate,
		DeleteWithoutTimeout: resourceVoiceConnectorTerminationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"calling_regions": {
				Type:     schema.TypeSet,
				Required: true,
				MinItems: 1,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringLenBetween(2, 2),
				},
			},
			"cidr_allow_list": {
				Type:     schema.TypeSet,
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
				ValidateFunc: validation.IntAtLeast(1),
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

func resourceVoiceConnectorTerminationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ChimeConn

	vcId := d.Get("voice_connector_id").(string)

	input := &chime.PutVoiceConnectorTerminationInput{
		VoiceConnectorId: aws.String(vcId),
	}

	termination := &chime.Termination{
		CidrAllowedList: flex.ExpandStringSet(d.Get("cidr_allow_list").(*schema.Set)),
		CallingRegions:  flex.ExpandStringSet(d.Get("calling_regions").(*schema.Set)),
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

	d.SetId(vcId)

	return resourceVoiceConnectorTerminationRead(ctx, d, meta)
}

func resourceVoiceConnectorTerminationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ChimeConn

	input := &chime.GetVoiceConnectorTerminationInput{
		VoiceConnectorId: aws.String(d.Id()),
	}

	resp, err := conn.GetVoiceConnectorTerminationWithContext(ctx, input)

	if !d.IsNewResource() && tfawserr.ErrMessageContains(err, chime.ErrCodeNotFoundException, "") {
		log.Printf("[WARN] Chime Voice Connector (%s) termination not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("error getting Chime Voice Connector (%s) termination: %s", d.Id(), err)
	}

	if resp == nil || resp.Termination == nil {
		return diag.Errorf("error getting Chime Voice Connector (%s) termination: empty response", d.Id())
	}

	d.Set("cps_limit", resp.Termination.CpsLimit)
	d.Set("disabled", resp.Termination.Disabled)
	d.Set("default_phone_number", resp.Termination.DefaultPhoneNumber)

	if err := d.Set("calling_regions", flex.FlattenStringList(resp.Termination.CallingRegions)); err != nil {
		return diag.Errorf("error setting termination calling regions (%s): %s", d.Id(), err)
	}
	if err := d.Set("cidr_allow_list", flex.FlattenStringList(resp.Termination.CidrAllowedList)); err != nil {
		return diag.Errorf("error setting termination cidr allow list (%s): %s", d.Id(), err)
	}

	d.Set("voice_connector_id", d.Id())

	return nil
}

func resourceVoiceConnectorTerminationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ChimeConn

	if d.HasChanges("calling_regions", "cidr_allow_list", "disabled", "cps_limit", "default_phone_number") {
		termination := &chime.Termination{
			CallingRegions:  flex.ExpandStringSet(d.Get("calling_regions").(*schema.Set)),
			CidrAllowedList: flex.ExpandStringSet(d.Get("cidr_allow_list").(*schema.Set)),
			CpsLimit:        aws.Int64(int64(d.Get("cps_limit").(int))),
		}

		if v, ok := d.GetOk("default_phone_number"); ok {
			termination.DefaultPhoneNumber = aws.String(v.(string))
		}

		if v, ok := d.GetOk("disabled"); ok {
			termination.Disabled = aws.Bool(v.(bool))
		}

		input := &chime.PutVoiceConnectorTerminationInput{
			VoiceConnectorId: aws.String(d.Id()),
			Termination:      termination,
		}

		_, err := conn.PutVoiceConnectorTerminationWithContext(ctx, input)

		if err != nil {
			return diag.Errorf("error updating Chime Voice Connector (%s) termination: %s", d.Id(), err)
		}
	}

	return resourceVoiceConnectorTerminationRead(ctx, d, meta)
}

func resourceVoiceConnectorTerminationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ChimeConn

	input := &chime.DeleteVoiceConnectorTerminationInput{
		VoiceConnectorId: aws.String(d.Id()),
	}

	_, err := conn.DeleteVoiceConnectorTerminationWithContext(ctx, input)

	if tfawserr.ErrMessageContains(err, chime.ErrCodeNotFoundException, "") {
		return nil
	}

	if err != nil {
		return diag.Errorf("error deleting Chime Voice Connector termination (%s): %s", d.Id(), err)
	}

	return nil
}
