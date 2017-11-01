package aws

import (
	"errors"

	"github.com/hashicorp/terraform/helper/schema"
)

func dataSourceAwsElasticacheReplicationGroup() *schema.Resource {
	return &schema.Resource{
		Read:   dataSourceAwsElasticacheReplicationGroupRead,
		Schema: map[string]*schema.Schema{},
	}
}

func dataSourceAwsElasticacheReplicationGroupRead(d *schema.ResourceData, meta interface{}) error {
	return errors.New("error")
}
