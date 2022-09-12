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
	"os"
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
							Type:     schema.TypeString,
							Required: true,
							ValidateDiagFunc: validation.ToDiagFunc(validation.All(
								validation.StringLenBetween(1, 1024),
								validation.StringMatch(regexp.MustCompile(`^[\p{L}\p{M}\p{S}\p{N}\p{P}\t\n\r  　]+$`), "must be a printable name"),
							)),
						},
						"given_name": {
							Type:     schema.TypeString,
							Required: true,
							ValidateDiagFunc: validation.ToDiagFunc(validation.All(
								validation.StringLenBetween(1, 1024),
								validation.StringMatch(regexp.MustCompile(`^[\p{L}\p{M}\p{S}\p{N}\p{P}\t\n\r  　]+$`), "must be a printable name"),
							)),
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

var lgr = log.New(os.Stderr, "DEBUG - ", 0)

func resourceUserUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).IdentityStoreConn

	in := &identitystore.UpdateUserInput{
		IdentityStoreId: aws.String(d.Get("identity_store_id").(string)),
		UserId:          aws.String(d.Get("user_id").(string)),
		Operations:      nil,
	}

	if d.HasChange("display_name") {
		lgr.Printf("display_name changed")
		in.Operations = append(in.Operations, types.AttributeOperation{
			AttributePath:  aws.String("displayName"),
			AttributeValue: document.NewLazyDocument(d.Get("display_name").(string)),
		})
	}

	if d.HasChange("emails") {
		lgr.Printf("emails changed")

		emails := expandEmails(d.Get("emails").([]interface{}))

		if len(emails) == 0 {
			emails = nil // The API requires a null to unset the field.
		}

		in.Operations = append(in.Operations, types.AttributeOperation{
			AttributePath:  aws.String("emails"),
			AttributeValue: document.NewLazyDocument(emails),
		})
	}

	if d.HasChange("name.0.family_name") {
		lgr.Printf("family_name changed")
		in.Operations = append(in.Operations, types.AttributeOperation{
			AttributePath:  aws.String("name.familyName"),
			AttributeValue: document.NewLazyDocument(d.Get("name.0.family_name").(string)),
		})
	}

	if d.HasChange("name.0.given_name") {
		lgr.Printf("given_name changed")
		in.Operations = append(in.Operations, types.AttributeOperation{
			AttributePath:  aws.String("name.givenName"),
			AttributeValue: document.NewLazyDocument(d.Get("name.0.given_name").(string)),
		})
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

	lgr.Printf("%+v %+v %+v", a.Primary, aws.ToString(a.Type), aws.ToString(a.Value))

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
