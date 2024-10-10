Multi-Operator-Manager

# Building
* `make build`


# Testing your compatible operator
`./multi-operator-manager test apply-configuration --test-dir=./test-data/apply-configuration/ --output-dir=../test-output --preserve-policy=KeepAlways`

The `../test-output` directory will be created and a `junit.xml` inside will summarize the results.

## Defining a test
An example is contained in `test-data`.
You can organize your tests however you wish, but every directory with a `test.yaml` is considered a test and must have
an `input-dir` and an `expected-output` dir.

TODO probably allow missing to mean no output.  It's painful otherwise.

### test.yaml
This repo contains examples, but to test your operator the operator binary name must be present.
