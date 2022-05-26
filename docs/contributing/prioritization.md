# How We Prioritize

## Intro

### What this document is

This document describes how we handle prioritization of work from a variety of input sources. Our focus is always to deliver tangible value to the practitioner on a predictable and frequent schedule, and we feel it is important to be transparent in how we weigh input in order to deliver on this goal.

### What this document is not

Due to the variety of input sources, the scale of the provider, and resource constraints, it is impossible to give a hard number on how each of the factors outlined in this document are weighted. Instead, the goal of the document is to give a transparent, but generalized assessment of each of the sources of input so that the community has a better idea of why things are prioritized the way they are. Additional information may be found in the [FAQ](https://github.com/hashicorp/terraform-provider-aws/blob/main/docs/contributing/faq.md#how-do-you-decide-what-gets-merged-for-each-release).

## Prioritization

We prioritze work based on a number of factors, including community feedback, issue/PR reactions, as well as the source of the request. While community feedback is heavily weighted, there are times where other factors take precedence. By their nature, some factors are less visible to the community, and so are outlined here as a way to be as transparent as possible. Each of the sources of input are detailed below.

### Community

Our large community of practitioners are vocal and immensely productive in contributing to the provider codebase. Unfortunately our current team capacity means that we are unable to give every issue or pull request the same level of attention. This means we need to prioritize the issues that provide the most value to the greatest number of practitioners.

We will always focus on the issues which have the most attention. The main rubric we have for assessing community wants is GitHub reactions. In addition to reactions, we look at comments, reactions to comments, and links to additional issues and PRs to help get a more holistic view of where the community stands. We try to ensure that for the issues where we have the most community support, we are responsive to that support and attempt to give timelines where-ever possible.

### Customer

Another source of work that must be weighted are escalations around particular feature requests and bugs from HashiCorp and AWS customers. Escalations typically come via several routes:

- Customer Support
- Sales Engineering
- AWS Solutions Architects contacting us on behalf of their clients.

These reports flow into an internal board and are triaged on a weekly basis to determine whether the escalation request should be prioritized for an upcoming release or added to the backlog to monitor for additional community support. During triage, we verify whether a GitHub issue or PR exists for the request and will create one if it does not exist. In this way, these requests are visible to the community to some degree. An escalation coming from a customer does not necessarily guarantee that it will be prioritized over requests made by the community. Instead, we assess them based on the following rubric:

- Does the issue have considerable community support?
- Does the issue pertain to one of our [Core Services](https://github.com/hashicorp/terraform-provider-aws/blob/main/docs/contributing/core-services.md)?

By weighing these factors, we can make a call to determine whether, and how it is to be prioritized.

### Partner

AWS Service Teams and Partner representatives regularly contact us to discuss upcoming features or new services. This work is often done under an NDA, so usually needs to be done in private. Often the ask is to enable Terraform support or an upcoming feature or service.

As with customer escalations, a request from a partner does not necessarily mean that it will be prioritized over other efforts; capacity restraints require us to prioritize major releases or prefer offerings in line with our [core services](https://github.com/hashicorp/terraform-provider-aws/blob/main/docs/contributing/core-services.md).

### Internal

#### SDK/Core Updates

We endeavor to keep in step with all minor SDK releases, so these are automatically pulled in by our GitHub automation. Major releases normally include breaking changes and usually require us to bump the provider itself to a major version. We plan to make one major version change a year and try to avoid any more than that.

#### Technical Debt

We always include capacity for technical debt work in every iteration, but engineers are free to include minor tech debt work on their own recognizance. For larger items, these are discussed and prioritized in an internal meeting aimed at reviewing technical debt.

#### Adverse User Experience or Security Vulnerabilities

Issues with the provider that provide a poor user experience (bugs, crashes), or involve a threat to security are always prioritized for inclusion. The severity of these will determine how soon they are included for release.
