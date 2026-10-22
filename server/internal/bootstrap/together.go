package bootstrap

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	app "github.com/nanfang/just-for-you/server/internal/application/together"
	"github.com/nanfang/just-for-you/server/internal/infrastructure/clock"
	"github.com/nanfang/just-for-you/server/internal/infrastructure/config"
	mysql "github.com/nanfang/just-for-you/server/internal/infrastructure/persistence/mysql/together"
	cosstorage "github.com/nanfang/just-for-you/server/internal/infrastructure/storage/cos"
	togetherhttp "github.com/nanfang/just-for-you/server/internal/interfaces/http/together"
)

func coverKey() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("cover")
	}
	return hex.EncodeToString(b)
}
func buildTogether(db *sql.DB, cfg config.Config) (*togetherhttp.Handler, error) {
	repo := mysql.NewRepository(db)
	storage, err := cosstorage.NewCOSStorage(cfg.COSBucketURL, cfg.COSSecretID, cfg.COSSecretKey)
	if err != nil {
		return nil, err
	}
	return togetherhttp.NewHandler(app.NewServiceWithStorage(repo, repo, clock.System{}, storage, coverKey)), nil
}
