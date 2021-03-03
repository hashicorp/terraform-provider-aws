package aws

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/cloudwatchlogs/finder"
	"log"
	"strings"
)

func resourceAwsCloudWatchQueryDefinition() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceAwsCloudWatchQueryDefinitionCreate,
		ReadContext:   resourceAwsCloudWatchQueryDefinitionRead,
		UpdateContext: resourceAwsCloudWatchQueryDefinitionUpdate,
		DeleteContext: resourceAwsCloudWatchQueryDefinitionDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceAwsCloudWatchQueryDefinitionImport,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"query": {
				Type:     schema.TypeString,
				Required: true,
			},
			"query_definition_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"log_groups": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func resourceAwsCloudWatchQueryDefinitionCreate(c context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).cloudwatchlogsconn
	params := getAwsCloudWatchQueryDefinitionInput(d)
	r, err := conn.PutQueryDefinition(params)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(aws.StringValue(r.QueryDefinitionId))
	if err := d.Set("query_definition_id", aws.StringValue(r.QueryDefinitionId)); err != nil {
		return diag.FromErr(err)
	}
	return resourceAwsCloudWatchQueryDefinitionRead(c, d, meta)
}

func getAwsCloudWatchQueryDefinitionInput(d *schema.ResourceData) *cloudwatchlogs.PutQueryDefinitionInput {
	name := d.Get("name").(string)
	logGroups := d.Get("log_groups").([]interface{})
	var lgs []*string

	for _, group := range logGroups {
		l := group.(string)
		lgs = append(lgs, aws.String(l))
	}

	query := d.Get("query").(string)
	return &cloudwatchlogs.PutQueryDefinitionInput{
		Name:          aws.String(name),
		LogGroupNames: lgs,
		QueryString:   aws.String(query),
	}
}

func resourceAwsCloudWatchQueryDefinitionRead(c context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).cloudwatchlogsconn
	name := d.Get("name").(string)
	id := d.Id()

	result, err := finder.QueryDefinition(conn, name, id)

	if err != nil {
		return diag.FromErr(err)
	}

	if result == nil {
		log.Printf("[WARN] cloudwatch query definition (%s) not found, removing from state", d.Id())
		d.SetId("")
	} else {
		if err := d.Set("query", aws.StringValue(result.QueryString)); err != nil {
			return diag.FromErr(err)
		}
		if err := d.Set("query_definition_id", aws.StringValue(result.QueryDefinitionId)); err != nil {
			return diag.FromErr(err)
		}
		if err := d.Set("log_groups", aws.StringValueSlice(result.LogGroupNames)); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

func resourceAwsCloudWatchQueryDefinitionUpdate(c context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).cloudwatchlogsconn
	queryId := d.Get("query_definition_id").(string)

	parms := getAwsCloudWatchQueryDefinitionInput(d)
	parms.QueryDefinitionId = aws.String(queryId)
	_, err := conn.PutQueryDefinition(parms)
	if err != nil {
		return diag.FromErr(err)
	}
	return resourceAwsCloudWatchQueryDefinitionRead(c, d, meta)
}

func resourceAwsCloudWatchQueryDefinitionDelete(c context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).cloudwatchlogsconn
	queryId := d.Get("query_definition_id").(string)

	params := &cloudwatchlogs.DeleteQueryDefinitionInput{QueryDefinitionId: aws.String(queryId)}
	_, err := conn.DeleteQueryDefinition(params)
	if err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func resourceAwsCloudWatchQueryDefinitionImport(c context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	name, id, err := parseImportFields(d.Id())
	if err != nil {
		return nil, err
	}

	if err := d.Set("name", name); err != nil {
		return nil, err
	}

	d.SetId(id)
	return []*schema.ResourceData{d}, nil
}

func parseImportFields(id string) (string, string, error) {
	// having underscores in the query name is valid. The last occurrence of the underscore should separate the ID
	// from the name of the query.
	malformed := "resource ID did not contain correct number of fields for import"
	lastUnd := strings.LastIndexByte(id, '_')

	// if there isn't an underscore, the import is malformed.
	if lastUnd < 0 {
		return "", "", fmt.Errorf(malformed)
	}

	name, qId := id[0:lastUnd], id[lastUnd+1:]

	// if either name or ID are the empty string, the import is malformed.
	if name == "" || qId == "" {
		return "", "", fmt.Errorf(malformed)
	}

	return name, qId, nil
}
