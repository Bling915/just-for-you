package bootstrap

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	identityapp "github.com/nanfang/just-for-you/server/internal/account/application"
	identitymysql "github.com/nanfang/just-for-you/server/internal/account/infrastructure/mysql"
	wechat "github.com/nanfang/just-for-you/server/internal/account/infrastructure/wechat"
	identityhttp "github.com/nanfang/just-for-you/server/internal/account/interfaces/http"
	cosstorage "github.com/nanfang/just-for-you/server/internal/infrastructure/storage/cos"
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
	repo := identitymysql.MySQLAccountStore{DB: db}
	avatars, err := cosstorage.NewCOSStorage(cfg.COSBucketURL, cfg.COSSecretID, cfg.COSSecretKey)
	if err != nil {
		return nil, tokens, avatars, err
	}
	provider := wechat.WeChatIdentityProvider{AppID: cfg.WeChatAppID, Secret: cfg.WeChatSecret, Client: &http.Client{Timeout: 5 * time.Second}}
	useCases := identityhttp.Dependencies{Login: identityapp.NewLogin(provider, repo, tokens, clock.System{}, auth.GenerateRefreshToken), RefreshSession: identityapp.NewRefreshSession(repo, tokens, clock.System{}, auth.GenerateRefreshToken), Logout: identityapp.NewLogout(repo, clock.System{}), GetProfile: identityapp.NewGetProfile(repo, avatars, clock.System{}), UpdateProfile: identityapp.NewUpdateProfile(repo, avatars, clock.System{}), RequestAvatarUpload: identityapp.NewRequestAvatarUpload(repo, avatars, clock.System{}, avatarKey)}
	return identityhttp.NewHandler(useCases), tokens, avatars, nil
}
