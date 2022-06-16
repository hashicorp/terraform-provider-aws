package athena

import (
	"bytes"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/athena"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func ResourceDatabase() *schema.Resource {
	return &schema.Resource{
		Create: resourceDatabaseCreate,
		Read:   resourceDatabaseRead,
		Update: schema.Noop,
		Delete: resourceDatabaseDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"acl_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"s3_acl_option": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(athena.S3AclOption_Values(), false),
							ForceNew:     true,
						},
					},
				},
			},
			"bucket": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"comment": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"encryption_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"encryption_option": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(athena.EncryptionOption_Values(), false),
							ForceNew:     true,
						},
						"kms_key": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
					},
				},
			},
			"expected_bucket_owner": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"force_destroy": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile("^[_a-z0-9]+$"), "must be lowercase letters, numbers, or underscore ('_')"),
			},
			"properties": {
				Type:     schema.TypeMap,
				Optional: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func resourceDatabaseCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).AthenaConn

	name := d.Get("name").(string)
	var queryString bytes.Buffer

	createStmt := fmt.Sprintf("create database `%s`", name)
	queryString.WriteString(createStmt)

	if v, ok := d.GetOk("comment"); ok && v.(string) != "" {
		commentStmt := fmt.Sprintf(" comment '%s'", strings.Replace(v.(string), "'", "\\'", -1))
		queryString.WriteString(commentStmt)
	}

	if v, ok := d.GetOk("properties"); ok && len(v.(map[string]interface{})) > 0 {
		var props []string
		for k, v := range v.(map[string]interface{}) {
			prop := fmt.Sprintf(" '%[1]s' = '%[2]s' ", k, v.(string))
			props = append(props, prop)
		}

		propStmt := fmt.Sprintf(" WITH DBPROPERTIES(%s)", strings.Join(props, ","))
		queryString.WriteString(propStmt)
	}

	queryString.WriteString(";")

	input := &athena.StartQueryExecutionInput{
		QueryString:         aws.String(queryString.String()),
		ResultConfiguration: expandResultConfiguration(d),
	}

	resp, err := conn.StartQueryExecution(input)

	if err != nil {
		return fmt.Errorf("error starting Athena Database (%s) query execution: %w", name, err)
	}

	if err := executeAndExpectNoRows(*resp.QueryExecutionId, "create", conn); err != nil {
		return err
	}

	d.SetId(name)

	return resourceDatabaseRead(d, meta)
}

func resourceDatabaseRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).AthenaConn

	input := &athena.GetDatabaseInput{
		DatabaseName: aws.String(d.Id()),
		CatalogName:  aws.String("AwsDataCatalog"),
	}
	res, err := conn.GetDatabase(input)

	if tfawserr.ErrMessageContains(err, athena.ErrCodeMetadataException, "not found") && !d.IsNewResource() {
		log.Printf("[WARN] Athena Database (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading Athena Database (%s): %w", d.Id(), err)
	}

	db := res.Database

	d.Set("name", db.Name)
	d.Set("comment", db.Description)
	d.Set("properties", db.Parameters)

	return nil
}

func resourceDatabaseDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).AthenaConn

	queryString := fmt.Sprintf("drop database `%s`", d.Id())
	if d.Get("force_destroy").(bool) {
		queryString += " cascade"
	}
	queryString += ";"

	input := &athena.StartQueryExecutionInput{
		QueryString:         aws.String(queryString),
		ResultConfiguration: expandResultConfiguration(d),
	}

	resp, err := conn.StartQueryExecution(input)
	if err != nil {
		return err
	}

	if err := executeAndExpectNoRows(*resp.QueryExecutionId, "delete", conn); err != nil {
		return err
	}

	return nil
}

func expandResultConfiguration(d *schema.ResourceData) *athena.ResultConfiguration {

	resultConfig := &athena.ResultConfiguration{
		OutputLocation:          aws.String("s3://" + d.Get("bucket").(string)),
		EncryptionConfiguration: expandResultConfigurationEncryptionConfig(d.Get("encryption_configuration").([]interface{})),
	}

	if v, ok := d.GetOk("expected_bucket_owner"); ok {
		resultConfig.ExpectedBucketOwner = aws.String(v.(string))
	}

	if v, ok := d.GetOk("acl_configuration"); ok && len(v.([]interface{})) > 0 {
		resultConfig.AclConfiguration = expandResultConfigurationACLConfig(v.([]interface{}))
	}

	return resultConfig
}

func expandResultConfigurationEncryptionConfig(config []interface{}) *athena.EncryptionConfiguration {
	if len(config) <= 0 {
		return nil
	}

	data := config[0].(map[string]interface{})

	encryptionConfig := &athena.EncryptionConfiguration{
		EncryptionOption: aws.String(data["encryption_option"].(string)),
	}

	if v, ok := data["kms_key"].(string); ok && v != "" {
		encryptionConfig.KmsKey = aws.String(v)
	}

	return encryptionConfig
}

func expandResultConfigurationACLConfig(config []interface{}) *athena.AclConfiguration {
	if len(config) <= 0 {
		return nil
	}

	data := config[0].(map[string]interface{})

	encryptionConfig := &athena.AclConfiguration{
		S3AclOption: aws.String(data["s3_acl_option"].(string)),
	}

	return encryptionConfig
}

func executeAndExpectNoRows(qeid, action string, conn *athena.Athena) error {
	rs, err := QueryExecutionResult(qeid, conn)
	if err != nil {
		return err
	}
	if len(rs.Rows) != 0 {
		return fmt.Errorf("Athena %s database, unexpected query result: %s", action, flattenResultSet(rs))
	}
	return nil
}

func QueryExecutionResult(qeid string, conn *athena.Athena) (*athena.ResultSet, error) {
	executionStateConf := &resource.StateChangeConf{
		Pending:    []string{athena.QueryExecutionStateQueued, athena.QueryExecutionStateRunning},
		Target:     []string{athena.QueryExecutionStateSucceeded},
		Refresh:    queryExecutionStateRefreshFunc(qeid, conn),
		Timeout:    10 * time.Minute,
		Delay:      3 * time.Second,
		MinTimeout: 3 * time.Second,
	}
	_, err := executionStateConf.WaitForState()
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

func queryExecutionStateRefreshFunc(qeid string, conn *athena.Athena) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &athena.GetQueryExecutionInput{
			QueryExecutionId: aws.String(qeid),
		}
		out, err := conn.GetQueryExecution(input)
		if err != nil {
			return nil, "failed", err
		}

		if out == nil || out.QueryExecution == nil || out.QueryExecution.Status == nil {
			return nil, "", nil
		}

		status := out.QueryExecution.Status

		if aws.StringValue(status.State) == athena.QueryExecutionStateFailed && status.StateChangeReason != nil {
			err = fmt.Errorf("reason: %s", aws.StringValue(status.StateChangeReason))
		}

		return out, aws.StringValue(out.QueryExecution.Status.State), err
	}
}

func flattenResultSet(rs *athena.ResultSet) string {
	ss := make([]string, 0)
	for _, row := range rs.Rows {
		for _, datum := range row.Data {
			ss = append(ss, aws.StringValue(datum.VarCharValue))
		}
	}
	return strings.Join(ss, "\n")
}
