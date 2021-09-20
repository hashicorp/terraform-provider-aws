package aws

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/chime"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func resourceAwsChimeVoiceConnectorOrigination() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceAwsChimeVoiceConnectorOriginationCreate,
		ReadWithoutTimeout:   resourceAwsChimeVoiceConnectorOriginationRead,
		UpdateWithoutTimeout: resourceAwsChimeVoiceConnectorOriginationUpdate,
		DeleteWithoutTimeout: resourceAwsChimeVoiceConnectorOriginationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"disabled": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"route": {
				Type:     schema.TypeSet,
				Required: true,
				MinItems: 1,
				MaxItems: 20,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"host": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.IsIPAddress,
						},
						"port": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      5060,
							ValidateFunc: validation.IsPortNumber,
						},
						"priority": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntBetween(1, 99),
						},
						"protocol": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(chime.OriginationRouteProtocol_Values(), false),
						},
						"weight": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntBetween(1, 99),
						},
					},
				},
			},
			"voice_connector_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceAwsChimeVoiceConnectorOriginationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ChimeConn

	vcId := d.Get("voice_connector_id").(string)

	input := &chime.PutVoiceConnectorOriginationInput{
		VoiceConnectorId: aws.String(vcId),
		Origination: &chime.Origination{
			Routes: expandOriginationRoutes(d.Get("route").(*schema.Set).List()),
		},
	}

	if v, ok := d.GetOk("disabled"); ok {
		input.Origination.Disabled = aws.Bool(v.(bool))
	}

	if _, err := conn.PutVoiceConnectorOriginationWithContext(ctx, input); err != nil {
		return diag.Errorf("error creating Chime Voice Connector (%s) origination: %s", vcId, err)
	}

	d.SetId(vcId)

	return resourceAwsChimeVoiceConnectorOriginationRead(ctx, d, meta)
}

func resourceAwsChimeVoiceConnectorOriginationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ChimeConn

	input := &chime.GetVoiceConnectorOriginationInput{
		VoiceConnectorId: aws.String(d.Id()),
	}

	resp, err := conn.GetVoiceConnectorOriginationWithContext(ctx, input)

	if !d.IsNewResource() && tfawserr.ErrMessageContains(err, chime.ErrCodeNotFoundException, "") {
		log.Printf("[WARN] Chime Voice Connector (%s) origination not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("error getting Chime Voice Connector (%s) origination: %s", d.Id(), err)
	}

	if resp == nil || resp.Origination == nil {
		return diag.Errorf("error getting Chime Voice Connector (%s) origination: empty response", d.Id())
	}

	d.Set("disabled", resp.Origination.Disabled)
	d.Set("voice_connector_id", d.Id())

	if err := d.Set("route", flattenOriginationRoutes(resp.Origination.Routes)); err != nil {
		return diag.Errorf("error setting Chime Voice Connector (%s) origination routes: %s", d.Id(), err)
	}

	return nil
}

func resourceAwsChimeVoiceConnectorOriginationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ChimeConn

	if d.HasChanges("route", "disabled") {
		input := &chime.PutVoiceConnectorOriginationInput{
			VoiceConnectorId: aws.String(d.Id()),
			Origination: &chime.Origination{
				Routes: expandOriginationRoutes(d.Get("route").(*schema.Set).List()),
			},
		}

		if v, ok := d.GetOk("disabled"); ok {
			input.Origination.Disabled = aws.Bool(v.(bool))
		}

		_, err := conn.PutVoiceConnectorOriginationWithContext(ctx, input)

		if err != nil {
			return diag.Errorf("error updating Chime Voice Connector (%s) origination: %s", d.Id(), err)
		}
	}

	return resourceAwsChimeVoiceConnectorOriginationRead(ctx, d, meta)
}

func resourceAwsChimeVoiceConnectorOriginationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ChimeConn

	input := &chime.DeleteVoiceConnectorOriginationInput{
		VoiceConnectorId: aws.String(d.Id()),
	}

	_, err := conn.DeleteVoiceConnectorOriginationWithContext(ctx, input)

	if tfawserr.ErrMessageContains(err, chime.ErrCodeNotFoundException, "") {
		return nil
	}

	if err != nil {
		return diag.Errorf("error deleting Chime Voice Connector (%s) origination: %s", d.Id(), err)
	}

	return nil
}

func expandOriginationRoutes(data []interface{}) []*chime.OriginationRoute {
	var originationRoutes []*chime.OriginationRoute

	for _, rItem := range data {
		item := rItem.(map[string]interface{})
		originationRoutes = append(originationRoutes, &chime.OriginationRoute{
			Host:     aws.String(item["host"].(string)),
			Port:     aws.Int64(int64(item["port"].(int))),
			Priority: aws.Int64(int64(item["priority"].(int))),
			Protocol: aws.String(item["protocol"].(string)),
			Weight:   aws.Int64(int64(item["weight"].(int))),
		})
	}

	return originationRoutes
}

func flattenOriginationRoutes(routes []*chime.OriginationRoute) []interface{} {
	var rawRoutes []interface{}

	for _, route := range routes {
		r := map[string]interface{}{
			"host":     aws.StringValue(route.Host),
			"port":     aws.Int64Value(route.Port),
			"priority": aws.Int64Value(route.Priority),
			"protocol": aws.StringValue(route.Protocol),
			"weight":   aws.Int64Value(route.Weight),
		}

		rawRoutes = append(rawRoutes, r)
	}

	return rawRoutes
}
