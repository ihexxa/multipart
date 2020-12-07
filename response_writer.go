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

// type CountablePipeWriter struct {
// 	*io.PipeWriter
// 	wrote int64
// }

// func (w *CountablePipeWriter) Write(b []byte) (int, error) {
// 	wrote, err := w.PipeWriter.Write(b)
// 	w.wrote += int64(wrote)
// 	return wrote, err
// }

// func (w *CountablePipeWriter) Wrote() int64 {
// 	return w.wrote
// }

// type IPipeWriter interface {
// 	Close() error
// 	CloseWithError(err error) error
// 	Write(data []byte) (n int, err error)
// }

type ResponseWriter struct {
	src       io.ReadSeeker
	headerBuf *bytes.Buffer
	parts     []*Part
	reader    io.Reader
	pw        *io.PipeWriter
	mw        *Writer
	// contentLength int64
	// fileName      string
	// boundary      string
	// mw            *Writer
	// pp            *CountablePipeWriter
}

func NewResponseWriter(src io.ReadSeeker, parts []*Part, outputHeaders bool) (*ResponseWriter, int64, error) {
	return NewResponseWriterWithBoudary(src, parts, randomBoundary(), outputHeaders)
}

func NewResponseWriterWithBoudary(src io.ReadSeeker, parts []*Part, boundary string, outputHeaders bool) (*ResponseWriter, int64, error) {
	var (
		err           error
		contentLength int64
	)
	rw := &ResponseWriter{
		src:       src,
		headerBuf: new(bytes.Buffer),
		parts:     parts,
	}

	switch len(parts) {
	case 0:
		return nil, 0, errors.New("no part to write")
	case 1:
		if outputHeaders {
			if err := writeStatus(rw.headerBuf, 206); err != nil {
				return nil, 0, err
			}

			headers := textproto.MIMEHeader{}
			headers.Add("Content-Range", fmt.Sprintf("bytes %s-%s/%s", parts[0].rangeStart, parts[0].rangeEnd, parts[0].fileSize))
			if err = writeHeaders(rw.headerBuf, headers); err != nil {
				return nil, 0, err
			}
		}

		rw.reader, rw.pw = io.Pipe()
		contentLength = int64(parts[0].rangeEndInt - parts[0].rangeStartInt + 1)
	default:
		if outputHeaders {
			if err := writeStatus(rw.headerBuf, 206); err != nil {
				return nil, 0, err
			}

			headers := textproto.MIMEHeader{}
			headers.Add("Content-Type", fmt.Sprintf("multipart/byteranges; boundary=%s", boundary))
			if err = writeHeaders(rw.headerBuf, headers); err != nil {
				return nil, 0, err
			}
		}

		mw := NewWriterWithBoundary(src, parts, boundary)
		rw.reader, rw.mw = mw, mw
		contentLength = mw.ContentLength()
	}

	return rw, contentLength, nil
}

// func (w *ResponseWriter) Write(fileName string) {
// 	w.WriteWithBoundary(fileName, randomBoundary())
// }

func (w *ResponseWriter) Write() {
	var err error
	if len(w.parts) == 1 {
		_, err = io.Copy(w.pw, w.headerBuf)
		if err != nil {
			w.pw.CloseWithError(err)
			return
		}
		_, err = w.pw.Write([]byte("\r\n"))
		if err != nil {
			w.pw.CloseWithError(err)
		}
		if err = writePartBody(w.src, w.pw, w.parts[0]); err != nil {
			w.pw.CloseWithError(err)
			return
		}
		w.pw.Close()
	} else {
		_, err = io.Copy(w.mw, w.headerBuf)
		if err != nil {
			w.pw.CloseWithError(err)
			return
		}
		// for _, part := range w.parts {
		// 	if err = w.mw.CreatePart(part.contentType, part.rangeStart, part.rangeEnd, part.fileSize); err != nil {
		// 		w.mw.CloseWithError(err)
		// 		return
		// 	}
		// 	if err = writePartBody(w.src, w.mw, part); err != nil {
		// 		w.mw.CloseWithError(err)
		// 		return
		// 	}
		// }

		if err = w.mw.WriteMultiParts(); err != nil {
			w.mw.CloseWithError(err)
			return
		}
		w.mw.Close()
	}
}

// if err != io.ErrClosedPipe {
// 	w.pw.CloseWithError(err)
// 	return false
// }
// panic(ErrClosed(err))

func (w *ResponseWriter) Read(p []byte) (n int, err error) {
	if len(w.parts) == 1 {
		return w.reader.Read(p)
	}
	return w.mw.Read(p)
}

// func writeMultiPart(src io.ReadSeeker, mw *Writer, fileName string, parts []*Part, contentLen *int64) {
// 	for _, part := range parts {
// 		if !mw.CreatePart(part.contentType, part.rangeStart, part.rangeEnd, part.fileSize) {
// 			return
// 		}
// 		if !writePartBody(src, mw, part) {
// 			return
// 		}
// 	}

// 	mw.Close()
// }

// func WriteResponse(src io.ReadSeeker, dst *io.PipeWriter, fileName string, parts []*Part, contentLen *int64) {
// 	writer := &CountablePipeWriter{
// 		boundary:   boundary,
// 		PipeWriter: dst,
// 		wrote:      0,
// 	}
// 	WriteResponseWithBoundary(src, writer, fileName, parts, contentLen, randomBoundary())
// }

// func WriteResponseWithBoundary(src io.ReadSeeker, dst *CountablePipeWriter, fileName string, parts []*Part, contentLen *int64, boundary string) {
// 	if len(parts) == 0 {
// 		writeErrResponse(dst, 401, errors.New("no part to write"))
// 		return
// 	} else if len(parts) == 1 {
// 		writeSinglePart(src, dst, fileName, parts[0], contentLen)
// 	}

// 	mw, ok := NewWriterWithBoundary(dst, boundary)
// 	if !ok {
// 		return
// 	}

// 	writeMultiPart(src, mw, fileName, parts, contentLen)
// }

// TODO: Add header: Content-Length
// func writeSinglePart(src io.ReadSeeker, writer *CountablePipeWriter, fileName string, part *Part, contentLen *int64) {
// 	if !writeStatus(writer, 206) {
// 		return
// 	}

// 	headers := textproto.MIMEHeader{}
// 	headers.Add("Content-Range", fmt.Sprintf("bytes %s-%s/%s", part.rangeStart, part.rangeEnd, part.fileSize))
// 	if !writeHeaders(writer, headers) {
// 		return
// 	}

// 	*contentLen = writer.Wrote() + part.rangeEndInt - part.rangeStartInt + 1
// 	if !writePartBody(src, writer, part) {
// 		return
// 	}

// 	writer.Close()
// }
