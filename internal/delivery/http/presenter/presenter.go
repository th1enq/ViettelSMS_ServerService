package presenter

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"github.com/th1enq/ViettelSMS_ServerService/internal/domain/response"
)

type (
	Presenter interface {
		// Error responses
		InvalidRequest(c *gin.Context, message string, err error)
		Conflict(c *gin.Context, message string, err error)
		InternalError(c *gin.Context, message string, err error)
		NotFound(c *gin.Context, message string, err error)

		// Success responses
		Created(c *gin.Context, message string, data interface{})
		Deleted(c *gin.Context, message string)
		Updated(c *gin.Context, message string, data interface{})
		Retrived(c *gin.Context, message string, data interface{})
		Imported(c *gin.Context, message string, data interface{})
	}

	presenter struct{}
)

func NewPresenter() Presenter {
	return &presenter{}
}

var PresenterWireSet = wire.NewSet(NewPresenter)

func (p *presenter) InvalidRequest(c *gin.Context, message string, err error) {
	c.JSON(http.StatusBadRequest, response.NewErrorResponse(
		response.CodeBadRequest,
		message,
		err.Error(),
	))
}

func (p *presenter) Conflict(c *gin.Context, message string, err error) {
	c.JSON(http.StatusConflict, response.NewErrorResponse(
		response.CodeConflict,
		message,
		err.Error(),
	))
}

func (p *presenter) InternalError(c *gin.Context, message string, err error) {
	c.JSON(http.StatusInternalServerError, response.NewErrorResponse(
		response.CodeInternalServerError,
		message,
		err.Error(),
	))
}

func (p *presenter) Created(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusCreated, response.NewSuccessResponse(
		response.CodeCreated,
		message,
		data,
	))
}

func (p *presenter) Deleted(c *gin.Context, message string) {
	c.JSON(http.StatusNoContent, response.NewSuccessResponse(
		response.CodeDeleted,
		message,
		nil,
	))
}

func (p *presenter) NotFound(c *gin.Context, message string, err error) {
	c.JSON(http.StatusNotFound, response.NewErrorResponse(
		response.CodeNotFound,
		message,
		err.Error(),
	))
}

func (p *presenter) Updated(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, response.NewSuccessResponse(
		response.CodeUpdated,
		message,
		data,
	))
}

func (p *presenter) Retrived(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, response.NewSuccessResponse(
		response.CodeUpdated,
		message,
		data,
	))
}

func (p *presenter) Imported(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, response.NewSuccessResponse(
		response.CodeSuccess,
		message,
		data,
	))
}
