# wheretogo

A small HTTP server which only returns 301


## Why?
When you host your website behind Cloudfront or an loadbalancer you only get a hostname instead of a set of IP
addresse, as an entrypoint because the actual IP addresse can changeregularly. When you now configure the DNS for your
domain you can set a CNAME record for `www.example.com`, but depending on your provider you **can't** set a CNAME for
`example.com` because according to the specs (RFC1034), you are technically not allowed to have a CNAME on your *domain
apex* (the root of your domain). More advanced DNS providers give you things like an `ALIAS` record which you can use
just like a CNAME but are also valid on the domain apex. 
When your DNS provider doesn't support such a feature you need something that has a static IP and return an HTTP
redirect to the `www` subdomain. This is where this tool comes into play :)

## Setup

### Simple
For smaller setups you can just use a super small single instance and run this tool. But to be honest, you are probably
better of with just a simple nginx that you can configure yourself and get a certificate for your domain.

### Complex
Because most of our customers had such a single host in their account, we wanted to have one high-available setup which
we can use for all our customers. Therefore we setup an AWS Network Loadblancer (which has static IPs) and configured
an ASG behind it to have a bunch of servers with wheretogo running. The NLB is configure to simply pass port 80 and 443
through to the instances and we then take care of SSL ourselfs. We do this by using Caddy which automatically requests
a valid TLS certifiacte as soon as you configure a new domain (which this tool automatically makes when you start it
with the `-with-cady` option). Clustering is also quite easy via a shared storage backend (We use EFS for our setup),
so you can run multiple server and don't have problems with ACME validation or certifacte renewal.
