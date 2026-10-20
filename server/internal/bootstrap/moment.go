package bootstrap

import (
	"database/sql"
	momentapp "github.com/nanfang/just-for-you/server/internal/application/moment"
	relationshipapp "github.com/nanfang/just-for-you/server/internal/application/relationship"
	"github.com/nanfang/just-for-you/server/internal/infrastructure/clock"
	"github.com/nanfang/just-for-you/server/internal/infrastructure/config"
	momentmysql "github.com/nanfang/just-for-you/server/internal/infrastructure/persistence/mysql/moment"
	cos "github.com/nanfang/just-for-you/server/internal/infrastructure/storage/cos"
	momenthttp "github.com/nanfang/just-for-you/server/internal/interfaces/http/moment"
)

func buildMoment(cfg config.Config, db *sql.DB, get *relationshipapp.GetRelationshipUseCase) (*momenthttp.Handler, error) {
	repo := momentmysql.NewMySQLRepository(db)
	photos, err := cos.NewCOSStorage(cfg.COSBucketURL, cfg.COSSecretID, cfg.COSSecretKey)
	if err != nil {
		return nil, err
	}
	relationships := momentmysql.NewActiveRelationshipReader(get)
	useCases := momenthttp.UseCases{RequestPhotoUpload: momentapp.NewRequestPhotoUploadUseCase(repo, photos, clock.System{}), PublishMoment: momentapp.NewPublishMomentUseCase(relationships, repo, repo, photos, clock.System{}), UpdateMoment: momentapp.NewUpdateMomentUseCase(relationships, repo, repo, photos, clock.System{}), GetTimeline: momentapp.NewGetTimelineUseCase(relationships, repo, photos, clock.System{}), GetHome: momentapp.NewGetHomeUseCase(relationships, repo, photos, clock.System{}), DeleteMoment: momentapp.NewDeleteMomentUseCase(relationships, repo), ReadModels: momentapp.NewReadModels(relationships, repo)}
	return momenthttp.NewHandler(useCases), nil
}
