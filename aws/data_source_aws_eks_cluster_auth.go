package aws

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/service/eks/token"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

func DataSourceClusterAuth() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceClusterAuthRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.NoZeroValues,
			},

			"token": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
		},
	}
}

func dataSourceClusterAuthRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).STSConn
	name := d.Get("name").(string)
	generator, err := token.NewGenerator(false, false)
	if err != nil {
		return fmt.Errorf("error getting token generator: %w", err)
	}
	token, err := generator.GetWithSTS(name, conn)
	if err != nil {
		return fmt.Errorf("error getting token: %w", err)
	}

	d.SetId(name)
	d.Set("token", token.Token)

	return nil
}
