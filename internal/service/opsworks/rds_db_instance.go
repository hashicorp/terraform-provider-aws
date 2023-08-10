// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package opsworks

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/opsworks"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @SDKResource("aws_opsworks_rds_db_instance")
func ResourceRDSDBInstance() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceRDSDBInstanceCreate,
		UpdateWithoutTimeout: resourceRDSDBInstanceUpdate,
		DeleteWithoutTimeout: resourceRDSDBInstanceDelete,
		ReadWithoutTimeout:   resourceRDSDBInstanceRead,

		Schema: map[string]*schema.Schema{
			"db_password": {
				Type:      schema.TypeString,
				Required:  true,
				Sensitive: true,
			},
			"db_user": {
				Type:     schema.TypeString,
				Required: true,
			},
			"rds_db_instance_arn": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"stack_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceRDSDBInstanceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := meta.(*conns.AWSClient).OpsWorksConn(ctx)

	dbInstanceARN := d.Get("rds_db_instance_arn").(string)
	stackID := d.Get("stack_id").(string)
	id := dbInstanceARN + stackID
	input := &opsworks.RegisterRdsDbInstanceInput{
		DbPassword:       aws.String(d.Get("db_password").(string)),
		DbUser:           aws.String(d.Get("db_user").(string)),
		RdsDbInstanceArn: aws.String(dbInstanceARN),
		StackId:          aws.String(stackID),
	}

	_, err := client.RegisterRdsDbInstanceWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "registering OpsWorks RDS DB Instance (%s): %s", id, err)
	}

	d.SetId(id)

	return append(diags, resourceRDSDBInstanceRead(ctx, d, meta)...)
}

func resourceRDSDBInstanceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OpsWorksConn(ctx)

	dbInstance, err := FindRDSDBInstanceByTwoPartKey(ctx, conn, d.Get("rds_db_instance_arn").(string), d.Get("stack_id").(string))

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] OpsWorks RDS DB Instance %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading OpsWorks RDS DB Instance (%s): %s", d.Id(), err)
	}

	d.Set("db_user", dbInstance.DbUser)
	d.Set("rds_db_instance_arn", dbInstance.RdsDbInstanceArn)
	d.Set("stack_id", dbInstance.StackId)

	return diags
}

func resourceRDSDBInstanceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := meta.(*conns.AWSClient).OpsWorksConn(ctx)

	input := &opsworks.UpdateRdsDbInstanceInput{
		RdsDbInstanceArn: aws.String(d.Get("rds_db_instance_arn").(string)),
	}

	if d.HasChange("db_password") {
		input.DbPassword = aws.String(d.Get("db_password").(string))
	}

	if d.HasChange("db_user") {
		input.DbUser = aws.String(d.Get("db_user").(string))
	}

	_, err := client.UpdateRdsDbInstanceWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating OpsWorks RDS DB Instance (%s): %s", d.Id(), err)
	}

	return append(diags, resourceRDSDBInstanceRead(ctx, d, meta)...)
}

func resourceRDSDBInstanceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := meta.(*conns.AWSClient).OpsWorksConn(ctx)

	log.Printf("[DEBUG] Deregistering OpsWorks RDS DB Instance: %s", d.Id())
	_, err := client.DeregisterRdsDbInstanceWithContext(ctx, &opsworks.DeregisterRdsDbInstanceInput{
		RdsDbInstanceArn: aws.String(d.Get("rds_db_instance_arn").(string)),
	})

	if tfawserr.ErrCodeEquals(err, opsworks.ErrCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deregistering OpsWorks RDS DB Instance (%s): %s", d.Id(), err)
	}

	return diags
}

func FindRDSDBInstanceByTwoPartKey(ctx context.Context, conn *opsworks.OpsWorks, dbInstanceARN, stackID string) (*opsworks.RdsDbInstance, error) {
	input := &opsworks.DescribeRdsDbInstancesInput{
		StackId: aws.String(stackID),
	}

	output, err := conn.DescribeRdsDbInstancesWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, opsworks.ErrCodeResourceNotFoundException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.RdsDbInstances) == 0 {
		return nil, tfresource.NewEmptyResultError(input)
	}

	for _, v := range output.RdsDbInstances {
		if aws.StringValue(v.RdsDbInstanceArn) == dbInstanceARN {
			return v, nil
		}
	}

	return nil, &retry.NotFoundError{}
}
