// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package servicediscovery

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicediscovery"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_service_discovery_public_dns_namespace", name="Public DNS Namespace")
// @Tags(identifierAttribute="arn")
func ResourcePublicDNSNamespace() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourcePublicDNSNamespaceCreate,
		ReadWithoutTimeout:   resourcePublicDNSNamespaceRead,
		UpdateWithoutTimeout: resourcePublicDNSNamespaceUpdate,
		DeleteWithoutTimeout: resourcePublicDNSNamespaceDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"hosted_zone": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validNamespaceName,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourcePublicDNSNamespaceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ServiceDiscoveryConn(ctx)

	name := d.Get("name").(string)
	input := &servicediscovery.CreatePublicDnsNamespaceInput{
		CreatorRequestId: aws.String(id.UniqueId()),
		Name:             aws.String(name),
		Tags:             getTagsIn(ctx),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	output, err := conn.CreatePublicDnsNamespaceWithContext(ctx, input)

	if err != nil {
		return diag.Errorf("creating Service Discovery Public DNS Namespace (%s): %s", name, err)
	}

	operation, err := WaitOperationSuccess(ctx, conn, aws.StringValue(output.OperationId))

	if err != nil {
		return diag.Errorf("waiting for Service Discovery Public DNS Namespace (%s) create: %s", name, err)
	}

	namespaceID, ok := operation.Targets[servicediscovery.OperationTargetTypeNamespace]

	if !ok {
		return diag.Errorf("creating Service Discovery Public DNS Namespace (%s): operation response missing Namespace ID", name)
	}

	d.SetId(aws.StringValue(namespaceID))

	return resourcePublicDNSNamespaceRead(ctx, d, meta)
}

func resourcePublicDNSNamespaceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ServiceDiscoveryConn(ctx)

	ns, err := FindNamespaceByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Service Discovery Public DNS Namespace %s not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading Service Discovery Public DNS Namespace (%s): %s", d.Id(), err)
	}

	arn := aws.StringValue(ns.Arn)
	d.Set("arn", arn)
	d.Set("description", ns.Description)
	if ns.Properties != nil && ns.Properties.DnsProperties != nil {
		d.Set("hosted_zone", ns.Properties.DnsProperties.HostedZoneId)
	} else {
		d.Set("hosted_zone", nil)
	}
	d.Set("name", ns.Name)

	return nil
}

func resourcePublicDNSNamespaceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ServiceDiscoveryConn(ctx)

	if d.HasChange("description") {
		input := &servicediscovery.UpdatePublicDnsNamespaceInput{
			Id: aws.String(d.Id()),
			Namespace: &servicediscovery.PublicDnsNamespaceChange{
				Description: aws.String(d.Get("description").(string)),
			},
			UpdaterRequestId: aws.String(id.UniqueId()),
		}

		output, err := conn.UpdatePublicDnsNamespaceWithContext(ctx, input)

		if err != nil {
			return diag.Errorf("updating Service Discovery Public DNS Namespace (%s): %s", d.Id(), err)
		}

		if output != nil && output.OperationId != nil {
			if _, err := WaitOperationSuccess(ctx, conn, aws.StringValue(output.OperationId)); err != nil {
				return diag.Errorf("waiting for Service Discovery Public DNS Namespace (%s) update: %s", d.Id(), err)
			}
		}
	}

	return resourcePublicDNSNamespaceRead(ctx, d, meta)
}

func resourcePublicDNSNamespaceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ServiceDiscoveryConn(ctx)

	log.Printf("[INFO] Deleting Service Discovery Public DNS Namespace: %s", d.Id())
	output, err := conn.DeleteNamespaceWithContext(ctx, &servicediscovery.DeleteNamespaceInput{
		Id: aws.String(d.Id()),
	})

	if err != nil {
		return diag.Errorf("deleting Service Discovery Public DNS Namespace (%s): %s", d.Id(), err)
	}

	if output != nil && output.OperationId != nil {
		if _, err := WaitOperationSuccess(ctx, conn, aws.StringValue(output.OperationId)); err != nil {
			return diag.Errorf("waiting for Service Discovery Public DNS Namespace (%s) delete: %s", d.Id(), err)
		}
	}

	return nil
}
