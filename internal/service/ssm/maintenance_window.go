package ssm

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceMaintenanceWindow() *schema.Resource {
	return &schema.Resource{
		Create: resourceMaintenanceWindowCreate,
		Read:   resourceMaintenanceWindowRead,
		Update: resourceMaintenanceWindowUpdate,
		Delete: resourceMaintenanceWindowDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},

			"schedule": {
				Type:     schema.TypeString,
				Required: true,
			},

			"duration": {
				Type:     schema.TypeInt,
				Required: true,
			},

			"cutoff": {
				Type:     schema.TypeInt,
				Required: true,
			},

			"allow_unassociated_targets": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			"enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},

			"end_date": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"schedule_timezone": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"schedule_offset": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(1, 6),
			},

			"start_date": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceMaintenanceWindowCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SSMConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	params := &ssm.CreateMaintenanceWindowInput{
		AllowUnassociatedTargets: aws.Bool(d.Get("allow_unassociated_targets").(bool)),
		Cutoff:                   aws.Int64(int64(d.Get("cutoff").(int))),
		Duration:                 aws.Int64(int64(d.Get("duration").(int))),
		Name:                     aws.String(d.Get("name").(string)),
		Schedule:                 aws.String(d.Get("schedule").(string)),
	}

	if len(tags) > 0 {
		params.Tags = Tags(tags.IgnoreAWS())
	}

	if v, ok := d.GetOk("end_date"); ok {
		params.EndDate = aws.String(v.(string))
	}

	if v, ok := d.GetOk("schedule_timezone"); ok {
		params.ScheduleTimezone = aws.String(v.(string))
	}

	if v, ok := d.GetOk("schedule_offset"); ok {
		params.ScheduleOffset = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("start_date"); ok {
		params.StartDate = aws.String(v.(string))
	}

	if v, ok := d.GetOk("description"); ok {
		params.Description = aws.String(v.(string))
	}

	resp, err := conn.CreateMaintenanceWindow(params)
	if err != nil {
		return fmt.Errorf("error creating SSM Maintenance Window: %s", err)
	}

	d.SetId(aws.StringValue(resp.WindowId))

	if !d.Get("enabled").(bool) {
		input := &ssm.UpdateMaintenanceWindowInput{
			Enabled:  aws.Bool(false),
			WindowId: aws.String(d.Id()),
		}

		_, err := conn.UpdateMaintenanceWindow(input)
		if err != nil {
			return fmt.Errorf("error disabling SSM Maintenance Window (%s): %s", d.Id(), err)
		}
	}

	return resourceMaintenanceWindowRead(d, meta)
}

func resourceMaintenanceWindowUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SSMConn

	// Replace must be set otherwise its not possible to remove optional attributes, e.g.
	// ValidationException: 1 validation error detected: Value '' at 'startDate' failed to satisfy constraint: Member must have length greater than or equal to 1
	params := &ssm.UpdateMaintenanceWindowInput{
		AllowUnassociatedTargets: aws.Bool(d.Get("allow_unassociated_targets").(bool)),
		Cutoff:                   aws.Int64(int64(d.Get("cutoff").(int))),
		Duration:                 aws.Int64(int64(d.Get("duration").(int))),
		Enabled:                  aws.Bool(d.Get("enabled").(bool)),
		Name:                     aws.String(d.Get("name").(string)),
		Replace:                  aws.Bool(true),
		Schedule:                 aws.String(d.Get("schedule").(string)),
		WindowId:                 aws.String(d.Id()),
	}

	if v, ok := d.GetOk("end_date"); ok {
		params.EndDate = aws.String(v.(string))
	}

	if v, ok := d.GetOk("schedule_timezone"); ok {
		params.ScheduleTimezone = aws.String(v.(string))
	}

	if v, ok := d.GetOk("schedule_offset"); ok {
		params.ScheduleOffset = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("start_date"); ok {
		params.StartDate = aws.String(v.(string))
	}

	if v, ok := d.GetOk("description"); ok {
		params.Description = aws.String(v.(string))
	}

	_, err := conn.UpdateMaintenanceWindow(params)
	if err != nil {
		return fmt.Errorf("error updating SSM Maintenance Window (%s): %w", d.Id(), err)
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Id(), ssm.ResourceTypeForTaggingMaintenanceWindow, o, n); err != nil {
			return fmt.Errorf("error updating SSM Maintenance Window (%s) tags: %w", d.Id(), err)
		}
	}

	return resourceMaintenanceWindowRead(d, meta)
}

func resourceMaintenanceWindowRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SSMConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	params := &ssm.GetMaintenanceWindowInput{
		WindowId: aws.String(d.Id()),
	}

	resp, err := conn.GetMaintenanceWindow(params)
	if err != nil {
		if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, ssm.ErrCodeDoesNotExistException) {
			log.Printf("[WARN] Maintenance Window %s not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error reading SSM Maintenance Window (%s): %w", d.Id(), err)
	}

	d.Set("allow_unassociated_targets", resp.AllowUnassociatedTargets)
	d.Set("cutoff", resp.Cutoff)
	d.Set("duration", resp.Duration)
	d.Set("enabled", resp.Enabled)
	d.Set("end_date", resp.EndDate)
	d.Set("name", resp.Name)
	d.Set("schedule_timezone", resp.ScheduleTimezone)
	d.Set("schedule_offset", resp.ScheduleOffset)
	d.Set("schedule", resp.Schedule)
	d.Set("start_date", resp.StartDate)
	d.Set("description", resp.Description)

	tags, err := ListTags(conn, d.Id(), ssm.ResourceTypeForTaggingMaintenanceWindow)

	if err != nil {
		return fmt.Errorf("error listing tags for SSM Maintenance Window (%s): %s", d.Id(), err)
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceMaintenanceWindowDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SSMConn

	log.Printf("[INFO] Deleting SSM Maintenance Window: %s", d.Id())

	params := &ssm.DeleteMaintenanceWindowInput{
		WindowId: aws.String(d.Id()),
	}

	_, err := conn.DeleteMaintenanceWindow(params)
	if err != nil {
		return fmt.Errorf("error deleting SSM Maintenance Window (%s): %s", d.Id(), err)
	}

	return nil
}
