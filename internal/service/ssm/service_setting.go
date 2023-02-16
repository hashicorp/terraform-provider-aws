package ssm

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	ResNameServiceSetting = "Service Setting"
)

func ResourceServiceSetting() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceServiceSettingUpdate,
		ReadWithoutTimeout:   resourceServiceSettingRead,
		UpdateWithoutTimeout: resourceServiceSettingUpdate,
		DeleteWithoutTimeout: resourceServiceSettingReset,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"setting_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"setting_value": {
				Type:     schema.TypeString,
				Required: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceServiceSettingUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SSMConn()

	log.Printf("[DEBUG] SSM service setting create: %s", d.Get("setting_id").(string))

	updateServiceSettingInput := &ssm.UpdateServiceSettingInput{
		SettingId:    aws.String(d.Get("setting_id").(string)),
		SettingValue: aws.String(d.Get("setting_value").(string)),
	}

	if _, err := conn.UpdateServiceSettingWithContext(ctx, updateServiceSettingInput); err != nil {
		return create.DiagError(names.SSM, create.ErrActionUpdating, ResNameServiceSetting, d.Get("setting_id").(string), err)
	}

	d.SetId(d.Get("setting_id").(string))

	if _, err := waitServiceSettingUpdated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
		return create.DiagError(names.SSM, create.ErrActionWaitingForUpdate, ResNameServiceSetting, d.Id(), err)
	}

	return append(diags, resourceServiceSettingRead(ctx, d, meta)...)
}

func resourceServiceSettingRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SSMConn()

	log.Printf("[DEBUG] Reading SSM Activation: %s", d.Id())

	output, err := FindServiceSettingByID(ctx, conn, d.Id())
	if err != nil {
		return create.DiagError(names.SSM, create.ErrActionReading, ResNameServiceSetting, d.Id(), err)
	}

	// AWS SSM service setting API requires the entire ARN as input,
	// but setting_id in the output is only a part of ARN.
	d.Set("setting_id", output.ARN)
	d.Set("setting_value", output.SettingValue)
	d.Set("arn", output.ARN)
	d.Set("status", output.Status)

	return diags
}

func resourceServiceSettingReset(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SSMConn()

	log.Printf("[DEBUG] Deleting SSM Service Setting: %s", d.Id())

	resetServiceSettingInput := &ssm.ResetServiceSettingInput{
		SettingId: aws.String(d.Get("setting_id").(string)),
	}

	_, err := conn.ResetServiceSettingWithContext(ctx, resetServiceSettingInput)
	if err != nil {
		return create.DiagError(names.SSM, create.ErrActionDeleting, ResNameServiceSetting, d.Id(), err)
	}

	if err := waitServiceSettingReset(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return create.DiagError(names.SSM, create.ErrActionWaitingForDeletion, ResNameServiceSetting, d.Id(), err)
	}

	return diags
}
