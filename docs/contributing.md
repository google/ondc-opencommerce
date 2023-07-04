# How to Contribute

We would love to accept your patches and contributions to this project.

## Before you begin

### Sign our Contributor License Agreement

Contributions to this project must be accompanied by a
[Contributor License Agreement](https://cla.developers.google.com/about) (CLA).
You (or your employer) retain the copyright to your contribution; this simply
gives us permission to use and redistribute your contributions as part of the
project.

If you or your current employer have already signed the Google CLA (even if it
was for a different project), you probably don't need to do it again.

Visit <https://cla.developers.google.com/> to see your current agreements or to
sign a new one.

### Review our Community Guidelines

This project follows [Google's Open Source Community
Guidelines](https://opensource.google/conduct/).

## Contribution process

### Code Reviews

All submissions, including submissions by project members, require review. We
use [GitHub pull requests](https://docs.github.com/articles/about-pull-requests)
for this purpose.

## Testing

We use `bazel` to test our services. You can test all services by running this command.
`bazel test //... --test_output=errors --define DOCKER_REGISTRY="" --define DOCKER_REPOSITORY=""`

you can perform the test on a specific target by running `bazel test //<PATH_TO_BUILD_FILE>:<TARGET>`.
For example,
```
bazel test //buyer-platform/bap-api:all
```

### Note

Some tests require a Spanner Emulator running on your local machine. To start Spanner Emulator locally, see [Emulate Cloud Spanner locally](https://cloud.google.com/spanner/docs/emulator).

By default, our test functions communicate with Spanner Emulator on port `9010` via `gRPC` protocol. If you specify a custom port when running Spanner Emulator, you need to set `SPANNER_EMULATOR_ADDRESS` test environment vari	able to that address when running `blaze test`.
For example,
`bazel test --test_env=SPANNER_EMULATOR_ADDRESS=localhost:9010 //…`
