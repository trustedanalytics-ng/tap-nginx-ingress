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

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/apis/extensions"
)

const (
	DefaultIP   string = "255.255.255.255"
	DefaultPort string = "1"
)

func getServiceNameToIPMapping(services *api.ServiceList) map[string]string {
	mapping := make(map[string]string)
	for _, svc := range services.Items {
		mapping[svc.Name] = svc.Spec.ClusterIP
	}
	return mapping
}

func make_feed(ingresses *extensions.IngressList, services *api.ServiceList) (*TemplateFeed, error) {
	serviceNameToIP := getServiceNameToIPMapping(services)
	feed := &TemplateFeed{}
	for _, ingress := range ingresses.Items {
		ingress_spec := ingress.Spec
		for _, rules := range ingress_spec.Rules {
			current_http_server := HttpServer{}

			current_http_server.ListenPort = "80"
			if val, ok := ingress.ObjectMeta.Annotations["useExternalSsl"]; ok && val == "true" {
				current_http_server.ListenPort = "443 ssl"
			}

			current_http_server.Name = rules.Host
			current_http_server.IngressName = ingress.Name
			feed.HttpServers = append(feed.HttpServers, &current_http_server)
			if rules.IngressRuleValue.HTTP == nil {
				log.Println("No http rules defined for ingress", ingress.Name)
				continue
			}
			for _, path := range rules.IngressRuleValue.HTTP.Paths {
				current_location := extractLocation(serviceNameToIP, ingress, path)
				current_http_server.Locations = append(current_http_server.Locations, current_location)
			}
		}
	}
	return feed, nil
}

func extractLocation(servicesIpCache map[string]string, ingress extensions.Ingress, path extensions.HTTPIngressPath) *Location {
	current_location := Location{}
	current_location.Namespace = ingress.Namespace
	current_location.LocationPath = path.Path
	current_location.ServiceName = path.Backend.ServiceName

	ip, present := servicesIpCache[path.Backend.ServiceName]
	if present {
		current_location.ClusterIP = ip
		current_location.Port = path.Backend.ServicePort.String()
	} else {
		//return feed, errors.New("unable to get cluster ip for service " +  path.Backend.ServiceName)
		log.Printf("Unable to get cluster ip for service %s - returning %s (port %s)", path.Backend.ServiceName, DefaultIP, DefaultPort)
		current_location.ClusterIP = DefaultIP
		current_location.Port = DefaultPort
	}

	current_location.Protocol = "http"
	if val, ok := ingress.ObjectMeta.Annotations["useSsl"]; ok && val == "true" {
		current_location.Protocol = "https"
	}
	return &current_location
}
