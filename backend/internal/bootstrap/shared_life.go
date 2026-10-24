package bootstrap

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"net/http"
	"time"

	anniversaryapp "github.com/nanfang/just-for-you/server/internal/anniversary/application"
	anniversarymysql "github.com/nanfang/just-for-you/server/internal/anniversary/infrastructure/mysql"
	anniversaryhttp "github.com/nanfang/just-for-you/server/internal/anniversary/interfaces/http"
	coupleapp "github.com/nanfang/just-for-you/server/internal/couple/application"
	dailyquestionapp "github.com/nanfang/just-for-you/server/internal/dailyquestion/application"
	dailyai "github.com/nanfang/just-for-you/server/internal/dailyquestion/infrastructure/ai"
	dailyquestionmysql "github.com/nanfang/just-for-you/server/internal/dailyquestion/infrastructure/mysql"
	dailyquestionhttp "github.com/nanfang/just-for-you/server/internal/dailyquestion/interfaces/http"
	cosstorage "github.com/nanfang/just-for-you/server/internal/media/infrastructure/cos"
	"github.com/nanfang/just-for-you/server/internal/platform/clock"
	"github.com/nanfang/just-for-you/server/internal/platform/config"
	todoapp "github.com/nanfang/just-for-you/server/internal/todo/application"
	todomysql "github.com/nanfang/just-for-you/server/internal/todo/infrastructure/mysql"
	todohttp "github.com/nanfang/just-for-you/server/internal/todo/interfaces/http"
	wishapp "github.com/nanfang/just-for-you/server/internal/wish/application"
	wishmysql "github.com/nanfang/just-for-you/server/internal/wish/infrastructure/mysql"
	wishhttp "github.com/nanfang/just-for-you/server/internal/wish/interfaces/http"
)

func coverKey() string {
	value := make([]byte, 16)
	if _, err := rand.Read(value); err != nil {
		return "cover"
	}
	return hex.EncodeToString(value)
}

func buildSharedLife(
	db *sql.DB,
	cfg config.Config,
	getCouple *coupleapp.GetCurrentCouple,
) ([]RouteRegistrar, error) {
	storage, err := cosstorage.NewObjectStorage(
		cfg.COSBucketURL,
		cfg.COSSecretID,
		cfg.COSSecretKey,
	)
	if err != nil {
		return nil, err
	}
	anniversaryStore := anniversarymysql.NewMySQLAnniversaryStore(db)
	anniversaryRelationships := newAnniversaryCoupleReader(getCouple)
	anniversaryHandler := anniversaryhttp.NewHandler(anniversaryhttp.Dependencies{
		List: anniversaryapp.NewListAnniversaries(
			anniversaryRelationships,
			anniversaryStore,
			clock.System{},
		),
		Get:    anniversaryapp.NewGetAnniversary(anniversaryRelationships, anniversaryStore),
		Create: anniversaryapp.NewCreateAnniversary(anniversaryRelationships, anniversaryStore, clock.System{}),
		Update: anniversaryapp.NewUpdateAnniversary(
			anniversaryRelationships,
			anniversaryStore,
			storage,
			clock.System{},
		),
		Delete: anniversaryapp.NewDeleteAnniversary(anniversaryRelationships, anniversaryStore),
		RequestCover: anniversaryapp.NewRequestCoverUpload(
			anniversaryRelationships,
			anniversaryStore,
			storage,
			clock.System{},
			coverKey,
		),
		CoverURL: anniversaryapp.NewCoverURL(storage, clock.System{}),
	})

	todoStore := todomysql.NewMySQLTodoStore(db)
	todoRelationships := newTodoCoupleReader(getCouple)
	todoHandler := todohttp.NewHandler(todohttp.Dependencies{
		List:     todoapp.NewListTodos(todoRelationships, todoStore),
		Create:   todoapp.NewCreateTodo(todoRelationships, todoStore, clock.System{}),
		Complete: todoapp.NewSetTodoCompleted(todoRelationships, todoStore, clock.System{}),
		Delete:   todoapp.NewDeleteTodo(todoRelationships, todoStore),
	})

	wishStore := wishmysql.NewMySQLWishStore(db)
	wishRelationships := newWishCoupleReader(getCouple)
	wishHandler := wishhttp.NewHandler(wishhttp.Dependencies{
		List:   wishapp.NewListWishes(wishRelationships, wishStore),
		Create: wishapp.NewCreateWish(wishRelationships, wishStore, clock.System{}),
		Complete: wishapp.NewCompleteWish(
			wishRelationships,
			wishmysql.NewMomentCoupleReader(db),
			wishStore,
			clock.System{},
		),
		Reopen: wishapp.NewReopenWish(wishRelationships, wishStore, clock.System{}),
		Delete: wishapp.NewDeleteWish(wishRelationships, wishStore),
	})

	dailyStore := dailyquestionmysql.NewMySQLQuestionStore(db)
	fallback := dailyai.NewFallbackQuestions()
	var primary dailyquestionapp.QuestionGenerator
	if cfg.AIBaseURL != "" && cfg.AIAPIKey != "" && cfg.AIModel != "" {
		primary = dailyai.NewOpenAICompatible(
			cfg.AIBaseURL,
			cfg.AIAPIKey,
			cfg.AIModel,
			&http.Client{Timeout: 3 * time.Second},
		)
	}
	var generator dailyquestionapp.QuestionGenerator = fallback
	if primary != nil {
		generator = dailyquestionapp.PersonalizedQuestionGenerator{
			Preferences: dailyStore,
			Primary:     primary,
			Fallback:    fallback,
		}
	}
	dailyRelationships := newDailyQuestionCoupleReader(getCouple)
	today := dailyquestionapp.NewGetDailyQuestion(
		dailyRelationships,
		dailyStore,
		generator,
		clock.System{},
		dailyStore,
	)
	submit := dailyquestionapp.NewSubmitDailyAnswer(
		dailyRelationships,
		dailyquestionmysql.NewMySQLAnswerSubmissionStore(db, dailyStore),
		clock.System{},
		today,
	)
	list := dailyquestionapp.NewListDailyQuestions(dailyRelationships, dailyStore)
	dailyHandler := dailyquestionhttp.NewHandler(dailyquestionhttp.Dependencies{
		GetToday:     today,
		SubmitAnswer: submit,
		List:         list,
	})

	return []RouteRegistrar{
		anniversaryHandler,
		todoHandler,
		wishHandler,
		dailyHandler,
	}, nil
}
