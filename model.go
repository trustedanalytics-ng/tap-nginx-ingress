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

const (
	nginxConf = `
events {
	worker_connections 1024;
}
http {
	# http://nginx.org/en/docs/http/ngx_http_core_module.html
	types_hash_max_size 2048;
	server_names_hash_max_size 2048;
	server_names_hash_bucket_size 2048;
	client_max_body_size 2048m;
	client_body_timeout 5m;
	include /etc/nginx/conf.d/*.conf;
	include /etc/nginx/sites-enabled/*;
	include /etc/nginx/ssl.conf;

	log_format	with_host	'$remote_addr - $remote_user [$time_local] $host "$request" '
					'$status $body_bytes_sent "$http_referer" '
					'"$http_user_agent" "$http_x_forwarded_for"';
	# standard log format with extra information about host

	access_log	/var/log/nginx/access.log	with_host;

	server {
		listen       80 default_server;
		listen       443 default_server;
		server_name  everythingelse;

		error_page 404 /404.html;

		location / {
			return 404;
		}

		# link the code to the file
		location = /404.html {
			root  /var/www/nginx/errors/;
		}
	}

	server {
		listen 5555;

		# liveness & readiness probes
		location /healthz {
			return 200;
		}
	}

	{{ range $httpserver := .HttpServers }}
	server {
		# Ingress: {{ $httpserver.IngressName }};
		listen {{ $httpserver.ListenPort }};
		server_name {{ $httpserver.Name }};
		proxy_read_timeout 900s;
		proxy_send_timeout 300s;
		proxy_connect_timeout 300s;
		proxy_buffering off;
		proxy_max_temp_file_size 0;
		{{ range $path := $httpserver.Locations }}
		location {{ $path.LocationPath }} {
			proxy_set_header Host $host;
			proxy_pass {{ $path.Protocol }}://{{ $path.ServiceName }}.{{ $path.Namespace }}:{{ $path.Port }};
			proxy_http_version 1.1;
			proxy_set_header Upgrade $http_upgrade;
			proxy_set_header Connection "upgrade";
			proxy_set_header X-Forwarded-Proto $scheme;
			# ServiceName: {{ $path.ServiceName }}
		}
		{{ end }}
	}

	{{ if eq  $httpserver.ListenPort "443 ssl" }}
	server {
		listen 80;
		server_name {{ $httpserver.Name }};
		return 302 https://$server_name$request_uri;
	}
	{{ end }}
	{{ end }}
}

`
)

type Location struct {
	LocationPath string
	ClusterIP    string
	Port         string
	ServiceName  string
	Protocol     string
	Namespace    string
}

type HttpServer struct {
	Locations   []*Location
	ListenPort  string
	Name        string
	IngressName string
}

type TemplateFeed struct {
	HttpServers []*HttpServer
	// TcpServers
	// UdpServers
}
