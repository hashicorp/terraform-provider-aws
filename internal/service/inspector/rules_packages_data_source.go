package inspector

import (
	"context"
	"sort"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/inspector"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

func DataSourceRulesPackages() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceRulesPackagesRead,

		Schema: map[string]*schema.Schema{
			"arns": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceRulesPackagesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).InspectorConn()

	output, err := findRulesPackageARNs(ctx, conn)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Inspector Rules Packages: %s", err)
	}

	arns := aws.StringValueSlice(output)
	sort.Strings(arns)

	d.SetId(meta.(*conns.AWSClient).Region)
	d.Set("arns", arns)

	return diags
}

func findRulesPackageARNs(ctx context.Context, conn *inspector.Inspector) ([]*string, error) {
	input := &inspector.ListRulesPackagesInput{}
	var output []*string

	err := conn.ListRulesPackagesPagesWithContext(ctx, input, func(page *inspector.ListRulesPackagesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.RulesPackageArns {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	return output, nil
}
