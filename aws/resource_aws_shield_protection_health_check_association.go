package aws

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/shield"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceAwsShieldProtectionHealthCheckAssociation() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceAwsShieldProtectionHealthCheckAssociationCreate,
		ReadContext:   resourceAwsShieldProtectionHealthCheckAssociationRead,
		DeleteContext: resourceAwsShieldProtectionHealthCheckAssociationDelete,

		Schema: map[string]*schema.Schema{
			"shield_protection_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"health_check_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceAwsShieldProtectionHealthCheckAssociationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).shieldconn

	input := &shield.DescribeProtectionInput{
		ProtectionId: aws.String(d.Get("shield_protection_id").(string)),
	}

	resp, err := conn.DescribeProtectionWithContext(ctx, input)
	if isAWSErr(err, shield.ErrCodeResourceNotFoundException, "") {
		log.Printf("[WARN] Shield Protection Health Check Association (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading state of Shield Protection Health Check Association: %s", err))
	}

	isHealthCheck := stringInSlice(d.Get("health_check_id").(string), aws.StringValueSlice(resp.Protection.HealthCheckIds))
	if !isHealthCheck {
		log.Printf("[WARN] Shield Protection Health Check Association (%s) not found, removing from state", d.Id())
		d.SetId("")
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

func resourceAwsShieldProtectionHealthCheckAssociationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).shieldconn

	healthCheckArn := arn.ARN{
		Partition: meta.(*AWSClient).partition,
		Service:   "route53",
		AccountID: meta.(*AWSClient).accountid,
		Resource:  fmt.Sprintf("healthcheck/%s", d.Get("health_check_id").(string)),
	}.String()
	log.Printf("[WARN] ARN: %s", healthCheckArn)

	input := &shield.AssociateHealthCheckInput{
		ProtectionId:   aws.String(d.Get("shield_protection_id").(string)),
		HealthCheckArn: &healthCheckArn,
	}

	_, err := conn.AssociateHealthCheckWithContext(ctx, input)

	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating Shield Protection Health Check Association: %s", err))
	}
	d.SetId(fmt.Sprintf("sphca-%s-%s", d.Get("shield_protection_id").(string), d.Get("health_check_id").(string)))

	return resourceAwsShieldProtectionHealthCheckAssociationRead(ctx, d, meta)
}

func resourceAwsShieldProtectionHealthCheckAssociationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).shieldconn

	input := &shield.DisassociateHealthCheckInput{
		ProtectionId:   aws.String(d.Get("shield_protection_id").(string)),
		HealthCheckArn: aws.String(d.Get("health_check_id").(string)),
	}

	_, err := conn.DisassociateHealthCheckWithContext(ctx, input)

	if err != nil {
		return diag.FromErr(fmt.Errorf("error destroying Shield Protection Health Check Association: %s", err))
	}

	return nil
}
