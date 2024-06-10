// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssm

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_ssm_maintenance_window", name="Maintenance Window")
// @Tags(identifierAttribute="id", resourceType="MaintenanceWindow")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/ssm;ssm.GetMaintenanceWindowOutput")
func resourceMaintenanceWindow() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceMaintenanceWindowCreate,
		ReadWithoutTimeout:   resourceMaintenanceWindowRead,
		UpdateWithoutTimeout: resourceMaintenanceWindowUpdate,
		DeleteWithoutTimeout: resourceMaintenanceWindowDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"allow_unassociated_targets": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"cutoff": {
				Type:     schema.TypeInt,
				Required: true,
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Optional: true,
			},
			names.AttrDuration: {
				Type:     schema.TypeInt,
				Required: true,
			},
			names.AttrEnabled: {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"end_date": {
				Type:     schema.TypeString,
				Optional: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
			},
			names.AttrSchedule: {
				Type:     schema.TypeString,
				Required: true,
			},
			"schedule_offset": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(1, 6),
			},
			"schedule_timezone": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"start_date": {
				Type:     schema.TypeString,
				Optional: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceMaintenanceWindowCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SSMClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &ssm.CreateMaintenanceWindowInput{
		AllowUnassociatedTargets: d.Get("allow_unassociated_targets").(bool),
		Cutoff:                   int32(d.Get("cutoff").(int)),
		Duration:                 aws.Int32(int32(d.Get(names.AttrDuration).(int))),
		Name:                     aws.String(name),
		Schedule:                 aws.String(d.Get(names.AttrSchedule).(string)),
		Tags:                     getTagsIn(ctx),
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("end_date"); ok {
		input.EndDate = aws.String(v.(string))
	}

	if v, ok := d.GetOk("schedule_offset"); ok {
		input.ScheduleOffset = aws.Int32(int32(v.(int)))
	}

	if v, ok := d.GetOk("schedule_timezone"); ok {
		input.ScheduleTimezone = aws.String(v.(string))
	}

	if v, ok := d.GetOk("start_date"); ok {
		input.StartDate = aws.String(v.(string))
	}

	output, err := conn.CreateMaintenanceWindow(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating SSM Maintenance Window (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.WindowId))

	if !d.Get(names.AttrEnabled).(bool) {
		input := &ssm.UpdateMaintenanceWindowInput{
			Enabled:  aws.Bool(false),
			WindowId: aws.String(d.Id()),
		}

		_, err := conn.UpdateMaintenanceWindow(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "disabling SSM Maintenance Window (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceMaintenanceWindowRead(ctx, d, meta)...)
}

func resourceMaintenanceWindowRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SSMClient(ctx)

	output, err := findMaintenanceWindowByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] SSM Maintenance Window %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SSM Maintenance Window (%s): %s", d.Id(), err)
	}

	d.Set("allow_unassociated_targets", output.AllowUnassociatedTargets)
	d.Set("cutoff", output.Cutoff)
	d.Set(names.AttrDescription, output.Description)
	d.Set(names.AttrDuration, output.Duration)
	d.Set(names.AttrEnabled, output.Enabled)
	d.Set("end_date", output.EndDate)
	d.Set(names.AttrName, output.Name)
	d.Set(names.AttrSchedule, output.Schedule)
	d.Set("schedule_offset", output.ScheduleOffset)
	d.Set("schedule_timezone", output.ScheduleTimezone)
	d.Set("start_date", output.StartDate)

	return diags
}

func resourceMaintenanceWindowUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SSMClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		// Replace must be set otherwise its not possible to remove optional attributes, e.g.
		// ValidationException: 1 validation error detected: Value '' at 'startDate' failed to satisfy constraint: Member must have length greater than or equal to 1
		input := &ssm.UpdateMaintenanceWindowInput{
			AllowUnassociatedTargets: aws.Bool(d.Get("allow_unassociated_targets").(bool)),
			Cutoff:                   aws.Int32(int32(d.Get("cutoff").(int))),
			Duration:                 aws.Int32(int32(d.Get(names.AttrDuration).(int))),
			Enabled:                  aws.Bool(d.Get(names.AttrEnabled).(bool)),
			Name:                     aws.String(d.Get(names.AttrName).(string)),
			Replace:                  aws.Bool(true),
			Schedule:                 aws.String(d.Get(names.AttrSchedule).(string)),
			WindowId:                 aws.String(d.Id()),
		}

		if v, ok := d.GetOk(names.AttrDescription); ok {
			input.Description = aws.String(v.(string))
		}

		if v, ok := d.GetOk("end_date"); ok {
			input.EndDate = aws.String(v.(string))
		}

		if v, ok := d.GetOk("schedule_offset"); ok {
			input.ScheduleOffset = aws.Int32(int32(v.(int)))
		}

		if v, ok := d.GetOk("schedule_timezone"); ok {
			input.ScheduleTimezone = aws.String(v.(string))
		}

		if v, ok := d.GetOk("start_date"); ok {
			input.StartDate = aws.String(v.(string))
		}

		_, err := conn.UpdateMaintenanceWindow(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating SSM Maintenance Window (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceMaintenanceWindowRead(ctx, d, meta)...)
}

func resourceMaintenanceWindowDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SSMClient(ctx)

	log.Printf("[INFO] Deleting SSM Maintenance Window: %s", d.Id())
	_, err := conn.DeleteMaintenanceWindow(ctx, &ssm.DeleteMaintenanceWindowInput{
		WindowId: aws.String(d.Id()),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting SSM Maintenance Window (%s): %s", d.Id(), err)
	}

	return diags
}

func findMaintenanceWindowByID(ctx context.Context, conn *ssm.Client, id string) (*ssm.GetMaintenanceWindowOutput, error) {
	input := &ssm.GetMaintenanceWindowInput{
		WindowId: aws.String(id),
	}

	output, err := conn.GetMaintenanceWindow(ctx, input)

	if errs.IsA[*awstypes.DoesNotExistException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}
