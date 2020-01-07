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
func (id *Request) subject(url string, mime string, data []byte, delimiter byte) (*Observable, error) {
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

	subject := NewSubject()

	go func() {
		defer func() {
			recover()
		}()

		// wait for connect
		<-subject.Connect

		// perform the request
		resp, err := id.client.Do(req)
		if err != nil {
			subject.Error <- err
		}

		// log.Println(resp.Header)
		contentLength := resp.ContentLength
		if contentLength == 0 {
			contentLength = -1
		}
		reader := bufio.NewReader(resp.Body)

		for {
			select {
			default:
				if delimiter == 0 {
					data, err := ioutil.ReadAll(reader)
					if err != nil {
						subject.Error <- err
						return
					}
					subject.Next <- data
					subject.Delay(1)
					subject.Complete <- true
					return
				}
				chunk, err := reader.ReadBytes(delimiter)
				if err != nil && err != io.EOF {
					subject.Error <- err
					return
				}
				chunkLength := int64(len(chunk))
				if chunkLength != 0 {
					subject.Next <- chunk
					subject.Delay(1)
					if contentLength > 0 {
						contentLength -= chunkLength
						if contentLength <= 0 {
							subject.Complete <- true
							return
						}
					}
				}
				// end of read
				if err == io.EOF {
					subject.Complete <- true
					return
				}
			}
		}
	}()

	return subject, nil
}

// ByteSubject export
func (id *Request) ByteSubject(url string, contentType string, payload []byte) (*Observable, error) {
	subject := NewSubject()

	subject.retryFn = func(observer *Observable) {
		httpSubject, err := id.subject(url, contentType, payload, 0)
		if err != nil {
			observer.Error <- err
			return
		}
		httpSubject.Pipe(observer)
	}
	subject.retryFn(subject)

	return subject, nil
}

// TextSubject export
func (id *Request) TextSubject(url string, payload []byte) (*Observable, error) {
	subject := NewSubject()

	subject.retryFn = func(observer *Observable) {
		httpSubject, err := id.subject(url, "text/plain", payload, 0)
		if err != nil {
			observer.Error <- err
			return
		}
		httpSubject.Map(func(event interface{}) interface{} {
			return ToByteString(event, "")
		})
		httpSubject.Pipe(observer)
	}
	subject.retryFn(subject)

	return subject, nil
}

// LineSubject export
func (id *Request) LineSubject(url string, payload []byte) (*Observable, error) {
	subject := NewSubject()

	subject.retryFn = func(observer *Observable) {
		httpSubject, err := id.subject(url, "text/plain", payload, byte('\n'))
		if err != nil {
			observer.Error <- err
			return
		}
		httpSubject.Pipe(observer)
	}
	subject.retryFn(subject)

	return subject, nil
}

// JSONSubject export
func (id *Request) JSONSubject(url string, payload []byte) (*Observable, error) {
	subject := NewSubject()

	subject.retryFn = func(observer *Observable) {
		httpSubject, err := id.subject(url, "application/json", payload, 0)
		if err != nil {
			observer.Error <- err
			return
		}
		httpSubject.Map(func(event interface{}) interface{} {
			data := ToByteArray(event, nil)
			var result interface{}
			err := json.Unmarshal(data, &result)
			if err != nil {
				subject.Error <- err
				return nil
			}
			return result
		})
		httpSubject.Pipe(observer)
	}
	subject.retryFn(subject)

	return subject, nil
}

// SSESubject export
func (id *Request) SSESubject(url string, payload []byte) (*Observable, error) {
	subject := NewSubject()

	subject.retryFn = func(observer *Observable) {
		// assumption, events will never exceed 10 lines
		lines := [10][]byte{}
		i := 0

		httpSubject, err := id.subject(url, "text/event-stream", payload, byte('\n'))
		if err != nil {
			observer.Error <- err
			return
		}
		httpSubject.Map(func(event interface{}) interface{} {
			line := ToByteArray(event, nil)
			if len(line) == 1 || i == 10 {
				// take a reference to lines
				buffer := lines[:i]
				i = 0
				return buffer
			}
			lines[i] = line
			i++
			return nil
		})
		httpSubject.Pipe(observer)
	}
	subject.retryFn(subject)

	return subject, nil
}
