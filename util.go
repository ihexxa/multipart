package multipart

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"io"
	"net/textproto"
	"sort"
)

func writeHeaders(w io.Writer, headers textproto.MIMEHeader) error {
	var buf bytes.Buffer

	keys := make([]string, 0, len(headers))
	for k := range headers {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		for _, v := range headers[k] {
			_, err := fmt.Fprintf(&buf, "%s: %s\r\n", k, v)
			if err != nil {
				return err
			}
		}
	}

	_, err := io.Copy(w, &buf)
	return err
}

func randomBoundary() string {
	var buf [30]byte
	_, err := io.ReadFull(rand.Reader, buf[:])
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("%x", buf[:])
}
