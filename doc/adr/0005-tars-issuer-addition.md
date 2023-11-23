# 5. Tars Issuer Addition

Date: 2023-11-17

## Status

Accepted

## Context

* Tars team needs to integrate ozone in their service, so their introspection API was needed to be merged in issuer's list

## Decision

* Tars issuer is added in the config and query param list in auth check API

## Consequences

* Tars service can call auth check api to validate their token on tars introspection API
