# Provider Design

_Please Note: This documentation is intended for Terraform AWS Provider code developers. Typical operators writing and applying Terraform configurations do not need to read or understand this material._

The Terraform AWS Provider follows the guidelines established in the [HashiCorp Provider Design Principles](https://www.terraform.io/docs/extend/hashicorp-provider-design-principles.html). That general documentation provides many high-level design points gleaned from years of experience with Terraform's design and implementation concepts. Sections below will expand on specific design details between that documentation and this provider, while others will capture other pertinent information that may not be covered there. Other pages of the contributing guide cover implementation details such as code, testing, and documentation specifics.

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

The AWS provider implements support for the [AWS](https://aws.amazon.com/) service APIs using the [AWS Go SDK](https://aws.amazon.com/sdk-for-go/). The API and SDK limits extend to the provider. In general, SDK operations manage the lifecycle of AWS components, such as creating, describing, updating, and deleting a database. Operations do not usually handle functionality within those components, such as executing a query on a database. If you are interested in other APIs/SDKs, we invite you to view the many Terraform Providers available, as each has a community of domain expertise.

Some examples of functionality that is not expected in this provider:

* Raw HTTP(S) handling. See the [Terraform HTTP Provider](https://registry.terraform.io/providers/hashicorp/http/latest) and [Terraform TLS Provider](https://registry.terraform.io/providers/hashicorp/tls/latest) instead.
* Kubernetes resource management beyond the EKS service APIs. See the [Terraform Kubernetes Provider](https://registry.terraform.io/providers/hashicorp/kubernetes/latest) instead.
* Active Directory or other protocol clients. See the [Terraform Active Directory Provider](https://registry.terraform.io/providers/hashicorp/ad/latest/docs) and other available provider instead.
* Functionality that requires additional software beyond the Terraform AWS Provider to be installed on the host executing Terraform. This currently includes the AWS CLI. See the [Terraform External Provider](https://registry.terraform.io/providers/hashicorp/external/latest) and other available providers instead.

## Infrastructure as Code Suitability

The provider maintainers' design goal is to cover as much of the AWS API as pragmatically possible. However, not every aspect of the API is compatible with an infrastructure-as-code (IaC) conception. If such limits affect you, we recommend that you open an AWS Support case and encourage others to do the same. Request that AWS components be made more self-contained and compatible with IaC. These AWS Support cases can also yield insights into the AWS service and API that are not well documented.

## Resource Type Considerations

The general heuristic of whether or not to create a separate Terraform resource is if the AWS service API implements create, read, and delete operations for a component. Terraform resources work best as the smallest blocks of infrastructure, where configurations can build abstractions on top of these such as [Terraform Modules](https://www.terraform.io/docs/modules/). However, not all AWS service API functionality falls cleanly into components to be managed with those types of operations. It may be necessary to consider other patterns of mapping API operations to Terraform resources and operations. This section highlights these design recommendations, such as when to consider implemantations within the same Terraform resource or as a separate Terraform resource.

Please note: the overall design and implementation across all AWS functionality is federated. This means that individual services may implement concepts and use terminology differently. As such, this guide may not be fully exhaustive in all situations. The goal is to provide enough hints about these concepts and towards specific terminology to point contributors in the correct direction, especially when researching prior implementations.

### Authorization and Acceptance Resources

Certain AWS services use an authorization-acceptance model for cross-account associations or access. Some examples of these can be found with:

* Direct Connect Association Proposals
* GuardDuty Member Invitations
* RAM Resource Share Associations
* Route 53 VPC Associations
* Security Hub Member Invitations

This type of API model either implements a form of an invitation/proposal identifier that can be used to accept the authorization, otherwise it may just be a matter of providing the desired AWS Account identifier as the target of an authorization. In the latter case, the acceptance is implicit when creating the other half of the association.

To model creating an association using an invitation or proposal, follow these guidelines.

* Follow the naming in the AWS service API to determine whether to use the term "invitation" or "proposal."
* For the originating account, create an "invitation" or "proposal" resource. Make sure that the AWS service API has operations for creating and reading invitations. 
* For the responding account, create an "accepter" resource. Ensure that the API has operations for accepting, reading, and rejecting invitations in the responding account. Map the operations as follows:
	* Create: Accepts the invitation.
	* Read: Reads the invitation to determine its status. Note that in some APIs, invitations expire and disappear, complicating associations. If a resource does not find an invitation, the developer should implement a fall back to read the API resource associated with the invitation/proposal.
	* Delete: Rejects or otherwise deletes the invitation.

To model the second type of association, implicit associations, create an "association" resource and, optionally, an "authorization" resource. Map create, read, and delete to the corresponding operations in the AWS service API.

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

The AWS API provides the ability to start, stop, enable, or disable some AWS components. Some examples include:

* Batch Job Queues
* CloudFront Distributions
* RDS DB Event Subscriptions

In this situation, provider developers should implement this ability within the resource instead of creating a separate resource. Since a practitioner cannot practically manage interaction with a resource's states in Terraform's declarative configuration, developers should implement the state management in the resource. This design provides consistency and future-proofing even where updating a resource in the current API is not problematic.

### Task Execution and Waiter Resources

Some AWS operations are asynchronous. Terraform requests that AWS perform a task. Initially, AWS only notifies Terraform that it received the request. Terraform then requests the status while awaiting completion. Examples of this include:

* ACM Certificate validation
* EC2 AMI copying
* RDS DB Cluster Snapshot management

In this situation, provider developers should create a separate resource representing the task, assuming that the AWS service API provides operations to start the task and read its status. Adding the task functionality to the parent resource muddies its infrastructure-management purpose. The maintainers prefer this approach even though there is some duplication of an existing resource. For example, the provider has a resource for copying an EC2 AMI in addition to the EC2 AMI resource itself. This modularity allows practitioners to manage the result of the task resource with another resource. 

For a related consideration, see the [Managing Resource Running State section](#managing-resource-running-state).

## Other Considerations

### AWS Credential Exfiltration

In the security-minded interest of the community, the provider will not implement data sources that will give operators the ability to reference or export the AWS credentials of the running provider. While there are valid use cases of this information, such as using those AWS credentials to execute AWS CLI calls as part of the same Terraform configuration, it also becomes possible for those credentials to be discovered and used completely outside of the current Terraform execution. Some of the concerns around this topic include:

* The values can potentially be visible in Terraform user interface output or logging where anyone with access to those can see and use them.
* The values must be stored in the Terraform state in plaintext, due to how Terraform operates. Anyone with access or another Terraform configuration that references the state can see and use them.
* Any new functionality related to this, while opt-in to implement, is also opt-in to prevent via security controls or policies. Since this introduces a weak by default security posture, organizations wishing to implement those controls may require advance notice before the release or may never be able to update.
