package v1

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/doug-martin/goqu/v9"
	fiber "github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/randhir3-cloud/GK-Circle-v2/api/config"
	"github.com/randhir3-cloud/GK-Circle-v2/api/constants"
	"github.com/randhir3-cloud/GK-Circle-v2/api/models"
	"github.com/randhir3-cloud/GK-Circle-v2/api/pkg/structs"
	"github.com/randhir3-cloud/GK-Circle-v2/api/utils"
)

type CourseNodeController struct {
	courseNodeModel *models.CourseNodeModel
	appConfig       *config.AppConfig
	logger          *zap.Logger
}

func InitCourseNodeController(db *goqu.Database, logger *zap.Logger, appConfig *config.AppConfig) (*CourseNodeController, error) {
	return &CourseNodeController{
		courseNodeModel: models.InitCourseNodeModel(db),
		appConfig:       appConfig,
		logger:          logger,
	}, nil
}

func (ctrl *CourseNodeController) CreateNode(c *fiber.Ctx) error {
	if _, _, err := authorizeCourseAdmin(c, ctrl.appConfig); err != nil {
		return mapCourseAuthError(c, ctrl.logger, err)
	}

	courseID, err := parseCourseIDParam(c)
	if err != nil {
		return utils.JSONFail(c, http.StatusBadRequest, constants.ErrCourseInvalidID)
	}

	var req structs.ReqCreateAdminCourseNode
	if err := json.Unmarshal(c.Body(), &req); err != nil {
		return utils.JSONFail(c, http.StatusBadRequest, err.Error())
	}

	params, err := validateCreateAdminCourseNode(courseID, req)
	if err != nil {
		return utils.JSONFail(c, http.StatusBadRequest, err.Error())
	}

	node, err := ctrl.courseNodeModel.CreateCourseNode(params)
	if err != nil {
		return ctrl.mapCreateError(c, err)
	}
	return utils.JSONSuccess(c, http.StatusCreated, toAdminCourseNodeResponse(node))
}

func (ctrl *CourseNodeController) ListRoots(c *fiber.Ctx) error {
	if _, _, err := authorizeCourseAdmin(c, ctrl.appConfig); err != nil {
		return mapCourseAuthError(c, ctrl.logger, err)
	}

	courseID, err := parseCourseIDParam(c)
	if err != nil {
		return utils.JSONFail(c, http.StatusBadRequest, constants.ErrCourseInvalidID)
	}

	nodes, err := ctrl.courseNodeModel.ListRootNodes(courseID)
	if err != nil {
		if errors.Is(err, models.ErrCourseNodeCourseNotFound) {
			return utils.JSONFail(c, http.StatusNotFound, constants.ErrCourseNotFound)
		}
		ctrl.logger.Error("error while listing course node roots", zap.Error(err))
		return utils.JSONError(c, http.StatusInternalServerError, constants.ErrListCourseNodes)
	}
	response := make([]structs.AdminCourseNodeResponse, 0, len(nodes))
	for _, node := range nodes {
		response = append(response, toAdminCourseNodeResponse(node))
	}
	return utils.JSONSuccess(c, http.StatusOK, response)
}

func (ctrl *CourseNodeController) GetTree(c *fiber.Ctx) error {
	if _, _, err := authorizeCourseAdmin(c, ctrl.appConfig); err != nil {
		return mapCourseAuthError(c, ctrl.logger, err)
	}

	courseID, err := parseCourseIDParam(c)
	if err != nil {
		return utils.JSONFail(c, http.StatusBadRequest, constants.ErrCourseInvalidID)
	}

	hierarchy, err := ctrl.courseNodeModel.GetHierarchy(courseID)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrCourseNodeCourseNotFound):
			return utils.JSONFail(c, http.StatusNotFound, constants.ErrCourseNotFound)
		case errors.Is(err, models.ErrCourseNodeHierarchyIntegrity):
			ctrl.logger.Error("course hierarchy integrity failure", zap.Error(err), zap.String("course_id", courseID.String()))
			return utils.JSONError(c, http.StatusInternalServerError, constants.ErrGetCourseHierarchy)
		default:
			ctrl.logger.Error("error while getting course hierarchy", zap.Error(err))
			return utils.JSONError(c, http.StatusInternalServerError, constants.ErrGetCourseHierarchy)
		}
	}
	return utils.JSONSuccess(c, http.StatusOK, toAdminCourseHierarchyResponse(hierarchy))
}

func (ctrl *CourseNodeController) GetByID(c *fiber.Ctx) error {
	if _, _, err := authorizeCourseAdmin(c, ctrl.appConfig); err != nil {
		return mapCourseAuthError(c, ctrl.logger, err)
	}

	courseID, err := parseCourseIDParam(c)
	if err != nil {
		return utils.JSONFail(c, http.StatusBadRequest, constants.ErrCourseInvalidID)
	}
	nodeID, err := parseNodeIDParam(c)
	if err != nil {
		return utils.JSONFail(c, http.StatusBadRequest, constants.ErrCourseNodeInvalidID)
	}

	node, err := ctrl.courseNodeModel.GetCourseNodeByID(courseID, nodeID)
	if err != nil {
		if errors.Is(err, models.ErrCourseNodeNotFound) {
			return utils.JSONFail(c, http.StatusNotFound, constants.ErrCourseNodeNotFound)
		}
		ctrl.logger.Error("error while getting course node", zap.Error(err))
		return utils.JSONError(c, http.StatusInternalServerError, constants.ErrGetCourseNode)
	}
	return utils.JSONSuccess(c, http.StatusOK, toAdminCourseNodeResponse(node))
}

func (ctrl *CourseNodeController) ListChildren(c *fiber.Ctx) error {
	if _, _, err := authorizeCourseAdmin(c, ctrl.appConfig); err != nil {
		return mapCourseAuthError(c, ctrl.logger, err)
	}

	courseID, err := parseCourseIDParam(c)
	if err != nil {
		return utils.JSONFail(c, http.StatusBadRequest, constants.ErrCourseInvalidID)
	}
	nodeID, err := parseNodeIDParam(c)
	if err != nil {
		return utils.JSONFail(c, http.StatusBadRequest, constants.ErrCourseNodeInvalidID)
	}

	nodes, err := ctrl.courseNodeModel.ListChildren(courseID, nodeID)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrCourseNodeCourseNotFound):
			return utils.JSONFail(c, http.StatusNotFound, constants.ErrCourseNotFound)
		case errors.Is(err, models.ErrCourseNodeParentNotFound),
			errors.Is(err, models.ErrCourseNodeCrossCourseParent):
			return utils.JSONFail(c, http.StatusNotFound, constants.ErrCourseNodeNotFound)
		default:
			ctrl.logger.Error("error while listing course node children", zap.Error(err))
			return utils.JSONError(c, http.StatusInternalServerError, constants.ErrListCourseNodes)
		}
	}
	response := make([]structs.AdminCourseNodeResponse, 0, len(nodes))
	for _, node := range nodes {
		response = append(response, toAdminCourseNodeResponse(node))
	}
	return utils.JSONSuccess(c, http.StatusOK, response)
}

func (ctrl *CourseNodeController) MoveNode(c *fiber.Ctx) error {
	if _, _, err := authorizeCourseAdmin(c, ctrl.appConfig); err != nil {
		return mapCourseAuthError(c, ctrl.logger, err)
	}

	courseID, err := parseCourseIDParam(c)
	if err != nil {
		return utils.JSONFail(c, http.StatusBadRequest, constants.ErrCourseInvalidID)
	}
	nodeID, err := parseNodeIDParam(c)
	if err != nil {
		return utils.JSONFail(c, http.StatusBadRequest, constants.ErrCourseNodeInvalidID)
	}

	var req structs.ReqMoveAdminCourseNode
	if err := json.Unmarshal(c.Body(), &req); err != nil {
		return utils.JSONFail(c, http.StatusBadRequest, err.Error())
	}

	position, err := parseRequiredPosition(req.Position)
	if err != nil {
		return utils.JSONFail(c, http.StatusBadRequest, err.Error())
	}
	newParentID, err := parseOptionalParentID(req.NewParentID, constants.ErrCourseNodeParentInvalid)
	if err != nil {
		return utils.JSONFail(c, http.StatusBadRequest, err.Error())
	}

	if err := ctrl.courseNodeModel.MoveNode(courseID, nodeID, newParentID, position); err != nil {
		return ctrl.mapMoveError(c, err)
	}
	return utils.JSONSuccess(c, http.StatusOK, "success")
}

func (ctrl *CourseNodeController) ReorderChildren(c *fiber.Ctx) error {
	if _, _, err := authorizeCourseAdmin(c, ctrl.appConfig); err != nil {
		return mapCourseAuthError(c, ctrl.logger, err)
	}

	courseID, err := parseCourseIDParam(c)
	if err != nil {
		return utils.JSONFail(c, http.StatusBadRequest, constants.ErrCourseInvalidID)
	}

	var req structs.ReqReorderAdminCourseNodes
	if err := json.Unmarshal(c.Body(), &req); err != nil {
		return utils.JSONFail(c, http.StatusBadRequest, err.Error())
	}

	parentID, err := parseOptionalParentID(req.ParentID, constants.ErrCourseNodeParentInvalid)
	if err != nil {
		return utils.JSONFail(c, http.StatusBadRequest, err.Error())
	}
	orderedIDs, err := parseOrderedNodeIDs(req.OrderedNodeIDs)
	if err != nil {
		return utils.JSONFail(c, http.StatusBadRequest, err.Error())
	}

	if err := ctrl.courseNodeModel.ReorderChildren(courseID, parentID, orderedIDs); err != nil {
		return ctrl.mapReorderError(c, err)
	}
	return utils.JSONSuccess(c, http.StatusOK, "success")
}

func (ctrl *CourseNodeController) DeleteSubtree(c *fiber.Ctx) error {
	if _, _, err := authorizeCourseAdmin(c, ctrl.appConfig); err != nil {
		return mapCourseAuthError(c, ctrl.logger, err)
	}

	courseID, err := parseCourseIDParam(c)
	if err != nil {
		return utils.JSONFail(c, http.StatusBadRequest, constants.ErrCourseInvalidID)
	}
	nodeID, err := parseNodeIDParam(c)
	if err != nil {
		return utils.JSONFail(c, http.StatusBadRequest, constants.ErrCourseNodeInvalidID)
	}

	if err := ctrl.courseNodeModel.DeleteSubtree(courseID, nodeID); err != nil {
		return ctrl.mapDeleteError(c, err)
	}
	return utils.JSONSuccess(c, http.StatusOK, "success")
}

func (ctrl *CourseNodeController) mapCreateError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, models.ErrCourseNodeCourseNotFound):
		return utils.JSONFail(c, http.StatusNotFound, constants.ErrCourseNotFound)
	case errors.Is(err, models.ErrCourseNodeTitleRequired):
		return utils.JSONFail(c, http.StatusBadRequest, constants.ErrCourseNodeTitleInvalid)
	case errors.Is(err, models.ErrCourseNodeTypeInvalid):
		return utils.JSONFail(c, http.StatusBadRequest, constants.ErrCourseNodeTypeInvalid)
	case errors.Is(err, models.ErrCourseNodePositionInvalid):
		return utils.JSONFail(c, http.StatusBadRequest, constants.ErrCourseNodePositionInvalid)
	case errors.Is(err, models.ErrCourseNodeParentNotFound),
		errors.Is(err, models.ErrCourseNodeCrossCourseParent):
		return utils.JSONFail(c, http.StatusBadRequest, constants.ErrCourseNodeParentInvalid)
	case errors.Is(err, models.ErrCourseNodePositionConflict):
		return utils.JSONFail(c, http.StatusConflict, constants.ErrCourseNodePositionConflict)
	default:
		ctrl.logger.Error("error while creating course node", zap.Error(err))
		return utils.JSONError(c, http.StatusInternalServerError, constants.ErrCreateCourseNode)
	}
}

func (ctrl *CourseNodeController) mapMoveError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, models.ErrCourseNodePositionInvalid),
		errors.Is(err, models.ErrCourseNodeInvalidMove),
		errors.Is(err, models.ErrCourseNodeCycle):
		return utils.JSONFail(c, http.StatusBadRequest, constants.ErrCourseNodeMoveInvalid)
	case errors.Is(err, models.ErrCourseNodeCourseNotFound):
		return utils.JSONFail(c, http.StatusNotFound, constants.ErrCourseNotFound)
	case errors.Is(err, models.ErrCourseNodeNotFound),
		errors.Is(err, models.ErrCourseNodeParentNotFound),
		errors.Is(err, models.ErrCourseNodeCrossCourseParent):
		return utils.JSONFail(c, http.StatusNotFound, constants.ErrCourseNodeNotFound)
	case errors.Is(err, models.ErrCourseNodePositionConflict),
		errors.Is(err, models.ErrCourseNodeSubtreeConflict):
		return utils.JSONFail(c, http.StatusConflict, constants.ErrCourseNodeMoveConflict)
	default:
		ctrl.logger.Error("error while moving course node", zap.Error(err))
		return utils.JSONError(c, http.StatusInternalServerError, constants.ErrMoveCourseNode)
	}
}

func (ctrl *CourseNodeController) mapReorderError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, models.ErrCourseNodeReorderDuplicate),
		errors.Is(err, models.ErrCourseNodeReorderMismatch),
		errors.Is(err, models.ErrCourseNodeNotFound):
		return utils.JSONFail(c, http.StatusBadRequest, constants.ErrCourseNodeReorderMismatch)
	case errors.Is(err, models.ErrCourseNodeCourseNotFound):
		return utils.JSONFail(c, http.StatusNotFound, constants.ErrCourseNotFound)
	case errors.Is(err, models.ErrCourseNodeParentNotFound),
		errors.Is(err, models.ErrCourseNodeCrossCourseParent):
		return utils.JSONFail(c, http.StatusNotFound, constants.ErrCourseNodeNotFound)
	case errors.Is(err, models.ErrCourseNodeReorderConflict):
		return utils.JSONFail(c, http.StatusConflict, constants.ErrCourseNodeReorderConflict)
	default:
		ctrl.logger.Error("error while reordering course nodes", zap.Error(err))
		return utils.JSONError(c, http.StatusInternalServerError, constants.ErrReorderCourseNodes)
	}
}

func (ctrl *CourseNodeController) mapDeleteError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, models.ErrCourseNodeCourseNotFound):
		return utils.JSONFail(c, http.StatusNotFound, constants.ErrCourseNotFound)
	case errors.Is(err, models.ErrCourseNodeNotFound):
		return utils.JSONFail(c, http.StatusNotFound, constants.ErrCourseNodeNotFound)
	case errors.Is(err, models.ErrCourseNodeDeleteConflict):
		return utils.JSONFail(c, http.StatusConflict, constants.ErrCourseNodeDeleteConflict)
	default:
		ctrl.logger.Error("error while deleting course node", zap.Error(err))
		return utils.JSONError(c, http.StatusInternalServerError, constants.ErrDeleteCourseNode)
	}
}

func parseNodeIDParam(c *fiber.Ctx) (uuid.UUID, error) {
	return uuid.Parse(strings.TrimSpace(c.Params(constants.NodeId)))
}

func validateCreateAdminCourseNode(courseID uuid.UUID, req structs.ReqCreateAdminCourseNode) (models.CreateCourseNodeParams, error) {
	params := models.CreateCourseNodeParams{CourseID: courseID}

	title := strings.TrimSpace(req.Title)
	if title == "" {
		return params, errors.New(constants.ErrCourseNodeTitleInvalid)
	}
	params.Title = title

	nodeType := models.CourseNodeType(strings.TrimSpace(req.NodeType))
	switch nodeType {
	case models.CourseNodeTypeSection, models.CourseNodeTypeSubject, models.CourseNodeTypeTopic:
		params.NodeType = nodeType
	default:
		return params, errors.New(constants.ErrCourseNodeTypeInvalid)
	}

	position, err := parseRequiredPosition(req.Position)
	if err != nil {
		return params, err
	}
	params.Position = position

	parentID, err := parseCreateParentID(req.ParentID)
	if err != nil {
		return params, err
	}
	params.ParentID = parentID
	return params, nil
}

func parseRequiredPosition(field structs.OptionalInteger) (int, error) {
	if !field.Present {
		return 0, errors.New(constants.ErrCourseNodePositionRequired)
	}
	if field.Null {
		return 0, errors.New(constants.ErrCourseNodePositionRequired)
	}
	if field.Value < 0 {
		return 0, errors.New(constants.ErrCourseNodePositionInvalid)
	}
	return field.Value, nil
}

func parseCreateParentID(field structs.OptionalString) (uuid.NullUUID, error) {
	return parseOptionalParentID(field, constants.ErrCourseNodeParentInvalid)
}

func parseOptionalParentID(field structs.OptionalString, invalidMessage string) (uuid.NullUUID, error) {
	if !field.Present || field.Null {
		return uuid.NullUUID{}, nil
	}
	raw := strings.TrimSpace(field.Value)
	if raw == "" {
		return uuid.NullUUID{}, errors.New(invalidMessage)
	}
	parentID, err := uuid.Parse(raw)
	if err != nil || parentID == uuid.Nil {
		return uuid.NullUUID{}, errors.New(invalidMessage)
	}
	return uuid.NullUUID{UUID: parentID, Valid: true}, nil
}

func parseOrderedNodeIDs(raw []string) ([]uuid.UUID, error) {
	if raw == nil || len(raw) == 0 {
		return nil, errors.New(constants.ErrCourseNodeReorderInvalid)
	}
	parsed := make([]uuid.UUID, 0, len(raw))
	seen := make(map[uuid.UUID]struct{}, len(raw))
	for _, entry := range raw {
		trimmed := strings.TrimSpace(entry)
		if trimmed == "" {
			return nil, errors.New(constants.ErrCourseNodeReorderInvalid)
		}
		id, err := uuid.Parse(trimmed)
		if err != nil || id == uuid.Nil {
			return nil, errors.New(constants.ErrCourseNodeReorderInvalid)
		}
		if _, exists := seen[id]; exists {
			return nil, errors.New(constants.ErrCourseNodeReorderInvalid)
		}
		seen[id] = struct{}{}
		parsed = append(parsed, id)
	}
	return parsed, nil
}

func toAdminCourseNodeResponse(node models.CourseNode) structs.AdminCourseNodeResponse {
	var parentID *uuid.UUID
	if node.ParentID.Valid {
		id := node.ParentID.UUID
		parentID = &id
	}
	return structs.AdminCourseNodeResponse{
		ID:        node.ID,
		CourseID:  node.CourseID,
		ParentID:  parentID,
		Title:     node.Title,
		NodeType:  string(node.NodeType),
		Status:    string(node.Status),
		Position:  node.Position,
		CreatedAt: node.CreatedAt.UTC().Format(time.RFC3339Nano),
		UpdatedAt: node.UpdatedAt.UTC().Format(time.RFC3339Nano),
	}
}

func toAdminCourseHierarchyResponse(hierarchy models.CourseHierarchy) structs.AdminCourseHierarchyResponse {
	roots := make([]structs.AdminCourseNodeHierarchyNode, 0, len(hierarchy.Roots))
	for _, root := range hierarchy.Roots {
		roots = append(roots, toAdminCourseNodeHierarchyNode(root))
	}
	return structs.AdminCourseHierarchyResponse{
		CourseID: hierarchy.CourseID,
		Roots:    roots,
	}
}

func toAdminCourseNodeHierarchyNode(node models.CourseNodeHierarchy) structs.AdminCourseNodeHierarchyNode {
	children := make([]structs.AdminCourseNodeHierarchyNode, 0, len(node.Children))
	for _, child := range node.Children {
		children = append(children, toAdminCourseNodeHierarchyNode(child))
	}
	return structs.AdminCourseNodeHierarchyNode{
		Node:     toAdminCourseNodeResponse(node.Node),
		Children: children,
	}
}
