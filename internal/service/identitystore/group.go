package identitystore

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/identitystore"
	"github.com/aws/aws-sdk-go-v2/service/identitystore/document"
	"github.com/aws/aws-sdk-go-v2/service/identitystore/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
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
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 1024),
			},
			"external_ids": {
				Type:     schema.TypeList,
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
	conn := meta.(*conns.AWSClient).IdentityStoreClient()

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

	d.SetId(fmt.Sprintf("%s/%s", aws.ToString(out.IdentityStoreId), aws.ToString(out.GroupId)))

	return resourceGroupRead(ctx, d, meta)
}

func resourceGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).IdentityStoreClient()

	identityStoreId, groupId, err := resourceGroupParseID(d.Id())

	if err != nil {
		return create.DiagError(names.IdentityStore, create.ErrActionReading, ResNameGroup, d.Id(), err)
	}

	out, err := FindGroupByTwoPartKey(ctx, conn, identityStoreId, groupId)

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
	conn := meta.(*conns.AWSClient).IdentityStoreClient()

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
	conn := meta.(*conns.AWSClient).IdentityStoreClient()

	log.Printf("[INFO] Deleting IdentityStore Group %s", d.Id())
	_, err := conn.DeleteGroup(ctx, &identitystore.DeleteGroupInput{
		IdentityStoreId: aws.String(d.Get("identity_store_id").(string)),
		GroupId:         aws.String(d.Get("group_id").(string)),
	})

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return nil
	}

	if err != nil {
		return create.DiagError(names.IdentityStore, create.ErrActionDeleting, ResNameGroup, d.Id(), err)
	}

	return nil
}

func FindGroupByTwoPartKey(ctx context.Context, conn *identitystore.Client, identityStoreID, groupID string) (*identitystore.DescribeGroupOutput, error) {
	in := &identitystore.DescribeGroupInput{
		GroupId:         aws.String(groupID),
		IdentityStoreId: aws.String(identityStoreID),
	}

	out, err := conn.DescribeGroup(ctx, in)

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}

	if err != nil {
		return nil, err
	}

	if out == nil || out.GroupId == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
}

func resourceGroupParseID(id string) (identityStoreId, groupId string, err error) {
	parts := strings.Split(id, "/")

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		err = errors.New("expected a resource id in the form: identity-store-id/group-id")
		return
	}

	return parts[0], parts[1], nil
}
