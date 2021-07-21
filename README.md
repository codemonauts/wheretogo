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

## Configuration
```yaml
"b.example.com":
  - name: "b_Moved-content"
    description: "Moved content"
    pathPrefix: "/images"
    match: "/images/(.*)"
    target: "https://blog.example.com/assets/$1"
```

The keys of the config file are the domains on which the tool will listen and what will configured in Caddy. So make
sure that these domains point to your server or otherwise Caddy won't be able to request a TLS certificate for them.

The value for the domainnames is a list of config blocks, where each entry represents one rule. The following keys are
available in this block (Bold = Required):

* **name**
  The unique name of this rule. This will be used in the log output when activating debug logging

* description
  This is just for the user to document the configuration file and not used in the code

* pathPrefix (default: `/`)
  The rule will only be triggered when the uri of the request matches this prefix. This is used to define a hierarchy
  when defining multiple rules for a domain

* match (default: `(.*)`)
  This must be a valid regular expression and is used on the uri of the request to define the location for the
  redirect. Capture groups can be defined and used in the target value

* **target**
  This will be used as the *Location* field of the redirect. You can use the capture groups you defined in the `match`
  section.

See the `config_testing.yml` for more example rules.
