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
	d.Set("free_tier_eligible", v.FreeTierEligible)
	d.Set("supported_usages_classes", v.SupportedUsageClasses)
	d.Set("supported_root_device_types", v.SupportedRootDeviceTypes)
	d.Set("bare_metal", v.BareMetal)
	d.Set("hypervisor", v.Hypervisor)
	d.Set("supported_architectures", v.ProcessorInfo.SupportedArchitectures)
	d.Set("sustained_clock_speed", v.ProcessorInfo.SustainedClockSpeedInGhz)
	d.Set("default_vcpus", v.VCpuInfo.DefaultVCpus)
	d.Set("default_cores", v.VCpuInfo.DefaultCores)
	d.Set("default_threads_per_core", v.VCpuInfo.DefaultThreadsPerCore)
	d.Set("valid_threads_per_core", v.VCpuInfo.ValidThreadsPerCore)
	d.Set("valid_cores", v.VCpuInfo.ValidCores)
	d.Set("memory_size", v.MemoryInfo.SizeInMiB)
	d.Set("instance_storage_supported", v.InstanceStorageSupported)
	if v.InstanceStorageInfo != nil {
		d.Set("total_instance_storage", v.InstanceStorageInfo.TotalSizeInGB)
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
			d.Set("instance_disks", diskList)
		}
	}
	d.Set("ebs_optimized_support", v.EbsInfo.EbsOptimizedSupport)
	d.Set("ebs_encryption_support", v.EbsInfo.EncryptionSupport)
	d.Set("network_performance", v.NetworkInfo.NetworkPerformance)
	d.Set("maximum_network_interfaces", v.NetworkInfo.MaximumNetworkInterfaces)
	d.Set("maximum_ipv4_addresses_per_interface", v.NetworkInfo.Ipv4AddressesPerInterface)
	d.Set("maximum_ipv6_addresses_per_interface", v.NetworkInfo.Ipv6AddressesPerInterface)
	d.Set("ipv6_supported", v.NetworkInfo.Ipv6Supported)
	d.Set("ena_support", v.NetworkInfo.EnaSupport)
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
		d.Set("gpus", gpuList)
		d.Set("total_gpu_memory", v.GpuInfo.TotalGpuMemoryInMiB)
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
		d.Set("fpgas", fpgaList)
		d.Set("total_fpga_memory", v.FpgaInfo.TotalFpgaMemoryInMiB)
	}
	d.Set("supported_placement_strategies", v.PlacementGroupInfo.SupportedStrategies)
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
		d.Set("accelerators", acceleratorList)
	}
	d.Set("hibernation_supported", v.HibernationSupported)
	d.Set("burstable_performance_supported", v.BurstablePerformanceSupported)
	d.Set("dedicated_hosts_supported", v.DedicatedHostsSupported)
	d.Set("auto_recovery_supported", v.AutoRecoverySupported)
	d.SetId(aws.StringValue(v.InstanceType))
	return nil
}
