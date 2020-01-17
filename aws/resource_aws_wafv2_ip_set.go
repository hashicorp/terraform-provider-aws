package aws

import (
	"fmt"
	"github.com/aws/aws-sdk-go/service/wafv2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceAwsWafv2IPSet() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsWafv2IPSetCreate,
		Read:   resourceAwsWafv2IPSetRead,
		Update: resourceAwsWafv2IPSetUpdate,
		Delete: resourceAwsWafv2IPSetDelete,

		Schema: map[string]*schema.Schema{
			"addresses": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"ip_address_version": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					wafv2.IPAddressVersionIpv4,
					wafv2.IPAddressVersionIpv6,
				}, false),
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"scope": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					wafv2.ScopeCloudfront,
					wafv2.ScopeRegional,
				}, false),
			},
			"tags": tagsSchema(),
		},
	}
}

func resourceAwsWafv2IPSetCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).wafv2conn
	var resp *wafv2.CreateIPSetOutput

	params := &wafv2.CreateIPSetInput{
		Addresses:        expandStringList(d.Get("addresses").([]interface{})),
		Description:      aws.String(d.Get("description").(string)),
		IPAddressVersion: aws.String(d.Get("ip_address_version").(string)),
		Name:             aws.String(d.Get("name").(string)),
		Scope:            aws.String(d.Get("scope").(string)),
	}

	err := resource.Retry(15*time.Minute, func() *resource.RetryError {
		var err error
		resp, err = conn.CreateIPSet(params)
		if err != nil {
			if isAWSErr(err, wafv2.ErrCodeWAFInternalErrorException, "AWS WAF couldn’t perform the operation because of a system problem") {
				return resource.RetryableError(err)
			}
			if isAWSErr(err, wafv2.ErrCodeWAFTagOperationException, "An error occurred during the tagging operation") {
				return resource.RetryableError(err)
			}
			if isAWSErr(err, wafv2.ErrCodeWAFTagOperationInternalErrorException, "AWS WAF couldn’t perform your tagging operation because of an internal error") {
				return resource.RetryableError(err)
			}
			if isAWSErr(err, wafv2.ErrCodeWAFOptimisticLockException, "AWS WAF couldn’t save your changes because you tried to update or delete a resource that has changed since you last retrieved it") {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})
	if isResourceTimeoutError(err) {
		_, err = conn.CreateIPSet(params)
	}
	if err != nil {
		return err
	}
	d.SetId(*resp.Summary.Id)

	return resourceAwsWafv2IPSetRead(d, meta)
}

func resourceAwsWafv2IPSetRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).wafv2conn

	params := &wafv2.GetIPSetInput{
		Id:    aws.String(d.Id()),
		Name:  aws.String(d.Get("name").(string)),
		Scope: aws.String(d.Get("scope").(string)),
	}

	resp, err := conn.GetIPSet(params)
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok && awsErr.Code() == wafv2.ErrCodeWAFNonexistentItemException {
			log.Printf("[WARN] WAFV2 IPSet (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}

		return err
	}

	d.Set("name", resp.IPSet.Name)
	d.Set("description", resp.IPSet.Description)
	d.Set("ip_address_version", resp.IPSet.IPAddressVersion)
	d.Set("arn", resp.IPSet.ARN)
	d.Set("addresses", flattenStringList(resp.IPSet.Addresses))

	return nil
}

func resourceAwsWafv2IPSetUpdate(d *schema.ResourceData, meta interface{}) error {
	return resourceAwsWafv2IPSetRead(d, meta)
}

func resourceAwsWafv2IPSetDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).wafv2conn
	var resp *wafv2.GetIPSetOutput
	params := &wafv2.GetIPSetInput{
		Id:    aws.String(d.Id()),
		Name:  aws.String(d.Get("name").(string)),
		Scope: aws.String(d.Get("scope").(string)),
	}
	log.Printf("[INFO] Deleting WAFV2 IPSet %s", d.Id())

	err := resource.Retry(15*time.Minute, func() *resource.RetryError {
		var err error
		resp, err = conn.GetIPSet(params)
		if err != nil {
			return resource.NonRetryableError(fmt.Errorf("Error getting lock token: %s", err))
		}

		_, err = conn.DeleteIPSet(&wafv2.DeleteIPSetInput{
			Id:        aws.String(d.Id()),
			Name:      aws.String(d.Get("name").(string)),
			Scope:     aws.String(d.Get("scope").(string)),
			LockToken: resp.LockToken,
		})

		if err != nil {
			if isAWSErr(err, wafv2.ErrCodeWAFInternalErrorException, "AWS WAF couldn’t perform the operation because of a system problem") {
				return resource.RetryableError(err)
			}
			if isAWSErr(err, wafv2.ErrCodeWAFOptimisticLockException, "AWS WAF couldn’t save your changes because you tried to update or delete a resource that has changed since you last retrieved it") {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})

	if isResourceTimeoutError(err) {
		_, err = conn.DeleteIPSet(&wafv2.DeleteIPSetInput{
			Id:        aws.String(d.Id()),
			Name:      aws.String(d.Get("name").(string)),
			Scope:     aws.String(d.Get("scope").(string)),
			LockToken: resp.LockToken,
		})
	}

	if err != nil {
		return fmt.Errorf("Error deleting WAFV2 IPSet: %s", err)
	}

	return nil
}
