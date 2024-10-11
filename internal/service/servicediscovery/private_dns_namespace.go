// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package servicediscovery

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/servicediscovery"
	awstypes "github.com/aws/aws-sdk-go-v2/service/servicediscovery/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_service_discovery_private_dns_namespace", name="Private DNS Namespace")
// @Tags(identifierAttribute="arn")
func resourcePrivateDNSNamespace() *schema.Resource {
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
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Optional: true,
			},
			"hosted_zone": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrName: {
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
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ServiceDiscoveryClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &servicediscovery.CreatePrivateDnsNamespaceInput{
		CreatorRequestId: aws.String(id.UniqueId()),
		Name:             aws.String(name),
		Tags:             getTagsIn(ctx),
		Vpc:              aws.String(d.Get("vpc").(string)),
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	output, err := conn.CreatePrivateDnsNamespace(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Service Discovery Private DNS Namespace (%s): %s", name, err)
	}

	operation, err := waitOperationSucceeded(ctx, conn, aws.ToString(output.OperationId))

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Service Discovery Private DNS Namespace (%s) create: %s", name, err)
	}

	d.SetId(operation.Targets[string(awstypes.OperationTargetTypeNamespace)])

	return append(diags, resourcePrivateDNSNamespaceRead(ctx, d, meta)...)
}

func resourcePrivateDNSNamespaceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ServiceDiscoveryClient(ctx)

	ns, err := findNamespaceByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Service Discovery Private DNS Namespace %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Service Discovery Private DNS Namespace (%s): %s", d.Id(), err)
	}

	arn := aws.ToString(ns.Arn)
	d.Set(names.AttrARN, arn)
	d.Set(names.AttrDescription, ns.Description)
	if ns.Properties != nil && ns.Properties.DnsProperties != nil {
		d.Set("hosted_zone", ns.Properties.DnsProperties.HostedZoneId)
	} else {
		d.Set("hosted_zone", nil)
	}
	d.Set(names.AttrName, ns.Name)

	return diags
}

func resourcePrivateDNSNamespaceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ServiceDiscoveryClient(ctx)

	if d.HasChange(names.AttrDescription) {
		input := &servicediscovery.UpdatePrivateDnsNamespaceInput{
			Id: aws.String(d.Id()),
			Namespace: &awstypes.PrivateDnsNamespaceChange{
				Description: aws.String(d.Get(names.AttrDescription).(string)),
			},
			UpdaterRequestId: aws.String(id.UniqueId()),
		}

		output, err := conn.UpdatePrivateDnsNamespace(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Service Discovery Private DNS Namespace (%s): %s", d.Id(), err)
		}

		if output != nil && output.OperationId != nil {
			if _, err := waitOperationSucceeded(ctx, conn, aws.ToString(output.OperationId)); err != nil {
				return sdkdiag.AppendErrorf(diags, "waiting for Service Discovery Private DNS Namespace (%s) update: %s", d.Id(), err)
			}
		}
	}

	return append(diags, resourcePrivateDNSNamespaceRead(ctx, d, meta)...)
}

func resourcePrivateDNSNamespaceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ServiceDiscoveryClient(ctx)

	log.Printf("[INFO] Deleting Service Discovery Private DNS Namespace: %s", d.Id())
	output, err := conn.DeleteNamespace(ctx, &servicediscovery.DeleteNamespaceInput{
		Id: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.NamespaceNotFound](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Service Discovery Private DNS Namespace (%s): %s", d.Id(), err)
	}

	if output != nil && output.OperationId != nil {
		if _, err := waitOperationSucceeded(ctx, conn, aws.ToString(output.OperationId)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Service Discovery Private DNS Namespace (%s) delete: %s", d.Id(), err)
		}
	}

	return diags
}
