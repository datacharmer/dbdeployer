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

version_file=https://raw.githubusercontent.com/datacharmer/dbdeployer/master/.build/VERSION

function check_exit_code {
    args="$1"
    exit_code=$?
    if [ "$exit_code" != "0" ]
    then
        echo "ERROR running $args"
        exit $exit_code
    fi
}

function exists_in_path {
    what=$1
    for dir in $(echo $PATH | tr ':' ' ')
    do
        wanted=$dir/$what
        if [ -x "$wanted" ]
        then
            echo "$wanted"
            return
        fi
    done
}

exists_dbdeployer=$(exists_in_path dbdeployer)
if [ -n "$exists_dbdeployer" ]
then
  echo "dbdeployer is already installed in $(dirname $exists_dbdeployer)"
  echo "Use 'dbdeployer update [--verbose]' to upgrade to the latest version"
  exit 1
fi

for tool in tar curl gzip shasum
do
    found_tool=$(exists_in_path $tool)
    if [ -z "$found_tool" ]
    then
        echo "tool '$tool' not found"
        exit 1
    fi
done

dbdeployer_version=$(curl -s $version_file)
check_exit_code "curl -s $version_file"

OS=$(uname -s | tr '[:upper:]' '[:lower:]')

if [ "$OS" != "linux" -a "$OS" != "darwin" ]
then
    echo "Operating system '$OS' not recognized"
    exit 1
fi
if [ "$OS" == "darwin" ]
then
    OS=osx
fi

origin=https://github.com/datacharmer/dbdeployer/releases/download/v${dbdeployer_version}
filename=dbdeployer-${dbdeployer_version}.${OS}.tar.gz
checksum_file=${filename}.sha256
ultimate_file=dbdeployer-${dbdeployer_version}.${OS}

for existing in $filename $checksum_file $ultimate_file 
do
    if [ -f "$existing" ]
    then
        echo "File '$existing' exists already"
        exit 1
    fi
done

curl -L -s -o "$filename" "$origin/$filename"
check_exit_code "curl -L -s -o $filename $origin/$filename"
if [ ! -f  "$filename" ]
then
    echo "File '$filename' not downloaded"
    exit 1
fi

curl -L -s -o "$checksum_file" "$origin/${filename}.sha256"
check_exit_code "curl -L -s -o $checksum_file $origin/${filename}.sha256"
if [ ! -f  "${checksum_file}" ]
then
    echo "File '$checksum_file' not downloaded"
    exit 1
fi

shasum -c "${checksum_file}"
check_exit_code "shasum -c $checksum_file"

tar -xzf "$filename"
check_exit_code "tar -xzf $filename"
if [ ! -f "${ultimate_file}" ]
then
    echo "File '${ultimate_file}' not extracted"
    exit 1
fi

chmod +x "${ultimate_file}"
check_exit_code "chmod +x ${ultimate_file}"

ls -lh
target=""
local_targets=($HOME/usr/local/bin $HOME/go/bin $HOME/bin)
if [[ $EUID -eq 0 ]]
then
  local_targets=(/usr/local/bin /usr/bin)
fi
for dir in ${local_targets[*]}
do
  for p in $(echo $PATH | tr ':' ' ')
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
if [ -n "$target" ]
then
  mv -f "${ultimate_file}" "${target}/dbdeployer"
  check_exit_code
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

for dir in $(echo $PATH | tr ':' ' ')
do
  if [ "$dir" == "$target" ]
  then
    target_found=1
  fi
done

echo "file ${ultimate_file} ready to move in \$PATH"
echo "e.g.:"
echo "sudo mv ${ultimate_file} ${target}/dbdeployer"

if [ -z "$target_found" ]
then
  echo "WARNING: no suitable target directory found in \$PATH"
  exit 1
fi