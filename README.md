# Elastic APM Cloud Native Buildpack

The buildpack installs the [elastic apm agent](https://www.elastic.co/guide/en/apm/agent/java/current/index.html) in a jvm.

## Usage

To package this buildpack for consumption:

```
$ ./scripts/package.sh --version <version-number>
```

This will create a `buildpackage.cnb` file under the `build` directory which you can use to build your app as follows:
`pack build <app-name> -p <path-to-app> -b build/buildpackage.cnb`

## Run Tests

To run all unit tests, run:
```
./scripts/unit.sh
```

To run all integration tests, run:
```
/scripts/integration.sh
```
