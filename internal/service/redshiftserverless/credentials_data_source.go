package redshiftserverless

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/redshiftserverless"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func DataSourceCredentials() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceCredentialsRead,

		Schema: map[string]*schema.Schema{
			"workgroup_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"db_name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"db_password": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
			"db_user": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"duration_seconds": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      900,
				ValidateFunc: validation.IntBetween(900, 3600),
			},
			"expiration": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceCredentialsRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RedshiftServerlessConn

	workgroupName := d.Get("workgroup_name").(string)
	input := &redshiftserverless.GetCredentialsInput{
		WorkgroupName:   aws.String(workgroupName),
		DurationSeconds: aws.Int64(int64(d.Get("duration_seconds").(int))),
	}

	if v, ok := d.GetOk("db_name"); ok {
		input.DbName = aws.String(v.(string))
	}

	creds, err := conn.GetCredentials(input)

	if err != nil {
		return fmt.Errorf("reading Redshift Serverless Credentials for Workgroup (%s): %w", workgroupName, err)
	}

	d.SetId(workgroupName)

	d.Set("db_password", creds.DbPassword)
	d.Set("db_user", creds.DbUser)
	d.Set("expiration", aws.TimeValue(creds.Expiration).Format(time.RFC3339))

	return nil
}
