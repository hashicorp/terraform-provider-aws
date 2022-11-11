package identitystore

import (
	"context"
	"errors"
	"regexp"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/identitystore"
	"github.com/aws/aws-sdk-go-v2/service/identitystore/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func DataSourceGroup() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceGroupRead,

		Schema: map[string]*schema.Schema{
			"alternate_identifier": {
				Type:          schema.TypeList,
				Optional:      true,
				MaxItems:      1,
				ConflictsWith: []string{"filter", "group_id"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"external_id": {
							Type:         schema.TypeList,
							Optional:     true,
							MaxItems:     1,
							ExactlyOneOf: []string{"alternate_identifier.0.external_id", "alternate_identifier.0.unique_attribute"},
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"id": {
										Type:     schema.TypeString,
										Required: true,
									},
									"issuer": {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
						},
						"unique_attribute": {
							Type:         schema.TypeList,
							Optional:     true,
							MaxItems:     1,
							ExactlyOneOf: []string{"alternate_identifier.0.external_id", "alternate_identifier.0.unique_attribute"},
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"attribute_path": {
										Type:     schema.TypeString,
										Required: true,
									},
									"attribute_value": {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
						},
					},
				},
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"display_name": {
				Type:     schema.TypeString,
				Computed: true,
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
			"filter": {
				Deprecated:    "Use the alternate_identifier attribute instead.",
				Type:          schema.TypeList,
				Optional:      true,
				MaxItems:      1,
				AtLeastOneOf:  []string{"alternate_identifier", "filter", "group_id"},
				ConflictsWith: []string{"alternate_identifier"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"attribute_path": {
							Type:     schema.TypeString,
							Required: true,
						},
						"attribute_value": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"group_id": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				AtLeastOneOf:  []string{"alternate_identifier", "filter", "group_id"},
				ConflictsWith: []string{"alternate_identifier"},
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 47),
					validation.StringMatch(regexp.MustCompile(`^([0-9a-f]{10}-|)[A-Fa-f0-9]{8}-[A-Fa-f0-9]{4}-[A-Fa-f0-9]{4}-[A-Fa-f0-9]{4}-[A-Fa-f0-9]{12}$`), "must match ([0-9a-f]{10}-|)[A-Fa-f0-9]{8}-[A-Fa-f0-9]{4}-[A-Fa-f0-9]{4}-[A-Fa-f0-9]{4}-[A-Fa-f0-9]{12}"),
				),
			},
			"identity_store_id": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 64),
					validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9-]*$`), "must match [a-zA-Z0-9-]"),
				),
			},
		},
	}
}

const (
	DSNameGroup = "Group Data Source"
)

func dataSourceGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).IdentityStoreClient

	identityStoreId := d.Get("identity_store_id").(string)

	var getGroupIdInput *identitystore.GetGroupIdInput

	if v, ok := d.GetOk("alternate_identifier"); ok && len(v.([]interface{})) > 0 {
		getGroupIdInput = &identitystore.GetGroupIdInput{
			AlternateIdentifier: expandAlternateIdentifier(v.([]interface{})[0].(map[string]interface{})),
			IdentityStoreId:     aws.String(identityStoreId),
		}
	} else if v, ok := d.GetOk("filter"); ok && len(v.([]interface{})) > 0 {
		getGroupIdInput = &identitystore.GetGroupIdInput{
			AlternateIdentifier: &types.AlternateIdentifierMemberUniqueAttribute{
				Value: *expandUniqueAttribute(v.([]interface{})[0].(map[string]interface{})),
			},
			IdentityStoreId: aws.String(identityStoreId),
		}
	}

	var groupId string

	if getGroupIdInput != nil {
		output, err := conn.GetGroupId(ctx, getGroupIdInput)

		if err != nil {
			var e *types.ResourceNotFoundException
			if errors.As(err, &e) {
				return diag.Errorf("no Identity Store Group found matching criteria; try different search")
			} else {
				return create.DiagError(names.IdentityStore, create.ErrActionReading, DSNameGroup, identityStoreId, err)
			}
		}

		groupId = aws.ToString(output.GroupId)
	}

	if v, ok := d.GetOk("group_id"); ok && v.(string) != "" {
		if groupId != "" && groupId != v.(string) {
			// We were given a filter, and it found a group different to this one.
			return diag.Errorf("no Identity Store Group found matching criteria; try different search")
		}

		groupId = v.(string)
	}

	group, err := findGroupByID(ctx, conn, identityStoreId, groupId)

	if err != nil {
		if tfresource.NotFound(err) {
			return diag.Errorf("no Identity Store Group found matching criteria; try different search")
		}

		return create.DiagError(names.IdentityStore, create.ErrActionReading, DSNameGroup, identityStoreId, err)
	}

	d.SetId(aws.ToString(group.GroupId))

	d.Set("description", group.Description)
	d.Set("display_name", group.DisplayName)
	d.Set("group_id", group.GroupId)

	if err := d.Set("external_ids", flattenExternalIds(group.ExternalIds)); err != nil {
		return create.DiagError(names.IdentityStore, create.ErrActionSetting, DSNameGroup, d.Id(), err)
	}

	return nil
}
