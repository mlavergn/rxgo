package rx

import (
	"bufio"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"time"
)

// RxHTTP type
type RxHTTP struct {
	client *http.Client
}

// NewRxHTTP init
func NewRxHTTP(timeout time.Duration) *RxHTTP {
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
			Timeout: timeout * time.Second,
		}).DialContext,
	}
	return &RxHTTP{
		client: &http.Client{
			Transport: httpTransport,
		},
	}
}

// Send export
func (id *RxHTTP) send(url string, mime string, delimiter byte, parser func(*RxObservable, interface{})) (*RxObservable, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
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

	subject := NewRxObservable()
	reader := bufio.NewReader(resp.Body)

	go func() {
		defer func() {
			subject.Complete <- true
		}()

		for {
			select {
			case <-subject.Complete:
				return
			default:
				if delimiter == 0 {
					data, err := ioutil.ReadAll(reader)
					if err != nil {
						log.Println(err)
						subject.Error <- err
						return
					}
					parser(subject, data)
				} else {
					chunk, err := reader.ReadBytes(delimiter)
					if err != nil {
						log.Println(err)
						subject.Error <- err
						return
					}
					parser(subject, chunk)
				}
			}
		}
	}()

	return subject, nil
}

// Text export
func (id *RxHTTP) Text(url string) (*RxObservable, error) {
	return id.send(url, "text/plain", 0, func(subject *RxObservable, raw interface{}) {
		text := ToByteString(raw, "")
		subject.Next <- text
		subject.Complete <- true
	})
}

// JSON export
func (id *RxHTTP) JSON(url string) (*RxObservable, error) {
	return id.send(url, "application/json", 0, func(subject *RxObservable, raw interface{}) {
		data := ToByteArray(raw, nil)
		var result interface{}
		err := json.Unmarshal(data, &result)
		if err != nil {
			subject.Error <- err
		}
		subject.Next <- result
		subject.Complete <- true
	})
}

// SSE export
func (id *RxHTTP) SSE(url string) (*RxObservable, error) {
	// assumption, events will never exceed 10 lines
	lines := [10][]byte{}
	i := 0

	return id.send(url, "text/event-stream", byte('\n'), func(subject *RxObservable, raw interface{}) {
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
