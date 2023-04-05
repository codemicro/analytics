# codemicro/analytics

A basic self-hosted analytics system that ingests logs from the Caddy HTTP server.

Note that this is designed to be used with [Authentik's proxy provider](https://goauthentik.io/docs/providers/proxy/) and may be insecure when used without adequate authentication in place. 

## Architecture

* Ingest server
  * A Go app listening on a TCP socket to ingest JSON logs from the Caddy web server
  * Groups requests into sessions and persists them to a database
* Datasette
  * Provides a frontend, data explorer and visualisation tool

## Docker

`Dockerfile` can be used for deployment. The directory containing the database and configuration file (which must be called `analytics.db` and `config.yml` respectively) should be mounted at `/analytics/run` within the container.

The ingest server listens on `0.0.0.0:7500` and Datasette listens on `0.0.0.0:8001`.

The following environment variables should be set:
* `BASE_URL`: the URL at which the Datasette instance can be accessed at.

## Caddy logging configuration

```
log {
    output net whateveraddr:7500
    format json
}
```

---

*Last commit containing the original web UI: `b817d6a23ff2ea51bd55a9a8209e44feeebb20ff`*
