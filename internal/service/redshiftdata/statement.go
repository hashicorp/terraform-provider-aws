package redshiftdata

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/redshiftdataapiservice"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceStatement() *schema.Resource {
	return &schema.Resource{
		Create: resourceStatementCreate,
		Read:   resourceStatementRead,
		Delete: schema.Noop,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"cluster_identifier": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"database": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"db_user": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"parameters": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				MinItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"value": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
					},
				},
			},
			"secret_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"sql": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"statement_name": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"with_event": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
			},
		},
	}
}

func resourceStatementCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RedshiftDataConn

	input := &redshiftdataapiservice.ExecuteStatementInput{
		ClusterIdentifier: aws.String(d.Get("cluster_identifier").(string)),
		Database:          aws.String(d.Get("database").(string)),
		Sql:               aws.String(d.Get("sql").(string)),
		WithEvent:         aws.Bool(d.Get("with_event").(bool)),
	}

	if v, ok := d.GetOk("db_user"); ok {
		input.DbUser = aws.String(v.(string))
	}

	if v, ok := d.GetOk("parameters"); ok && len(v.([]interface{})) > 0 {
		input.Parameters = expandParameters(v.([]interface{}))
	}

	if v, ok := d.GetOk("secret_arn"); ok {
		input.SecretArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("statement_name"); ok {
		input.StatementName = aws.String(v.(string))
	}

	output, err := conn.ExecuteStatement(input)

	if err != nil {
		return fmt.Errorf("executing Redshift Data Statement: %w", err)
	}

	d.SetId(aws.StringValue(output.Id))

	if _, err := waitStatementFinished(conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return fmt.Errorf("waiting for Redshift Data Statement (%s) to finish: %w", d.Id(), err)
	}

	return resourceStatementRead(d, meta)
}

func resourceStatementRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RedshiftDataConn

	sub, err := FindStatementByID(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Redshift Data Statement (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("reading Redshift Data Statement (%s): %w", d.Id(), err)
	}

	d.Set("cluster_identifier", sub.ClusterIdentifier)
	d.Set("secret_arn", sub.SecretArn)
	d.Set("database", d.Get("database").(string))
	d.Set("db_user", d.Get("db_user").(string))
	d.Set("sql", sub.QueryString)

	if err := d.Set("parameters", flattenParameters(sub.QueryParameters)); err != nil {
		return fmt.Errorf("setting parameters: %w", err)
	}

	return nil
}

func expandParameter(tfMap map[string]interface{}) *redshiftdataapiservice.SqlParameter {
	if tfMap == nil {
		return nil
	}

	apiObject := &redshiftdataapiservice.SqlParameter{}

	if v, ok := tfMap["name"].(string); ok {
		apiObject.Name = aws.String(v)
	}

	if v, ok := tfMap["value"].(string); ok {
		apiObject.Value = aws.String(v)
	}

	return apiObject
}

func expandParameters(tfList []interface{}) []*redshiftdataapiservice.SqlParameter {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*redshiftdataapiservice.SqlParameter

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandParameter(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func flattenParameter(apiObject *redshiftdataapiservice.SqlParameter) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Name; v != nil {
		tfMap["name"] = aws.StringValue(v)
	}

	if v := apiObject.Value; v != nil {
		tfMap["value"] = aws.StringValue(v)
	}
	return tfMap
}

func flattenParameters(apiObjects []*redshiftdataapiservice.SqlParameter) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenParameter(apiObject))
	}

	return tfList
}
