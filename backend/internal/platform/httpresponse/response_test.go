package httpresponse

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestWriteDataKeepsResponseEnvelope(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	WriteData(context, http.StatusCreated, gin.H{"id": 7})
	var body map[string]map[string]int
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if recorder.Code != http.StatusCreated || body["data"]["id"] != 7 {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
}

func TestWriteErrorKeepsErrorEnvelope(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	WriteError(context, http.StatusBadRequest, "INVALID", "参数错误")
	var body map[string]ErrorBody
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if got := body["error"]; recorder.Code != http.StatusBadRequest || got.Code != "INVALID" || got.Message != "参数错误" {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
}
