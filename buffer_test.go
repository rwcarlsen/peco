package peco

import (
	"io/ioutil"
	"strings"
	"testing"
)

func TestInputReaderToRawLineBuffer(t *testing.T) {
	buf := strings.NewReader(`
1. Foo
2. Bar
3. Baz
`)
	rdr := NewInputReader(ioutil.NopCloser(buf))
	rawbuf := NewRawLineBuffer()

	go rdr.Loop()

	<-rdr.ReadyCh()

	rawbuf.Accept(rdr)

	for l := range rawbuf.OutputCh() {
		t.Logf("Received new line %#v", l)
	}

	if rawbuf.Size() != 3 {
		t.Errorf("Expected 3 entries in RawLineBuffer, got %d", rawbuf.Size())
	}
}

func TestInputReaderToRawLineBufferToRegexpFilter(t *testing.T) {
	buf := strings.NewReader(`
1. Foo
2. Bar
3. Baz
`)
	rdr := NewInputReader(ioutil.NopCloser(buf))
	rawbuf := NewRawLineBuffer()

	go rdr.Loop()

	<-rdr.ReadyCh()

	rawbuf.Accept(rdr)
	rf := RegexpFilter{
		flags: regexpFlagList(ignoreCaseFlags),
		query: `\d\. b`,
	}
	rf.Accept(rawbuf)

	flb := NewRawLineBuffer()

	flb.Accept(rf)

	for l := range flb.OutputCh() {
		t.Logf("Received new line %#v", l)
	}

	if flb.Size() != 2 {
		t.Errorf("Expected 2 entries in RawLineBuffer, got %d", flb.Size())
	}
}

func TestBuffer(t *testing.T) {
	rawbuf := NewRawLineBuffer()
	for _, l := range []string{"Alice", "Bob", "Charlie"} {
		rawbuf.AppendLine(NewRawLine(l, false))
	}

	if rawbuf.Size() != 3 {
		t.Errorf("Expected to read 3 lines, got %d", rawbuf.Size())
	}

	f := RegexpFilter{
		flags: regexpFlagList(ignoreCaseFlags),
		query: `c`,
	}

	rawbuf.Replay()
	done := make(chan struct{})
	f.Accept(rawbuf)

	buf := NewRawLineBuffer()
	buf.onEnd = func() { done<-struct{}{} }
	buf.Accept(f)

	for loop := true; loop; {
		select {
		case <-done:
			loop = false
		case <-buf.outputCh:
		}
	}

	if buf.Size() != 2 {
		t.Errorf("Expected to match 2 lines, got %d", buf.Size())
	}

	for i, v := range []string{"Alice", "Charlie"} {
		l, err := buf.LineAt(i)
		if err != nil {
			t.Errorf("Failed to get line at %d: %s", i, err)
			continue
		}

		if l.DisplayString() != v {
			t.Errorf("Expected filtered output at %d to be '%s', got '%s'", i, v, l.DisplayString())
		}
	}
}

func TestBufferPaging(t *testing.T) {
	rawbuf := NewRawLineBuffer()
	for _, l := range []string{"Alice", "Bob", "Charlie", "David", "Eve", "Frank", "George", "Hugh"} {
		rawbuf.AppendLine(NewRawLine(l, false))
	}

	pc := PageCrop{perPage: 4, currentPage: 2}
	pagebuf := pc.Crop(rawbuf)

	for i, v := range []string{"Eve", "Frank", "George", "Hugh"} {
		l, err := pagebuf.LineAt(i)
		if err != nil {
			t.Errorf("Failed to get line at %d: %s", i, err)
			continue
		}

		if l.DisplayString() != v {
			t.Errorf("Expected filtered output at %d to be '%s', got '%s'", i, v, l.DisplayString())
		}
	}

	rawbuf.Replay()

	// Also test regexp filter + paging
	rf := RegexpFilter{
		flags: regexpFlagList(ignoreCaseFlags),
		query: `a`,
	}
	pc.perPage = 2

	rf.Accept(rawbuf)

	done := make(chan struct{})
	buf := NewRawLineBuffer()
	buf.onEnd = func() { done <- struct{}{} }
	buf.Accept(rf)

	for loop := true; loop; {
		select {
		case <-done:
			loop = false
		case <-buf.outputCh:
		}
	}

	pagebuf = pc.Crop(buf)

	for i, v := range []string{"David", "Frank"} {
		l, err := pagebuf.LineAt(i)
		if err != nil {
			t.Errorf("Failed to get line at %d: %s", i, err)
			continue
		}

		if l.DisplayString() != v {
			t.Errorf("Expected filtered output at %d to be '%s', got '%s'", i, v, l.DisplayString())
		}
	}
}
