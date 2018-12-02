package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/terraform/helper/schema"
)

func dataSourceAwsS3PresignedUrl() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsS3PresignedUrlRead,

		Schema: map[string]*schema.Schema{
			"bucket": {
				Type:     schema.TypeString,
				Required: true,
			},
			"key": {
				Type:     schema.TypeString,
				Required: true,
			},
			"expiration_time": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"put": {
				Type:     schema.TypeBool,
				Default:  false,
				Optional: true,
			},
			"url": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceAwsS3PresignedUrlRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).s3conn

	bucket := d.Get("bucket").(string)
	key := d.Get("key").(string)
	expirationTime := time.Duration(d.Get("expiration_time").(int))
	put := d.Get("put").(bool)

	presignedURL, err := dataSourceAwsS3PresignedUrlCreatePresignedUrl(conn, bucket, key, expirationTime, put)

	if err != nil {
		return fmt.Errorf("Error failed to sign request: %q", err)
	}

	d.SetId(time.Now().UTC().String())
	d.Set("url", presignedURL)

	return nil
}

func dataSourceAwsS3PresignedUrlCreatePresignedUrl(conn *s3.S3, bucket string, key string, expirationTime time.Duration, put bool) (string, error) {
	if put {
		log.Print("[DEBUG] Creating S3 PUT request")
		req, _ := conn.PutObjectRequest(&s3.PutObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(key),
		})

		log.Print("[DEBUG] Signing S3 PUT request")
		return req.Presign(expirationTime * time.Second)
	}

	log.Print("[DEBUG] Creating S3 GET request")
	req, _ := conn.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})

	log.Print("[DEBUG] Signing S3 GET request")
	return req.Presign(expirationTime * time.Second)
}
