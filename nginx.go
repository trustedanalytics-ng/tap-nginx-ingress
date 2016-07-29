/*
Copyright 2015 The Kubernetes Authors All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"fmt"
	"log"
	"time"

	"github.com/hpcloud/tail"
)

func follow_logs() {
	go follow_logfile("/var/log/nginx/access.log")
	go follow_logfile("/var/log/nginx/error.log")
}

func follow_logfile(filepath string) {
	// TODO: add stop mechanism
	log.Println("tailing file ", filepath)
	for {
		t, err := tail.TailFile(filepath, tail.Config{Follow: true})
		if err != nil {
			log.Println("Tailing file error: ", err)
			time.Sleep(1000 * time.Millisecond)
			continue
		}
		for line := range t.Lines {
			fmt.Println("NGINX: ", line.Text)
		}
	}
}

func start_nginx() bool {
	if !shellOut("nginx") {
		log.Println("Failed to start nginx!")
		return false
	}
	return true
}

func reload_nginx() bool {
	if !shellOut("nginx -s reload") {
		log.Println("Failed to reload nginx!")
		return false
	}
	return true
}
