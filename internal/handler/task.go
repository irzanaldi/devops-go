package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/irzan-aldi/task-manager/internal/model"
	"github.com/irzan-aldi/task-manager/internal/service"
	"github.com/irzan-aldi/task-manager/pkg/response"
)

type TaskHandler struct {
	taskService *service.TaskService
}

func NewTaskHandler(taskService *service.TaskService) *TaskHandler {
	return &TaskHandler{taskService: taskService}
}

func (h *TaskHandler) Create(c *gin.Context) {
	userID := c.GetUint("userID")
	boardID, err := strconv.ParseUint(c.Param("boardId"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid board ID")
		return
	}

	var req model.CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	task, err := h.taskService.Create(uint(boardID), userID, req)
	if err != nil {
		response.Error(c, http.StatusForbidden, err.Error())
		return
	}

	response.Success(c, http.StatusCreated, "Task created", task)
}

func (h *TaskHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid task ID")
		return
	}

	task, err := h.taskService.GetByID(uint(id))
	if err != nil {
		response.Error(c, http.StatusNotFound, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Task detail", task)
}

func (h *TaskHandler) GetByBoardID(c *gin.Context) {
	boardID, err := strconv.ParseUint(c.Param("boardId"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid board ID")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	tasks, total, err := h.taskService.GetByBoardID(uint(boardID), page, limit)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to fetch tasks")
		return
	}

	response.Success(c, http.StatusOK, "Board tasks", gin.H{
		"tasks": tasks,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

func (h *TaskHandler) GetMyTasks(c *gin.Context) {
	userID := c.GetUint("userID")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	tasks, total, err := h.taskService.GetMyTasks(userID, page, limit)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to fetch tasks")
		return
	}

	response.Success(c, http.StatusOK, "My tasks", gin.H{
		"tasks": tasks,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

func (h *TaskHandler) Update(c *gin.Context) {
	userID := c.GetUint("userID")
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid task ID")
		return
	}

	var req model.UpdateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	task, err := h.taskService.Update(uint(id), userID, req)
	if err != nil {
		response.Error(c, http.StatusForbidden, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Task updated", task)
}

func (h *TaskHandler) Delete(c *gin.Context) {
	userID := c.GetUint("userID")
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid task ID")
		return
	}

	if err := h.taskService.Delete(uint(id), userID); err != nil {
		response.Error(c, http.StatusForbidden, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Task deleted", nil)
}
