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
all: clean container

TAG = 0.2
PREFIX = tap-nginx-ingress

build: clean build_anywhere

prepare_dirs:
	mkdir -p ./temp/src/github.com/trustedanalytics-ng/tap-ingress
	$(eval REPOFILES=$(shell pwd)/*)
	ln -sf $(REPOFILES) temp/src/github.com/trustedanalytics-ng/tap-ingress
	
build_anywhere: prepare_dirs
	$(eval GOPATH=$(shell cd ./temp; pwd))
	$(eval APP_DIR_LIST=$(shell GOPATH=$(GOPATH) go list ./temp/src/github.com/trustedanalytics-ng/tap-ingress/... | grep -v /vendor/))
	GOPATH=$(GOPATH) CGO_ENABLED=0 go build -a -installsuffix cgo -ldflags '-w' -o controller -tags netgo $(APP_DIR_LIST)

container: build
	docker build -t $(PREFIX):$(TAG) .

pushlocal: container
	docker tag $(PREFIX):$(TAG) 127.0.0.1:30000/$(PREFIX):$(TAG)
	docker push 127.0.0.1:30000/$(PREFIX):$(TAG)

clean:
	rm -f controller
	rm -rf ./temp/

test:
	go test --cover -tags netgo -ldflags '-w' $(APP_DIR_LIST)
