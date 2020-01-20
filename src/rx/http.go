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
	"strings"
	"sync"
	"time"
)

// HTTPRequest type
type HTTPRequest struct {
	client *http.Client
}

var tlsConfigOnce sync.Once
var tlsConfig *tls.Config

var httpClient *http.Client
var httpClientTimeout time.Duration

// NewHTTPRequest init
func NewHTTPRequest(timeout time.Duration) *HTTPRequest {
	// only need to configure tls.Config once
	tlsConfigOnce.Do(func() {
		rootCAs, _ := x509.SystemCertPool()
		if rootCAs == nil {
			rootCAs = x509.NewCertPool()
		}

		// currently based on Linux CA location
		caCert, err := ioutil.ReadFile("/etc/ssl/ca-bundle.crt")
		if err == nil {
			rootCAs.AppendCertsFromPEM(caCert)
		}

		tlsConfig = &tls.Config{
			InsecureSkipVerify: true,
			RootCAs:            rootCAs,
		}
	})

	// if timeout is unchanged, reuse http.Client
	if timeout != httpClientTimeout {
		httpTransport := &http.Transport{
			TLSClientConfig: tlsConfig,
			DialContext: (&net.Dialer{
				Timeout: timeout,
			}).DialContext,
		}

		httpClient = &http.Client{
			Transport: httpTransport,
		}
	}

	return &HTTPRequest{
		client: httpClient,
	}
}

// httpSubject export
func (id *HTTPRequest) httpSubject(url string, mime string, data []byte, delimiter byte) (*Observable, error) {
	log.Println("HTTPRequest.httpSubject")
	var req *http.Request
	var err error
	if data != nil {
		req, err = http.NewRequest(http.MethodPost, url, bytes.NewBuffer(data))
	} else {
		req, err = http.NewRequest(http.MethodGet, url, nil)
	}

	if err != nil {
		log.Println("HTTPRequest.httpSubject", err)
		return nil, err
	}

	req.Header.Add("Accept", mime)

	subject := NewSubject()
	subject.UID = "httpSubject." + subject.UID

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		wg.Done()
		// wait for connect
		<-subject.connect

		// perform the request
		resp, err := id.client.Do(req)
		if err != nil {
			subject.Error <- err
			return
		}

		// log.Println(resp.Header)
		contentLength := resp.ContentLength
		if contentLength == 0 {
			contentLength = -1
		}
		dlog.Println(subject.UID, "HTTPRequest.httpSubject.ContentLength", contentLength)
		reader := bufio.NewReader(resp.Body)

		for {
			if delimiter == 0 {
				data, err := ioutil.ReadAll(reader)
				if err != nil {
					log.Println(subject.UID, "HTTPRequest.httpSubject.Error", err)
					subject.Error <- err
					return
				}
				subject.Next <- data
				log.Println(subject.UID, "HTTPRequest.httpSubject.Complete via ReadAll")
				subject.Yield()
				subject.Complete <- true
				return
			}
			chunk, err := reader.ReadBytes(delimiter)
			if err != nil && err != io.EOF {
				log.Println(subject.UID, "HTTPRequest.httpSubject.Error", err)
				subject.Error <- err
				return
			}
			chunkLength := int64(len(chunk))
			if chunkLength != 0 {
				dlog.Println(subject.UID, "HTTPRequest.httpSubject.Next")
				subject.Next <- chunk
				if contentLength > 0 {
					contentLength -= chunkLength
					if contentLength <= 0 {
						dlog.Println(subject.UID, "HTTPRequest.httpSubject.Complete via ContentLength")
						subject.Yield()
						subject.Complete <- true
						return
					}
				}
			}
			// end of read
			if err == io.EOF {
				dlog.Println(subject.UID, "HTTPRequest.httpSubject.Complete via EOF")
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
func (id *HTTPRequest) ByteSubject(url string, contentType string, payload []byte) (*Observable, error) {
	log.Println("HTTPRequest.ByteSubject")
	subject := NewSubject()

	subject.Resubscribe(func(observer *Observable) error {
		log.Println(observer.UID, "HTTPRequest.ByteSubject.Resubscribe")
		httpSubject, err := id.httpSubject(url, contentType, payload, 0)
		if err != nil {
			return err
		}
		httpSubject.UID = "ByteSubject." + observer.UID
		httpSubject.Pipe(observer)
		dlog.Println(httpSubject.UID, "HTTPRequest.ByteSubject.Subscribed")
		return nil
	})

	return subject, nil
}

// TextSubject export
func (id *HTTPRequest) TextSubject(url string, payload []byte) (*Observable, error) {
	log.Println("HTTPRequest.TextSubject")
	subject := NewSubject()

	subject.Resubscribe(func(observer *Observable) error {
		log.Println(observer.UID, "HTTPRequest.TextSubject.Resubscribe")
		httpSubject, err := id.httpSubject(url, "text/plain", payload, 0)
		if err != nil {
			return err
		}
		httpSubject.Map(func(event interface{}) interface{} {
			return ToByteString(event, "")
		})
		httpSubject.UID = "TextSubject." + observer.UID
		httpSubject.Pipe(observer)
		dlog.Println(httpSubject.UID, "HTTPRequest.TextSubject.Subscribed")
		return nil
	})

	return subject, nil
}

// LineSubject export
func (id *HTTPRequest) LineSubject(url string, payload []byte) (*Observable, error) {
	log.Println("HTTPRequest.LineSubject")
	subject := NewSubject()

	subject.Resubscribe(func(observer *Observable) error {
		log.Println(observer.UID, "HTTPRequest.LineSubject.Resubscribe")
		httpSubject, err := id.httpSubject(url, "text/plain", payload, byte('\n'))
		if err != nil {
			return err
		}
		httpSubject.UID = "LineSubject." + observer.UID
		httpSubject.Pipe(observer)
		dlog.Println(httpSubject.UID, "HTTPRequest.LineSubject.Subscribed")
		return nil
	})

	return subject, nil
}

// JSONSubject export
func (id *HTTPRequest) JSONSubject(url string, payload []byte) (*Observable, error) {
	log.Println("HTTPRequest.JSONSubject")
	subject := NewSubject()

	subject.Resubscribe(func(observer *Observable) error {
		log.Println(observer.UID, "HTTPRequest.JSONSubject.Resubscribe")
		httpSubject, err := id.httpSubject(url, "application/json", payload, 0)
		if err != nil {
			return err
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
		httpSubject.UID = "JSONSubject." + observer.UID
		httpSubject.Pipe(observer)
		dlog.Println(httpSubject.UID, "HTTPRequest.JSONSubject.Subscribed")
		return nil
	})

	return subject, nil
}

// SSESubject export
func (id *HTTPRequest) SSESubject(url string, payload []byte) (*Observable, error) {
	log.Println("HTTPRequest.SSESubject")
	subject := NewSubject()

	subject.Resubscribe(func(observer *Observable) error {
		log.Println(observer.UID, "HTTPRequest.SSESubject.Resubscribe")
		// assumption, events will never exceed 10 lines
		lines := [10][]byte{}
		i := 0

		httpSubject, err := id.httpSubject(url, "text/event-stream", payload, byte('\n'))
		if err != nil {
			return err
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
		}).Map(func(event interface{}) interface{} {
			if event == nil {
				return nil
			}
			sse := map[string]interface{}{}
			lines := ToByteArrayArray(event, nil)
			for i = 0; i < len(lines); i++ {
				line := string(lines[i])
				split := strings.Index(line, ":")
				sse[line[:split]] = strings.TrimSpace(line[split+1:])
			}
			return sse
		})
		httpSubject.UID = "SSESubject." + observer.UID
		httpSubject.Pipe(observer)
		dlog.Println(httpSubject.UID, "HTTPRequest.SSESubject.Subscribed")
		return nil
	})

	return subject, nil
}
