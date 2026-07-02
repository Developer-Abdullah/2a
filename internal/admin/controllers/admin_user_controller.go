package controllers

import (
	"net/http"
	"github.com/gin-gonic/gin"
)

type AdminUserController struct{}

func NewAdminUserController() *AdminUserController { return &AdminUserController{} }

func (ctrl *AdminUserController) List(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"success": true, "data": []interface{}{}})
}