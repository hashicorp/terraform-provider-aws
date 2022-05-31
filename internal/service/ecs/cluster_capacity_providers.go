package ecs

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func ResourceClusterCapacityProviders() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceClusterCapacityProvidersPut,
		ReadContext:   resourceClusterCapacityProvidersRead,
		UpdateContext: resourceClusterCapacityProvidersPut,
		DeleteContext: resourceClusterCapacityProvidersDelete,

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
	conn := meta.(*conns.AWSClient).ECSConn

	clusterName := d.Get("cluster_name").(string)

	input := &ecs.PutClusterCapacityProvidersInput{
		Cluster:                         aws.String(clusterName),
		CapacityProviders:               flex.ExpandStringSet(d.Get("capacity_providers").(*schema.Set)),
		DefaultCapacityProviderStrategy: expandCapacityProviderStrategy(d.Get("default_capacity_provider_strategy").(*schema.Set)),
	}

	log.Printf("[DEBUG] Updating ECS cluster capacity providers: %s", input)

	err := retryClusterCapacityProvidersPut(ctx, conn, input)

	if err != nil {
		return diag.Errorf("error updating ECS Cluster (%s) Capacity Providers: %s", clusterName, err)
	}

	if _, err := waitClusterAvailable(ctx, conn, clusterName); err != nil {
		return diag.Errorf("error waiting for ECS Cluster (%s) to become available while putting Capacity Providers: %s", clusterName, err)
	}

	d.SetId(clusterName)

	return resourceClusterCapacityProvidersRead(ctx, d, meta)
}

func resourceClusterCapacityProvidersRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ECSConn

	cluster, err := FindClusterByNameOrARN(ctx, conn, d.Id())

	if tfresource.NotFound(err) {
		diag.Errorf("[WARN] ECS Cluster (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("error reading ECS Cluster (%s): %s", d.Id(), err)
	}

	// Status==INACTIVE means deleted cluster
	if aws.StringValue(cluster.Status) == "INACTIVE" {
		diag.Errorf("[WARN] ECS Cluster (%s) deleted, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err := d.Set("capacity_providers", aws.StringValueSlice(cluster.CapacityProviders)); err != nil {
		return diag.Errorf("error setting capacity_providers: %s", err)
	}

	d.Set("cluster_name", cluster.ClusterName)

	if err := d.Set("default_capacity_provider_strategy", flattenCapacityProviderStrategy(cluster.DefaultCapacityProviderStrategy)); err != nil {
		return diag.Errorf("error setting default_capacity_provider_strategy: %s", err)
	}

	return nil
}

func resourceClusterCapacityProvidersDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ECSConn

	input := &ecs.PutClusterCapacityProvidersInput{
		Cluster:                         aws.String(d.Id()),
		CapacityProviders:               []*string{},
		DefaultCapacityProviderStrategy: []*ecs.CapacityProviderStrategyItem{},
	}

	log.Printf("[DEBUG] Removing ECS Cluster (%s) Capacity Providers", d.Id())

	err := retryClusterCapacityProvidersPut(ctx, conn, input)

	if tfawserr.ErrCodeEquals(err, ecs.ErrCodeClusterNotFoundException) {
		return nil
	}

	if err != nil {
		return diag.Errorf("error deleting ECS Cluster (%s) Capacity Providers: %s", d.Id(), err)
	}

	if _, err := waitClusterAvailable(ctx, conn, d.Id()); err != nil {
		return diag.Errorf("error waiting for ECS Cluster (%s) to become available while deleting Capacity Providers: %s", d.Id(), err)
	}

	return nil
}

func retryClusterCapacityProvidersPut(ctx context.Context, conn *ecs.ECS, input *ecs.PutClusterCapacityProvidersInput) error {
	err := resource.RetryContext(ctx, clusterUpdateTimeout, func() *resource.RetryError {
		_, err := conn.PutClusterCapacityProvidersWithContext(ctx, input)
		if err != nil {
			if tfawserr.ErrMessageContains(err, ecs.ErrCodeClientException, "Cluster was not ACTIVE") {
				return resource.RetryableError(err)
			}
			if tfawserr.ErrCodeEquals(err, ecs.ErrCodeResourceInUseException) {
				return resource.RetryableError(err)
			}
			if tfawserr.ErrCodeEquals(err, ecs.ErrCodeUpdateInProgressException) {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})

	if tfresource.TimedOut(err) {
		_, err = conn.PutClusterCapacityProvidersWithContext(ctx, input)
	}

	return err
}
