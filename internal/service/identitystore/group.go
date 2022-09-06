package identitystore

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/identitystore"
	types "github.com/aws/aws-sdk-go-v2/service/identitystore/types"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func ResourceGroup() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceGroupCreate,
		ReadWithoutTimeout:   resourceGroupRead,
		DeleteWithoutTimeout: resourceGroupDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 1024),
			},
			"display_name": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
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

	output, err := conn.CreateGroup(ctx, input)

	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating Identity Store Group (%s): %w", identityStoreId, err))
	}

	if output == nil || output.GroupId == nil {
		return diag.FromErr(fmt.Errorf("error creating Identity Store Group (%s): empty output", identityStoreId))
	}

	d.SetId(aws.StringValue(output.GroupId))

	return resourceGroupRead(ctx, d, meta)
}

func resourceGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).IdentityStoreConn

	input := &identitystore.DescribeGroupInput{
		GroupId:         aws.String(d.Id()),
		IdentityStoreId: aws.String(d.Get("identity_store_id").(string)),
	}

	output, err := conn.DescribeGroup(ctx, input)

	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading Identity Store Group (%s): %w", d.Id(), err))
	}

	if output == nil || output.GroupId == nil {
		return diag.FromErr(fmt.Errorf("error reading Identity Store Group (%s): empty output", d.Id()))
	}

	d.Set("group_id", output.GroupId)
	d.Set("identity_store_id", output.IdentityStoreId)
	d.Set("description", output.Description)
	d.Set("display_name", output.DisplayName)

	if err := d.Set("external_ids", flattenExternalIds(output.ExternalIds)); err != nil {
		return diag.Errorf("error setting external_ids: %s", err)
	}

	return nil
}

func resourceGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).IdentityStoreConn

	input := &identitystore.DeleteGroupInput{
		GroupId:         aws.String(d.Id()),
		IdentityStoreId: aws.String(d.Get("identity_store_id").(string)),
	}

	_, err := conn.DeleteGroup(ctx, input)

	if err != nil {
		return diag.FromErr(fmt.Errorf("error deleting Identity Store Group (%s): %w", d.Id(), err))
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
		m["id"] = aws.StringValue(v)
	}

	if v := apiObject.Issuer; v != nil {
		m["issuer"] = aws.StringValue(v)
	}

	return m
}
