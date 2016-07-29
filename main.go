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
	"log"
	"os"
	"reflect"
	"text/template"

	"time"

	"github.com/davecgh/go-spew/spew"

	"k8s.io/kubernetes/pkg/api"
	client "k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/util/flowcontrol"
)

func main() {
	var ingClient client.IngressInterface
	var svcClient client.ServiceInterface
	if kubeClient, err := client.NewInCluster(); err != nil {
		log.Fatalf("Failed to create client: %v.", err)
	} else {
		ingClient = kubeClient.Extensions().Ingress(api.NamespaceAll)
		svcClient = kubeClient.Services(api.NamespaceAll)
	}
	tmpl, _ := template.New("nginx").Parse(nginxConf)
	rateLimiter := flowcontrol.NewTokenBucketRateLimiter(0.1, 1)
	last_template_feed := &TemplateFeed{}

	nc := &NginxConfigGenerator{
		ingClient:          ingClient,
		svcClient:          svcClient,
		tmpl:               tmpl,
		last_template_feed: last_template_feed,
	}

	nc.updateConfigInFile()
	// ^^ ignoring returned error as it should be fully handled within

	if !start_nginx() {
		log.Fatal("Failed to start nginx!")
	}
	follow_logs()
	for {
		time.Sleep(5000 * time.Millisecond)
		rateLimiter.Accept()
		changed, _ := nc.updateConfigInFile()
		if changed && !reload_nginx() {
			log.Fatal("Failed to reload nginx!")
		}
	}
}

type NginxConfigGenerator struct {
	ingClient          client.IngressInterface
	svcClient          client.ServiceInterface
	tmpl               *template.Template
	last_template_feed *TemplateFeed
}

// updateConfigInFile generates Nginx config file and compares it with previous version.
// first return value informs about change in configuration,
// second return value is error that doesn't need handling, but could be used in tests
func (nc *NginxConfigGenerator) updateConfigInFile() (bool, error) {
	ingresses, err := nc.ingClient.List(api.ListOptions{})
	if err != nil {
		log.Printf("Error retrieving ingresses: %v", err)
		return false, err
	}
	services, err := nc.svcClient.List(api.ListOptions{})
	if err != nil {
		log.Printf("Error retrieving services: %v", err)
		return false, err
	}
	template_feed, err := make_feed(ingresses, services)
	if err != nil {
		log.Fatalf("Error generating template feed data: %v for data: ingresses: %v and services: %v", err, ingresses, services)
		return false, err
	}
	if reflect.DeepEqual(template_feed, nc.last_template_feed) {
		log.Println("No changes detected.")
		return false, nil
	}
	log.Println("updated template feed:")
	spew.Dump(template_feed)

	if w, err := os.Create("/etc/nginx/nginx.conf"); err != nil {
		log.Fatalf("Failed to open %v: %v", nginxConf, err)
	} else if err := nc.tmpl.Execute(w, template_feed); err != nil {
		log.Fatalf("Failed to write template %v", err)
	}
	nc.last_template_feed = template_feed
	return true, nil
}
