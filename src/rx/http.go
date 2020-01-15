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
	"sync"
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

// httpSubject export
func (id *Request) httpSubject(url string, mime string, data []byte, delimiter byte) (*Observable, error) {
	log.Println("httpSubject")
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

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		wg.Done()
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
		dlog.Println(subject.UID, "httpSubject.ContentLength", contentLength)
		reader := bufio.NewReader(resp.Body)

		for {
			if delimiter == 0 {
				data, err := ioutil.ReadAll(reader)
				if err != nil {
					dlog.Println(subject.UID, "httpSubject.Error", err)
					subject.Error <- err
					return
				}
				subject.Next <- data
				dlog.Println(subject.UID, "httpSubject.Complete via ReadAll")
				subject.Yield()
				subject.Complete <- true
				return
			}
			chunk, err := reader.ReadBytes(delimiter)
			if err != nil && err != io.EOF {
				dlog.Println(subject.UID, "httpSubject.Error", err)
				subject.Error <- err
				return
			}
			chunkLength := int64(len(chunk))
			if chunkLength != 0 {
				dlog.Println(subject.UID, "httpSubject.Next")
				subject.Next <- chunk
				if contentLength > 0 {
					contentLength -= chunkLength
					if contentLength <= 0 {
						dlog.Println(subject.UID, "httpSubject.Complete via ContentLength")
						subject.Yield()
						subject.Complete <- true
						return
					}
				}
			}
			// end of read
			if err == io.EOF {
				dlog.Println(subject.UID, "httpSubject.Complete via EOF")
				subject.Yield()
				subject.Complete <- true
				return
			}
		}
	}()

	wg.Wait()
	return subject, nil
}

// ByteSubject export
func (id *Request) ByteSubject(url string, contentType string, payload []byte) (*Observable, error) {
	log.Println("ByteSubject")
	subject := NewSubject()

	subject.resubscribeFn = func(observer *Observable) {
		httpSubject, err := id.httpSubject(url, contentType, payload, 0)
		if err != nil {
			observer.Error <- err
			return
		}
		httpSubject.UID = "httpSubject:" + observer.UID
		httpSubject.Pipe(observer)
	}
	subject.resubscribeFn(subject)

	return subject, nil
}

// TextSubject export
func (id *Request) TextSubject(url string, payload []byte) (*Observable, error) {
	log.Println("TextSubject")
	subject := NewSubject()

	subject.resubscribeFn = func(observer *Observable) {
		httpSubject, err := id.httpSubject(url, "text/plain", payload, 0)
		if err != nil {
			observer.Error <- err
			return
		}
		httpSubject.Map(func(event interface{}) interface{} {
			return ToByteString(event, "")
		})
		httpSubject.UID = "httpSubject:" + observer.UID
		httpSubject.Pipe(observer)
	}
	subject.resubscribeFn(subject)

	return subject, nil
}

// LineSubject export
func (id *Request) LineSubject(url string, payload []byte) (*Observable, error) {
	log.Println("LineSubject")
	subject := NewSubject()

	subject.resubscribeFn = func(observer *Observable) {
		httpSubject, err := id.httpSubject(url, "text/plain", payload, byte('\n'))
		if err != nil {
			observer.Error <- err
			return
		}
		httpSubject.UID = "httpSubject:" + observer.UID
		httpSubject.Pipe(observer)
	}
	subject.resubscribeFn(subject)

	return subject, nil
}

// JSONSubject export
func (id *Request) JSONSubject(url string, payload []byte) (*Observable, error) {
	log.Println("JSONSubject")
	subject := NewSubject()

	subject.resubscribeFn = func(observer *Observable) {
		httpSubject, err := id.httpSubject(url, "application/json", payload, 0)
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
		httpSubject.UID = "httpSubject:" + observer.UID
		httpSubject.Pipe(observer)
	}
	subject.resubscribeFn(subject)

	return subject, nil
}

// SSESubject export
func (id *Request) SSESubject(url string, payload []byte) (*Observable, error) {
	log.Println("SSESubject")
	subject := NewSubject()

	subject.resubscribeFn = func(observer *Observable) {
		// assumption, events will never exceed 10 lines
		lines := [10][]byte{}
		i := 0

		httpSubject, err := id.httpSubject(url, "text/event-stream", payload, byte('\n'))
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
		httpSubject.UID = "httpSubject:" + observer.UID
		httpSubject.Pipe(observer)
	}
	subject.resubscribeFn(subject)

	return subject, nil
}
