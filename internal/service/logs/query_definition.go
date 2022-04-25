package logs

import (
	"context"
	"fmt"
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceQueryDefinition() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceQueryDefinitionCreate,
		ReadWithoutTimeout:   resourceQueryDefinitionRead,
		UpdateWithoutTimeout: resourceQueryDefinitionUpdate,
		DeleteWithoutTimeout: resourceQueryDefinitionDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceQueryDefinitionImport,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 255),
					validation.StringMatch(regexp.MustCompile(`^([^:*\/]+\/?)*[^:*\/]+$`), "cannot contain a colon or asterisk and cannot start or end with a slash"),
				),
			},
			"query_string": {
				Type:     schema.TypeString,
				Required: true,
			},
			"query_definition_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"log_group_names": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validLogGroupName,
				},
			},
		},
	}
}

func resourceQueryDefinitionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).LogsConn

	input := getQueryDefinitionInput(d)
	r, err := conn.PutQueryDefinitionWithContext(ctx, input)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(aws.StringValue(r.QueryDefinitionId))
	d.Set("query_definition_id", r.QueryDefinitionId) // TODO: is this needed?

	return resourceQueryDefinitionRead(ctx, d, meta)
}

func getQueryDefinitionInput(d *schema.ResourceData) *cloudwatchlogs.PutQueryDefinitionInput {
	result := &cloudwatchlogs.PutQueryDefinitionInput{
		Name:          aws.String(d.Get("name").(string)),
		LogGroupNames: flex.ExpandStringList(d.Get("log_group_names").([]interface{})),
		QueryString:   aws.String(d.Get("query_string").(string)),
	}

	if d.Id() != "" {
		result.QueryDefinitionId = aws.String(d.Id())
	}

	return result
}

func resourceQueryDefinitionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).LogsConn

	result, err := FindQueryDefinition(ctx, conn, d.Get("name").(string), d.Id())

	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading CloudWatch query definition (%s): %w", d.Id(), err))
	}

	if result == nil {
		log.Printf("[WARN] CloudWatch query definition (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	d.Set("name", result.Name)
	d.Set("query_string", result.QueryString)
	d.Set("query_definition_id", result.QueryDefinitionId)
	if err := d.Set("log_group_names", aws.StringValueSlice(result.LogGroupNames)); err != nil {
		return diag.FromErr(fmt.Errorf("error setting log_group_names: %w", err))
	}

	return nil
}

func resourceQueryDefinitionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).LogsConn

	_, err := conn.PutQueryDefinitionWithContext(ctx, getQueryDefinitionInput(d))
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceQueryDefinitionRead(ctx, d, meta)
}

func resourceQueryDefinitionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).LogsConn

	input := &cloudwatchlogs.DeleteQueryDefinitionInput{
		QueryDefinitionId: aws.String(d.Id()),
	}
	_, err := conn.DeleteQueryDefinitionWithContext(ctx, input)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceQueryDefinitionImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	arn, err := arn.Parse(d.Id())
	if err != nil {
		return nil, fmt.Errorf("unexpected format for ID (%s), expected a CloudWatch query definition ARN", d.Id())
	}

	if arn.Service != cloudwatchlogs.ServiceName {
		return nil, fmt.Errorf("unexpected format for ID (%s), expected a CloudWatch query definition ARN", d.Id())
	}

	matcher := regexp.MustCompile("^query-definition:(" + verify.UUIDRegexPattern + ")$")
	matches := matcher.FindStringSubmatch(arn.Resource)
	if len(matches) != 2 {
		return nil, fmt.Errorf("unexpected format for ID (%s), expected a CloudWatch query definition ARN", d.Id())
	}

	d.SetId(matches[1])

	return []*schema.ResourceData{d}, nil
}
