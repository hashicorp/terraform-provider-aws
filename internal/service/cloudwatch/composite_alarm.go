// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudwatch

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_cloudwatch_composite_alarm", name="Composite Alarm")
// @Tags(identifierAttribute="arn")
func ResourceCompositeAlarm() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceCompositeAlarmCreate,
		ReadWithoutTimeout:   resourceCompositeAlarmRead,
		UpdateWithoutTimeout: resourceCompositeAlarmUpdate,
		DeleteWithoutTimeout: resourceCompositeAlarmDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"actions_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
				ForceNew: true,
			},
			"alarm_actions": {
				Type:     schema.TypeSet,
				Optional: true,
				Set:      schema.HashString,
				MaxItems: 5,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: verify.ValidARN,
				},
			},
			"alarm_description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 1024),
			},
			"alarm_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(0, 255),
			},
			"alarm_rule": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 10240),
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"insufficient_data_actions": {
				Type:     schema.TypeSet,
				Optional: true,
				Set:      schema.HashString,
				MaxItems: 5,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: verify.ValidARN,
				},
			},
			"ok_actions": {
				Type:     schema.TypeSet,
				Optional: true,
				Set:      schema.HashString,
				MaxItems: 5,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: verify.ValidARN,
				},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceCompositeAlarmCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).CloudWatchConn(ctx)

	name := d.Get("alarm_name").(string)
	input := expandPutCompositeAlarmInput(ctx, d)

	_, err := conn.PutCompositeAlarmWithContext(ctx, input)

	// Some partitions (e.g. ISO) may not support tag-on-create.
	if input.Tags != nil && errs.IsUnsupportedOperationInPartitionError(conn.PartitionID, err) {
		input.Tags = nil

		_, err = conn.PutCompositeAlarmWithContext(ctx, input)
	}

	if err != nil {
		return diag.Errorf("creating CloudWatch Composite Alarm (%s): %s", name, err)
	}

	d.SetId(name)

	// For partitions not supporting tag-on-create, attempt tag after create.
	if tags := getTagsIn(ctx); input.Tags == nil && len(tags) > 0 {
		alarm, err := FindCompositeAlarmByName(ctx, conn, d.Id())

		if err != nil {
			return diag.Errorf("reading CloudWatch Composite Alarm (%s): %s", d.Id(), err)
		}

		err = createTags(ctx, conn, aws.StringValue(alarm.AlarmArn), tags)

		// If default tags only, continue. Otherwise, error.
		if v, ok := d.GetOk(names.AttrTags); (!ok || len(v.(map[string]interface{})) == 0) && errs.IsUnsupportedOperationInPartitionError(conn.PartitionID, err) {
			return resourceCompositeAlarmRead(ctx, d, meta)
		}

		if err != nil {
			return diag.Errorf("setting CloudWatch Composite Alarm (%s) tags: %s", d.Id(), err)
		}
	}

	return resourceCompositeAlarmRead(ctx, d, meta)
}

func resourceCompositeAlarmRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).CloudWatchConn(ctx)

	alarm, err := FindCompositeAlarmByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] CloudWatch Composite Alarm %s not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading CloudWatch Composite Alarm (%s): %s", d.Id(), err)
	}

	d.Set("actions_enabled", alarm.ActionsEnabled)
	d.Set("alarm_actions", aws.StringValueSlice(alarm.AlarmActions))
	d.Set("alarm_description", alarm.AlarmDescription)
	d.Set("alarm_name", alarm.AlarmName)
	d.Set("alarm_rule", alarm.AlarmRule)
	d.Set("arn", alarm.AlarmArn)
	d.Set("insufficient_data_actions", aws.StringValueSlice(alarm.InsufficientDataActions))
	d.Set("ok_actions", aws.StringValueSlice(alarm.OKActions))

	return nil
}

func resourceCompositeAlarmUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).CloudWatchConn(ctx)

	if d.HasChangesExcept("tags", "tags_all") {
		input := expandPutCompositeAlarmInput(ctx, d)

		_, err := conn.PutCompositeAlarmWithContext(ctx, input)

		if err != nil {
			return diag.Errorf("updating CloudWatch Composite Alarm (%s): %s", d.Id(), err)
		}
	}

	return resourceCompositeAlarmRead(ctx, d, meta)
}

func resourceCompositeAlarmDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).CloudWatchConn(ctx)

	log.Printf("[INFO] Deleting CloudWatch Composite Alarm: %s", d.Id())
	_, err := conn.DeleteAlarmsWithContext(ctx, &cloudwatch.DeleteAlarmsInput{
		AlarmNames: aws.StringSlice([]string{d.Id()}),
	})

	if tfawserr.ErrCodeEquals(err, cloudwatch.ErrCodeResourceNotFound) {
		return nil
	}

	if err != nil {
		return diag.Errorf("deleting CloudWatch Composite Alarm (%s): %s", d.Id(), err)
	}

	return nil
}

func FindCompositeAlarmByName(ctx context.Context, conn *cloudwatch.CloudWatch, name string) (*cloudwatch.CompositeAlarm, error) {
	input := &cloudwatch.DescribeAlarmsInput{
		AlarmNames: aws.StringSlice([]string{name}),
		AlarmTypes: aws.StringSlice([]string{cloudwatch.AlarmTypeCompositeAlarm}),
	}

	output, err := conn.DescribeAlarmsWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, cloudwatch.ErrCodeResourceNotFound) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.CompositeAlarms) == 0 || output.CompositeAlarms[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output.CompositeAlarms); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output.CompositeAlarms[0], nil
}

func expandPutCompositeAlarmInput(ctx context.Context, d *schema.ResourceData) *cloudwatch.PutCompositeAlarmInput {
	apiObject := &cloudwatch.PutCompositeAlarmInput{
		ActionsEnabled: aws.Bool(d.Get("actions_enabled").(bool)),
		Tags:           getTagsIn(ctx),
	}

	if v, ok := d.GetOk("alarm_actions"); ok {
		apiObject.AlarmActions = flex.ExpandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("alarm_description"); ok {
		apiObject.AlarmDescription = aws.String(v.(string))
	}

	if v, ok := d.GetOk("alarm_name"); ok {
		apiObject.AlarmName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("alarm_rule"); ok {
		apiObject.AlarmRule = aws.String(v.(string))
	}

	if v, ok := d.GetOk("insufficient_data_actions"); ok {
		apiObject.InsufficientDataActions = flex.ExpandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("ok_actions"); ok {
		apiObject.OKActions = flex.ExpandStringSet(v.(*schema.Set))
	}

	return apiObject
}
