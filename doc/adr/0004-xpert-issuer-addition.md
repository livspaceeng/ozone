# 4. Xpert Issuer Addition

Date: 2023-09-19

## Status

Accepted

## Context

* Xpert team needs to integrate ozone in their service, so their introspection API was needed to be merged in issuer's list

## Decision

* Xpert issuer is added in the config and query param list in auth check API

## Consequences

* Xpert service can call auth check api to validate their token on xpert's introspection API
