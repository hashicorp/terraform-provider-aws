package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/guardduty"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsGuardDutyFilter() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsGuardDutyFilterCreate,
		Read:   resourceAwsGuardDutyFilterRead,
		// Update: resourceAwsGuardDutyFilterUpdate,
		Delete: resourceAwsGuardDutyFilterDelete,

		// Importer: &schema.ResourceImporter{
		// 	State: schema.ImportStatePassthrough,
		// },
		Schema: map[string]*schema.Schema{ // TODO: add validations
			"detector_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true, // perhaps remove here and below, when Update is back
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true, // perhaps remove here and below, when Update is back
			},
			"tags": tagsSchemaForceNew(),
			"finding_criteria": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Required: true,
				ForceNew: true, // perhaps remove here and below, when Update is back
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"criterion": &schema.Schema{
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"account_id": &schema.Schema{
										Type:     schema.TypeSet,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"equals": {
													Type:     schema.TypeList,
													Optional: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
												"not_equals": {
													Type:     schema.TypeList,
													Optional: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
											},
										},
									},
									"region": &schema.Schema{
										Type:     schema.TypeSet,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"equals": {
													Type:     schema.TypeList,
													Optional: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
												"not_equals": {
													Type:     schema.TypeList,
													Optional: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
											},
										},
									},
									"confidence": &schema.Schema{
										Type:     schema.TypeSet,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"equals": {
													Type:     schema.TypeList,
													Optional: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
												"greater_than": {
													Type:     schema.TypeInt,
													Optional: true,
												},
												"greater_than_or_equal": {
													Type:     schema.TypeInt,
													Optional: true,
												},
												"less_than": {
													Type:     schema.TypeInt,
													Optional: true,
												},
												"less_than_or_equal": {
													Type:     schema.TypeInt,
													Optional: true,
												},
												"not_equals": {
													Type:     schema.TypeList,
													Optional: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
											},
										},
									},
									"id": &schema.Schema{
										Type:     schema.TypeSet,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"equals": {
													Type:     schema.TypeList,
													Optional: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
												"greater_than": {
													Type:     schema.TypeInt,
													Optional: true,
												},
												"greater_than_or_equal": {
													Type:     schema.TypeInt,
													Optional: true,
												},
												"less_than": {
													Type:     schema.TypeInt,
													Optional: true,
												},
												"less_than_or_equal": {
													Type:     schema.TypeInt,
													Optional: true,
												},
												"not_equals": {
													Type:     schema.TypeList,
													Optional: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
											},
										},
									},
									"resource.accessKeyDetails.accessKeyId": &schema.Schema{
										Type:     schema.TypeSet,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"equals": {
													Type:     schema.TypeList,
													Optional: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
												"not_equals": {
													Type:     schema.TypeList,
													Optional: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
											},
										},
									},
									"resource.accessKeyDetails.principalId": &schema.Schema{
										Type:     schema.TypeSet,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"equals": {
													Type:     schema.TypeList,
													Optional: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
												"not_equals": {
													Type:     schema.TypeList,
													Optional: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
											},
										},
									},
									"resource.accessKeyDetails.userName": &schema.Schema{
										Type:     schema.TypeSet,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"equals": {
													Type:     schema.TypeList,
													Optional: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
												"not_equals": {
													Type:     schema.TypeList,
													Optional: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
											},
										},
									},
									"resource.accessKeyDetails.userType": &schema.Schema{
										Type:     schema.TypeSet,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"equals": {
													Type:     schema.TypeList,
													Optional: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
												"not_equals": {
													Type:     schema.TypeList,
													Optional: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
											},
										},
									},
									"resource.instanceDetails.iamInstanceProfile.id": &schema.Schema{
										Type:     schema.TypeSet,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"equals": {
													Type:     schema.TypeList,
													Optional: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
												"not_equals": {
													Type:     schema.TypeList,
													Optional: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
											},
										},
									},
									"resource.instanceDetails.imageId": &schema.Schema{
										Type:     schema.TypeSet,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"equals": {
													Type:     schema.TypeList,
													Optional: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
												"not_equals": {
													Type:     schema.TypeList,
													Optional: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
											},
										},
									},
									"resource.instanceDetails.instanceId": &schema.Schema{
										Type:     schema.TypeSet,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"equals": {
													Type:     schema.TypeList,
													Optional: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
												"not_equals": {
													Type:     schema.TypeList,
													Optional: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
											},
										},
									},
									"resource.instanceDetails.networkInterfaces.ipv6Addresses": &schema.Schema{
										Type:     schema.TypeSet,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"equals": {
													Type:     schema.TypeList,
													Optional: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
												"not_equals": {
													Type:     schema.TypeList,
													Optional: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
											},
										},
									},
									"resource.instanceDetails.networkInterfaces.privateIpAddresses.privateIpAddress": &schema.Schema{
										Type:     schema.TypeSet,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"equals": {
													Type:     schema.TypeList,
													Optional: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
												"not_equals": {
													Type:     schema.TypeList,
													Optional: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
											},
										},
									},
									"resource.instanceDetails.networkInterfaces.publicDnsName": &schema.Schema{
										Type:     schema.TypeSet,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"equals": {
													Type:     schema.TypeList,
													Optional: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
												"not_equals": {
													Type:     schema.TypeList,
													Optional: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
											},
										},
									},
									"resource.instanceDetails.networkInterfaces.publicIp": &schema.Schema{
										Type:     schema.TypeSet,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"equals": {
													Type:     schema.TypeList,
													Optional: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
												"not_equals": {
													Type:     schema.TypeList,
													Optional: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
											},
										},
									},
									"resource.instanceDetails.networkInterfaces.securityGroups.groupId": &schema.Schema{
										Type:     schema.TypeSet,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"equals": {
													Type:     schema.TypeList,
													Optional: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
												"not_equals": {
													Type:     schema.TypeList,
													Optional: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
											},
										},
									},
									"resource.instanceDetails.networkInterfaces.securityGroups.groupName": &schema.Schema{
										Type:     schema.TypeSet,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"equals": {
													Type:     schema.TypeList,
													Optional: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
												"not_equals": {
													Type:     schema.TypeList,
													Optional: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
											},
										},
									},
									"resource.instanceDetails.networkInterfaces.subnetId": &schema.Schema{
										Type:     schema.TypeSet,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"equals": {
													Type:     schema.TypeList,
													Optional: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
												"not_equals": {
													Type:     schema.TypeList,
													Optional: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
											},
										},
									},
									"resource.instanceDetails.networkInterfaces.vpcId": &schema.Schema{
										Type:     schema.TypeSet,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"equals": {
													Type:     schema.TypeList,
													Optional: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
												"not_equals": {
													Type:     schema.TypeList,
													Optional: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
											},
										},
									},
									"resource.instanceDetails.tags.key": &schema.Schema{
										Type:     schema.TypeSet,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"equals": {
													Type:     schema.TypeList,
													Optional: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
												"not_equals": {
													Type:     schema.TypeList,
													Optional: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
											},
										},
									},
									"resource.instanceDetails.tags.value": &schema.Schema{
										Type:     schema.TypeSet,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"equals": {
													Type:     schema.TypeList,
													Optional: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
												"not_equals": {
													Type:     schema.TypeList,
													Optional: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
											},
										},
									},
									"resource.resourceType": &schema.Schema{
										Type:     schema.TypeSet,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"equals": {
													Type:     schema.TypeList,
													Optional: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
												"not_equals": {
													Type:     schema.TypeList,
													Optional: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
											},
										},
									},
									"service.action.actionType": &schema.Schema{
										Type:     schema.TypeSet,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"equals": {
													Type:     schema.TypeList,
													Optional: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
												"not_equals": {
													Type:     schema.TypeList,
													Optional: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
											},
										},
									},
									"service.action.awsApiCallAction.api": &schema.Schema{
										Type:     schema.TypeSet,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"equals": {
													Type:     schema.TypeList,
													Optional: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
												"not_equals": {
													Type:     schema.TypeList,
													Optional: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
											},
										},
									},
									"service.action.awsApiCallAction.callerType": &schema.Schema{
										Type:     schema.TypeSet,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"equals": {
													Type:     schema.TypeList,
													Optional: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
												"not_equals": {
													Type:     schema.TypeList,
													Optional: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
											},
										},
									},
									"service.action.awsApiCallAction.remoteIpDetails.city.cityName": &schema.Schema{
										Type:     schema.TypeSet,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"equals": {
													Type:     schema.TypeList,
													Optional: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
												"not_equals": {
													Type:     schema.TypeList,
													Optional: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
											},
										},
									},
									"service.action.awsApiCallAction.remoteIpDetails.country.countryName": &schema.Schema{
										Type:     schema.TypeSet,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"equals": {
													Type:     schema.TypeList,
													Optional: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
												"not_equals": {
													Type:     schema.TypeList,
													Optional: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
											},
										},
									},
									"service.action.awsApiCallAction.remoteIpDetails.ipAddressV4": &schema.Schema{
										Type:     schema.TypeSet,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"equals": {
													Type:     schema.TypeList,
													Optional: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
												"not_equals": {
													Type:     schema.TypeList,
													Optional: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
											},
										},
									},
									"service.action.awsApiCallAction.remoteIpDetails.organization.asn": &schema.Schema{
										Type:     schema.TypeSet,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"equals": {
													Type:     schema.TypeList,
													Optional: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
												"not_equals": {
													Type:     schema.TypeList,
													Optional: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
											},
										},
									},
									"service.action.awsApiCallAction.remoteIpDetails.organization.asnOrg": &schema.Schema{
										Type:     schema.TypeSet,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"equals": {
													Type:     schema.TypeList,
													Optional: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
												"not_equals": {
													Type:     schema.TypeList,
													Optional: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
											},
										},
									},
									"service.action.awsApiCallAction.serviceName": &schema.Schema{
										Type:     schema.TypeSet,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"equals": {
													Type:     schema.TypeList,
													Optional: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
												"not_equals": {
													Type:     schema.TypeList,
													Optional: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
											},
										},
									},
									"service.action.dnsRequestAction.domain": &schema.Schema{
										Type:     schema.TypeSet,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"equals": {
													Type:     schema.TypeList,
													Optional: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
												"not_equals": {
													Type:     schema.TypeList,
													Optional: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
											},
										},
									},
									"service.action.networkConnectionAction.blocked": &schema.Schema{
										Type:     schema.TypeSet,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"equals": {
													Type:     schema.TypeList,
													Optional: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
												"not_equals": {
													Type:     schema.TypeList,
													Optional: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
											},
										},
									},
									"service.action.networkConnectionAction.connectionDirection": &schema.Schema{
										Type:     schema.TypeSet,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"equals": {
													Type:     schema.TypeList,
													Optional: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
												"not_equals": {
													Type:     schema.TypeList,
													Optional: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
											},
										},
									},
									"service.action.networkConnectionAction.localPortDetails.port": &schema.Schema{
										Type:     schema.TypeSet,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"equals": {
													Type:     schema.TypeList,
													Optional: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
												"not_equals": {
													Type:     schema.TypeList,
													Optional: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
											},
										},
									},
									"service.action.networkConnectionAction.protocol": &schema.Schema{
										Type:     schema.TypeSet,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"equals": {
													Type:     schema.TypeList,
													Optional: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
												"not_equals": {
													Type:     schema.TypeList,
													Optional: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
											},
										},
									},
									"service.action.networkConnectionAction.remoteIpDetails.city.cityName": &schema.Schema{
										Type:     schema.TypeSet,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"equals": {
													Type:     schema.TypeList,
													Optional: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
												"not_equals": {
													Type:     schema.TypeList,
													Optional: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
											},
										},
									},
									"service.action.networkConnectionAction.remoteIpDetails.country.countryName": &schema.Schema{
										Type:     schema.TypeSet,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"equals": {
													Type:     schema.TypeList,
													Optional: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
												"not_equals": {
													Type:     schema.TypeList,
													Optional: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
											},
										},
									},
									"service.action.networkConnectionAction.remoteIpDetails.ipAddressV4": &schema.Schema{
										Type:     schema.TypeSet,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"equals": {
													Type:     schema.TypeList,
													Optional: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
												"not_equals": {
													Type:     schema.TypeList,
													Optional: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
											},
										},
									},
									"service.action.networkConnectionAction.remoteIpDetails.organization.asn": &schema.Schema{
										Type:     schema.TypeSet,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"equals": {
													Type:     schema.TypeList,
													Optional: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
												"not_equals": {
													Type:     schema.TypeList,
													Optional: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
											},
										},
									},
									"service.action.networkConnectionAction.remoteIpDetails.organization.asnOrg": &schema.Schema{
										Type:     schema.TypeSet,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"equals": {
													Type:     schema.TypeList,
													Optional: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
												"not_equals": {
													Type:     schema.TypeList,
													Optional: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
											},
										},
									},
									"service.action.networkConnectionAction.remotePortDetails.port": &schema.Schema{
										Type:     schema.TypeSet,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"equals": {
													Type:     schema.TypeList,
													Optional: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
												"not_equals": {
													Type:     schema.TypeList,
													Optional: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
											},
										},
									},
									"service.additionalInfo.threatListName": &schema.Schema{
										Type:     schema.TypeSet,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"equals": {
													Type:     schema.TypeList,
													Optional: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
												"not_equals": {
													Type:     schema.TypeList,
													Optional: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
											},
										},
									},
									"service.archived": &schema.Schema{
										Type:     schema.TypeSet,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"equals": {
													Type:     schema.TypeBool,
													Optional: true,
												},
												"not_equals": {
													Type:     schema.TypeBool,
													Optional: true,
												},
											},
										},
									},
									"service.resourceRole": &schema.Schema{
										Type:     schema.TypeSet,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"equals": {
													Type:     schema.TypeList,
													Optional: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
												"not_equals": {
													Type:     schema.TypeList,
													Optional: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
											},
										},
									},
									"severity": &schema.Schema{
										Type:     schema.TypeSet,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"equals": {
													Type:     schema.TypeList,
													Optional: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
												"not_equals": {
													Type:     schema.TypeList,
													Optional: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
											},
										},
									},
									"type": &schema.Schema{
										Type:     schema.TypeSet,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"equals": {
													Type:     schema.TypeList,
													Optional: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
												"not_equals": {
													Type:     schema.TypeList,
													Optional: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
											},
										},
									},
									"updatedAt": &schema.Schema{
										Type:     schema.TypeSet,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"equals": {
													Type:     schema.TypeList,
													Optional: true,
													Elem:     &schema.Schema{Type: schema.TypeInt},
												},
												"not_equals": {
													Type:     schema.TypeList,
													Optional: true,
													Elem:     &schema.Schema{Type: schema.TypeInt},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			"action": {
				Type:     schema.TypeString, // should have a new type or a validation for NOOP/ARCHIVE
				Optional: true,
				ForceNew: true, // perhaps remove here and below, when Update is back
			},
			"rank": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true, // perhaps remove here and below, when Update is back
			},
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Second),
			Update: schema.DefaultTimeout(60 * time.Second),
		},
	}
}

func resourceAwsGuardDutyFilterCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).guarddutyconn

	input := guardduty.CreateFilterInput{
		DetectorId:  aws.String(d.Get("detector_id").(string)),
		Name:        aws.String(d.Get("name").(string)),
		Description: aws.String(d.Get("description").(string)),
		Rank:        aws.Int64(int64(d.Get("rank").(int))),
	}

	// building `FindingCriteria`
	findingCriteria := d.Get("finding_criteria").([]interface{})[0].(map[string]interface{})
	criterion := findingCriteria["criterion"].(*schema.Set).List()[0].(map[string]interface{})
	condition := criterion["region"].(*schema.Set).List()[0].(map[string]interface{})

	interfaceForEquals := condition["equals"].([]interface{})

	equals := make([]string, len(interfaceForEquals))
	for i, v := range interfaceForEquals {
		equals[i] = string(v.(string)) // Maybe aws.string?
	}

	interfaceForNotEquals := condition["not_equals"].([]interface{})

	notEquals := make([]string, len(interfaceForNotEquals))
	for i, v := range interfaceForNotEquals {
		notEquals[i] = string(v.(string))
	}

	tagsInterface := d.Get("tags").(map[string]interface{})
	if len(tagsInterface) > 0 {
		tags := make(map[string]*string, len(tagsInterface))
		for i, v := range tagsInterface {
			tags[i] = aws.String(v.(string))
		}

		input.Tags = tags
	}

	//input.FindingCriteria = &guardduty.FindingCriteria{
	//	Criterion: map[string]*guardduty.Condition{
	//		"condition": &guardduty.Condition{
	//			Equals:             aws.StringSlice(equals),
	//			GreaterThan:        aws.Int64(int64(condition["greater_than"].(int))),
	//			GreaterThanOrEqual: aws.Int64(int64(condition["greater_than_or_equal"].(int))),
	//			LessThan:           aws.Int64(int64(condition["less_than"].(int))),
	//			LessThanOrEqual:    aws.Int64(int64(condition["less_than_or_equal"].(int))),
	//			NotEquals:          aws.StringSlice(notEquals),
	//		},
	//	},
	//}
	// Currently it works only for region, must be expanded to all other resources
	input.FindingCriteria = &guardduty.FindingCriteria{
		Criterion: map[string]*guardduty.Condition{
			"region": &guardduty.Condition{
				Equals: aws.StringSlice(equals),
			},
		},
	}
	log.Printf("[DEBUG] Creating FindingCriteria map: %#v", findingCriteria)

	// Setting the default value for `action`
	action := "NOOP"

	if len(d.Get("action").(string)) > 0 {
		action = d.Get("action").(string)
	}

	input.Action = aws.String(action)

	log.Printf("[DEBUG] Creating GuardDuty Filter: %s", input)
	output, err := conn.CreateFilter(&input)
	if err != nil {
		return fmt.Errorf("Creating GuardDuty Filter failed: %s", err.Error())
	}

	d.SetId(*output.Name)

	return resourceAwsGuardDutyFilterRead(d, meta)
}

func resourceAwsGuardDutyFilterRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).guarddutyconn
	detectorId := d.Get("detector_id").(string)
	name := d.Get("name").(string)

	input := guardduty.GetFilterInput{
		DetectorId: aws.String(detectorId),
		FilterName: aws.String(name),
	}

	log.Printf("[DEBUG] Reading GuardDuty Filter: %s", input)
	filter, err := conn.GetFilter(&input)

	if err != nil {
		if isAWSErr(err, guardduty.ErrCodeBadRequestException, "The request is rejected because the input detectorId is not owned by the current account.") {
			log.Printf("[WARN] GuardDuty detector %q not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Reading GuardDuty Filter '%s' failed: %s", name, err.Error())
	}

	d.Set("action", filter.Action) // Make sure I really want to set all these attrs
	d.Set("description", filter.Description)
	d.Set("rank", filter.Rank)
	d.Set("name", d.Id())

	return nil
}

// func resourceAwsGuardDutyFilterUpdate(d *schema.ResourceData, meta interface{}) error {
// 	conn := meta.(*AWSClient).guarddutyconn
//
// 	accountID, detectorID, err := decodeGuardDutyMemberID(d.Id())
// 	if err != nil {
// 		return err
// 	}
//
// 	if d.HasChange("invite") {
// 		if d.Get("invite").(bool) {
// 			input := &guardduty.InviteMembersInput{
// 				DetectorId:               aws.String(detectorID),
// 				AccountIds:               []*string{aws.String(accountID)},
// 				DisableEmailNotification: aws.Bool(d.Get("disable_email_notification").(bool)),
// 				Message:                  aws.String(d.Get("invitation_message").(string)),
// 			}
//
// 			log.Printf("[INFO] Inviting GuardDuty Member: %s", input)
// 			output, err := conn.InviteMembers(input)
// 			if err != nil {
// 				return fmt.Errorf("error inviting GuardDuty Member %q: %s", d.Id(), err)
// 			}
//
// 			// {"unprocessedAccounts":[{"result":"The request is rejected because the current account has already invited or is already the GuardDuty master of the given member account ID.","accountId":"067819342479"}]}
// 			if len(output.UnprocessedAccounts) > 0 {
// 				return fmt.Errorf("error inviting GuardDuty Member %q: %s", d.Id(), aws.StringValue(output.UnprocessedAccounts[0].Result))
// 			}
// 		} else {
// 			input := &guardduty.DisassociateMembersInput{
// 				AccountIds: []*string{aws.String(accountID)},
// 				DetectorId: aws.String(detectorID),
// 			}
// 			log.Printf("[INFO] Disassociating GuardDuty Member: %s", input)
// 			_, err := conn.DisassociateMembers(input)
// 			if err != nil {
// 				return fmt.Errorf("error disassociating GuardDuty Member %q: %s", d.Id(), err)
// 			}
// 		}
// 	}
//
// 	return resourceAwsGuardDutyFilterRead(d, meta)
// }

func resourceAwsGuardDutyFilterDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).guarddutyconn

	detectorId := d.Get("detector_id").(string)
	name := d.Get("name").(string)

	input := guardduty.DeleteFilterInput{
		FilterName: aws.String(name),
		DetectorId: aws.String(detectorId),
	}

	log.Printf("[DEBUG] Delete GuardDuty Filter: %s", input)

	_, err := conn.DeleteFilter(&input)
	if err != nil {
		return fmt.Errorf("Deleting GuardDuty Filter '%s' failed: %s", d.Id(), err.Error())
	}
	return nil
}
