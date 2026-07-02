package controllers

import (
	"net/http"
	"github.com/gin-gonic/gin"
)

type AdminExportController struct{}

func NewAdminExportController() *AdminExportController { return &AdminExportController{} }

func (ctrl *AdminExportController) Export(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"success": true, "data": map[string]interface{}{}})
}