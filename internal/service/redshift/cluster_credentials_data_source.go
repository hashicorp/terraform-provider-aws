package redshift

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/redshift"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
)

func DataSourceClusterCredentials() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceClusterCredentialsRead,

		Schema: map[string]*schema.Schema{
			"auto_create": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"cluster_identifier": {
				Type:     schema.TypeString,
				Required: true,
			},
			"db_name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"db_user": {
				Type:     schema.TypeString,
				Required: true,
			},
			"db_password": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"db_groups": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
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

func dataSourceClusterCredentialsRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RedshiftConn

	clusterId := d.Get("cluster_identifier").(string)
	input := &redshift.GetClusterCredentialsInput{
		ClusterIdentifier: aws.String(clusterId),
		DbUser:            aws.String(d.Get("db_user").(string)),
		AutoCreate:        aws.Bool(d.Get("auto_create").(bool)),
		DurationSeconds:   aws.Int64(int64(d.Get("duration_seconds").(int))),
	}

	if v, ok := d.GetOk("db_name"); ok {
		input.DbName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("db_groups"); ok && v.(*schema.Set).Len() > 0 {
		input.DbGroups = flex.ExpandStringSet(v.(*schema.Set))
	}

	creds, err := conn.GetClusterCredentials(input)

	if err != nil {
		return fmt.Errorf("reading Redshift Cluster Credentials for Cluster (%s): %w", clusterId, err)
	}

	d.SetId(clusterId)

	d.Set("db_user", creds.DbUser)
	d.Set("db_password", creds.DbPassword)
	d.Set("expiration", aws.TimeValue(creds.Expiration).Format(time.RFC3339))

	return nil
}
