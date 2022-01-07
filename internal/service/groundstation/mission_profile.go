package groundstation

import (
	"fmt"
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/groundstation"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceMissionProfile() *schema.Resource {
	return &schema.Resource{
		Create: resourceMissionProfileCreate,
		Read:   resourceMissionProfileRead,
		Update: resourceMissionProfileUpdate,
		Delete: resourceMissionProfileDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"contact_post_pass_duration_seconds": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(1, 21600),
			},
			"contact_pre_pass_duration_seconds": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(1, 21600),
			},
			"dataflow_edges": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Schema{
					Type: schema.TypeList,
					Elem: &schema.Schema{
						Type:         schema.TypeString,
						ValidateFunc: verify.ValidARN,
						MinItems:     2,
						MaxItems:     2,
					},
				},
			},
			"minimum_viable_contact_duration_seconds": {
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validation.IntBetween(1, 21600),
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile("^[ a-zA-Z0-9_:-]{1,256}$"), "must contain only alphanumeric characters, hyphens, underscores, and colons. Must be between 1 and 256 characters long."),
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
			"tracking_config_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
		},
	}
}
func resourceMissionProfileCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).GroundStationConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	input := &groundstation.CreateMissionProfileInput{
		ContactPostPassDurationSeconds:      aws.Int64(int64(d.Get("contact_post_pass_duration_seconds").(int))),
		ContactPrePassDurationSeconds:       aws.Int64(int64(d.Get("contact_pre_pass_duration_seconds").(int))),
		MinimumViableContactDurationSeconds: aws.Int64(int64(d.Get("minimum_viable_contact_duration_seconds").(int))),
		Name:                                aws.String(d.Get("name").(string)),
		Tags:                                Tags(tags.IgnoreAWS()),
		TrackingConfigArn:                   aws.String(d.Get("tracking_config_arn").(string)),
	}
	log.Printf("[DEBUG] Creating Ground Station Mission Profile: %s", input)
	output, err := conn.CreateMissionProfile(input)
	if err != nil {
		return fmt.Errorf("error creating Ground Station Mission Profile: %s", err)
	}

	d.SetId(aws.StringValue(output.MissionProfileId))
	input.SetDataflowEdges(d.Get("dataflow_edges").([][]string))
	return resourceMissionProfileRead(d, meta)
}

func Tags(keyValueTags tftags.KeyValueTags) {
	panic("unimplemented")
}
func resourceMissionProfileRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).GroundStationConn
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	input := &groundstation.GetMissionProfileInput{
		MissionProfileId: aws.String(d.Id()),
	}
	log.Printf("[DEBUG] Reading Ground Station Mission Profile: %s", input)
	output, err := conn.GetMissionProfile(input)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, groundstation.ErrCodeResourceNotFoundException) {
			log.Printf("[WARN] Ground Station Mission Profile %q not found - removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error reading Ground Station Mission Profile: %s", err)
	}

	d.Set("contact_post_pass_duration_seconds", output.ContactPostPassDurationSeconds)
	d.Set("contact_pre_pass_duration_seconds", output.ContactPrePassDurationSeconds)
	d.Set("dataflow_edges", flattenDataflowEdges(output.DataflowEdges))
	d.Set("minimum_viable_contact_duration_seconds", output.MinimumViableContactDurationSeconds)
	d.Set("name", output.Name)
	d.Set("tracking_config_arn", output.TrackingConfigArn)
	tags := KeyValueTags(output.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)
	if err := d.Set("tags", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}
	return nil
}

func resourceMissionProfileUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).GroundStationConn

	input := &groundstation.UpdateMissionProfileInput{
		ContactPostPassDurationSeconds:      aws.Int64(int64(d.Get("contact_post_pass_duration_seconds").(int))),
		ContactPrePassDurationSeconds:       aws.Int64(int64(d.Get("contact_pre_pass_duration_seconds").(int))),,
		MinimumViableContactDurationSeconds: aws.Int64(int64(d.Get("minimum_viable_contact_duration_seconds").(int))),
		Name:                                aws.String(d.Get("name").(string)),
		TrackingConfigArn:                   aws.String(d.Get("tracking_config_arn").(string)),
		MissionProfileId:                    aws.String(d.Id()),
	}
	log.Printf("[DEBUG] Updating Ground Station Mission Profile: %s", input)
	_, err := conn.UpdateMissionProfile(input)
	if err != nil {
		return fmt.Errorf("error updating Ground Station Mission Profile: %s", err)
	}
	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Id(), o, n, "MissionProfile"); err != nil {
			return fmt.Errorf("error updating tags: %s", err)
		}
	}
	input.SetDataflowEdges(d.Get("dataflow_edges").([][]string))
	return resourceMissionProfileRead(d, meta)
}

func resourceMissionProfileDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).GroundStationConn
	input := &groundstation.DeleteMissionProfileInput{
		MissionProfileId: aws.String(d.Id()),
	}
	log.Printf("[DEBUG] Deleting Ground Station Mission Profile: %s", input)
	_, err := conn.DeleteMissionProfile(input)
	if err != nil {
		return fmt.Errorf("error deleting Ground Station Mission Profile: %s", err)
	}

	return nil
}
