package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3control"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/service/s3control/waiter"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceAccountPublicAccessBlock() *schema.Resource {
	return &schema.Resource{
		Create: resourceAccountPublicAccessBlockCreate,
		Read:   resourceAccountPublicAccessBlockRead,
		Update: resourceAccountPublicAccessBlockUpdate,
		Delete: resourceAccountPublicAccessBlockDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"account_id": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidAccountID,
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

func resourceAccountPublicAccessBlockCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).S3ControlConn

	accountID := meta.(*conns.AWSClient).AccountID
	if v, ok := d.GetOk("account_id"); ok {
		accountID = v.(string)
	}

	input := &s3control.PutPublicAccessBlockInput{
		AccountId: aws.String(accountID),
		PublicAccessBlockConfiguration: &s3control.PublicAccessBlockConfiguration{
			BlockPublicAcls:       aws.Bool(d.Get("block_public_acls").(bool)),
			BlockPublicPolicy:     aws.Bool(d.Get("block_public_policy").(bool)),
			IgnorePublicAcls:      aws.Bool(d.Get("ignore_public_acls").(bool)),
			RestrictPublicBuckets: aws.Bool(d.Get("restrict_public_buckets").(bool)),
		},
	}

	log.Printf("[DEBUG] Creating S3 Account Public Access Block: %s", input)
	_, err := conn.PutPublicAccessBlock(input)
	if err != nil {
		return fmt.Errorf("error creating S3 Account Public Access Block: %s", err)
	}

	d.SetId(accountID)

	return resourceAccountPublicAccessBlockRead(d, meta)
}

func resourceAccountPublicAccessBlockRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).S3ControlConn

	input := &s3control.GetPublicAccessBlockInput{
		AccountId: aws.String(d.Id()),
	}

	// Retry for eventual consistency on creation
	var output *s3control.GetPublicAccessBlockOutput
	err := resource.Retry(waiter.propagationTimeout, func() *resource.RetryError {
		var err error
		output, err = conn.GetPublicAccessBlock(input)

		if d.IsNewResource() && tfawserr.ErrCodeEquals(err, s3control.ErrCodeNoSuchPublicAccessBlockConfiguration) {
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

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, s3control.ErrCodeNoSuchPublicAccessBlockConfiguration) {
		log.Printf("[WARN] S3 Account Public Access Block (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading S3 Account Public Access Block: %s", err)
	}

	if output == nil || output.PublicAccessBlockConfiguration == nil {
		return fmt.Errorf("error reading S3 Account Public Access Block (%s): missing public access block configuration", d.Id())
	}

	d.Set("account_id", d.Id())
	d.Set("block_public_acls", output.PublicAccessBlockConfiguration.BlockPublicAcls)
	d.Set("block_public_policy", output.PublicAccessBlockConfiguration.BlockPublicPolicy)
	d.Set("ignore_public_acls", output.PublicAccessBlockConfiguration.IgnorePublicAcls)
	d.Set("restrict_public_buckets", output.PublicAccessBlockConfiguration.RestrictPublicBuckets)

	return nil
}

func resourceAccountPublicAccessBlockUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).S3ControlConn

	input := &s3control.PutPublicAccessBlockInput{
		AccountId: aws.String(d.Id()),
		PublicAccessBlockConfiguration: &s3control.PublicAccessBlockConfiguration{
			BlockPublicAcls:       aws.Bool(d.Get("block_public_acls").(bool)),
			BlockPublicPolicy:     aws.Bool(d.Get("block_public_policy").(bool)),
			IgnorePublicAcls:      aws.Bool(d.Get("ignore_public_acls").(bool)),
			RestrictPublicBuckets: aws.Bool(d.Get("restrict_public_buckets").(bool)),
		},
	}

	log.Printf("[DEBUG] Updating S3 Account Public Access Block: %s", input)
	_, err := conn.PutPublicAccessBlock(input)
	if err != nil {
		return fmt.Errorf("error updating S3 Account Public Access Block (%s): %s", d.Id(), err)
	}

	if d.HasChange("block_public_acls") {
		if _, err := waiter.waitPublicAccessBlockConfigurationBlockPublicACLsUpdated(conn, d.Id(), d.Get("block_public_acls").(bool)); err != nil {
			return fmt.Errorf("error waiting for S3 Account Public Access Block (%s) block_public_acls update: %w", d.Id(), err)
		}
	}

	if d.HasChange("block_public_policy") {
		if _, err := waiter.waitPublicAccessBlockConfigurationBlockPublicPolicyUpdated(conn, d.Id(), d.Get("block_public_policy").(bool)); err != nil {
			return fmt.Errorf("error waiting for S3 Account Public Access Block (%s) block_public_policy update: %w", d.Id(), err)
		}
	}

	if d.HasChange("ignore_public_acls") {
		if _, err := waiter.waitPublicAccessBlockConfigurationIgnorePublicACLsUpdated(conn, d.Id(), d.Get("ignore_public_acls").(bool)); err != nil {
			return fmt.Errorf("error waiting for S3 Account Public Access Block (%s) ignore_public_acls update: %w", d.Id(), err)
		}
	}

	if d.HasChange("restrict_public_buckets") {
		if _, err := waiter.waitPublicAccessBlockConfigurationRestrictPublicBucketsUpdated(conn, d.Id(), d.Get("restrict_public_buckets").(bool)); err != nil {
			return fmt.Errorf("error waiting for S3 Account Public Access Block (%s) restrict_public_buckets update: %w", d.Id(), err)
		}
	}

	return resourceAccountPublicAccessBlockRead(d, meta)
}

func resourceAccountPublicAccessBlockDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).S3ControlConn

	input := &s3control.DeletePublicAccessBlockInput{
		AccountId: aws.String(d.Id()),
	}

	_, err := conn.DeletePublicAccessBlock(input)

	if tfawserr.ErrMessageContains(err, s3control.ErrCodeNoSuchPublicAccessBlockConfiguration, "") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting S3 Account Public Access Block (%s): %s", d.Id(), err)
	}

	return nil
}
