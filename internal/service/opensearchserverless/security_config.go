package opensearchserverless

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/opensearchserverless"
	"github.com/aws/aws-sdk-go-v2/service/opensearchserverless/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func ResourceSecurityConfig() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceSecurityConfigCreate,
		ReadWithoutTimeout:   resourceSecurityConfigRead,
		UpdateWithoutTimeout: resourceSecurityConfigUpdate,
		DeleteWithoutTimeout: resourceSecurityConfigDelete,

		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				const idSeparator = "/"
				parts := strings.Split(d.Id(), idSeparator)
				if len(parts) != 3 || parts[0] == "" || parts[1] == "" || parts[2] == "" {
					return nil, fmt.Errorf("unexpected format for ID (%[1]s), expected saml/account-id/name", d.Id())
				}

				d.SetId(d.Id())
				d.Set("name", parts[2])

				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(3, 32),
				ForceNew:     true,
			},
			"config_version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 1000),
			},
			"saml_options": {
				Type:     schema.TypeList,
				Required: true, // API docs suggest this is optional, but it returns an error if not provided
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"group_attribute": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(1, 2048),
						},
						"metadata": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(1, 20480),
						},
						"session_timeout": {
							Type:         schema.TypeInt,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validation.IntBetween(5, 1440),
						},
						"user_attribute": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(1, 2048),
						},
					},
				},
			},
			"type": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: enum.Validate[types.SecurityConfigType]()},
		},
	}
}

const (
	ResNameSecurityConfig = "Security Config"
)

func resourceSecurityConfigCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).OpenSearchServerlessClient()

	in := &opensearchserverless.CreateSecurityConfigInput{
		ClientToken: aws.String(resource.UniqueId()),
		Name:        aws.String(d.Get("name").(string)),
		Type:        types.SecurityConfigType(d.Get("type").(string)),
	}

	if v, ok := d.GetOk("description"); ok {
		in.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("saml_options"); ok {
		in.SamlOptions = expandSAMLOptions(v.([]interface{}))
	}

	out, err := conn.CreateSecurityConfig(ctx, in)
	if err != nil {
		return create.DiagError(names.OpenSearchServerless, create.ErrActionCreating, ResNameSecurityConfig, d.Get("name").(string), err)
	}

	if out == nil || out.SecurityConfigDetail == nil {
		return create.DiagError(names.OpenSearchServerless, create.ErrActionCreating, ResNameSecurityConfig, d.Get("name").(string), errors.New("empty output"))
	}

	d.SetId(aws.ToString(out.SecurityConfigDetail.Id))

	return resourceSecurityConfigRead(ctx, d, meta)
}

func resourceSecurityConfigRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).OpenSearchServerlessClient()
	out, err := findSecurityConfigByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] OpenSearchServerless Security Config (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return create.DiagError(names.OpenSearchServerless, create.ErrActionReading, ResNameSecurityConfig, d.Id(), err)
	}

	d.Set("description", out.Description)
	d.Set("type", "saml") // GetSecurityConfig doesn't return this field so hard-code it for now
	d.Set("saml_options", flattenSAMLOptions(out.SamlOptions))
	return nil
}

func resourceSecurityConfigUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).OpenSearchServerlessClient()

	update := false

	in := &opensearchserverless.UpdateSecurityConfigInput{
		ClientToken:   aws.String(resource.UniqueId()),
		Id:            aws.String(d.Id()),
		ConfigVersion: aws.String(d.Get("config_version").(string)),
	}

	if d.HasChanges("description") {
		in.Description = aws.String(d.Get("description").(string))
		update = true
	}

	if d.HasChanges("saml_options") {
		in.SamlOptions = expandSAMLOptions(d.Get("saml_options").([]interface{}))
		update = true
	}

	if !update {
		return nil
	}

	log.Printf("[DEBUG] Updating OpenSearchServerless Security Config (%s): %#v", d.Id(), in)
	_, err := conn.UpdateSecurityConfig(ctx, in)
	if err != nil {
		return create.DiagError(names.OpenSearchServerless, create.ErrActionUpdating, ResNameSecurityConfig, d.Id(), err)
	}

	return resourceSecurityConfigRead(ctx, d, meta)
}

func resourceSecurityConfigDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).OpenSearchServerlessClient()

	log.Printf("[INFO] Deleting OpenSearchServerless Security Config %s", d.Id())

	_, err := conn.DeleteSecurityConfig(ctx, &opensearchserverless.DeleteSecurityConfigInput{
		ClientToken: aws.String(resource.UniqueId()),
		Id:          aws.String(d.Id()),
	})

	if err != nil {
		var nfe *types.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return nil
		}

		return create.DiagError(names.OpenSearchServerless, create.ErrActionDeleting, ResNameSecurityConfig, d.Id(), err)
	}

	return nil
}
