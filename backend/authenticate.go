package main

import (
	"net/http"
	"github.com/gin-gonic/gin"
)

func authenticate(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"message": "This system does not support authentication"})
}