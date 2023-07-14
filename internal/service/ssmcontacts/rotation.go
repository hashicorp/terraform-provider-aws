package ssmcontacts

import (
	"context"
	"errors"
	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"log"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssmcontacts"
	"github.com/aws/aws-sdk-go-v2/service/ssmcontacts/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_ssmcontacts_rotation")
func ResourceRotation() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceRotationCreate,
		ReadWithoutTimeout:   resourceRotationRead,
		UpdateWithoutTimeout: resourceRotationUpdate,
		DeleteWithoutTimeout: resourceRotationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"contact_ids": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"recurrence": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"daily_settings": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Schema{
								Type:             schema.TypeString,
								ValidateDiagFunc: validation.ToDiagFunc(handOffTimeValidator),
							},
						},
						"monthly_settings": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"day_of_month": {
										Type:     schema.TypeInt,
										Required: true,
									},
									"hand_off_time": {
										Type:             schema.TypeString,
										Required:         true,
										ValidateDiagFunc: validation.ToDiagFunc(handOffTimeValidator),
									},
								},
							},
						},
						"number_of_on_calls": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"recurrence_multiplier": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"shift_coverages": {
							Type:     schema.TypeList,
							Optional: true,
							Computed: true, // This is computed to allow for clearing the diff to handle erroneous diffs
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"coverage_times": {
										Type:     schema.TypeList,
										MaxItems: 1,
										Required: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"end_time": {
													Type:             schema.TypeString,
													Optional:         true,
													ValidateDiagFunc: validation.ToDiagFunc(handOffTimeValidator),
												},
												"start_time": {
													Type:             schema.TypeString,
													Optional:         true,
													ValidateDiagFunc: validation.ToDiagFunc(handOffTimeValidator),
												},
											},
										},
									},
									"day_of_week": {
										Type:             schema.TypeString,
										Required:         true,
										ValidateDiagFunc: enum.Validate[types.DayOfWeek](),
									},
								},
							},
						},
						"weekly_settings": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"day_of_week": {
										Type:             schema.TypeString,
										Required:         true,
										ValidateDiagFunc: enum.Validate[types.DayOfWeek](),
									},
									"hand_off_time": {
										Type:             schema.TypeString,
										Required:         true,
										ValidateDiagFunc: validation.ToDiagFunc(handOffTimeValidator),
									},
								},
							},
						},
					},
				},
			},
			"start_time": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: validation.ToDiagFunc(validation.IsRFC3339Time),
			},
			"time_zone_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: customdiff.Sequence(
			func(_ context.Context, diff *schema.ResourceDiff, _ interface{}) error {
				if diff.HasChange("recurrence.0.shift_coverages") {
					o, n := diff.GetChange("recurrence.0.shift_coverages")

					sortShiftCoverages(o.([]interface{}))
					sortShiftCoverages(n.([]interface{}))

					isEqual := cmp.Diff(o, n) == ""

					if isEqual {
						return diff.Clear("recurrence.0.shift_coverages")
					}
				}

				return nil
			},
			verify.SetTagsDiff,
		),
	}
}

const (
	ResNameRotation = "Rotation"
)

var handOffTimeValidator = validation.StringMatch(regexp.MustCompile(`^\d\d:\d\d$`), "Time must be in 24-hour time format, e.g. \"01:00\"")

func resourceRotationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).SSMContactsClient(ctx)

	in := &ssmcontacts.CreateRotationInput{
		ContactIds: flex.ExpandStringValueList(d.Get("contact_ids").([]interface{})),
		Name:       aws.String(d.Get("name").(string)),
		Recurrence: expandRecurrence(d.Get("recurrence").([]interface{}), ctx),
		Tags:       getTagsIn(ctx),
		TimeZoneId: aws.String(d.Get("time_zone_id").(string)),
	}

	if v, ok := d.GetOk("start_time"); ok {
		startTime, _ := time.Parse(time.RFC3339, v.(string))
		in.StartTime = aws.Time(startTime)
	}

	out, err := conn.CreateRotation(ctx, in)
	if err != nil {
		return create.DiagError(names.SSMContacts, create.ErrActionCreating, ResNameRotation, d.Get("name").(string), err)
	}

	if out == nil {
		return create.DiagError(names.SSMContacts, create.ErrActionCreating, ResNameRotation, d.Get("name").(string), errors.New("empty output"))
	}

	d.SetId(aws.ToString(out.RotationArn))

	return resourceRotationRead(ctx, d, meta)
}

func resourceRotationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).SSMContactsClient(ctx)

	out, err := FindRotationByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] SSMContacts Rotation (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return create.DiagError(names.SSMContacts, create.ErrActionReading, ResNameRotation, d.Id(), err)
	}

	d.Set("arn", out.RotationArn)
	d.Set("contact_ids", out.ContactIds)
	d.Set("name", out.Name)
	d.Set("time_zone_id", out.TimeZoneId)

	if out.StartTime != nil {
		d.Set("start_time", out.StartTime.Format(time.RFC3339))
	}

	if err := d.Set("recurrence", flattenRecurrence(out.Recurrence, ctx)); err != nil {
		return create.DiagError(names.SSMContacts, create.ErrActionSetting, DSNameRotation, d.Id(), err)
	}

	return nil
}

func resourceRotationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).SSMContactsClient(ctx)

	update := false

	in := &ssmcontacts.UpdateRotationInput{
		RotationId: aws.String(d.Id()),
	}

	if d.HasChanges("contact_ids") {
		in.ContactIds = flex.ExpandStringValueList(d.Get("contact_ids").([]interface{}))
		update = true
	}

	// Recurrence is a required field, but we don't want to force an update every time if no changes
	in.Recurrence = expandRecurrence(d.Get("recurrence").([]interface{}), ctx)
	if d.HasChanges("recurrence") {
		update = true
	}

	if d.HasChanges("start_time") {
		startTime, _ := time.Parse(time.RFC3339, d.Get("start_time").(string))
		in.StartTime = aws.Time(startTime)
		update = true
	}

	if d.HasChanges("time_zone_id") {
		in.TimeZoneId = aws.String(d.Get("time_zone_id").(string))
		update = true
	}

	if !update {
		return nil
	}

	log.Printf("[DEBUG] Updating SSMContacts Rotation (%s): %#v", d.Id(), in)
	_, err := conn.UpdateRotation(ctx, in)
	if err != nil {
		return create.DiagError(names.SSMContacts, create.ErrActionUpdating, ResNameRotation, d.Id(), err)
	}

	return resourceRotationRead(ctx, d, meta)
}

func resourceRotationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).SSMContactsClient(ctx)

	log.Printf("[INFO] Deleting SSMContacts Rotation %s", d.Id())

	_, err := conn.DeleteRotation(ctx, &ssmcontacts.DeleteRotationInput{
		RotationId: aws.String(d.Id()),
	})

	if err != nil {
		var nfe *types.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return nil
		}

		return create.DiagError(names.SSMContacts, create.ErrActionDeleting, ResNameRotation, d.Id(), err)
	}

	return nil
}
