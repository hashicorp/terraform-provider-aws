# Relationship Resource Design Standards

**Summary:** Align on design standards for relationship management resources in the Terraform AWS Provider.  
**Created**: 2022-07-11  

---

The goal of this document is to assess the design of existing "relationship" resources in the Terraform AWS Provider and determine if a consistent set of rules can be defined for implementing them. For the purpose of this document, a "relationship" resource is defined as a resource which manages either a direct relationship between two standalone resources ("one-to-one", ie. [`aws_ssoadmin_permission_boundary_attachment`](https://registry.terraform.io/providers/-/aws/latest/docs/resources/ssoadmin_permissions_boundary_attachment)), or a variable number of child relationships to a parent resource ("one-to-many", ie. [`aws_iam_role_policy_attachment`](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/iam_role_policy_attachment)). Resources and AWS APIs with this function will often contain suffixes like "attachment", "assignment", "registration", or "rule".

A documented standard for implementing relationship-styled resources will inform how new resources are written, and provide guidelines to refer back to when the community requests features which may not align with internal best practices.

## Background

The first form of relationship resources ("one-to-one") typically have a straightforward, singular API design and provider implementation given only a single relationship exists. The second form of resources ("one-to-many") often has to balance two provider design principles:

* [Resources should represent a single API object](https://developer.hashicorp.com/terraform/plugin/best-practices/hashicorp-provider-design-principles#resources-should-represent-a-single-api-object)
* [Resource and attribute schema should closely match the underlying API](https://developer.hashicorp.com/terraform/plugin/best-practices/hashicorp-provider-design-principles#resource-and-attribute-schema-should-closely-match-the-underlying-api)

In these cases, the smallest possible piece of infrastructure may be a single parent-child relationship, while the AWS APIs may accept and return lists of parent-child relationships. The first principle would favor a resource representing a single relationship, while the second principle suggests a resource should manage a variable number of relationships. Additionally, practitioners coming from the AWS CLI or SDK might also have expectations about how resource schemas should be shaped compared to CLI flags or SDK inputs.

An analysis of existing resources can inform which of these principles maintainers have given precedence to up to this point. The table below documents existing relationship resources in the AWS provider. This table should not be considered exhaustive[^1], but covers a large majority of the resources implementing the patterns discussed above.

| **Resource Name** | **Form** | **AWS API** | **Terraform** |
| --- | --- | --- | --- |
| [aws_alb_target_group_attachment](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/lb_target_group_attachment) | One-to-many | [Plural](https://docs.aws.amazon.com/elasticloadbalancing/latest/APIReference/API_RegisterTargets.html) | [Singular](https://github.com/hashicorp/terraform-provider-aws/blob/v5.7.0/internal/service/elbv2/target_group_attachment.go#L76-L79) |
| [aws_autoscaling_attachment](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/autoscaling_attachment) (ELB) | One-to-many | [Plural](https://docs.aws.amazon.com/autoscaling/ec2/APIReference/API_AttachLoadBalancers.html) | [Singular](https://github.com/hashicorp/terraform-provider-aws/blob/v5.7.0/internal/service/autoscaling/attachment.go#L58-L61) |
| [aws_autoscaling_attachment](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/autoscaling_attachment) (Target Group ARN) | One-to-many | [Plural](https://docs.aws.amazon.com/autoscaling/ec2/APIReference/API_AttachLoadBalancerTargetGroups.html)| [Singular](https://github.com/hashicorp/terraform-provider-aws/blob/v5.7.0/internal/service/autoscaling/attachment.go#L74-L77) |
| [aws_autoscaling_traffic_source_attachment](https://registry.terraform.io/providers/-/aws/latest/docs/resources/autoscaling_traffic_source_attachment) | One-to-many | [Plural](https://docs.aws.amazon.com/autoscaling/ec2/APIReference/API_AttachTrafficSources.html)| [Singular](https://github.com/hashicorp/terraform-provider-aws/blob/v5.7.0/internal/service/autoscaling/traffic_source_attachment.go#L78-L81) |
| [aws_cognito_identity_pool_roles_attachment](https://registry.terraform.io/providers/-/aws/latest/docs/resources/cognito_identity_pool_roles_attachment)[^2] | One-to-many | [Plural](https://docs.aws.amazon.com/cognitoidentity/latest/APIReference/API_SetIdentityPoolRoles.html) | [Plural](https://github.com/hashicorp/terraform-provider-aws/blob/v5.7.0/internal/service/cognitoidentity/pool_roles_attachment.go#L123-L126) |
| [aws_ec2_transit_gateway_peering_attachment](https://registry.terraform.io/providers/-/aws/latest/docs/resources/ec2_transit_gateway_peering_attachment) | One-to-one | [Singular](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_CreateTransitGatewayPeeringAttachment.html) | [Singular](https://github.com/hashicorp/terraform-provider-aws/blob/v5.7.0/internal/service/ec2/transitgateway_peering_attachment.go#L75-L81) |
| [aws_ec2_transit_gateway_vpc_attachment](https://registry.terraform.io/providers/-/aws/latest/docs/resources/ec2_transit_gateway_vpc_attachment) | One-to-one | [Singular](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_CreateTransitGatewayVpcAttachment.html) | [Singular](https://github.com/hashicorp/terraform-provider-aws/blob/v5.7.0/internal/service/ec2/transitgateway_vpc_attachment.go#L102-L112) |
| [aws_elb_attachment](https://registry.terraform.io/providers/-/aws/latest/docs/resources/elb_attachment) | One-to-many | [Plural](https://docs.aws.amazon.com/elasticloadbalancing/2012-06-01/APIReference/API_RegisterInstancesWithLoadBalancer.html) | [Singular](https://github.com/hashicorp/terraform-provider-aws/blob/v5.7.0/internal/service/elb/attachment.go#L54-L57) |
| [aws_iam_group_policy_attachment](https://registry.terraform.io/providers/-/aws/latest/docs/resources/iam_group_policy_attachment) | One-to-many | [Singular](https://docs.aws.amazon.com/IAM/latest/APIReference/API_AttachGroupPolicy.html)| [Singular](https://github.com/hashicorp/terraform-provider-aws/blob/v5.7.0/internal/service/iam/group_policy_attachment.go#L153-L156) |
| [aws_iam_policy_attachment](https://registry.terraform.io/providers/-/aws/latest/docs/resources/iam_policy_attachment)[^3] | One-to-many | [Singular](https://docs.aws.amazon.com/IAM/latest/APIReference/API_AttachRolePolicy.html) | [Plural](https://github.com/hashicorp/terraform-provider-aws/blob/v5.7.0/internal/service/iam/policy_attachment.go#L219-L230) |
| [aws_iam_role_policy_attachment](https://registry.terraform.io/providers/-/aws/latest/docs/resources/iam_role_policy_attachment) | One-to-many | [Singular](https://docs.aws.amazon.com/IAM/latest/APIReference/API_AttachRolePolicy.html) | [Singular](https://github.com/hashicorp/terraform-provider-aws/blob/v5.7.0/internal/service/iam/role_policy_attachment.go#L160-L163) |
| [aws_iam_user_policy_attachment](https://registry.terraform.io/providers/-/aws/latest/docs/resources/iam_user_policy_attachment) | One-to-many | [Singular](https://docs.aws.amazon.com/IAM/latest/APIReference/API_AttachUserPolicy.html) | [Singular](https://github.com/hashicorp/terraform-provider-aws/blob/v5.7.0/internal/service/iam/user_policy_attachment.go#L156-L159) |
| [aws_internet_gateway_attachment](https://registry.terraform.io/providers/-/aws/latest/docs/resources/internet_gateway_attachment) | One-to-one| [Singular](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_AttachInternetGateway.html) | [Singular](https://github.com/hashicorp/terraform-provider-aws/blob/v5.7.0/internal/service/ec2/vpc_internet_gateway.go#L190-L193) |
| [aws_iot_policy_attachment](https://registry.terraform.io/providers/-/aws/latest/docs/resources/iot_policy_attachment) | One-to-many | [Singular](https://docs.aws.amazon.com/iot/latest/apireference/API_AttachPolicy.html) | [Singular](https://github.com/hashicorp/terraform-provider-aws/blob/v5.7.0/internal/service/iot/policy_attachment.go#L48-L51) |
| [aws_iot_thing_principal_attachment](https://registry.terraform.io/providers/-/aws/latest/docs/resources/iot_thing_principal_attachment) | One-to-many | [Singular](https://docs.aws.amazon.com/iot/latest/apireference/API_AttachThingPrincipal.html) | [Singular](https://github.com/hashicorp/terraform-provider-aws/blob/v5.7.0/internal/service/iot/thing_principal_attachment.go#L49-L52) |
| [aws_lb_target_group_attachment](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/lb_target_group_attachment) | One-to-many | [Plural](https://docs.aws.amazon.com/elasticloadbalancing/latest/APIReference/API_RegisterTargets.html) | [Singular](https://github.com/hashicorp/terraform-provider-aws/blob/v5.7.0/internal/service/elbv2/target_group_attachment.go#L76-L79) |
| [aws_lightsail_disk_attachment](https://registry.terraform.io/providers/-/aws/latest/docs/resources/lightsail_disk_attachment) | One-to-many | [Singular](https://docs.aws.amazon.com/lightsail/2016-11-28/api-reference/API_AttachDisk.html)| [Singular](https://github.com/hashicorp/terraform-provider-aws/blob/v5.7.0/internal/service/lightsail/disk_attachment.go#L57-L61) |
| [aws_lightsail_lb_attachment](https://registry.terraform.io/providers/-/aws/latest/docs/resources/lightsail_lb_attachment) | One-to-many | [Plural](https://docs.aws.amazon.com/lightsail/2016-11-28/api-reference/API_AttachInstancesToLoadBalancer.html) | [Singular](https://github.com/hashicorp/terraform-provider-aws/blob/v5.7.0/internal/service/lightsail/lb_attachment.go#L59-L62) |
| [aws_lightsail_lb_certificate_attachment](https://registry.terraform.io/providers/-/aws/latest/docs/resources/lightsail_lb_certificate_attachment) | One-to-one| [Singular](https://docs.aws.amazon.com/lightsail/2016-11-28/api-reference/API_AttachLoadBalancerTlsCertificate.html)| [Singular](https://github.com/hashicorp/terraform-provider-aws/blob/v5.7.0/internal/service/lightsail/lb_certificate_attachment.go#L60-L63) |
| [aws_lightsail_static_ip_attachment](https://registry.terraform.io/providers/-/aws/latest/docs/resources/lightsail_static_ip_attachment) | One-to-one| [Singular](https://docs.aws.amazon.com/lightsail/2016-11-28/api-reference/API_AttachStaticIp.html)| [Singular](https://github.com/hashicorp/terraform-provider-aws/blob/v5.7.0/internal/service/lightsail/static_ip_attachment.go#L50-L53) |
| [aws_network_interface_attachment](https://registry.terraform.io/providers/-/aws/latest/docs/resources/network_interface_attachment) | One-to-many | [Singular](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_AttachNetworkInterface.html)| [Singular](https://github.com/hashicorp/terraform-provider-aws/blob/v5.7.0/internal/service/ec2/vpc_network_interface.go#L1058-L1062) |
| [aws_network_interface_sg_attachment](https://registry.terraform.io/providers/-/aws/latest/docs/resources/network_interface_sg_attachment) | One-to-many | [Plural](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_ModifyNetworkInterfaceAttribute.html) | [Singular](https://github.com/hashicorp/terraform-provider-aws/blob/v5.7.0/internal/service/ec2/vpc_network_interface_sg_attachment.go#L63-L82) |
| [aws_networkmanager_connect_attachment](https://registry.terraform.io/providers/-/aws/latest/docs/resources/networkmanager_connect_attachment) | One-to-one| [Singular](https://docs.aws.amazon.com/networkmanager/latest/APIReference/API_CreateConnectAttachment.html) | [Singular](https://github.com/hashicorp/terraform-provider-aws/blob/v5.7.0/internal/service/networkmanager/connect_attachment.go#L143-L149) |
| [aws_networkmanager_core_network_policy_attachment](https://registry.terraform.io/providers/-/aws/latest/docs/resources/networkmanager_core_network_policy_attachment) | One-to-one| [Singular](https://docs.aws.amazon.com/networkmanager/latest/APIReference/API_PutCoreNetworkPolicy.html)| [Singular](https://github.com/hashicorp/terraform-provider-aws/blob/v5.7.0/internal/service/networkmanager/core_network.go#L513-L545) |
| [aws_networkmanager_site_to_site_vpn_attachment](https://registry.terraform.io/providers/-/aws/latest/docs/resources/networkmanager_site_to_site_vpn_attachment) | One-to-one| [Singular](https://docs.aws.amazon.com/networkmanager/latest/APIReference/API_CreateSiteToSiteVpnAttachment.html) | [Singular](https://github.com/hashicorp/terraform-provider-aws/blob/v5.7.0/internal/service/networkmanager/site_to_site_vpn_attachment.go#L108-L112) |
| [aws_networkmanager_transit_gateway_registration](https://registry.terraform.io/providers/-/aws/latest/docs/resources/networkmanager_transit_gateway_registration) | One-to-one| [Singular](https://docs.aws.amazon.com/networkmanager/latest/APIReference/API_RegisterTransitGateway.html)| [Singular](https://github.com/hashicorp/terraform-provider-aws/blob/v5.7.0/internal/service/networkmanager/transit_gateway_registration.go#L63-L66) |
| [aws_networkmanager_transit_gateway_route_table_attachment](https://registry.terraform.io/providers/-/aws/latest/docs/resources/networkmanager_transit_gateway_route_table_attachment) | One-to-one| [Singular](https://docs.aws.amazon.com/networkmanager/latest/APIReference/API_CreateTransitGatewayRouteTableAttachment.html)| [Singular](https://github.com/hashicorp/terraform-provider-aws/blob/v5.7.0/internal/service/networkmanager/transit_gateway_route_table_attachment.go#L109-L113) |
| [aws_networkmanager_vpc_attachment](https://registry.terraform.io/providers/-/aws/latest/docs/resources/networkmanager_vpc_attachment) | One-to-one| [Singular](https://docs.aws.amazon.com/networkmanager/latest/APIReference/API_CreateVpcAttachment.html) | [Singular](https://github.com/hashicorp/terraform-provider-aws/blob/v5.7.0/internal/service/networkmanager/vpc_attachment.go#L133-L138) |
| [aws_organizations_policy_attachment](https://registry.terraform.io/providers/-/aws/latest/docs/resources/organizations_policy_attachment) | One-to-many | [Singular](https://docs.aws.amazon.com/organizations/latest/APIReference/API_AttachPolicy.html) | [Singular](https://github.com/hashicorp/terraform-provider-aws/blob/v5.7.0/internal/service/organizations/policy_attachment.go#L61-L64) |
| [aws_quicksight_iam_policy_assignment](https://registry.terraform.io/providers/-/aws/latest/docs/resources/quicksight_iam_policy_assignment) | One-to-many | [Plural](https://docs.aws.amazon.com/quicksight/latest/APIReference/API_CreateIAMPolicyAssignment.html) | [Plural](https://github.com/hashicorp/terraform-provider-aws/blob/v5.7.0/internal/service/quicksight/iam_policy_assignment.go#L131-L145) |
| [aws_security_group](https://registry.terraform.io/providers/-/aws/latest/docs/resources/security_group) (Egress)[^4] | One-to-many | [Plural](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_AuthorizeSecurityGroupEgress.html)| [Plural](https://github.com/hashicorp/terraform-provider-aws/blob/v5.7.0/internal/service/ec2/vpc_security_group.go#L781-L784) |
| [aws_security_group](https://registry.terraform.io/providers/-/aws/latest/docs/resources/security_group) (Ingress)[^4] | One-to-many | [Plural](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_AuthorizeSecurityGroupIngress.html) | [Plural](https://github.com/hashicorp/terraform-provider-aws/blob/v5.7.0/internal/service/ec2/vpc_security_group.go#L788-L791k) |
| [aws_security_group_rule](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/security_group_rule) (Egress)| One-to-many | [Plural](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_AuthorizeSecurityGroupEgress.html)| [Singular](https://github.com/hashicorp/terraform-provider-aws/blob/v5.7.0/internal/service/ec2/vpc_security_group_rule.go#L189-L205) |
| [aws_security_group_rule](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/security_group_rule) (Ingress) | One-to-many | [Plural](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_AuthorizeSecurityGroupIngress.html) | [Singular](https://github.com/hashicorp/terraform-provider-aws/blob/v5.7.0/internal/service/ec2/vpc_security_group_rule.go#L172-L187) |
| [aws_sesv2_dedicated_ip_assignment](https://registry.terraform.io/providers/-/aws/latest/docs/resources/sesv2_dedicated_ip_assignment) | One-to-one| [Singular](https://docs.aws.amazon.com/ses/latest/APIReference-V2/API_PutDedicatedIpInPool.html)| [Singular](https://github.com/hashicorp/terraform-provider-aws/blob/v5.7.0/internal/service/sesv2/dedicated_ip_assignment.go#L66-L69) |
| [aws_ssoadmin_account_assignment](https://registry.terraform.io/providers/-/aws/latest/docs/resources/ssoadmin_account_assignment) | One-to-one| [Singular](https://docs.aws.amazon.com/singlesignon/latest/APIReference/API_CreateAccountAssignment.html) | [Singular](https://github.com/hashicorp/terraform-provider-aws/blob/v5.7.0/internal/service/ssoadmin/account_assignment.go#L106-L113) |
| [aws_ssoadmin_customer_managed_policy_attachment](https://registry.terraform.io/providers/-/aws/latest/docs/resources/ssoadmin_customer_managed_policy_attachment) | One-to-many | [Singular](https://docs.aws.amazon.com/singlesignon/latest/APIReference/API_AttachCustomerManagedPolicyReferenceToPermissionSet.html) | [Singular](https://github.com/hashicorp/terraform-provider-aws/blob/v5.7.0/internal/service/ssoadmin/customer_managed_policy_attachment.go#L90-L94) |
| [aws_ssoadmin_managed_policy_attachment](https://registry.terraform.io/providers/-/aws/latest/docs/resources/ssoadmin_managed_policy_attachment) | One-to-many | [Singular](https://docs.aws.amazon.com/singlesignon/latest/APIReference/API_AttachManagedPolicyToPermissionSet.html)| [Singular](https://github.com/hashicorp/terraform-provider-aws/blob/v5.7.0/internal/service/ssoadmin/managed_policy_attachment.go#L69-L73) |
| [aws_ssoadmin_permissions_boundary_attachment](https://registry.terraform.io/providers/-/aws/latest/docs/resources/ssoadmin_permissions_boundary_attachment) | One-to-one| [Singular](https://docs.aws.amazon.com/singlesignon/latest/APIReference/API_PutPermissionsBoundaryToPermissionSet.html) | [Singular](https://github.com/hashicorp/terraform-provider-aws/blob/v5.7.0/internal/service/ssoadmin/permissions_boundary_attachment.go#L105-L109) |
| [aws_volume_attachment](https://registry.terraform.io/providers/-/aws/latest/docs/resources/volume_attachment) | One-to-many | [Singular](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_AttachVolume.html)| [Singular](https://github.com/hashicorp/terraform-provider-aws/blob/v5.7.0/internal/service/ec2/ebs_volume_attachment.go#L106-L110) |
| [aws_vpc_security_group_egress_rule](https://registry.terraform.io/providers/-/aws/latest/docs/resources/vpc_security_group_egress_rule) | One-to-many | [Plural](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_AuthorizeSecurityGroupEgress.html)| [Singular](https://github.com/hashicorp/terraform-provider-aws/blob/v5.7.0/internal/service/ec2/vpc_security_group_egress_rule.go#L37-L40) |
| [aws_vpc_security_group_ingress_rule](https://registry.terraform.io/providers/-/aws/latest/docs/resources/vpc_security_group_ingress_rule) | One-to-many | [Plural](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_AuthorizeSecurityGroupIngress.html) | [Singular](https://github.com/hashicorp/terraform-provider-aws/blob/v5.7.0/internal/service/ec2/vpc_security_group_ingress_rule.go#L55-L58) |
| [aws_vpclattice_target_group_attachment](https://registry.terraform.io/providers/-/aws/latest/docs/resources/vpclattice_target_group_attachment) | One-to-many | [Plural](https://docs.aws.amazon.com/vpc-lattice/latest/APIReference/API_RegisterTargets.html)| [Singular](https://github.com/hashicorp/terraform-provider-aws/blob/v5.7.0/internal/service/vpclattice/target_group_attachment.go#L81-L84) |
| [aws_vpn_gateway_attachment](https://registry.terraform.io/providers/-/aws/latest/docs/resources/vpn_gateway_attachment) | One-to-one| [Singular](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_AttachVpnGateway.html)| [Singular](https://github.com/hashicorp/terraform-provider-aws/blob/v5.7.0/internal/service/ec2/vpnsite_gateway_attachment.go#L48-L51) |

[^1]: Due to the volume of resources with "rule" in the name (~70), only the prominent security group rule resources were included in the analysis above. While "rule" resources often follow the same relationship-style design, the ~40 examples above provided enough initial data to inform design standards.
[^2]: The structure of this API precludes it from being implemented in a singular fashion.
[^3]: Creates exclusive attachments.
[^4]: Creates exclusive rules.

Of the 44 resources documented above, 29 are of the "one-to-many" form and 17 have "plural" AWS APIs (ie. accept a list of child resources to be attached to a single parent). Of these 17, 13 resources (76%) use a "singular" Terraform implementation, where a list with one item is sent to the Create/Read/Update API, rather than allowing a single resource to manage multiple relationships. Of the remaining 4 with "plural" Terraform implementations, 2 do so to exclusively manage child relationships (`aws_security_group` Ingress/Egress variants), and one requires a "plural" implementation simply because of API limitations.

These metrics indicate a strong historical preference for representing a single API object over aligning the schema to the underlying AWS API. The primary exceptions to this are when exclusive management of all child resources is desired, such as [security group ingress/egress rules](https://registry.terraform.io/providers/-/aws/latest/docs/resources/security_group), or [IAM policy attachments](https://registry.terraform.io/providers/-/aws/latest/docs/resources/iam_policy_attachment).

## Proposal

The best practice for net-new "one-to-many" relationship resources should be to implement singular versions. Feature requests related to changing the singular nature of an existing relationship resource should be avoided unless necessary for the underlying API to function properly.

Variation from this pattern should only be done when:

1. There is a valid use case for a single resource to retain exclusive management of all parent/child relationships.
2. Manipulating the underlying AWS APIs to work with singular relationships is not possible or introduces unnecessary complexity.

### "One-to-many" Design Example

Given a fictional "plural" API `AttachChildren` with a request body like:

```
{
  "ParentId": "string",
  "Children": [
    {
      "ChildId: "string"
    }
  ]
}
```

The corresponding Terraform resource would only represent a single parent/child relationship with a configuration like:

```terraform
resource "aws_parent" "example" {
  name = "foo"
}

resource "aws_child" "example" {
  name = "bar"
}

resource "aws_child_attachment" "example" {
  parent_id = aws_parent.example.id
  child_id  = aws_child.example.id
}
```

## References

* The following feature request and PR initiated the discussion for this analysis:
    * [https://github.com/hashicorp/terraform-provider-aws/issues/9901](https://github.com/hashicorp/terraform-provider-aws/issues/9901)
    * [https://github.com/hashicorp/terraform-provider-aws/pull/32380](https://github.com/hashicorp/terraform-provider-aws/pull/32380)
