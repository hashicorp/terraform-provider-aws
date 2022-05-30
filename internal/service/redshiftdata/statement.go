package redshiftdata

import (
	"fmt"
	"log"

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

	request := &redshiftdataapiservice.ExecuteStatementInput{
		ClusterIdentifier: aws.String(d.Get("cluster_identifier").(string)),
		Database:          aws.String(d.Get("database").(string)),
		Sql:               aws.String(d.Get("sql").(string)),
		WithEvent:         aws.Bool(d.Get("with_event").(bool)),
	}

	if v, ok := d.GetOk("secret_arn"); ok {
		request.SecretArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("db_user"); ok {
		request.DbUser = aws.String(v.(string))
	}

	if v, ok := d.GetOk("statement_name"); ok {
		request.StatementName = aws.String(v.(string))
	}

	output, err := conn.ExecuteStatement(request)
	if err != nil || output.Id == nil {
		return fmt.Errorf("Error Executing Redshift Data Statement %s: %s", d.Get("cluster_identifier").(string), err)
	}

	d.SetId(aws.StringValue(output.Id))

	if _, err := waitStatementFinished(conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for Redshift Data Statement (%s) to be finished: %w", d.Id(), err)
	}

	return resourceStatementRead(d, meta)
}

func resourceStatementRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RedshiftDataConn

	sub, err := FindStatementById(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Redshift Data Statement (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading Redshift Data Statement %s: %w", d.Id(), err)
	}

	d.Set("cluster_identifier", sub.ClusterIdentifier)
	d.Set("secret_arn", sub.SecretArn)
	d.Set("database", sub.Database)
	d.Set("db_user", sub.DbUser)
	d.Set("sql", sub.QueryString)

	return nil
}
