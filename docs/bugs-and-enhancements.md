# Making Small Changes to Existing Resources

Most contributions to the provider will take the form of small additions or bug-fixes on existing resources/data sources. In this case the existing resource will give you the best guidance on how the change should be structured, but we require the following to allow the change to be merged:

- __Acceptance test coverage of new behavior__: Existing resources each
   have a set of [acceptance tests](running-and-writing-acceptance-tests.md) covering their functionality.
   These tests should exercise all the behavior of the resource. Whether you are
   adding something or fixing a bug, the idea is to have an acceptance test that
   fails if your code were to be removed. Sometimes it is sufficient to
   "enhance" an existing test by adding an assertion or tweaking the config
   that is used, but it's often better to add a new test. You can copy/paste an
   existing test and follow the conventions you see there, modifying the test
   to exercise the behavior of your code.
- __Documentation updates__: If your code makes any changes that need to
   be documented, you should include those [documentation changes](documentation-changes.md)
   in the same PR. This includes things like new resource attributes or changes in default values.
- __Well-formed Code__: Do your best to follow existing conventions you
   see in the codebase, and ensure your code is formatted with `go fmt`.
   The PR reviewers can help out on this front, and may provide comments with
   suggestions on how to improve the code.
- __Dependency updates__: Create a separate PR if you are updating dependencies.
   This is to avoid conflicts as version updates tend to be fast-
   moving targets. We will plan to merge the PR with this change first.
- __Changelog entry__: Assuming the code change affects Terraform operators,
   the relevant PR ought to include a user-facing [changelog entry](changelog-process.md)
   describing the new behavior.
