package opsworks

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/opsworks"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func ResourceRDSDBInstance() *schema.Resource {
	return &schema.Resource{
		Create: resourceRDSDBInstanceCreate,
		Update: resourceRDSDBInstanceUpdate,
		Delete: resourceRDSDBInstanceDelete,
		Read:   resourceRDSDBInstanceRead,

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

func resourceRDSDBInstanceCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*conns.AWSClient).OpsWorksConn

	dbInstanceARN := d.Get("rds_db_instance_arn").(string)
	stackID := d.Get("stack_id").(string)
	id := dbInstanceARN + stackID
	input := &opsworks.RegisterRdsDbInstanceInput{
		DbPassword:       aws.String(d.Get("db_password").(string)),
		DbUser:           aws.String(d.Get("db_user").(string)),
		RdsDbInstanceArn: aws.String(dbInstanceARN),
		StackId:          aws.String(stackID),
	}

	log.Printf("[DEBUG] Registering OpsWorks RDS DB Instance: %s", input)
	_, err := client.RegisterRdsDbInstance(input)

	if err != nil {
		return fmt.Errorf("registering OpsWorks RDS DB Instance (%s): %w", id, err)
	}

	d.SetId(id)

	return resourceRDSDBInstanceRead(d, meta)
}

func resourceRDSDBInstanceRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).OpsWorksConn

	dbInstance, err := FindRDSDBInstanceByTwoPartKey(conn, d.Get("rds_db_instance_arn").(string), d.Get("stack_id").(string))

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] OpsWorks RDS DB Instance %s not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("reading OpsWorks RDS DB Instance (%s): %w", d.Id(), err)
	}

	d.Set("db_user", dbInstance.DbUser)
	d.Set("rds_db_instance_arn", dbInstance.RdsDbInstanceArn)
	d.Set("stack_id", dbInstance.StackId)

	return nil
}

func resourceRDSDBInstanceUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*conns.AWSClient).OpsWorksConn

	input := &opsworks.UpdateRdsDbInstanceInput{
		RdsDbInstanceArn: aws.String(d.Get("rds_db_instance_arn").(string)),
	}

	if d.HasChange("db_password") {
		input.DbPassword = aws.String(d.Get("db_password").(string))
	}

	if d.HasChange("db_user") {
		input.DbUser = aws.String(d.Get("db_user").(string))
	}

	log.Printf("[DEBUG] Updating OpsWorks RDS DB Instance: %s", input)
	_, err := client.UpdateRdsDbInstance(input)

	if err != nil {
		return fmt.Errorf("updating OpsWorks RDS DB Instance (%s): %w", d.Id(), err)
	}

	return resourceRDSDBInstanceRead(d, meta)
}

func resourceRDSDBInstanceDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*conns.AWSClient).OpsWorksConn

	log.Printf("[DEBUG] Deregistering OpsWorks RDS DB Instance: %s", d.Id())
	_, err := client.DeregisterRdsDbInstance(&opsworks.DeregisterRdsDbInstanceInput{
		RdsDbInstanceArn: aws.String(d.Get("rds_db_instance_arn").(string)),
	})

	if tfawserr.ErrCodeEquals(err, opsworks.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("deregistering OpsWorks RDS DB Instance (%s): %w", d.Id(), err)
	}

	return nil
}

func FindRDSDBInstanceByTwoPartKey(conn *opsworks.OpsWorks, dbInstanceARN, stackID string) (*opsworks.RdsDbInstance, error) {
	input := &opsworks.DescribeRdsDbInstancesInput{
		StackId: aws.String(stackID),
	}

	output, err := conn.DescribeRdsDbInstances(input)

	if tfawserr.ErrCodeEquals(err, opsworks.ErrCodeResourceNotFoundException) {
		return nil, &resource.NotFoundError{
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

	return nil, &resource.NotFoundError{}
}
