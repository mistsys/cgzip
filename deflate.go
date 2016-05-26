package cgzip

import "io"

func NewFlateWriter(w io.Writer) *Writer {
	z, _ := NewFlateWriterLevelBuffer(w, Z_DEFAULT_COMPRESSION, DEFAULT_COMPRESSED_BUFFER_SIZE)
	return z
}

func NewFlateWriterLevel(w io.Writer, level int) (*Writer, error) {
	return NewFlateWriterLevelBuffer(w, level, DEFAULT_COMPRESSED_BUFFER_SIZE)
}

func NewFlateWriterLevelBuffer(w io.Writer, level, bufferSize int) (*Writer, error) {
	z := &Writer{w: w, out: make([]byte, bufferSize)}
	err := z.strm.deflateInitWindowBits(
		level,
		-15, // Raw deflate, largest window size
	)
	if err != nil {
		return nil, err
	}
	return z, nil
}
