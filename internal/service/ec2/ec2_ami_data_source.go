package ec2

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"regexp"
	"sort"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

func DataSourceAMI() *schema.Resource {
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
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"block_device_mappings": {
				Type:     schema.TypeSet,
				Computed: true,
				Set:      amiBlockDeviceMappingHash,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"device_name": {
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
						"virtual_name": {
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
			"creation_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"deprecation_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
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
			"filter": DataSourceFiltersSchema(),
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
			"most_recent": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name_regex": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringIsValidRegExp,
			},
			"owner_id": {
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
				Set:      amiProductCodesHash,
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
			"state": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"state_reason": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"tags": tftags.TagsSchemaComputed(),
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
	conn := meta.(*conns.AWSClient).EC2Conn()
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	input := &ec2.DescribeImagesInput{
		IncludeDeprecated: aws.Bool(d.Get("include_deprecated").(bool)),
	}

	if v, ok := d.GetOk("executable_users"); ok {
		input.ExecutableUsers = flex.ExpandStringList(v.([]interface{}))
	}

	if v, ok := d.GetOk("filter"); ok {
		input.Filters = BuildFiltersDataSource(v.(*schema.Set))
	}

	if v, ok := d.GetOk("owners"); ok && len(v.([]interface{})) > 0 {
		input.Owners = flex.ExpandStringList(v.([]interface{}))
	}

	images, err := FindImages(ctx, conn, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 AMIs: %s", err)
	}

	var filteredImages []*ec2.Image
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

	if len(filteredImages) < 1 {
		return sdkdiag.AppendErrorf(diags, "Your query returned no results. Please change your search criteria and try again.")
	}

	if len(filteredImages) > 1 {
		if !d.Get("most_recent").(bool) {
			return sdkdiag.AppendErrorf(diags, "Your query returned more than one result. Please try a more "+
				"specific search criteria, or set `most_recent` attribute to true.")
		}
		sort.Slice(filteredImages, func(i, j int) bool {
			itime, _ := time.Parse(time.RFC3339, aws.StringValue(filteredImages[i].CreationDate))
			jtime, _ := time.Parse(time.RFC3339, aws.StringValue(filteredImages[j].CreationDate))
			return itime.Unix() > jtime.Unix()
		})
	}

	image := filteredImages[0]

	d.SetId(aws.StringValue(image.ImageId))
	d.Set("architecture", image.Architecture)
	imageArn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Region:    meta.(*conns.AWSClient).Region,
		Service:   ec2.ServiceName,
		Resource:  fmt.Sprintf("image/%s", d.Id()),
	}.String()
	d.Set("arn", imageArn)
	if err := d.Set("block_device_mappings", flattenAMIBlockDeviceMappings(image.BlockDeviceMappings)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting block_device_mappings: %s", err)
	}
	d.Set("boot_mode", image.BootMode)
	d.Set("creation_date", image.CreationDate)
	d.Set("deprecation_time", image.DeprecationTime)
	d.Set("description", image.Description)
	d.Set("ena_support", image.EnaSupport)
	d.Set("hypervisor", image.Hypervisor)
	d.Set("image_id", image.ImageId)
	d.Set("image_location", image.ImageLocation)
	d.Set("image_owner_alias", image.ImageOwnerAlias)
	d.Set("image_type", image.ImageType)
	d.Set("imds_support", image.ImdsSupport)
	d.Set("kernel_id", image.KernelId)
	d.Set("name", image.Name)
	d.Set("owner_id", image.OwnerId)
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
	d.Set("state", image.State)
	if err := d.Set("state_reason", flattenAMIStateReason(image.StateReason)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting state_reason: %s", err)
	}
	d.Set("tpm_support", image.TpmSupport)
	d.Set("usage_operation", image.UsageOperation)
	d.Set("virtualization_type", image.VirtualizationType)

	if err := d.Set("tags", KeyValueTags(image.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	return diags
}

func flattenAMIBlockDeviceMappings(m []*ec2.BlockDeviceMapping) *schema.Set {
	s := &schema.Set{
		F: amiBlockDeviceMappingHash,
	}
	for _, v := range m {
		mapping := map[string]interface{}{
			"device_name":  aws.StringValue(v.DeviceName),
			"virtual_name": aws.StringValue(v.VirtualName),
		}

		if v.Ebs != nil {
			ebs := map[string]interface{}{
				"delete_on_termination": fmt.Sprintf("%t", aws.BoolValue(v.Ebs.DeleteOnTermination)),
				"encrypted":             fmt.Sprintf("%t", aws.BoolValue(v.Ebs.Encrypted)),
				"iops":                  fmt.Sprintf("%d", aws.Int64Value(v.Ebs.Iops)),
				"throughput":            fmt.Sprintf("%d", aws.Int64Value(v.Ebs.Throughput)),
				"volume_size":           fmt.Sprintf("%d", aws.Int64Value(v.Ebs.VolumeSize)),
				"snapshot_id":           aws.StringValue(v.Ebs.SnapshotId),
				"volume_type":           aws.StringValue(v.Ebs.VolumeType),
			}

			mapping["ebs"] = ebs
		}

		log.Printf("[DEBUG] aws_ami - adding block device mapping: %v", mapping)
		s.Add(mapping)
	}
	return s
}

func flattenAMIProductCodes(m []*ec2.ProductCode) *schema.Set {
	s := &schema.Set{
		F: amiProductCodesHash,
	}
	for _, v := range m {
		code := map[string]interface{}{
			"product_code_id":   aws.StringValue(v.ProductCodeId),
			"product_code_type": aws.StringValue(v.ProductCodeType),
		}
		s.Add(code)
	}
	return s
}

func amiRootSnapshotId(image *ec2.Image) string {
	if image.RootDeviceName == nil {
		return ""
	}
	for _, bdm := range image.BlockDeviceMappings {
		if bdm.DeviceName == nil ||
			aws.StringValue(bdm.DeviceName) != aws.StringValue(image.RootDeviceName) {
			continue
		}
		if bdm.Ebs != nil && bdm.Ebs.SnapshotId != nil {
			return aws.StringValue(bdm.Ebs.SnapshotId)
		}
	}
	return ""
}

func flattenAMIStateReason(m *ec2.StateReason) map[string]interface{} {
	s := make(map[string]interface{})
	if m != nil {
		s["code"] = aws.StringValue(m.Code)
		s["message"] = aws.StringValue(m.Message)
	} else {
		s["code"] = "UNSET"
		s["message"] = "UNSET"
	}
	return s
}

func amiBlockDeviceMappingHash(v interface{}) int {
	var buf bytes.Buffer
	// All keys added in alphabetical order.
	m := v.(map[string]interface{})
	buf.WriteString(fmt.Sprintf("%s-", m["device_name"].(string)))
	if d, ok := m["ebs"]; ok {
		if len(d.(map[string]interface{})) > 0 {
			e := d.(map[string]interface{})
			buf.WriteString(fmt.Sprintf("%s-", e["delete_on_termination"].(string)))
			buf.WriteString(fmt.Sprintf("%s-", e["encrypted"].(string)))
			buf.WriteString(fmt.Sprintf("%s-", e["iops"].(string)))
			buf.WriteString(fmt.Sprintf("%s-", e["volume_size"].(string)))
			buf.WriteString(fmt.Sprintf("%s-", e["volume_type"].(string)))
		}
	}
	if d, ok := m["no_device"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", d.(string)))
	}
	if d, ok := m["virtual_name"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", d.(string)))
	}
	if d, ok := m["snapshot_id"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", d.(string)))
	}
	return create.StringHashcode(buf.String())
}

func amiProductCodesHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	// All keys added in alphabetical order.
	buf.WriteString(fmt.Sprintf("%s-", m["product_code_id"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["product_code_type"].(string)))
	return create.StringHashcode(buf.String())
}
