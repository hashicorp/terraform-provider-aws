package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

func DataSourceEBSDefaultKMSKey() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceEBSDefaultKMSKeyRead,

		Schema: map[string]*schema.Schema{
			"key_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}
func dataSourceEBSDefaultKMSKeyRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	res, err := conn.GetEbsDefaultKmsKeyId(&ec2.GetEbsDefaultKmsKeyIdInput{})
	if err != nil {
		return fmt.Errorf("Error reading EBS default KMS key: %w", err)
	}

	d.SetId(meta.(*conns.AWSClient).Region)
	d.Set("key_arn", res.KmsKeyId)

	return nil
}
