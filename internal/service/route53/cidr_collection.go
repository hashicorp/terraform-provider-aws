package route53

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apprunner"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func ResourceCidrCollection() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: ResourceCidrCollectionCreate,
		ReadWithoutTimeout:   ResourceCidrCollectionRead,
		UpdateWithoutTimeout: ResourceCidrCollectionUpdate,
		DeleteWithoutTimeout: ResourceCidrCollectionDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"caller_reference": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"version": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func ResourceCidrCollectionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).Route53Conn

	name := d.Get("name").(string)

	input := &route53.CreateCidrCollectionInput{
		Name:            aws.String(name),
		CallerReference: aws.String(resource.UniqueId()),
	}

	output, err := conn.CreateCidrCollectionWithContext(ctx, input)

	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating Route 53 CIDR collection (%s): %w", name, err))
	}

	if output == nil || output.Collection == nil {
		return diag.FromErr(fmt.Errorf("error creating Route 53 CIDR collection (%s): empty output", name))
	}

	d.SetId(aws.StringValue(output.Collection.Arn))

	return ResourceCidrCollectionRead(ctx, d, meta)
}

func ResourceCidrCollectionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).Route53Conn

	input := &route53.ListCidrCollectionsInput{}

	for {
		res, err := conn.ListCidrCollectionsWithContext(ctx, input)

		if err != nil {
			return diag.Errorf("Error reading Route 53 CIDR collections: %w", err)
		}

		for _, cidrCollection := range res.CidrCollections {
			if cidrCollection.Arn == aws.String(d.Id()) {
				d.Set("arn", cidrCollection.Arn)
				d.Set("name", cidrCollection.Name)
				d.Set("id", cidrCollection.Id)
				d.Set("version", cidrCollection.Version)
				return nil
			}
		}

		// Loop till we find our cidr collection or we reach the end
		if res.NextToken != nil {
			input.NextToken = res.NextToken
		} else {
			break
		}
	}

	// no cidr collection found
	log.Printf("[WARN] Route53 CIDR collection (%s) not found, removing from state", d.Id())
	d.SetId("")
	return nil
}

func ResourceCidrCollectionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).Route53Conn
}

func ResourceCidrCollectionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).Route53Conn

	input := &route53.DeleteCidrCollectionInput{
		Id: aws.String(d.Id()),
	}

	_, err := conn.DeleteCidrCollectionWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, apprunner.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("error deleting Route 53 CIDR collection (%s): %w", d.Id(), err))
	}

	return nil
}
