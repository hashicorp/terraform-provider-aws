package aws

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceAwsRoute53DelegationSet() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsRoute53DelegationSetCreate,
		Read:   resourceAwsRoute53DelegationSetRead,
		Delete: resourceAwsRoute53DelegationSetDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"reference_name": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(0, 128),
			},

			"name_servers": {
				Type:     schema.TypeList,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},
		},
	}
}

func resourceAwsRoute53DelegationSetCreate(d *schema.ResourceData, meta interface{}) error {
	r53 := meta.(*AWSClient).r53conn

	callerRef := resource.UniqueId()
	if v, ok := d.GetOk("reference_name"); ok {
		callerRef = strings.Join([]string{
			v.(string), "-", callerRef,
		}, "")
	}
	input := &route53.CreateReusableDelegationSetInput{
		CallerReference: aws.String(callerRef),
	}

	log.Printf("[DEBUG] Creating Route53 reusable delegation set: %#v", input)
	out, err := r53.CreateReusableDelegationSet(input)
	if err != nil {
		return err
	}
	log.Printf("[DEBUG] Route53 reusable delegation set created: %#v", out)

	set := out.DelegationSet
	d.SetId(cleanDelegationSetId(*set.Id))

	return resourceAwsRoute53DelegationSetRead(d, meta)
}

func resourceAwsRoute53DelegationSetRead(d *schema.ResourceData, meta interface{}) error {
	r53 := meta.(*AWSClient).r53conn

	input := &route53.GetReusableDelegationSetInput{
		Id: aws.String(cleanDelegationSetId(d.Id())),
	}
	log.Printf("[DEBUG] Reading Route53 reusable delegation set: %#v", input)
	out, err := r53.GetReusableDelegationSet(input)
	if err != nil {
		if tfawserr.ErrMessageContains(err, route53.ErrCodeNoSuchDelegationSet, "") {
			d.SetId("")
			return nil

		}
		return err
	}
	log.Printf("[DEBUG] Route53 reusable delegation set received: %#v", out)

	set := out.DelegationSet
	d.Set("name_servers", aws.StringValueSlice(set.NameServers))

	arn := arn.ARN{
		Partition: meta.(*AWSClient).partition,
		Service:   "route53",
		Resource:  fmt.Sprintf("delegationset/%s", d.Id()),
	}.String()
	d.Set("arn", arn)

	return nil
}

func resourceAwsRoute53DelegationSetDelete(d *schema.ResourceData, meta interface{}) error {
	r53 := meta.(*AWSClient).r53conn

	input := &route53.DeleteReusableDelegationSetInput{
		Id: aws.String(cleanDelegationSetId(d.Id())),
	}
	log.Printf("[DEBUG] Deleting Route53 reusable delegation set: %#v", input)
	_, err := r53.DeleteReusableDelegationSet(input)
	if tfawserr.ErrMessageContains(err, route53.ErrCodeNoSuchDelegationSet, "") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting Route53 reusable delegation set (%s): %w", d.Id(), err)
	}

	return nil
}

func cleanDelegationSetId(id string) string {
	return strings.TrimPrefix(id, "/delegationset/")
}
