# Changelog

### v0.8.0

- Add the `Zero()` method to the `Asserts` type

### v0.7.0

- Migrate to Go 1.18

### v0.6.5

- Fix checking for nil error in `Asserts.ErrorContains()`
- Fix generator to resist concurrent calls

### v0.6.4

- Web simulator now provides some convenience methods

### v0.6.3

- Web package now uses `httptest` package

### v0.6.2

- Migrate web response to `http.Response`
- Add helper for request body

### v0.6.1

- Change web response to use the standard `http.Response`
- Add helper for response body

### v0.6.0

- New `web` package for handler tests

### v0.5.2

- Optimize output of last change

### v0.5.1

- Fix output in case of failing web body assertions

### v0.5.0

- `Asserts` now contains `NotOK()` and `AnyError()` for error handling
- `Asserts` created with `NewTesting()` now uses the `Failable` as `Printer`
- `Generator` now contains `OneOf()`

### v0.4.0

- `Asserts` now contains `NotPanics()` and `PanicsWith()`
- `Asserts.OK()` now also handles `func() error`
- Migrate Tester into private helper functions
- Drop unused output in `Asserts` unit test

### v0.3.4

- Length tester now counts runes instead of bytes in case of strings

### v0.3.3

- Extend `Asserts.OK()` for more types
- Extend `Asserts.NoError()` to not only check for nil but also in case of a `T.Err()` instance to return no error
- Same for `Asserts.ErrorMatch()` and `Asserts.ErrorContains()`

### v0.3.2

- Add `Asserts.OK()` as a simple alias for `Asserts.True()`
- Fix the public embedding of the Tester to Asserts

### v0.3.1

- Fix output of `Asserts.ErrorMatch()`

### v0.3.0

- Extracted from Tideland Go Library as part of splitting the GoLib
