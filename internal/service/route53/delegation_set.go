package route53

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

func ResourceDelegationSet() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceDelegationSetCreate,
		ReadWithoutTimeout:   resourceDelegationSetRead,
		DeleteWithoutTimeout: resourceDelegationSetDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
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

func resourceDelegationSetCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	r53 := meta.(*conns.AWSClient).Route53Conn()

	callerRef := resource.UniqueId()
	if v, ok := d.GetOk("reference_name"); ok {
		callerRef = strings.Join([]string{
			v.(string), "-", callerRef,
		}, "")
	}
	input := &route53.CreateReusableDelegationSetInput{
		CallerReference: aws.String(callerRef),
	}

	out, err := r53.CreateReusableDelegationSetWithContext(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Route 53 Reusable Delegation Set: %s", err)
	}

	set := out.DelegationSet
	d.SetId(CleanDelegationSetID(*set.Id))

	return append(diags, resourceDelegationSetRead(ctx, d, meta)...)
}

func resourceDelegationSetRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	r53 := meta.(*conns.AWSClient).Route53Conn()

	input := &route53.GetReusableDelegationSetInput{
		Id: aws.String(CleanDelegationSetID(d.Id())),
	}
	out, err := r53.GetReusableDelegationSetWithContext(ctx, input)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, route53.ErrCodeNoSuchDelegationSet) {
			d.SetId("")
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "reading Route 53 Reusable Delegation Set: %s", err)
	}

	set := out.DelegationSet
	d.Set("name_servers", aws.StringValueSlice(set.NameServers))

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "route53",
		Resource:  fmt.Sprintf("delegationset/%s", d.Id()),
	}.String()
	d.Set("arn", arn)

	return diags
}

func resourceDelegationSetDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	r53 := meta.(*conns.AWSClient).Route53Conn()

	input := &route53.DeleteReusableDelegationSetInput{
		Id: aws.String(CleanDelegationSetID(d.Id())),
	}
	_, err := r53.DeleteReusableDelegationSetWithContext(ctx, input)
	if tfawserr.ErrCodeEquals(err, route53.ErrCodeNoSuchDelegationSet) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Route53 Reusable Delegation Set (%s): %s", d.Id(), err)
	}

	return diags
}

func CleanDelegationSetID(id string) string {
	return strings.TrimPrefix(id, "/delegationset/")
}
