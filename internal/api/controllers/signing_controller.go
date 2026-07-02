package controllers

import (
	"context"
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"

	"platform/internal/domain"
	"platform/internal/tenant"
	"platform/pkg/response"
)

type SigningJobRepository interface {
	GetJobStatus(ctx context.Context, tenantID string, jobID string) (*domain.SigningJobStatusDTO, error)
}

type SigningController struct {
	repo SigningJobRepository
}

func NewSigningController(repo SigningJobRepository) *SigningController {
	return &SigningController{repo: repo}
}

func (ctrl *SigningController) StreamProgress(c *gin.Context) {
	currentTenant := tenant.GetFromContext(c)
	if currentTenant == nil {
		response.Fail(c, http.StatusBadRequest, "tenant_missing", "tenant context missing")
		return
	}

	status, err := ctrl.repo.GetJobStatus(c.Request.Context(), currentTenant.ID.String(), c.Param("job_id"))
	if err != nil {
		if err == sql.ErrNoRows {
			response.Fail(c, http.StatusNotFound, "job_not_found", "signing job not found")
			return
		}
		response.Fail(c, http.StatusInternalServerError, "job_status_failed", "failed to get signing job status")
		return
	}

	response.Success(c, status)
}
