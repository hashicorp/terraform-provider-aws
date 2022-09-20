package identitystore

import (
	"context"
	"errors"
	"fmt"
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
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
	"log"
	"regexp"
	"strings"
)

func ResourceUser() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceUserCreate,
		ReadWithoutTimeout:   resourceUserRead,
		UpdateWithoutTimeout: resourceUserUpdate,
		DeleteWithoutTimeout: resourceUserDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"display_name": {
				Type:     schema.TypeString,
				Required: true,
				ValidateDiagFunc: validation.ToDiagFunc(validation.All(
					validation.StringLenBetween(1, 1024),
					validation.StringMatch(regexp.MustCompile(`^[\p{L}\p{M}\p{S}\p{N}\p{P}\t\n\r  　]+$`), "must be a printable name"),
				)),
			},
			"emails": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"primary": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"type": {
							Type:     schema.TypeString,
							Optional: true,
							ValidateDiagFunc: validation.ToDiagFunc(validation.All(
								validation.StringLenBetween(1, 1024),
								validation.StringMatch(regexp.MustCompile(`^[\p{L}\p{M}\p{S}\p{N}\p{P}\t\n\r  　]+$`), "must be a printable type"),
							)),
						},
						"value": {
							Type:     schema.TypeString,
							Optional: true,
							ValidateDiagFunc: validation.ToDiagFunc(validation.All(
								validation.StringLenBetween(1, 1024),
								validation.StringMatch(regexp.MustCompile(`^[\p{L}\p{M}\p{S}\p{N}\p{P}\t\n\r  　]+$`), "must be a printable email"),
							)),
						},
					},
				},
			},
			"identity_store_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"name": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"family_name": {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: resourceUserValidateName,
						},
						"formatted": {
							Type:             schema.TypeString,
							Optional:         true,
							ValidateDiagFunc: resourceUserValidateName,
						},
						"given_name": {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: resourceUserValidateName,
						},
						},
					},
				},
			},
			"user_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"user_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateDiagFunc: validation.ToDiagFunc(validation.All(
					validation.StringLenBetween(1, 128),
					validation.StringMatch(regexp.MustCompile(`^[\p{L}\p{M}\p{S}\p{N}\p{P}]+$`), "must be a user name"),
				)),
			},
		},
	}
}

const (
	ResNameUser = "User"
)

func resourceUserCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).IdentityStoreConn

	in := &identitystore.CreateUserInput{
		DisplayName:     aws.String(d.Get("display_name").(string)),
		IdentityStoreId: aws.String(d.Get("identity_store_id").(string)),
		UserName:        aws.String(d.Get("user_name").(string)),
	}

	if v, ok := d.GetOk("emails"); ok && len(v.([]interface{})) > 0 {
		in.Emails = expandEmails(v.([]interface{}))
	}

	if v, ok := d.GetOk("name"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		in.Name = expandName(v.([]interface{})[0].(map[string]interface{}))
	}

	out, err := conn.CreateUser(ctx, in)
	if err != nil {
		return create.DiagError(names.IdentityStore, create.ErrActionCreating, ResNameUser, d.Get("identity_store_id").(string), err)
	}

	if out == nil || out.UserId == nil {
		return create.DiagError(names.IdentityStore, create.ErrActionCreating, ResNameUser, d.Get("identity_store_id").(string), errors.New("empty output"))
	}

	d.SetId(fmt.Sprintf("%s/%s", aws.ToString(out.IdentityStoreId), aws.ToString(out.UserId)))

	return resourceUserRead(ctx, d, meta)
}

func resourceUserRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).IdentityStoreConn

	identityStoreId, userId, err := resourceUserParseID(d.Id())

	if err != nil {
		return create.DiagError(names.IdentityStore, create.ErrActionReading, ResNameUser, d.Id(), err)
	}

	out, err := findUserByID(ctx, conn, identityStoreId, userId)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] IdentityStore User (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return create.DiagError(names.IdentityStore, create.ErrActionReading, ResNameUser, d.Id(), err)
	}

	d.Set("display_name", out.DisplayName)
	d.Set("identity_store_id", out.IdentityStoreId)
	d.Set("user_id", out.UserId)
	d.Set("user_name", out.UserName)

	if err := d.Set("emails", flattenEmails(out.Emails)); err != nil {
		return create.DiagError(names.IdentityStore, create.ErrActionSetting, ResNameUser, d.Id(), err)
	}

	if err := d.Set("name", []interface{}{flattenName(out.Name)}); err != nil {
		return create.DiagError(names.IdentityStore, create.ErrActionSetting, ResNameUser, d.Id(), err)
	}

	return nil
}

func resourceUserUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).IdentityStoreConn

	in := &identitystore.UpdateUserInput{
		IdentityStoreId: aws.String(d.Get("identity_store_id").(string)),
		UserId:          aws.String(d.Get("user_id").(string)),
		Operations:      nil,
	}

	// IMPLEMENTATION NOTE.
	//
	// Complex types, such as the `emails` field, don't allow field by field
	// updates, and require that the entire sub-object is modified.
	//
	// In those sub-objects, to remove a field, it must not be present at all
	// in the updated attribute value.
	//
	// However, structs such as types.Email don't specify omitempty in their
	// struct tags, so the document.NewLazyDocument marshaller will write out
	// nulls.
	//
	// This is why, for those complex fields, a custom Expand function is
	// provided that converts the Go SDK type (e.g. types.Email) into a field
	// by field representation of what the API would expect.

	fieldsToUpdate := []struct {
		Attribute string
		Field     string
		Expand    func(interface{}) interface{}
	}{
		{
			Attribute: "display_name",
			Field:     "displayName",
		},
		{
			Attribute: "name.0.family_name",
			Field:     "name.familyName",
		},
		{
			Attribute: "name.0.formatted",
			Field:     "name.formatted",
		},
		{
			Attribute: "name.0.given_name",
			Field:     "name.givenName",
		},
		{
			Attribute: "emails",
			Field:     "emails",
			Expand: func(value interface{}) interface{} {
				emails := expandEmails(value.([]interface{}))

				var result []interface{}

				// The API requires a null to unset the list, so in the case
				// of no emails, a nil result is preferable.
				for _, email := range emails {
					m := map[string]interface{}{}

					m["primary"] = email.Primary

					if v := email.Type; v != nil {
						m["type"] = v
					}

					if v := email.Value; v != nil {
						m["value"] = v
					}

					result = append(result, m)
				}

				return result
			},
		},
	}

	for _, fieldToUpdate := range fieldsToUpdate {
		if d.HasChange(fieldToUpdate.Attribute) {
			value := d.Get(fieldToUpdate.Attribute)

			if expand := fieldToUpdate.Expand; expand != nil {
				value = expand(value)
			}

			in.Operations = append(in.Operations, types.AttributeOperation{
				AttributePath:  aws.String(fieldToUpdate.Field),
				AttributeValue: document.NewLazyDocument(value),
			})
		}
	}

	if len(in.Operations) > 0 {
		log.Printf("[DEBUG] Updating IdentityStore User (%s): %#v", d.Id(), in)
		_, err := conn.UpdateUser(ctx, in)
		if err != nil {
			return create.DiagError(names.IdentityStore, create.ErrActionUpdating, ResNameUser, d.Id(), err)
		}
	}

	return resourceUserRead(ctx, d, meta)
}

func resourceUserDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).IdentityStoreConn

	log.Printf("[INFO] Deleting IdentityStore User %s", d.Id())

	_, err := conn.DeleteUser(ctx, &identitystore.DeleteUserInput{
		IdentityStoreId: aws.String(d.Get("identity_store_id").(string)),
		UserId:          aws.String(d.Get("user_id").(string)),
	})

	if err != nil {
		var nfe *types.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return nil
		}

		return create.DiagError(names.IdentityStore, create.ErrActionDeleting, ResNameUser, d.Id(), err)
	}

	return nil
}

func findUserByID(ctx context.Context, conn *identitystore.Client, identityStoreId, userId string) (*identitystore.DescribeUserOutput, error) {
	in := &identitystore.DescribeUserInput{
		IdentityStoreId: aws.String(identityStoreId),
		UserId:          aws.String(userId),
	}

	out, err := conn.DescribeUser(ctx, in)

	if err != nil {
		var e *types.ResourceNotFoundException
		if errors.As(err, &e) {
			return nil, &resource.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		} else {
			return nil, err
		}
	}

	if out == nil || out.UserId == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
}

func flattenName(apiObject *types.Name) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := apiObject.FamilyName; v != nil {
		m["family_name"] = aws.ToString(v)
	}

	if v := apiObject.Formatted; v != nil {
		m["formatted"] = aws.ToString(v)
	}

	if v := apiObject.GivenName; v != nil {
		m["given_name"] = aws.ToString(v)
	}

	return m
}

func expandName(tfMap map[string]interface{}) *types.Name {
	if tfMap == nil {
		return nil
	}

	a := &types.Name{}

	if v, ok := tfMap["family_name"].(string); ok && v != "" {
		a.FamilyName = aws.String(v)
	}

	if v, ok := tfMap["formatted"].(string); ok && v != "" {
		a.Formatted = aws.String(v)
	}

	if v, ok := tfMap["given_name"].(string); ok && v != "" {
		a.GivenName = aws.String(v)
	}

	return a
}

func flattenEmail(apiObject *types.Email) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{}

	m["primary"] = apiObject.Primary

	if v := apiObject.Type; v != nil {
		m["type"] = aws.ToString(v)
	}

	if v := apiObject.Value; v != nil {
		m["value"] = aws.ToString(v)
	}

	return m
}

func expandEmail(tfMap map[string]interface{}) *types.Email {
	if tfMap == nil {
		return nil
	}

	a := &types.Email{}

	a.Primary = tfMap["primary"].(bool)

	if v, ok := tfMap["type"].(string); ok && v != "" {
		a.Type = aws.String(v)
	}

	if v, ok := tfMap["value"].(string); ok && v != "" {
		a.Value = aws.String(v)
	}

	return a
}

func flattenEmails(apiObjects []types.Email) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var l []interface{}

	for _, apiObject := range apiObjects {
		apiObject := apiObject
		l = append(l, flattenEmail(&apiObject))
	}

	return l
}

func expandEmails(tfList []interface{}) []types.Email {
	s := make([]types.Email, 0, len(tfList))

	for _, r := range tfList {
		m, ok := r.(map[string]interface{})

		if !ok {
			continue
		}

		a := expandEmail(m)

		if a == nil {
			continue
		}

		s = append(s, *a)
	}

	return s
}

func resourceUserParseID(id string) (identityStoreId, userId string, err error) {
	parts := strings.Split(id, "/")

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		err = errors.New("???")
		return
	}

	return parts[0], parts[1], nil
}

var resourceUserValidateName = validation.ToDiagFunc(validation.All(
	validation.StringLenBetween(1, 1024),
	validation.StringMatch(regexp.MustCompile(`^[\p{L}\p{M}\p{S}\p{N}\p{P}\t\n\r  　]+$`), "must be a printable name"),
))
