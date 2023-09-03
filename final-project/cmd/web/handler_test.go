package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/MatsuoTakuro/final-project/data"
)

type optAssert func(params optParams)

type optParams struct {
	t   *testing.T
	ctx context.Context
	w   *httptest.ResponseRecorder
}

func Test_Handlers(t *testing.T) {
	HTMLTmplPath = "./templates"
	ManualTmplPath = "./../../pdf/manual.pdf"
	ManualOutputTempPath = "./../../tmp/%d_manual.pdf"

	tests := map[string]struct {
		path               string
		method             string
		queryParams        url.Values
		rawBody            url.Values
		expectedStatusCode int
		handler            http.HandlerFunc
		sessionData        map[string]any
		expectedHTML       []string
		optAsserts         []optAssert
	}{
		"home page": {
			path:               HOME_PATH,
			method:             http.MethodGet,
			rawBody:            nil,
			expectedStatusCode: http.StatusOK,
			handler:            testServer.HomePage,
			sessionData:        nil,
			expectedHTML:       []string{`<h1 class="mt-5">Home</h1>`},
			optAsserts:         nil,
		},
		"login page": {
			path:               LOGIN_PATH,
			method:             http.MethodGet,
			rawBody:            nil,
			expectedStatusCode: http.StatusOK,
			handler:            testServer.LoginPage,
			sessionData:        nil,
			expectedHTML:       []string{`<h1 class="mt-5">Login</h1>`},
			optAsserts:         nil,
		},
		"logout": {
			path:               LOGOUT_PATH,
			method:             http.MethodGet,
			rawBody:            nil,
			expectedStatusCode: http.StatusSeeOther,
			handler:            testServer.Logout,
			sessionData:        nil,
			expectedHTML:       nil,
			optAsserts:         nil,
		},
		"login": {
			path:   LOGIN_PATH,
			method: http.MethodPost,
			rawBody: url.Values{
				EMAIL_ATTR:    {"admin@example.com"},
				PASSWORD_ATTR: {"abc123abc123abc123abc123"},
			},
			expectedStatusCode: http.StatusSeeOther,
			handler:            testServer.Login,
			sessionData:        nil,
			expectedHTML:       nil,
			optAsserts: []optAssert{
				func(params optParams) {
					if !testServer.Session.Exists(params.ctx, USER_ID_CTX) {
						params.t.Errorf("expected session to contain %s", USER_ID_CTX)
					}
				},
			},
		},
		"subscribe to plan": {
			path:    SUBSCRIBE_PATH,
			method:  http.MethodGet,
			rawBody: nil,
			queryParams: url.Values{
				PLAN_ID_CTX: {"1"},
			},
			expectedStatusCode: http.StatusSeeOther,
			handler:            testServer.SubcribeToPlan,
			sessionData: map[string]any{
				USER_CTX: data.User{
					ID:        1,
					Email:     "admin@example.com",
					FirstName: "Admin",
					LastName:  "User",
					IsActive:  data.Active,
				},
			},
			expectedHTML: nil,
			optAsserts:   nil,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()

			if tt.queryParams != nil {
				tt.path += "?" + tt.queryParams.Encode()
			}
			var reqBody = &strings.Reader{}
			if tt.rawBody != nil {
				reqBody = strings.NewReader(tt.rawBody.Encode())
			}
			rawReq, _ := http.NewRequest(tt.method, tt.path, reqBody)
			r := newReqWithSession(rawReq)

			if len(tt.sessionData) > 0 {
				for k, v := range tt.sessionData {
					testServer.Session.Put(r.Context(), k, v)
				}
			}

			tt.handler.ServeHTTP(w, r)

			// wait for the async job to finish if it is fired off
			testServer.AsyncJob.Wait()

			if w.Code != tt.expectedStatusCode {
				t.Errorf("expected status code %d; got %d", tt.expectedStatusCode, w.Code)
			}

			if len(tt.expectedHTML) > 0 {
				got := w.Body.String()
				for _, v := range tt.expectedHTML {
					if !strings.Contains(got, v) {
						t.Errorf("expected %s to contain %s", got, v)
					}
				}
			}

			if len(tt.optAsserts) > 0 {
				for _, a := range tt.optAsserts {
					a(optParams{t: t, ctx: r.Context(), w: w})
				}
			}
		})
	}
}
