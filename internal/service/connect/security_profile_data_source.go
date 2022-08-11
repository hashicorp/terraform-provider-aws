package connect

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/connect"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

func DataSourceSecurityProfile() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceSecurityProfileRead,
		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"instance_id": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 100),
			},
			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{"name", "security_profile_id"},
			},
			"organization_resource_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"permissions": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"security_profile_id": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{"security_profile_id", "name"},
			},
			"tags": tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourceSecurityProfileRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ConnectConn
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	instanceID := d.Get("instance_id").(string)

	input := &connect.DescribeSecurityProfileInput{
		InstanceId: aws.String(instanceID),
	}

	if v, ok := d.GetOk("security_profile_id"); ok {
		input.SecurityProfileId = aws.String(v.(string))
	} else if v, ok := d.GetOk("name"); ok {
		name := v.(string)
		securityProfileSummary, err := dataSourceGetSecurityProfileSummaryByName(ctx, conn, instanceID, name)

		if err != nil {
			return diag.FromErr(fmt.Errorf("error finding Connect Security Profile Summary by name (%s): %w", name, err))
		}

		if securityProfileSummary == nil {
			return diag.FromErr(fmt.Errorf("error finding Connect Security Profile Summary by name (%s): not found", name))
		}

		input.SecurityProfileId = securityProfileSummary.Id
	}

	resp, err := conn.DescribeSecurityProfileWithContext(ctx, input)

	if err != nil {
		return diag.FromErr(fmt.Errorf("error getting Connect Security Profile: %w", err))
	}

	if resp == nil || resp.SecurityProfile == nil {
		return diag.FromErr(fmt.Errorf("error getting Connect Security Profile: empty response"))
	}

	securityProfile := resp.SecurityProfile

	d.Set("arn", resp.SecurityProfile.Arn)
	d.Set("description", resp.SecurityProfile.Description)
	d.Set("instance_id", instanceID)
	d.Set("organization_resource_id", resp.SecurityProfile.OrganizationResourceId)
	d.Set("security_profile_id", resp.SecurityProfile.Id)
	d.Set("name", resp.SecurityProfile.SecurityProfileName)

	// reading permissions requires a separate API call
	permissions, err := getSecurityProfilePermissions(ctx, conn, instanceID, *resp.SecurityProfile.Id)

	if err != nil {
		return diag.FromErr(fmt.Errorf("error finding Connect Security Profile Permissions for Security Profile (%s): %w", *resp.SecurityProfile.Id, err))
	}

	if permissions != nil {
		d.Set("permissions", flex.FlattenStringSet(permissions))
	}

	if err := d.Set("tags", KeyValueTags(securityProfile.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return diag.FromErr(fmt.Errorf("error setting tags: %s", err))
	}

	d.SetId(fmt.Sprintf("%s:%s", instanceID, aws.StringValue(resp.SecurityProfile.Id)))

	return nil
}

func dataSourceGetSecurityProfileSummaryByName(ctx context.Context, conn *connect.Connect, instanceID, name string) (*connect.SecurityProfileSummary, error) {
	var result *connect.SecurityProfileSummary

	input := &connect.ListSecurityProfilesInput{
		InstanceId: aws.String(instanceID),
		MaxResults: aws.Int64(ListSecurityProfilesMaxResults),
	}

	err := conn.ListSecurityProfilesPagesWithContext(ctx, input, func(page *connect.ListSecurityProfilesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, qs := range page.SecurityProfileSummaryList {
			if qs == nil {
				continue
			}

			if aws.StringValue(qs.Name) == name {
				result = qs
				return false
			}
		}

		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}
