package controller

import (
	"errors"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/th1enq/ViettelSMS_ServerService/internal/delivery/http/presenter"
	"github.com/th1enq/ViettelSMS_ServerService/internal/domain/dto"
	domain "github.com/th1enq/ViettelSMS_ServerService/internal/domain/errors"
	server_usecase "github.com/th1enq/ViettelSMS_ServerService/internal/usecase/server"
	"go.uber.org/zap"
)

type Controller struct {
	usecase   server_usecase.UseCase
	logger    *zap.Logger
	presenter presenter.Presenter
}

func NewController(
	usecase server_usecase.UseCase,
	logger *zap.Logger,
	presenter presenter.Presenter,
) *Controller {
	return &Controller{
		usecase:   usecase,
		logger:    logger,
		presenter: presenter,
	}
}

// CreateServer godoc
// @Summary Create a new server
// @Description Create a new server with the provided information
// @Tags server
// @Accept json
// @Produce json
// @Param server body response.CreateServerParams true "Server information"
// @Success 201 {object} response.APIResponse{data=response.ServerResponse}
// @Failure 400 {object} response.APIResponse
// @Failure 409 {object} response.APIResponse
// @Failure 500 {object} response.APIResponse
// @Security BearerAuth
// @Router /api/v1/server [post]
func (s *Controller) Create(c *gin.Context) {
	s.logger.Info("Create server request received")

	var req dto.CreateServerParams
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		s.logger.Warn("Failed to bind request body", zap.Error(err))
		s.presenter.InvalidRequest(c, "Invalid request body", err)
	}

	if err := s.usecase.CreateServer(c.Request.Context(), req); err != nil {
		if errors.Is(err, domain.ErrServerExist) {
			s.logger.Warn("Server already exists", zap.String("server_id", req.ServerID), zap.String("server_name", req.ServerName))
			s.presenter.Conflict(c, "Server already exists", err)
		} else {
			s.logger.Error("Failed to create server", zap.Error(err))
			s.presenter.InternalError(c, "Failed to create server", err)
		}
		return
	}

	s.logger.Info("Server created successfully", zap.Any("server", req))
	s.presenter.Created(c, "Server created successfully")
}

// DeleteServer godoc
// @Summary Delete server
// @Description Delete a server by ID
// @Tags server
// @Accept json
// @Produce json
// @Param id path int true "Server ID"
// @Success 200 {object} response.APIResponse
// @Failure 404 {object} response.APIResponse
// @Failure 409 {object} response.APIResponse
// @Failure 500 {object} response.APIResponse
// @Security BearerAuth
// @Router /api/v1/server/{id} [delete]
func (s *Controller) Delete(c *gin.Context) {
	s.logger.Info("Delete server request received")
	serverID := c.Param("server_id")

	if err := s.usecase.DeleteServer(c.Request.Context(), serverID); err != nil {
		if errors.Is(err, domain.ErrServerNotFound) {
			s.logger.Warn("Server not found", zap.String("server_id", serverID))
			s.presenter.NotFound(c, "Server not found", err)
		} else {
			s.logger.Error("Failed to delete server", zap.Error(err))
			s.presenter.InternalError(c, "Failed to delete server", err)
		}
		return
	}

	s.logger.Info("Server deleted successfully", zap.String("server_id", serverID))
	s.presenter.Deleted(c, "Server deleted successfully")
}

// UpdateServer godoc
// @Summary Update server
// @Description Update server information
// @Tags server
// @Accept json
// @Produce json
// @Param id path int true "Server ID"
// @Param updateInfo body dto.UpdateServerParams true "Server update information"
// @Success 200 {object} response.APIResponse{data=response.ServerResponse}
// @Failure 400 {object} response.APIResponse
// @Failure 404 {object} response.APIResponse
// @Failure 409 {object} response.APIResponse
// @Failure 500 {object} response.APIResponse
// @Security BearerAuth
// @Router /api/v1/server/{id} [put]
func (s *Controller) Update(c *gin.Context) {
	s.logger.Info("Update server request received")

	serverID := c.Param("server_id")
	var req dto.UpdateServerParams
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		s.logger.Warn("Failed to bind request body", zap.Error(err))
		s.presenter.InvalidRequest(c, "Invalid request body", err)
		return
	}

	if err := s.usecase.UpdateServer(c.Request.Context(), serverID, req); err != nil {
		if errors.Is(err, domain.ErrServerNotFound) {
			s.logger.Warn("Server not found", zap.String("server_id", serverID))
			s.presenter.NotFound(c, "Server not found", err)
		} else if errors.Is(err, domain.ErrServerExist) {
			s.logger.Warn("Server already exists", zap.Any("request update", req))
			s.presenter.Conflict(c, "Server already exists", err)
		} else {
			s.logger.Error("Failed to update server", zap.Error(err))
			s.presenter.InternalError(c, "Failed to update server", err)
		}
		return
	}

	s.logger.Info("Server updated successfully", zap.String("server_id", serverID))
	s.presenter.Updated(c, "Server updated successfully")
}

// ViewServers godoc
// @Summary View servers
// @Description Get list of servers with optional filters and pagination
// @Tags server
// @Accept json
// @Produce json
// @Param server_name query string false "Filter by server name"
// @Param status query string false "Filter by status"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(10)
// @Param sort query string false "Sort field" default(server_name)
// @Param order query string false "Sort order" default(asc)
// @Success 200 {object} response.APIResponse{data=dto.ServerListResponse}
// @Failure 400 {object} response.APIResponse
// @Failure 500 {object} response.APIResponse
// @Security BearerAuth
// @Router /api/v1/server [get]
func (s *Controller) View(c *gin.Context) {
	s.logger.Info("View server request received")

	var (
		filter     dto.ServerFilterOptions
		pagination dto.ServerPaginationOptions
	)

	if err := c.ShouldBindBodyWithJSON(&filter); err != nil {
		s.logger.Warn("Failed to bind filter options", zap.Error(err))
		s.presenter.InvalidRequest(c, "Invalid filter options", err)
		return
	}

	if err := c.ShouldBindBodyWithJSON(&pagination); err != nil {
		s.logger.Warn("Failed to bind pagination options", zap.Error(err))
		s.presenter.InvalidRequest(c, "Invalid pagination options", err)
		return
	}

	server, total, err := s.usecase.ViewServer(c.Request.Context(), filter, pagination)
	if err != nil {
		s.logger.Error("Failed to view servers", zap.Error(err))
		s.presenter.InternalError(c, "Failed to view servers", err)
		return
	}
	s.logger.Info("Servers retrieved successfully", zap.Int("total", total))
	s.presenter.Retrived(c, "Servers retrieved successfully", server)
}

// ImportServers godoc
// @Summary Import servers from Excel file
// @Description Import multiple servers from an Excel file
// @Tags server
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "Excel file"
// @Success 200 {object} response.APIResponse{data=dto.ImportResult}
// @Failure 400 {object} response.APIResponse
// @Failure 500 {object} response.APIResponse
// @Security BearerAuth
// @Router /api/v1/server/import [post]
func (s *Controller) Import(c *gin.Context) {
	s.logger.Info("Import server request received")

	file, err := c.FormFile("file")
	if err != nil {
		s.logger.Warn("Failed to get file from request", zap.Error(err))
		s.presenter.InvalidRequest(c, "Invalid file", err)
		return
	}

	filePath := fmt.Sprintf("/tmp/%s_%s", uuid.New().String(), file.Filename)
	if err := c.SaveUploadedFile(file, filePath); err != nil {
		s.logger.Error("Failed to save uploaded file", zap.String("file_path", filePath), zap.Error(err))
		s.presenter.InternalError(c, "Failed to save uploaded file", err)
		return
	}

	result, err := s.usecase.ImportServer(c.Request.Context(), filePath)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidFile) {
			s.logger.Warn("Invalid file format", zap.Error(err))
			s.presenter.InvalidRequest(c, "Invalid file format", err)
		} else {
			s.logger.Error("Failed to import server", zap.Error(err))
			s.presenter.InternalError(c, "Failed to import server", err)
		}
		return
	}

	s.logger.Info("Server imported successfully", zap.String("file_name", file.Filename))
	s.presenter.Imported(c, "Server imported successfully", result)
}

func (s *Controller) Export(c *gin.Context) {
	s.logger.Info("Export server request received")

	var (
		filter     dto.ServerFilterOptions
		pagination dto.ServerPaginationOptions
	)

	if err := c.ShouldBindBodyWithJSON(&filter); err != nil {
		s.logger.Warn("Failed to bind filter options", zap.Error(err))
		s.presenter.InvalidRequest(c, "Invalid filter options", err)
		return
	}

	if err := c.ShouldBindBodyWithJSON(&pagination); err != nil {
		s.logger.Warn("Failed to bind pagination options", zap.Error(err))
		s.presenter.InvalidRequest(c, "Invalid pagination options", err)
		return
	}

	filePath, err := s.usecase.ExportServer(c.Request.Context(), filter, pagination)
	if err != nil {
		s.logger.Error("Failed to export servers", zap.Error(err))
		s.presenter.InternalError(c, "Failed to export servers", err)
		return
	}

	s.logger.Info("Servers exported successfully", zap.String("file_path", filePath))
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Content-Disposition", "attachment; filename=servers.xlsx")
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.File(filePath)
}
