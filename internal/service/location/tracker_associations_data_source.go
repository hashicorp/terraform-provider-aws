package location

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/locationservice"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func DataSourceTrackerAssociations() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceTrackerAssociationsRead,

		Schema: map[string]*schema.Schema{
			"consumer_arns": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"tracker_name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 100),
			},
		},
	}
}

const (
	DSNameTrackerAssociations = "Tracker Associations Data Source"
)

func dataSourceTrackerAssociationsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).LocationConn()

	name := d.Get("tracker_name").(string)

	in := &locationservice.ListTrackerConsumersInput{
		TrackerName: aws.String(name),
	}

	var arns []string

	err := conn.ListTrackerConsumersPagesWithContext(ctx, in, func(page *locationservice.ListTrackerConsumersOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, arn := range page.ConsumerArns {
			if arn == nil {
				continue
			}

			arns = append(arns, aws.StringValue(arn))
		}

		return !lastPage
	})

	if err != nil {
		return create.DiagError(names.Location, create.ErrActionReading, DSNameTrackerAssociations, name, err)
	}

	d.SetId(meta.(*conns.AWSClient).Region)
	d.Set("consumer_arns", arns)

	return nil
}
