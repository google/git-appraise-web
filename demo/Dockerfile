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
FROM debian

RUN apt-get update && apt-get upgrade -y && \
    apt-get install -y -qq --no-install-recommends \
      ca-certificates \
      git \
      supervisor \
      wget && \
    mkdir -p /opt/bin && \
    mkdir -p /var/log/supervisor && \
    mkdir -p /var/log/app_engine/custom_logs

ADD supervisord.conf /etc/supervisor/conf.d/supervisord.conf
ADD pull.sh /opt/bin/pull.sh

RUN chmod u+x /opt/bin/pull.sh

RUN wget -O /opt/go1.6.2.linux-amd64.tar.gz \
      https://storage.googleapis.com/golang/go1.6.2.linux-amd64.tar.gz && \
    tar -C /usr/local -xzf /opt/go1.6.2.linux-amd64.tar.gz && \
    export PATH=${PATH}:/usr/local/go/bin/:/opt/bin/ && \
    export GOPATH=/opt/ && \
    go get github.com/google/git-appraise/git-appraise && \
    go get github.com/google/git-appraise-web/git-appraise-web && \
    rm -rf /opt/go1.4.2.linux-amd64.tar.gz && \
    rm -rf /usr/local

EXPOSE 8080

CMD '/usr/bin/supervisord'
