# 2. Restructuring, Caching of tokens, Unit Tests, Sonar Integration and B3 Propagation opentelemetry

Date: 2023-02-10

## Status

Accepted

## Context

* To remove common code usage and restructuring controllers and services and handle 5xx errors 
* Handling semicolon delimitation in url 
* Two request take place in check API in which Hydra request can be cached to save time
* Following SLIs to overcome SLOs
* keto go client will be more convenient to call keto server instead of using http client

## Decision

* Restructuring of whole service
* Added logic to fetch query params on its own
* Cached hydra tokens
* Implemented Unit tests and sonar integration 
* Implemented B3 Propagator opentelemetry 
* Keto go client implemented to call keto server v0.11

## Consequences

* 5xx errors resolved with auth token and semicolon delimiter in url
* Saved one jump to hydra to fetch token subject
* Proper implementation of SLIs with writing unit tests and opentelemetry will help in more detailed information on requests
