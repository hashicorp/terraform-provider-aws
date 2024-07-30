from collections import defaultdict
import json
import re
import os
import subprocess
import tempfile
from typing import TypedDict
import git
import shutil

AWS_PROVIDER = "https://github.com/hashicorp/terraform-provider-aws"
PLUGIN_SDK = "github.com/hashicorp/terraform-plugin-sdk/v2"
PATCHED_SDK = "github.com/cloutierMat/terraform-plugin-sdk"
PATCH_TAG_FILE = "patched_tags.json"
TAGS_REGEX = r"v\d+\.\d+\.\d+"
a = defaultdict(lambda: 0)


class Tag(TypedDict):
    tag: str
    sdk_version: str | None
    sdk_patched: bool
    terraform_patched: bool


def find_all_tags(repo: git.Repo) -> list[Tag]:
    return [
        Tag(
            tag=str(tag),
            sdk_version=None,
            sdk_patched=False,
            terraform_patched=False,
        )
        for tag in sorted(repo.tags, key=lambda t: t.commit.committed_datetime)
    ]


def enrich_tag(tag: Tag, repo: git.Repo) -> Tag:
    print(f"checking_out {tag['tag']}")
    repo.git.checkout(f"tags/{tag['tag']}", force=True)
    process = subprocess.Popen(
        f"go list -m {PLUGIN_SDK}".split(),
        cwd=aws_repo.common_dir,
        stdout=subprocess.PIPE,
    )
    plugin_sdk, _ = process.communicate()
    plugin_sdk = plugin_sdk.decode()

    if not plugin_sdk:
        return tag

    if PLUGIN_SDK in plugin_sdk:
        found = re.search(TAGS_REGEX, plugin_sdk) or [None]
        tag["sdk_version"] = found[0]
    if PATCHED_SDK in plugin_sdk:
        tag["sdk_patched"] = True

    return tag

def is_git_repo(path):
    try:
        _ = git.Repo(path).git_dir
        return True
    except git.exc.InvalidGitRepositoryError:
        return False

def list_all_tags_for_remote_git_repo(repo_url):
    """
    Given a repository URL, list all tags for that repository
    without cloning it.

    This function use "git ls-remote", so the
    "git" command line program must be available.
    """
    # Run the 'git' command to fetch and list remote tags
    result = subprocess.run([
        "git", "ls-remote", "--tags", repo_url
    ], stdout=subprocess.PIPE, text=True)

    # Process the output to extract tag names
    output_lines = result.stdout.splitlines()
    tags = [
        line.split("refs/tags/")[-1] for line in output_lines
        if "refs/tags/" in line and "^{}" not in line
    ]

    return tags

if __name__ == "__main__":
    patched_tag = {}
    if os.path.exists(PATCH_TAG_FILE):
        with open(PATCH_TAG_FILE, "r") as file:
            patched_tag = json.load(file) or {}

    temp_dir = tempfile.mkdtemp(prefix="temp_aws_provider_")
    # temp_dir = "/var/folders/jm/jdck4c8n6wz1b79b7pb6lmpw0000gn/T/temp_aws_provider_t_d6o08w"
    try:
        if is_git_repo(temp_dir):
            aws_repo = git.Repo(temp_dir)
        else:
            print(f"Cloning into {temp_dir}...")
            aws_repo = git.Repo.clone_from(url=AWS_PROVIDER, to_path=temp_dir, depth=1)
            print(f"Finished cloning {temp_dir}.")

        aws_repo.remotes.origin.fetch(refspec="refs/tags/*:refs/tags/*")

        unlisted = [
            tag for tag in find_all_tags(aws_repo) if tag["tag"] not in patched_tag
        ]
        added = {tag["tag"]: enrich_tag(tag, aws_repo) for tag in unlisted}
        patched_tag.update(added)
    finally:
        pass
        # shutil.rmtree(temp_dir)

    with open(PATCH_TAG_FILE, "+r") as file:
        json.dump({k: patched_tag[k] for k in sorted(patched_tag)}, file, indent=4)