// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
)

// @SDKDataSource("aws_ami_ids")
func DataSourceAMIIDs() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceAMIIDsRead,

		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(20 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"executable_users": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"filter": DataSourceFiltersSchema(),
			"ids": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"include_deprecated": {
				Type:     schema.TypeBool,
				Default:  false,
				Optional: true,
			},
			"name_regex": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringIsValidRegExp,
			},
			"owners": {
				Type:     schema.TypeList,
				Required: true,
				MinItems: 1,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.NoZeroValues,
				},
			},
			"sort_ascending": {
				Type:     schema.TypeBool,
				Default:  false,
				Optional: true,
			},
		},
	}
}

func dataSourceAMIIDsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn(ctx)

	input := &ec2.DescribeImagesInput{
		IncludeDeprecated: aws.Bool(d.Get("include_deprecated").(bool)),
		Owners:            flex.ExpandStringList(d.Get("owners").([]interface{})),
	}

	if v, ok := d.GetOk("executable_users"); ok {
		input.ExecutableUsers = flex.ExpandStringList(v.([]interface{}))
	}

	if v, ok := d.GetOk("filter"); ok {
		input.Filters = BuildFiltersDataSource(v.(*schema.Set))
	}

	images, err := FindImages(ctx, conn, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 AMIs: %s", err)
	}

	var filteredImages []*ec2.Image
	imageIDs := make([]string, 0)

	if v, ok := d.GetOk("name_regex"); ok {
		r := regexp.MustCompile(v.(string))
		for _, image := range images {
			name := aws.StringValue(image.Name)

			// Check for a very rare case where the response would include no
			// image name. No name means nothing to attempt a match against,
			// therefore we are skipping such image.
			if name == "" {
				continue
			}

			if r.MatchString(name) {
				filteredImages = append(filteredImages, image)
			}
		}
	} else {
		filteredImages = images[:]
	}

	sort.Slice(filteredImages, func(i, j int) bool {
		itime, _ := time.Parse(time.RFC3339, aws.StringValue(filteredImages[i].CreationDate))
		jtime, _ := time.Parse(time.RFC3339, aws.StringValue(filteredImages[j].CreationDate))
		if d.Get("sort_ascending").(bool) {
			return itime.Unix() < jtime.Unix()
		}
		return itime.Unix() > jtime.Unix()
	})
	for _, image := range filteredImages {
		imageIDs = append(imageIDs, aws.StringValue(image.ImageId))
	}

	d.SetId(fmt.Sprintf("%d", create.StringHashcode(input.String())))
	d.Set("ids", imageIDs)

	return diags
}
