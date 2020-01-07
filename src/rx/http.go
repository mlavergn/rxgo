package rx

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"time"
)

// Request type
type Request struct {
	client *http.Client
}

// NewRequest init
func NewRequest(timeout time.Duration) *Request {
	rootCAs, _ := x509.SystemCertPool()
	if rootCAs == nil {
		rootCAs = x509.NewCertPool()
	}

	// currently based on Linux CA location
	caCert, err := ioutil.ReadFile("/etc/ssl/ca-bundle.crt")
	if err == nil {
		rootCAs.AppendCertsFromPEM(caCert)
	}

	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
		RootCAs:            rootCAs,
	}

	httpTransport := &http.Transport{
		TLSClientConfig: tlsConfig,
		DialContext: (&net.Dialer{
			Timeout: timeout,
		}).DialContext,
	}
	return &Request{
		client: &http.Client{
			Transport: httpTransport,
		},
	}
}

// subject export
func (id *Request) subject(url string, mime string, data []byte, delimiter byte, parser func(*Observable, interface{})) (*Observable, error) {
	var req *http.Request
	var err error
	if data != nil {
		req, err = http.NewRequest(http.MethodPost, url, bytes.NewBuffer(data))
	} else {
		req, err = http.NewRequest(http.MethodGet, url, nil)
	}

	if err != nil {
		log.Println(err)
		return nil, err
	}

	req.Header.Add("Accept", mime)

	// perform the request
	resp, err := id.client.Do(req)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	// log.Println(resp.Header)
	contentLength := resp.ContentLength
	if contentLength == 0 {
		contentLength = -1
	}

	subject := NewSubject()
	reader := bufio.NewReader(resp.Body)

	go func(contentLength int64) {
		defer func() {
			subject.Complete <- true
		}()

		// wait for connect
		<-subject.Connect

	loop:
		for {
			select {
			default:
				if delimiter == 0 {
					data, err := ioutil.ReadAll(reader)
					if err != nil {
						log.Println(err)
						subject.Error <- err
						return
					}
					parser(subject, data)
					subject.Delay(1)
					break loop
				} else {
					chunk, err := reader.ReadBytes(delimiter)
					if err != nil && err != io.EOF {
						log.Println(err)
						subject.Error <- err
						return
					}
					chunkLength := int64(len(chunk))
					if chunkLength != 0 {
						parser(subject, chunk)
						subject.Delay(1)
						if contentLength > 0 {
							contentLength -= chunkLength
							if contentLength <= 0 {
								return
							}
						}
					}
					// end of read
					if err == io.EOF {
						return
					}
				}
			}
		}
	}(contentLength)

	return subject, nil
}

// ByteSubject export
func (id *Request) ByteSubject(url string, contentType string, payload []byte) (*Observable, error) {
	retryFn := func() (*Observable, error) {
		return id.subject(url, contentType, payload, 0, func(subject *Observable, raw interface{}) {
			subject.Next <- raw
		})
	}
	subject, err := retryFn()
	if subject != nil {
		subject.retry = retryFn
	}
	return subject, err
}

// TextSubject export
func (id *Request) TextSubject(url string, payload []byte) (*Observable, error) {
	retryFn := func() (*Observable, error) {
		return id.subject(url, "text/plain", payload, 0, func(subject *Observable, raw interface{}) {
			text := ToByteString(raw, "")
			subject.Next <- text
		})
	}
	subject, err := retryFn()
	if subject != nil {
		subject.retry = retryFn
	}
	return subject, err

}

// LineSubject export
func (id *Request) LineSubject(url string, payload []byte) (*Observable, error) {
	retryFn := func() (*Observable, error) {
		return id.subject(url, "text/plain", payload, byte('\n'), func(subject *Observable, raw interface{}) {
			subject.Next <- raw
		})
	}
	subject, err := retryFn()
	if subject != nil {
		subject.retry = retryFn
	}
	return subject, err

}

// JSONSubject export
func (id *Request) JSONSubject(url string, payload []byte) (*Observable, error) {
	retryFn := func() (*Observable, error) {
		return id.subject(url, "application/json", payload, 0, func(subject *Observable, raw interface{}) {
			data := ToByteArray(raw, nil)
			var result interface{}
			err := json.Unmarshal(data, &result)
			if err != nil {
				subject.Error <- err
			}
			subject.Next <- result
		})
	}
	subject, err := retryFn()
	if subject != nil {
		subject.retry = retryFn
	}
	return subject, err

}

// SSESubject export
func (id *Request) SSESubject(url string, payload []byte) (*Observable, error) {
	retryFn := func() (*Observable, error) {
		// assumption, events will never exceed 10 lines
		lines := [10][]byte{}
		i := 0

		return id.subject(url, "text/event-stream", payload, byte('\n'), func(subject *Observable, raw interface{}) {
			line := ToByteArray(raw, nil)
			if len(line) == 1 || i == 10 {
				// take a reference to lines
				buffer := lines
				subject.Next <- buffer[:i]
				i = 0
			} else {
				lines[i] = line
				i++
			}
		})
	}
	subject, err := retryFn()
	if subject != nil {
		subject.retry = retryFn
	}
	return subject, err

}
