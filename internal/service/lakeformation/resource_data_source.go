package lakeformation

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lakeformation"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func DataSourceResource() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceResourceRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
			"last_modified": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"role_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceResourceRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).LakeFormationConn

	input := &lakeformation.DescribeResourceInput{}

	if v, ok := d.GetOk("arn"); ok {
		input.ResourceArn = aws.String(v.(string))
	}

	output, err := conn.DescribeResource(input)

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, lakeformation.ErrCodeEntityNotFoundException) {
		log.Printf("[WARN] Resource Lake Formation Resource (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading data source, Lake Formation Resource (arn: %s): %w", aws.StringValue(input.ResourceArn), err)
	}

	if output == nil || output.ResourceInfo == nil {
		return fmt.Errorf("error reading data source, Lake Formation Resource: empty response")
	}

	d.SetId(aws.StringValue(input.ResourceArn))
	// d.Set("arn", output.ResourceInfo.ResourceArn) // output not including resource arn currently
	d.Set("role_arn", output.ResourceInfo.RoleArn)
	if output.ResourceInfo.LastModified != nil { // output not including last modified currently
		d.Set("last_modified", output.ResourceInfo.LastModified.Format(time.RFC3339))
	}

	return nil
}
