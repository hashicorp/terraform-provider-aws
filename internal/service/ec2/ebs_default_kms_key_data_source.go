package ec2

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

func DataSourceEBSDefaultKMSKey() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceEBSDefaultKMSKeyRead,

		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(20 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"key_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}
func dataSourceEBSDefaultKMSKeyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn()

	res, err := conn.GetEbsDefaultKmsKeyIdWithContext(ctx, &ec2.GetEbsDefaultKmsKeyIdInput{})
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "Error reading EBS default KMS key: %s", err)
	}

	d.SetId(meta.(*conns.AWSClient).Region)
	d.Set("key_arn", res.KmsKeyId)

	return diags
}
