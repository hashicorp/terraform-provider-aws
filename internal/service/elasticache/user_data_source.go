package elasticache

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func DataSourceUser() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceUserRead,

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

func dataSourceUserRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ElastiCacheConn()

	user, err := FindUserByID(ctx, conn, d.Get("user_id").(string))
	if tfresource.NotFound(err) {
		return sdkdiag.AppendErrorf(diags, "reading ElastiCache Cache Cluster (%s): Not found. Please change your search criteria and try again: %s", d.Get("user_id").(string), err)
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ElastiCache Cache Cluster (%s): %s", d.Get("user_id").(string), err)
	}

	d.SetId(aws.StringValue(user.UserId))

	d.Set("access_string", user.AccessString)
	d.Set("engine", user.Engine)
	d.Set("user_id", user.UserId)
	d.Set("user_name", user.UserName)

	return diags
}
