package aws

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"log"
)

func dataSourceAwsEc2InstanceType() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsEc2InstanceTypeRead,

		Schema: map[string]*schema.Schema{
			"instance_type": {
				Type:     schema.TypeString,
				Required: true,
			},

			"current_generation": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"free_tier_eligible": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"supported_usages_classes": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"supported_root_device_types": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"bare_metal": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"hypervisor": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
			"supported_architectures": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"sustained_clock_speed": {
				Type:     schema.TypeFloat,
				Computed: true,
			},
			"default_vcpus": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"default_cores": {
				Type:     schema.TypeInt,
				Computed: true,
				Optional: true,
			},
			"default_threads_per_core": {
				Type:     schema.TypeInt,
				Computed: true,
				Optional: true,
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
			"memory_size": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"instance_storage_supported": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"total_instance_storage": {
				Type:     schema.TypeInt,
				Computed: true,
				Optional: true,
			},
			"instance_disks": {
				Type:     schema.TypeSet,
				Computed: true,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"size": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"count": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"type": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"ebs_optimized_support": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ebs_encryption_support": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"network_performance": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"maximum_network_interfaces": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"maximum_ipv4_addresses_per_interface": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"maximum_ipv6_addresses_per_interface": {
				Type:     schema.TypeInt,
				Computed: true,
				Optional: true,
			},
			"ipv6_supported": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"ena_support": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"gpus": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"manufacturer": {
							Type:     schema.TypeString,
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
					},
				},
			},
			"total_gpu_memory": {
				Type:     schema.TypeInt,
				Computed: true,
				Optional: true,
			},
			"fpgas": {
				Type:     schema.TypeSet,
				Computed: true,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"manufacturer": {
							Type:     schema.TypeString,
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
					},
				},
			},
			"total_fpga_memory": {
				Type:     schema.TypeInt,
				Computed: true,
				Optional: true,
			},
			"supported_placement_strategies": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"accelerators": {
				Type:     schema.TypeSet,
				Computed: true,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"manufacturer": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"count": {
							Type:     schema.TypeInt,
							Computed: true,
						},
					},
				},
			},
			"hibernation_supported": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"burstable_performance_supported": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"dedicated_hosts_supported": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"auto_recovery_supported": {
				Type:     schema.TypeBool,
				Computed: true,
			},
		},
	}
}

func dataSourceAwsEc2InstanceTypeRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	params := &ec2.DescribeInstanceTypesInput{}

	instanceType := d.Get("instance_type").(string)
	params.InstanceTypes = []*string{aws.String(instanceType)}
	log.Printf("[DEBUG] Reading instances types: %s", params)

	resp, err := conn.DescribeInstanceTypes(params)
	if err != nil {
		return err
	}
	if len(resp.InstanceTypes) == 0 {
		return fmt.Errorf("no Instance Type found for %s", instanceType)
	}
	if len(resp.InstanceTypes) > 1 {
		return fmt.Errorf("multiple instance types found for type %s", instanceType)
	}
	v := resp.InstanceTypes[0]
	d.Set("instance_type", v.InstanceType)
	d.Set("current_generation", v.CurrentGeneration)
	if err := d.Set("free_tier_eligible", aws.BoolValue(v.FreeTierEligible)); err != nil {
		return fmt.Errorf("error setting free_tier_eligible: %s", err)
	}
	if err := d.Set("supported_usages_classes", aws.StringValueSlice(v.SupportedUsageClasses)); err != nil {
		return fmt.Errorf("error setting supported_usages_classes: %s", err)
	}
	if err := d.Set("supported_root_device_types", aws.StringValueSlice(v.SupportedRootDeviceTypes)); err != nil {
		return fmt.Errorf("error setting supported_root_device_types: %s", err)
	}
	if err := d.Set("bare_metal", aws.BoolValue(v.BareMetal)); err != nil {
		return fmt.Errorf("error setting bare_metal: %s", err)
	}
	if v.Hypervisor != nil {
		if err := d.Set("hypervisor", aws.StringValue(v.Hypervisor)); err != nil {
			return fmt.Errorf("error setting hypervisor: %s", err)
		}
	}
	if err := d.Set("supported_architectures", aws.StringValueSlice(v.ProcessorInfo.SupportedArchitectures)); err != nil {
		return fmt.Errorf("error setting supported_architectures: %s", err)
	}
	if err := d.Set("sustained_clock_speed", aws.Float64Value(v.ProcessorInfo.SustainedClockSpeedInGhz)); err != nil {
		return fmt.Errorf("error setting sustained_clock_speed: %s", err)
	}
	if err := d.Set("default_vcpus", aws.Int64Value(v.VCpuInfo.DefaultVCpus)); err != nil {
		return fmt.Errorf("error setting default_vcpus: %s", err)
	}
	if v.VCpuInfo.DefaultCores != nil {
		if err := d.Set("default_cores", aws.Int64Value(v.VCpuInfo.DefaultCores)); err != nil {
			return fmt.Errorf("error setting default_cores: %s", err)
		}
	}
	if v.VCpuInfo.DefaultThreadsPerCore != nil {
		if err := d.Set("default_threads_per_core", aws.Int64Value(v.VCpuInfo.DefaultThreadsPerCore)); err != nil {
			return fmt.Errorf("error setting default_threads_per_core: %s", err)
		}
	}
	if v.VCpuInfo.ValidThreadsPerCore != nil {
		if err := d.Set("valid_threads_per_core", aws.Int64ValueSlice(v.VCpuInfo.ValidThreadsPerCore)); err != nil {
			return fmt.Errorf("error setting valid_threads_per_core: %s", err)
		}
	}
	if v.VCpuInfo.ValidCores != nil {
		if err := d.Set("valid_cores", aws.Int64ValueSlice(v.VCpuInfo.ValidCores)); err != nil {
			return fmt.Errorf("error setting valid_cores: %s", err)
		}
	}
	if err := d.Set("memory_size", aws.Int64Value(v.MemoryInfo.SizeInMiB)); err != nil {
		return fmt.Errorf("error setting memory_size: %s", err)
	}
	if err := d.Set("instance_storage_supported", aws.BoolValue(v.InstanceStorageSupported)); err != nil {
		return fmt.Errorf("error setting instance_storage_supported: %s", err)
	}
	if v.InstanceStorageInfo != nil {
		if err := d.Set("total_instance_storage", aws.Int64Value(v.InstanceStorageInfo.TotalSizeInGB)); err != nil {
			return fmt.Errorf("error setting total_instance_storage: %s", err)
		}
		if v.InstanceStorageInfo.Disks != nil {
			diskList := make([]interface{}, len(v.InstanceStorageInfo.Disks))
			for i, dk := range v.InstanceStorageInfo.Disks {
				disk := map[string]interface{}{
					"size":  aws.Int64Value(dk.SizeInGB),
					"count": aws.Int64Value(dk.Count),
					"type":  aws.StringValue(dk.Type),
				}
				diskList[i] = disk
			}
			if err := d.Set("instance_disks", diskList); err != nil {
				return fmt.Errorf("error setting instance_disks: %s", err)
			}
		}
	}
	if err := d.Set("ebs_optimized_support", aws.StringValue(v.EbsInfo.EbsOptimizedSupport)); err != nil {
		return fmt.Errorf("error setting ebs_optimized_support: %s", err)
	}
	if err := d.Set("ebs_encryption_support", aws.StringValue(v.EbsInfo.EncryptionSupport)); err != nil {
		return fmt.Errorf("error setting ebs_encryption_support: %s", err)
	}
	if err := d.Set("network_performance", aws.StringValue(v.NetworkInfo.NetworkPerformance)); err != nil {
		return fmt.Errorf("error setting network_performance: %s", err)
	}
	if err := d.Set("maximum_network_interfaces", aws.Int64Value(v.NetworkInfo.MaximumNetworkInterfaces)); err != nil {
		return fmt.Errorf("error setting maximum_network_interfaces: %s", err)
	}
	if err := d.Set("maximum_ipv4_addresses_per_interface", aws.Int64Value(v.NetworkInfo.Ipv4AddressesPerInterface)); err != nil {
		return fmt.Errorf("error setting ipv4_addresses_per_interface: %s", err)
	}
	if err := d.Set("maximum_ipv6_addresses_per_interface", aws.Int64Value(v.NetworkInfo.Ipv6AddressesPerInterface)); err != nil {
		return fmt.Errorf("error setting ipv6_addresses_per_interface: %s", err)
	}
	if err := d.Set("ipv6_supported", aws.BoolValue(v.NetworkInfo.Ipv6Supported)); err != nil {
		return fmt.Errorf("error setting ipv6_supported: %s", err)
	}
	if err := d.Set("ena_support", aws.StringValue(v.NetworkInfo.EnaSupport)); err != nil {
		return fmt.Errorf("error setting ena_support: %s", err)
	}
	if v.GpuInfo != nil {
		gpuList := make([]interface{}, len(v.GpuInfo.Gpus))
		for i, gp := range v.GpuInfo.Gpus {
			gpu := map[string]interface{}{
				"manufacturer": aws.StringValue(gp.Manufacturer),
				"name":         aws.StringValue(gp.Name),
				"count":        aws.Int64Value(gp.Count),
				"memory_size":  aws.Int64Value(gp.MemoryInfo.SizeInMiB),
			}
			gpuList[i] = gpu
		}
		if err := d.Set("gpus", gpuList); err != nil {
			return fmt.Errorf("error setting gpu: %s", err)
		}
		if err := d.Set("total_gpu_memory", aws.Int64Value(v.GpuInfo.TotalGpuMemoryInMiB)); err != nil {
			return fmt.Errorf("error setting total_gpu_memory: %s", err)
		}
	}
	if v.FpgaInfo != nil {
		fpgaList := make([]interface{}, len(v.FpgaInfo.Fpgas))
		for i, fpg := range v.FpgaInfo.Fpgas {
			fpga := map[string]interface{}{
				"manufacturer": aws.StringValue(fpg.Manufacturer),
				"name":         aws.StringValue(fpg.Name),
				"count":        aws.Int64Value(fpg.Count),
				"memory_size":  aws.Int64Value(fpg.MemoryInfo.SizeInMiB),
			}
			fpgaList[i] = fpga
		}
		if err := d.Set("fpgas", fpgaList); err != nil {
			return fmt.Errorf("error setting fpga: %s", err)
		}
		if err := d.Set("total_fpga_memory", aws.Int64Value(v.FpgaInfo.TotalFpgaMemoryInMiB)); err != nil {
			return fmt.Errorf("error setting total_fpga_memory: %s", err)
		}
	}
	if err := d.Set("supported_placement_strategies", aws.StringValueSlice(v.PlacementGroupInfo.SupportedStrategies)); err != nil {
		return fmt.Errorf("error setting supported_placement_strategies: %s", err)
	}
	if v.InferenceAcceleratorInfo != nil {
		acceleratorList := make([]interface{}, len(v.InferenceAcceleratorInfo.Accelerators))
		for i, accl := range v.InferenceAcceleratorInfo.Accelerators {
			accelerator := map[string]interface{}{
				"manufacturer": aws.StringValue(accl.Manufacturer),
				"name":         aws.StringValue(accl.Name),
				"count":        aws.Int64Value(accl.Count),
			}
			acceleratorList[i] = accelerator
		}
		if err := d.Set("accelerators", acceleratorList); err != nil {
			return fmt.Errorf("error setting fpga: %s", err)
		}
	}
	if err := d.Set("hibernation_supported", aws.BoolValue(v.HibernationSupported)); err != nil {
		return fmt.Errorf("error setting hibernation_supported: %s", err)
	}
	if err := d.Set("burstable_performance_supported", aws.BoolValue(v.BurstablePerformanceSupported)); err != nil {
		return fmt.Errorf("error setting burstable_performance_supported: %s", err)
	}
	if err := d.Set("dedicated_hosts_supported", aws.BoolValue(v.DedicatedHostsSupported)); err != nil {
		return fmt.Errorf("error setting dedicated_hosts_supported: %s", err)
	}
	if err := d.Set("auto_recovery_supported", aws.BoolValue(v.AutoRecoverySupported)); err != nil {
		return fmt.Errorf("error setting auto_recovery_supported: %s", err)
	}
	d.SetId(aws.StringValue(v.InstanceType))
	return nil
}
