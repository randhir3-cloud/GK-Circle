package models

import (
	"database/sql"

	"github.com/randhir3-cloud/GK-Circle-v2/api/config"
	"github.com/randhir3-cloud/GK-Circle-v2/api/constants"
	"github.com/doug-martin/goqu/v9"
)

// PlayedQuizReviewAccessContext links a played-quiz row to its live session and quiz.
type PlayedQuizReviewAccessContext struct {
	UserID         string `db:"user_id"`
	SessionAdminID string `db:"session_admin_id"`
	QuizID         string `db:"quiz_id"`
}

func (model *UserPlayedQuizModel) GetReviewAccessContext(userPlayedQuizID string) (*PlayedQuizReviewAccessContext, error) {
	var accessCtx PlayedQuizReviewAccessContext

	found, err := model.db.From(goqu.T(UserPlayedQuizTable).As("upq")).
		Select(
			goqu.I("upq.user_id"),
			goqu.I("aq.admin_id").As("session_admin_id"),
			goqu.I("aq.quiz_id").As("quiz_id"),
		).
		InnerJoin(
			goqu.T(ActiveQuizzesTable).As("aq"),
			goqu.On(goqu.I("upq.active_quiz_id").Eq(goqu.I("aq.id"))),
		).
		Where(goqu.Ex{"upq.id": userPlayedQuizID}).
		ScanStruct(&accessCtx)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, sql.ErrNoRows
	}

	return &accessCtx, nil
}

func HasQuizReviewPreviewAccess(
	requestingUserID string,
	editor *User,
	cfg *config.AppConfig,
	sharedQuizzesModel *SharedQuizzesModel,
	accessCtx PlayedQuizReviewAccessContext,
) (bool, error) {
	if requestingUserID == accessCtx.UserID || requestingUserID == accessCtx.SessionAdminID {
		return true, nil
	}

	if editor == nil {
		return false, nil
	}

	isCreator, err := sharedQuizzesModel.CheckQuizCreatorExists(accessCtx.QuizID, editor.ID)
	if err != nil {
		return false, err
	}
	if isCreator {
		return true, nil
	}

	if cfg.Quiz.IsPublicQuizAdmin(editor.Email) {
		isPublic, err := sharedQuizzesModel.IsQuizPublic(accessCtx.QuizID)
		if err != nil {
			return false, err
		}
		if isPublic {
			return true, nil
		}
	}

	permission, err := sharedQuizzesModel.GetPermissionByQuizAndUser(accessCtx.QuizID, editor.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}

	switch permission {
	case constants.ReadPermission, constants.WritePermission, constants.SharePermission:
		return true, nil
	default:
		return false, nil
	}
}

func (model *UserPlayedQuizModel) CanAccessReview(
	requestingUserID string,
	editor *User,
	cfg *config.AppConfig,
	sharedQuizzesModel *SharedQuizzesModel,
	userPlayedQuizID string,
) (bool, error) {
	accessCtx, err := model.GetReviewAccessContext(userPlayedQuizID)
	if err != nil {
		return false, err
	}

	return HasQuizReviewPreviewAccess(requestingUserID, editor, cfg, sharedQuizzesModel, *accessCtx)
}
