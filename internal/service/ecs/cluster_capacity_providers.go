// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ecs

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_ecs_cluster_capacity_providers", name="Cluster Capacity Providers")
func resourceClusterCapacityProviders() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceClusterCapacityProvidersPut,
		ReadWithoutTimeout:   resourceClusterCapacityProvidersRead,
		UpdateWithoutTimeout: resourceClusterCapacityProvidersPut,
		DeleteWithoutTimeout: resourceClusterCapacityProvidersDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"capacity_providers": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			names.AttrClusterName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				// The API accepts both an ARN and a name in a generic "cluster"
				// parameter, but allowing that would force the resource to guess
				// which one to return on read.
				ValidateFunc: validateClusterName,
			},
			"default_capacity_provider_strategy": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"base": {
							Type:         schema.TypeInt,
							Default:      0,
							Optional:     true,
							ValidateFunc: validation.IntBetween(0, 100000),
						},
						"capacity_provider": {
							Type:     schema.TypeString,
							Required: true,
						},
						names.AttrWeight: {
							Type:         schema.TypeInt,
							Default:      0,
							Optional:     true,
							ValidateFunc: validation.IntBetween(0, 1000),
						},
					},
				},
			},
		},
	}
}

func resourceClusterCapacityProvidersPut(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).ECSClient(ctx)

	clusterName := d.Get(names.AttrClusterName).(string)
	input := &ecs.PutClusterCapacityProvidersInput{
		CapacityProviders:               flex.ExpandStringValueSet(d.Get("capacity_providers").(*schema.Set)),
		Cluster:                         aws.String(clusterName),
		DefaultCapacityProviderStrategy: expandCapacityProviderStrategyItems(d.Get("default_capacity_provider_strategy").(*schema.Set)),
	}

	err := retryClusterCapacityProvidersPut(ctx, conn, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating ECS Cluster Capacity Providers (%s): %s", clusterName, err)
	}

	if d.IsNewResource() {
		d.SetId(clusterName)
	}

	if _, err := waitClusterAvailable(ctx, conn, clusterName); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for ECS Cluster Capacity Providers (%s) update: %s", d.Id(), err)
	}

	return append(diags, resourceClusterCapacityProvidersRead(ctx, d, meta)...)
}

func resourceClusterCapacityProvidersRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ECSClient(ctx)

	cluster, err := findClusterByNameOrARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		sdkdiag.AppendErrorf(diags, "[WARN] ECS Cluster Capacity Providers (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ECS Cluster (%s): %s", d.Id(), err)
	}

	if err := d.Set("capacity_providers", cluster.CapacityProviders); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting capacity_providers: %s", err)
	}
	d.Set(names.AttrClusterName, cluster.ClusterName)
	if err := d.Set("default_capacity_provider_strategy", flattenCapacityProviderStrategyItems(cluster.DefaultCapacityProviderStrategy)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting default_capacity_provider_strategy: %s", err)
	}

	return diags
}

func resourceClusterCapacityProvidersDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ECSClient(ctx)

	input := &ecs.PutClusterCapacityProvidersInput{
		CapacityProviders:               []string{},
		Cluster:                         aws.String(d.Id()),
		DefaultCapacityProviderStrategy: []awstypes.CapacityProviderStrategyItem{},
	}

	log.Printf("[DEBUG] Deleting ECS Cluster Capacity Providers: %s", d.Id())
	err := retryClusterCapacityProvidersPut(ctx, conn, input)

	if errs.IsA[*awstypes.ClusterNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting ECS Cluster Capacity Providers (%s): %s", d.Id(), err)
	}

	if _, err := waitClusterAvailable(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for ECS Cluster Capacity Providers (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func retryClusterCapacityProvidersPut(ctx context.Context, conn *ecs.Client, input *ecs.PutClusterCapacityProvidersInput) error {
	const (
		timeout = 10 * time.Minute
	)
	_, err := tfresource.RetryWhen(ctx, timeout,
		func() (interface{}, error) {
			return conn.PutClusterCapacityProviders(ctx, input)
		},
		func(err error) (bool, error) {
			if errs.IsAErrorMessageContains[*awstypes.ClientException](err, "Cluster was not ACTIVE") {
				return true, err
			}

			if errs.IsA[*awstypes.ResourceInUseException](err) || errs.IsA[*awstypes.UpdateInProgressException](err) {
				return true, err
			}

			return false, err
		},
	)

	return err
}
