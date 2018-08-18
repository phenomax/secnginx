# SecNginX

**SecNginX** is a toolbox, which helps to:

* Build the latest stable NginX with selected modules from source
* Setup a basic file structure (Based on [server-configs-nginx](https://github.com/h5bp/server-configs-nginx))
* Apply best practice Security Headers and TLS-Config
* Provide hybrid RSA/ECDSA certificates
* Submit RSA/ECDSA certificates to [all](https://www.gstatic.com/ct/log_list/all_logs_list.json) Certificate Transparency Logs, currenctly active in Chrome

The aim of this project is to provide a fast solution to setup a secure, efficient and minimal NginX server. If you are searching for a more individual nginx build tool, shoot an eye on [nginx-build](https://github.com/cubicdaiya/nginx-build).

## Setup

* Download and extract the [lastest release](https://github.com/phenomax/secnginx/releases)
* Make it executable `chmod +x secnginx`
* Edit `config.toml` to your desires
* Start NginX installation `./secnginx install` - Check optional parameters with `./secnginx help install`

## Applied NginX Enhancements/Extensions (by default)

* OpenSSL 1.1.1-pre (TLS 1.3) - Version is configurable
* [Dynamic TLS Records](https://blog.cloudflare.com/optimizing-tls-over-tcp-to-reduce-latency/) patch to optimize latency
* [Dynamic CORS rules](https://github.com/x-v8/ngx_http_cors_filter)
* [Brotli](https://github.com/google/ngx_brotli) Compression algorithm
* [Nginx-CT](https://github.com/grahamedgecombe/nginx-ct) for using the Certificate Transparency TLS Extension **Important Note:** CT signature validation [is currently not supported](https://github.com/grahamedgecombe/nginx-ct/issues/36) in TLSv1.3
* [Headers-More](https://github.com/openresty/headers-more-nginx-module) for advanced output headers
* [Cookie Flags](https://github.com/AirisX/nginx_cookie_flag_module) Set Cookie Flags in NginX - `HttpOnly` is preset for all cookies in the delivered NginX config
* Up to date SSL and cipher list configuration
* Generate strong 4096bit Diffie-Hellmann parameters

## Further steps to consider

* Request RSA and ECDSA certificates from letsencrypt and setup [HSTS-Preload](https://hstspreload.org/)

    ### Example commands using [lego](https://github.com/xenolf/lego/)

    * `./lego -a -m contact@example.com -d example.com --webroot /var/www/ --path /etc/nginx/ssl/ecdsa -k ec384 run` for an EC384 certificate
    * `./lego -a -m contact@example.com -d example.com --webroot /var/www/ --path /etc/nginx/ssl/rsa -k rsa4096  run` for an RSA4096 certificate

* Submit your received certificates to various CT Logs using `secnginx submit-ct --input <path to public key> -output <path to output folder>`
* Setup a [CAA](https://support.dnsimple.com/articles/caa-record/)-DNS Record
* Check the existing `ssl_basic.conf` settings (especially the headers!)
* Check [Mozillas Web Security Guidelines](https://infosec.mozilla.org/guidelines/web_security)
* Setup AAAA-DNS Records to use IPv6
* Check your [Security Headers](https://securityheaders.io)
* Check your overall SSL deployment: [SSL Labs](https://www.ssllabs.com/ssltest/)

## Setup TLSA/DANE

**Please don't forget to setup [DNSSEC](https://support.dnsimple.com/articles/dnssec/) before using TLSA/DANE!**

When using short lived certificates, like these being issued by letsencrypt, you probably want to create your own Certificate Signing Request (CSR),
because ACME clients like [lego](https://github.com/xenolf/lego/) will generate a new private key for every renewal. As a consequence your certificates public key will change,
which results in the need to change your DANE DNS records on every cert renewal.

In case you want to deploy hybrid ECDSA/RSA certificates, follow this steps

* Create secure RSA and ECDSA private keys. Don't forget to store them somewhere safe!
    * `openssl genrsa -out rsa_privkey.pem 4096`
    * `openssl ecparam -name secp384r1 -genkey -out ecdsa_privkey.pem`
* Create CSR based on your private keys, *don't specify a challenge password!*
    * `openssl req -out rsa_csr.csr -key rsa_privkey.pem -new`
    * `openssl req -out ecdsa_csr.csr -key ecdsa_privkey.pem -new`
* Submit your generated CSR to [lego](https://github.com/xenolf/lego/) or another ACME client
    * `./lego -m contact@example.com -a --csr="ecdsa_csr.csr" --webroot /var/www/ --path /etc/nginx/ssl/ecdsa run`
    * `./lego -m contact@example.com -a --csr="rsa_csr.csr" --webroot /var/www/ --path /etc/nginx/ssl/rsa run`
* Generate your required DNS Records using [SSL-Tools TLSA Record Generator](https://ssl-tools.net/tlsa-generator) and publish them

## FAQ

### Why don't use BoringSSL ?

* No support for DHE key exchange
* No hybrid RSA/ECDSA certificates
* No OCSP stapling

### Why don't use LibreSSL ?

* No support for Certificate Transparency Timestamps
