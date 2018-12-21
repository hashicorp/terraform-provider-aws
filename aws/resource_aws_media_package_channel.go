package aws

import (
	"fmt"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/mediapackage"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
)

func resourceAwsMediaPackageChannel() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsMediaPackageChannelCreate,
		Read:   resourceAwsMediaPackageChannelRead,
		Update: resourceAwsMediaPackageChannelUpdate,
		Delete: resourceAwsMediaPackageChannelDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
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
			"ingest_endpoints": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"password": {
							Type:     schema.TypeString,
							Computed: true,
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
	}
}

func resourceAwsMediaPackageChannelCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).mediapackageconn

	input := &mediapackage.CreateChannelInput{
		Id:          aws.String(d.Get("channel_id").(string)),
		Description: aws.String(d.Get("description").(string)),
	}

	_, err := conn.CreateChannel(input)
	if err != nil {
		return err
	}

	d.SetId(d.Get("channel_id").(string))
	return resourceAwsMediaPackageChannelRead(d, meta)
}

func resourceAwsMediaPackageChannelRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).mediapackageconn

	input := &mediapackage.DescribeChannelInput{
		Id: aws.String(d.Id()),
	}
	resp, err := conn.DescribeChannel(input)
	if err != nil {
		return err
	}
	d.Set("arn", resp.Arn)
	d.Set("description", resp.Description)
	d.Set("ingest_endpoints", convertIngestEndpoints(resp.HlsIngest.IngestEndpoints))

	return nil
}

func resourceAwsMediaPackageChannelUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).mediapackageconn

	input := &mediapackage.UpdateChannelInput{
		Id:          aws.String(d.Id()),
		Description: aws.String(d.Get("description").(string)),
	}

	_, err := conn.UpdateChannel(input)
	if err != nil {
		return err
	}

	return resourceAwsMediaPackageChannelRead(d, meta)
}

func resourceAwsMediaPackageChannelDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).mediapackageconn

	input := &mediapackage.DeleteChannelInput{
		Id: aws.String(d.Id()),
	}
	_, err := conn.DeleteChannel(input)
	if err != nil {
		if isAWSErr(err, mediapackage.ErrCodeNotFoundException, "") {
			return nil
		}
		return err
	}

	err = resource.Retry(5*time.Minute, func() *resource.RetryError {
		dcinput := &mediapackage.DescribeChannelInput{
			Id: aws.String(d.Id()),
		}
		_, err := conn.DescribeChannel(dcinput)
		if err != nil {
			if isAWSErr(err, mediapackage.ErrCodeNotFoundException, "") {
				return nil
			}
			return resource.NonRetryableError(err)
		}
		return resource.RetryableError(fmt.Errorf("Media Package Channel (%s) still exists", d.Id()))
	})
	if err != nil {
		return fmt.Errorf("error waiting for Media Package Channel (%s) deletion: %s", d.Id(), err)
	}

	return nil
}

func convertIngestEndpoints(es []*mediapackage.IngestEndpoint) (ingestEndpoints []map[string]interface{}) {
	for _, e := range es {
		endpoint := map[string]interface{}{
			"password": *e.Password,
			"url":      *e.Url,
			"username": *e.Username,
		}

		ingestEndpoints = append(ingestEndpoints, endpoint)
	}

	return
}
