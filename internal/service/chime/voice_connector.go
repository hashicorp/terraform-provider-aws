package chime

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/chime"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func ResourceVoiceConnector() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVoiceConnectorCreate,
		ReadContext:   resourceVoiceConnectorRead,
		UpdateContext: resourceVoiceConnectorUpdate,
		DeleteContext: resourceVoiceConnectorDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"aws_region": {
				Type:         schema.TypeString,
				ForceNew:     true,
				Optional:     true,
				Default:      chime.VoiceConnectorAwsRegionUsEast1,
				ValidateFunc: validation.StringInSlice([]string{chime.VoiceConnectorAwsRegionUsEast1, chime.VoiceConnectorAwsRegionUsWest2}, false),
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			"outbound_host_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"require_encryption": {
				Type:     schema.TypeBool,
				Required: true,
			},
		},
	}
}

func resourceVoiceConnectorCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ChimeConn

	createInput := &chime.CreateVoiceConnectorInput{
		Name:              aws.String(d.Get("name").(string)),
		RequireEncryption: aws.Bool(d.Get("require_encryption").(bool)),
	}

	if v, ok := d.GetOk("aws_region"); ok {
		createInput.AwsRegion = aws.String(v.(string))
	}

	resp, err := conn.CreateVoiceConnectorWithContext(ctx, createInput)
	if err != nil || resp.VoiceConnector == nil {
		return diag.Errorf("Error creating Chime Voice connector: %s", err)
	}

	d.SetId(aws.StringValue(resp.VoiceConnector.VoiceConnectorId))

	return resourceVoiceConnectorRead(ctx, d, meta)
}

func resourceVoiceConnectorRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ChimeConn

	getInput := &chime.GetVoiceConnectorInput{
		VoiceConnectorId: aws.String(d.Id()),
	}

	resp, err := conn.GetVoiceConnectorWithContext(ctx, getInput)
	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, chime.ErrCodeNotFoundException) {
		log.Printf("[WARN] Chime Voice connector %s not found", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil || resp.VoiceConnector == nil {
		return diag.Errorf("Error getting Voice connector (%s): %s", d.Id(), err)
	}

	d.Set("aws_region", resp.VoiceConnector.AwsRegion)
	d.Set("outbound_host_name", resp.VoiceConnector.OutboundHostName)
	d.Set("require_encryption", resp.VoiceConnector.RequireEncryption)
	d.Set("name", resp.VoiceConnector.Name)

	return nil
}

func resourceVoiceConnectorUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ChimeConn

	if d.HasChanges("name", "require_encryption") {
		updateInput := &chime.UpdateVoiceConnectorInput{
			VoiceConnectorId:  aws.String(d.Id()),
			Name:              aws.String(d.Get("name").(string)),
			RequireEncryption: aws.Bool(d.Get("require_encryption").(bool)),
		}

		if _, err := conn.UpdateVoiceConnectorWithContext(ctx, updateInput); err != nil {
			return diag.Errorf("Error updating Voice connector (%s): %s", d.Id(), err)
		}
	}
	return resourceVoiceConnectorRead(ctx, d, meta)
}

func resourceVoiceConnectorDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ChimeConn

	input := &chime.DeleteVoiceConnectorInput{
		VoiceConnectorId: aws.String(d.Id()),
	}

	if _, err := conn.DeleteVoiceConnectorWithContext(ctx, input); err != nil {
		if tfawserr.ErrCodeEquals(err, chime.ErrCodeNotFoundException) {
			log.Printf("[WARN] Chime Voice connector %s not found", d.Id())
			return nil
		}
		return diag.Errorf("Error deleting Voice connector (%s)", d.Id())
	}
	return nil
}
