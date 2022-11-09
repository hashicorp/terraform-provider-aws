package appsync

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appsync"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceDomainName() *schema.Resource {
	return &schema.Resource{
		Create: resourceDomainNameCreate,
		Read:   resourceDomainNameRead,
		Update: resourceDomainNameUpdate,
		Delete: resourceDomainNameDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"appsync_domain_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"certificate_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"domain_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"hosted_zone_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceDomainNameCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).AppSyncConn

	params := &appsync.CreateDomainNameInput{
		CertificateArn: aws.String(d.Get("certificate_arn").(string)),
		Description:    aws.String(d.Get("description").(string)),
		DomainName:     aws.String(d.Get("domain_name").(string)),
	}

	resp, err := conn.CreateDomainName(params)
	if err != nil {
		return fmt.Errorf("error creating Appsync Domain Name: %w", err)
	}

	d.SetId(aws.StringValue(resp.DomainNameConfig.DomainName))

	return resourceDomainNameRead(d, meta)
}

func resourceDomainNameRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).AppSyncConn

	domainName, err := FindDomainNameByID(conn, d.Id())
	if domainName == nil && !d.IsNewResource() {
		log.Printf("[WARN] AppSync Domain Name (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error getting Appsync Domain Name %q: %w", d.Id(), err)
	}

	d.Set("domain_name", domainName.DomainName)
	d.Set("description", domainName.Description)
	d.Set("certificate_arn", domainName.CertificateArn)
	d.Set("hosted_zone_id", domainName.HostedZoneId)
	d.Set("appsync_domain_name", domainName.AppsyncDomainName)

	return nil
}

func resourceDomainNameUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).AppSyncConn

	params := &appsync.UpdateDomainNameInput{
		DomainName: aws.String(d.Id()),
	}

	if d.HasChange("description") {
		params.Description = aws.String(d.Get("description").(string))
	}

	_, err := conn.UpdateDomainName(params)
	if err != nil {
		return fmt.Errorf("error updating Appsync Domain Name %q: %w", d.Id(), err)
	}

	return resourceDomainNameRead(d, meta)
}

func resourceDomainNameDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).AppSyncConn

	input := &appsync.DeleteDomainNameInput{
		DomainName: aws.String(d.Id()),
	}

	err := resource.Retry(5*time.Minute, func() *resource.RetryError {
		_, err := conn.DeleteDomainName(input)
		if tfawserr.ErrCodeEquals(err, appsync.ErrCodeConcurrentModificationException) {
			return resource.RetryableError(fmt.Errorf("deleting Appsync Domain Name %q: %w", d.Id(), err))
		}
		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})
	if tfresource.TimedOut(err) {
		_, err = conn.DeleteDomainName(input)
	}
	if err != nil {
		return fmt.Errorf("error deleting Appsync Domain Name %q: %w", d.Id(), err)
	}

	return nil
}
