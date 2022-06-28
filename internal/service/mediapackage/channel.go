package mediapackage

import (
	"fmt"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/mediapackage"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceChannel() *schema.Resource {
	return &schema.Resource{
		Create: resourceChannelCreate,
		Read:   resourceChannelRead,
		Update: resourceChannelUpdate,
		Delete: resourceChannelDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"channel_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^[\w-]+$`), "must only contain alphanumeric characters, dashes or underscores"),
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "Managed by Terraform",
			},
			"hls_ingest": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ingest_endpoints": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"password": {
										Type:      schema.TypeString,
										Computed:  true,
										Sensitive: true,
									},
									"url": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"username": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
					},
				},
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceChannelCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).MediaPackageConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	input := &mediapackage.CreateChannelInput{
		Id:          aws.String(d.Get("channel_id").(string)),
		Description: aws.String(d.Get("description").(string)),
	}

	if len(tags) > 0 {
		input.Tags = Tags(tags.IgnoreAWS())
	}

	resp, err := conn.CreateChannel(input)
	if err != nil {
		return fmt.Errorf("error creating MediaPackage Channel: %s", err)
	}

	d.SetId(aws.StringValue(resp.Id))

	return resourceChannelRead(d, meta)
}

func resourceChannelRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).MediaPackageConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	input := &mediapackage.DescribeChannelInput{
		Id: aws.String(d.Id()),
	}
	resp, err := conn.DescribeChannel(input)
	if err != nil {
		return fmt.Errorf("error describing MediaPackage Channel: %s", err)
	}
	d.Set("arn", resp.Arn)
	d.Set("channel_id", resp.Id)
	d.Set("description", resp.Description)

	if err := d.Set("hls_ingest", flattenHLSIngest(resp.HlsIngest)); err != nil {
		return fmt.Errorf("error setting hls_ingest: %s", err)
	}

	tags := KeyValueTags(resp.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceChannelUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).MediaPackageConn

	input := &mediapackage.UpdateChannelInput{
		Id:          aws.String(d.Id()),
		Description: aws.String(d.Get("description").(string)),
	}

	_, err := conn.UpdateChannel(input)
	if err != nil {
		return fmt.Errorf("error updating MediaPackage Channel: %s", err)
	}

	arn := d.Get("arn").(string)
	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, arn, o, n); err != nil {
			return fmt.Errorf("error updating MediaPackage Channel (%s) tags: %s", arn, err)
		}
	}

	return resourceChannelRead(d, meta)
}

func resourceChannelDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).MediaPackageConn

	input := &mediapackage.DeleteChannelInput{
		Id: aws.String(d.Id()),
	}
	_, err := conn.DeleteChannel(input)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, mediapackage.ErrCodeNotFoundException) {
			return nil
		}
		return fmt.Errorf("error deleting MediaPackage Channel: %s", err)
	}

	dcinput := &mediapackage.DescribeChannelInput{
		Id: aws.String(d.Id()),
	}
	err = resource.Retry(5*time.Minute, func() *resource.RetryError {
		_, err := conn.DescribeChannel(dcinput)
		if err != nil {
			if tfawserr.ErrCodeEquals(err, mediapackage.ErrCodeNotFoundException) {
				return nil
			}
			return resource.NonRetryableError(err)
		}
		return resource.RetryableError(fmt.Errorf("MediaPackage Channel (%s) still exists", d.Id()))
	})
	if tfresource.TimedOut(err) {
		_, err = conn.DescribeChannel(dcinput)
	}
	if err != nil {
		return fmt.Errorf("error waiting for MediaPackage Channel (%s) deletion: %s", d.Id(), err)
	}

	return nil
}

func flattenHLSIngest(h *mediapackage.HlsIngest) []map[string]interface{} {
	if h.IngestEndpoints == nil {
		return []map[string]interface{}{
			{"ingest_endpoints": []map[string]interface{}{}},
		}
	}

	var ingestEndpoints []map[string]interface{}
	for _, e := range h.IngestEndpoints {
		endpoint := map[string]interface{}{
			"password": aws.StringValue(e.Password),
			"url":      aws.StringValue(e.Url),
			"username": aws.StringValue(e.Username),
		}

		ingestEndpoints = append(ingestEndpoints, endpoint)
	}

	return []map[string]interface{}{
		{"ingest_endpoints": ingestEndpoints},
	}
}
