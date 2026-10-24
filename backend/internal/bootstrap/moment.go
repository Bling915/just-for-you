package bootstrap

import (
	"database/sql"

	anniversarymysql "github.com/nanfang/just-for-you/server/internal/anniversary/infrastructure/mysql"
	coupleapp "github.com/nanfang/just-for-you/server/internal/couple/application"
	cos "github.com/nanfang/just-for-you/server/internal/media/infrastructure/cos"
	momentapp "github.com/nanfang/just-for-you/server/internal/moment/application"
	momentmysql "github.com/nanfang/just-for-you/server/internal/moment/infrastructure/mysql"
	momenthttp "github.com/nanfang/just-for-you/server/internal/moment/interfaces/http"
	"github.com/nanfang/just-for-you/server/internal/platform/clock"
	"github.com/nanfang/just-for-you/server/internal/platform/config"
)

func buildMoment(cfg config.Config, db *sql.DB, get *coupleapp.GetCurrentCouple) (*momenthttp.Handler, error) {
	repo := momentmysql.NewMySQLMomentStore(db)
	photos, err := cos.NewObjectStorage(cfg.COSBucketURL, cfg.COSSecretID, cfg.COSSecretKey)
	if err != nil {
		return nil, err
	}
	relationships := newMomentCoupleReader(get)
	dependencies := momenthttp.Dependencies{RequestPhotoUpload: momentapp.NewRequestPhotoUpload(repo, photos, clock.System{}), PublishMoment: momentapp.NewPublishMoment(relationships, repo, repo, photos, clock.System{}), UpdateMoment: momentapp.NewUpdateMoment(relationships, repo, repo, photos, clock.System{}), GetTimeline: momentapp.NewGetMomentTimeline(relationships, repo, photos, clock.System{}), GetHome: momentapp.NewGetHome(relationships, repo, photos, clock.System{}, newMomentAnniversaryReader(anniversarymysql.NewPinnedReader(db))), DeleteMoment: momentapp.NewDeleteMoment(relationships, repo), ShareToCouple: momentapp.NewShareToCouple(relationships, repo, clock.System{}), ReactToMoment: momentapp.NewReactToMoment(relationships, repo, clock.System{}), SavePerspective: momentapp.NewSavePerspective(relationships, repo, repo, photos, clock.System{}), GetResurfacedMemory: momentapp.NewGetResurfacedMemory(relationships, repo, photos, clock.System{}), ReadModels: momentapp.NewReadModels(relationships, repo)}
	return momenthttp.NewHandler(dependencies), nil
}
