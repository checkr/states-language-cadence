# States Language on Cadence

This is an implementation of [States Language](https://states-language.net/spec.html) on top of the [Cadence Workflow](https://cadenceworkflow.io/) engine.

Amazon States Language (ASL) is a JSON specification describing state machines and workflows. It is the description that powers AWS Step Functions. Cadence is a robust workflow engine created and open sourced by Uber.
This project allows you to mimic Step Functions outside of AWS infrastructure on inside a Cadence workflow.

[Spec here](./STATE_SPEC.md)

#### Updating vendors

```
make mod-vendor
```
