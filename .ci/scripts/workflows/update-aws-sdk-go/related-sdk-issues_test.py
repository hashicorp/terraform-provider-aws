# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

import json
from pathlib import Path
import subprocess
import tempfile
import unittest


PREFIX = "github.com/aws/aws-sdk-go-v2"
MODULE = PREFIX + "/service/networkmanager"


def issue(number, body, **kwargs):
    return {"number": number, "title": "", "body": body, "state": "open", **kwargs}


class RelatedSDKIssuesTest(unittest.TestCase):
    def setUp(self):
        self.before = {"Require": [
            {"Path": PREFIX, "Version": "v1.47.4"},
            {"Path": MODULE, "Version": "v1.50.0"},
        ]}
        self.after = {"Require": [
            {"Path": PREFIX, "Version": "v1.48.0"},
            {"Path": MODULE, "Version": "v1.53.0"},
        ]}
        self.tags = [
            {"ref": "refs/tags/release-2026-10-22", "sha": "a" * 40},
            {"ref": "refs/tags/release-2026-11-16", "sha": "b" * 40},
            {"ref": "refs/tags/release-2026-11-17", "sha": "c" * 40},
            {"ref": "refs/tags/v1.47.4", "sha": "a" * 40},
            {"ref": "refs/tags/v1.48.0", "sha": "b" * 40},
            {"ref": "refs/tags/service/networkmanager/v1.50.0", "sha": "a" * 40},
            {"ref": "refs/tags/service/networkmanager/v1.53.0", "sha": "c" * 40},
        ]

    def relations(self, issues):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            for name, value in (("before", self.before), ("after", self.after), ("tags", self.tags)):
                (root / name).write_text(json.dumps(value))
            (root / "issues").write_text("\n".join(json.dumps(i) for i in issues))
            return subprocess.check_output([
                "jq", "-sr",
                "--slurpfile", "before", str(root / "before"),
                "--slurpfile", "after", str(root / "after"),
                "--slurpfile", "tags", str(root / "tags"),
                "-f", str(Path(__file__).with_name("related-sdk-issues.jq")),
                str(root / "issues"),
            ], text=True).strip()

    def test_module_upgrade_interval_and_identity(self):
        self.assertEqual(self.relations([
            issue(1, f"`{MODULE}` v1.50.0"),
            issue(2, f"`{MODULE}` v1.52.0"),
            issue(3, f"`{MODULE}` v1.53.0"),
            issue(4, f"`{MODULE}` v1.54.0"),
            issue(5, f"`{PREFIX}/service/other` v1.52.0"),
        ]), "Relates #2.\nRelates #3.")

    def test_all_explicit_requirements_must_be_satisfied(self):
        self.assertEqual(self.relations([
            issue(1, f"`{MODULE}` v1.52.0; `{PREFIX}` v1.47.4"),
            issue(2, f"`{MODULE}` v1.52.0; `{PREFIX}` v1.49.0"),
            issue(3, f"`{MODULE}` v1.50.0; `{PREFIX}` v1.47.4"),
        ]), "Relates #1.")

    def test_release_dates_and_links_include_intermediate_releases(self):
        self.assertEqual(self.relations([
            issue(1, "Requires AWS SDK for Go v2 [Release 2026-11-16](https://example.com)."),
            issue(2, "https://github.com/aws/aws-sdk-go-v2/releases/tag/release-2026-11-16"),
            issue(3, "Requires AWS SDK for Go v2 Release 2026-10-22"),
            issue(4, "Requires AWS SDK for Go v2 Release 2026-11-18"),
            issue(5, "An unrelated release on 2026-11-16"),
        ]), "Relates #1.\nRelates #2.")

    def test_example_issues_and_commit_links(self):
        for tag in self.tags:
            if tag["sha"] == "a" * 40:
                tag["sha"] = "4aaad745c73dd432ea4d2188a350401437dd42ae"
        self.assertEqual(self.relations([
            issue(51065, "Requires AWS SDK for Go v2 [Release 2026-11-16]"
                  "(https://github.com/aws/aws-sdk-go-v2/releases/tag/release-2026-11-16) "
                  f"(`{MODULE}` v1.52.0)."),
            issue(50613, "Requires AWS SDK for Go v2 [Release 2026-10-22]"
                  "(https://github.com/aws/aws-sdk-go-v2/commit/4aaad745c73dd432ea4d2188a350401437dd42ae)."),
            issue(6, "https://github.com/aws/aws-sdk-go-v2/commit/" + "b" * 40),
            issue(7, "https://github.com/aws/aws-sdk-go-v2/commit/bbbbbbb"),
            issue(8, "https://github.com/somewhere/else/commit/" + "b" * 40),
        ]), "Relates #6.\nRelates #7.\nRelates #51065.")

    def test_module_requirement_takes_precedence_over_release(self):
        self.assertEqual(self.relations([
            issue(1, f"`{MODULE}` v1.50.0; Requires AWS SDK for Go v2 Release 2026-11-16"),
            issue(2, f"`{MODULE}` v1.54.0; Requires AWS SDK for Go v2 Release 2026-11-16"),
        ]), "No linked issue.")

    def test_missing_tags_do_not_guess_a_generic_release(self):
        self.tags = []
        self.assertEqual(self.relations([
            issue(1, "Requires AWS SDK for Go v2 Release 2026-11-16"),
            issue(2, f"`{MODULE}` v1.52.0"),
        ]), "Relates #2.")

    def test_only_unique_open_issues_and_title_references(self):
        self.assertEqual(self.relations([
            issue(9, None, title=f"Requires {MODULE} v1.52.0"),
            issue(9, f"`{MODULE}` v1.52.0"),
            issue(10, f"`{MODULE}` v1.52.0", state="closed"),
            issue(11, f"`{MODULE}` v1.52.0", pull_request={}),
            issue(12, None),
        ]), "Relates #9.")

    def test_already_satisfied_and_prerelease_requirements(self):
        self.assertEqual(self.relations([
            issue(1, f"`{MODULE}` v1.52.0-rc.1"),
        ]), "No linked issue.")
        self.after = self.before
        self.assertEqual(self.relations([
            issue(1, f"`{MODULE}` v1.52.0"),
            issue(2, "Requires AWS SDK for Go v2 Release 2026-11-16"),
        ]), "No linked issue.")

    def test_numeric_release_revisions(self):
        self.tags[0]["ref"] = "refs/tags/release-2026-11-16.2"
        self.tags[1]["ref"] = "refs/tags/release-2026-11-16.10"
        self.tags = [t for t in self.tags if t["ref"] != "refs/tags/release-2026-11-17"]
        for tag in self.tags:
            if tag["sha"] == "c" * 40:
                tag["sha"] = "b" * 40
        self.assertEqual(self.relations([
            issue(1, "Requires AWS SDK for Go v2 Release 2026-11-16.3"),
            issue(2, "Requires AWS SDK for Go v2 Release 2026-11-16.10"),
            issue(3, "Requires AWS SDK for Go v2 Release 2026-11-16.11"),
        ]), "Relates #1.\nRelates #2.")

    def test_new_module_requirement_was_not_previously_satisfied(self):
        self.before["Require"] = [m for m in self.before["Require"] if m["Path"] != MODULE]
        self.assertEqual(self.relations([
            issue(1, f"`{MODULE}` v1.49.0."),
        ]), "Relates #1.")

    def test_numeric_module_version_ordering(self):
        self.before["Require"][1]["Version"] = "v1.9.9"
        self.after["Require"][1]["Version"] = "v1.10.1"
        self.assertEqual(self.relations([
            issue(1, f"`{MODULE}` v1.10.0"),
            issue(2, f"`{MODULE}` v1.9.9"),
            issue(3, f"`{MODULE}` v1.10.2"),
        ]), "Relates #1.")


if __name__ == "__main__":
    unittest.main()
