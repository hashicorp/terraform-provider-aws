package qldb

import (
	"fmt"
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/qldb"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func DataSourceLedger() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceLedgerRead,
		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"name": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 32),
					validation.StringMatch(regexp.MustCompile(`^[A-Za-z0-9_-]+`), "must contain only alphanumeric characters, underscores, and hyphens"),
				),
			},

			"permissions_mode": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"deletion_protection": {
				Type:     schema.TypeBool,
				Computed: true,
			},
		},
	}
}

func dataSourceLedgerRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).QLDBConn

	target := d.Get("name")

	req := &qldb.DescribeLedgerInput{
		Name: aws.String(target.(string)),
	}

	log.Printf("[DEBUG] Reading QLDB Ledger: %s", req)
	resp, err := conn.DescribeLedger(req)

	if err != nil {
		return fmt.Errorf("Error describing ledger: %w", err)
	}

	d.SetId(aws.StringValue(resp.Name))
	d.Set("arn", resp.Arn)
	d.Set("deletion_protection", resp.DeletionProtection)
	d.Set("permissions_mode", resp.PermissionsMode)

	return nil
}
