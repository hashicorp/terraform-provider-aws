// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ecs

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @SDKResource("aws_ecs_cluster_capacity_providers")
func ResourceClusterCapacityProviders() *schema.Resource {
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
			"cluster_name": {
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
						"weight": {
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
	conn := meta.(*conns.AWSClient).ECSConn(ctx)

	clusterName := d.Get("cluster_name").(string)
	input := &ecs.PutClusterCapacityProvidersInput{
		CapacityProviders:               flex.ExpandStringSet(d.Get("capacity_providers").(*schema.Set)),
		Cluster:                         aws.String(clusterName),
		DefaultCapacityProviderStrategy: expandCapacityProviderStrategy(d.Get("default_capacity_provider_strategy").(*schema.Set)),
	}

	err := retryClusterCapacityProvidersPut(ctx, conn, input)

	if err != nil {
		return diag.Errorf("updating ECS Cluster Capacity Providers (%s): %s", clusterName, err)
	}

	if d.IsNewResource() {
		d.SetId(clusterName)
	}

	if _, err := waitClusterAvailable(ctx, conn, clusterName); err != nil {
		return diag.Errorf("waiting for ECS Cluster Capacity Providers (%s) update: %s", d.Id(), err)
	}

	return resourceClusterCapacityProvidersRead(ctx, d, meta)
}

func resourceClusterCapacityProvidersRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ECSConn(ctx)

	cluster, err := FindClusterByNameOrARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		diag.Errorf("[WARN] ECS Cluster (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading ECS Cluster (%s): %s", d.Id(), err)
	}

	if err := d.Set("capacity_providers", aws.StringValueSlice(cluster.CapacityProviders)); err != nil {
		return diag.Errorf("setting capacity_providers: %s", err)
	}
	d.Set("cluster_name", cluster.ClusterName)
	if err := d.Set("default_capacity_provider_strategy", flattenCapacityProviderStrategy(cluster.DefaultCapacityProviderStrategy)); err != nil {
		return diag.Errorf("setting default_capacity_provider_strategy: %s", err)
	}

	return nil
}

func resourceClusterCapacityProvidersDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ECSConn(ctx)

	input := &ecs.PutClusterCapacityProvidersInput{
		CapacityProviders:               []*string{},
		Cluster:                         aws.String(d.Id()),
		DefaultCapacityProviderStrategy: []*ecs.CapacityProviderStrategyItem{},
	}

	log.Printf("[DEBUG] Deleting ECS Cluster Capacity Providers: %s", d.Id())
	err := retryClusterCapacityProvidersPut(ctx, conn, input)

	if tfawserr.ErrCodeEquals(err, ecs.ErrCodeClusterNotFoundException) {
		return nil
	}

	if err != nil {
		return diag.Errorf("deleting ECS Cluster Capacity Providers (%s): %s", d.Id(), err)
	}

	if _, err := waitClusterAvailable(ctx, conn, d.Id()); err != nil {
		return diag.Errorf("waiting for ECS Cluster Capacity Providers (%s) delete: %s", d.Id(), err)
	}

	return nil
}

func retryClusterCapacityProvidersPut(ctx context.Context, conn *ecs.ECS, input *ecs.PutClusterCapacityProvidersInput) error {
	_, err := tfresource.RetryWhen(ctx, clusterUpdateTimeout,
		func() (interface{}, error) {
			return conn.PutClusterCapacityProvidersWithContext(ctx, input)
		},
		func(err error) (bool, error) {
			if tfawserr.ErrMessageContains(err, ecs.ErrCodeClientException, "Cluster was not ACTIVE") {
				return true, err
			}

			if tfawserr.ErrCodeEquals(err, ecs.ErrCodeResourceInUseException, ecs.ErrCodeUpdateInProgressException) {
				return true, err
			}

			return false, err
		},
	)

	return err
}
