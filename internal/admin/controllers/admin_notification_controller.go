package controllers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"

	"platform/internal/notify"
	"platform/internal/tenant"
)

type AdminNotificationStore interface {
	Create(ctx context.Context, tenantID uuid.UUID, title, body string, targetAppID *string, nType string) (string, error)
}

type AdminNotificationController struct {
	store AdminNotificationStore
	queue *asynq.Client
}

func NewAdminNotificationController(store AdminNotificationStore, queue *asynq.Client) *AdminNotificationController {
	return &AdminNotificationController{store: store, queue: queue}
}

type sendNotificationRequest struct {
	Title       string  `json:"title" binding:"required"`
	Body        string  `json:"body" binding:"required"`
	TargetAppID *string `json:"target_app_id"`
	Type        string  `json:"type"`
}

// Send persists a notification and enqueues it for push delivery.
func (ctrl *AdminNotificationController) Send(c *gin.Context) {
	t := tenant.GetFromContext(c)
	if t == nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": gin.H{"message": "tenant context missing"}})
		return
	}

	var req sendNotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": gin.H{"message": "invalid request"}})
		return
	}

	id, err := ctrl.store.Create(c.Request.Context(), t.ID, req.Title, req.Body, req.TargetAppID, req.Type)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": gin.H{"message": err.Error()}})
		return
	}

	if task, err := notify.NewPushTask(notify.PushPayload{NotificationID: id, TenantID: t.ID.String(), Title: req.Title, Body: req.Body}); err == nil {
		_, _ = ctrl.queue.EnqueueContext(c.Request.Context(), task, asynq.Queue("notifications"))
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"id": id, "queued": true}})
}
