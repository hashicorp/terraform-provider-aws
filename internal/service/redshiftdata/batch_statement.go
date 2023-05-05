package redshiftdata

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/redshiftdataapiservice"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// @SDKResource("aws_redshiftdata_batch_statement")
func ResourceBatchStatement() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceBatchStatementCreate,
		ReadWithoutTimeout:   resourceBatchStatementRead,
		DeleteWithoutTimeout: schema.NoopContext,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"cluster_identifier": {
				Type:     schema.TypeString,
				Optional: true,
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
			"secret_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"sqls": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
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
			"workgroup_name": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
		},
	}
}

func resourceBatchStatementCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftDataConn()

	input := &redshiftdataapiservice.BatchExecuteStatementInput{
		Database:  aws.String(d.Get("database").(string)),
		WithEvent: aws.Bool(d.Get("with_event").(bool)),
	}

	if v, ok := d.GetOk("cluster_identifier"); ok {
		input.ClusterIdentifier = aws.String(v.(string))
	}

	if v, ok := d.GetOk("db_user"); ok {
		input.DbUser = aws.String(v.(string))
	}

	if v, ok := d.GetOk("sqls"); ok && len(v.([]interface{})) > 0 {
		input.Sqls = flex.ExpandStringList(v.([]interface{}))
	}

	if v, ok := d.GetOk("secret_arn"); ok {
		input.SecretArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("statement_name"); ok {
		input.StatementName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("workgroup_name"); ok {
		input.WorkgroupName = aws.String(v.(string))
	}

	output, err := conn.BatchExecuteStatementWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "executing Redshift Data Batch Statement: %s", err)
	}

	d.SetId(aws.StringValue(output.Id))

	if _, err := waitStatementFinished(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Redshift Data Statement (%s) to finish: %s", d.Id(), err)
	}

	return append(diags, resourceBatchStatementRead(ctx, d, meta)...)
}

func resourceBatchStatementRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftDataConn()

	sub, err := FindStatementByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Redshift Data Batch Statement (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Redshift Batch Data Statement (%s): %s", d.Id(), err)
	}

	d.Set("cluster_identifier", sub.ClusterIdentifier)
	d.Set("database", d.Get("database").(string))
	d.Set("db_user", d.Get("db_user").(string))
	f := flattenSqlStatements(sub.SubStatements)
	d.Set("sqls", f)
	d.Set("secret_arn", sub.SecretArn)
	d.Set("workgroup_name", sub.WorkgroupName)

	return diags
}

func flattenSqlStatements(apiObjects []*redshiftdataapiservice.SubStatementData) []string {
	if len(apiObjects) == 0 {
		return nil
	}

	// tfMap := map[string]interface{}{}
	var tfStrings []string

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		if v := apiObject.QueryString; v != nil {
			// tfMap["query_string"] = aws.StringValue(v)
			tfStrings = append(tfStrings, aws.StringValue(v))
		}
	}

	return tfStrings
}
