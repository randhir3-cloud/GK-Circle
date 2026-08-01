package routes

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"sync"

	"go.uber.org/zap"

	goqu "github.com/doug-martin/goqu/v9"
	"github.com/gofiber/contrib/swagger"
	"github.com/gofiber/contrib/websocket"
	fiber "github.com/gofiber/fiber/v2"
	"github.com/randhir3-cloud/GK-Circle-v2/api/config"
	"github.com/randhir3-cloud/GK-Circle-v2/api/constants"
	controller "github.com/randhir3-cloud/GK-Circle-v2/api/controllers/api/v1"
	"github.com/randhir3-cloud/GK-Circle-v2/api/internal/email"
	"github.com/randhir3-cloud/GK-Circle-v2/api/middlewares"
	pMetrics "github.com/randhir3-cloud/GK-Circle-v2/api/pkg/prometheus"
	"github.com/randhir3-cloud/GK-Circle-v2/api/pkg/redis"
	"github.com/randhir3-cloud/GK-Circle-v2/api/services"
	goredis "github.com/redis/go-redis/v9"
	"net/http"
)

var mu sync.Mutex

// Setup func
func Setup(app *fiber.App, goqu *goqu.Database, logger *zap.Logger, config config.AppConfig, pMetrics *pMetrics.PrometheusMetrics) error {
	mu.Lock()
	defer mu.Unlock()

	// plugins
	app.Use(middlewares.LogHandler(logger, pMetrics))

	swagger_file_path := "./assets/swagger.json"
	swagger_new_file_path := "./assets/new_swagger.json"

	err := newSwagger(swagger_file_path, swagger_new_file_path, config.WebUrl)
	if err != nil {
		return err
	}

	app.Use(swagger.New(swagger.Config{
		BasePath: "/api/v1/",
		FilePath: swagger_new_file_path,
		Path:     "docs",
		Title:    "Swagger API Docs",
	}))

	router := app.Group("/api")

	err = setupHealthCheckController(router, goqu, logger)
	if err != nil {
		return err
	}

	err = setupMetricsController(router, goqu, logger, pMetrics)
	if err != nil {
		return err
	}

	redis, err := redis.InitRedisPubSub(goqu, config.RedisClient, logger)

	if err != nil {
		return err
	}

	// middleware initialization
	middleware := middlewares.NewMiddleware(config, logger, goqu)

	v1 := router.Group("/v1")

	v1.Use("/socket", func(c *fiber.Ctx) error {

		if websocket.IsWebSocketUpgrade(c) {
			c.Locals(constants.MiddlewareError, nil)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	// FinalScoreboard
	err = setUpFinalScoreBoardController(v1, goqu, logger, middleware)
	if err != nil {
		return err
	}
	err = setUpAnalyticsBoardController(v1, goqu, logger, config, middleware)
	if err != nil {
		return err
	}

	err = setupAuthController(v1, goqu, logger, config, middleware)
	if err != nil {
		return err
	}

	err = setupUserController(v1, goqu, logger, middleware, config)
	if err != nil {
		return err
	}

	err = setupQuizSocketController(v1, goqu, logger, middleware, config, redis)
	if err != nil {
		return err
	}

	err = setupQuizController(v1, goqu, logger, middleware, config)
	if err != nil {
		return err
	}

	err = setupQuizCategoryController(v1, goqu, logger, middleware, config)
	if err != nil {
		return err
	}

	err = setupCourseController(v1, goqu, logger, middleware, config)
	if err != nil {
		return err
	}

	err = setupQuestionController(v1, goqu, logger, middleware, config)
	if err != nil {
		return err
	}

	err = setupQuestionCollectionController(v1, goqu, logger, middleware, config)
	if err != nil {
		return err
	}

	err = setupTestSnapshotController(v1, goqu, logger, middleware, config)
	if err != nil {
		return err
	}

	err = setupAssessmentAttemptController(v1, goqu, logger, middleware, config, redis)
	if err != nil {
		return err
	}

	err = setupAssessmentAnalyticsController(v1, goqu, logger, middleware, redis)
	if err != nil {
		return err
	}

	err = setupLearnerAnalyticsController(v1, goqu, logger, middleware, redis)
	if err != nil {
		return err
	}

	err = setupInstructorAnalyticsController(v1, goqu, logger, middleware, redis)
	if err != nil {
		return err
	}

	err = setupInstructorExportController(v1, goqu, logger, middleware, config)
	if err != nil {
		return err
	}

	err = setupQuizResultAdminController(v1, goqu, logger, middleware, config, redis)
	if err != nil {
		return err
	}

	err = setupUserPlayedQuizeController(v1, goqu, logger, middleware, config)
	if err != nil {
		return err
	}

	err = setupSharedQuizzesController(v1, goqu, logger, middleware, config)
	if err != nil {
		return err
	}

	return nil
}

func newSwagger(file_name, new_file, port string) error {
	// Verify Swagger file exists
	if _, err := os.Stat(file_name); os.IsNotExist(err) {
		return fmt.Errorf("%s file does not exist", file_name)
	}

	// Read Swagger Spec into memory
	rawSpec, err := os.ReadFile(file_name)
	if err != nil {
		return fmt.Errorf("failed to read provided Swagger file (%s): %v", file_name, err.Error())
	}

	// Validate we have valid JSON or YAML
	var jsonData map[string]interface{}
	errJSON := json.Unmarshal(rawSpec, &jsonData)
	if errJSON != nil {
		return fmt.Errorf("swagger-json is not in valid format")
	}
	jsonData["host"] = port

	newData, err := json.MarshalIndent(jsonData, "", "   ")
	if err != nil {
		return fmt.Errorf("error during host change in swagger")
	}

	file, err := os.Create(new_file)
	if err != nil {
		return fmt.Errorf("error creating file: %v", err)
	}
	defer file.Close()

	_, err = file.Write(newData)

	return err
}

func setupAuthController(v1 fiber.Router, goqu *goqu.Database, logger *zap.Logger, config config.AppConfig, middlewares middlewares.Middleware) error {
	authController, err := controller.NewAuthController(goqu, logger, config)
	if err != nil {
		return err
	}

	if config.Kratos.IsEnabled {
		kratos := v1.Group("/kratos")
		kratos.Get("/auth", authController.DoKratosAuth)
		kratos.Get("/whoami", authController.GetRegisteredUser)
		kratos.Put("/user", middlewares.KratosAuthenticated, middlewares.VerificationRequired, authController.UpadateRegisteredUser)
		kratos.Delete("/user", middlewares.KratosAuthenticated, middlewares.VerificationRequired, authController.DeleteRegisteredUser)
		kratos.Delete("/e2e-cleanup", authController.E2ECleanup)
	}
	return nil
}

func setupUserController(v1 fiber.Router, goqu *goqu.Database, logger *zap.Logger, middlewares middlewares.Middleware, config config.AppConfig) error {
	userController, err := controller.NewUserController(goqu, logger, config)
	if err != nil {
		return err
	}

	// user route
	userRouter := v1.Group("/user")
	userRouter.Get("/who", middlewares.Authenticated, userController.GetUserMeta)
	userRouter.Post(fmt.Sprintf("/:%s", constants.Username), userController.CreateGuestUser)

	return nil
}

func setupHealthCheckController(api fiber.Router, goqu *goqu.Database, logger *zap.Logger) error {
	healthController, err := controller.NewHealthController(goqu, logger)
	if err != nil {
		return err
	}

	healthz := api.Group("/healthz")
	healthz.Get("/", healthController.Overall)
	healthz.Get("/db", healthController.Db)
	return nil
}

func setupMetricsController(api fiber.Router, db *goqu.Database, logger *zap.Logger, pMetrics *pMetrics.PrometheusMetrics) error {
	metricsController, err := controller.InitMetricsController(db, logger, pMetrics)
	if err != nil {
		return nil
	}

	api.Get("/metrics", metricsController.Metrics)
	return nil
}

func setupQuizSocketController(v1 fiber.Router, db *goqu.Database, logger *zap.Logger, middleware middlewares.Middleware, config config.AppConfig, redis *redis.RedisPubSub) error {
	quizSocketController, err := controller.InitQuizConfig(db, &config, logger, redis)
	if err != nil {
		return err
	}

	// CustomAuthenticated (not kratos-only) so guests can host public quizzes too.
	// GetOrActivateSession still enforces admin_id == userId, so a guest can only
	// arrange a session they themselves created.
	v1.Get(fmt.Sprintf("/socket/admin/arrange/:%s", constants.SessionIDParam), middleware.CheckSessionId, middleware.CustomAuthenticated, websocket.New(quizSocketController.Arrange))
	v1.Get(fmt.Sprintf("/socket/join/:%s", constants.QuizSessionInvitationCode), middleware.CheckSessionCode, middleware.CustomAuthenticated, websocket.New(quizSocketController.Join))
	v1.Post("/quiz/answer", middleware.Authenticated, middleware.CustomAuthenticated, quizSocketController.SetAnswer)
	v1.Get("/quiz/terminate", middleware.Authenticated, quizSocketController.Terminate)

	return nil
}

func setupQuizController(v1 fiber.Router, db *goqu.Database, logger *zap.Logger, middleware middlewares.Middleware, config config.AppConfig) error {
	quizController, err := controller.InitQuizController(db, logger, &config)
	if err != nil {
		return err
	}

	admin := v1.Group("/admin")
	admin.Use(middleware.KratosAuthenticated)
	admin.Use(middleware.VerificationRequired)

	// Public quizzes listing — must be registered BEFORE the auth-guarded /quizzes group
	// so it is not protected by KratosAuthenticated.
	v1.Get("/quizzes/public", quizController.GetPublicQuizzes)

	// Hosting a public quiz is open to guests as well as registered users, so it uses
	// Authenticated (guest JWT or kratos) instead of the kratos-only /quizzes group.
	// The handler itself enforces that the quiz is actually public.
	v1.Post(fmt.Sprintf("/quizzes/:%s/public_session", constants.QuizId), middleware.Authenticated, quizController.GeneratePublicSession)

	quizzes := v1.Group("/quizzes")
	quizzes.Use(middleware.KratosAuthenticated)
	quizzes.Use(middleware.VerificationRequired)

	quizzes.Post(fmt.Sprintf("/:%s/demo_session", constants.QuizId), quizController.GenerateDemoSession)
	quizzes.Post(fmt.Sprintf("/:%s/upload", constants.QuizTitle), middleware.ValidateCsv, middleware.KratosAuthenticated, quizController.CreateQuizByCsv)
	quizzes.Post("/", quizController.CreateQuiz)
	quizzes.Get("/", quizController.GetAdminUploadedQuizzes)
	quizzes.Put(fmt.Sprintf("/:%s/settings", constants.QuizId), middleware.QuizPermission, middleware.VerifyQuizEditAccess, quizController.UpdateQuizSettings)
	quizzes.Delete(fmt.Sprintf("/:%s", constants.QuizId), middleware.QuizPermission, middleware.VerifyQuizEditAccess, quizController.DeleteQuizById)

	report := admin.Group("/reports")
	report.Get("/list", quizController.ListQuizzesAnalysis)
	report.Get(fmt.Sprintf("/:%s/analysis", constants.ActiveQuizId), middleware.KratosAuthenticated, quizController.GetQuizAnalysis)
	return nil
}

func setupQuizCategoryController(v1 fiber.Router, db *goqu.Database, logger *zap.Logger, middleware middlewares.Middleware, config config.AppConfig) error {
	quizCategoryController, err := controller.InitQuizCategoryController(db, logger, &config)
	if err != nil {
		return err
	}

	// Category listing is public — the homepage groups public quizzes by category.
	// Registered BEFORE the auth-guarded /categories group so it stays open.
	v1.Get("/categories", quizCategoryController.ListCategories)

	// Managing categories requires Kratos auth; on top of that, each handler
	// enforces the public-quiz admin email allowlist (same gate as publishing).
	categories := v1.Group("/categories")
	categories.Use(middleware.KratosAuthenticated)
	categories.Use(middleware.VerificationRequired)

	categories.Post("/", quizCategoryController.CreateCategory)
	categories.Put(fmt.Sprintf("/:%s", constants.CategoryId), quizCategoryController.UpdateCategory)
	categories.Delete(fmt.Sprintf("/:%s", constants.CategoryId), quizCategoryController.DeleteCategory)

	return nil
}

func setupCourseController(v1 fiber.Router, db *goqu.Database, logger *zap.Logger, middleware middlewares.Middleware, config config.AppConfig) error {
	courseController, err := controller.InitCourseController(db, logger, &config)
	if err != nil {
		return err
	}
	courseNodeController, err := controller.InitCourseNodeController(db, logger, &config)
	if err != nil {
		return err
	}
	learningItemController, err := controller.InitLearningItemController(db, logger, &config)
	if err != nil {
		return err
	}
	learnerLearningItemController, err := controller.InitLearnerLearningItemController(db, logger, &config)
	if err != nil {
		return err
	}
	learnerCourseEnrollmentController, err := controller.InitLearnerCourseEnrollmentController(db, logger, &config)
	if err != nil {
		return err
	}
	learnerCourseController, err := controller.InitLearnerCourseController(db, logger, &config)
	if err != nil {
		return err
	}

	// Admin Course APIs reuse KratosAuthenticated for identity, then the existing
	// administrative allowlist inside the controller for Course authorisation.
	admin := v1.Group("/admin")
	admin.Use(middleware.KratosAuthenticated)
	admin.Use(middleware.VerificationRequired)

	courses := admin.Group("/courses")
	courses.Post("/", courseController.CreateCourse)
	courses.Get("/", courseController.ListCourses)

	// CourseNode routes must register /tree and /reorder before /:node_id.
	nodes := courses.Group(fmt.Sprintf("/:%s/nodes", constants.CourseId))
	nodes.Post("/", courseNodeController.CreateNode)
	nodes.Get("/", courseNodeController.ListRoots)
	nodes.Get("/tree", courseNodeController.GetTree)
	nodes.Post("/reorder", courseNodeController.ReorderChildren)

	// LearningItem routes nest under a node; register /reorder and /move before /:item_id.
	learningItems := nodes.Group(fmt.Sprintf("/:%s/learning-items", constants.NodeId))
	learningItems.Post("/", learningItemController.Create)
	learningItems.Get("/", learningItemController.List)
	learningItems.Post("/reorder", learningItemController.Reorder)
	learningItems.Post("/move", learningItemController.Move)
	learningItems.Get(fmt.Sprintf("/:%s", constants.ItemId), learningItemController.GetByID)
	learningItems.Patch(fmt.Sprintf("/:%s", constants.ItemId), learningItemController.Update)
	learningItems.Delete(fmt.Sprintf("/:%s", constants.ItemId), learningItemController.Delete)

	nodes.Get(fmt.Sprintf("/:%s", constants.NodeId), courseNodeController.GetByID)
	nodes.Get(fmt.Sprintf("/:%s/children", constants.NodeId), courseNodeController.ListChildren)
	nodes.Patch(fmt.Sprintf("/:%s/move", constants.NodeId), courseNodeController.MoveNode)
	nodes.Delete(fmt.Sprintf("/:%s", constants.NodeId), courseNodeController.DeleteSubtree)

	courses.Get(fmt.Sprintf("/:%s", constants.CourseId), courseController.GetCourse)
	courses.Patch(fmt.Sprintf("/:%s", constants.CourseId), courseController.UpdateCourse)

	// Learner Course APIs: Public catalog & outlines + Authenticated enrollment & LearningItem delivery.
	learner := v1.Group("/learner")
	learnerCourses := learner.Group("/courses")
	learnerCourses.Get("/", learnerCourseController.ListPublishedCourses)
	learnerCourses.Get(fmt.Sprintf("/:%s", constants.CourseId), learnerCourseController.GetPublishedCourse)
	learnerCourses.Get(fmt.Sprintf("/:%s/nodes/tree", constants.CourseId), learnerCourseController.GetPublishedOutline)

	learnerAuth := learner.Group("")
	learnerAuth.Use(middleware.KratosAuthenticated)
	learnerAuthCourses := learnerAuth.Group("/courses")
	learnerAuthCourses.Get(fmt.Sprintf("/:%s/enrollment", constants.CourseId), learnerCourseEnrollmentController.GetEnrollment)
	learnerAuthCourses.Post(fmt.Sprintf("/:%s/enrollment", constants.CourseId), learnerCourseEnrollmentController.Enroll)
	learnerAuthCourses.Delete(fmt.Sprintf("/:%s/enrollment", constants.CourseId), learnerCourseEnrollmentController.Unenroll)
	learnerNodes := learnerAuthCourses.Group(fmt.Sprintf("/:%s/nodes", constants.CourseId))
	learnerItems := learnerNodes.Group(fmt.Sprintf("/:%s/learning-items", constants.NodeId))
	learnerItems.Get("/", learnerLearningItemController.List)
	learnerItems.Get(fmt.Sprintf("/:%s", constants.ItemId), learnerLearningItemController.GetByID)

	return nil
}

func setupQuestionController(v1 fiber.Router, db *goqu.Database, logger *zap.Logger, middleware middlewares.Middleware, config config.AppConfig) error {
	questionController, err := controller.InitQuestionController(db, logger, &config)
	if err != nil {
		return err
	}

	questionRouter := v1.Group(fmt.Sprintf("/quizzes/:%s/questions", constants.QuizId))
	questionRouter.Use(middleware.KratosAuthenticated, middleware.QuizPermission)

	questionRouter.Get("/", questionController.ListQuestionsWithAnswerByQuizId)
	questionRouter.Post("/", middleware.VerifyQuizEditAccess, questionController.CreateQuestion)
	questionRouter.Post("/upload", middleware.VerifyQuizEditAccess, middleware.ValidateCsv, questionController.ImportQuestionsByCsv)
	questionRouter.Post("/import-jobs", middleware.VerifyQuizEditAccess, middleware.ValidateCsv, questionController.CreateQuestionImportJob)
	questionRouter.Get(fmt.Sprintf("/import-jobs/:%s", constants.ImportJobId), middleware.VerifyQuizEditAccess, questionController.GetQuestionImportJob)
	questionRouter.Post(fmt.Sprintf("/import-jobs/:%s/commit", constants.ImportJobId), middleware.VerifyQuizEditAccess, questionController.CommitQuestionImportJob)
	questionRouter.Get(fmt.Sprintf("/:%s", constants.QuestionId), middleware.VerifyQuizEditAccess, questionController.GetQuestionById)
	questionRouter.Get(fmt.Sprintf("/:%s/revisions", constants.QuestionId), middleware.VerifyQuizEditAccess, questionController.ListQuestionRevisions)
	questionRouter.Put(fmt.Sprintf("/:%s", constants.QuestionId), middleware.VerifyQuizEditAccess, questionController.UpdateQuestionById)
	questionRouter.Delete(fmt.Sprintf("/:%s", constants.QuestionId), middleware.VerifyQuizEditAccess, questionController.DeleteQuestionById)

	return nil
}

func setupQuestionCollectionController(v1 fiber.Router, db *goqu.Database, logger *zap.Logger, middleware middlewares.Middleware, config config.AppConfig) error {
	collectionController, err := controller.InitQuestionCollectionController(db, logger, &config)
	if err != nil {
		return err
	}

	collectionRouter := v1.Group(fmt.Sprintf("/quizzes/:%s/collections", constants.QuizId))
	collectionRouter.Use(middleware.KratosAuthenticated, middleware.QuizPermission, middleware.VerifyQuizEditAccess)

	collectionRouter.Get("/", collectionController.ListCollections)
	collectionRouter.Post("/", collectionController.CreateCollection)
	collectionRouter.Get(fmt.Sprintf("/:%s", constants.CollectionId), collectionController.GetCollection)
	collectionRouter.Patch(fmt.Sprintf("/:%s", constants.CollectionId), collectionController.UpdateCollection)
	collectionRouter.Delete(fmt.Sprintf("/:%s", constants.CollectionId), collectionController.DeleteCollection)
	collectionRouter.Put(fmt.Sprintf("/:%s/members", constants.CollectionId), collectionController.ReplaceMembers)
	collectionRouter.Get(fmt.Sprintf("/:%s/resolve", constants.CollectionId), collectionController.ResolveCollection)

	return nil
}

func setupTestSnapshotController(v1 fiber.Router, db *goqu.Database, logger *zap.Logger, middleware middlewares.Middleware, config config.AppConfig) error {
	snapshotController, err := controller.InitTestSnapshotController(db, logger, &config)
	if err != nil {
		return err
	}

	snapshotRouter := v1.Group(fmt.Sprintf("/quizzes/:%s/test-snapshots", constants.QuizId))
	snapshotRouter.Use(middleware.KratosAuthenticated, middleware.QuizPermission, middleware.VerifyQuizEditAccess)

	snapshotRouter.Get("/", snapshotController.ListSnapshots)
	snapshotRouter.Post("/", snapshotController.CreateSnapshot)
	snapshotRouter.Get(fmt.Sprintf("/:%s", constants.SnapshotId), snapshotController.GetSnapshot)
	snapshotRouter.Get(fmt.Sprintf("/:%s/learner", constants.SnapshotId), snapshotController.GetLearnerSnapshot)

	return nil
}

func setupAssessmentAttemptController(v1 fiber.Router, db *goqu.Database, logger *zap.Logger, middleware middlewares.Middleware, config config.AppConfig, redisClient *redis.RedisPubSub) error {
	attemptController, err := controller.InitAssessmentAttemptController(db, logger, &config)
	if err != nil {
		return err
	}
	if redisClient != nil && redisClient.PubSubModel != nil {
		cache := services.NewLearnerAnalyticsCache(&redisClient.PubSubModel.Client, logger)
		attemptController.SetLearnerAnalyticsCache(cache)
	}

	attemptRouter := v1.Group(fmt.Sprintf("/quizzes/:%s/attempts", constants.QuizId))
	attemptRouter.Use(middleware.EnsureCorrelationID, middleware.KratosAuthenticated)

	attemptRouter.Get("/instructions", attemptController.GetInstructions)
	attemptRouter.Get("/", attemptController.ListMyAttempts)
	attemptRouter.Post("/", attemptController.CreateAttempt)
	attemptRouter.Get(fmt.Sprintf("/:%s/resume", constants.AttemptId), attemptController.ResumeAttempt)
	attemptRouter.Get(fmt.Sprintf("/:%s/status", constants.AttemptId), attemptController.GetAttemptStatus)
	attemptRouter.Post(fmt.Sprintf("/:%s/submit", constants.AttemptId), attemptController.SubmitAttempt)
	attemptRouter.Get(fmt.Sprintf("/:%s/result", constants.AttemptId), attemptController.GetAttemptResult)
	attemptRouter.Put(fmt.Sprintf("/:%s/answers/:%s", constants.AttemptId, constants.QuestionId), attemptController.AutosaveAnswer)
	attemptRouter.Get(fmt.Sprintf("/:%s", constants.AttemptId), attemptController.GetMyAttempt)

	editorAttemptRouter := v1.Group(fmt.Sprintf("/quizzes/:%s/attempts", constants.QuizId))
	editorAttemptRouter.Use(middleware.EnsureCorrelationID, middleware.KratosAuthenticated, middleware.QuizPermission, middleware.VerifyQuizEditAccess)
	editorAttemptRouter.Get(fmt.Sprintf("/:%s/editor", constants.AttemptId), attemptController.GetEditorAttempt)

	return nil
}

func setupAssessmentAnalyticsController(v1 fiber.Router, db *goqu.Database, logger *zap.Logger, middleware middlewares.Middleware, redisClient *redis.RedisPubSub) error {
	analyticsController := controller.NewAssessmentAnalyticsController(db, logger)
	if redisClient != nil && redisClient.PubSubModel != nil {
		cache := services.NewLearnerAnalyticsCache(&redisClient.PubSubModel.Client, logger)
		analyticsController.SetLearnerAnalyticsCache(cache)
	}

	analyticsRouter := v1.Group(fmt.Sprintf("/quizzes/:%s/attempts/:%s/analytics/events", constants.QuizId, constants.AttemptId))
	analyticsRouter.Use(middleware.EnsureCorrelationID, middleware.KratosAuthenticated)
	analyticsRouter.Post("/", analyticsController.RecordClientTelemetryBatch)
	analyticsRouter.Get("/", analyticsController.GetAttemptEvents)
	return nil
}

func setupLearnerAnalyticsController(v1 fiber.Router, db *goqu.Database, logger *zap.Logger, middleware middlewares.Middleware, redisClient *redis.RedisPubSub) error {
	var redisGoClient *goredis.Client
	if redisClient != nil && redisClient.PubSubModel != nil {
		redisGoClient = &redisClient.PubSubModel.Client
	}
	analyticsController := controller.NewLearnerAnalyticsController(db, redisGoClient, logger)

	analyticsRouter := v1.Group("/analytics")
	analyticsRouter.Use(middleware.EnsureCorrelationID, middleware.KratosAuthenticated)
	analyticsRouter.Get("/dashboard", analyticsController.GetDashboard)
	analyticsRouter.Get("/activity", analyticsController.GetRecentActivity)
	analyticsRouter.Get("/trends", analyticsController.GetPerformanceTrends)
	analyticsRouter.Get("/subjects", analyticsController.GetSubjectPerformance)
	analyticsRouter.Get(fmt.Sprintf("/attempts/:%s/timeline", constants.AttemptId), analyticsController.GetAttemptTimeline)
	return nil
}

func setupInstructorAnalyticsController(v1 fiber.Router, db *goqu.Database, logger *zap.Logger, middleware middlewares.Middleware, redisClient *redis.RedisPubSub) error {
	var redisGoClient *goredis.Client
	if redisClient != nil && redisClient.PubSubModel != nil {
		redisGoClient = &redisClient.PubSubModel.Client
	}
	cache := services.NewInstructorAnalyticsCache(redisGoClient, logger)
	svc := services.NewInstructorAnalyticsService(db, cache, logger)
	analyticsController := controller.NewInstructorAnalyticsController(svc, logger)

	portfolioRouter := v1.Group("/instructor/analytics")
	portfolioRouter.Use(middleware.EnsureCorrelationID, middleware.KratosAuthenticated)
	portfolioRouter.Get("/overview", analyticsController.GetOverview)
	portfolioRouter.Get("/quizzes", analyticsController.GetQuizList)
	portfolioRouter.Get("/learners", analyticsController.GetLearnerList)
	portfolioRouter.Get("/releases", analyticsController.GetReleaseMonitoring)
	portfolioRouter.Get("/timeline", analyticsController.GetTimeline)

	quizRouter := v1.Group(fmt.Sprintf("/quizzes/:%s/analytics", constants.QuizId))
	quizRouter.Use(middleware.EnsureCorrelationID, middleware.KratosAuthenticated, middleware.LoadQuizAnalyticsContext, middleware.VerifyQuizAnalyticsAccess)
	quizRouter.Get("/summary", analyticsController.GetQuizSummary)
	quizRouter.Get("/attempts", analyticsController.GetQuizAttempts)
	quizRouter.Get("/questions", analyticsController.GetQuestionMetrics)
	quizRouter.Get("/engagement", analyticsController.GetEngagement)
	return nil
}

// final score board controller setup
func setUpFinalScoreBoardController(v1 fiber.Router, goqu *goqu.Database, logger *zap.Logger, middlewares middlewares.Middleware) error {
	finalScoreBoardController, err := controller.NewFinalScoreBoardController(goqu, logger)
	if err != nil {
		return err
	}

	finalScoreBoardControllerAdmin, err := controller.NewFinalScoreBoardAdminController(goqu, logger)
	if err != nil {
		return err
	}

	finalScore := v1.Group("/final_score")
	finalScore.Get("/user", middlewares.Authenticated, middlewares.VerifyPlayedQuizReviewAccess, finalScoreBoardController.GetScore)
	finalScore.Get("/admin", middlewares.KratosAuthenticated, finalScoreBoardControllerAdmin.GetScoreForAdmin)

	return nil
}

func setUpAnalyticsBoardController(v1 fiber.Router, goqu *goqu.Database, logger *zap.Logger, config config.AppConfig, middlewares middlewares.Middleware) error {
	analyticsBoardUserController, err := controller.NewAnalyticsBoardUserController(goqu, logger, &config)
	if err != nil {
		return err
	}

	analyticsBoardAdminController, err := controller.NewAnalyticsBoardAdminController(goqu, logger, &config)
	if err != nil {
		return err
	}

	analyticsBoard := v1.Group("/analytics_board")
	analyticsBoard.Get("/user", middlewares.Authenticated, middlewares.VerifyPlayedQuizReviewAccess, analyticsBoardUserController.GetAnalyticsForUser)
	analyticsBoard.Get("/admin", middlewares.KratosAuthenticated, analyticsBoardAdminController.GetAnalyticsForAdmin)

	return nil
}

func setupUserPlayedQuizeController(v1 fiber.Router, goqu *goqu.Database, logger *zap.Logger, middlewares middlewares.Middleware, config config.AppConfig) error {
	userPlayedQuizeController, err := controller.NewUserPlayedQuizeController(goqu, logger, &config)
	if err != nil {
		return err
	}

	userRouter := v1.Group("/user_played_quizes")
	userRouter.Get("/", middlewares.KratosAuthenticated, userPlayedQuizeController.ListUserPlayedQuizes)
	userRouter.Get(fmt.Sprintf("/:%s", constants.UserPlayedQuizId), middlewares.Authenticated, middlewares.VerifyPlayedQuizReviewAccess, userPlayedQuizeController.ListUserPlayedQuizesWithQuestionById)
	userRouter.Post(fmt.Sprintf("/:%s", constants.QuizSessionInvitationCode), middlewares.Authenticated, userPlayedQuizeController.PlayedQuizValidation)
	return nil
}

func setupSharedQuizzesController(v1 fiber.Router, goqu *goqu.Database, logger *zap.Logger, middlewares middlewares.Middleware, config config.AppConfig) error {
	// Initialize Phase 2 transactional email service
	if err := config.Email.Validate(config.Env); err != nil {
		logger.Fatal("Failed to validate transactional email config", zap.Error(err))
	}

	renderer, err := email.NewTemplateRenderer()
	if err != nil {
		logger.Fatal("Failed to initialize template renderer", zap.Error(err))
	}

	urlBuilder, err := email.NewAppURLBuilder(config.WebUrl, config.Env)
	if err != nil {
		logger.Fatal("Failed to initialize app URL builder", zap.Error(err))
	}

	client := &http.Client{
		Timeout: config.Email.HTTPTimeout,
	}

	provider, err := email.NewProvider(config.Email, logger, client, email.SystemClock{}, email.ContextSleeper{})
	if err != nil {
		logger.Fatal("Failed to initialize email provider", zap.Error(err))
	}

	txEmailService := email.NewTransactionalEmailService(
		config.Email,
		provider,
		renderer,
		urlBuilder,
		logger,
		email.SystemClock{},
		email.NopMetricsHook{},
	)

	sharedQuizzesController, err := controller.NewSharedQuizzesController(goqu, logger, &config, txEmailService)
	if err != nil {
		return err
	}

	sharedQuizzesRouter := v1.Group("/shared_quizzes")
	sharedQuizzesRouter.Use(middlewares.KratosAuthenticated)

	sharedQuizzesRouter.Get("/", sharedQuizzesController.ListSharedQuizzes)
	sharedQuizzesRouter.Post(fmt.Sprintf("/:%s", constants.QuizId), middlewares.QuizPermission, middlewares.VerifyQuizShareAccess, sharedQuizzesController.ShareQuiz)
	sharedQuizzesRouter.Get(fmt.Sprintf("/:%s", constants.QuizId), middlewares.QuizPermission, middlewares.VerifyQuizShareAccess, sharedQuizzesController.ListQuizAuthorizedUsers)
	sharedQuizzesRouter.Put(fmt.Sprintf("/:%s", constants.QuizId), middlewares.QuizPermission, middlewares.VerifyQuizShareAccess, sharedQuizzesController.UpdateUserPermissionOfQuiz)
	sharedQuizzesRouter.Delete(fmt.Sprintf("/:%s", constants.QuizId), middlewares.QuizPermission, middlewares.VerifyQuizShareAccess, sharedQuizzesController.DeleteUserPermissionOfQuiz)
	return nil
}

func setupQuizResultAdminController(v1 fiber.Router, db *goqu.Database, logger *zap.Logger, middleware middlewares.Middleware, config config.AppConfig, redisClient *redis.RedisPubSub) error {
	adminService := services.NewQuizResultAdminService(db, logger)
	if redisClient != nil && redisClient.PubSubModel != nil {
		cache := services.NewLearnerAnalyticsCache(&redisClient.PubSubModel.Client, logger)
		adminService.SetLearnerAnalyticsCache(cache)
	}
	adminController := controller.NewQuizResultAdminController(adminService, logger)

	resultsAdminRouter := v1.Group(fmt.Sprintf("/quizzes/:%s/results", constants.QuizId))
	resultsAdminRouter.Use(middleware.EnsureCorrelationID, middleware.KratosAuthenticated, middleware.QuizPermission, middleware.VerifyQuizEditAccess)

	resultsAdminRouter.Patch("/settings", adminController.UpdateResultSettings)
	resultsAdminRouter.Post("/release", adminController.ReleaseResults)
	resultsAdminRouter.Get("/status", adminController.GetReleaseStatus)

	return nil
}

func setupInstructorExportController(v1 fiber.Router, db *goqu.Database, logger *zap.Logger, middleware middlewares.Middleware, cfg config.AppConfig) error {
	// Build storage provider based on config.
	var storage services.StorageProvider
	var storageErr error

	providerName := cfg.Report.StorageProvider
	if providerName == "" {
		if cfg.Env == "production" {
			providerName = "s3"
		} else {
			providerName = "local"
		}
	}

	switch providerName {
	case "s3":
		storage, storageErr = services.NewS3StorageProvider(
			context.Background(),
			cfg.Report.S3Bucket,
			cfg.Report.S3Region,
			cfg.Report.S3Endpoint,
			"", "", // credentials from env / EC2 metadata
			logger,
		)
	default:
		basePath := cfg.Report.LocalStoragePath
		if basePath == "" {
			basePath = "./data/reports"
		}
		storage, storageErr = services.NewLocalStorageProvider(basePath, cfg.Secret, logger)
	}
	if storageErr != nil {
		logger.Warn("export: storage provider init failed; export disabled", zap.Error(storageErr))
		return nil // Non-fatal: system runs without export if storage misconfigured.
	}

	// Build services.
	analyticsSvc := services.NewInstructorAnalyticsService(db, services.NewInstructorAnalyticsCache(nil, logger), logger)
	retentionDays := cfg.Report.RetentionDays
	if retentionDays == 0 {
		retentionDays = 30
	}
	exportSvc := services.NewExportService(db, analyticsSvc, storage, retentionDays, logger)

	emailSvc := services.NewEmailService(logger, &cfg.SMTP)
	maxAttach := cfg.Report.MaxAttachmentBytes
	if maxAttach == 0 {
		maxAttach = 10 * 1024 * 1024
	}
	reportEmailSvc := services.NewReportEmailService(emailSvc, db, storage, maxAttach, logger)

	// Build and start worker + scheduler.
	workerPoolSize := cfg.Report.WorkerPoolSize
	if workerPoolSize == 0 {
		workerPoolSize = 3
	}
	reclaimInterval := cfg.Report.ReclaimIntervalSeconds
	if reclaimInterval == 0 {
		reclaimInterval = 30
	}
	bgCtx := context.Background()
	logger.Info("workers initialized", zap.Int("pool_size", workerPoolSize), zap.Int("reclaim_interval_seconds", reclaimInterval))
	worker, jobQueue := services.NewReportWorker(db, exportSvc, reportEmailSvc, workerPoolSize, reclaimInterval, logger)
	worker.Start(bgCtx)

	schedulerInterval := cfg.Report.SchedulerIntervalSeconds
	if schedulerInterval == 0 {
		schedulerInterval = 60
	}

	// Validate scheduler timeout
	if val, present := os.LookupEnv("REPORT_SCHEDULER_TIMEOUT_SECONDS"); present {
		if val == "0" {
			return fmt.Errorf("invalid REPORT_SCHEDULER_TIMEOUT_SECONDS: zero duration is rejected")
		}
		if timeoutVal, err := strconv.Atoi(val); err != nil {
			return fmt.Errorf("invalid REPORT_SCHEDULER_TIMEOUT_SECONDS: malformed integer %q: %w", val, err)
		} else if timeoutVal < 0 {
			return fmt.Errorf("invalid REPORT_SCHEDULER_TIMEOUT_SECONDS: negative duration %d is rejected", timeoutVal)
		} else if timeoutVal > 300 {
			return fmt.Errorf("invalid REPORT_SCHEDULER_TIMEOUT_SECONDS: exceeds reasonable maximum of 300s (got %d)", timeoutVal)
		}
	}

	schedulerTimeout := cfg.Report.SchedulerTimeoutSeconds
	if schedulerTimeout == 0 {
		schedulerTimeout = 10 // safe default
	}

	scheduler := services.NewReportScheduler(db, jobQueue, schedulerInterval, schedulerTimeout, logger)
	scheduler.Start(bgCtx)

	retentionSvc := services.NewReportRetentionService(db, storage, logger)
	retentionSvc.Start(bgCtx)

	exportController := controller.NewInstructorExportController(exportSvc, jobQueue, storage, logger)

	// Portfolio-scope routes (owner only).
	portfolioRouter := v1.Group("/instructor/reports")
	portfolioRouter.Use(middleware.EnsureCorrelationID, middleware.KratosAuthenticated)
	portfolioRouter.Post("/exports", exportController.RequestExport)
	portfolioRouter.Get("/exports/:report_id", exportController.GetExportStatus)
	portfolioRouter.Get("/exports/:report_id/download", exportController.DownloadReport)
	portfolioRouter.Delete("/exports/:report_id", exportController.DeleteReport)
	portfolioRouter.Get("/history", exportController.GetHistory)
	portfolioRouter.Get("/audit", exportController.GetAuditLog)
	portfolioRouter.Post("/schedules", exportController.CreateSchedule)
	portfolioRouter.Get("/schedules", exportController.ListSchedules)
	portfolioRouter.Get("/schedules/:schedule_id", exportController.GetSchedule)
	portfolioRouter.Patch("/schedules/:schedule_id", exportController.UpdateSchedule)
	portfolioRouter.Delete("/schedules/:schedule_id", exportController.DeleteSchedule)

	// Per-quiz-scope routes (collaborators allowed via VerifyQuizAnalyticsAccess).
	quizRouter := v1.Group(fmt.Sprintf("/quizzes/:%s/reports", constants.QuizId))
	quizRouter.Use(middleware.EnsureCorrelationID, middleware.KratosAuthenticated, middleware.LoadQuizAnalyticsContext, middleware.VerifyQuizAnalyticsAccess)
	quizRouter.Post("/exports", exportController.RequestQuizExport)
	quizRouter.Get("/history", exportController.GetQuizHistory)
	quizRouter.Delete("/exports/:report_id", exportController.DeleteQuizReport)

	return nil
}
