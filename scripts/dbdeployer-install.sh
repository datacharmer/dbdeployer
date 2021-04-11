#!/usr/bin/env bash
# DBDeployer - The MySQL Sandbox
# Copyright © 2006-2020 Giuseppe Maxia
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# dbdeployer installer version 1.0.0 - Released 2020-12-29
# Usage:
# curl -s https://raw.githubusercontent.com/datacharmer/dbdeployer/master/scripts/dbdeployer-install.sh | bash
# or
# curl -L -s https://bit.ly/dbdeployer | bash

# Quality set for bash scripts:
# [set -e] fails on error
set -e

# [set -u ] stops the script on unset variables
set -u

# [set -o pipefail] triggers a failure if one of the commands in the pipe fail
set -o pipefail

# File containing latest version of dbdeployer
version_file=https://raw.githubusercontent.com/datacharmer/dbdeployer/master/common/VERSION

# check_exit_code checks the return code of the previous command
# exits the script if it is non-zero
function check_exit_code {
    args="$1"
    exit_code=$?
    if [[ $exit_code -ne 0 ]]
    then
        echo "ERROR running $args"
        exit $exit_code
    fi
}

# exists_in_path looks for an executable in each directory from $PATH
# If found, returns the full path
# When not found, returns an empty string
function exists_in_path {
    what=$1
    for dir in $(echo "$PATH" | tr ':' ' ')
    do
        wanted=$dir/$what
        if [[ -x "$wanted" ]]
        then
            echo "$wanted"
            return
        fi
    done
}

# (STEP 1) checks whether dbdeployer is already installed
# If it is, exits, recommending to use the internal upgrade feature instead
exists_dbdeployer=$(exists_in_path dbdeployer)
if [ -n "$exists_dbdeployer" ]
then
  echo "dbdeployer is already installed in $(dirname $exists_dbdeployer)"
  echo "Use 'dbdeployer update [--verbose]' to upgrade to the latest version"
  exit 1
fi

# (STEP 2) checks that the needed tools are installed in the system
for tool in tar curl gzip shasum
do
    found_tool=$(exists_in_path $tool)
    if [ -z "$found_tool" ]
    then
        echo "tool '$tool' not found"
        exit 1
    fi
done

# (STEP 3) collects the latest version from GitHub
dbdeployer_version=$(curl -s $version_file)
check_exit_code "curl -s $version_file"
if [ -z "${dbdeployer_version}" ]
then
    echo "error collecting version from ${version_file}"
    exit 1
fi

OS=$(uname -s | tr '[:upper:]' '[:lower:]')

# (STEP 4) checks the operating system
# Only Linux and MacOS are recognized
if [ "$OS" != "linux" -a "$OS" != "darwin" ]
then
    echo "Operating system '$OS' not recognized"
    exit 1
fi

# While "darwin" is what the system itself returns,
# "osx" is the identifier in dbdeployer binaries
if [ "$OS" == "darwin" ]
then
    OS=osx
fi

# Base URL for dbdeployer downloads
origin=https://github.com/datacharmer/dbdeployer/releases/download/v${dbdeployer_version}

# Name of the archive we are looking for
filename=dbdeployer-${dbdeployer_version}.${OS}.tar.gz

# Name of the file containing the checksum
checksum_file=${filename}.sha256

# Name of the executable inside the archive
ultimate_file=dbdeployer-${dbdeployer_version}.${OS}

# (STEP 5) Checks for already existing files.
# If any of the files that we need to download exist already,
# the program terminates
for existing in $filename $checksum_file $ultimate_file 
do
    if [ -f "$existing" ]
    then
        echo "File '$existing' exists already"
        exit 1
    fi
done

# Up to this point, no new files have been created

# (STEP 6) downloads the dbdeployer archive from GitHub
curl -L -s -o "$filename" "$origin/$filename"
check_exit_code "curl -L -s -o $filename $origin/$filename"

# (STEP 7) checks that the file was downloaded
if [ ! -f  "$filename" ]
then
    echo "File '$filename' not downloaded"
    exit 1
fi

# (STEP 8) downloads the checksum file
curl -L -s -o "$checksum_file" "$origin/${filename}.sha256"
check_exit_code "curl -L -s -o $checksum_file $origin/${filename}.sha256"

# (STEP 9) checks that the checksum file was downloaded
if [ ! -f  "${checksum_file}" ]
then
    echo "File '$checksum_file' not downloaded"
    exit 1
fi

# (STEP 10) probes the checksum
shasum -c "${checksum_file}"
check_exit_code "shasum -c $checksum_file"

# (STEP 11) unpacks the archive
tar -xzf "$filename"
check_exit_code "tar -xzf $filename"

# (STEP 12) checks that the executable was extracted
if [ ! -f "${ultimate_file}" ]
then
    echo "File '${ultimate_file}' not extracted"
    exit 1
fi

# (STEP 13) change attributes to the executable
chmod +x "${ultimate_file}"
check_exit_code "chmod +x ${ultimate_file}"

# (STEP 14) shows the downloaded files
ls -lh "${ultimate_file}" "${checksum_file}" "${filename}"
target=""

# local_targets are the places where a non-privileged user could store the executable
local_targets=("$HOME/usr/local/bin" "$HOME/go/bin" "$HOME/bin")

# if the local user is root
if [[ $EUID -eq 0 ]]
then
  # global paths for executables become local targets
  local_targets=(/usr/local/bin /usr/bin)
fi

# (STEP 15) checks if any of the destination directories is in $PATH
for dir in ${local_targets[*]}
do
  for p in $(echo "$PATH" | tr ':' ' ')
  do
    if [ "$dir" == "$p" -a -d "$dir" ]
    then
      target=$dir
      break
    fi
    if [ -n "$target" ]
    then
      break
    fi
  done
  echo "$dir"
done

# (STEP 16) copies the executable to the target directory and exits with success
if [ -n "$target" ]
then
  mv -f "${ultimate_file}" "${target}/dbdeployer"
  check_exit_code "mv -f '${ultimate_file}' '${target}/dbdeployer'"
  echo "File $ultimate_file copied to ${target}/dbdeployer"
  exit 0
fi

# If we have reached this point, the script is running as a non-root user
# and there were no suitable local directories in PATH

target=/usr/local/bin
target_found=""
if [ ! -d "$target" ]
then
    target=/usr/bin
fi

for dir in $(echo "$PATH" | tr ':' ' ')
do
  if [ "$dir" == "$target" ]
  then
    target_found=1
    break
  fi
done

# (STEP 16-a) suggest manual installation when automated one was not possible
echo "file ${ultimate_file} ready to move in \$PATH"
echo "e.g.:"
echo "sudo mv ${ultimate_file} ${target}/dbdeployer"

# (STEP 16-b)" warn about ultimate targets '/usr/bin' and '/usr/bin' missing from $PATH
if [ -z "$target_found" ]
then
  echo "WARNING: no suitable target directory found in \$PATH"
  exit 1
fi
