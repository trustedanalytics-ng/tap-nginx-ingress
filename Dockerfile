# Copyright 2015 The Kubernetes Authors. All rights reserved.
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

#FROM nginx:stable
FROM tapimages:8080/nginx:stable

RUN rm -f /var/log/nginx/access.log
RUN rm -f /var/log/nginx/error.log

RUN openssl dhparam -out /etc/nginx/dhparams.pem 2048

COPY controller /
COPY default.conf /etc/nginx/nginx.conf
COPY ssl.conf /etc/nginx/

RUN mkdir -p /etc/nginx/ssl

WORKDIR /etc/nginx
CMD ["/controller"]
