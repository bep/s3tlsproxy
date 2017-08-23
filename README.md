# s3tlsproxy

**NOTE: This is not ready for use.**

WORK IN PROGRESS: Amazon S3 cache with auto TLS and virtual host support.

**The main use case for this tool is to host a set of [Hugo](https://gohugo.io/) sites with automatic https, backed by one or more Amazon S3 buckets, with a cache and as little administration need as possible.**

Planned features:

* TLS via [https://letsencrypt.org/](https://letsencrypt.org/)
* Virtual hosts support with S3 bucket sharing
* Hot-Reloading of server, including adding and removing virtual hosts
* Cache to save money on S3 bandwidth
* Load balancing and HA via DNS
* Cross-site (i.e. all servers in domain) cache purge from signed URLs
