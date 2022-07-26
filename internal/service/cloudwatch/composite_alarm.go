package cloudwatch

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceCompositeAlarm() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceCompositeAlarmCreate,
		ReadContext:   resourceCompositeAlarmRead,
		UpdateContext: resourceCompositeAlarmUpdate,
		DeleteContext: resourceCompositeAlarmDelete,

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
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceCompositeAlarmCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).CloudWatchConn
	name := d.Get("alarm_name").(string)

	input := expandPutCompositeAlarmInput(d, meta)

	_, err := conn.PutCompositeAlarmWithContext(ctx, &input)

	// Some partitions (i.e., ISO) may not support tag-on-create
	if input.Tags != nil && verify.CheckISOErrorTagsUnsupported(conn.PartitionID, err) {
		log.Printf("[WARN] failed creating CloudWatch Composite Alarm (%s) with tags: %s. Trying create without tags.", name, err)
		input.Tags = nil

		_, err = conn.PutCompositeAlarmWithContext(ctx, &input)
	}

	if err != nil {
		return diag.Errorf("failed creating CloudWatch Composite Alarm (%s): %s", name, err)
	}

	d.SetId(name)

	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	// Some partitions (i.e., ISO) may not support tag-on-create, attempt tag after create
	if input.Tags == nil && len(tags) > 0 {
		alarm, err := FindCompositeAlarmByName(ctx, conn, name)

		if err != nil {
			return diag.Errorf("error reading CloudWatch Composite Alarm (%s): %s", name, err)
		}

		err = UpdateTags(conn, aws.StringValue(alarm.AlarmArn), nil, tags)

		// If default tags only, log and continue. Otherwise, error.
		if v, ok := d.GetOk("tags"); (!ok || len(v.(map[string]interface{})) == 0) && verify.CheckISOErrorTagsUnsupported(conn.PartitionID, err) {
			log.Printf("[WARN] failed adding tags after create for CloudWatch Composite Alarm (%s): %s", d.Id(), err)
			return resourceCompositeAlarmRead(ctx, d, meta)
		}

		if err != nil {
			return diag.Errorf("failed adding tags after create for CloudWatch Composite Alarm (%s): %s", d.Id(), err)
		}
	}

	return resourceCompositeAlarmRead(ctx, d, meta)
}

func resourceCompositeAlarmRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).CloudWatchConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig
	name := d.Id()

	alarm, err := FindCompositeAlarmByName(ctx, conn, name)
	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, cloudwatch.ErrCodeResourceNotFound) {
		log.Printf("[WARN] CloudWatch Composite Alarm %s not found, removing from state", name)
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("error reading CloudWatch Composite Alarm (%s): %s", name, err)
	}

	if alarm == nil {
		if d.IsNewResource() {
			return diag.Errorf("error reading CloudWatch Composite Alarm (%s): not found", name)
		}

		log.Printf("[WARN] CloudWatch Composite Alarm %s not found, removing from state", name)
		d.SetId("")
		return nil
	}

	d.Set("actions_enabled", alarm.ActionsEnabled)

	if err := d.Set("alarm_actions", flex.FlattenStringSet(alarm.AlarmActions)); err != nil {
		return diag.Errorf("error setting alarm_actions: %s", err)
	}

	d.Set("alarm_description", alarm.AlarmDescription)
	d.Set("alarm_name", alarm.AlarmName)
	d.Set("alarm_rule", alarm.AlarmRule)
	d.Set("arn", alarm.AlarmArn)

	if err := d.Set("insufficient_data_actions", flex.FlattenStringSet(alarm.InsufficientDataActions)); err != nil {
		return diag.Errorf("error setting insufficient_data_actions: %s", err)
	}

	if err := d.Set("ok_actions", flex.FlattenStringSet(alarm.OKActions)); err != nil {
		return diag.Errorf("error setting ok_actions: %s", err)
	}

	tags, err := ListTags(conn, aws.StringValue(alarm.AlarmArn))

	// Some partitions (i.e., ISO) may not support tagging, giving error
	if verify.CheckISOErrorTagsUnsupported(conn.PartitionID, err) {
		log.Printf("[WARN] failed listing tags for CloudWatch Composite Alarm (%s): %s", d.Id(), err)
		return nil
	}

	if err != nil {
		return diag.Errorf("failed listing tags for CloudWatch Composite Alarm (%s): %s", d.Id(), err)
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return diag.FromErr(fmt.Errorf("error setting tags: %w", err))
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return diag.FromErr(fmt.Errorf("error setting tags_all: %w", err))
	}

	return nil
}

func resourceCompositeAlarmUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).CloudWatchConn
	name := d.Id()

	input := expandPutCompositeAlarmInput(d, meta)

	_, err := conn.PutCompositeAlarmWithContext(ctx, &input)
	if err != nil {
		return diag.Errorf("error updating CloudWatch Composite Alarm (%s): %s", name, err)
	}

	arn := d.Get("arn").(string)
	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		err := UpdateTags(conn, arn, o, n)

		// Some partitions (i.e., ISO) may not support tagging, giving error
		if verify.CheckISOErrorTagsUnsupported(conn.PartitionID, err) {
			log.Printf("[WARN] failed updating tags for CloudWatch Composite Alarm (%s): %s", d.Id(), err)
			return resourceCompositeAlarmRead(ctx, d, meta)
		}

		if err != nil {
			return diag.Errorf("failed updating tags for CloudWatch Composite Alarm (%s): %s", d.Id(), err)
		}
	}

	return resourceCompositeAlarmRead(ctx, d, meta)
}

func resourceCompositeAlarmDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).CloudWatchConn
	name := d.Id()

	input := cloudwatch.DeleteAlarmsInput{
		AlarmNames: aws.StringSlice([]string{name}),
	}

	_, err := conn.DeleteAlarmsWithContext(ctx, &input)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, cloudwatch.ErrCodeResourceNotFound) {
			return nil
		}
		return diag.Errorf("error deleting CloudWatch Composite Alarm (%s): %s", name, err)
	}

	return nil
}

func expandPutCompositeAlarmInput(d *schema.ResourceData, meta interface{}) cloudwatch.PutCompositeAlarmInput {
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	out := cloudwatch.PutCompositeAlarmInput{
		ActionsEnabled: aws.Bool(d.Get("actions_enabled").(bool)),
	}

	if v, ok := d.GetOk("alarm_actions"); ok {
		out.AlarmActions = flex.ExpandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("alarm_description"); ok {
		out.AlarmDescription = aws.String(v.(string))
	}

	if v, ok := d.GetOk("alarm_name"); ok {
		out.AlarmName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("alarm_rule"); ok {
		out.AlarmRule = aws.String(v.(string))
	}

	if v, ok := d.GetOk("insufficient_data_actions"); ok {
		out.InsufficientDataActions = flex.ExpandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("ok_actions"); ok {
		out.OKActions = flex.ExpandStringSet(v.(*schema.Set))
	}

	if len(tags) > 0 {
		out.Tags = Tags(tags.IgnoreAWS())
	}

	return out
}
