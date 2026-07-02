package controllers

import (
	"net/http"
	"github.com/gin-gonic/gin"
)

type AdminStatsController struct{}

func NewAdminStatsController() *AdminStatsController { return &AdminStatsController{} }

func (ctrl *AdminStatsController) Overview(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"success": true, "data": map[string]interface{}{}})
}