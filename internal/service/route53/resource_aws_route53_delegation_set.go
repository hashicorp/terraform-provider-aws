package route53

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
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceDelegationSet() *schema.Resource {
	return &schema.Resource{
		Create: resourceDelegationSetCreate,
		Read:   resourceDelegationSetRead,
		Delete: resourceDelegationSetDelete,
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

func resourceDelegationSetCreate(d *schema.ResourceData, meta interface{}) error {
	r53 := meta.(*conns.AWSClient).Route53Conn

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

	return resourceDelegationSetRead(d, meta)
}

func resourceDelegationSetRead(d *schema.ResourceData, meta interface{}) error {
	r53 := meta.(*conns.AWSClient).Route53Conn

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
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "route53",
		Resource:  fmt.Sprintf("delegationset/%s", d.Id()),
	}.String()
	d.Set("arn", arn)

	return nil
}

func resourceDelegationSetDelete(d *schema.ResourceData, meta interface{}) error {
	r53 := meta.(*conns.AWSClient).Route53Conn

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
