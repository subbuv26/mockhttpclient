/*
   Copyright 2021, Subba Reddy Veeramreddy

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package mockhttpclient

import (
	"bytes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"io/ioutil"
	"net/http"
	"testing"
)

func TestMockHTTPClient(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Mock HTTP Client Suite")
}

var _ = Describe("Mock HTTP Client Tests", func() {
	Describe("Getting New Client", func() {
		var responseMap ResponseConfigMap
		BeforeEach(func() {
			responseMap = make(ResponseConfigMap)
			responseMap[http.MethodPost] = &ResponseConfig{
				Responses: []*http.Response{
					{
						StatusCode: 200,
						Header:     http.Header{},
						Body:       ioutil.NopCloser(bytes.NewReader([]byte("body"))),
					},
				},
			}

		})

		It("Create a client", func() {
			responseMap[http.MethodPost].MaxRun = 2
			client, err := NewMockHTTPClient(responseMap)
			Expect(err).To(BeNil(), "Failed to Create Client")
			Expect(client).NotTo(BeNil(), "Failed to Create Client")
		})

		It("Empty Responses", func() {
			responseMap[http.MethodPost].MaxRun = 2
			responseMap[http.MethodPost].Responses = nil
			client, err := NewMockHTTPClient(responseMap)
			Expect(err).NotTo(BeNil(), "Failed to Validate Response Configuration")
			Expect(client).To(BeNil(), "Failed to Validate Response Configuration")
		})

		It("Invalid maxRun (0)", func() {
			client, err := NewMockHTTPClient(responseMap)
			Expect(err).To(BeNil(), "Failed to Create Client")
			Expect(client).NotTo(BeNil(), "Failed to Create Client")
			Expect(responseMap[http.MethodPost].MaxRun).To(Equal(1), "Failed to set MaxRun")
		})

		It("Invalid maxRun (-ve)", func() {
			responseMap[http.MethodPost].MaxRun = -2
			client, err := NewMockHTTPClient(responseMap)
			Expect(err).NotTo(BeNil(), "Failed to Validate Response Configuration")
			Expect(client).To(BeNil(), "Failed to Validate Response Configuration")
		})

		It("Invalid maxRun (-ve)", func() {
			client, err := NewMockHTTPClient(nil)
			Expect(err).NotTo(BeNil(), "Failed to Validate Response Configuration")
			Expect(client).To(BeNil(), "Failed to Validate Response Configuration")
		})
	})

	Describe("Send Requests with the Client", func() {
		var responseMap ResponseConfigMap
		var client *http.Client
		var resp1, resp2, resp3 *http.Response
		var postReq, getReq *http.Request

		BeforeEach(func() {
			resp1 = &http.Response{
				StatusCode: 200,
				Header:     http.Header{},
				Body:       ioutil.NopCloser(bytes.NewReader([]byte("body"))),
			}

			resp2 = &http.Response{
				StatusCode: 404,
				Header:     http.Header{},
				Body:       ioutil.NopCloser(bytes.NewReader([]byte("body"))),
			}

			resp3 = &http.Response{
				StatusCode: 503,
				Header:     http.Header{},
				Body:       ioutil.NopCloser(bytes.NewReader([]byte("body"))),
			}

			postReq, _ = http.NewRequest("POST", "http://1.2.3.4", bytes.NewBuffer([]byte("{}")))
			getReq, _ = http.NewRequest("GET", "http://1.2.3.4", bytes.NewBuffer([]byte("")))

			responseMap = make(ResponseConfigMap)
		})

		It("Basic Request and Response", func() {
			responseMap[http.MethodGet] = &ResponseConfig{}
			responseMap[http.MethodGet].Responses = []*http.Response{resp1}

			client, _ = NewMockHTTPClient(responseMap)
			resp, err := client.Do(getReq)
			Expect(err).To(BeNil())
			Expect(resp).To(Equal(resp1))
		})

		It("Continuous Requests and Responses", func() {
			responseMap[http.MethodPost] = &ResponseConfig{}
			responseMap[http.MethodPost].Responses = []*http.Response{resp1, resp3}
			responseMap[http.MethodGet] = &ResponseConfig{}
			responseMap[http.MethodGet].Responses = []*http.Response{resp1, resp2, resp3}

			client, _ = NewMockHTTPClient(responseMap)

			rsp, err := client.Do(postReq)
			Expect(err).To(BeNil())
			Expect(rsp).To(Equal(resp1))

			rsp, err = client.Do(getReq)
			Expect(err).To(BeNil())
			Expect(rsp).To(Equal(resp1))

			rsp, err = client.Do(getReq)
			Expect(err).To(BeNil())
			Expect(rsp).To(Equal(resp2))

			rsp, err = client.Do(postReq)
			Expect(err).To(BeNil())
			Expect(rsp).To(Equal(resp3))

			rsp, err = client.Do(getReq)
			Expect(err).To(BeNil())
			Expect(rsp).To(Equal(resp3))
		})

		It("Run over multiple Cycles", func() {
			responseMap[http.MethodGet] = &ResponseConfig{}
			responseMap[http.MethodGet].Responses = []*http.Response{resp1, resp2, resp3}
			responseMap[http.MethodGet].MaxRun = 9
			client, _ = NewMockHTTPClient(responseMap)

			for i := 0; i < 3; i++ {
				for _, resp := range responseMap[http.MethodGet].Responses {
					rsp, err := client.Do(getReq)
					Expect(err).To(BeNil())
					Expect(rsp).To(Equal(resp))
				}
			}
		})

		It("Out Run maxRun", func() {
			responseMap[http.MethodPost] = &ResponseConfig{}
			responseMap[http.MethodPost].Responses = []*http.Response{resp1, resp2, resp3}
			responseMap[http.MethodPost].MaxRun = 3
			client, _ = NewMockHTTPClient(responseMap)
			for _, resp := range responseMap[http.MethodPost].Responses {
				rsp, err := client.Do(postReq)
				Expect(err).To(BeNil())
				Expect(rsp).To(Equal(resp))
			}
			rsp, err := client.Do(postReq)
			Expect(err).NotTo(BeNil())
			Expect(rsp).To(BeNil())
		})

		It("Nil Response", func() {
			responseMap[http.MethodPost] = &ResponseConfig{}
			responseMap[http.MethodPost].Responses = []*http.Response{nil}
			client, _ = NewMockHTTPClient(responseMap)
			resp, err := client.Do(postReq)
			Expect(err).NotTo(BeNil())
			Expect(resp).To(BeNil())
		})

		It("No Responses for http Method", func() {
			responseMap[http.MethodPost] = &ResponseConfig{}
			responseMap[http.MethodPost].Responses = []*http.Response{resp1}
			client, _ = NewMockHTTPClient(responseMap)
			resp, err := client.Do(getReq)
			Expect(err).NotTo(BeNil())
			Expect(resp).To(BeNil())
		})
	})
})
