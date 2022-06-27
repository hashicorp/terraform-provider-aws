package connect

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/connect"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

func ResourceSecurityProfile() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceSecurityProfileCreate,
		ReadContext:   resourceSecurityProfileRead,
		UpdateContext: resourceSecurityProfileUpdate,
		DeleteContext: resourceSecurityProfileDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 250),
			},
			"instance_id": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 100),
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"organization_resource_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"permissions": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 500,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringLenBetween(1, 128),
				},
			},
			"security_profile_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},
	}
}

func resourceSecurityProfileCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ConnectConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	instanceID := d.Get("instance_id").(string)
	securityProfileName := d.Get("name").(string)

	input := &connect.CreateSecurityProfileInput{
		InstanceId:          aws.String(instanceID),
		SecurityProfileName: aws.String(securityProfileName),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("permissions"); ok && v.(*schema.Set).Len() > 0 {
		input.Permissions = flex.ExpandStringSet(v.(*schema.Set))
	}

	if len(tags) > 0 {
		input.Tags = Tags(tags.IgnoreAWS())
	}

	log.Printf("[DEBUG] Creating Connect Security Profile %s", input)
	output, err := conn.CreateSecurityProfileWithContext(ctx, input)

	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating Connect Security Profile (%s): %w", securityProfileName, err))
	}

	if output == nil {
		return diag.FromErr(fmt.Errorf("error creating Connect Security Profile (%s): empty output", securityProfileName))
	}

	d.SetId(fmt.Sprintf("%s:%s", instanceID, aws.StringValue(output.SecurityProfileId)))

	return resourceSecurityProfileRead(ctx, d, meta)
}

func resourceSecurityProfileRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ConnectConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	instanceID, securityProfileID, err := SecurityProfileParseID(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	resp, err := conn.DescribeSecurityProfileWithContext(ctx, &connect.DescribeSecurityProfileInput{
		InstanceId:        aws.String(instanceID),
		SecurityProfileId: aws.String(securityProfileID),
	})

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, connect.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] Connect Security Profile (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("error getting Connect Security Profile (%s): %w", d.Id(), err))
	}

	if resp == nil || resp.SecurityProfile == nil {
		return diag.FromErr(fmt.Errorf("error getting Connect Security Profile (%s): empty response", d.Id()))
	}

	d.Set("arn", resp.SecurityProfile.Arn)
	d.Set("description", resp.SecurityProfile.Description)
	d.Set("instance_id", instanceID)
	d.Set("organization_resource_id", resp.SecurityProfile.OrganizationResourceId)
	d.Set("security_profile_id", resp.SecurityProfile.Id)
	d.Set("name", resp.SecurityProfile.SecurityProfileName)

	// reading permissions requires a separate API call
	permissions, err := getSecurityProfilePermissions(ctx, conn, instanceID, securityProfileID)

	if err != nil {
		return diag.FromErr(fmt.Errorf("error finding Connect Security Profile Permissions for Security Profile (%s): %w", securityProfileID, err))
	}

	if permissions != nil {
		d.Set("permissions", flex.FlattenStringSet(permissions))
	}

	tags := KeyValueTags(resp.SecurityProfile.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return diag.FromErr(fmt.Errorf("error setting tags: %w", err))
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return diag.FromErr(fmt.Errorf("error setting tags_all: %w", err))
	}

	return nil
}

func resourceSecurityProfileUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ConnectConn

	instanceID, securityProfileID, err := SecurityProfileParseID(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	input := &connect.UpdateSecurityProfileInput{
		InstanceId:        aws.String(instanceID),
		SecurityProfileId: aws.String(securityProfileID),
	}

	if d.HasChange("description") {
		input.Description = aws.String(d.Get("description").(string))
	}

	if d.HasChange("permissions") {
		input.Permissions = flex.ExpandStringSet(d.Get("permissions").(*schema.Set))
	}

	_, err = conn.UpdateSecurityProfileWithContext(ctx, input)

	if err != nil {
		return diag.FromErr(fmt.Errorf("[ERROR] Error updating SecurityProfile (%s): %w", d.Id(), err))
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")
		if err := UpdateTags(conn, d.Id(), o, n); err != nil {
			return diag.FromErr(fmt.Errorf("error updating tags: %w", err))
		}
	}

	return resourceSecurityProfileRead(ctx, d, meta)
}

func resourceSecurityProfileDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ConnectConn

	instanceID, securityProfileID, err := SecurityProfileParseID(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	_, err = conn.DeleteSecurityProfileWithContext(ctx, &connect.DeleteSecurityProfileInput{
		InstanceId:        aws.String(instanceID),
		SecurityProfileId: aws.String(securityProfileID),
	})

	if err != nil {
		return diag.FromErr(fmt.Errorf("error deleting SecurityProfile (%s): %w", d.Id(), err))
	}

	return nil
}

func SecurityProfileParseID(id string) (string, string, error) {
	parts := strings.SplitN(id, ":", 2)

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format of ID (%s), expected instanceID:securityProfileID", id)
	}

	return parts[0], parts[1], nil
}

func getSecurityProfilePermissions(ctx context.Context, conn *connect.Connect, instanceID, securityProfileID string) ([]*string, error) {
	var result []*string

	input := &connect.ListSecurityProfilePermissionsInput{
		InstanceId:        aws.String(instanceID),
		MaxResults:        aws.Int64(ListSecurityProfilePermissionsMaxResults),
		SecurityProfileId: aws.String(securityProfileID),
	}

	err := conn.ListSecurityProfilePermissionsPagesWithContext(ctx, input, func(page *connect.ListSecurityProfilePermissionsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		result = append(result, page.Permissions...)

		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}
