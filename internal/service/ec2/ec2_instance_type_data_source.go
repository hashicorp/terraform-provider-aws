// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_ec2_instance_type", name="Instance Type")
func dataSourceInstanceType() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceInstanceTypeRead,

		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(20 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"auto_recovery_supported": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"bandwidth_weightings": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"bare_metal": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"boot_modes": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"burstable_performance_supported": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"current_generation": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"dedicated_hosts_supported": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"default_cores": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"default_network_card_index": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"default_threads_per_core": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"default_vcpus": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"ebs_encryption_support": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ebs_nvme_support": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ebs_optimized_support": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ebs_performance_baseline_bandwidth": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"ebs_performance_baseline_throughput": {
				Type:     schema.TypeFloat,
				Computed: true,
			},
			"ebs_performance_baseline_iops": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"ebs_performance_maximum_bandwidth": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"ebs_performance_maximum_throughput": {
				Type:     schema.TypeFloat,
				Computed: true,
			},
			"ebs_performance_maximum_iops": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"efa_supported": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"efa_maximum_interfaces": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"ena_srd_supported": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"ena_support": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"encryption_in_transit_supported": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"fpgas": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"count": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"manufacturer": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"memory_size": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						names.AttrName: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"free_tier_eligible": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"gpus": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"count": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"manufacturer": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"memory_size": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						names.AttrName: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"hibernation_supported": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"hypervisor": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"inference_accelerators": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"count": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"manufacturer": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"memory_size": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						names.AttrName: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"instance_disks": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"count": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						names.AttrSize: {
							Type:     schema.TypeInt,
							Computed: true,
						},
						names.AttrType: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"instance_storage_supported": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			names.AttrInstanceType: {
				Type:     schema.TypeString,
				Required: true,
			},
			"ipv6_supported": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"maximum_ipv4_addresses_per_interface": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"maximum_ipv6_addresses_per_interface": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"maximum_network_cards": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"maximum_network_interfaces": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"media_accelerators": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"count": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"manufacturer": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"memory_size": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						names.AttrName: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"memory_size": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"network_cards": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"baseline_bandwidth": {
							Type:     schema.TypeFloat,
							Computed: true,
						},
						"index": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"maximum_interfaces": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"performance": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"peak_bandwidth": {
							Type:     schema.TypeFloat,
							Computed: true,
						},
					},
				},
			},
			"network_performance": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"neuron_devices": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"core_count": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"core_version": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"count": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"memory_size": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						names.AttrName: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"nitro_enclaves_support": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"nitro_tpm_support": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"nitro_tpm_supported_versions": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"phc_support": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"supported_architectures": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"supported_cpu_features": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"supported_placement_strategies": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"supported_root_device_types": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"supported_usages_classes": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"supported_virtualization_types": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"sustained_clock_speed": {
				Type:     schema.TypeFloat,
				Computed: true,
			},
			"total_fpga_memory": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"total_gpu_memory": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"total_inference_memory": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"total_instance_storage": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"total_media_memory": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"total_neuron_device_memory": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"valid_cores": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeInt},
			},
			"valid_threads_per_core": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeInt},
			},
		},
	}
}

func dataSourceInstanceTypeRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	name := d.Get(names.AttrInstanceType).(string)
	v, err := findInstanceTypeByName(ctx, conn, name)

	if err == nil && (v.EbsInfo == nil || v.NetworkInfo == nil || v.PlacementGroupInfo == nil || v.ProcessorInfo == nil || v.VCpuInfo == nil) {
		err = tfresource.NewEmptyResultError(name)
	}

	if err != nil {
		return sdkdiag.AppendFromErr(diags, tfresource.SingularDataSourceFindError("EC2 Instance Type", err))
	}

	d.SetId(string(v.InstanceType))
	d.Set("auto_recovery_supported", v.AutoRecoverySupported)
	d.Set("bare_metal", v.BareMetal)
	d.Set("bandwidth_weightings", v.NetworkInfo.BandwidthWeightings)
	d.Set("boot_modes", v.SupportedBootModes)
	d.Set("burstable_performance_supported", v.BurstablePerformanceSupported)
	d.Set("current_generation", v.CurrentGeneration)
	d.Set("dedicated_hosts_supported", v.DedicatedHostsSupported)
	d.Set("default_cores", v.VCpuInfo.DefaultCores)
	d.Set("default_network_card_index", v.NetworkInfo.DefaultNetworkCardIndex)
	d.Set("default_threads_per_core", v.VCpuInfo.DefaultThreadsPerCore)
	d.Set("default_vcpus", v.VCpuInfo.DefaultVCpus)
	d.Set("ebs_encryption_support", v.EbsInfo.EncryptionSupport)
	d.Set("ebs_nvme_support", v.EbsInfo.NvmeSupport)
	d.Set("ebs_optimized_support", v.EbsInfo.EbsOptimizedSupport)
	if v := v.EbsInfo.EbsOptimizedInfo; v != nil {
		d.Set("ebs_performance_baseline_bandwidth", v.BaselineBandwidthInMbps)
		d.Set("ebs_performance_baseline_throughput", v.BaselineThroughputInMBps)
		d.Set("ebs_performance_baseline_iops", v.BaselineIops)
		d.Set("ebs_performance_maximum_bandwidth", v.MaximumBandwidthInMbps)
		d.Set("ebs_performance_maximum_throughput", v.MaximumThroughputInMBps)
		d.Set("ebs_performance_maximum_iops", v.MaximumIops)
	}
	d.Set("efa_supported", v.NetworkInfo.EfaSupported)
	if v := v.NetworkInfo.EfaInfo; v != nil {
		d.Set("efa_maximum_interfaces", v.MaximumEfaInterfaces)
	}
	d.Set("ena_srd_supported", v.NetworkInfo.EnaSrdSupported)
	d.Set("ena_support", v.NetworkInfo.EnaSupport)
	d.Set("encryption_in_transit_supported", v.NetworkInfo.EncryptionInTransitSupported)
	if v := v.FpgaInfo; v != nil {
		tfList := make([]any, len(v.Fpgas))
		for i, v := range v.Fpgas {
			tfMap := map[string]any{
				"count":        aws.ToInt32(v.Count),
				"manufacturer": aws.ToString(v.Manufacturer),
				"memory_size":  aws.ToInt32(v.MemoryInfo.SizeInMiB),
				names.AttrName: aws.ToString(v.Name),
			}
			tfList[i] = tfMap
		}
		d.Set("fpgas", tfList)
		d.Set("total_fpga_memory", v.TotalFpgaMemoryInMiB)
	}
	d.Set("free_tier_eligible", v.FreeTierEligible)
	if v := v.GpuInfo; v != nil {
		tfList := make([]any, len(v.Gpus))
		for i, v := range v.Gpus {
			tfMap := map[string]any{
				"count":        aws.ToInt32(v.Count),
				"manufacturer": aws.ToString(v.Manufacturer),
				"memory_size":  aws.ToInt32(v.MemoryInfo.SizeInMiB),
				names.AttrName: aws.ToString(v.Name),
			}
			tfList[i] = tfMap
		}
		d.Set("gpus", tfList)
		d.Set("total_gpu_memory", v.TotalGpuMemoryInMiB)
	}
	d.Set("hibernation_supported", v.HibernationSupported)
	d.Set("hypervisor", v.Hypervisor)
	if v := v.InferenceAcceleratorInfo; v != nil {
		tfList := make([]any, len(v.Accelerators))
		for i, v := range v.Accelerators {
			tfMap := map[string]any{
				"count":        aws.ToInt32(v.Count),
				"manufacturer": aws.ToString(v.Manufacturer),
				"memory_size":  aws.ToInt32(v.MemoryInfo.SizeInMiB),
				names.AttrName: aws.ToString(v.Name),
			}
			tfList[i] = tfMap
		}
		d.Set("inference_accelerators", tfList)
		d.Set("total_inference_memory", v.TotalInferenceMemoryInMiB)
	}
	if v := v.InstanceStorageInfo; v != nil {
		if v := v.Disks; v != nil {
			tfList := make([]any, len(v))
			for i, v := range v {
				tfMap := map[string]any{
					"count":        aws.ToInt32(v.Count),
					names.AttrSize: aws.ToInt64(v.SizeInGB),
					names.AttrType: v.Type,
				}
				tfList[i] = tfMap
			}
			d.Set("instance_disks", tfList)
		}
		d.Set("total_instance_storage", v.TotalSizeInGB)
	}
	d.Set("instance_storage_supported", v.InstanceStorageSupported)
	d.Set(names.AttrInstanceType, v.InstanceType)
	d.Set("ipv6_supported", v.NetworkInfo.Ipv6Supported)
	d.Set("maximum_ipv4_addresses_per_interface", v.NetworkInfo.Ipv4AddressesPerInterface)
	d.Set("maximum_ipv6_addresses_per_interface", v.NetworkInfo.Ipv6AddressesPerInterface)
	d.Set("maximum_network_cards", v.NetworkInfo.MaximumNetworkCards)
	d.Set("maximum_network_interfaces", v.NetworkInfo.MaximumNetworkInterfaces)
	if v := v.MediaAcceleratorInfo; v != nil {
		tfList := make([]any, len(v.Accelerators))
		for i, v := range v.Accelerators {
			tfMap := map[string]any{
				"count":        aws.ToInt32(v.Count),
				"manufacturer": aws.ToString(v.Manufacturer),
				"memory_size":  aws.ToInt32(v.MemoryInfo.SizeInMiB),
				names.AttrName: aws.ToString(v.Name),
			}
			tfList[i] = tfMap
		}
		d.Set("media_accelerators", tfList)
		d.Set("total_media_memory", v.TotalMediaMemoryInMiB)
	}
	d.Set("memory_size", v.MemoryInfo.SizeInMiB)
	if v := v.NeuronInfo; v != nil {
		tfList := make([]any, len(v.NeuronDevices))
		for i, v := range v.NeuronDevices {
			tfMap := map[string]any{
				"count":        aws.ToInt32(v.Count),
				"core_count":   aws.ToInt32(v.CoreInfo.Count),
				"core_version": aws.ToInt32(v.CoreInfo.Version),
				"memory_size":  aws.ToInt32(v.MemoryInfo.SizeInMiB),
				names.AttrName: aws.ToString(v.Name),
			}
			tfList[i] = tfMap
		}
		d.Set("neuron_devices", tfList)
		d.Set("total_neuron_device_memory", v.TotalNeuronDeviceMemoryInMiB)
	}
	d.Set("nitro_enclaves_support", v.NitroEnclavesSupport)
	d.Set("nitro_tpm_support", v.NitroTpmSupport)
	var nitroTPMSupportedVersions []string
	if v.NitroTpmInfo != nil {
		nitroTPMSupportedVersions = v.NitroTpmInfo.SupportedVersions
	}
	d.Set("nitro_tpm_supported_versions", nitroTPMSupportedVersions)
	d.Set("network_performance", v.NetworkInfo.NetworkPerformance)
	if v := v.NetworkInfo; v != nil {
		tfList := make([]any, len(v.NetworkCards))
		for i, v := range v.NetworkCards {
			tfMap := map[string]any{
				"baseline_bandwidth": aws.ToFloat64(v.BaselineBandwidthInGbps),
				"index":              aws.ToInt32(v.NetworkCardIndex),
				"maximum_interfaces": aws.ToInt32(v.MaximumNetworkInterfaces),
				"peak_bandwidth":     aws.ToFloat64(v.PeakBandwidthInGbps),
				"performance":        aws.ToString(v.NetworkPerformance),
			}
			tfList[i] = tfMap
		}
		d.Set("network_cards", tfList)
	}
	d.Set("phc_support", v.PhcSupport)
	d.Set("supported_architectures", v.ProcessorInfo.SupportedArchitectures)
	d.Set("supported_cpu_features", v.ProcessorInfo.SupportedFeatures)
	d.Set("supported_placement_strategies", v.PlacementGroupInfo.SupportedStrategies)
	d.Set("supported_root_device_types", v.SupportedRootDeviceTypes)
	d.Set("supported_usages_classes", v.SupportedUsageClasses)
	d.Set("supported_virtualization_types", v.SupportedVirtualizationTypes)
	d.Set("sustained_clock_speed", v.ProcessorInfo.SustainedClockSpeedInGhz)
	d.Set("valid_cores", v.VCpuInfo.ValidCores)
	d.Set("valid_threads_per_core", v.VCpuInfo.ValidThreadsPerCore)

	return diags
}
