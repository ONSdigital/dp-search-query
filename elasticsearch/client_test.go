package elasticsearch

import (
	"context"
	"errors"
	esauth "github.com/ONSdigital/dp-elasticsearch/v2/awsauth"
	dphttp "github.com/ONSdigital/dp-net/http"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestSearch(t *testing.T) {

	Convey("When Search is called", t, func() {
		// Define a mock struct to be used in your unit tests of myFunc.

		Convey("Then a request with the search action should be posted", func() {
			dphttpMock := &dphttp.ClienterMock{
				DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
					return newResponse("moo"), nil
				},
			}

			var testSigner *esauth.Signer

			client := New("http://localhost:999", dphttpMock, false, testSigner, "es", "eu-west-1")

			res, err := client.Search(context.Background(), "index", "doctype", []byte("search request"))
			So(err, ShouldBeNil)
			So(res, ShouldNotBeEmpty)
			So(dphttpMock.DoCalls(), ShouldHaveLength, 1)
			actualRequest := dphttpMock.DoCalls()[0].Req
			So(actualRequest.URL.String(), ShouldResemble, "http://localhost:999/index/doctype/_search")
			So(actualRequest.Method, ShouldResemble, "POST")
			body, err := ioutil.ReadAll(actualRequest.Body)
			So(err, ShouldBeNil)
			So(string(body), ShouldResemble, "search request")

		})

		Convey("Then a returned error should be passed back", func() {
			dphttpMock := &dphttp.ClienterMock{
				DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
					return nil, errors.New("http error")
				},
			}

			var testSigner *esauth.Signer

			client := New("http://localhost:999", dphttpMock, false, testSigner, "es", "eu-west-1")

			_, err := client.Search(context.Background(), "index", "doctype", []byte("search request"))
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldResemble, "http error")
			So(dphttpMock.DoCalls(), ShouldHaveLength, 1)
			actualRequest := dphttpMock.DoCalls()[0].Req
			So(actualRequest.URL.String(), ShouldResemble, "http://localhost:999/index/doctype/_search")
			So(actualRequest.Method, ShouldResemble, "POST")
			body, err := ioutil.ReadAll(actualRequest.Body)
			So(err, ShouldBeNil)
			So(string(body), ShouldResemble, "search request")

		})

	})
}

func TestMultiSearch(t *testing.T) {

	Convey("When MultiSearch is called", t, func() {

		Convey("Then a request with the multi search action should be posted", func() {

			dphttpMock := &dphttp.ClienterMock{
				DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
					return newResponse("moo"), nil
				},
			}

			var testSigner *esauth.Signer

			client := New("http://localhost:999", dphttpMock, false, testSigner, "es", "eu-west-1")

			res, err := client.MultiSearch(context.Background(), "index", "doctype", []byte("multiSearch request"))
			So(err, ShouldBeNil)
			So(res, ShouldNotBeEmpty)
			So(dphttpMock.DoCalls(), ShouldHaveLength, 1)
			actualRequest := dphttpMock.DoCalls()[0].Req
			So(actualRequest.URL.String(), ShouldResemble, "http://localhost:999/index/doctype/_msearch")
			So(actualRequest.Method, ShouldResemble, "POST")
			body, err := ioutil.ReadAll(actualRequest.Body)
			So(err, ShouldBeNil)
			So(string(body), ShouldResemble, "multiSearch request")
		})

		Convey("Then a returned error should be passed back", func() {
			dphttpMock := &dphttp.ClienterMock{
				DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
					return nil, errors.New("http error")
				},
			}
			var testSigner *esauth.Signer

			client := New("http://localhost:999", dphttpMock, false, testSigner, "es", "eu-west-1")

			_, err := client.MultiSearch(context.Background(), "index", "doctype", []byte("search request"))
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldResemble, "http error")
			So(dphttpMock.DoCalls(), ShouldHaveLength, 1)
			actualRequest := dphttpMock.DoCalls()[0].Req
			So(actualRequest.URL.String(), ShouldResemble, "http://localhost:999/index/doctype/_msearch")
			So(actualRequest.Method, ShouldResemble, "POST")
			body, err := ioutil.ReadAll(actualRequest.Body)
			So(err, ShouldBeNil)
			So(string(body), ShouldResemble, "search request")

		})
	})
}

func TestGetStatus(t *testing.T) {

	Convey("When GetStatus is called", t, func() {

		Convey("Then a GET request with the status action should be called", func() {

			dphttpMock := &dphttp.ClienterMock{
				DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
					return newResponse("moo"), nil
				},
			}

			var testSigner *esauth.Signer

			client := New("http://localhost:999", dphttpMock, false, testSigner, "es", "eu-west-1")

			res, err := client.GetStatus(context.Background())
			So(err, ShouldBeNil)
			So(res, ShouldNotBeEmpty)
			So(dphttpMock.DoCalls(), ShouldHaveLength, 1)
			actualRequest := dphttpMock.DoCalls()[0].Req
			So(actualRequest.URL.String(), ShouldResemble, "http://localhost:999/_cat/health")
			So(actualRequest.Method, ShouldResemble, "GET")
		})

		Convey("Then a returned error should be passed back", func() {
			dphttpMock := &dphttp.ClienterMock{
				DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
					return nil, errors.New("http error")
				},
			}

			var testSigner *esauth.Signer

			client := New("http://localhost:999", dphttpMock, false, testSigner, "es", "eu-west-1")

			_, err := client.GetStatus(context.Background())
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldResemble, "http error")
			So(dphttpMock.DoCalls(), ShouldHaveLength, 1)
			actualRequest := dphttpMock.DoCalls()[0].Req
			So(actualRequest.URL.String(), ShouldResemble, "http://localhost:999/_cat/health")
			So(actualRequest.Method, ShouldResemble, "GET")

		})

		Convey("Then a signing error should be passed back", func() {
			dphttpMock := &dphttp.ClienterMock{
				DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
					return newResponse("moo"), nil
				},
			}

			var testSigner *esauth.Signer

			client := New("http://localhost:999", dphttpMock, true, testSigner, "es", "eu-west-1")

			_, err := client.GetStatus(context.Background())
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldResemble, "v4 signer missing. Cannot sign request")

		})
	})
}

func newResponse(body string) *http.Response {
	recorder := httptest.NewRecorder()
	recorder.WriteString(body)
	return recorder.Result()
}
