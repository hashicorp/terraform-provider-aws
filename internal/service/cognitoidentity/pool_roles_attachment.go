package cognitoidentity

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cognitoidentity"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func ResourcePoolRolesAttachment() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourcePoolRolesAttachmentCreate,
		ReadWithoutTimeout:   resourcePoolRolesAttachmentRead,
		UpdateWithoutTimeout: resourcePoolRolesAttachmentUpdate,
		DeleteWithoutTimeout: resourcePoolRolesAttachmentDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"identity_pool_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"role_mapping": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"identity_provider": {
							Type:     schema.TypeString,
							Required: true,
						},
						"ambiguous_role_resolution": {
							Type:     schema.TypeString,
							Optional: true, // Required if Type equals Token or Rules.
							ValidateFunc: validation.StringInSlice([]string{
								cognitoidentity.AmbiguousRoleResolutionTypeAuthenticatedRole,
								cognitoidentity.AmbiguousRoleResolutionTypeDeny,
							}, false),
						},
						"mapping_rule": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 25,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"claim": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validRoleMappingsRulesClaim,
									},
									"match_type": {
										Type:     schema.TypeString,
										Required: true,
										ValidateFunc: validation.StringInSlice([]string{
											cognitoidentity.MappingRuleMatchTypeEquals,
											cognitoidentity.MappingRuleMatchTypeContains,
											cognitoidentity.MappingRuleMatchTypeStartsWith,
											cognitoidentity.MappingRuleMatchTypeNotEqual,
										}, false),
									},
									"role_arn": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: verify.ValidARN,
									},
									"value": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringLenBetween(1, 128),
									},
								},
							},
						},
						"type": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								cognitoidentity.RoleMappingTypeToken,
								cognitoidentity.RoleMappingTypeRules,
							}, false),
						},
					},
				},
			},

			"roles": {
				Type:     schema.TypeMap,
				Required: true,
				ForceNew: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: verify.ValidARN,
				},
			},
		},
	}
}

func resourcePoolRolesAttachmentCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CognitoIdentityConn()

	// Validates role keys to be either authenticated or unauthenticated,
	// since ValidateFunc validates only the value not the key.
	if errors := validRoles(d.Get("roles").(map[string]interface{})); len(errors) > 0 {
		return sdkdiag.AppendErrorf(diags, "Error validating Roles: %v", errors)
	}

	params := &cognitoidentity.SetIdentityPoolRolesInput{
		IdentityPoolId: aws.String(d.Get("identity_pool_id").(string)),
		Roles:          expandIdentityPoolRoles(d.Get("roles").(map[string]interface{})),
	}

	if v, ok := d.GetOk("role_mapping"); ok {
		errors := validateRoleMappings(v.(*schema.Set).List())

		if len(errors) > 0 {
			return sdkdiag.AppendErrorf(diags, "Error validating ambiguous role resolution: %v", errors)
		}

		params.RoleMappings = expandIdentityPoolRoleMappingsAttachment(v.(*schema.Set).List())
	}

	log.Printf("[DEBUG] Creating Cognito Identity Pool Roles Association: %#v", params)
	_, err := conn.SetIdentityPoolRolesWithContext(ctx, params)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "Error creating Cognito Identity Pool Roles Association: %s", err)
	}

	d.SetId(d.Get("identity_pool_id").(string))

	return append(diags, resourcePoolRolesAttachmentRead(ctx, d, meta)...)
}

func resourcePoolRolesAttachmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CognitoIdentityConn()
	log.Printf("[DEBUG] Reading Cognito Identity Pool Roles Association: %s", d.Id())

	ip, err := conn.GetIdentityPoolRolesWithContext(ctx, &cognitoidentity.GetIdentityPoolRolesInput{
		IdentityPoolId: aws.String(d.Id()),
	})
	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, cognitoidentity.ErrCodeResourceNotFoundException) {
		create.LogNotFoundRemoveState(names.CognitoIdentity, create.ErrActionReading, ResNamePoolRolesAttachment, d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return create.DiagError(names.CognitoIdentity, create.ErrActionReading, ResNamePoolRolesAttachment, d.Id(), err)
	}

	d.Set("identity_pool_id", ip.IdentityPoolId)

	if err := d.Set("roles", aws.StringValueMap(ip.Roles)); err != nil {
		return sdkdiag.AppendErrorf(diags, "Error setting roles error: %#v", err)
	}

	if err := d.Set("role_mapping", flattenIdentityPoolRoleMappingsAttachment(ip.RoleMappings)); err != nil {
		return sdkdiag.AppendErrorf(diags, "Error setting role mappings error: %#v", err)
	}

	return diags
}

func resourcePoolRolesAttachmentUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CognitoIdentityConn()

	// Validates role keys to be either authenticated or unauthenticated,
	// since ValidateFunc validates only the value not the key.
	if errors := validRoles(d.Get("roles").(map[string]interface{})); len(errors) > 0 {
		return sdkdiag.AppendErrorf(diags, "Error validating Roles: %v", errors)
	}

	params := &cognitoidentity.SetIdentityPoolRolesInput{
		IdentityPoolId: aws.String(d.Get("identity_pool_id").(string)),
		Roles:          expandIdentityPoolRoles(d.Get("roles").(map[string]interface{})),
	}

	if d.HasChange("role_mapping") {
		v, ok := d.GetOk("role_mapping")
		var mappings []interface{}

		if ok {
			errors := validateRoleMappings(v.(*schema.Set).List())

			if len(errors) > 0 {
				return sdkdiag.AppendErrorf(diags, "Error validating ambiguous role resolution: %v", errors)
			}
			mappings = v.(*schema.Set).List()
		} else {
			mappings = []interface{}{}
		}

		params.RoleMappings = expandIdentityPoolRoleMappingsAttachment(mappings)
	}

	log.Printf("[DEBUG] Updating Cognito Identity Pool Roles Association: %#v", params)
	_, err := conn.SetIdentityPoolRolesWithContext(ctx, params)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "Error updating Cognito Identity Pool Roles Association: %s", err)
	}

	d.SetId(d.Get("identity_pool_id").(string))

	return append(diags, resourcePoolRolesAttachmentRead(ctx, d, meta)...)
}

func resourcePoolRolesAttachmentDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CognitoIdentityConn()
	log.Printf("[DEBUG] Deleting Cognito Identity Pool Roles Association: %s", d.Id())

	_, err := conn.SetIdentityPoolRolesWithContext(ctx, &cognitoidentity.SetIdentityPoolRolesInput{
		IdentityPoolId: aws.String(d.Id()),
		Roles:          expandIdentityPoolRoles(make(map[string]interface{})),
		RoleMappings:   expandIdentityPoolRoleMappingsAttachment([]interface{}{}),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "Error deleting Cognito identity pool roles association: %s", err)
	}

	return diags
}

// Validating that each role_mapping ambiguous_role_resolution
// is defined when "type" equals Token or Rules.
func validateRoleMappings(roleMappings []interface{}) []error {
	errors := make([]error, 0)

	for _, r := range roleMappings {
		rm := r.(map[string]interface{})

		// If Type equals "Token" or "Rules", ambiguous_role_resolution must be defined.
		// This should be removed as soon as we can have a ValidateFuncAgainst callable on the schema.
		if err := validRoleMappingsAmbiguousRoleResolutionAgainstType(rm); len(err) > 0 {
			errors = append(errors, fmt.Errorf("Role Mapping %q: %v", rm["identity_provider"].(string), err))
		}

		// Validating that Rules Configuration is defined when Type equals Rules
		// but not defined when Type equals Token.
		if err := validRoleMappingsRulesConfiguration(rm); len(err) > 0 {
			errors = append(errors, fmt.Errorf("Role Mapping %q: %v", rm["identity_provider"].(string), err))
		}
	}

	return errors
}
