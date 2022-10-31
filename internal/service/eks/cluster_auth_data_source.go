package eks

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
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
	generator, err := NewGenerator(false, false)
	if err != nil {
		return fmt.Errorf("error getting token generator: %w", err)
	}
	toke, err := generator.GetWithSTS(name, conn)
	if err != nil {
		return fmt.Errorf("error getting token: %w", err)
	}

	var identifier string
	if v, ok := d.GetOk("id"); ok {
		identifier = v.(string)
	} else {
		identifier = d.Get("name").(string)
	}

	d.SetId(identifier)
	d.Set("token", toke.Token)

	return nil
}
