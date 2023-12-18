package controllers

import (
	"fmt"
	"sfilter/user/models"

	"github.com/gin-gonic/gin"
)

func getRegisterParams(c *gin.Context) (*models.RegisterForm, error) {
	var creds models.RegisterForm
	err := c.ShouldBind(&creds)
	if err != nil {
		return nil, fmt.Errorf("wrong params")
	}

	return &creds, nil
}
