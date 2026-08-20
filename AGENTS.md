# AGENTS.md

Guidance for AI agents working in this repository. Read this before making changes.
See `README.md` for the user-facing explanation of what the project does.

## What this repo is

`multi-operator-manager` (MOM) is an OpenShift Go module that is **both a CLI tool and a
library**:

- The **tool** (`multi-operator-manager` binary) tests, debugs, and analyzes OpenShift
  operators *offline* (no live cluster).
- The **library** (`pkg/library/...`) is imported by real operators so they become
  "compatible" — i.e. expose the three MOM verbs.

A compatible operator turns its reconcile logic into a function of its inputs: given input
resources, it computes the mutations it *would* make, without touching a cluster. It does
this via three subcommands: `input-resources`, `apply-configuration`, `output-resources`.

## The core: `pkg/library`

`pkg/library` is the heart of the project — the shared implementation of the operator
contract, and the part other repos import. It does two jobs:

- **Defines the contract.** It provides the ready-made cobra commands for the three verbs and
  the types they exchange, so an operator doesn't implement them from scratch. An operator
  becomes compatible by importing these packages and supplying three things: its input list,
  its output list, and its decision logic (`pkg/sampleoperator` is the worked example).
- **Runs and checks the results.** It contains the machinery that drives an operator once
  (`SimpleOperatorStarter` and the run-once helpers), writes the resulting mutations to disk,
  and compares two result sets for equivalence — the same logic the test harness relies on.

Its three subpackages map one-to-one to the verbs:

- `libraryinputresources/` — the `input-resources` verb and input types.
- `libraryapplyconfiguration/` — the `apply-configuration` verb, the run-once machinery, and
  mutation output/comparison.
- `libraryoutputresources/` — the `output-resources` verb and output types.

Because this is the public surface other operators depend on, treat changes to its exported
types and function signatures as **breaking changes** and weigh them accordingly.

For a real-world consumer of this contract, see the
[cluster-authentication-operator example in the README](README.md#real-world-example-cluster-authentication-operator)
(`pkg/sampleoperator` is the in-repo example).

## Commands

```
make build                          # build both binaries (multi-operator-manager, sample-operator) into repo root
make check                          # verify + unit tests (run before considering a change done)
make test-unit                      # go test ./pkg/... ./cmd/...
make verify                         # gofmt/vet/etc. via build-machinery-go
make test-operator-integration      # build + run sample-operator against test-data, compare vs expected output
make update-test-operator-integration   # same, but REGENERATE the expected output (use after intended changes)
make test-e2e                       # requires openshift-tests in PATH; rarely needed locally
make clean                          # remove built binary + test-output
```

Run a single package's tests directly with `go test ./pkg/library/libraryinputresources/...`.

## Repo layout

- `cmd/` — two `main` packages: `multi-operator-manager` and `sample-operator`.
- `pkg/cmd/multi-operator-manager/` — the cobra CLI tree (`test`, `sample-operator`,
  `create-input-resources`).
- `pkg/library/` — the reusable contract; **this is the public API other repos import**:
  - `libraryinputresources/` — `input-resources` verb + input types.
  - `libraryapplyconfiguration/` — `apply-configuration` verb, the operator launch/run-once
    machinery, and mutation output/comparison.
  - `libraryoutputresources/` — `output-resources` verb + output types.
- `pkg/sampleoperator/` — reference compatible operator (the example to copy from).
- `pkg/test/testapplyconfiguration/` — the test harness behind `test apply-configuration`.
- `test-data/apply-configuration/*/` — integration fixtures: each dir with a `test.yaml` has
  an `input-dir/` and an `expected-output/` (the known-good result to compare against).
- `vendor/` — dependencies are **vendored**; dep changes go through `go mod` + build-machinery
  vendoring, not hand edits.

## Conventions and gotchas

- **Output must be reproducible in content, and order-independent.** `apply-configuration`
  should compute the same *set* of desired mutations for the same inputs, regardless of the
  order controllers happen to run in. To support this, time is injected via `Clock` / `--now`
  and reads go through the injected `MutationTrackingClient` (in `ApplyConfigurationInput`)
  rather than a live cluster or disk; and `SimpleOperatorStarter` shuffles controllers before
  running them and rejects duplicate controller names.

  Practical consequence: do **not** introduce `time.Now()`, `math/rand`, env/file reads, or
  network/cluster calls into operator logic — use the injected clock and client instead, and
  don't rely on controller run order.
- **A test passes when the *meaningful* output matches — not when the files are identical.**
  The harness does not require the run's output to be byte-for-byte equal to
  `expected-output/`. Instead it compares the actual resource mutations field-by-field, and it
  ignores events entirely (`EquivalentApplyConfigurationResultIgnoringEvents` in
  `equivalence.go`). So differences that are only in events, formatting, or ordering will
  still pass; only a difference in the actual mutations an operator wants to make will fail a
  test.
- **Do not hand-edit `expected-output/` or `test-output/`.** These are generated. To update
  the expected output after an intentional behavior change, run
  `make update-test-operator-integration` (or the harness with `--replace-expected-output`),
  then review the diff.
- **Output is structured by cluster type + verb + GVR.** Mutations land under
  `Configuration/`, `Management/`, or `UserWorkload/`, then `Create`/`ApplyStatus`/etc. When
  reasoning about a change, look at which of these buckets shifts.
- **`output-resources` must be complete.** Any GVR an operator mutates must be declared
  there; mutations to undeclared resources are filtered out (a bug signal).

## When adding or changing behavior

1. Make the code change (keep output reproducible/order-independent — see above).
2. `make check` for unit tests + verify.
3. If integration output changes intentionally, regenerate the expected output with
   `make update-test-operator-integration` and inspect the diff to confirm only the intended
   mutations changed.
4. If you add a new operator capability, mirror the pattern in `pkg/sampleoperator/` and add
   a fixture under `test-data/apply-configuration/`.
