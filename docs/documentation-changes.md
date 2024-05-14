# End User Documentation Changes

All practitioner-focused documentation is found in the `/website` folder of the repository.

```
├── website/docs
    ├── r/                     # Documentation for resources
    ├── d/                     # Documentation for data sources
    ├── guides/                # Long format guides for provider level configuration or provider upgrades.
    ├── cdktf/                 # Documentation for CDKTF generated in other programming languages
    └── index.html.markdown    # Home page and all provider level documentation.
└── examples/                  # Large example configurations
```

!!!note
    The CDKTF documentation is generated based on resource and data source documentation. Files in the `cdktf/` folder should not be edited directly.

For any documentation change please raise a pull request including and adhering to the following:

- __Reasoning for Change__: Documentation updates should include an explanation for why the update is needed. If the change is a correction that aligns with AWS behavior, please include a link to the AWS Documentation in the PR.
- __Prefer AWS Documentation__: Documentation about AWS service features and valid argument values that are likely to update over time should link to AWS service user guides and API references where possible.
- __Large Example Configurations__: Example Terraform configuration that includes multiple resource definitions should be added to the repository `examples` directory instead of an individual resource documentation page. Each directory under `examples` should be self-contained to call `terraform apply` without special configuration.
- __Avoid Terraform Configuration Language Features__: Individual resource documentation pages and examples should refrain from highlighting particular Terraform configuration language syntax workarounds or features such as `variable`, `local`, `count`, and built-in functions.
