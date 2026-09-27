package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/irzan-aldi/task-manager/internal/model"
	"github.com/irzan-aldi/task-manager/internal/service"
	"github.com/irzan-aldi/task-manager/pkg/response"
)

type BoardHandler struct {
	boardService *service.BoardService
}

func NewBoardHandler(boardService *service.BoardService) *BoardHandler {
	return &BoardHandler{boardService: boardService}
}

func (h *BoardHandler) Create(c *gin.Context) {
	userID := c.GetUint("userID")
	var req model.CreateBoardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	board, err := h.boardService.Create(userID, req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, http.StatusCreated, "Board created", board)
}

func (h *BoardHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid board ID")
		return
	}

	board, err := h.boardService.GetByID(uint(id))
	if err != nil {
		response.Error(c, http.StatusNotFound, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Board detail", board)
}

func (h *BoardHandler) GetMyBoards(c *gin.Context) {
	userID := c.GetUint("userID")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	boards, total, err := h.boardService.GetByOwnerID(userID, page, limit)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to fetch boards")
		return
	}

	response.Success(c, http.StatusOK, "My boards", gin.H{
		"boards": boards,
		"total":  total,
		"page":   page,
		"limit":  limit,
	})
}

func (h *BoardHandler) Update(c *gin.Context) {
	userID := c.GetUint("userID")
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid board ID")
		return
	}

	var req model.UpdateBoardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	board, err := h.boardService.Update(uint(id), userID, req)
	if err != nil {
		response.Error(c, http.StatusForbidden, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Board updated", board)
}

func (h *BoardHandler) Delete(c *gin.Context) {
	userID := c.GetUint("userID")
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid board ID")
		return
	}

	if err := h.boardService.Delete(uint(id), userID); err != nil {
		response.Error(c, http.StatusForbidden, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Board deleted", nil)
}
