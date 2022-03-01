# Changelog

## v0.6.5

* (C) Fix checking for nil error in Asserts.ErrorContains()
* (C) Fix generator to resist concurrent calls

## v0.6.4

* (C) Web simulator now provides some convenience methods

## v0.6.3

* (C) Web package now uses httptest package

## v0.6.2

* (C) Migrate web response to http.Response
* (A) Add helper for request body

## v0.6.1

* (C) Change web response to use the standard http.Response
* (A) Add helper for response body

## v0.6.0

* (A) New web package for handler tests

## v0.5.2

* (C) Optimize output of last change

## v0.5.1

* (C) Fix output in case of failing web body assertions

## v0.5.0

* (A) Asserts now contains NotOK() and AnyError()
* (C) Asserts created with NewTesting() now uses the Failable as Printer if
      it implements the according interface
* (A) Generator now contains OneOf()

## v0.4.0

* (A) Asserts now contains NotPanics() and PanicsWith()
* (C) Asserts.OK() now also handles func() error
* (C) Migrate Tester into private helper functions
* (D) Drop unused output in Asserts unit test

## v0.3.4

* (C) Length tester now counts runes instead of bytes in case of strings

## v0.3.3

* (C) Extend Asserts.OK() for more types
* (C) Extend Asserts.NoError() to not only check for nil but also in case
  of a T.Err() instance to return no error
* (C) Same for Asserts.ErrorMatch() and Asserts.ErrorContains()

## v0.3.2

* (C) Add Asserts.OK() as a simple alias for Asserts.True()
* (C) Fix the public embedding of the Tester to Asserts

## v0.3.1

* (C) Fix output of Asserts.ErrorMatch()

## v0.3.0

* (C) Extracted from Tideland Go Library as part of split

