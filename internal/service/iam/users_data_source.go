package iam

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

func DataSourceUsers() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceUsersRead,
		Schema: map[string]*schema.Schema{
			"arns": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"name_regex": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringIsValidRegExp,
			},
			"names": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"path_prefix": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func dataSourceUsersRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMConn()

	nameRegex := d.Get("name_regex").(string)
	pathPrefix := d.Get("path_prefix").(string)

	results, err := FindUsers(ctx, conn, nameRegex, pathPrefix)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IAM users: %s", err)
	}

	d.SetId(meta.(*conns.AWSClient).Region)

	var arns, names []string

	for _, r := range results {
		names = append(names, aws.StringValue(r.UserName))
		arns = append(arns, aws.StringValue(r.Arn))
	}

	if err := d.Set("names", names); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting names: %s", err)
	}

	if err := d.Set("arns", arns); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting arns: %s", err)
	}

	return diags
}
