# 3. decoding url params

Date: 2023-06-05

## Status

Accepted

## Context

URL encoding in get query was not working properly to validate relation tuple

## Decision

Decode url params from request and send it to keto server to validate relation tuple

## Consequences

URL params are decoded and validation of relation tuple with encoding params working properly
