package elasticsearch

import (
	"context"
	"fmt"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	elasticsearch "github.com/aws/aws-sdk-go/service/elasticsearchservice"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"golang.org/x/exp/slices"
)

var (
	vpcEndpointStillInProgress = []string{"CREATING", "UPDATING", "DELETING"}
)

func ResourceDomainVpcEndpoint() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDomainVpcEndpointCreate,
		ReadContext:   resourceDomainVpcEndpointRead,
		UpdateContext: resourceDomainVpcEndpointUpdate,
		DeleteContext: resourceDomainVpcEndpointDelete,

		Schema: map[string]*schema.Schema{
			"domain_arn": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The Amazon Resource Name (ARN) of the domain associated with the endpoint.",
				ForceNew:    true,
			},
			"vpc_options": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"security_group_ids": {
							Type:        schema.TypeSet,
							Optional:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Set:         schema.HashString,
							Description: "The list of security group IDs associated with the VPC endpoints for the domain. If you do not provide a security group ID, OpenSearch Service uses the default security group for the VPC.",
						},
						"subnet_ids": {
							Type:        schema.TypeSet,
							Required:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Set:         schema.HashString,
							Description: "A list of subnet IDs associated with the VPC endpoints for the domain. If your domain uses multiple Availability Zones, you need to provide two subnet IDs, one per zone. Otherwise, provide only one.",
						},
					},
				},
			},
			"owner": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The creator of the endpoint.",
			},
			"connection_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The connection endpoint ID for connecting to the domain.",
			},
		},
	}
}

func resourceDomainVpcEndpointDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ElasticsearchConn

	id := d.Id()
	_, err := conn.DeleteVpcEndpointWithContext(ctx, &elasticsearch.DeleteVpcEndpointInput{
		VpcEndpointId: aws.String(id),
	})
	if tfawserr.ErrCodeEquals(err, elasticsearch.ErrCodeResourceNotFoundException) {
		return nil
	}
	if err != nil {
		return diag.FromErr(err)
	}

	if err := WaitForDomainVPCEndpoint(ctx, conn, id, d.Timeout(schema.TimeoutDelete)); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func resourceDomainVpcEndpointUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ElasticsearchConn

	if !d.HasChanges() {
		return resourceDomainVpcEndpointRead(ctx, d, meta)
	}

	id := d.Id()
	input := &elasticsearch.UpdateVpcEndpointInput{
		VpcEndpointId: aws.String(id),
	}

	if !d.HasChange("vpc_options") {
		return nil
	}
	options := d.Get("vpc_options").([]interface{})
	s := options[0].(map[string]interface{})
	input.VpcOptions = expandVPCOptions(s)

	_, err := conn.UpdateVpcEndpointWithContext(ctx, input)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := WaitForDomainVPCEndpoint(ctx, conn, id, d.Timeout(schema.TimeoutUpdate)); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceDomainVpcEndpointCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ElasticsearchConn

	input := &elasticsearch.CreateVpcEndpointInput{
		ClientToken: aws.String(resource.UniqueId()),
		DomainArn:   aws.String(d.Get("domain_arn").(string)),
	}

	if v, ok := d.GetOk("vpc_options"); ok {
		options := v.([]interface{})
		if options[0] == nil {
			return diag.Errorf("At least one field is expected inside vpc_options")
		}

		s := options[0].(map[string]interface{})
		input.VpcOptions = expandVPCOptions(s)
	}

	output, err := conn.CreateVpcEndpointWithContext(ctx, input)
	if err != nil {
		return diag.FromErr(err)
	}

	id := aws.ToString(output.VpcEndpoint.VpcEndpointId)
	d.SetId(id)

	if err := WaitForDomainVPCEndpoint(ctx, conn, id, d.Timeout(schema.TimeoutCreate)); err != nil {
		return diag.Errorf("error waiting for Elasticsearch Domain VPC Endpoint (%s) create: %s", d.Id(), err)
	}

	return resourceDomainVpcEndpointRead(ctx, d, meta)
}

func resourceDomainVpcEndpointRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ElasticsearchConn

	id := d.Id()

	endpoint, err := FindVPCEndpointByID(ctx, conn, id)
	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Elasticsearch VPC Endpoint (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	if err != nil {
		return diag.Errorf("error reading Elasticsearch VPC Endpoint (%s): %s", d.Id(), err)
	}

	d.Set("domain_arn", endpoint.DomainArn)
	d.Set("owner", endpoint.VpcEndpointOwner)
	d.Set("connection_id", endpoint.Endpoint)

	if err := d.Set("vpc_options", flattenVPCDerivedInfo(endpoint.VpcOptions)); err != nil {
		return diag.Errorf("error setting vpc_options: %s", err)
	}

	return nil
}

func WaitForDomainVPCEndpoint(ctx context.Context, conn *elasticsearch.ElasticsearchService, id string, timeout time.Duration) error {
	err := resource.RetryContext(ctx, timeout, func() *resource.RetryError {
		vpcEndpoint, err := FindVPCEndpointByID(ctx, conn, id)
		if err != nil {
			return resource.NonRetryableError(err)
		}
		if slices.Contains(vpcEndpointStillInProgress, aws.ToString(vpcEndpoint.Status)) {
			return resource.RetryableError(fmt.Errorf("waiting for %s to be finished. Current status: %s", aws.ToString(vpcEndpoint.VpcEndpointId), aws.ToString(vpcEndpoint.Status)))
		}
		return nil
	})

	if tfresource.TimedOut(err) {
		vpcEndpoint, err := FindVPCEndpointByID(ctx, conn, id)
		if err != nil {
			if tfresource.NotFound(err) {
				return nil
			}
			return fmt.Errorf("error describing Elasticsearch domain: %s", err)
		}

		if vpcEndpoint != nil && slices.Contains(vpcEndpointStillInProgress, aws.ToString(vpcEndpoint.Status)) {
			return nil
		}
	}
	return err
}
