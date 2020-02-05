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

// HTTPClient type
type HTTPClient struct {
	client *http.Client
}

var tlsConfigOnce sync.Once
var tlsConfig *tls.Config

var httpClient *http.Client
var httpClientTimeout = -1 * time.Second

// NewHTTPClient init
func NewHTTPClient(timeout time.Duration) *HTTPClient {
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

	return &HTTPClient{
		client: httpClient,
	}
}

// subject export
func (id *HTTPClient) subject(url string, mime string, data []byte, delimiter byte) (*Observable, error) {
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
	req.Header.Add("Connection", "close")

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
			subject.onError(err)
			return
		}

		// log.Println(resp.Header)
		contentLength := resp.ContentLength
		if contentLength == 0 {
			contentLength = -1
		}
		dlog.Println(subject.UID, "HTTPRequest.httpSubject.ContentLength", contentLength)
		reader := bufio.NewReader(resp.Body)
		defer resp.Body.Close()

		for {
			if delimiter == 0 {
				data, err := ioutil.ReadAll(reader)
				if err != nil {
					log.Println(subject.UID, "HTTPRequest.httpSubject.Error", err)
					subject.onError(err)
					return
				}
				subject.onNext(data)
				log.Println(subject.UID, "HTTPRequest.httpSubject.Complete via ReadAll")
				subject.Yield()
				subject.onComplete(subject)
				return
			}
			chunk, err := reader.ReadBytes(delimiter)
			if err != nil && err != io.EOF {
				log.Println(subject.UID, "HTTPRequest.httpSubject.Error", err)
				subject.onError(err)
				return
			}
			chunkLength := len(chunk)
			if chunkLength != 0 {
				dlog.Println(subject.UID, "HTTPRequest.httpSubject.Next")
				subject.onNext(chunk)
				if contentLength > 0 {
					contentLength -= int64(chunkLength)
					if contentLength <= 0 {
						dlog.Println(subject.UID, "HTTPRequest.httpSubject.Complete via ContentLength")
						subject.Yield()
						subject.onComplete(subject)
						return
					}
				}
			}
			// end of read
			if err == io.EOF {
				dlog.Println(subject.UID, "HTTPRequest.httpSubject.Complete via EOF")
				subject.Yield()
				subject.onComplete(subject)
				return
			}
		}
	}()

	wg.Wait()
	return subject, nil
}

// NewHTTPByteSubject HTTP response of Observable<[]byte>
func NewHTTPByteSubject(url string, contentType string, payload []byte) (*Observable, error) {
	log.Println("HTTPRequest.ByteSubject")
	subject := NewSubject()
	client := NewHTTPClient(10 * time.Second)

	subject.Resubscribe(func(observer *Observable) error {
		log.Println(observer.UID, "HTTPRequest.ByteSubject.Resubscribe")
		httpSubject, err := client.subject(url, contentType, payload, 0)
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

// NewHTTPTextSubject HTTP response of Observable<string>
func NewHTTPTextSubject(url string, payload []byte) (*Observable, error) {
	log.Println("HTTPRequest.TextSubject")
	subject := NewSubject()
	client := NewHTTPClient(10 * time.Second)

	subject.Resubscribe(func(observer *Observable) error {
		log.Println(observer.UID, "HTTPRequest.TextSubject.Resubscribe")
		httpSubject, err := client.subject(url, "text/plain", payload, 0)
		if err != nil {
			return err
		}
		httpSubject.Map(func(event interface{}) interface{} {
			if event == nil {
				return ""
			}
			return string(event.([]byte))
		})
		httpSubject.UID = "TextSubject." + observer.UID
		httpSubject.Pipe(observer)
		dlog.Println(httpSubject.UID, "HTTPRequest.TextSubject.Subscribed")
		return nil
	})

	return subject, nil
}

// NewHTTPLineSubject HTTP response of Observable<[]byte> delimited by newlines
func NewHTTPLineSubject(url string, payload []byte) (*Observable, error) {
	log.Println("HTTPRequest.LineSubject")
	subject := NewSubject()
	client := NewHTTPClient(10 * time.Second)

	subject.Resubscribe(func(observer *Observable) error {
		log.Println(observer.UID, "HTTPRequest.LineSubject.Resubscribe")
		httpSubject, err := client.subject(url, "text/plain", payload, byte('\n'))
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

// NewHTTPJSONSubject export
func NewHTTPJSONSubject(url string, payload []byte) (*Observable, error) {
	log.Println("HTTPRequest.JSONSubject")
	subject := NewSubject()
	client := NewHTTPClient(10 * time.Second)

	subject.Resubscribe(func(observer *Observable) error {
		log.Println(observer.UID, "HTTPRequest.JSONSubject.Resubscribe")
		httpSubject, err := client.subject(url, "application/json", payload, 0)
		if err != nil {
			return err
		}
		httpSubject.Map(func(event interface{}) interface{} {
			if event == nil {
				return nil
			}
			data := event.([]byte)
			var result interface{}
			err := json.Unmarshal(data, &result)
			if err != nil {
				subject.onError(err)
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

// NewHTTPSSESubject export
func NewHTTPSSESubject(url string, payload []byte) (*Observable, error) {
	log.Println("HTTPRequest.SSESubject")
	subject := NewSubject()
	client := NewHTTPClient(10 * time.Second)

	subject.Resubscribe(func(observer *Observable) error {
		log.Println(observer.UID, "HTTPRequest.SSESubject.Resubscribe")
		// assumption, events will never exceed 10 lines
		lineMax := 10
		lines := make([][]byte, 10)
		i := 0

		httpSubject, err := client.subject(url, "text/event-stream", payload, byte('\n'))
		if err != nil {
			return err
		}
		httpSubject.Map(func(event interface{}) interface{} {
			var line []byte
			if event == nil {
				line = []byte{}
			} else {
				line = event.([]byte)
			}
			if len(line) == 1 || i == lineMax {
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
			sse := make(map[string]interface{}, lineMax)
			lines := event.([][]byte)
			for i = 0; i < len(lines); i++ {
				line := string(lines[i])
				split := strings.Index(line, ":")
				sse[line[:split]] = strings.TrimSpace(line[split+1:])
			}
			return sse
		}).Filter(func(event interface{}) bool {
			return event != nil
		})
		httpSubject.UID = "SSESubject." + observer.UID
		httpSubject.Pipe(observer)
		dlog.Println(httpSubject.UID, "HTTPRequest.SSESubject.Subscribed")
		return nil
	})

	return subject, nil
}
