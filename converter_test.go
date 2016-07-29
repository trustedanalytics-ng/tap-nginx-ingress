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
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/apis/extensions"
)

const (
	expectedHost        string = "localhost"
	expectedIP          string = "1.1.1.1"
	expectedPath        string = "Path"
	expectedNameIngress string = "MyName"
	expectedNameService string = "Bucket"
	expectedPort        string = "80"
)

func getProperServiceList() *api.ServiceList {
	item := api.Service{}
	item.Name = expectedNameService
	item.Spec = api.ServiceSpec{ClusterIP: expectedIP}
	return &api.ServiceList{Items: []api.Service{item}}
}

func getProperIngresItem() extensions.Ingress {
	path := extensions.HTTPIngressPath{}
	path.Path = expectedPath

	httpIng := &extensions.HTTPIngressRuleValue{}
	httpIng.Paths = []extensions.HTTPIngressPath{path}

	inrule := extensions.IngressRule{}
	inrule.Host = expectedHost
	inrule.HTTP = httpIng

	ingressItem := extensions.Ingress{}
	ingressItem.Name = expectedNameIngress

	ingressItem.Spec = extensions.IngressSpec{Rules: []extensions.IngressRule{inrule}}
	return ingressItem
}

func TestGet_getServiceNameToIPMapping_generation(t *testing.T) {
	Convey("Test getServiceNameToIPMapping correctnes", t, func() {
		serviceList := getProperServiceList()
		cache := getServiceNameToIPMapping(serviceList)
		ip, present := cache[expectedNameService]

		So(present, ShouldBeTrue)
		So(ip, ShouldEqual, expectedIP)
		So(len(cache), ShouldEqual, 1)
	})
}

func Test_make_feed(t *testing.T) {
	Convey("For empty ingress items returns empty list", t, func() {
		serviceList := getProperServiceList()
		ingresesList := &extensions.IngressList{Items: []extensions.Ingress{}}

		feed, err := make_feed(ingresesList, serviceList)

		So(err, ShouldBeNil)
		So(feed, ShouldResemble, &TemplateFeed{})
	})

	Convey("For empty ingress spec rules items returns empty list", t, func() {
		serviceList := getProperServiceList()

		ingressItem := getProperIngresItem()
		ingressItem.Spec.Rules = []extensions.IngressRule{}
		ingresesList := &extensions.IngressList{Items: []extensions.Ingress{ingressItem}}

		feed, err := make_feed(ingresesList, serviceList)

		So(err, ShouldBeNil)
		So(feed, ShouldResemble, &TemplateFeed{})
	})

	Convey("For empty IngressRuleValue returns list without locations", t, func() {
		serviceList := getProperServiceList()

		ingressItem := getProperIngresItem()
		ingressItem.Spec.Rules[0].HTTP = nil //http ingress rule value
		ingresesList := &extensions.IngressList{Items: []extensions.Ingress{ingressItem}}

		feed, err := make_feed(ingresesList, serviceList)

		So(err, ShouldBeNil)
		So(len(feed.HttpServers), ShouldEqual, 1)
		So(feed.HttpServers[0].ListenPort, ShouldEqual, expectedPort)
		So(feed.HttpServers[0].Name, ShouldEqual, expectedHost)
		So(feed.HttpServers[0].IngressName, ShouldEqual, expectedNameIngress)
		So(len(feed.HttpServers[0].Locations), ShouldEqual, 0)
	})

	Convey("For proper Ingress returns default feed", t, func() {
		serviceList := getProperServiceList()
		ingressItem := getProperIngresItem()
		ingresesList := &extensions.IngressList{Items: []extensions.Ingress{ingressItem}}

		feed, err := make_feed(ingresesList, serviceList)

		So(err, ShouldBeNil)
		So(len(feed.HttpServers), ShouldEqual, 1)
		So(feed.HttpServers[0].ListenPort, ShouldEqual, expectedPort)
		So(feed.HttpServers[0].Name, ShouldEqual, expectedHost)
		So(feed.HttpServers[0].IngressName, ShouldEqual, expectedNameIngress)
		So(len(feed.HttpServers[0].Locations), ShouldEqual, 1)
		So(feed.HttpServers[0].Locations[0].LocationPath, ShouldEqual, expectedPath)
	})
}

func Test_extractLocation(t *testing.T) {
	Convey("For proper ingress and path returns location with provided ip", t, func() {
		serviceNameToIP := make(map[string]string)
		serviceNameToIP[expectedNameService] = expectedIP

		ingressItem := getProperIngresItem()
		path := extensions.HTTPIngressPath{Path: expectedPath}
		path.Backend = extensions.IngressBackend{ServiceName: expectedNameService}

		location := extractLocation(serviceNameToIP, ingressItem, path)

		So(location.LocationPath, ShouldEqual, expectedPath)
		So(location.ClusterIP, ShouldEqual, expectedIP)
	})

	Convey(fmt.Sprintf("For proper ingress without path returns location %s ip", DefaultIP), t, func() {
		serviceNameToIP := make(map[string]string)
		serviceNameToIP[expectedNameService] = expectedIP

		ingressItem := getProperIngresItem()
		path := extensions.HTTPIngressPath{Path: expectedPath}

		location := extractLocation(serviceNameToIP, ingressItem, path)

		So(location.LocationPath, ShouldEqual, expectedPath)
		So(location.ClusterIP, ShouldEqual, DefaultIP)
		So(location.Port, ShouldEqual, DefaultPort)
		//make sure default is different than expected
		So(expectedIP, ShouldNotEqual, DefaultIP)
		So(expectedPort, ShouldNotEqual, DefaultPort)
	})
}
