---
type: technical
date: 2026-03-19
author: sarah
anchor: env0/services/payment-service/
linear_issue: ENG-892
stale: false
---

# Chose SQS over direct invocation for vendor API

The vendor API rate-limits at 100 req/s. Direct invocation from the Lambda
would hit this limit during peak hours. SQS provides natural backpressure
and retry semantics without custom rate-limiting code.

Considered alternatives:
- Direct invocation with client-side rate limiting — rejected
- SNS + SQS fan-out — overkill for a single consumer
