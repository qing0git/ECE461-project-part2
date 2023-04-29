package main

import (
	"github.com/gin-gonic/gin"
)

type PackageQuery struct {
	Name string `json:"name"`
	Version string `json:"version,omitempty"`
}

type UpdatePackageRequest struct {
	Metadata struct {
		Name string `json:"Name"`
		Version string `json:"Version"`
		ID string `json:"ID"`
	} `json:"metadata"`
	Data struct {
		Content string `json:"Content,omitempty"`
		URL string `json:"URL,omitempty"`
		JSProgram string `json:"JSProgram"`
	} `json:"data"`
}

type CreatePackageRequest struct {
	Content string `json:"Content,omitempty"`
	JSProgram string `json:"JSProgram"`
	URL string `json:"URL,omitempty"`
}

type PackageByRegex struct {
	RegEx string `json:"RegEx"`
}

func main() {
	// Initialize Gin
	router := gin.Default()
	router.POST("/packages", searchPackages)
	router.DELETE("/reset", resetRegistry)
	router.GET("/package/:id", getPackageByID)
	router.PUT("/package/:id", updatePackageByID)
	router.DELETE("/package/:id", deletePackageByID)
	router.POST("/package", createPackage)
	router.GET("/package/:id/rate", ratePackage)
	router.PUT("/authenticate", authenticate)
	router.GET("/package/byName/:name", getPackageByName)
	router.DELETE("/package/byName/:name", deletePackageByName)
	router.POST("/package/byRegEx", searchByRegEx)
	// Start the server
	router.Run(":8080")
}
