// Copyright 2012, Google Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cgzip

import (
	"io"
	"runtime"
)

// err starts out as nil
// we will call inflateEnd when we set err to a value:
// - whatever error is returned by the underlying reader
// - io.EOF if Close was called
type reader struct {
	r      io.Reader
	in     []byte
	strm   zstream
	err    error
	skipIn bool
}

func NewReader(r io.Reader) (io.ReadCloser, error) {
	return NewReaderBuffer(r, DEFAULT_COMPRESSED_BUFFER_SIZE)
}

func NewReaderBuffer(r io.Reader, bufferSize int) (io.ReadCloser, error) {
	z := &reader{r: r, in: make([]byte, bufferSize)}
	if err := z.strm.inflateInit(); err != nil {
		return nil, err
	}
	// make sure we clean up any buffers allocated by the C code if we aren't Closed by the caller
	runtime.SetFinalizer(z, (*reader).Close)
	return z, nil
}

func (z *reader) Read(p []byte) (int, error) {
	if z.err != nil {
		return 0, z.err
	}

	if len(p) == 0 {
		return 0, nil
	}

	// read and deflate until the output buffer is full
	z.strm.setOutBuf(p, len(p))

	for {
		// if we have no data to inflate, read more
		if !z.skipIn && z.strm.availIn() == 0 {
			var n int
			n, z.err = z.r.Read(z.in)
			// If we got data and EOF, pretend we didn't get the
			// EOF.  That way we will return the right values
			// upstream.  Note this will trigger another read
			// later on, that should return (0, EOF).
			if n > 0 && z.err == io.EOF {
				z.err = nil
			}

			// FIXME(alainjobart) this code is not compliant with
			// the Reader interface. We should process all the
			// data we got from the reader, and then return the
			// error, whatever it is.
			if (z.err != nil && z.err != io.EOF) || (n == 0 && z.err == io.EOF) {
				z.strm.inflateEnd()
				runtime.SetFinalizer(z, nil) // finalizer isn't needed once we've called inflateEnd()
				return 0, z.err
			}

			z.strm.setInBuf(z.in, n)
		} else {
			z.skipIn = false
		}

		// inflate some
		ret, err := z.strm.inflate(zNoFlush)
		if err != nil {
			z.err = err
			z.strm.inflateEnd()
			runtime.SetFinalizer(z, nil) // finalizer isn't needed once we've called inflateEnd()
			return 0, z.err
		}

		// if we read something, we're good
		have := len(p) - z.strm.availOut()
		if have > 0 {
			z.skipIn = ret == Z_OK && z.strm.availOut() == 0
			return have, z.err
		}
	}
}

// Close closes the Reader. It does not close the underlying io.Reader.
func (z *reader) Close() error {
	if z.err != nil {
		if z.err != io.EOF {
			return z.err
		}
		return nil
	}
	z.strm.inflateEnd()
	runtime.SetFinalizer(z, nil) // finalizer isn't needed once we've called inflateEnd()
	z.err = io.EOF
	return nil
}
