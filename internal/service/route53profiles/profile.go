package route53profiles

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/route53profiles"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

func ResourceProfile() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceProfileCreate,
		ReadWithoutTimeout:   resourceProfileRead,
		DeleteWithoutTimeout: resourceProfileDelete,
		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:         schema.TypeString,
				ForceNew:     true,
				Required:     true,
				ValidateFunc: validation.NoZeroValues,
			},
		},
	}
}

func resourceProfileCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diag diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53ProfilesClient(ctx)

	name := d.Get("name").(string)
	input := &route53profiles.CreateProfileInput{
		Name: aws.String(name),
		Tags: getTagsIn(ctx),
	}
	output, err := conn.CreateProfile(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diag, "creating Route53 Profile (%s): %s", name, err)
	}
	d.SetId(aws.ToString(output.Profile.Id))
	return append(diag, resourceProfileRead(ctx, d, meta)...)
}

func resourceProfileRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diag diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53ProfilesClient(ctx)

	input := &route53profiles.GetProfileInput{
		ProfileId: aws.String(d.Id()),
	}

	route53Profile, err := conn.GetProfile(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diag, "reading Route53 Profile (%s): %s", d.Id(), err)
	}

	d.Set("arn", route53Profile.Profile.Arn)
	d.Set("client_token", route53Profile.Profile.ClientToken)
	d.Set("id", route53Profile.Profile.Id)

	return diag
}

func resourceProfileDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diag diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53ProfilesClient(ctx)

	log.Printf("[INFO] Deleting Route53 Profile: %s", d.Id())

	var token *string

	for {
		resources, err := conn.ListProfileResourceAssociations(ctx, &route53profiles.ListProfileResourceAssociationsInput{
			ProfileId: aws.String(d.Id()),
			NextToken: token,
		})
		if err != nil {
			return sdkdiag.AppendErrorf(diag, "disassociate Route53 Profile resources (%s): %s", d.Id(), err)
		} else if resources.NextToken == nil {
			break
		} else {
			token = resources.NextToken
		}
	}
	_, err := conn.DeleteProfile(ctx, &route53profiles.DeleteProfileInput{
		ProfileId: aws.String(d.Id()),
	})
	if err != nil {
		return sdkdiag.AppendErrorf(diag, "delete Route53 Profile (%s): %s", d.Id(), err)
	}

	return diag
}
