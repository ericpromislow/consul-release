package consul_test

import (
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"

	"github.com/cloudfoundry-incubator/consul-release/src/acceptance-tests/testing/consul"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("HTTPKV", func() {
	Describe("Get", func() {
		AfterEach(func() {
			consul.ResetBodyReader()
		})

		It("gets the key-value based on the key", func() {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				Expect(req.URL.Path).To(Equal("/v1/kv/some-key"))

				Expect(req.Method).To(Equal("GET"))

				params, err := url.ParseQuery(req.URL.RawQuery)
				Expect(err).NotTo(HaveOccurred())

				_, ok := params["raw"]
				Expect(ok).To(BeTrue())

				w.Write([]byte("some-value"))
			}))

			kv := consul.NewHTTPKV(server.URL)

			value, err := kv.Get("some-key")
			Expect(err).NotTo(HaveOccurred())

			Expect(value).To(Equal("some-value"))
		})

		Context("failure cases", func() {
			Context("when consul cant find a value based on the key", func() {
				It("should return an error", func() {
					server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
						w.WriteHeader(http.StatusNotFound)
					}))

					kv := consul.NewHTTPKV(server.URL)

					_, err := kv.Get("some-key")
					Expect(err).To(MatchError(errors.New("key \"some-key\" not found")))
				})
			})

			Context("when consul returns some other non-200 status code", func() {
				It("should return an error", func() {
					server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
						w.WriteHeader(http.StatusBadGateway)
					}))

					kv := consul.NewHTTPKV(server.URL)

					_, err := kv.Get("some-key")
					Expect(err).To(MatchError(errors.New("consul http error: 502 Bad Gateway")))
				})
			})

			Context("when the consul address is invalid", func() {
				It("returns an error", func() {
					kv := consul.NewHTTPKV("banana://some-bad-address")

					_, err := kv.Get("some-key")
					Expect(err).To(MatchError(ContainSubstring("unsupported protocol")))
				})
			})

			Context("when consul returns garbage", func() {
				It("returns an error", func() {
					server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
						w.Write([]byte("true"))
					}))

					consul.SetBodyReader(func(io.Reader) ([]byte, error) {
						return []byte{}, errors.New("bad things happened")
					})

					kv := consul.NewHTTPKV(server.URL)

					_, err := kv.Get("some-key")
					Expect(err).To(MatchError(errors.New("bad things happened")))
				})
			})
		})
	})

	Describe("Set", func() {
		AfterEach(func() {
			consul.ResetBodyReader()
		})

		It("sets a key-value pair over HTTP", func() {
			var wasCalled bool
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				defer req.Body.Close()
				wasCalled = true

				Expect(req.URL.Path).To(Equal("/v1/kv/some-key"))

				body, err := ioutil.ReadAll(req.Body)
				Expect(err).NotTo(HaveOccurred())

				Expect(string(body)).To(Equal("some-value"))
				Expect(req.Method).To(Equal("PUT"))

				w.Write([]byte("true"))
			}))

			kv := consul.NewHTTPKV(server.URL)

			err := kv.Set("some-key", "some-value")
			Expect(err).NotTo(HaveOccurred())
			Expect(wasCalled).To(BeTrue())
		})

		Context("failure cases", func() {
			Context("when consul fails to save the value", func() {
				It("returns an error", func() {
					server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
						w.Write([]byte("false"))
					}))

					kv := consul.NewHTTPKV(server.URL)

					err := kv.Set("some-key", "some-value")
					Expect(err).To(MatchError(errors.New("failed to save to kv store")))
				})
			})

			Context("when consul returns garbage", func() {
				It("returns an error", func() {
					server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
						w.Write([]byte("true"))
					}))

					consul.SetBodyReader(func(io.Reader) ([]byte, error) {
						return []byte{}, errors.New("bad things happened")
					})

					kv := consul.NewHTTPKV(server.URL)

					err := kv.Set("some-key", "some-value")
					Expect(err).To(MatchError(errors.New("bad things happened")))
				})
			})

			Context("when the consul address is invalid", func() {
				It("returns an error", func() {
					kv := consul.NewHTTPKV("banana://some-bad-address")

					err := kv.Set("some-key", "some-value")
					Expect(err).To(MatchError(ContainSubstring("unsupported protocol")))
				})

				It("returns an error", func() {
					kv := consul.NewHTTPKV("banana://%%%%%")

					err := kv.Set("some-key", "some-value")
					Expect(err).To(MatchError(ContainSubstring("percent-encoded characters")))
				})
			})
		})
	})
})
