package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/terraform/helper/schema"
)

const maxS3ObjectListReqSize = 1000

func dataSourceAwsS3BucketObjects() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsS3BucketObjectsRead,

		Schema: map[string]*schema.Schema{
			"bucket": {
				Type:     schema.TypeString,
				Required: true,
			},
			"prefix": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"delimiter": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"encoding_type": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"max_keys": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"start_after": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"fetch_owner": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"keys": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"common_prefixes": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"owners": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceAwsS3BucketObjectsRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).s3conn

	bucket := d.Get("bucket").(string)
	prefix := d.Get("prefix").(string)

	input := &s3.HeadBucketInput{
		Bucket: aws.String(bucket),
	}

	log.Printf("[DEBUG] Reading S3 bucket: %s", input)
	_, err := conn.HeadBucket(input)

	if err != nil {
		return fmt.Errorf("Failed listing S3 bucket object keys: %s Bucket: %q", err, bucket)
	}

	d.SetId(fmt.Sprintf("%s_%s", bucket, prefix))

	arn := arn.ARN{
		Partition: meta.(*AWSClient).partition,
		Service:   "s3",
		Resource:  bucket,
	}.String()
	d.Set("arn", arn)

	listInput := s3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
	}

	if prefix != "" {
		listInput.Prefix = aws.String(prefix)
	}

	if s, ok := d.GetOk("delimiter"); ok {
		listInput.Delimiter = aws.String(s.(string))
	}

	if s, ok := d.GetOk("encoding_type"); ok {
		listInput.EncodingType = aws.String(s.(string))
	}

	// MaxKeys attribute refers to max keys returned in a single request
	// (i.e., page size), not the total number of keys returned if you page
	// through the results. This reduces # requests to fewest possible.
	maxKeys := -1
	if max, ok := d.GetOk("max_keys"); ok {
		maxKeys = max.(int)
		if maxKeys > maxS3ObjectListReqSize {
			listInput.MaxKeys = aws.Int64(int64(maxS3ObjectListReqSize))
		} else {
			listInput.MaxKeys = aws.Int64(int64(maxKeys))
		}
	}

	if s, ok := d.GetOk("start_after"); ok {
		listInput.StartAfter = aws.String(s.(string))
	}

	if b, ok := d.GetOk("fetch_owner"); ok {
		listInput.FetchOwner = aws.Bool(b.(bool))
	}

	keys, prefixes, owners, err := listS3Objects(conn, listInput, maxKeys)
	if err != nil {
		return err
	}
	d.Set("keys", keys)
	d.Set("common_prefixes", prefixes)
	d.Set("owners", owners)

	return nil
}

func listS3Objects(conn *s3.S3, input s3.ListObjectsV2Input, maxKeys int) ([]string, []string, []string, error) {
	var objectList []string
	var commonPrefixList []string
	var ownerList []string
	var continueToken *string
	for {
		//page through keys
		input.ContinuationToken = continueToken

		log.Printf("[DEBUG] Requesting page of S3 bucket (%s) object keys", *input.Bucket)
		listOutput, err := conn.ListObjectsV2(&input)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("Failed listing S3 bucket object keys: %s Bucket: %q", err, *input.Bucket)
		}

		for _, content := range listOutput.Contents {
			objectList = append(objectList, *content.Key)
			if input.FetchOwner != nil && *input.FetchOwner {
				ownerList = append(ownerList, *content.Owner.ID)
			}
			if maxKeys > -1 && len(objectList) >= maxKeys {
				break
			}
		}

		for _, commonPrefix := range listOutput.CommonPrefixes {
			commonPrefixList = append(commonPrefixList, *commonPrefix.Prefix)
		}

		// stop requesting if no more results OR all wanted keys done
		if !*listOutput.IsTruncated || (maxKeys > -1 && len(objectList) >= maxKeys) {
			break
		}
		continueToken = listOutput.NextContinuationToken
	}

	return objectList, commonPrefixList, ownerList, nil
}
