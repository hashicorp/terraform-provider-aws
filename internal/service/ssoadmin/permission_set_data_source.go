package ssoadmin

import (
	"context"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssoadmin"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func DataSourcePermissionSet() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourcePermissionSetRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: verify.ValidARN,
				ExactlyOneOf: []string{"arn", "name"},
			},

			"created_date": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"instance_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},

			"name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 32),
					validation.StringMatch(regexp.MustCompile(`[\w+=,.@-]+`), "must match [\\w+=,.@-]"),
				),
				ExactlyOneOf: []string{"name", "arn"},
			},

			"relay_state": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"session_duration": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"tags": tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourcePermissionSetRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SSOAdminConn()
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	instanceArn := d.Get("instance_arn").(string)

	var permissionSet *ssoadmin.PermissionSet

	if v, ok := d.GetOk("arn"); ok {
		arn := v.(string)

		input := &ssoadmin.DescribePermissionSetInput{
			InstanceArn:      aws.String(instanceArn),
			PermissionSetArn: aws.String(arn),
		}

		output, err := conn.DescribePermissionSetWithContext(ctx, input)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading SSO Admin Permission Set (%s): %s", arn, err)
		}

		if output == nil {
			return sdkdiag.AppendErrorf(diags, "reading SSO Admin Permission Set (%s): empty output", arn)
		}

		permissionSet = output.PermissionSet
	} else if v, ok := d.GetOk("name"); ok {
		name := v.(string)
		var describeErr error

		input := &ssoadmin.ListPermissionSetsInput{
			InstanceArn: aws.String(instanceArn),
		}

		err := conn.ListPermissionSetsPagesWithContext(ctx, input, func(page *ssoadmin.ListPermissionSetsOutput, lastPage bool) bool {
			if page == nil {
				return !lastPage
			}

			for _, permissionSetArn := range page.PermissionSets {
				if permissionSetArn == nil {
					continue
				}

				output, describeErr := conn.DescribePermissionSetWithContext(ctx, &ssoadmin.DescribePermissionSetInput{
					InstanceArn:      aws.String(instanceArn),
					PermissionSetArn: permissionSetArn,
				})

				if describeErr != nil {
					return false
				}

				if output == nil || output.PermissionSet == nil {
					continue
				}

				if aws.StringValue(output.PermissionSet.Name) == name {
					permissionSet = output.PermissionSet
					return false
				}
			}

			return !lastPage
		})

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "listing SSO Permission Sets: %s", err)
		}

		if describeErr != nil {
			return sdkdiag.AppendErrorf(diags, "reading SSO Permission Set (%s): %s", name, describeErr)
		}
	}

	if permissionSet == nil {
		return sdkdiag.AppendErrorf(diags, "reading SSO Permission Set: not found")
	}

	arn := aws.StringValue(permissionSet.PermissionSetArn)

	d.SetId(arn)
	d.Set("arn", arn)
	d.Set("created_date", permissionSet.CreatedDate.Format(time.RFC3339))
	d.Set("description", permissionSet.Description)
	d.Set("instance_arn", instanceArn)
	d.Set("name", permissionSet.Name)
	d.Set("session_duration", permissionSet.SessionDuration)
	d.Set("relay_state", permissionSet.RelayState)

	tags, err := ListTags(ctx, conn, arn, instanceArn)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing tags for SSO Permission Set (%s): %s", arn, err)
	}

	if err := d.Set("tags", tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	return diags
}
