package bootstrap

import (
	"database/sql"
	identityapp "github.com/nanfang/just-for-you/server/internal/application/identity"
	"github.com/nanfang/just-for-you/server/internal/infrastructure/clock"
	"github.com/nanfang/just-for-you/server/internal/infrastructure/config"
	wechat "github.com/nanfang/just-for-you/server/internal/infrastructure/identity/wechat"
	identitymysql "github.com/nanfang/just-for-you/server/internal/infrastructure/persistence/mysql/identity"
	jwt "github.com/nanfang/just-for-you/server/internal/infrastructure/security/jwt"
	identityhttp "github.com/nanfang/just-for-you/server/internal/interfaces/http/identity"
	"net/http"
	"time"
)

func buildIdentity(cfg config.Config, db *sql.DB) (*identityhttp.Handler, jwt.JWTManager) {
	tokens := jwt.NewJWTManager([]byte(cfg.JWTSecret))
	repo := identitymysql.MySQLRepository{DB: db}
	provider := wechat.WeChatIdentityProvider{AppID: cfg.WeChatAppID, Secret: cfg.WeChatSecret, Client: &http.Client{Timeout: 5 * time.Second}}
	useCases := identityhttp.UseCases{Login: identityapp.NewLoginUseCase(provider, repo, tokens, clock.System{}, jwt.GenerateRefreshToken), RefreshSession: identityapp.NewRefreshSessionUseCase(repo, tokens, clock.System{}, jwt.GenerateRefreshToken), Logout: identityapp.NewLogoutUseCase(repo, clock.System{}), GetProfile: identityapp.NewGetProfileUseCase(repo), UpdateProfile: identityapp.NewUpdateProfileUseCase(repo)}
	return identityhttp.NewHandler(useCases), tokens
}
