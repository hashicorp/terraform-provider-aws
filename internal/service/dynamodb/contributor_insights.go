package dynamodb

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func ResourceContributorInsights() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceContributorInsightsCreate,
		ReadWithoutTimeout:   resourceContributorInsightsRead,
		DeleteWithoutTimeout: resourceContributorInsightsDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"index_name": {
				Type:     schema.TypeString,
				ForceNew: true,
				Optional: true,
			},
			"table_name": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceContributorInsightsCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).DynamoDBConn

	in := &dynamodb.UpdateContributorInsightsInput{
		ContributorInsightsAction: aws.String(dynamodb.ContributorInsightsActionEnable),
	}

	if v, ok := d.GetOk("table_name"); ok {
		in.TableName = aws.String(v.(string))
	}

	var indexName string
	if v, ok := d.GetOk("index_name"); ok {
		in.IndexName = aws.String(v.(string))
		indexName = v.(string)
	}

	out, err := conn.UpdateContributorInsightsWithContext(ctx, in)
	if err != nil {
		return diag.Errorf("creating dynamodb ContributorInsights for table (%s): %s", d.Get("table_name").(string), err)
	}

	id := encodeContributorInsightsID(aws.StringValue(out.TableName), indexName, meta.(*conns.AWSClient).AccountID)
	d.SetId(id)

	if err := waitContributorInsightsCreated(ctx, conn, aws.StringValue(out.TableName), indexName, d.Timeout(schema.TimeoutCreate)); err != nil {
		return diag.Errorf("waiting for dynamodb ContributorInsights (%s) create: %s", d.Id(), err)
	}

	return resourceContributorInsightsRead(ctx, d, meta)
}

func resourceContributorInsightsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).DynamoDBConn

	var indexName string
	if v, ok := d.GetOk("index_name"); ok {
		indexName = v.(string)
	}

	tableName, indexName, _, err := decodeContributorInsightsID(d.Id())
	if err != nil {
		return diag.Errorf("unable to decode ContributorInsights ID (%s): %s", d.Id(), err)
	}

	out, err := FindContributorInsights(ctx, conn, tableName, indexName)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] DynamoDB ContributorInsights (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading DynamoDB ContributorInsights (%s): %s", d.Id(), err)
	}

	d.Set("index_name", out.IndexName)
	d.Set("table_name", out.TableName)
	d.Set("status", out.ContributorInsightsStatus)

	return nil
}

func resourceContributorInsightsDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).DynamoDBConn

	log.Printf("[INFO] Deleting DynamoDB ContributorInsights %s", d.Id())

	tableName, indexName, _, err := decodeContributorInsightsID(d.Id())
	if err != nil {
		return diag.Errorf("unable to decode DynamoDB ContributorInsights ID (%s): %s", d.Id(), err)
	}

	in := &dynamodb.UpdateContributorInsightsInput{
		ContributorInsightsAction: aws.String(dynamodb.ContributorInsightsActionDisable),
		TableName:                 aws.String(tableName),
	}

	if indexName != "" {
		in.IndexName = aws.String(indexName)
	}

	_, err = conn.UpdateContributorInsightsWithContext(ctx, in)

	if tfawserr.ErrCodeEquals(err, dynamodb.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return diag.Errorf("deleting DynamoDB ContributorInsights (%s): %s", d.Id(), err)
	}

	if err := waitContributorInsightsDeleted(ctx, conn, tableName, indexName, d.Timeout(schema.TimeoutDelete)); err != nil {
		return diag.Errorf("waiting for DynamoDB ContributorInsights (%s) to be deleted: %s", d.Id(), err)
	}

	return nil
}

func encodeContributorInsightsID(tableName, indexName, accountID string) string {
	if indexName != "" {
		return fmt.Sprintf("%s-%s/%s", tableName, indexName, accountID)
	}

	return fmt.Sprintf("%s/%s", tableName, accountID)
}

func decodeContributorInsightsID(id string) (string, string, string, error) {
	idParts := strings.Split(id, "/")
	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		return "", "", "", fmt.Errorf("expected ID in the form of table_name/account_id, given: %q", id)
	}

	var tableName, indexName, accountID string

	tableName = idParts[0]
	if strings.Contains(tableName, "-") {
		parts := strings.Split(tableName, "-")
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			return "", "", "", fmt.Errorf("expected ID in the form of table_name-index_name, given: %q", id)
		}
		tableName = parts[0]
		indexName = parts[1]
	}

	accountID = idParts[1]

	return tableName, indexName, accountID, nil
}
