// Tideland Go Audit - Generators - Unit Tests
//
// Copyright (C) 2013-2021 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the New BSD license.

package generators_test

//--------------------
// IMPORTS
//--------------------

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"tideland.dev/go/audit/asserts"
	"tideland.dev/go/audit/generators"
)

//--------------------
// TESTS
//--------------------

// TestBuildDate tests the generation of dates.
func TestBuildDate(t *testing.T) {
	assert := asserts.NewTesting(t, asserts.FailStop)
	layouts := []string{
		time.ANSIC,
		time.UnixDate,
		time.RubyDate,
		time.RFC822,
		time.RFC822Z,
		time.RFC850,
		time.RFC1123,
		time.RFC1123Z,
		time.RFC3339,
		time.RFC3339Nano,
		time.Kitchen,
		time.Stamp,
		time.StampMilli,
		time.StampMicro,
		time.StampNano,
	}

	for _, layout := range layouts {
		ts, t := generators.BuildTime(layout, 0)
		tsp, err := time.Parse(layout, ts)
		assert.Nil(err)
		assert.Equal(t, tsp)

		ts, t = generators.BuildTime(layout, -30*time.Minute)
		tsp, err = time.Parse(layout, ts)
		assert.Nil(err)
		assert.Equal(t, tsp)

		ts, t = generators.BuildTime(layout, time.Hour)
		tsp, err = time.Parse(layout, ts)
		assert.Nil(err)
		assert.Equal(t, tsp)
	}
}

// TestBytes tests the generation of bytes.
func TestBytes(t *testing.T) {
	assert := asserts.NewTesting(t, asserts.FailStop)
	gen := generators.New(generators.FixedRand())

	// Test individual bytes.
	for i := 0; i < 10000; i++ {
		lo := gen.Byte(0, 255)
		hi := gen.Byte(0, 255)
		n := gen.Byte(lo, hi)
		if hi < lo {
			lo, hi = hi, lo
		}
		assert.True(lo <= n && n <= hi)
	}

	// Test byte slices.
	ns := gen.Bytes(1, 200, 1000)
	assert.Length(ns, 1000)
	for _, n := range ns {
		assert.True(n >= 1 && n <= 200)
	}

	// Test UUIDs.
	for i := 0; i < 10000; i++ {
		uuid := gen.UUID()
		assert.Length(uuid, 16)
	}
}

// TestInts tests the generation of ints.
func TestInts(t *testing.T) {
	assert := asserts.NewTesting(t, asserts.FailStop)
	gen := generators.New(generators.FixedRand())

	// Test individual ints.
	for i := 0; i < 10000; i++ {
		lo := gen.Int(-100, 100)
		hi := gen.Int(-100, 100)
		n := gen.Int(lo, hi)
		if hi < lo {
			lo, hi = hi, lo
		}
		assert.True(lo <= n && n <= hi)
	}

	// Test int slices.
	ns := gen.Ints(0, 500, 10000)
	assert.Length(ns, 10000)
	for _, n := range ns {
		assert.True(n >= 0 && n <= 500)
	}

	// Test the generation of percent.
	for i := 0; i < 10000; i++ {
		p := gen.Percent()
		assert.True(p >= 0 && p <= 100)
	}

	// Test the flipping of coins.
	ct := 0
	cf := 0
	for i := 0; i < 10000; i++ {
		c := gen.FlipCoin(50)
		if c {
			ct++
		} else {
			cf++
		}
	}
	assert.About(float64(ct), float64(cf), 500)
}

// TestOneOf tests the generation of selections.
func TestOneOf(t *testing.T) {
	assert := asserts.NewTesting(t, asserts.FailStop)
	gen := generators.New(generators.FixedRand())
	stuff := []interface{}{1, true, "three", 47.11, []byte{'A', 'B', 'C'}}

	for i := 0; i < 10000; i++ {
		m := gen.OneOf(stuff...)
		assert.Contains(m, stuff)

		b := gen.OneByteOf(1, 2, 3, 4, 5)
		assert.True(b >= 1 && b <= 5)

		r := gen.OneRuneOf("abcdef")
		assert.True(r >= 'a' && r <= 'f')

		n := gen.OneIntOf(1, 2, 3, 4, 5)
		assert.True(n >= 1 && n <= 5)

		s := gen.OneStringOf("one", "two", "three", "four", "five")
		assert.Substring(s, "one/two/three/four/five")

		d := gen.OneDurationOf(1*time.Second, 2*time.Second, 3*time.Second)
		assert.True(d >= 1*time.Second && d <= 3*time.Second)
	}
}

// TestWords tests the generation of words.
func TestWords(t *testing.T) {
	assert := asserts.NewTesting(t, asserts.FailStop)
	gen := generators.New(generators.FixedRand())

	// Test single words.
	for i := 0; i < 10000; i++ {
		w := gen.Word()
		for _, r := range w {
			assert.True(r >= 'a' && r <= 'z')
		}
	}

	// Test limited words.
	for i := 0; i < 10000; i++ {
		lo := gen.Int(generators.MinWordLen, generators.MaxWordLen)
		hi := gen.Int(generators.MinWordLen, generators.MaxWordLen)
		w := gen.LimitedWord(lo, hi)
		wl := len(w)
		if hi < lo {
			lo, hi = hi, lo
		}
		assert.True(lo <= wl && wl <= hi, info("WL %d LO %d HI %d", wl, lo, hi))
	}
}

// TestPattern tests the generation based on patterns.
func TestPattern(t *testing.T) {
	assert := asserts.NewTesting(t, asserts.FailStop)
	gen := generators.New(generators.FixedRand())
	assertPattern := func(pattern, runes string) {
		set := make(map[rune]bool)
		for _, r := range runes {
			set[r] = true
		}
		for i := 0; i < 10; i++ {
			result := gen.Pattern(pattern)
			for _, r := range result {
				assert.True(set[r], pattern, result, runes)
			}
		}
	}

	assertPattern("^^", "^")
	assertPattern("^0^0^0^0^0", "0123456789")
	assertPattern("^1^1^1^1^1", "123456789")
	assertPattern("^o^o^o^o^o", "01234567")
	assertPattern("^h^h^h^h^h", "0123456789abcdef")
	assertPattern("^H^H^H^H^H", "0123456789ABCDEF")
	assertPattern("^a^a^a^a^a", "abcdefghijklmnopqrstuvwxyz")
	assertPattern("^A^A^A^A^A", "ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	assertPattern("^c^c^c^c^c", "bcdfghjklmnpqrstvwxyz")
	assertPattern("^C^C^C^C^C", "BCDFGHJKLMNPQRSTVWXYZ")
	assertPattern("^v^v^v^v^v", "aeiou")
	assertPattern("^V^V^V^V^V", "AEIOU")
	assertPattern("^z^z^z^z^z", "abcdefghijklmnopqrstuvwxyz0123456789")
	assertPattern("^Z^Z^Z^Z^Z", "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	assertPattern("^1^0.^0^0^0,^0^0 €", "0123456789 .,€")
}

// TestText tests the generation of text.
func TestText(t *testing.T) {
	assert := asserts.NewTesting(t, asserts.FailStop)
	gen := generators.New(generators.FixedRand())
	names := gen.Names(4)

	for i := 0; i < 10000; i++ {
		s := gen.Sentence()
		ws := strings.Split(s, " ")
		lws := len(ws)
		assert.True(2 <= lws && lws <= 15, info("S: %v SL: %d", s, lws))
		assert.True('A' <= s[0] && s[0] <= 'Z', info("SUC: %v", s[0]))
	}

	for i := 0; i < 10; i++ {
		s := gen.SentenceWithNames(names)
		assert.NotEmpty(s)
	}

	for i := 0; i < 10000; i++ {
		p := gen.Paragraph()
		ss := strings.Split(p, ". ")
		lss := len(ss)
		assert.True(2 <= lss && lss <= 10, info("PL: %d", lss))
		for _, s := range ss {
			ws := strings.Split(s, " ")
			lws := len(ws)
			assert.True(2 <= lws && lws <= 15, info("S: %v PSL: %d", s, lws))
			assert.True('A' <= s[0] && s[0] <= 'Z', info("PSUC: %v", s[0]))
		}
	}

	for i := 0; i < 10; i++ {
		s := gen.ParagraphWithNames(names)
		assert.NotEmpty(s)
	}
}

// TestName tests the generation of names.
func TestName(t *testing.T) {
	assert := asserts.NewTesting(t, asserts.FailStop)
	gen := generators.New(generators.FixedRand())

	assert.Equal(generators.ToUpperFirst("yadda"), "Yadda")

	for i := 0; i < 10000; i++ {
		first, middle, last := gen.Name()

		assert.Match(first, `[A-Z][a-z]+(-[A-Z][a-z]+)?`)
		assert.Match(middle, `[A-Z][a-z]+(-[A-Z][a-z]+)?`)
		assert.Match(last, `[A-Z]['a-zA-Z]+`)

		first, middle, last = gen.MaleName()

		assert.Match(first, `[A-Z][a-z]+(-[A-Z][a-z]+)?`)
		assert.Match(middle, `[A-Z][a-z]+(-[A-Z][a-z]+)?`)
		assert.Match(last, `[A-Z]['a-zA-Z]+`)

		first, middle, last = gen.FemaleName()

		assert.Match(first, `[A-Z][a-z]+(-[A-Z][a-z]+)?`)
		assert.Match(middle, `[A-Z][a-z]+(-[A-Z][a-z]+)?`)
		assert.Match(last, `[A-Z]['a-zA-Z]+`)

		count := gen.Int(0, 5)
		names := gen.Names(count)

		assert.Length(names, count)
		for _, name := range names {
			assert.Match(name, `[A-Z][a-z]+(-[A-Z][a-z]+)?\s([A-Z]\.\s)?[A-Z]['a-zA-Z]+`)
		}
	}
}

// TestDomain tests the generation of domains.
func TestDomain(t *testing.T) {
	assert := asserts.NewTesting(t, asserts.FailStop)
	gen := generators.New(generators.FixedRand())

	for i := 0; i < 00100; i++ {
		domain := gen.Domain()

		assert.Match(domain, `^[a-z0-9.-]+\.[a-z]{2,4}$`)
	}
}

// TestURL tests the generation of URLs.
func TestURL(t *testing.T) {
	assert := asserts.NewTesting(t, asserts.FailStop)
	gen := generators.New(generators.FixedRand())

	for i := 0; i < 10000; i++ {
		url := gen.URL()

		assert.Match(url, `(http|ftp|https):\/\/[\w\-_]+(\.[\w\-_]+)+([\w\-\.,@?^=%&amp;:/~\+#]*[\w\-\@?^=%&amp;/~\+#])?`)
	}
}

// TestEMail tests the generation of e-mail addresses.
func TestEMail(t *testing.T) {
	assert := asserts.NewTesting(t, asserts.FailStop)
	gen := generators.New(generators.FixedRand())

	for i := 0; i < 10000; i++ {
		addr := gen.EMail()

		assert.Match(addr, `^[a-z0-9._%+-]+@[a-z0-9.-]+\.[a-z]{2,4}$`)
	}
}

// TestTimes tests the generation of durations and times.
func TestTimes(t *testing.T) {
	assert := asserts.NewTesting(t, asserts.FailStop)
	gen := generators.New(generators.FixedRand())

	for i := 0; i < 10000; i++ {
		// Test durations.
		lo := gen.Duration(time.Second, time.Minute)
		hi := gen.Duration(time.Second, time.Minute)
		d := gen.Duration(lo, hi)
		if hi < lo {
			lo, hi = hi, lo
		}
		assert.True(lo <= d && d <= hi, "High / Low")

		// Test times.
		loc := time.Local
		now := time.Now()
		dur := gen.Duration(24*time.Hour, 30*24*time.Hour)
		t := gen.Time(loc, now, dur)
		assert.True(t.Equal(now) || t.After(now), "Equal or after now")
		assert.True(t.Before(now.Add(dur)) || t.Equal(now.Add(dur)), "Before or equal now plus duration")
	}

	sleeps := map[int]time.Duration{
		1: 1 * time.Millisecond,
		2: 2 * time.Millisecond,
		3: 3 * time.Millisecond,
		4: 4 * time.Millisecond,
		5: 5 * time.Millisecond,
	}
	for i := 0; i < 1000; i++ {
		sleep := gen.SleepOneOf(sleeps[1], sleeps[2], sleeps[3], sleeps[4], sleeps[5])
		s := int(sleep) / 1000000
		_, ok := sleeps[s]
		assert.True(ok, "Chosen duration is one the arguments")
	}
}

//--------------------
// HELPER
//--------------------

var info = fmt.Sprintf

// EOF
