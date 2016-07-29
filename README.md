# TAP fork of:

https://github.com/kubernetes/contrib/tree/master/ingress/controllers/nginx-alpha

# Nginx Ingress Controller

Major change from upstream (not-compatible) are:

- Using ClusterIPs, not SkyDNS.
- Will work on supporting steam (TCP) and UDP flows.
- Trusts TAP CA

In far future:
- integrate with ACME protocol of let's encrypt.

## Overall flow:

Fetch Services, Fetch Ingresses
Generate template
reload nginx

## Deploying the controller

Please see tap-deply, roles, k8s-app, ingress.yml.
Example usage in dashboard.yml, service-catalog.yml and uaa.yml.

