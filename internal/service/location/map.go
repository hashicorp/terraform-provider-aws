package location

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/locationservice"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceMap() *schema.Resource {
	return &schema.Resource{
		Create: resourceMapCreate,
		Read:   resourceMapRead,
		Update: resourceMapUpdate,
		Delete: resourceMapDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"configuration": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"style": {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validation.StringLenBetween(1, 100),
						},
					},
				},
			},
			"create_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 1000),
			},
			"map_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"map_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 100),
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
			"update_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceMapCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).LocationConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	input := &locationservice.CreateMapInput{}

	if v, ok := d.GetOk("configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.Configuration = expandConfiguration(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("map_name"); ok {
		input.MapName = aws.String(v.(string))
	}

	if len(tags) > 0 {
		input.Tags = Tags(tags.IgnoreAWS())
	}

	output, err := conn.CreateMap(input)

	if err != nil {
		return fmt.Errorf("error creating map: %w", err)
	}

	if output == nil {
		return fmt.Errorf("error creating map: empty result")
	}

	d.SetId(aws.StringValue(output.MapName))

	return resourceMapRead(d, meta)
}

func resourceMapRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).LocationConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	input := &locationservice.DescribeMapInput{
		MapName: aws.String(d.Id()),
	}

	output, err := conn.DescribeMap(input)

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, locationservice.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] Location Service Map (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error getting Location Service Map (%s): %w", d.Id(), err)
	}

	if output == nil {
		return fmt.Errorf("error getting Location Service Map (%s): empty response", d.Id())
	}

	if output.Configuration != nil {
		d.Set("configuration", []interface{}{flattenConfiguration(output.Configuration)})
	} else {
		d.Set("configuration", nil)
	}

	d.Set("create_time", aws.TimeValue(output.CreateTime).Format(time.RFC3339))
	d.Set("description", output.Description)
	d.Set("map_arn", output.MapArn)
	d.Set("map_name", output.MapName)
	d.Set("update_time", aws.TimeValue(output.UpdateTime).Format(time.RFC3339))

	tags := KeyValueTags(output.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceMapUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).LocationConn

	if d.HasChange("description") {
		input := &locationservice.UpdateMapInput{
			MapName: aws.String(d.Id()),
		}

		if v, ok := d.GetOk("description"); ok {
			input.Description = aws.String(v.(string))
		}

		_, err := conn.UpdateMap(input)

		if err != nil {
			return fmt.Errorf("error updating Location Service Map (%s): %w", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Get("map_arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating tags for Location Service Map (%s): %w", d.Id(), err)
		}
	}

	return resourceMapRead(d, meta)
}

func resourceMapDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).LocationConn

	input := &locationservice.DeleteMapInput{
		MapName: aws.String(d.Id()),
	}

	_, err := conn.DeleteMap(input)

	if tfawserr.ErrCodeEquals(err, locationservice.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting Location Service Map (%s): %w", d.Id(), err)
	}

	return nil
}

func expandConfiguration(tfMap map[string]interface{}) *locationservice.MapConfiguration {
	if tfMap == nil {
		return nil
	}

	apiObject := &locationservice.MapConfiguration{}

	if v, ok := tfMap["style"].(string); ok && v != "" {
		apiObject.Style = aws.String(v)
	}

	return apiObject
}

func flattenConfiguration(apiObject *locationservice.MapConfiguration) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Style; v != nil {
		tfMap["style"] = aws.StringValue(v)
	}

	return tfMap
}
