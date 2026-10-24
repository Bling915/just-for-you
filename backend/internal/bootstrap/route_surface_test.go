package bootstrap

import (
	"context"
	"sort"
	"testing"

	accounthttp "github.com/nanfang/just-for-you/server/internal/account/interfaces/http"
	anniversaryhttp "github.com/nanfang/just-for-you/server/internal/anniversary/interfaces/http"
	companionshiphttp "github.com/nanfang/just-for-you/server/internal/companionship/interfaces/http"
	coupleapp "github.com/nanfang/just-for-you/server/internal/couple/application"
	coupledomain "github.com/nanfang/just-for-you/server/internal/couple/domain"
	couplehttp "github.com/nanfang/just-for-you/server/internal/couple/interfaces/http"
	dailyquestionhttp "github.com/nanfang/just-for-you/server/internal/dailyquestion/interfaces/http"
	momenthttp "github.com/nanfang/just-for-you/server/internal/moment/interfaces/http"
	todohttp "github.com/nanfang/just-for-you/server/internal/todo/interfaces/http"
	wishhttp "github.com/nanfang/just-for-you/server/internal/wish/interfaces/http"
)

type soloTestExecutor struct{}

func (soloTestExecutor) Execute(
	context.Context,
	coupleapp.EnableSoloTestCommand,
) (coupledomain.Couple, error) {
	return coupledomain.Couple{}, nil
}

func TestEveryContextRegistersItsRouteSurfaceOnce(t *testing.T) {
	router := NewRouter(Dependencies{
		Tokens: testTokens{},
		Registrars: []RouteRegistrar{
			accounthttp.NewHandler(accounthttp.Dependencies{}),
			couplehttp.NewHandler(couplehttp.Dependencies{EnableSoloTest: soloTestExecutor{}}),
			companionshiphttp.NewHandler(companionshiphttp.Dependencies{}),
			momenthttp.NewHandler(momenthttp.Dependencies{}),
			anniversaryhttp.NewHandler(anniversaryhttp.Dependencies{}),
			todohttp.NewHandler(todohttp.Dependencies{}),
			wishhttp.NewHandler(wishhttp.Dependencies{}),
			dailyquestionhttp.NewHandler(dailyquestionhttp.Dependencies{}),
		},
	})
	want := []string{
		"GET /healthz",
		"POST /v1/auth/logout",
		"POST /v1/auth/refresh",
		"POST /v1/auth/wechat",
		"GET /v1/users/me",
		"POST /v1/users/me/avatar-upload",
		"PUT /v1/users/me",
		"POST /v1/couples/bind",
		"POST /v1/couples/invitations",
		"GET /v1/couples/invitations/:code",
		"GET /v1/couples/me",
		"DELETE /v1/couples/me",
		"POST /v1/couples/solo-test",
		"GET /v1/presence",
		"PUT /v1/presence",
		"POST /v1/nudges",
		"GET /v1/notification-preferences",
		"PUT /v1/notification-preferences",
		"POST /v1/uploads",
		"GET /v1/moments",
		"POST /v1/moments",
		"PATCH /v1/moments/:id",
		"DELETE /v1/moments/:id",
		"POST /v1/moments/:id/share-to-couple",
		"POST /v1/moments/:id/reactions",
		"DELETE /v1/moments/:id/reactions/:reactionId",
		"PUT /v1/moments/:id/perspective",
		"DELETE /v1/moments/:id/perspective",
		"GET /v1/home",
		"GET /v1/moments/calendar",
		"GET /v1/moments/map",
		"GET /v1/reports/monthly",
		"GET /v1/memories/resurface",
		"GET /v1/anniversaries",
		"POST /v1/anniversaries",
		"GET /v1/anniversaries/:id",
		"PATCH /v1/anniversaries/:id",
		"POST /v1/anniversaries/:id/cover-upload",
		"DELETE /v1/anniversaries/:id",
		"GET /v1/todos",
		"POST /v1/todos",
		"PATCH /v1/todos/:id",
		"DELETE /v1/todos/:id",
		"GET /v1/wishes",
		"POST /v1/wishes",
		"PATCH /v1/wishes/:id",
		"DELETE /v1/wishes/:id",
		"GET /v1/daily-question",
		"PUT /v1/daily-question/answer",
		"GET /v1/daily-questions",
	}
	got := make([]string, 0, len(router.Routes()))
	for _, route := range router.Routes() {
		got = append(got, route.Method+" "+route.Path)
	}
	sort.Strings(got)
	sort.Strings(want)
	if len(got) != len(want) {
		t.Fatalf("route count=%d want=%d\ngot=%v", len(got), len(want), got)
	}
	for index := range want {
		if got[index] != want[index] {
			t.Fatalf("route[%d]=%q want=%q\ngot=%v", index, got[index], want[index], got)
		}
	}
}
