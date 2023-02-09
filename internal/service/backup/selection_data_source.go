package backup

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/backup"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

func DataSourceSelection() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceSelectionRead,

		Schema: map[string]*schema.Schema{
			"plan_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"selection_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"iam_role_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"resources": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceSelectionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).BackupConn()

	input := &backup.GetBackupSelectionInput{
		BackupPlanId: aws.String(d.Get("plan_id").(string)),
		SelectionId:  aws.String(d.Get("selection_id").(string)),
	}

	resp, err := conn.GetBackupSelectionWithContext(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "Error getting Backup Selection: %s", err)
	}

	d.SetId(aws.StringValue(resp.SelectionId))
	d.Set("iam_role_arn", resp.BackupSelection.IamRoleArn)
	d.Set("name", resp.BackupSelection.SelectionName)

	if resp.BackupSelection.Resources != nil {
		if err := d.Set("resources", aws.StringValueSlice(resp.BackupSelection.Resources)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting resources: %s", err)
		}
	}

	return diags
}
