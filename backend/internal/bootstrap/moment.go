package bootstrap

import (
	"database/sql"
	momentapp "github.com/nanfang/just-for-you/server/internal/application/moment"
	coupleapp "github.com/nanfang/just-for-you/server/internal/couple/application"
	momentmysql "github.com/nanfang/just-for-you/server/internal/infrastructure/persistence/mysql/moment"
	togethermysql "github.com/nanfang/just-for-you/server/internal/infrastructure/persistence/mysql/together"
	momenthttp "github.com/nanfang/just-for-you/server/internal/interfaces/http/moment"
	cos "github.com/nanfang/just-for-you/server/internal/media/infrastructure/cos"
	"github.com/nanfang/just-for-you/server/internal/platform/clock"
	"github.com/nanfang/just-for-you/server/internal/platform/config"
)

func buildMoment(cfg config.Config, db *sql.DB, get *coupleapp.GetCurrentCouple) (*momenthttp.Handler, error) {
	repo := momentmysql.NewMySQLRepository(db)
	photos, err := cos.NewObjectStorage(cfg.COSBucketURL, cfg.COSSecretID, cfg.COSSecretKey)
	if err != nil {
		return nil, err
	}
	relationships := momentmysql.NewActiveCoupleReader(get)
	useCases := momenthttp.UseCases{RequestPhotoUpload: momentapp.NewRequestPhotoUploadUseCase(repo, photos, clock.System{}), PublishMoment: momentapp.NewPublishMomentUseCase(relationships, repo, repo, photos, clock.System{}), UpdateMoment: momentapp.NewUpdateMomentUseCase(relationships, repo, repo, photos, clock.System{}), GetTimeline: momentapp.NewGetTimelineUseCase(relationships, repo, photos, clock.System{}), GetHome: momentapp.NewGetHomeUseCase(relationships, repo, photos, clock.System{}, togethermysql.NewHighlightReader(db)), DeleteMoment: momentapp.NewDeleteMomentUseCase(relationships, repo), ShareToCouple: momentapp.NewShareToCoupleUseCase(relationships, repo, clock.System{}), ReactToMoment: momentapp.NewReactToMomentUseCase(relationships, repo, clock.System{}), SavePerspective: momentapp.NewSavePerspectiveUseCase(relationships, repo, repo, photos, clock.System{}), GetResurfacedMemory: momentapp.NewGetResurfacedMemoryUseCase(relationships, repo, photos, clock.System{}), ReadModels: momentapp.NewReadModels(relationships, repo)}
	return momenthttp.NewHandler(useCases), nil
}
