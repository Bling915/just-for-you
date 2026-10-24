package couplehttp

import (
	"bytes"
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	application "github.com/nanfang/just-for-you/server/internal/couple/application"
	domain "github.com/nanfang/just-for-you/server/internal/couple/domain"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type fakeCreate struct{}

func (fakeCreate) Execute(context.Context, application.CreateInvitationCommand) (domain.Invitation, error) {
	return domain.Invitation{}, nil
}

type fakePreview struct{}

func (fakePreview) Execute(context.Context, application.PreviewInvitationQuery) (domain.Invitation, error) {
	return domain.Invitation{}, nil
}

type fakeBind struct{ err error }

func (f fakeBind) Execute(context.Context, application.BindCoupleCommand) (domain.Couple, error) {
	return domain.Couple{}, f.err
}

type fakeGet struct{}

func (fakeGet) Execute(context.Context, application.GetCoupleQuery) (application.CurrentCouple, error) {
	return application.CurrentCouple{}, domain.ErrNoActiveCouple
}

type fakeGetProfile struct{}

func (fakeGetProfile) Execute(context.Context, application.GetCoupleQuery) (application.CurrentCouple, error) {
	return application.CurrentCouple{
		Couple:  domain.NewCouple(9, 7, 8, time.Date(2026, 10, 20, 12, 0, 0, 0, time.UTC)),
		Partner: application.PartnerProfile{UserID: 8, DisplayName: "小南", AvatarURL: "cos://avatar"},
	}, nil
}

type fakeEnd struct{}

func (fakeEnd) Execute(context.Context, application.EndCoupleCommand) error { return nil }

type fakeSoloTest struct{}

func (fakeSoloTest) Execute(_ context.Context, command application.EnableSoloTestCommand) (domain.Couple, error) {
	return domain.NewCouple(12, command.UserID, 1007, time.Date(2026, 10, 20, 12, 0, 0, 0, time.UTC)), nil
}
func fakeUseCases(bindErr error) Dependencies {
	return Dependencies{CreateInvitation: fakeCreate{}, PreviewInvitation: fakePreview{}, BindCouple: fakeBind{bindErr}, GetCouple: fakeGet{}, EndCouple: fakeEnd{}}
}
func TestBindMapsStableErrorCodes(t *testing.T) {
	cases := []struct {
		err  error
		code string
	}{{domain.ErrInvitationExpired, "INVITE_EXPIRED"}, {domain.ErrInvitationUsed, "INVITE_USED"}, {domain.ErrCannotBindSelf, "CANNOT_BIND_SELF"}, {domain.ErrAlreadyBound, "ALREADY_BOUND"}}
	for _, item := range cases {
		t.Run(item.code, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			router := gin.New()
			group := router.Group("/v1", func(c *gin.Context) { c.Set("userID", int64(2)) })
			NewHandler(fakeUseCases(item.err)).RegisterProtected(group)
			request := httptest.NewRequest(http.MethodPost, "/v1/couples/bind", bytes.NewBufferString(`{"code":"AB3K9Z"}`))
			request.Header.Set("Content-Type", "application/json")
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)
			if !strings.Contains(response.Body.String(), `"code":"`+item.code+`"`) {
				t.Fatalf("body=%s", response.Body.String())
			}
		})
	}
}
func TestGetMapsNoActiveCouple(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	group := router.Group("/v1", func(c *gin.Context) { c.Set("userID", int64(2)) })
	NewHandler(fakeUseCases(nil)).RegisterProtected(group)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/v1/couples/me", nil))
	if response.Code != http.StatusNotFound || !strings.Contains(response.Body.String(), `"code":"NO_ACTIVE_COUPLE"`) {
		t.Fatalf("body=%s", response.Body.String())
	}
}

func TestGetReturnsPartnerProfileWithoutOpenID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	group := router.Group("/v1", func(c *gin.Context) { c.Set("userID", int64(7)) })
	useCases := fakeUseCases(nil)
	useCases.GetCouple = fakeGetProfile{}
	NewHandler(useCases).RegisterProtected(group)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/v1/couples/me", nil))
	body := response.Body.String()
	if response.Code != http.StatusOK || !strings.Contains(body, `"partner"`) || !strings.Contains(body, `"displayName":"小南"`) || !strings.Contains(body, `"soloTest":false`) {
		t.Fatalf("status=%d body=%s", response.Code, body)
	}
	if strings.Contains(strings.ToLower(body), "openid") {
		t.Fatalf("private identity leaked: %s", body)
	}
}

func TestSoloTestRouteIsRegisteredOnlyWhenEnabled(t *testing.T) {
	gin.SetMode(gin.TestMode)
	request := func(useCases Dependencies) *httptest.ResponseRecorder {
		router := gin.New()
		group := router.Group("/v1", func(c *gin.Context) { c.Set("userID", int64(7)) })
		NewHandler(useCases).RegisterProtected(group)
		response := httptest.NewRecorder()
		router.ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/v1/couples/solo-test", nil))
		return response
	}

	disabled := request(fakeUseCases(nil))
	if disabled.Code != http.StatusNotFound {
		t.Fatalf("disabled status=%d", disabled.Code)
	}
	enabledUseCases := fakeUseCases(nil)
	enabledUseCases.EnableSoloTest = fakeSoloTest{}
	enabled := request(enabledUseCases)
	if enabled.Code != http.StatusCreated || !strings.Contains(enabled.Body.String(), `"userAId":7`) || !strings.Contains(enabled.Body.String(), `"userBId":1007`) {
		t.Fatalf("enabled status=%d body=%s", enabled.Code, enabled.Body.String())
	}
}

var _ = errors.Is
