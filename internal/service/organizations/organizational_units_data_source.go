package organizations

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/organizations"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

func DataSourceOrganizationalUnits() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceOrganizationalUnitsRead,

		Schema: map[string]*schema.Schema{
			"parent_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"children": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"arn": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceOrganizationalUnitsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OrganizationsConn()

	parent_id := d.Get("parent_id").(string)

	params := &organizations.ListOrganizationalUnitsForParentInput{
		ParentId: aws.String(parent_id),
	}

	var children []*organizations.OrganizationalUnit

	err := conn.ListOrganizationalUnitsForParentPagesWithContext(ctx, params,
		func(page *organizations.ListOrganizationalUnitsForParentOutput, lastPage bool) bool {
			children = append(children, page.OrganizationalUnits...)

			return !lastPage
		})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing Organizations Organization Units for parent (%s): %s", parent_id, err)
	}

	d.SetId(parent_id)

	if err := d.Set("children", FlattenOrganizationalUnits(children)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting children: %s", err)
	}

	return diags
}

func FlattenOrganizationalUnits(ous []*organizations.OrganizationalUnit) []map[string]interface{} {
	if len(ous) == 0 {
		return nil
	}
	var result []map[string]interface{}
	for _, ou := range ous {
		result = append(result, map[string]interface{}{
			"arn":  aws.StringValue(ou.Arn),
			"id":   aws.StringValue(ou.Id),
			"name": aws.StringValue(ou.Name),
		})
	}
	return result
}
