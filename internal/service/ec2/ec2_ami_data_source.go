// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_ami", name="AMI")
// @Tags
// @Testing(tagsTest=false)
func dataSourceAMI() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceAMIRead,

		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(20 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"architecture": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"block_device_mappings": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrDeviceName: {
							Type:     schema.TypeString,
							Computed: true,
						},
						"ebs": {
							Type:     schema.TypeMap,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"no_device": {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrVirtualName: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"boot_mode": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrCreationDate: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"deprecation_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ena_support": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"executable_users": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrFilter: customFiltersSchema(),
			"hypervisor": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"image_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"image_location": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"image_owner_alias": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"image_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"imds_support": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"include_deprecated": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"kernel_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrMostRecent: {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name_regex": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringIsValidRegExp,
			},
			names.AttrOwnerID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"owners": {
				Type:     schema.TypeList,
				Optional: true,
				MinItems: 1,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.NoZeroValues,
				},
			},
			"platform": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"platform_details": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"product_codes": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"product_code_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"product_code_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"public": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"ramdisk_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"root_device_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"root_device_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"root_snapshot_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"sriov_net_support": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrState: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"state_reason": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
			"tpm_support": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"usage_operation": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"virtualization_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceAMIRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	input := &ec2.DescribeImagesInput{
		IncludeDeprecated: aws.Bool(d.Get("include_deprecated").(bool)),
	}

	if v, ok := d.GetOk("executable_users"); ok {
		input.ExecutableUsers = flex.ExpandStringValueList(v.([]interface{}))
	}

	if v, ok := d.GetOk(names.AttrFilter); ok {
		input.Filters = newCustomFilterList(v.(*schema.Set))
	}

	if v, ok := d.GetOk("owners"); ok && len(v.([]interface{})) > 0 {
		input.Owners = flex.ExpandStringValueList(v.([]interface{}))
	}

	images, err := findImages(ctx, conn, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 AMIs: %s", err)
	}

	var filteredImages []awstypes.Image
	if v, ok := d.GetOk("name_regex"); ok {
		r := regexache.MustCompile(v.(string))
		for _, image := range images {
			name := aws.ToString(image.Name)

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

	if len(filteredImages) < 1 {
		return sdkdiag.AppendErrorf(diags, "Your query returned no results. Please change your search criteria and try again.")
	}

	if len(filteredImages) > 1 {
		if !d.Get(names.AttrMostRecent).(bool) {
			return sdkdiag.AppendErrorf(diags, "Your query returned more than one result. Please try a more "+
				"specific search criteria, or set `most_recent` attribute to true.")
		}
		sort.Slice(filteredImages, func(i, j int) bool {
			itime, _ := time.Parse(time.RFC3339, aws.ToString(filteredImages[i].CreationDate))
			jtime, _ := time.Parse(time.RFC3339, aws.ToString(filteredImages[j].CreationDate))
			return itime.Unix() > jtime.Unix()
		})
	}

	image := filteredImages[0]

	d.SetId(aws.ToString(image.ImageId))
	d.Set("architecture", image.Architecture)
	imageArn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Region:    meta.(*conns.AWSClient).Region,
		Service:   names.EC2,
		Resource:  fmt.Sprintf("image/%s", d.Id()),
	}.String()
	d.Set(names.AttrARN, imageArn)
	if err := d.Set("block_device_mappings", flattenAMIBlockDeviceMappings(image.BlockDeviceMappings)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting block_device_mappings: %s", err)
	}
	d.Set("boot_mode", image.BootMode)
	d.Set(names.AttrCreationDate, image.CreationDate)
	d.Set("deprecation_time", image.DeprecationTime)
	d.Set(names.AttrDescription, image.Description)
	d.Set("ena_support", image.EnaSupport)
	d.Set("hypervisor", image.Hypervisor)
	d.Set("image_id", image.ImageId)
	d.Set("image_location", image.ImageLocation)
	d.Set("image_owner_alias", image.ImageOwnerAlias)
	d.Set("image_type", image.ImageType)
	d.Set("imds_support", image.ImdsSupport)
	d.Set("kernel_id", image.KernelId)
	d.Set(names.AttrName, image.Name)
	d.Set(names.AttrOwnerID, image.OwnerId)
	d.Set("platform", image.Platform)
	d.Set("platform_details", image.PlatformDetails)
	if err := d.Set("product_codes", flattenAMIProductCodes(image.ProductCodes)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting product_codes: %s", err)
	}
	d.Set("public", image.Public)
	d.Set("ramdisk_id", image.RamdiskId)
	d.Set("root_device_name", image.RootDeviceName)
	d.Set("root_device_type", image.RootDeviceType)
	d.Set("root_snapshot_id", amiRootSnapshotId(image))
	d.Set("sriov_net_support", image.SriovNetSupport)
	d.Set(names.AttrState, image.State)
	if err := d.Set("state_reason", flattenAMIStateReason(image.StateReason)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting state_reason: %s", err)
	}
	d.Set("tpm_support", image.TpmSupport)
	d.Set("usage_operation", image.UsageOperation)
	d.Set("virtualization_type", image.VirtualizationType)

	setTagsOut(ctx, image.Tags)

	return diags
}

func flattenAMIBlockDeviceMappings(apiObjects []awstypes.BlockDeviceMapping) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfMap := map[string]interface{}{
			names.AttrDeviceName:  aws.ToString(apiObject.DeviceName),
			names.AttrVirtualName: aws.ToString(apiObject.VirtualName),
		}

		if apiObject := apiObject.Ebs; apiObject != nil {
			ebs := map[string]interface{}{
				names.AttrDeleteOnTermination: flex.BoolToStringValue(apiObject.DeleteOnTermination),
				names.AttrEncrypted:           flex.BoolToStringValue(apiObject.Encrypted),
				names.AttrIOPS:                flex.Int32ToStringValue(apiObject.Iops),
				names.AttrSnapshotID:          aws.ToString(apiObject.SnapshotId),
				names.AttrThroughput:          flex.Int32ToStringValue(apiObject.Throughput),
				names.AttrVolumeSize:          flex.Int32ToStringValue(apiObject.VolumeSize),
				names.AttrVolumeType:          apiObject.VolumeType,
			}

			tfMap["ebs"] = ebs
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenAMIProductCodes(apiObjects []awstypes.ProductCode) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfList = append(tfList, map[string]interface{}{
			"product_code_id":   aws.ToString(apiObject.ProductCodeId),
			"product_code_type": apiObject.ProductCodeType,
		})
	}

	return tfList
}

func amiRootSnapshotId(image awstypes.Image) string {
	if image.RootDeviceName == nil {
		return ""
	}
	for _, bdm := range image.BlockDeviceMappings {
		if bdm.DeviceName == nil ||
			aws.ToString(bdm.DeviceName) != aws.ToString(image.RootDeviceName) {
			continue
		}
		if bdm.Ebs != nil && bdm.Ebs.SnapshotId != nil {
			return aws.ToString(bdm.Ebs.SnapshotId)
		}
	}
	return ""
}

func flattenAMIStateReason(m *awstypes.StateReason) map[string]interface{} {
	s := make(map[string]interface{})
	if m != nil {
		s["code"] = aws.ToString(m.Code)
		s[names.AttrMessage] = aws.ToString(m.Message)
	} else {
		s["code"] = "UNSET"
		s[names.AttrMessage] = "UNSET"
	}
	return s
}
