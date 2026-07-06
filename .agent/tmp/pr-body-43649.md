## Rollback Plan

If a change needs to be reverted, we will publish an updated version of the library.

## Changes to Security Controls

No. This change adds a read-only data source and introduces no changes to access controls, encryption, or logging.

### Description

Adds a new data source `aws_vpclattice_service_network_resource_associations` that lists the resource associations for a VPC Lattice service network. Exactly one of `service_network_identifier` or `resource_configuration_identifier` must be supplied; the data source returns the matching association summaries (ARN, IDs, DNS entries, status, etc.).

Implementation notes:
- Terraform Plugin Framework data source using AutoFlex (`fwflex.Flatten`) to map the API summaries onto the schema.
- Mutual exclusivity of the two identifiers enforced by a single `stringvalidator.ExactlyOneOf`.
- Paginated `List` call via `NewListServiceNetworkResourceAssociationsPaginator`.

### Relations

Closes #43649

### References

- [VPC Lattice `ListServiceNetworkResourceAssociations` API](https://docs.aws.amazon.com/vpc-lattice/latest/APIReference/API_ListServiceNetworkResourceAssociations.html)

### Output from Acceptance Testing

<!-- DRAFT: acceptance test output pending; will be added before requesting review. -->

```console
% make testacc TESTS=TestAccVPCLatticeServiceNetworkResourceAssociationsDataSource_ PKG=vpclattice

... (pending)
```
