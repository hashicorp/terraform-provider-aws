package route53recoveryreadiness

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53recoveryreadiness"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceRecoveryGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceRecoveryGroupCreate,
		Read:   resourceRecoveryGroupRead,
		Update: resourceRecoveryGroupUpdate,
		Delete: resourceRecoveryGroupDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cells": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"recovery_group_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceRecoveryGroupCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).Route53RecoveryReadinessConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	input := &route53recoveryreadiness.CreateRecoveryGroupInput{
		RecoveryGroupName: aws.String(d.Get("recovery_group_name").(string)),
		Cells:             flex.ExpandStringList(d.Get("cells").([]interface{})),
	}

	resp, err := conn.CreateRecoveryGroup(input)
	if err != nil {
		return fmt.Errorf("error creating Route53 Recovery Readiness RecoveryGroup: %w", err)
	}

	d.SetId(aws.StringValue(resp.RecoveryGroupName))

	if len(tags) > 0 {
		arn := aws.StringValue(resp.RecoveryGroupArn)
		if err := tftags.Route53recoveryreadinessUpdateTags(conn, arn, nil, tags); err != nil {
			return fmt.Errorf("error adding Route53 Recovery Readiness RecoveryGroup (%s) tags: %w", d.Id(), err)
		}
	}

	return resourceRecoveryGroupRead(d, meta)
}

func resourceRecoveryGroupRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).Route53RecoveryReadinessConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	input := &route53recoveryreadiness.GetRecoveryGroupInput{
		RecoveryGroupName: aws.String(d.Id()),
	}
	resp, err := conn.GetRecoveryGroup(input)

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, route53recoveryreadiness.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] Route53RecoveryReadiness Recovery Group (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error describing Route53 Recovery Readiness RecoveryGroup: %s", err)
	}

	d.Set("arn", resp.RecoveryGroupArn)
	d.Set("recovery_group_name", resp.RecoveryGroupName)
	d.Set("cells", resp.Cells)

	tags, err := tftags.Route53recoveryreadinessListTags(conn, d.Get("arn").(string))

	if err != nil {
		return fmt.Errorf("error listing tags for Route53 Recovery Readiness RecoveryGroup (%s): %w", d.Id(), err)
	}

	tags = tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceRecoveryGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).Route53RecoveryReadinessConn

	input := &route53recoveryreadiness.UpdateRecoveryGroupInput{
		RecoveryGroupName: aws.String(d.Id()),
		Cells:             flex.ExpandStringList(d.Get("cells").([]interface{})),
	}

	_, err := conn.UpdateRecoveryGroup(input)

	if err != nil {
		return fmt.Errorf("error updating Route53 Recovery Readiness RecoveryGroup: %s", err)
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")
		arn := d.Get("arn").(string)
		if err := tftags.Route53recoveryreadinessUpdateTags(conn, arn, o, n); err != nil {
			return fmt.Errorf("error updating Route53 Recovery Readiness RecoveryGroup (%s) tags: %w", d.Id(), err)
		}
	}

	return resourceRecoveryGroupRead(d, meta)
}

func resourceRecoveryGroupDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).Route53RecoveryReadinessConn

	input := &route53recoveryreadiness.DeleteRecoveryGroupInput{
		RecoveryGroupName: aws.String(d.Id()),
	}

	_, err := conn.DeleteRecoveryGroup(input)

	if err != nil {
		if tfawserr.ErrMessageContains(err, route53recoveryreadiness.ErrCodeResourceNotFoundException, "") {
			return nil
		}
		return fmt.Errorf("error deleting Route53 Recovery Readiness RecoveryGroup: %s", err)
	}

	gcinput := &route53recoveryreadiness.GetRecoveryGroupInput{
		RecoveryGroupName: aws.String(d.Id()),
	}

	err = resource.Retry(d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		_, err := conn.GetRecoveryGroup(gcinput)
		if err != nil {
			if tfawserr.ErrMessageContains(err, route53recoveryreadiness.ErrCodeResourceNotFoundException, "") {
				return nil
			}
			return resource.NonRetryableError(err)
		}
		return resource.RetryableError(fmt.Errorf("Route 53 Recovery Readiness RecoveryGroup (%s) still exists", d.Id()))
	})

	if tfresource.TimedOut(err) {
		_, err = conn.GetRecoveryGroup(gcinput)
	}

	if err != nil {
		return fmt.Errorf("error waiting for Route 53 Recovery Readiness RecoveryGroup (%s) deletion: %s", d.Id(), err)
	}

	return nil
}
