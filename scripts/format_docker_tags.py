#!/usr/bin/env python3
#
# Outputs a list of Docker tags appropriate for the docker file.
# ./format_docker_tags.py <"unstable" | a SemVer version string> [suffix]
import sys

version = sys.argv[1]
version_numbers = version.split(".")
version_tags = [ ".".join(version_numbers[:i+1]) for i in range(0, len(version_numbers)) ]

# Latest is also a version_tag
if version != "unstable":
    version_tags.append("latest")

suffix = ""
if len(sys.argv) == 3:
    suffix = sys.argv[2]

base = ["natsukagami/kjudge", "ghcr.io/natsukagami/kjudge"]

for b in base:
    for v in version_tags:
        print(b + ":" + v + suffix)
