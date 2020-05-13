#!/bin/bash

# Copyright 2019 Google Inc. All rights reserved.
# 
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
# 
# http://www.apache.org/licenses/LICENSE-2.0
# 
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

mkdir -p /opt/src
cd /opt/src
echo "Cloning the repos in ${CLONE_REPOS}..." >&2
for repo in ${CLONE_REPOS}; do
  echo "Cloning the repo \"${repo}\"" >&2
  git clone --config "remote.origin.fetch=+refs/pull/*:refs/pull/*" --config "remote.origin.fetch=+refs/devtools/*:refs/devtools/*" "${repo}"
done

for repo in `find /opt/src -type d -name '.git'`; do
  cd "$(dirname "${repo}")"
  /opt/bin/git-appraise pull
done

echo "Starting the 'git-appraise-web' program" >&2
cd /opt/src
/opt/bin/git-appraise-web --port ${PORT}

