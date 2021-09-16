package multipart

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/textproto"
)

var ErrClosed = func(err error) error {
	return fmt.Errorf("pipe is closed before writing %s", err)
}

type MultipartReader struct {
	src           ReadSeekCloser
	outputHeaders bool
	contentLen    int64
	parts         []*Part
	boundary      string
	w             *io.PipeWriter
	r             *io.PipeReader
	transformer   *Transformer
}

func NewMultipartReader(src ReadSeekCloser, parts []*Part) (*MultipartReader, error) {
	return NewMultipartReaderWithBoudary(src, parts, randomBoundary())
}

func NewMultipartReaderWithBoudary(src ReadSeekCloser, parts []*Part, boundary string) (*MultipartReader, error) {
	r, w := io.Pipe()
	mpReader := &MultipartReader{
		src:         src,
		parts:       parts,
		boundary:    boundary,
		w:           w,
		r:           r,
		transformer: NewTransformerWithBoundary(src, parts, boundary),
	}

	switch len(parts) {
	case 0:
		return nil, errors.New("no part to write")
	case 1:
		// rw.reader, rw.pw = io.Pipe()
		mpReader.contentLen = int64(parts[0].rangeEndInt - parts[0].rangeStartInt + 1)
	default:
		// rw.reader, rw.mw = mw, mw
		mpReader.contentLen = mpReader.transformer.ContentLength()
	}

	return mpReader, nil
}

func (mr *MultipartReader) ContentLength() int64 {
	return mr.contentLen
}

func (mr *MultipartReader) SetOutputHeaders(val bool) {
	mr.outputHeaders = val
}

func (mr *MultipartReader) Start() {
	var err error
	headerBuf := new(bytes.Buffer)

	if len(mr.parts) == 1 {
		if mr.outputHeaders {
			if err := writeStatus(headerBuf, 206); err != nil {
				mr.w.CloseWithError(err)
			}

			headers := textproto.MIMEHeader{}
			headers.Add("Content-Range", fmt.Sprintf("bytes %s-%s/%s", mr.parts[0].rangeStart, mr.parts[0].rangeEnd, mr.parts[0].fileSize))
			if err = writeHeaders(headerBuf, headers); err != nil {
				mr.w.CloseWithError(err)
			}
		}

		_, err = io.Copy(mr.w, headerBuf)
		if err != nil {
			mr.w.CloseWithError(err)
			return
		}
		_, err = mr.w.Write([]byte("\r\n"))
		if err != nil {
			mr.w.CloseWithError(err)
		}
		if err = writePartBody(mr.src, mr.w, mr.parts[0]); err != nil {
			mr.w.CloseWithError(err)
			return
		}
	} else {
		if mr.outputHeaders {
			if err := writeStatus(headerBuf, 206); err != nil {
				mr.w.CloseWithError(err)
			}

			headers := textproto.MIMEHeader{}
			headers.Add("Content-Type", fmt.Sprintf("multipart/byteranges; boundary=%s", mr.boundary))
			if err = writeHeaders(headerBuf, headers); err != nil {
				mr.w.CloseWithError(err)
			}
		}

		_, err = io.Copy(mr.w, headerBuf)
		if err != nil {
			mr.w.CloseWithError(err)
			return
		}

		if err = mr.transformer.WriteMultiParts(mr.w); err != nil {
			mr.w.CloseWithError(err)
			return
		}
	}

	// source file should be closed by user
	mr.w.CloseWithError(nil) // TODO: log error
}

func (mr *MultipartReader) Read(p []byte) (n int, err error) {
	return mr.r.Read(p)
}

func (mr *MultipartReader) Close() error {
	return mr.r.Close()
}
