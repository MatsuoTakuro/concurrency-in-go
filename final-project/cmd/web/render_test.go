package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestServer_UpdateDefaultData(t *testing.T) {
	rawReq, _ := http.NewRequest("GET", "/", nil)
	req := newReqWithSession(rawReq)

	wantFlashMsg := "flash"
	wantWarnMsg := "warning"
	wantErrMsg := "error"
	testServer.Session.Put(req.Context(), FLASH_CTX, wantFlashMsg)
	testServer.Session.Put(req.Context(), WARNING_CTX, wantWarnMsg)
	testServer.Session.Put(req.Context(), ERROR_CTX, wantErrMsg)

	tmplData := testServer.UpdateDefaultData(&TemplateData{}, req)

	if tmplData.Flash != wantFlashMsg {
		t.Error("failed to get flash data")
	}

	if tmplData.Warning != wantWarnMsg {
		t.Error("failed to get warning data")
	}

	if tmplData.Error != wantErrMsg {
		t.Error("failed to get error data")
	}
}

func TestServer_IsAuthenticated(t *testing.T) {
	rawReq, _ := http.NewRequest("GET", "/", nil)
	req := newReqWithSession(rawReq)

	isAuth := testServer.IsAuthenticated(req)
	if isAuth {
		t.Errorf("expected false, got %t", isAuth)
	}

	testServer.Session.Put(req.Context(), USER_ID_CTX, 1)
	isAuth = testServer.IsAuthenticated(req)
	if !isAuth {
		t.Errorf("expected true, got %t", isAuth)
	}
}

func TestServer_render(t *testing.T) {
	HTMLTmplPath = "./templates"

	w := httptest.NewRecorder()
	rawReq, _ := http.NewRequest("GET", "/", nil)
	r := newReqWithSession(rawReq)

	testServer.render(w, r, HOME_PAGE, &TemplateData{})

	if w.Code != http.StatusOK {
		t.Errorf("expected %d, got %d", http.StatusOK, w.Code)
	}
}
