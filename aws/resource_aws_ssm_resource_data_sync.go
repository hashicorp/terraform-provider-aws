package aws

import (
	"bytes"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/hashicorp/terraform/helper/hashcode"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsSsmResourceDataSync() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsSsmResourceDataSyncCreate,
		Read:   resourceAwsSsmResourceDataSyncRead,
		Delete: resourceAwsSsmResourceDataSyncDelete,

		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"destination": {
				Type:     schema.TypeSet,
				Required: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"key": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"bucket": {
							Type:     schema.TypeString,
							Required: true,
						},
						"prefix": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"region": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
				Set: func(v interface{}) int {
					var buf bytes.Buffer
					m := v.(map[string]interface{})
					buf.WriteString(fmt.Sprintf("%s-", strings.ToLower(m["bucket"].(string))))
					buf.WriteString(fmt.Sprintf("%s-", strings.ToLower(m["region"].(string))))
					if val, ok := m["key"]; ok {
						buf.WriteString(fmt.Sprintf("%s-", strings.ToLower(val.(string))))
					}
					if val, ok := m["prefix"]; ok {
						buf.WriteString(fmt.Sprintf("%s-", strings.ToLower(val.(string))))
					}
					return hashcode.String(buf.String())
				},
			},
		},
	}
}

func resourceAwsSsmResourceDataSyncCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ssmconn

	input := &ssm.CreateResourceDataSyncInput{
		S3Destination: &ssm.ResourceDataSyncS3Destination{
			SyncFormat: aws.String(ssm.ResourceDataSyncS3FormatJsonSerDe),
		},
		SyncName: aws.String(d.Get("name").(string)),
	}
	destination := d.Get("destination").(*schema.Set).List()
	for _, v := range destination {
		dest := v.(map[string]interface{})
		input.S3Destination.SetBucketName(dest["bucket"].(string))
		input.S3Destination.SetRegion(dest["region"].(string))
		if val, ok := dest["key"].(string); ok && val != "" {
			input.S3Destination.SetAWSKMSKeyARN(val)
		}
		if val, ok := dest["prefix"].(string); ok && val != "" {
			input.S3Destination.SetPrefix(val)
		}
	}

	_, err := conn.CreateResourceDataSync(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case ssm.ErrCodeResourceDataSyncAlreadyExistsException:
				if err := resourceAwsSsmResourceDataSyncDeleteAndCreate(meta, input); err != nil {
					return err
				}
			default:
				return err
			}
		} else {
			return err
		}
	}
	return resourceAwsSsmResourceDataSyncRead(d, meta)
}

func resourceAwsSsmResourceDataSyncRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ssmconn

	nextToken := ""
	found := false
	for {
		input := &ssm.ListResourceDataSyncInput{}
		if nextToken != "" {
			input.NextToken = aws.String(nextToken)
		}
		resp, err := conn.ListResourceDataSync(input)
		if err != nil {
			return err
		}
		for _, v := range resp.ResourceDataSyncItems {
			if *v.SyncName == d.Get("name").(string) {
				found = true
			}
		}
		if found || *resp.NextToken == "" {
			break
		}
		nextToken = *resp.NextToken
	}
	if !found {
		log.Printf("[INFO] No Resource Data Sync found for SyncName: %s", d.Get("name").(string))
	}
	return nil
}

func resourceAwsSsmResourceDataSyncDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ssmconn

	input := &ssm.DeleteResourceDataSyncInput{
		SyncName: aws.String(d.Get("name").(string)),
	}

	_, err := conn.DeleteResourceDataSync(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case ssm.ErrCodeResourceDataSyncNotFoundException:
				return nil
			default:
				return err
			}
		}
		return err
	}
	return nil
}

func resourceAwsSsmResourceDataSyncDeleteAndCreate(meta interface{}, input *ssm.CreateResourceDataSyncInput) error {
	conn := meta.(*AWSClient).ssmconn

	delinput := &ssm.DeleteResourceDataSyncInput{
		SyncName: input.SyncName,
	}

	_, err := conn.DeleteResourceDataSync(delinput)
	if err != nil {
		return err
	}
	_, err = conn.CreateResourceDataSync(input)
	if err != nil {
		return err
	}
	return nil
}
