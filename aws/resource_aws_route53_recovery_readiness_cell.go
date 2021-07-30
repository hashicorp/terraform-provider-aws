package aws

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53recoveryreadiness"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceAwsRoute53RecoveryReadinessCell() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsRoute53RecoveryReadinessCellCreate,
		Read:   resourceAwsRoute53RecoveryReadinessCellRead,
		Update: resourceAwsRoute53RecoveryReadinessCellUpdate,
		Delete: resourceAwsRoute53RecoveryReadinessCellDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"cell_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cell_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"cells": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"parent_readiness_scopes": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func resourceAwsRoute53RecoveryReadinessCellCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).route53recoveryreadinessconn

	input := &route53recoveryreadiness.CreateCellInput{
		CellName: aws.String(d.Get("cell_name").(string)),
		Cells:    expandStringList(d.Get("cells").([]interface{})),
	}

	resp, err := conn.CreateCell(input)
	if err != nil {
		return fmt.Errorf("error creating Route53 Recovery Readiness Cell: %w", err)
	}

	d.SetId(aws.StringValue(resp.CellName))

	return resourceAwsRoute53RecoveryReadinessCellRead(d, meta)
}

func resourceAwsRoute53RecoveryReadinessCellRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).route53recoveryreadinessconn

	input := &route53recoveryreadiness.GetCellInput{
		CellName: aws.String(d.Id()),
	}
	resp, err := conn.GetCell(input)
	if err != nil {
		return fmt.Errorf("error describing Route53 Recovery Readiness Cell: %s", err)
	}
	d.Set("cell_arn", resp.CellArn)
	d.Set("cell_name", resp.CellName)
	d.Set("cells", resp.Cells)
	d.Set("parent_readiness_scopes", resp.ParentReadinessScopes)

	return nil
}

func resourceAwsRoute53RecoveryReadinessCellUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).route53recoveryreadinessconn

	input := &route53recoveryreadiness.UpdateCellInput{
		CellName: aws.String(d.Id()),
		Cells:    expandStringList(d.Get("cells").([]interface{})),
	}

	_, err := conn.UpdateCell(input)
	if err != nil {
		return fmt.Errorf("error updating Route53 Recovery Readiness Cell: %s", err)
	}

	return resourceAwsRoute53RecoveryReadinessCellRead(d, meta)
}

func resourceAwsRoute53RecoveryReadinessCellDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).route53recoveryreadinessconn

	input := &route53recoveryreadiness.DeleteCellInput{
		CellName: aws.String(d.Id()),
	}
	_, err := conn.DeleteCell(input)
	if err != nil {
		if isAWSErr(err, route53recoveryreadiness.ErrCodeResourceNotFoundException, "") {
			return nil
		}
		return fmt.Errorf("error deleting Route53 Recovery Readiness Cell: %s", err)
	}

	gcinput := &route53recoveryreadiness.GetCellInput{
		CellName: aws.String(d.Id()),
	}
	err = resource.Retry(5*time.Minute, func() *resource.RetryError {
		_, err := conn.GetCell(gcinput)
		if err != nil {
			if isAWSErr(err, route53recoveryreadiness.ErrCodeResourceNotFoundException, "") {
				return nil
			}
			return resource.NonRetryableError(err)
		}
		return resource.RetryableError(fmt.Errorf("Route 53 Recovery Readiness Cell (%s) still exists", d.Id()))
	})
	if isResourceTimeoutError(err) {
		_, err = conn.GetCell(gcinput)
	}
	if err != nil {
		return fmt.Errorf("error waiting for Route 53 Recovery Readiness Cell (%s) deletion: %s", d.Id(), err)
	}

	return nil
}
