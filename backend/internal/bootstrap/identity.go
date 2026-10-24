package bootstrap

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	identityapp "github.com/nanfang/just-for-you/server/internal/application/identity"
	wechat "github.com/nanfang/just-for-you/server/internal/infrastructure/identity/wechat"
	identitymysql "github.com/nanfang/just-for-you/server/internal/infrastructure/persistence/mysql/identity"
	cosstorage "github.com/nanfang/just-for-you/server/internal/infrastructure/storage/cos"
	identityhttp "github.com/nanfang/just-for-you/server/internal/interfaces/http/identity"
	"github.com/nanfang/just-for-you/server/internal/platform/auth"
	"github.com/nanfang/just-for-you/server/internal/platform/clock"
	"github.com/nanfang/just-for-you/server/internal/platform/config"
	"net/http"
	"time"
)

func avatarKey() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b)
}
func buildIdentity(cfg config.Config, db *sql.DB) (*identityhttp.Handler, auth.TokenService, cosstorage.COSStorage, error) {
	tokens := auth.NewTokenService([]byte(cfg.JWTSecret))
	repo := identitymysql.MySQLRepository{DB: db}
	avatars, err := cosstorage.NewCOSStorage(cfg.COSBucketURL, cfg.COSSecretID, cfg.COSSecretKey)
	if err != nil {
		return nil, tokens, avatars, err
	}
	provider := wechat.WeChatIdentityProvider{AppID: cfg.WeChatAppID, Secret: cfg.WeChatSecret, Client: &http.Client{Timeout: 5 * time.Second}}
	useCases := identityhttp.UseCases{Login: identityapp.NewLoginUseCase(provider, repo, tokens, clock.System{}, auth.GenerateRefreshToken), RefreshSession: identityapp.NewRefreshSessionUseCase(repo, tokens, clock.System{}, auth.GenerateRefreshToken), Logout: identityapp.NewLogoutUseCase(repo, clock.System{}), GetProfile: identityapp.NewGetProfileUseCase(repo, avatars, clock.System{}), UpdateProfile: identityapp.NewUpdateProfileUseCase(repo, avatars, clock.System{}), RequestAvatarUpload: identityapp.NewRequestAvatarUploadUseCase(repo, avatars, clock.System{}, avatarKey)}
	return identityhttp.NewHandler(useCases), tokens, avatars, nil
}
