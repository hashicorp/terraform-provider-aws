package identitystore

import (
	"context"
	"errors"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/identitystore"
	"github.com/aws/aws-sdk-go-v2/service/identitystore/document"
	types "github.com/aws/aws-sdk-go-v2/service/identitystore/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func ResourceGroup() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceGroupCreate,
		ReadWithoutTimeout:   resourceGroupRead,
		UpdateWithoutTimeout: resourceGroupUpdate,
		DeleteWithoutTimeout: resourceGroupDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 1024),
			},
			"display_name": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 1024),
			},
			"external_ids": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"issuer": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"group_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"identity_store_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

const (
	ResNameGroup = "Group"
)

func resourceGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).IdentityStoreConn

	identityStoreId := d.Get("identity_store_id").(string)

	input := &identitystore.CreateGroupInput{
		IdentityStoreId: aws.String(identityStoreId),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("display_name"); ok {
		input.DisplayName = aws.String(v.(string))
	}

	out, err := conn.CreateGroup(ctx, input)

	if err != nil {
		return create.DiagError(names.IdentityStore, create.ErrActionCreating, ResNameGroup, d.Get("identity_store_id").(string), err)
	}

	if out == nil || out.GroupId == nil {
		return create.DiagError(names.IdentityStore, create.ErrActionCreating, ResNameGroup, d.Get("identity_store_id").(string), errors.New("empty output"))
	}

	d.SetId(aws.ToString(out.GroupId))

	return resourceGroupRead(ctx, d, meta)
}

func resourceGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).IdentityStoreConn

	input := &identitystore.DescribeGroupInput{
		GroupId:         aws.String(d.Id()),
		IdentityStoreId: aws.String(d.Get("identity_store_id").(string)),
	}

	out, err := conn.DescribeGroup(ctx, input)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] IdentityStore Group (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return create.DiagError(names.IdentityStore, create.ErrActionReading, ResNameGroup, d.Id(), err)
	}

	d.Set("group_id", out.GroupId)
	d.Set("identity_store_id", out.IdentityStoreId)
	d.Set("description", out.Description)
	d.Set("display_name", out.DisplayName)

	if err := d.Set("external_ids", flattenExternalIds(out.ExternalIds)); err != nil {
		return create.DiagError(names.IdentityStore, create.ErrActionSetting, ResNameGroup, d.Id(), err)
	}

	return nil
}

func resourceGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).IdentityStoreConn

	in := &identitystore.UpdateGroupInput{
		GroupId:         aws.String(d.Get("group_id").(string)),
		IdentityStoreId: aws.String(d.Get("identity_store_id").(string)),
		Operations:      nil,
	}

	if d.HasChange("display_name") {
		in.Operations = append(in.Operations, types.AttributeOperation{
			AttributePath:  aws.String("displayName"),
			AttributeValue: document.NewLazyDocument(d.Get("display_name").(string)),
		})
	}

	if len(in.Operations) > 0 {
		log.Printf("[DEBUG] Updating IdentityStore Group (%s): %#v", d.Id(), in)
		_, err := conn.UpdateGroup(ctx, in)
		if err != nil {
			return create.DiagError(names.IdentityStore, create.ErrActionUpdating, ResNameGroup, d.Id(), err)
		}
	}

	return resourceGroupRead(ctx, d, meta)
}

func resourceGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).IdentityStoreConn

	log.Printf("[INFO] Deleting IdentityStore Group %s", d.Id())

	input := &identitystore.DeleteGroupInput{
		GroupId:         aws.String(d.Id()),
		IdentityStoreId: aws.String(d.Get("identity_store_id").(string)),
	}

	_, err := conn.DeleteGroup(ctx, input)

	if err != nil {
		var nfe *types.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return nil
		}

		return create.DiagError(names.IdentityStore, create.ErrActionDeleting, ResNameGroup, d.Id(), err)
	}

	return nil
}

func flattenExternalIds(apiObjects []types.ExternalId) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var l []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == (types.ExternalId{}) {
			continue
		}

		l = append(l, flattenExternalId(apiObject))
	}

	return l
}

func flattenExternalId(apiObject types.ExternalId) map[string]interface{} {
	if apiObject == (types.ExternalId{}) {
		return nil
	}

	m := map[string]interface{}{}

	if v := apiObject.Id; v != nil {
		m["id"] = aws.ToString(v)
	}

	if v := apiObject.Issuer; v != nil {
		m["issuer"] = aws.ToString(v)
	}

	return m
}
