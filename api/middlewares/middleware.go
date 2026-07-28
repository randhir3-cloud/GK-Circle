package middlewares

import (
	"github.com/randhir3-cloud/GK-Circle-v2/api/config"
	"github.com/randhir3-cloud/GK-Circle-v2/api/models"
	"github.com/doug-martin/goqu/v9"
	"go.uber.org/zap"
)

type Middleware struct {
	Config               config.AppConfig
	Logger               *zap.Logger
	Db                   *goqu.Database
	userModel            *models.UserModel
	sharedQuizzesModel   *models.SharedQuizzesModel
	userPlayedQuizModel  *models.UserPlayedQuizModel
}

func NewMiddleware(cfg config.AppConfig, logger *zap.Logger, db *goqu.Database) Middleware {

	userModel, _ := models.InitUserModel(db, logger)
	sharedQuizzesModel := models.InitSharedQuizzesModel(db, logger)
	userPlayedQuizModel := models.InitUserPlayedQuizModel(db)

	return Middleware{
		Config:              cfg,
		Logger:              logger,
		Db:                  db,
		userModel:           &userModel,
		sharedQuizzesModel:  sharedQuizzesModel,
		userPlayedQuizModel: userPlayedQuizModel,
	}
}
