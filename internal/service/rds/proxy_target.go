// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @SDKResource("aws_db_proxy_target")
func ResourceProxyTarget() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceProxyTargetCreate,
		ReadWithoutTimeout:   resourceProxyTargetRead,
		DeleteWithoutTimeout: resourceProxyTargetDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"db_cluster_identifier": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validIdentifier,
				ExactlyOneOf: []string{
					"db_instance_identifier",
					"db_cluster_identifier",
				},
			},
			"db_instance_identifier": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validIdentifier,
				ExactlyOneOf: []string{
					"db_instance_identifier",
					"db_cluster_identifier",
				},
			},
			"db_proxy_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validIdentifier,
			},
			"endpoint": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"port": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"rds_resource_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"target_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"target_group_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validIdentifier,
			},
			"tracked_cluster_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"type": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceProxyTargetCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSConn(ctx)

	dbProxyName := d.Get("db_proxy_name").(string)
	targetGroupName := d.Get("target_group_name").(string)
	input := &rds.RegisterDBProxyTargetsInput{
		DBProxyName:     aws.String(dbProxyName),
		TargetGroupName: aws.String(targetGroupName),
	}

	if v, ok := d.GetOk("db_cluster_identifier"); ok {
		input.DBClusterIdentifiers = aws.StringSlice([]string{v.(string)})
	}

	if v, ok := d.GetOk("db_instance_identifier"); ok {
		input.DBInstanceIdentifiers = aws.StringSlice([]string{v.(string)})
	}

	outputRaw, err := tfresource.RetryWhenAWSErrMessageContains(ctx, 5*time.Minute,
		func() (interface{}, error) {
			return conn.RegisterDBProxyTargetsWithContext(ctx, input)
		},
		rds.ErrCodeInvalidDBInstanceStateFault, "CREATING")
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "registering RDS DB Proxy (%s/%s) Target: %s", dbProxyName, targetGroupName, err)
	}

	dbProxyTarget := outputRaw.(*rds.RegisterDBProxyTargetsOutput).DBProxyTargets[0]

	d.SetId(strings.Join([]string{dbProxyName, targetGroupName, aws.StringValue(dbProxyTarget.Type), aws.StringValue(dbProxyTarget.RdsResourceId)}, "/"))

	return append(diags, resourceProxyTargetRead(ctx, d, meta)...)
}

func ProxyTargetParseID(id string) (string, string, string, string, error) {
	idParts := strings.SplitN(id, "/", 4)
	if len(idParts) != 4 || idParts[0] == "" || idParts[1] == "" || idParts[2] == "" || idParts[3] == "" {
		return "", "", "", "", fmt.Errorf("unexpected format of ID (%s), expected db_proxy_name/target_group_name/type/id", id)
	}
	return idParts[0], idParts[1], idParts[2], idParts[3], nil
}

func resourceProxyTargetRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSConn(ctx)

	dbProxyName, targetGroupName, targetType, rdsResourceId, err := ProxyTargetParseID(d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading RDS Proxy Target (%s): %s", d.Id(), err)
	}

	dbProxyTarget, err := FindDBProxyTarget(ctx, conn, dbProxyName, targetGroupName, targetType, rdsResourceId)

	if tfawserr.ErrCodeEquals(err, rds.ErrCodeDBProxyNotFoundFault) {
		log.Printf("[WARN] RDS DB Proxy Target (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if tfawserr.ErrCodeEquals(err, rds.ErrCodeDBProxyTargetGroupNotFoundFault) {
		log.Printf("[WARN] RDS DB Proxy Target (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading RDS DB Proxy Target (%s): %s", d.Id(), err)
	}

	if dbProxyTarget == nil {
		log.Printf("[WARN] RDS DB Proxy Target (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	d.Set("db_proxy_name", dbProxyName)
	d.Set("endpoint", dbProxyTarget.Endpoint)
	d.Set("port", dbProxyTarget.Port)
	d.Set("rds_resource_id", dbProxyTarget.RdsResourceId)
	d.Set("target_arn", dbProxyTarget.TargetArn)
	d.Set("target_group_name", targetGroupName)
	d.Set("tracked_cluster_id", dbProxyTarget.TrackedClusterId)
	d.Set("type", dbProxyTarget.Type)

	if aws.StringValue(dbProxyTarget.Type) == rds.TargetTypeRdsInstance {
		d.Set("db_instance_identifier", dbProxyTarget.RdsResourceId)
	} else {
		d.Set("db_cluster_identifier", dbProxyTarget.RdsResourceId)
	}

	return diags
}

func resourceProxyTargetDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSConn(ctx)

	params := rds.DeregisterDBProxyTargetsInput{
		DBProxyName:     aws.String(d.Get("db_proxy_name").(string)),
		TargetGroupName: aws.String(d.Get("target_group_name").(string)),
	}

	if v, ok := d.GetOk("db_instance_identifier"); ok {
		params.DBInstanceIdentifiers = []*string{aws.String(v.(string))}
	}

	if v, ok := d.GetOk("db_cluster_identifier"); ok {
		params.DBClusterIdentifiers = []*string{aws.String(v.(string))}
	}

	log.Printf("[DEBUG] Deregister DB Proxy target: %#v", params)
	_, err := conn.DeregisterDBProxyTargetsWithContext(ctx, &params)

	if tfawserr.ErrCodeEquals(err, rds.ErrCodeDBProxyNotFoundFault) {
		return diags
	}

	if tfawserr.ErrCodeEquals(err, rds.ErrCodeDBProxyTargetGroupNotFoundFault) {
		return diags
	}

	if tfawserr.ErrCodeEquals(err, rds.ErrCodeDBProxyTargetNotFoundFault) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deregistering DB Proxy target: %s", err)
	}

	return diags
}
