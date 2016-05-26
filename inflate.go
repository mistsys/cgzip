package cgzip

import "io"

func NewFlateReader(r io.Reader) (io.ReadCloser, error) {
	return NewFlateReaderBuffer(r, DEFAULT_COMPRESSED_BUFFER_SIZE)
}

func NewFlateReaderBuffer(r io.Reader, bufferSize int) (io.ReadCloser, error) {
	z := &reader{r: r, in: make([]byte, bufferSize)}
	err := z.strm.inflateInitWindowBits(
		-15, // Expect no header or trailer
	)
	if err != nil {
		return nil, err
	}
	return z, nil
}
