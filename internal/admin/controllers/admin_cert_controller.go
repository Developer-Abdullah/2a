package controllers

import (
	"net/http"
	"github.com/gin-gonic/gin"
)

type AdminCertController struct{}

func NewAdminCertController() *AdminCertController { return &AdminCertController{} }

func (ctrl *AdminCertController) List(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"success": true, "data": []interface{}{}})
}