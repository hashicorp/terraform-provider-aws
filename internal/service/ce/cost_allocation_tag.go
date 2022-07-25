package ce

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/costexplorer"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func ResourceCostAllocationTag() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceCostAllocationTagUpdate,
		ReadContext:   resourceCostAllocationTagRead,
		UpdateContext: resourceCostAllocationTagUpdate,
		DeleteContext: resourceCostAllocationTagDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"status": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(costexplorer.CostAllocationTagStatus_Values(), false),
			},
			"tag_key": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 1024),
			},
			"type": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceCostAllocationTagRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).CEConn

	costAllocTag, err := FindCostAllocationTagByKey(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		names.LogNotFoundRemoveState(names.CE, names.ErrActionReading, ResCostAllocationTag, d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return names.DiagError(names.CE, names.ErrActionReading, ResCostAllocationTag, d.Id(), err)
	}

	d.Set("tag_key", costAllocTag.TagKey)
	d.Set("status", costAllocTag.Status)
	d.Set("type", costAllocTag.Type)

	return nil
}

func resourceCostAllocationTagUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	key := d.Get("tag_key").(string)

	updateTagStatus(ctx, d, meta, false)

	d.SetId(key)

	return resourceCostAllocationTagRead(ctx, d, meta)
}

func resourceCostAllocationTagDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return updateTagStatus(ctx, d, meta, true)
}

func updateTagStatus(ctx context.Context, d *schema.ResourceData, meta interface{}, delete bool) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).CEConn

	key := d.Get("tag_key").(string)
	tagStatus := &costexplorer.CostAllocationTagStatusEntry{
		TagKey: aws.String(key),
		Status: aws.String(d.Get("status").(string)),
	}

	if delete {
		tagStatus.Status = aws.String(costexplorer.CostAllocationTagStatusInactive)
	}

	input := &costexplorer.UpdateCostAllocationTagsStatusInput{
		CostAllocationTagsStatus: []*costexplorer.CostAllocationTagStatusEntry{tagStatus},
	}

	_, err := conn.UpdateCostAllocationTagsStatusWithContext(ctx, input)

	if err != nil {
		return names.DiagError(names.CE, names.ErrActionUpdating, ResCostAllocationTag, d.Id(), err)
	}

	return nil
}
