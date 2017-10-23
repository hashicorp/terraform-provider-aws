package aws

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/athena"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsAthenaDatabase() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsAthenaDatabaseCreate,
		Read:   resourceAwsAthenaDatabaseRead,
		//Update: resourceAwsAthenaDatabaseUpdate,
		Delete: resourceAwsAthenaDatabaseDelete,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"bucket": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceAwsAthenaDatabaseCreate(d *schema.ResourceData, meta interface{}) error {
	athenaconn := meta.(*AWSClient).athenaconn

	athenainput := &athena.StartQueryExecutionInput{
		QueryString: aws.String(createDatabaseQueryString(d.Get("name").(string))),
		ResultConfiguration: &athena.ResultConfiguration{
			OutputLocation: aws.String("s3://" + d.Get("bucket").(string)),
		},
	}

	athenaresp, err := athenaconn.StartQueryExecution(athenainput)
	if err != nil {
		return err
	}

	if err := checkCreateDatabaseQueryExecution(*athenaresp.QueryExecutionId, d, meta); err != nil {
		return err
	}
	d.SetId(d.Get("name").(string))
	return resourceAwsAthenaDatabaseRead(d, meta)
}

func resourceAwsAthenaDatabaseRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).athenaconn

	bucket := d.Get("bucket").(string)
	input := &athena.StartQueryExecutionInput{
		QueryString: aws.String(showDatabaseQueryString()),
		ResultConfiguration: &athena.ResultConfiguration{
			OutputLocation: aws.String("s3://" + bucket),
		},
	}

	resp, err := conn.StartQueryExecution(input)
	if err != nil {
		return err
	}

	if err := checkShowDatabaseQueryExecution(*resp.QueryExecutionId, d, meta); err != nil {
		return err
	}
	return nil
}

func resourceAwsAthenaDatabaseUpdate(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func resourceAwsAthenaDatabaseDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).athenaconn

	bucket := d.Get("bucket").(string)
	input := &athena.StartQueryExecutionInput{
		QueryString: aws.String(dropDatabaseQueryString(d.Get("name").(string))),
		ResultConfiguration: &athena.ResultConfiguration{
			OutputLocation: aws.String("s3://" + bucket),
		},
	}

	resp, err := conn.StartQueryExecution(input)
	if err != nil {
		return err
	}

	if err := checkDropDatabaseQueryExecution(*resp.QueryExecutionId, d, meta); err != nil {
		return err
	}
	return nil
}

func createDatabaseQueryString(databaseName string) string {
	return fmt.Sprintf("create database %s;", databaseName)
}

func checkCreateDatabaseQueryExecution(qeid string, d *schema.ResourceData, meta interface{}) error {
	rs, err := queryExecutionResult(qeid, meta)
	if err != nil {
		return err
	}
	if len(rs.Rows) != 0 {
		return fmt.Errorf("[ERROR] Athena create database, unexpected query result: %s", flattenAthenaResultSet(rs))
	}
	return nil
}

func showDatabaseQueryString() string {
	return fmt.Sprint("show databases;")
}

func checkShowDatabaseQueryExecution(qeid string, d *schema.ResourceData, meta interface{}) error {
	rs, err := queryExecutionResult(qeid, meta)
	if err != nil {
		return err
	}
	found := false
	dbName := d.Get("name").(string)
	for _, row := range rs.Rows {
		for _, datum := range row.Data {
			if *datum.VarCharValue == dbName {
				found = true
			}
		}
	}
	if !found {
		return fmt.Errorf("[ERROR] Athena not found database: %s, query result: %s", dbName, flattenAthenaResultSet(rs))
	}
	return nil
}

func dropDatabaseQueryString(databaseName string) string {
	return fmt.Sprintf("drop database %s;", databaseName)
}

func checkDropDatabaseQueryExecution(qeid string, d *schema.ResourceData, meta interface{}) error {
	rs, err := queryExecutionResult(qeid, meta)
	if err != nil {
		return err
	}
	if len(rs.Rows) != 0 {
		return fmt.Errorf("[ERROR] Athena drop database, unexpected query result: %s", flattenAthenaResultSet(rs))
	}
	return nil
}

func queryExecutionResult(qeid string, meta interface{}) (*athena.ResultSet, error) {
	conn := meta.(*AWSClient).athenaconn

	input := &athena.GetQueryExecutionInput{
		QueryExecutionId: aws.String(qeid),
	}
	err := resource.Retry(5*time.Minute, func() *resource.RetryError {
		out, err := conn.GetQueryExecution(input)
		if err != nil {
			return resource.NonRetryableError(err)
		}
		switch *out.QueryExecution.Status.State {
		case athena.QueryExecutionStateQueued, athena.QueryExecutionStateRunning:
			log.Printf("[DEBUG] Executing Athena Query...")
			return resource.RetryableError(nil)
		case athena.QueryExecutionStateSucceeded:
			return nil
		case athena.QueryExecutionStateFailed:
			return resource.NonRetryableError(fmt.Errorf("[Error] QueryExecution Failed"))
		case athena.QueryExecutionStateCancelled:
			return resource.NonRetryableError(fmt.Errorf("[Error] QueryExecution Canceled"))
		default:
			return resource.NonRetryableError(fmt.Errorf("[Error] Unexpected QueryExecution State: %s", *out.QueryExecution.Status.State))
		}
	})
	if err != nil {
		return nil, err
	}

	qrinput := &athena.GetQueryResultsInput{
		QueryExecutionId: aws.String(qeid),
	}
	resp, err := conn.GetQueryResults(qrinput)
	if err != nil {
		return nil, err
	}
	return resp.ResultSet, nil
}

func flattenAthenaResultSet(rs *athena.ResultSet) string {
	ss := make([]string, 0)
	for _, row := range rs.Rows {
		for _, datum := range row.Data {
			ss = append(ss, *datum.VarCharValue)
		}
	}
	return strings.Join(ss, "\n")
}
