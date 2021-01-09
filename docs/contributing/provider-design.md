# Provider Design

_Please Note: This documentation is intended for Terraform AWS Provider code developers. Typical operators writing and applying Terraform configurations do not need to read or understand this material._

The Terraform AWS Provider follows the guidelines established in the [HashiCorp Provider Design Principles](https://www.terraform.io/docs/extend/hashicorp-provider-design-principles.html). That general documentation provides many high level design points after years of experience with Terraform's overall design and implementation of infrastructure as code concepts. Certain sections below will expand on specific design details between that documentation and this provider, while others will capture other pertinent information which may not be covered there. Other pages of the contributing guide cover implementation details such as code, testing, and documentation specifics.

- [API and SDK Boundary](#api-and-sdk-boundary)
- [Infrastructure as Code Suitability](#infrastructure-as-code-suitability)
- [Resource Type Considerations](#resource-type-considerations)
    - [Authorization and Acceptance Resources](#authorization-and-acceptance-resources)
    - [Cross-Service Functionality](#cross-service-functionality)
    - [IAM Resource-Based Policy Resources](#iam-resource-based-policy-resources)
    - [Managing Resource Running State](#managing-resource-running-state)
    - [Task Execution and Waiter Resources](#task-execution-and-waiter-resources)
- [Other Considerations](#other-considerations)
    - [AWS Credential Exfiltration](#aws-credential-exfiltration)

## API and SDK Boundary

The provider implements support for the [AWS](https://aws.amazon.com/) service APIs using the [AWS Go SDK](https://aws.amazon.com/sdk-for-go/) and its available functionality. Overall these operations are for the lifecycle management of AWS components (such as creating, describing, updating, and deleting a database) and not for handling functionality within those components (such as executing a query on that database). It is expected that other Terraform Providers will be implemented with the appropriate API/SDK for those other systems and protocols, which are maintained with separate domain expertise in that area.

Some examples of functionality that is not expected in this provider:

* Raw HTTP(S) handling. See the [Terraform HTTP Provider](https://registry.terraform.io/providers/hashicorp/http/latest) and [Terraform TLS Provider](https://registry.terraform.io/providers/hashicorp/tls/latest) instead.
* Kubernetes resource management beyond the EKS service APIs. See the [Terraform Kubernetes Provider](https://registry.terraform.io/providers/hashicorp/kubernetes/latest) instead.
* Active Directory or other protocol clients. See the [Terraform Active Directory Provider](https://registry.terraform.io/providers/hashicorp/ad/latest/docs) and other available provider instead.
* Functionality that requires additional software beyond the Terraform AWS Provider to be installed on the host executing Terraform. This currently includes the AWS CLI. See the [Terraform External Provider](https://registry.terraform.io/providers/hashicorp/external/latest) and other available providers instead.

## Infrastructure as Code Suitability

While a goal of the provider is to gain as much resource coverage as possible, from a design perspective, it is necessary to be pragmatic that there are APIs or portions of APIs that may not be wholly compatible with infrastructure as code concepts and managing them with that methodology. In these cases, it is recommended to open an AWS Support case (potentially encouraging others to do the same) to either fix or enhance the AWS service API so components are more self-contained and compatible with these concepts. Sometimes these types of AWS Support cases can also provide some good insight into the AWS service expectations or API behaviors that may not be well documented.

## Resource Type Considerations

The general heuristic of when to create or not create a separate Terraform resource is if the AWS service API implements create, read, and delete operations for a component. Terraform resources work best as the smallest blocks of infrastructure, where configurations can build abstractions on top of these such as [Terraform Modules](https://www.terraform.io/docs/modules/). However, not all AWS service API functionality falls cleanly into components to be managed with those types of operations. It may be necessary to consider other patterns of mapping API operations to Terraform resources and operations. This section highlights these design recommendations, such as when to consider implemantations within the same Terraform resource or as a separate Terraform resource.

Please note: the overall design and implementation across all AWS functionality is federated. This means that individual services may implement concepts and use terminology differently. As such, this guide may not be fully exhaustive in all situations. The goal is to provide enough hints about these concepts and towards specific terminology to point contributors in the correct direction, especially when researching prior implementations.

### Authorization and Acceptance Resources

Certain AWS services use an authorization-acceptance model for cross-account associations or access. Some examples of these can be found with:

* Direct Connect Association Proposals
* GuardDuty Member Invitations
* RAM Resource Share Associations
* Route 53 VPC Associations
* Security Hub Member Invitations

This type of API model either implements a form of an invitation/proposal identifier that can be used to accept the authorization, otherwise it may just be a matter of providing the desired AWS Account identifier as the target of an authorization. In the latter case, the acceptance is implicit when creating the other half of the association.

To model the separate invitation/proposal in Terraform resources:

* Typically the authorization side has separate and sufficient creation and read API functionality to create an "invite" or "proposal" resource.
* If the acceptance side has separate and sufficient accept, read, and reject API functionality, then an "accepter" resource may be created, where the operations are mapped as:
    * Create: Accepts the invitation/proposal.
    * Read: Reads the invitation/proposal to determine status. If not found, then it should fallback to reading the API resource associated with the invitation/proposal/association. As evidenced in some previous API implementations, the invitation/proposal may be temporary and removed from the API after an indeterminate amount of time which may not be documented or easily discoverable from testing.
    * Delete: Rejects or otherwise deletes the invitation/proposal.

Otherwise, to model an implicit acceptance in Terraform resources:

* If the authorization side has separate and sufficient creation and read API functionality, an "authorization" resource is created.
* An "association" resource is created, where the operations are mapped as:
    * Create: Create the association
    * Read: Reads the association
    * Delete: Deletes the association

### Cross-Service Functionality

Many AWS service APIs build on top of other AWS services. Some examples of these include:

* EKS Node Groups managing Auto Scaling Groups
* Lambda Functions managing EC2 ENIs
* Transfer Servers managing EC2 VPC Endpoints

In certain of these cross-service API implementations, there can be lacking management or description capabilities of the other service functionality, which can make the Terraform resource implementation seem incomplete or unsuccessful in end-to-end configurations. Given the overall “resources should represent a single API object” goal from the [HashiCorp Provider Design Principles](https://www.terraform.io/docs/extend/hashicorp-provider-design-principles.html), it is expected that a single resource in this provider will only communicate with a single AWS service API in a self-contained manner. As such, cross-service resources will not be implementated.

In particular, some of the rationale behind this design decision is:

* Unexpected IAM permissions being necessary for the resource. In restrictive environments, these permissions may not be available or acceptable for security purposes.
* Unexpected CloudTrail logs being generated for the resource.
* Unexpected API endpoints configuration being necessary for those using custom endpoints, such as VPC endpoints.
* Unexpected changes to the AWS service internals. Given that these cross-service implementations require being based on internal details the current AWS service implementation, these details can change over time and may not be considered as a breaking change by the AWS service API during an API upgrade.

As a poignant real world example of the last point, the provider was relying on the description of Lambda Function created EC2 ENIs for the purposes of trying to delete lingering ENIs due to a common misconfiguration. This functionality was helpful for operators as the issue was hard to diagnose due to a generic EC2 API error message. Years after the implementation, the Lambda API was updated and immediately many operators started reporting Terraform executions were failing. Downgrading the provider could not help, many configurations depended on resources and behaviors in recent versions. For other environments that were running many versions behind, forcing an immediate provider upgrade with the fix could also have other unrelated or unexpected changes. In the end, both HashiCorp and AWS needed to perform a large scale outreach about upgrading or fixing the misconfiguration, causing lost time for many operators and the provider maintainers.

### IAM Resource-Based Policy Resources

Many AWS resources support the concept of an [IAM resource-based policy](https://docs.aws.amazon.com/IAM/latest/UserGuide/access_policies_identity-vs-resource.html) for specifying IAM permissions associated the resource. Some examples of these include:

* ECR Repository Policies
* EFS File System Policies
* SNS Topic Policies

When implementing this support in the Terraform AWS Provider, it is preferred to create a separate resource for a few reasons:

* Many of these policies require the Amazon Resource Name (ARN) of the resource in the policy itself. It is difficult to workaround this requirement with custom difference handling within a self-contained resource.
* Sometimes policies between two resources need to be written where they cross-reference each other resource's ARN within each policy. Without a separate resource, this introduces a configuration cycle.
* Splitting the resources allows operators to logically split their infrastructure on purely operational and security boundaries with separate configurations/modules.
* Splitting the resources prevents any separate policy API calls from needing to be permitted in the main resource in environments with restrictive IAM permissions, which can be undesirable.

In rare cases, it may be necessary to implement the policy handling within the associated resource itself. This should only be reserved for cases where the policy is _required_ during resource creation.

### Managing Resource Running State

Certain AWS components allow the ability to start/stop or enable/disable the resource within the API. Some examples of these include:

* Batch Job Queues
* CloudFront Distributions
* RDS DB Event Subscriptions

It is recommended that this ability be implemented within the parent Terraform resource instead of creating a separate resource.  Trying to manage updates that require the resource to be in a specific running state becomes impractical for operators to manage within a Terraform configuration. This is generally a consistency and future-proofing design decision, even if there is no problematic resource update use case within the current API.

### Task Execution and Waiter Resources

Certain AWS functionality operates with the execution of individual tasks or external validation processes that have a completion status. Some examples of these include:

* ACM Certificate validation
* EC2 AMI copying
* RDS DB Cluster Snapshot management

The general expectation in these cases is that the AWS service APIs provide sufficient lifecycle management to start the execution/validation and read the status of that operation. At the end of the operation, the result may be something that can be fully managed on its own with update and delete support (e.g. with EC2 AMI copying, a copy of the AMI is a manageable AMI). In these cases, the preference is to create a separate Terraform resource, rather than implement the functionality within the parent Terraform resource. This recommendation is even at the expense of duplicated resource and schema handling.

One related consideration is the ability to manage the running state of a resource (e.g. start/stop it). See the [Managing Resource Running State section](#managing-resource-running-state) for specific details.

## Other Considerations

### AWS Credential Exfiltration

In the security-minded interest of the community, the provider will not implement data sources that will give operators the ability to reference or export the AWS credentials of the running provider. While there are valid use cases of this information, such as using those AWS credentials to execute AWS CLI calls as part of the same Terraform configuration, it also becomes possible for those credentials to be discovered and used completely outside of the current Terraform execution. Some of the concerns around this topic include:

* The values can potentially be visible in Terraform user interface output or logging where anyone with access to those can see and use them.
* The values must be stored in the Terraform state in plaintext, due to how Terraform operates. Anyone with access or another Terraform configuration that references the state can see and use them.
* Any new functionality related to this, while opt-in to implement, is also opt-in to prevent via security controls or policies. Since this introduces a weak by default security posture, organizations wishing to implement those controls may require advance notice before the release or may never be able to update.
