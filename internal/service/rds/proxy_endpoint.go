package rds

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceProxyEndpoint() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceProxyEndpointCreate,
		ReadWithoutTimeout:   resourceProxyEndpointRead,
		DeleteWithoutTimeout: resourceProxyEndpointDelete,
		UpdateWithoutTimeout: resourceProxyEndpointUpdate,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		CustomizeDiff: verify.SetTagsDiff,
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(60 * time.Minute),
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
				ValidateFunc: validIdentifier,
			},
			"db_proxy_endpoint_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validIdentifier,
			},
			"endpoint": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"is_default": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
			"target_role": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      rds.DBProxyEndpointTargetRoleReadWrite,
				ValidateFunc: validation.StringInSlice(rds.DBProxyEndpointTargetRole_Values(), false),
			},
			"vpc_id": {
				Type:     schema.TypeString,
				Computed: true,
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
		},
	}
}

func resourceProxyEndpointCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	dbProxyName := d.Get("db_proxy_name").(string)
	dbProxyEndpointName := d.Get("db_proxy_endpoint_name").(string)

	params := rds.CreateDBProxyEndpointInput{
		DBProxyName:         aws.String(dbProxyName),
		DBProxyEndpointName: aws.String(dbProxyEndpointName),
		TargetRole:          aws.String(d.Get("target_role").(string)),
		VpcSubnetIds:        flex.ExpandStringSet(d.Get("vpc_subnet_ids").(*schema.Set)),
		Tags:                Tags(tags.IgnoreAWS()),
	}

	if v := d.Get("vpc_security_group_ids").(*schema.Set); v.Len() > 0 {
		params.VpcSecurityGroupIds = flex.ExpandStringSet(v)
	}

	_, err := conn.CreateDBProxyEndpointWithContext(ctx, &params)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "Creating RDS DB Proxy Endpoint (%s/%s): %s", dbProxyName, dbProxyEndpointName, err)
	}

	d.SetId(strings.Join([]string{dbProxyName, dbProxyEndpointName}, "/"))

	if _, err := waitDBProxyEndpointAvailable(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for RDS DB Proxy Endpoint (%s) to become available: %s", d.Id(), err)
	}

	return append(diags, resourceProxyEndpointRead(ctx, d, meta)...)
}

func resourceProxyEndpointRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	dbProxyEndpoint, err := FindDBProxyEndpoint(ctx, conn, d.Id())

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, rds.ErrCodeDBProxyNotFoundFault) {
		log.Printf("[WARN] RDS DB Proxy Endpoint (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, rds.ErrCodeDBProxyEndpointNotFoundFault) {
		log.Printf("[WARN] RDS DB Proxy Endpoint (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading RDS DB Proxy Endpoint (%s): %s", d.Id(), err)
	}

	if dbProxyEndpoint == nil {
		if d.IsNewResource() {
			return sdkdiag.AppendErrorf(diags, "reading RDS DB Proxy Endpoint (%s): not found after creation", d.Id())
		}

		log.Printf("[WARN] RDS DB Proxy Endpoint (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	endpointArn := aws.StringValue(dbProxyEndpoint.DBProxyEndpointArn)
	d.Set("arn", endpointArn)
	d.Set("db_proxy_name", dbProxyEndpoint.DBProxyName)
	d.Set("endpoint", dbProxyEndpoint.Endpoint)
	d.Set("db_proxy_endpoint_name", dbProxyEndpoint.DBProxyEndpointName)
	d.Set("is_default", dbProxyEndpoint.IsDefault)
	d.Set("target_role", dbProxyEndpoint.TargetRole)
	d.Set("vpc_id", dbProxyEndpoint.VpcId)
	d.Set("target_role", dbProxyEndpoint.TargetRole)
	d.Set("vpc_subnet_ids", flex.FlattenStringSet(dbProxyEndpoint.VpcSubnetIds))
	d.Set("vpc_security_group_ids", flex.FlattenStringSet(dbProxyEndpoint.VpcSecurityGroupIds))

	tags, err := ListTags(ctx, conn, endpointArn)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing tags for RDS DB Proxy Endpoint (%s): %s", endpointArn, err)
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags_all: %s", err)
	}

	return diags
}

func resourceProxyEndpointUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSConn()

	if d.HasChange("vpc_security_group_ids") {
		params := rds.ModifyDBProxyEndpointInput{
			DBProxyEndpointName: aws.String(d.Get("db_proxy_endpoint_name").(string)),
			VpcSecurityGroupIds: flex.ExpandStringSet(d.Get("vpc_security_group_ids").(*schema.Set)),
		}

		_, err := conn.ModifyDBProxyEndpointWithContext(ctx, &params)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating DB Proxy Endpoint: %s", err)
		}

		if _, err := waitDBProxyEndpointAvailable(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for RDS DB Proxy Endpoint (%s) to become modified: %s", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(ctx, conn, d.Get("arn").(string), o, n); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating RDS DB Proxy Endpoint (%s) tags: %s", d.Get("arn").(string), err)
		}
	}

	return append(diags, resourceProxyEndpointRead(ctx, d, meta)...)
}

func resourceProxyEndpointDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSConn()

	params := rds.DeleteDBProxyEndpointInput{
		DBProxyEndpointName: aws.String(d.Get("db_proxy_endpoint_name").(string)),
	}

	log.Printf("[DEBUG] Delete DB Proxy Endpoint: %#v", params)
	_, err := conn.DeleteDBProxyEndpointWithContext(ctx, &params)

	if err != nil {
		if tfawserr.ErrCodeEquals(err, rds.ErrCodeDBProxyNotFoundFault) || tfawserr.ErrCodeEquals(err, rds.ErrCodeDBProxyEndpointNotFoundFault) {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "Deleting DB Proxy Endpoint: %s", err)
	}

	if _, err := waitDBProxyEndpointDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		if tfawserr.ErrCodeEquals(err, rds.ErrCodeDBProxyNotFoundFault) || tfawserr.ErrCodeEquals(err, rds.ErrCodeDBProxyEndpointNotFoundFault) {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "waiting for RDS DB Proxy Endpoint (%s) to become deleted: %s", d.Id(), err)
	}

	return diags
}
