package s3

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func ResourceBucketPublicAccessBlock() *schema.Resource {
	return &schema.Resource{
		Create: resourceBucketPublicAccessBlockCreate,
		Read:   resourceBucketPublicAccessBlockRead,
		Update: resourceBucketPublicAccessBlockUpdate,
		Delete: resourceBucketPublicAccessBlockDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"bucket": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"block_public_acls": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			"block_public_policy": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			"ignore_public_acls": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			"restrict_public_buckets": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
		},
	}
}

func resourceBucketPublicAccessBlockCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).S3Conn
	bucket := d.Get("bucket").(string)

	input := &s3.PutPublicAccessBlockInput{
		Bucket: aws.String(bucket),
		PublicAccessBlockConfiguration: &s3.PublicAccessBlockConfiguration{
			BlockPublicAcls:       aws.Bool(d.Get("block_public_acls").(bool)),
			BlockPublicPolicy:     aws.Bool(d.Get("block_public_policy").(bool)),
			IgnorePublicAcls:      aws.Bool(d.Get("ignore_public_acls").(bool)),
			RestrictPublicBuckets: aws.Bool(d.Get("restrict_public_buckets").(bool)),
		},
	}

	log.Printf("[DEBUG] S3 bucket: %s, public access block: %v", bucket, input.PublicAccessBlockConfiguration)
	err := resource.Retry(1*time.Minute, func() *resource.RetryError {
		_, err := conn.PutPublicAccessBlock(input)

		if tfawserr.ErrCodeEquals(err, s3.ErrCodeNoSuchBucket) {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})
	if tfresource.TimedOut(err) {
		_, err = conn.PutPublicAccessBlock(input)
	}
	if err != nil {
		return fmt.Errorf("error creating public access block policy for S3 bucket (%s): %s", bucket, err)
	}

	d.SetId(bucket)
	return resourceBucketPublicAccessBlockRead(d, meta)
}

func resourceBucketPublicAccessBlockRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).S3Conn

	input := &s3.GetPublicAccessBlockInput{
		Bucket: aws.String(d.Id()),
	}

	// Retry for eventual consistency on creation
	var output *s3.GetPublicAccessBlockOutput
	err := resource.Retry(propagationTimeout, func() *resource.RetryError {
		var err error
		output, err = conn.GetPublicAccessBlock(input)

		if d.IsNewResource() && tfawserr.ErrCodeEquals(err, s3.ErrCodeNoSuchBucket) {
			return resource.RetryableError(err)
		}

		if d.IsNewResource() && tfawserr.ErrCodeEquals(err, ErrCodeNoSuchPublicAccessBlockConfiguration) {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		output, err = conn.GetPublicAccessBlock(input)
	}

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, ErrCodeNoSuchPublicAccessBlockConfiguration) {
		log.Printf("[WARN] S3 Bucket Public Access Block (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, s3.ErrCodeNoSuchBucket) {
		log.Printf("[WARN] S3 Bucket (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading S3 bucket Public Access Block (%s): %w", d.Id(), err)
	}

	if output == nil || output.PublicAccessBlockConfiguration == nil {
		return fmt.Errorf("error reading S3 Bucket Public Access Block (%s): empty response", d.Id())
	}

	d.Set("bucket", d.Id())
	d.Set("block_public_acls", output.PublicAccessBlockConfiguration.BlockPublicAcls)
	d.Set("block_public_policy", output.PublicAccessBlockConfiguration.BlockPublicPolicy)
	d.Set("ignore_public_acls", output.PublicAccessBlockConfiguration.IgnorePublicAcls)
	d.Set("restrict_public_buckets", output.PublicAccessBlockConfiguration.RestrictPublicBuckets)

	return nil
}

func resourceBucketPublicAccessBlockUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).S3Conn

	input := &s3.PutPublicAccessBlockInput{
		Bucket: aws.String(d.Id()),
		PublicAccessBlockConfiguration: &s3.PublicAccessBlockConfiguration{
			BlockPublicAcls:       aws.Bool(d.Get("block_public_acls").(bool)),
			BlockPublicPolicy:     aws.Bool(d.Get("block_public_policy").(bool)),
			IgnorePublicAcls:      aws.Bool(d.Get("ignore_public_acls").(bool)),
			RestrictPublicBuckets: aws.Bool(d.Get("restrict_public_buckets").(bool)),
		},
	}

	log.Printf("[DEBUG] Updating S3 bucket Public Access Block: %s", input)
	_, err := conn.PutPublicAccessBlock(input)
	if err != nil {
		return fmt.Errorf("error updating S3 Bucket Public Access Block (%s): %s", d.Id(), err)
	}

	// Workaround API eventual consistency issues. This type of logic should not normally be used.
	// We cannot reliably determine when the Read after Update might be properly updated.
	// Rather than introduce complicated retry logic, we presume that a lack of an update error
	// means our update succeeded with our expected values.
	d.Set("block_public_acls", input.PublicAccessBlockConfiguration.BlockPublicAcls)
	d.Set("block_public_policy", input.PublicAccessBlockConfiguration.BlockPublicPolicy)
	d.Set("ignore_public_acls", input.PublicAccessBlockConfiguration.IgnorePublicAcls)
	d.Set("restrict_public_buckets", input.PublicAccessBlockConfiguration.RestrictPublicBuckets)

	// Skip normal Read after Update due to eventual consistency issues
	return nil
}

func resourceBucketPublicAccessBlockDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).S3Conn

	input := &s3.DeletePublicAccessBlockInput{
		Bucket: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] S3 bucket: %s, delete public access block", d.Id())
	_, err := conn.DeletePublicAccessBlock(input)

	if tfawserr.ErrCodeEquals(err, ErrCodeNoSuchPublicAccessBlockConfiguration) {
		return nil
	}

	if tfawserr.ErrCodeEquals(err, s3.ErrCodeNoSuchBucket) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting S3 Bucket Public Access Block (%s): %s", d.Id(), err)
	}

	return nil
}
