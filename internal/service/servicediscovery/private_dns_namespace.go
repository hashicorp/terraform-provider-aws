// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package servicediscovery

import (
	"context"
	"fmt"
	"log"
	"strings"

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

// @SDKResource("aws_service_discovery_private_dns_namespace", name="Private DNS Namespace")
// @Tags(identifierAttribute="arn")
func ResourcePrivateDNSNamespace() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourcePrivateDNSNamespaceCreate,
		ReadWithoutTimeout:   resourcePrivateDNSNamespaceRead,
		UpdateWithoutTimeout: resourcePrivateDNSNamespaceUpdate,
		DeleteWithoutTimeout: resourcePrivateDNSNamespaceDelete,

		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				idParts := strings.Split(d.Id(), ":")
				if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
					return nil, fmt.Errorf("Unexpected format of ID (%q), expected NAMESPACE_ID:VPC_ID", d.Id())
				}
				d.SetId(idParts[0])
				d.Set("vpc", idParts[1])
				return []*schema.ResourceData{d}, nil
			},
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
			"vpc": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourcePrivateDNSNamespaceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ServiceDiscoveryConn(ctx)

	name := d.Get("name").(string)
	input := &servicediscovery.CreatePrivateDnsNamespaceInput{
		CreatorRequestId: aws.String(id.UniqueId()),
		Name:             aws.String(name),
		Tags:             getTagsIn(ctx),
		Vpc:              aws.String(d.Get("vpc").(string)),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	output, err := conn.CreatePrivateDnsNamespaceWithContext(ctx, input)

	if err != nil {
		return diag.Errorf("creating Service Discovery Private DNS Namespace (%s): %s", name, err)
	}

	operation, err := WaitOperationSuccess(ctx, conn, aws.StringValue(output.OperationId))

	if err != nil {
		return diag.Errorf("waiting for Service Discovery Private DNS Namespace (%s) create: %s", name, err)
	}

	namespaceID, ok := operation.Targets[servicediscovery.OperationTargetTypeNamespace]

	if !ok {
		return diag.Errorf("creating Service Discovery Private DNS Namespace (%s): operation response missing Namespace ID", name)
	}

	d.SetId(aws.StringValue(namespaceID))

	return resourcePrivateDNSNamespaceRead(ctx, d, meta)
}

func resourcePrivateDNSNamespaceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ServiceDiscoveryConn(ctx)

	ns, err := FindNamespaceByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Service Discovery Private DNS Namespace %s not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading Service Discovery Private DNS Namespace (%s): %s", d.Id(), err)
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

func resourcePrivateDNSNamespaceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ServiceDiscoveryConn(ctx)

	if d.HasChange("description") {
		input := &servicediscovery.UpdatePrivateDnsNamespaceInput{
			Id: aws.String(d.Id()),
			Namespace: &servicediscovery.PrivateDnsNamespaceChange{
				Description: aws.String(d.Get("description").(string)),
			},
			UpdaterRequestId: aws.String(id.UniqueId()),
		}

		output, err := conn.UpdatePrivateDnsNamespaceWithContext(ctx, input)

		if err != nil {
			return diag.Errorf("updating Service Discovery Private DNS Namespace (%s): %s", d.Id(), err)
		}

		if output != nil && output.OperationId != nil {
			if _, err := WaitOperationSuccess(ctx, conn, aws.StringValue(output.OperationId)); err != nil {
				return diag.Errorf("waiting for Service Discovery Private DNS Namespace (%s) update: %s", d.Id(), err)
			}
		}
	}

	return resourcePrivateDNSNamespaceRead(ctx, d, meta)
}

func resourcePrivateDNSNamespaceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ServiceDiscoveryConn(ctx)

	log.Printf("[INFO] Deleting Service Discovery Private DNS Namespace: %s", d.Id())
	output, err := conn.DeleteNamespaceWithContext(ctx, &servicediscovery.DeleteNamespaceInput{
		Id: aws.String(d.Id()),
	})

	if err != nil {
		return diag.Errorf("deleting Service Discovery Private DNS Namespace (%s): %s", d.Id(), err)
	}

	if output != nil && output.OperationId != nil {
		if _, err := WaitOperationSuccess(ctx, conn, aws.StringValue(output.OperationId)); err != nil {
			return diag.Errorf("waiting for Service Discovery Private DNS Namespace (%s) delete: %s", d.Id(), err)
		}
	}

	return nil
}
