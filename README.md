# Lemur
Repository for projects concerning the Lemur certificate manager.

## Lemur-Client
Web-based self-service portal for Lemur certificates. Authentication via Okta.

_NOTES_: 
* You must export the `LEMUR_USER` and `LEMUR_PASS` variables as these are no longer automatically pulled out of the secrets file.
* Lemur-client must be run in a context in which it has a route to the requisite Lemur instance.
  * Example: At my employer, running lemurclient from my personal kubernetes cluster, with no VPCs set up, will fail to provide a service since the Lemur installation runs in ms-pipeline.
* The web server will attempt to set up and refresh its own SSL certificates using the Let's Encrypt authority. HOWEVER, take note when developing/testing/validating that _there is a limit of 5 duplicate certificates per week!_ See [the Let's Encrypt docs](https://letsencrypt.org/docs/rate-limits/)
* Certificates are requested/generated based on the first of the month. If you revoke the certificate in the middle of the month, the API request _should_ handle it by searching for only _valid_ certs, but as of now the behavior is undefined. Beware.

## Setup
`source setup.sh`
This script installs govendor and ensures the src/lemur/ directory has a vendor/ directory with local copies of all dependencies.
If you wish to work on a different Go project within the same terminal session, you may revert the virtualenv-like by invoking the function `revert`.

## Doing the Things
### Build
`./build.sh`
Creates an executable in `bin/lemur-client`.

### Run
`./bin/lemur-client [ options ]`
You're running a web service!

### Test
`cd src/lemur/; go test`
Run unit tests in \*\_test.go files.


## Options

* -config The path to configuration yaml.
* -host_port Port on which to listen for web traffic. Overridden by environment variable.
* -admin_port Port on which to listen for web traffic for admin tasks.. Overridden by environment variable.
* -statsd_host Host to which to send statsd metrics. Overridden by environment variable.
* -statsd_port Port on statsd_host. Overridden by environment variable.

## Environment variables

* LEMUR_USER Username with which to authenticate to Lemur service
* LEMUR_PASS Password with which to authenticate to Lemur service
* HOST_PORT Port on which to listen for web traffic. Overrides -host_port option.
* ADMIN_PORT Port on which to listen for web traffic for admin tasks. Overrides -admin_port option. Cannot be the same as HOST_PORT.
* STATSD_HOST Host to which to send statsd metrics. Overrides -statsd_host option.
* STATSD_PORT Port on STATSD_HOST. Overrides -statsd_port option.
