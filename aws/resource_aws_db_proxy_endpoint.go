package aws

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/rds/finder"
)

func resourceAwsDbProxyEndpoint() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsDbProxyEndpointCreate,
		Read:   resourceAwsDbProxyEndpointRead,
		Delete: resourceAwsDbProxyEndpointDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"db_proxy_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateRdsIdentifier,
			},
			"db_proxy_endpoint_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateRdsIdentifier,
			},
			"tags": tagsSchema(),
			"target_role": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      rds.DBProxyEndpointTargetRoleReadWrite,
				ValidateFunc: validation.StringInSlice(rds.DBProxyEndpointTargetRole_Values(), false),
			},
			"vpc_security_group_ids": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"vpc_subnet_ids": {
				Type:     schema.TypeSet,
				Required: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"endpoint": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"vpc_id": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"is_default": {
				Type:     schema.TypeBool,
				Computed: true,
			},
		},
	}
}

func resourceAwsDbProxyEndpointCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).rdsconn

	dbProxyName := d.Get("db_proxy_name").(string)
	dbProxyEndpointName := d.Get("db_proxy_endpoint_name").(string)

	params := rds.CreateDBProxyEndpointInput{
		DBProxyName:         aws.String(dbProxyName),
		DBProxyEndpointName: aws.String(dbProxyEndpointName),
		TargetRole:          aws.String(d.Get("target_role").(string)),
		VpcSubnetIds:        expandStringSet(d.Get("vpc_subnet_ids").(*schema.Set)),
	}

	if v := d.Get("vpc_security_group_ids").(*schema.Set); v.Len() > 0 {
		params.VpcSecurityGroupIds = expandStringSet(v)
	}

	resp, err := conn.CreateDBProxyEndpoint(&params)

	if err != nil {
		return fmt.Errorf("error Creating RDS DB Proxy Endpoint (%s/%s): %w", dbProxyName, dbProxyEndpointName, err)
	}

	dbProxyEndpoint := resp.DBProxyEndpoint

	d.SetId(strings.Join([]string{dbProxyName, dbProxyEndpointName, aws.StringValue(dbProxyEndpoint.DBProxyEndpointArn)}, "/"))

	return resourceAwsDbProxyTargetRead(d, meta)
}

func resourceAwsDbProxyEndpointParseID(id string) (string, string, string, error) {
	idParts := strings.SplitN(id, "/", 3)
	if len(idParts) != 3 || idParts[0] == "" || idParts[1] == "" || idParts[2] == "" {
		return "", "", "", fmt.Errorf("unexpected format of ID (%s), expected db_proxy_name/db_proxy_endpoint_name/arn", id)
	}
	return idParts[0], idParts[1], idParts[2], nil
}

func resourceAwsDbProxyEndpointRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).rdsconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	dbProxyName, dbProxyEndpointName, dbProxyEndpointArn, err := resourceAwsDbProxyEndpointParseID(d.Id())
	if err != nil {
		return err
	}

	dbProxyEndpoint, err := finder.DBProxyEndpoint(conn, dbProxyName, dbProxyEndpointName, dbProxyEndpointArn)

	if tfawserr.ErrCodeEquals(err, rds.ErrCodeDBProxyNotFoundFault) {
		log.Printf("[WARN] RDS DB Proxy Endpoint (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if tfawserr.ErrCodeEquals(err, rds.ErrCodeDBProxyEndpointNotFoundFault) {
		log.Printf("[WARN] RDS DB Proxy Endpoint (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading RDS DB Proxy Endpoint (%s): %w", d.Id(), err)
	}

	if dbProxyEndpoint == nil {
		log.Printf("[WARN] RDS DB Proxy Endpoint (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	endpointArn := aws.StringValue(dbProxyEndpoint.DBProxyEndpointArn)
	d.Set("arn", endpointArn)
	d.Set("db_proxy_name", dbProxyName)
	d.Set("endpoint", dbProxyEndpoint.Endpoint)
	d.Set("db_proxy_endpoint_name", dbProxyEndpointName)
	d.Set("is_default", dbProxyEndpoint.IsDefault)
	d.Set("target_role", dbProxyEndpoint.TargetRole)
	d.Set("vpc_id", dbProxyEndpoint.VpcId)
	d.Set("target_role", dbProxyEndpoint.TargetRole)
	d.Set("vpc_subnet_ids", flattenStringSet(dbProxyEndpoint.VpcSubnetIds))
	d.Set("vpc_security_group_ids", flattenStringSet(dbProxyEndpoint.VpcSecurityGroupIds))

	tags, err := keyvaluetags.RdsListTags(conn, endpointArn)

	if err != nil {
		return fmt.Errorf("Error listing tags for RDS DB Proxy Endpoint (%s): %w", endpointArn, err)
	}

	if err := d.Set("tags", tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("Error setting tags: %w", err)
	}

	return nil
}

func resourceAwsDbProxyEndpointDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).rdsconn

	params := rds.DeleteDBProxyEndpointInput{
		DBProxyEndpointName: aws.String(d.Get("db_proxy_endpoint_name").(string)),
	}

	log.Printf("[DEBUG] Delete DB Proxy Endpoint: %#v", params)
	_, err := conn.DeleteDBProxyEndpoint(&params)

	if tfawserr.ErrCodeEquals(err, rds.ErrCodeDBProxyNotFoundFault) {
		return nil
	}

	if tfawserr.ErrCodeEquals(err, rds.ErrCodeDBProxyTargetGroupNotFoundFault) {
		return nil
	}

	if tfawserr.ErrCodeEquals(err, rds.ErrCodeDBProxyTargetNotFoundFault) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("Error Deleting DB Proxy Endpoint: %w", err)
	}

	return nil
}
