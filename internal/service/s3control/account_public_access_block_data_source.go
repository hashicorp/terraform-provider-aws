package s3control

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3control"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func DataSourceAccountPublicAccessBlock() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceAccountPublicAccessBlockRead,

		Schema: map[string]*schema.Schema{
			"account_id": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: verify.ValidAccountID,
			},
			"block_public_acls": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"block_public_policy": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"ignore_public_acls": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"restrict_public_buckets": {
				Type:     schema.TypeBool,
				Computed: true,
			},
		},
	}
}

func dataSourceAccountPublicAccessBlockRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).S3ControlConn

	accountID := meta.(*conns.AWSClient).AccountID
	if v, ok := d.GetOk("account_id"); ok {
		accountID = v.(string)
	}

	input := &s3control.GetPublicAccessBlockInput{
		AccountId: aws.String(accountID),
	}

	log.Printf("[DEBUG] Reading Account access block: %s", input)

	output, err := conn.GetPublicAccessBlock(input)

	if err != nil {
		return diag.Errorf("error reading S3 Account Public Access Block: %s", err)
	}

	if output == nil || output.PublicAccessBlockConfiguration == nil {
		return diag.Errorf("error reading S3 Account Public Access Block (%s): missing public access block configuration", accountID)
	}

	d.SetId(accountID)
	d.Set("block_public_acls", output.PublicAccessBlockConfiguration.BlockPublicAcls)
	d.Set("block_public_policy", output.PublicAccessBlockConfiguration.BlockPublicPolicy)
	d.Set("ignore_public_acls", output.PublicAccessBlockConfiguration.IgnorePublicAcls)
	d.Set("restrict_public_buckets", output.PublicAccessBlockConfiguration.RestrictPublicBuckets)

	return nil
}
