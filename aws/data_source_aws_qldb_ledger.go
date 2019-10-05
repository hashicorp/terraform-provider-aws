package aws

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/qldb"
	"log"
	"regexp"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func dataSourceAwsQLDBLedger() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsQLDBLedgerRead,
		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"name": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
					value := v.(string)
					validNamePattern := "^[A-Za-z0-9_-]+$"
					validName, nameMatchErr := regexp.MatchString(validNamePattern, value)
					if !validName || nameMatchErr != nil {
						errors = append(errors, fmt.Errorf(
							"%q must match regex '%v'", k, validNamePattern))
					}
					return
				},
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

func dataSourceAwsQLDBLedgerRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).qldbconn

	target := d.Get("name")

	req := &qldb.DescribeLedgerInput{
		Name: aws.String(target.(string)),
	}

	log.Printf("[DEBUG] Reading QLDB Ledger: %s", req)
	resp, err := conn.DescribeLedger(req)

	if err != nil {
		return fmt.Errorf("Error describing ledger: %s", err)
	}

	d.SetId(time.Now().UTC().String())
	d.Set("arn", resp.Arn)
	// This is hardcoded because AWS SDK does not allow returning Permissions Mode
	d.Set("permissions_mode", qldb.PermissionsModeAllowAll)
	d.Set("deletion_protection", resp.DeletionProtection)

	return nil
}
