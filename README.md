# Multi-Operator-Manager (MOM)

MOM is a command-line tool (binary) that lets you **test, debug, and compare OpenShift
operators offline — without a live cluster**.

Normally an operator only runs against a live cluster: it watches resources and writes
changes back, deciding and acting in the same step. That makes it hard to test and
impossible to simply ask *"what would you do?"* without it actually doing it. MOM changes
that: you hand an operator a folder of resources describing the cluster's current state, and
it prints the exact changes it *would* make — without applying any of them.

To work with MOM, an operator exposes three commands ("verbs"):

| Verb                  | In plain terms                                                              |
| --------------------- | -------------------------------------------------------------------------- |
| `input-resources`     | "Everything I need to read before I can decide anything."                  |
| `apply-configuration` | "Given those resources, the changes I'd make (create/update/delete/status)." |
| `output-resources`    | "The complete list of things I'm ever allowed to change."                  |

You don't write these from scratch: the packages under `pkg/library/` provide the commands,
and the operator just supplies three functions — its input list, its output list, and its
decision logic (see `pkg/sampleoperator` for a working example). Once an operator exposes
these three verbs, it is **compatible** and MOM can drive it.

Because every compatible operator answers in the same standard format, MOM can **test** an
operator (compare its output to a known-good result), **debug** a real cluster (replay its
state through the operator on your laptop), and **compare** operators to each other (e.g.
spot two that would fight over the same resource) — all without touching a running cluster.

The reference implementation lives under `pkg/sampleoperator` and is exposed through the
`sample-operator` subcommand.

## Why MOM — example scenarios

### 1. CI regression testing, no cluster required

You change a controller in your operator and want to prove it only affected what you
intended. `apply-configuration` is a pure function — given a directory of input resources
it prints the exact mutations it *would* make — so CI can diff that output against the
checked-in expected output:

```
./multi-operator-manager test apply-configuration \
  --test-dir=./test-data/apply-configuration/ \
  --output-dir=./test-output \
  --preserve-policy=KeepAlways
```

**What you learn:** whether your change altered the operator's decisions. The run is
deterministic (time is pinned via `now` in each `test.yaml`) and takes milliseconds — no
cluster, no flakiness. `test-output/junit.xml` summarizes pass/fail per test.

### 2. Debug a real cluster from a must-gather

A customer cluster is misbehaving and all you have is a must-gather. Build an input
directory from it, then run `apply-configuration` to see what the operator *would* do
against that exact state — on your laptop, with no cluster access:

```
# 1. capture the resources the operator declares it reads
<path-to-operator> input-resources > pertinent-resources.yaml

# 2. extract just those resources from the must-gather into an input dir
./multi-operator-manager create-input-resources from-must-gather \
  --must-gather-dir=<dump> \
  --input-resources=pertinent-resources.yaml \
  --output-dir=<input-dir>

# 3. run the operator against that input dir
<path-to-operator> apply-configuration \
  --input-dir=<input-dir> \
  --output-dir=<result-dir>
```

**What you learn:** the operator's intended creates/updates/deletes against real cluster
state, reproducibly, offline.

### 3. Multi-operator conflict and dependency analysis

Every compatible operator declares what it reads (`input-resources`) and the complete set
it may mutate (`output-resources`). Line those lists up across operators:

```
<operator-a> output-resources > a-out.yaml
<operator-b> output-resources > b-out.yaml
<operator-a> input-resources  > a-in.yaml
```

**What you learn:**
- A resource (GVR) appearing in two operators' `output-resources` is a latent **write
  conflict** (they will fight over it on a live cluster).
- Matching one operator's `input-resources` against another's `output-resources` reveals a
  **dependency / ordering** edge.
- A mutation emitted by `apply-configuration` for a GVR that is *not* in `output-resources`
  is an out-of-bounds bug.

This is the "multi" the project is named for: reducing each operator to comparable,
declarative pieces so a whole cluster's worth of them can be reasoned about together.

## Building

```
make build
```

This produces two binaries in the repo root:

| Binary                    | What it is                                                                        |
| ------------------------- | --------------------------------------------------------------------------------- |
| `multi-operator-manager`  | The MOM tool itself — runs the tests, builds inputs from must-gather, etc.         |
| `sample-operator`         | A reference **compatible** operator, used to exercise MOM and as a copyable example. |

`multi-operator-manager` is what you run against *your* operators; `sample-operator` is a
minimal operator that already speaks the three verbs, so you can see the whole flow end to
end without wiring up a real operator.

## Command overview

```
multi-operator-manager
├── test
│   └── apply-configuration     # run compatible operators against fixtures and diff output
├── sample-operator             # reference compatible operator
│   ├── input-resources
│   ├── apply-configuration
│   └── output-resources
└── create-input-resources
    └── from-must-gather        # build input resources from a must-gather dump
```

## Testing your compatible operator

```
./multi-operator-manager test apply-configuration \
  --test-dir=./test-data/apply-configuration/ \
  --output-dir=./test-output \
  --preserve-policy=KeepAlways
```

The `./test-output` directory is created and a `junit.xml` inside summarizes the results.

Useful flags for `test apply-configuration`:

- `--test-dir` (required) — directory of tests, searched recursively.
- `--output-dir` (required) — where results (and `junit.xml`) are written.
- `--preserve-policy` — how much of the run output to keep.
- `--replace-expected-output` — delete each test's `expected-output` and replace it with
  the current run's values (use to regenerate the expected output).

### Defining a test

Examples live in `test-data`. You can organize tests however you like: **every directory
containing a `test.yaml` is a test**, and must have an `input-dir` and an `expected-output`
directory. `test.yaml` must name the operator binary under test (`binaryName`).

> TODO: allow a missing `expected-output` to mean "no output". It's painful otherwise.

## The operator contract (verb flags)

Each compatible operator exposes the three verbs. As implemented by the shared libraries:

- `input-resources` / `output-resources` — print the declared resource lists.
- `apply-configuration`:
  - `--input-dir` — directory holding the input resources.
  - `--output-dir` — directory where computed mutations are written.
  - `--controllers` — controllers to enable: `*` (all), `foo` (enable `foo`),
    `-foo` (disable `foo`). Default: `*`.
  - `--now` — the value to use for `time.Now` during the run (for deterministic output).

## Real-world example: cluster-authentication-operator

[cluster-authentication-operator](https://github.com/openshift/cluster-authentication-operator)
is a production operator that adopts the MOM contract. It's a useful reference for how to
make your own operator compatible, because MOM support there is almost entirely **additive** —
the normal operator keeps working unchanged.

It adds a small `mom` command group and reuses its existing controller graph:

- **Wiring** (`cmd/authentication-operator/main.go`) — registers the three verbs as
  subcommands (`apply-configuration`, `input-resources`, `output-resources`).
- **The two declarations** (`pkg/cmd/mom/input_resources_command.go`,
  `output_resources_command.go`) — thin wrappers around the MOM libraries that declare what
  the operator reads and what it is allowed to write (the latter partitioned into
  configuration / management / user-workload resources).
- **The run** (`pkg/cmd/mom/apply_configuration_command.go`) — builds the operator from the
  MOM-supplied input and calls `RunOnce`, returning desired mutations instead of applying
  them.
- **The shared core** (`pkg/operator/replacement_starter.go`) — the key design point: one
  controller graph serves both modes. A production constructor uses real clients; a MOM
  constructor swaps in the mutation-tracking client and a static feature-gate accessor (no
  live cluster to observe). Both register each controller in two forms — a long-running
  `Run` for production and a single-shot `Sync` for MOM's `RunOnce`.

The pattern to copy for your own operator:

1. Refactor operator construction so one starter exposes both `Run` (continuous) and `Sync`
   (run-once) forms of each controller.
2. Add a MOM input constructor that swaps real clients for the mutation-tracking client.
3. Add three thin commands that declare inputs, declare outputs, and call `RunOnce`.

## Building input resources from a must-gather

Replay what an operator *would* do against a real cluster snapshot. Supply the operator's
declared inputs as a "pertinent resources" file (its `input-resources` output), and the
command extracts exactly those objects from the must-gather:

```
<operator-binary> input-resources > pertinent-resources.yaml

./multi-operator-manager create-input-resources from-must-gather \
  --must-gather-dir=<dir> \
  --input-resources=pertinent-resources.yaml \
  --output-dir=<dir>
```

- `--must-gather-dir` — location of the must-gather output.
- `--input-resources` — file listing the pertinent resources to extract (the operator's
  `input-resources` output).
- `--output-dir` — where the minimal output is written.
- `--operator-binary` — intended to derive the pertinent resources by calling the operator
  directly, but **not yet implemented** (currently errors); use `--input-resources` instead.

## Testing this repo

```
make test-operator-integration
```

runs the `sample-operator` against the local test data in `test-data`.
