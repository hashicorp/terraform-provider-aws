package ssmcontacts

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
	"time"
)

// @SDKDataSource("aws_ssmcontacts_rotation")
func DataSourceRotation() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceRotationRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Required: true,
			},
			"contact_ids": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"recurrence": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"daily_settings": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"monthly_settings": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"day_of_month": {
										Type:     schema.TypeInt,
										Computed: true,
									},
									"hand_off_time": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
						"number_of_on_calls": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"recurrence_multiplier": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"shift_coverages": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"coverage_times": {
										Type:     schema.TypeList,
										Computed: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"end_time": {
													Type:     schema.TypeString,
													Computed: true,
												},
												"start_time": {
													Type:     schema.TypeString,
													Computed: true,
												},
											},
										},
									},
									"day_of_week": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
						"weekly_settings": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"day_of_week": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"hand_off_time": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
					},
				},
			},
			"start_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"time_zone_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
		},
	}
}

const (
	DSNameRotation = "Rotation Data Source"
)

func dataSourceRotationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).SSMContactsClient(ctx)
	arn := d.Get("arn").(string)

	out, err := FindRotationByID(ctx, conn, arn)
	if err != nil {
		return create.DiagError(names.SSMContacts, create.ErrActionReading, DSNameRotation, arn, err)
	}

	d.SetId(aws.ToString(out.RotationArn))

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

	tags, err := listTags(ctx, conn, d.Id())
	if err != nil {
		return create.DiagError(names.SSMIncidents, create.ErrActionReading, DSNameRotation, d.Id(), err)
	}

	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	//lintignore:AWSR002
	if err := d.Set("tags", tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return create.DiagError(names.SSMContacts, create.ErrActionSetting, DSNameRotation, d.Id(), err)
	}

	return nil
}
