# Changelog

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

