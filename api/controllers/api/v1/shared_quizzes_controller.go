package v1

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	goqu "github.com/doug-martin/goqu/v9"
	fiber "github.com/gofiber/fiber/v2"
	"github.com/randhir3-cloud/GK-Circle-v2/api/config"
	"github.com/randhir3-cloud/GK-Circle-v2/api/constants"
	quizUtilsHelper "github.com/randhir3-cloud/GK-Circle-v2/api/helpers/utils"
	"github.com/randhir3-cloud/GK-Circle-v2/api/internal/email"
	"github.com/randhir3-cloud/GK-Circle-v2/api/models"
	"github.com/randhir3-cloud/GK-Circle-v2/api/pkg/structs"
	"github.com/randhir3-cloud/GK-Circle-v2/api/services"
	"github.com/randhir3-cloud/GK-Circle-v2/api/utils"
	"go.uber.org/zap"
	validator "gopkg.in/go-playground/validator.v9"
)

type SharedQuizzes struct {
	sharedQuizzesModel *models.SharedQuizzesModel
	userModel          *models.UserModel
	emailService       *services.EmailService
	txEmailService     *email.TransactionalEmailService
	logger             *zap.Logger
	config             *config.AppConfig
}

func NewSharedQuizzesController(goqu *goqu.Database, logger *zap.Logger, config *config.AppConfig, txEmailService *email.TransactionalEmailService) (*SharedQuizzes, error) {

	sharedQuizzesModel := models.InitSharedQuizzesModel(goqu, logger)
	userModel, err := models.InitUserModel(goqu, logger)
	if err != nil {
		return nil, err
	}

	emailService := services.NewEmailService(logger, &config.SMTP)

	return &SharedQuizzes{
		sharedQuizzesModel: sharedQuizzesModel,
		userModel:          &userModel,
		emailService:       emailService,
		txEmailService:     txEmailService,
		logger:             logger,
		config:             config,
	}, nil
}

// ShareQuiz to insert data for share the quiz.
// swagger:route POST /v1/shared_quizzes ShareQuiz RequestShareQuiz
//
// share quiz to other user.
//
//		Consumes:
//		- application/json
//
//		Schemes: http, https
//
//		Responses:
//		  200: ResponseOkWithMessage
//	     400: GenericResFailNotFound
//		  500: GenericResError
func (sqctrl *SharedQuizzes) ShareQuiz(c *fiber.Ctx) error {
	user, ok := quizUtilsHelper.ConvertType[models.User](c.Locals(constants.ContextUser))
	if !ok {
		sqctrl.logger.Error(constants.ErrConvertTypeUser)
		return utils.JSONError(c, http.StatusInternalServerError, constants.ErrShareQuiz)
	}

	quizId := c.Params(constants.QuizId)
	sqctrl.logger.Debug("SharedQuizzes.ShareQuiz called", zap.Any("quizId", quizId), zap.Any("userId", user.ID))
	if quizId == "" {
		return utils.JSONError(c, http.StatusBadRequest, "No quiz_id found")
	}

	sqctrl.logger.Debug("validate req", zap.Any("Body", c.Body()))
	var shareQuizReq structs.ReqShareQuiz
	err := json.Unmarshal(c.Body(), &shareQuizReq)
	if err != nil {
		return utils.JSONFail(c, http.StatusBadRequest, err.Error())
	}

	validate := validator.New()
	err = validate.Struct(shareQuizReq)
	if err != nil {
		return utils.JSONFail(c, http.StatusBadRequest, utils.ValidatorErrorString(err))
	}
	sqctrl.logger.Debug("validate req success", zap.Any("shareQuizReq", shareQuizReq))

	// Insert for share quiz
	id, err := sqctrl.sharedQuizzesModel.InsertSharedQuiz(models.SharedQuizzes{
		QuizId:     quizId,
		SharedTo:   shareQuizReq.Email,
		SharedBy:   user.ID,
		Permission: shareQuizReq.Permission,
	})
	if err != nil {
		sqctrl.logger.Error(constants.ErrShareQuiz, zap.Error(err))
		return utils.JSONError(c, http.StatusInternalServerError, constants.ErrShareQuiz)
	}

	// Send Email to user to notify using Phase 2 TransactionalEmailService
	recipientName := "Candidate"
	emailParts := strings.Split(shareQuizReq.Email, "@")
	if len(emailParts) > 0 && emailParts[0] != "" {
		recipientName = emailParts[0]
	}

	inviterName := strings.TrimSpace(user.FirstName + " " + user.LastName)
	if inviterName == "" {
		inviterName = user.Email
	}

	_, sendErr := sqctrl.txEmailService.SendQuizInvitation(
		c.UserContext(),
		email.QuizInvitationInput{
			InvitationID: fmt.Sprintf("%v", id),
			Recipient: email.EmailRecipient{
				Name:    recipientName,
				Address: shareQuizReq.Email,
			},
			InviterName: inviterName,
			QuizTitle:   "GK Circle Quiz",
			QuizID:      quizId,
			ExpiresAt:   time.Now().Add(7 * 24 * time.Hour), // 7 days expiration
		},
	)

	if sendErr != nil {
		sqctrl.logger.Error("Failed to send quiz invitation email", zap.Error(sendErr))
		if errors.Is(sendErr, context.Canceled) || errors.Is(sendErr, context.DeadlineExceeded) {
			return utils.JSONError(c, http.StatusRequestTimeout, "Request timeout while sending invitation")
		}
		return utils.JSONError(c, http.StatusServiceUnavailable, "Unable to send the invitation email")
	}

	sqctrl.logger.Debug("SharedQuizzes.ShareQuiz success", zap.Any("quizId", quizId), zap.Any("userId", user.ID))
	return utils.JSONSuccess(c, http.StatusOK, id)
}

// ListQuizAuthorizedUsers to List authorized users for perticular quiz.
// swagger:route GET /v1/shared_quizzes/{quiz_id} ShareQuiz RequestListQuizAuthorizedUsers
//
// List authorized users for perticular quiz.
//
//		Consumes:
//		- application/json
//
//		Schemes: http, https
//
//		Responses:
//		  200: ResponseListQuizAuthorizedUsers
//	     400: GenericResFailNotFound
//		  500: GenericResError
func (sqctrl *SharedQuizzes) ListQuizAuthorizedUsers(c *fiber.Ctx) error {
	quizId := c.Params(constants.QuizId)
	sqctrl.logger.Debug("SharedQuizzes.ListQuizAuthorizedUsers called", zap.Any("quizId", quizId))
	if quizId == "" {
		return utils.JSONError(c, http.StatusBadRequest, "No quiz_id found")
	}

	quizAuthorizedUsers, err := sqctrl.sharedQuizzesModel.ListQuizAuthorizedUsersByQuizId(quizId)
	if err != nil {
		sqctrl.logger.Error(constants.ErrFetchAuthorizedUsersError, zap.Error(err))
		return utils.JSONError(c, http.StatusInternalServerError, constants.ErrFetchAuthorizedUsersError)
	}
	sqctrl.logger.Debug("SharedQuizzes.ListQuizAuthorizedUsers success", zap.Any("quizAuthorizedUsers", quizAuthorizedUsers))

	return utils.JSONSuccess(c, http.StatusOK, quizAuthorizedUsers)
}

// UpdateUserPermissionOfQuiz to Update authorized user permission for perticular quiz.
// swagger:route PUT /v1/shared_quizzes/{quiz_id} ShareQuiz RequestUpdateUserPermissionOfQuiz
//
// Update authorized user permission for perticular quiz.
//
//		Consumes:
//		- application/json
//
//		Schemes: http, https
//
//		Responses:
//		  200: ResponseOkWithMessage
//	     400: GenericResFailNotFound
//		  500: GenericResError
func (sqctrl *SharedQuizzes) UpdateUserPermissionOfQuiz(c *fiber.Ctx) error {
	sharedQuizId := c.Query(constants.SharedQuizId)
	sqctrl.logger.Debug("SharedQuizzes.UpdateUserPermissionOfQuiz called", zap.Any(constants.SharedQuizId, sharedQuizId))
	if sharedQuizId == "" {
		return utils.JSONError(c, http.StatusBadRequest, constants.BadRequestSharedQuizIdNotFound)
	}

	sqctrl.logger.Debug("validate req", zap.Any("Body", c.Body()))
	var shareQuizReq structs.ReqShareQuiz
	err := json.Unmarshal(c.Body(), &shareQuizReq)
	if err != nil {
		return utils.JSONFail(c, http.StatusBadRequest, err.Error())
	}

	validate := validator.New()
	err = validate.Struct(shareQuizReq)
	if err != nil {
		return utils.JSONFail(c, http.StatusBadRequest, utils.ValidatorErrorString(err))
	}
	sqctrl.logger.Debug("validate req success", zap.Any("shareQuizReq", shareQuizReq))

	err = sqctrl.sharedQuizzesModel.UpdateUserPermissionById(sharedQuizId, shareQuizReq)
	if err != nil {
		sqctrl.logger.Error(constants.ErrUpdateUserPermissionForQuiz, zap.Error(err))
		return utils.JSONError(c, http.StatusInternalServerError, constants.ErrUpdateUserPermissionForQuiz)
	}
	sqctrl.logger.Debug("SharedQuizzes.UpdateUserPermissionOfQuiz success", zap.Any("shareQuizReq", shareQuizReq))

	return utils.JSONSuccess(c, http.StatusOK, "User permission updated successfully!")
}

// DeleteUserPermissionOfQuiz to Delete authorized user permission for perticular quiz.
// swagger:route DELETE /v1/shared_quizzes/{quiz_id} ShareQuiz RequestDeleteUserPermissionOfQuiz
//
// Delete authorized user permission for perticular quiz.
//
//		Consumes:
//		- application/json
//
//		Schemes: http, https
//
//		Responses:
//		  200: ResponseOkWithMessage
//	     400: GenericResFailNotFound
//		  500: GenericResError
func (sqctrl *SharedQuizzes) DeleteUserPermissionOfQuiz(c *fiber.Ctx) error {
	sharedQuizId := c.Query(constants.SharedQuizId)
	sqctrl.logger.Debug("SharedQuizzes.DeleteUserPermissionOfQuiz called", zap.Any("sharedQuizId", sharedQuizId))
	if sharedQuizId == "" {
		return utils.JSONError(c, http.StatusBadRequest, constants.BadRequestSharedQuizIdNotFound)
	}

	err := sqctrl.sharedQuizzesModel.DeleteUserPermissionById(sharedQuizId)
	if err != nil {
		sqctrl.logger.Error(constants.ErrDeleteUserPermissionForQuiz, zap.Error(err))
		return utils.JSONError(c, http.StatusInternalServerError, constants.ErrDeleteUserPermissionForQuiz)
	}
	sqctrl.logger.Debug("SharedQuizzes.DeleteUserPermissionOfQuiz success", zap.Any("sharedQuizId", sharedQuizId))

	return utils.JSONSuccess(c, http.StatusOK, "User permission deleted successfully!")
}

// ListSharedQuizzes to List shared quiz for perticular user (only shared with the user or shared by the user).
// swagger:route GET /v1/shared_quizzes ShareQuiz RequestListSharedQuizzes
//
// List shared quiz for perticular user (only shared with the user or shared by the user).
//
//		Consumes:
//		- application/json
//
//		Schemes: http, https
//
//		Responses:
//		  200: ResponseListSharedQuizzes
//	     400: GenericResFailNotFound
//		  500: GenericResError
func (sqctrl *SharedQuizzes) ListSharedQuizzes(c *fiber.Ctx) error {
	requestType := c.Query("type")
	user, ok := quizUtilsHelper.ConvertType[models.User](c.Locals(constants.ContextUser))
	if !ok {
		sqctrl.logger.Error(constants.ErrConvertTypeUser)
		return utils.JSONError(c, http.StatusInternalServerError, constants.ErrListShareQuiz)
	}
	sqctrl.logger.Debug("SharedQuizzes.ListSharedQuizzes called", zap.Any("userId", user.ID), zap.Any("requestType", requestType))

	var sharedTo, sharedBy string
	switch requestType {
	case "shared_by_me":
		// Quizzes shared by the user
		sharedBy = user.ID
	case "shared_with_me":
		// Quizzes shared with the user
		sharedTo = user.Email
	default:
		return utils.JSONError(c, http.StatusBadRequest, "Invalid request type. Use 'shared_by_me' or 'shared_with_me'.")
	}

	sharedQuizzes, err := sqctrl.sharedQuizzesModel.ListSharedQuizzes(sharedBy, sharedTo)
	if err != nil {
		sqctrl.logger.Error(constants.ErrListShareQuiz, zap.Error(err))
		return utils.JSONError(c, http.StatusInternalServerError, constants.ErrListShareQuiz)
	}
	sqctrl.logger.Debug("SharedQuizzes.ListSharedQuizzes success", zap.Any("sharedQuizzes", sharedQuizzes))

	return utils.JSONSuccess(c, http.StatusOK, sharedQuizzes)
}
