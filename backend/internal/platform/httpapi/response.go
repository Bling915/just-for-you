package httpapi

import (
	"github.com/nanfang/just-for-you/server/internal/platform/auth"
	"github.com/nanfang/just-for-you/server/internal/platform/httpresponse"
)

var WriteData = httpresponse.WriteData
var WriteError = httpresponse.WriteError
var UserID = auth.UserID
