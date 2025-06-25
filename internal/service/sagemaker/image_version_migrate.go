// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sagemaker

import (
	"context"
	"fmt"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/sagemaker/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func resourceImageVersionV0() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"base_image": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"container_image": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"horovod": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"image_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"image_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"job_type": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: enum.Validate[awstypes.JobType](),
			},
			"ml_framework": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringMatch(regexache.MustCompile(`^[a-zA-Z]+ ?\d+\.\d+(\.\d+)?$`), ""),
			},
			"processor": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: enum.Validate[awstypes.Processor](),
			},
			"programming_lang": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringMatch(regexache.MustCompile(`^[a-zA-Z]+ ?\d+\.\d+(\.\d+)?$`), ""),
			},
			"release_notes": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 255),
			},
			"vendor_guidance": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: enum.Validate[awstypes.VendorGuidance](),
			},
			names.AttrVersion: {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func imageVersionStateUpgradeV0(_ context.Context, rawState map[string]any, meta any) (map[string]any, error) {
	if rawState == nil {
		rawState = map[string]any{}
	}

	// The version attribute must be asserted into an int, otherwise fmt will assume
	// a float64 when a numerical any value is detected.
	version, ok := rawState[names.AttrVersion].(int)
	if !ok {
		version = 1
	}

	rawState[names.AttrID] = fmt.Sprintf("%s,%d", rawState["image_name"], version)

	return rawState, nil
}
