# Run Tests

```bash
go test -v -failfast -count 1
```

where:
- `-v` is to enable verbose output
- `-failfast` makes the test stop after first failure
- `-count 1` is to avoid Go test caching for example when testing a rebuilt snap

## Environment variables 

Some environment variables can modify the test functionality. Refer to these in
[the documentation](https://pkg.go.dev/github.com/canonical/matter-snap-testing@v1.0.0-beta.3/env)
of the `matter-snap-testing` Go package.
