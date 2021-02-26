package aws

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"strings"
)

func awsCloudWatchQueryDefinition() *schema.Resource {
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
	d.SetId(*r.QueryDefinitionId)
	d.Set("query_definition_id", *r.QueryDefinitionId)
	return resourceAwsCloudWatchQueryDefinitionRead(c, d, meta)
}

func getAwsCloudWatchQueryDefinitionInput(d *schema.ResourceData) *cloudwatchlogs.PutQueryDefinitionInput {
	name := d.Get("name").(string)
	logGroups := d.Get("log_groups").([]interface{})
	var lgs []*string

	for _, group := range logGroups {
		l := group.(string)
		lgs = append(lgs, &l)
	}

	query := d.Get("query").(string)
	return &cloudwatchlogs.PutQueryDefinitionInput{
		Name:          &name,
		LogGroupNames: lgs,
		QueryString:   &query,
	}
}

func resourceAwsCloudWatchQueryDefinitionRead(c context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).cloudwatchlogsconn
	name := d.Get("name").(string)
	input := &cloudwatchlogs.DescribeQueryDefinitionsInput{
		QueryDefinitionNamePrefix: &name,
	}

	qdResp, err := conn.DescribeQueryDefinitions(input)
	if err != nil {
		return diag.FromErr(err)
	}

	for _, qd := range qdResp.QueryDefinitions {
		if *qd.QueryDefinitionId == d.Id() {
			if err := d.Set("name", *qd.Name); err != nil {
				return diag.FromErr(err)
			}
			if err := d.Set("query", *qd.QueryString); err != nil {
				return diag.FromErr(err)
			}

			var logGroups []string
			for _, lg := range qd.LogGroupNames {
				logGroups = append(logGroups, *lg)
			}
			if err := d.Set("log_groups", logGroups); err != nil {
				return diag.FromErr(err)
			}
			return nil
		}
	}

	// disappears case
	d.SetId("")
	return nil
}

func resourceAwsCloudWatchQueryDefinitionUpdate(c context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).cloudwatchlogsconn
	queryId := d.Get("query_definition_id").(string)

	parms := getAwsCloudWatchQueryDefinitionInput(d)
	parms.QueryDefinitionId = &queryId
	_, err := conn.PutQueryDefinition(parms)
	if err != nil {
		return diag.FromErr(err)
	}
	return resourceAwsCloudWatchQueryDefinitionRead(c, d, meta)
}

func resourceAwsCloudWatchQueryDefinitionDelete(c context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).cloudwatchlogsconn
	queryId := d.Get("query_definition_id").(string)

	parms := &cloudwatchlogs.DeleteQueryDefinitionInput{QueryDefinitionId: &queryId}
	_, err := conn.DeleteQueryDefinition(parms)
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

	d.Set("name", name)
	d.SetId(id)
	return []*schema.ResourceData{d}, nil
}

func parseImportFields(id string) (string, string, error) {
	tokens := strings.Split(id, "_")
	if len(tokens) != 2 || tokens[0] == "" || tokens[1] == "" {
		return "", "", fmt.Errorf(`failed parsing resource ID: did not contain the correct number of fields for import`)
	}

	return tokens[0], tokens[1], nil
}
