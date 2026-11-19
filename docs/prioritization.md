<!-- Copyright IBM Corp. 2014, 2026 -->
<!-- SPDX-License-Identifier: MPL-2.0 -->

# How We Prioritize

## Overview

### What This Document Is

This document describes how we prioritize work from a variety of input sources. Our focus is to deliver tangible value to practitioners on a predictable, frequent schedule. Being transparent about how we weigh input helps us meet that goal.

### What This Document Is Not

This document does not give exact weights for each prioritization factor. The variety of input sources, the scale of the provider, and resource constraints make that impossible. Instead, it offers a transparent, generalized assessment of each input source so the community can better understand why work is prioritized the way it is. See the [FAQ](faq.md#how-do-you-decide-what-gets-merged-for-each-release) for more information.

## Sources of Work

We prioritize work based on a range of factors, including community feedback (such as issue and pull request reactions) and the source of the request. While community feedback is heavily weighted, other factors sometimes take precedence. Security issues and serious regressions are always prioritized. For everything else, community support is the primary signal, and customer, partner, and internal needs are weighed against it. Some of these sources are less visible to the community, so we outline each one below.

### Community

Our large community of practitioners is vocal and contributes heavily to the provider codebase. Because team capacity is limited, we cannot give every issue or pull request the same level of attention, so we prioritize the issues that provide the most value to the most practitioners.

Community attention is a primary input, and GitHub reactions are the main rubric we use to assess it. In addition to reactions, we look at comments, reactions to comments, and links to additional issues and pull requests to help get a more holistic view of where the community stands. For the items with the most community support, we aim to be responsive and to share timelines where we can.

### Customer

We also receive escalations for specific feature requests and bugs from HashiCorp/IBM and AWS customers. Escalations typically come via several routes:

- Customer Support
- Sales Engineering
- AWS Solutions Architects contacting us on behalf of their clients

These reports flow into an internal board and are triaged weekly to decide whether to prioritize the escalation for an upcoming release or add it to the backlog until it gathers more community support. If the request has no corresponding GitHub issue or pull request, we create one, which gives the community some visibility.

A customer escalation does not guarantee priority over community requests. Instead, we assess each escalation with the following questions:

- Does the request have considerable community support?
- Does the request pertain to one of our [core services](core-services.md)?

We weigh the answers to these questions to decide whether, and how, to prioritize the request.

### Partner

AWS service teams and partner representatives regularly contact us to discuss upcoming features or new services. This work is often done under NDA, so it usually has to happen in private.

As with customer escalations, a partner request does not guarantee priority. Because of capacity constraints, we favor major service launches and offerings that align with our [core services](core-services.md).

### Internal

#### SDK Updates

We keep pace with minor releases of the AWS SDK for Go v2, which our GitHub automation pulls in automatically. Major SDK releases can include breaking changes, which may require a major provider version bump. We plan for at most one major provider version per year.

#### Technical Debt

Every planning cycle includes capacity for technical debt work, and engineers may also take on minor technical debt work at their own discretion. Larger items are discussed and prioritized in an internal meeting specifically aimed at reviewing and addressing technical debt.

#### Regressions, Crashes, and Security Issues

Issues that degrade the user experience (such as regressions and crashes) or pose a security threat are always prioritized. The severity determines how quickly we release a fix.
