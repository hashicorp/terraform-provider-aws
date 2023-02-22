package configservice

import (
	"context"
	"errors"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/configservice"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func ResourceConfigurationRecorderStatus() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceConfigurationRecorderStatusPut,
		ReadWithoutTimeout:   resourceConfigurationRecorderStatusRead,
		UpdateWithoutTimeout: resourceConfigurationRecorderStatusPut,
		DeleteWithoutTimeout: resourceConfigurationRecorderStatusDelete,

		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				d.Set("name", d.Id())
				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"is_enabled": {
				Type:     schema.TypeBool,
				Required: true,
			},
		},
	}
}

func resourceConfigurationRecorderStatusPut(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConfigServiceConn()

	name := d.Get("name").(string)
	d.SetId(name)

	if d.HasChange("is_enabled") {
		isEnabled := d.Get("is_enabled").(bool)
		if isEnabled {
			log.Printf("[DEBUG] Starting AWSConfig Configuration recorder %q", name)
			startInput := configservice.StartConfigurationRecorderInput{
				ConfigurationRecorderName: aws.String(name),
			}
			_, err := conn.StartConfigurationRecorderWithContext(ctx, &startInput)
			if err != nil {
				return sdkdiag.AppendErrorf(diags, "Failed to start Configuration Recorder: %s", err)
			}
		} else {
			log.Printf("[DEBUG] Stopping AWSConfig Configuration recorder %q", name)
			stopInput := configservice.StopConfigurationRecorderInput{
				ConfigurationRecorderName: aws.String(name),
			}
			_, err := conn.StopConfigurationRecorderWithContext(ctx, &stopInput)
			if err != nil {
				return sdkdiag.AppendErrorf(diags, "Failed to stop Configuration Recorder: %s", err)
			}
		}
	}

	return append(diags, resourceConfigurationRecorderStatusRead(ctx, d, meta)...)
}

func resourceConfigurationRecorderStatusRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConfigServiceConn()

	name := d.Id()
	statusInput := configservice.DescribeConfigurationRecorderStatusInput{
		ConfigurationRecorderNames: []*string{aws.String(name)},
	}
	statusOut, err := conn.DescribeConfigurationRecorderStatusWithContext(ctx, &statusInput)
	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, configservice.ErrCodeNoSuchConfigurationRecorderException) {
		create.LogNotFoundRemoveState(names.ConfigService, create.ErrActionReading, ResNameConfigurationRecorderStatus, d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return create.DiagError(names.ConfigService, create.ErrActionReading, ResNameConfigurationRecorderStatus, d.Id(), err)
	}

	numberOfStatuses := len(statusOut.ConfigurationRecordersStatus)
	if !d.IsNewResource() && numberOfStatuses < 1 {
		create.LogNotFoundRemoveState(names.ConfigService, create.ErrActionReading, ResNameConfigurationRecorderStatus, d.Id())
		d.SetId("")
		return diags
	}

	if d.IsNewResource() && numberOfStatuses < 1 {
		return create.DiagError(names.ConfigService, create.ErrActionReading, ResNameConfigurationRecorderStatus, d.Id(), errors.New("not found after creation"))
	}

	if numberOfStatuses > 1 {
		return sdkdiag.AppendErrorf(diags, "Expected exactly 1 Configuration Recorder (status), received %d: %#v",
			numberOfStatuses, statusOut.ConfigurationRecordersStatus)
	}

	d.Set("is_enabled", statusOut.ConfigurationRecordersStatus[0].Recording)

	return diags
}

func resourceConfigurationRecorderStatusDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConfigServiceConn()
	input := configservice.StopConfigurationRecorderInput{
		ConfigurationRecorderName: aws.String(d.Get("name").(string)),
	}
	_, err := conn.StopConfigurationRecorderWithContext(ctx, &input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "Stopping Configuration Recorder failed: %s", err)
	}

	return diags
}
