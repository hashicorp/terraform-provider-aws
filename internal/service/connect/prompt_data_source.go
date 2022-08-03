package connect

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/connect"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func DataSourcePrompt() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourcePromptRead,
		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"instance_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"prompt_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourcePromptRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ConnectConn

	instanceID := d.Get("instance_id").(string)
	name := d.Get("name").(string)

	promptSummary, err := dataSourceGetPromptSummaryByName(ctx, conn, instanceID, name)

	if err != nil {
		return diag.FromErr(fmt.Errorf("error finding Connect Prompt Summary by name (%s): %w", name, err))
	}

	if promptSummary == nil {
		return diag.FromErr(fmt.Errorf("error finding Connect Prompt Summary by name (%s): not found", name))
	}

	d.Set("arn", promptSummary.Arn)
	d.Set("instance_id", instanceID)
	d.Set("prompt_id", promptSummary.Id)
	d.Set("name", promptSummary.Name)

	d.SetId(fmt.Sprintf("%s:%s", instanceID, aws.StringValue(promptSummary.Id)))

	return nil
}

func dataSourceGetPromptSummaryByName(ctx context.Context, conn *connect.Connect, instanceID, name string) (*connect.PromptSummary, error) {
	var result *connect.PromptSummary

	input := &connect.ListPromptsInput{
		InstanceId: aws.String(instanceID),
		MaxResults: aws.Int64(ListPromptsMaxResults),
	}

	err := conn.ListPromptsPagesWithContext(ctx, input, func(page *connect.ListPromptsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, cf := range page.PromptSummaryList {
			if cf == nil {
				continue
			}

			if aws.StringValue(cf.Name) == name {
				result = cf
				return false
			}
		}

		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}
