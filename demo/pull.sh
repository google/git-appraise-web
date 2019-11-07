#!/bin/bash

# Copyright 2016 Google Inc. All rights reserved.
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

# We must output "READY" before supervisord will send events
echo "READY"

# Read in the event header sent by supervisord
read

# We must output "RESULT 2\nOK" to let supervisord know the event has been accepted
echo "RESULT 2"
echo "OK"

# Now we do the actual pulls
for repo in `find /opt/src -type d -name '.git'`; do
    cd "$(dirname "${repo}")" && \
    git pull origin master && \
    git fetch origin '+refs/heads/*:refs/remotes/origin/*' && \
    git fetch origin '+refs/devtools/*:refs/devtools/*' && \
    git fetch origin '+refs/pull/*:refs/pull/*' && \
    /opt/bin/git-appraise pull
done
