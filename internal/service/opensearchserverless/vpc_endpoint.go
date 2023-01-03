package opensearchserverless

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/opensearchserverless"
	"github.com/aws/aws-sdk-go-v2/service/opensearchserverless/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func ResourceVPCEndpoint() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceVPCEndpointCreate,
		ReadWithoutTimeout:   resourceVPCEndpointRead,
		UpdateWithoutTimeout: resourceVPCEndpointUpdate,
		DeleteWithoutTimeout: resourceVPCEndpointDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(3, 32),
				ForceNew:     true,
			},
			"security_group_ids": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				MinItems: 1,
				MaxItems: 5,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"subnet_ids": {
				Type:     schema.TypeSet,
				Required: true,
				MinItems: 1,
				MaxItems: 6,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"vpc_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
			},
		},
	}
}

const (
	ResNameVPCEndpoint = "VPC Endpoint"
)

func resourceVPCEndpointCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).OpenSearchServerlessClient()

	in := &opensearchserverless.CreateVpcEndpointInput{
		ClientToken: aws.String(resource.UniqueId()),
		Name:        aws.String(d.Get("name").(string)),
		VpcId:       aws.String(d.Get("vpc_id").(string)),
		SubnetIds:   flex.ExpandStringValueSet(d.Get("subnet_ids").(*schema.Set)),
	}

	if v, ok := d.GetOk("security_group_ids"); ok {
		in.SecurityGroupIds = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	out, err := conn.CreateVpcEndpoint(ctx, in)
	if err != nil {
		return create.DiagError(names.OpenSearchServerless, create.ErrActionCreating, ResNameVPCEndpoint, d.Get("name").(string), err)
	}

	if out == nil || out.CreateVpcEndpointDetail == nil {
		return create.DiagError(names.OpenSearchServerless, create.ErrActionCreating, ResNameVPCEndpoint, d.Get("name").(string), errors.New("empty output"))
	}

	d.SetId(aws.ToString(out.CreateVpcEndpointDetail.Id))

	if _, err := waitVPCEndpointCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return create.DiagError(names.OpenSearchServerless, create.ErrActionWaitingForCreation, ResNameVPCEndpoint, d.Id(), err)
	}

	return resourceVPCEndpointRead(ctx, d, meta)
}

func resourceVPCEndpointRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).OpenSearchServerlessClient()
	out, err := findVPCEndpointByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] OpenSearchServerless VpcEndpoint (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return create.DiagError(names.OpenSearchServerless, create.ErrActionReading, ResNameVPCEndpoint, d.Id(), err)
	}

	d.Set("name", out.Name)
	d.Set("security_group_ids", flex.FlattenStringValueSet(out.SecurityGroupIds))
	d.Set("subnet_ids", flex.FlattenStringValueSet(out.SubnetIds))
	d.Set("vpc_id", out.VpcId)

	return nil
}

func resourceVPCEndpointUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).OpenSearchServerlessClient()

	update := false

	in := &opensearchserverless.UpdateVpcEndpointInput{
		ClientToken: aws.String(resource.UniqueId()),
		Id:          aws.String(d.Id()),
	}

	if d.HasChange("security_group_ids") {
		old, new := d.GetChange("security_group_ids")
		if add := flex.ExpandStringValueSet(new.(*schema.Set).Difference(old.(*schema.Set))); len(add) > 0 {
			in.AddSecurityGroupIds = add
		}

		if del := flex.ExpandStringValueSet(old.(*schema.Set).Difference(new.(*schema.Set))); len(del) > 0 {
			in.RemoveSecurityGroupIds = del
		}
		update = true
	}

	if d.HasChange("subnet_ids") {
		old, new := d.GetChange("subnet_ids")
		if add := flex.ExpandStringValueSet(new.(*schema.Set).Difference(old.(*schema.Set))); len(add) > 0 {
			in.AddSubnetIds = add
		}

		if del := flex.ExpandStringValueSet(old.(*schema.Set).Difference(new.(*schema.Set))); len(del) > 0 {
			in.RemoveSubnetIds = del
		}
		update = true
	}

	if !update {
		return nil
	}

	log.Printf("[DEBUG] Updating OpenSearchServerless VpcEndpoint (%s): %#v", d.Id(), in)
	out, err := conn.UpdateVpcEndpoint(ctx, in)
	if err != nil {
		return create.DiagError(names.OpenSearchServerless, create.ErrActionUpdating, ResNameVPCEndpoint, d.Id(), err)
	}

	if _, err := waitVPCEndpointUpdated(ctx, conn, string(out.UpdateVpcEndpointDetail.Status), d.Timeout(schema.TimeoutUpdate)); err != nil {
		return create.DiagError(names.OpenSearchServerless, create.ErrActionWaitingForUpdate, ResNameVPCEndpoint, d.Id(), err)
	}

	return resourceVPCEndpointRead(ctx, d, meta)
}

func resourceVPCEndpointDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).OpenSearchServerlessClient()

	log.Printf("[INFO] Deleting OpenSearchServerless VpcEndpoint %s", d.Id())

	_, err := conn.DeleteVpcEndpoint(ctx, &opensearchserverless.DeleteVpcEndpointInput{
		ClientToken: aws.String(resource.UniqueId()),
		Id:          aws.String(d.Id()),
	})

	if err != nil {
		var nfe *types.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return nil
		}

		return create.DiagError(names.OpenSearchServerless, create.ErrActionDeleting, ResNameVPCEndpoint, d.Id(), err)
	}

	if _, err := waitVPCEndpointDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return create.DiagError(names.OpenSearchServerless, create.ErrActionWaitingForDeletion, ResNameVPCEndpoint, d.Id(), err)
	}

	return nil
}
