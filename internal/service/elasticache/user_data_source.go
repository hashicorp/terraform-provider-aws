package elasticache

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elasticache"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func DataSourceUser() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceUserRead,

		Schema: map[string]*schema.Schema{
			"access_string": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"engine": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"no_password_required": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"passwords": {
				Type:      schema.TypeSet,
				Optional:  true,
				Elem:      &schema.Schema{Type: schema.TypeString},
				Set:       schema.HashString,
				Sensitive: true,
			},
			"user_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"user_name": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func dataSourceUserRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ElastiCacheConn

	params := &elasticache.DescribeUsersInput{
		UserId: aws.String(d.Get("user_id").(string)),
	}

	log.Printf("[DEBUG] Reading ElastiCache User: %s", params)
	response, err := conn.DescribeUsers(params)
	if err != nil {
		return err
	}

	if len(response.Users) != 1 {
		return fmt.Errorf("[ERROR] Query returned wrong number of results. Please change your search criteria and try again.")
	}

	user := response.Users[0]

	d.SetId(aws.StringValue(user.UserId))

	d.Set("access_string", user.AccessString)
	d.Set("engine", user.Engine)
	d.Set("user_id", user.UserId)
	d.Set("user_name", user.UserName)

	return nil
}
