package s3

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

// @SDKDataSource("aws_s3_bucket_server_side_encryption_configuration")
func DataSourceBucketServerSideEncryptionConfiguration() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceBucketServerSideEncryptionConfigurationRead,

		Schema: map[string]*schema.Schema{
			"bucket": {
				Type:     schema.TypeString,
				Required: true,
			},
			"kms_master_key_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"sse_algorithm": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"bucket_key_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
		},
	}
}

func dataSourceBucketServerSideEncryptionConfigurationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3Conn()

	bucket := d.Get("bucket").(string)

	input := &s3.GetBucketEncryptionInput{
		Bucket: aws.String(bucket),
	}

	log.Printf("[INFO] Reading S3 Bucket (%s) server side encryption configuration", bucket)

	// Wait 3 seconds before querying SDK to avoid API returning "AES256" as "sse_algorithm" if just set.
	time.Sleep(3 * time.Second)

	output, err := conn.GetBucketEncryptionWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "[ERROR] Reading S3 bucket encryption configuration: %s", output)
	}

	log.Printf("[INFO] S3 Bucket (%s) server side encryption configuration is: %v", bucket, output)

	for _, rule := range output.ServerSideEncryptionConfiguration.Rules {
		sseAlgorithm := rule.ApplyServerSideEncryptionByDefault.SSEAlgorithm
		d.SetId(bucket)
		d.Set("kms_master_key_id", rule.ApplyServerSideEncryptionByDefault.KMSMasterKeyID)
		d.Set("sse_algorithm", sseAlgorithm)
		d.Set("bucket_key_enabled", rule.BucketKeyEnabled)
	}

	return diags
}
