package shield

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/shield"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func ResourceProtectionHealthCheckAssociation() *schema.Resource {
	return &schema.Resource{
		Create: ResourceProtectionHealthCheckAssociationCreate,
		Read:   ResourceProtectionHealthCheckAssociationRead,
		Delete: ResourceProtectionHealthCheckAssociationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"shield_protection_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"health_check_arn": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func ResourceProtectionHealthCheckAssociationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ShieldConn

	protectionId := d.Get("shield_protection_id").(string)
	healthCheckArn := d.Get("health_check_arn").(string)
	id := ProtectionHealthCheckAssociationCreateResourceID(protectionId, healthCheckArn)

	input := &shield.AssociateHealthCheckInput{
		ProtectionId:   aws.String(protectionId),
		HealthCheckArn: aws.String(healthCheckArn),
	}

	_, err := conn.AssociateHealthCheck(input)
	if err != nil {
		return fmt.Errorf("error associating Route53 Health Check (%s) with Shield Protected resource (%s): %s", d.Get("health_check_arn"), d.Get("shield_protection_id"), err)
	}
	d.SetId(id)
	return ResourceProtectionHealthCheckAssociationRead(d, meta)
}

func ResourceProtectionHealthCheckAssociationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ShieldConn

	protectionId, healthCheckArn, err := ProtectionHealthCheckAssociationParseResourceID(d.Id())

	if err != nil {
		return fmt.Errorf("error parsing Shield Protection and Route53 Health Check Association ID: %w", err)
	}

	input := &shield.DescribeProtectionInput{
		ProtectionId: aws.String(protectionId),
	}

	resp, err := conn.DescribeProtection(input)

	if tfawserr.ErrCodeEquals(err, shield.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] Shield Protection itself (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading Shield Protection Health Check Association (%s): %s", d.Id(), err)
	}

	isHealthCheck := stringInSlice(strings.Split(healthCheckArn, "/")[1], aws.StringValueSlice(resp.Protection.HealthCheckIds))
	if !isHealthCheck {
		log.Printf("[WARN] Shield Protection Health Check Association (%s) not found, removing from state", d.Id())
		d.SetId("")
	}

	d.Set("health_check_arn", healthCheckArn)
	d.Set("shield_protection_id", resp.Protection.Id)

	return nil
}

func ResourceProtectionHealthCheckAssociationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ShieldConn

	protectionId, healthCheckId, err := ProtectionHealthCheckAssociationParseResourceID(d.Id())

	if err != nil {
		return fmt.Errorf("error parsing Shield Protection and Route53 Health Check Association ID: %w", err)
	}

	input := &shield.DisassociateHealthCheckInput{
		ProtectionId:   aws.String(protectionId),
		HealthCheckArn: aws.String(healthCheckId),
	}

	_, err = conn.DisassociateHealthCheck(input)

	if err != nil {
		return fmt.Errorf("error disassociating Route53 Health Check (%s) from Shield Protected resource (%s): %s", d.Get("health_check_arn"), d.Get("shield_protection_id"), err)
	}
	return nil
}

func stringInSlice(expected string, list []string) bool {
	for _, item := range list {
		if item == expected {
			return true
		}
	}
	return false
}
