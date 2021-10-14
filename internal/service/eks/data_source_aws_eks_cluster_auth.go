package aws

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/service/eks/token"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	tfeks "github.com/hashicorp/terraform-provider-aws/internal/service/eks"
	tfeks "github.com/hashicorp/terraform-provider-aws/internal/service/eks"
	tfeks "github.com/hashicorp/terraform-provider-aws/internal/service/eks"
	tfeks "github.com/hashicorp/terraform-provider-aws/internal/service/eks"
	tfeks "github.com/hashicorp/terraform-provider-aws/internal/service/eks"
	tfeks "github.com/hashicorp/terraform-provider-aws/internal/service/eks"
	tfeks "github.com/hashicorp/terraform-provider-aws/internal/service/eks"
	tfeks "github.com/hashicorp/terraform-provider-aws/internal/service/eks"
	tfeks "github.com/hashicorp/terraform-provider-aws/internal/service/eks"
	tfeks "github.com/hashicorp/terraform-provider-aws/internal/service/eks"
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
	generator, err := tfeks.NewGenerator(false, false)
	if err != nil {
		return fmt.Errorf("error getting token generator: %w", err)
	}
	token, err := generator.GetWithSTS(name, conn)
	if err != nil {
		return fmt.Errorf("error getting token: %w", err)
	}

	d.SetId(name)
	d.Set("token", tfeks.Token)

	return nil
}
