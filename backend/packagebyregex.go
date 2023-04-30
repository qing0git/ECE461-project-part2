package main

import (
	"context"
	"net/http"
	"regexp"
	"log"
	firebase "firebase.google.com/go"
	"github.com/gin-gonic/gin"
	"google.golang.org/api/iterator"
)

func searchByRegEx(c *gin.Context) {
	// Get the regex pattern from the request body
	var req PackageByRegex
	err := c.ShouldBindJSON(&req)
	if err != nil {
		log.Println("searchByRegEx error:", err)
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request body"})
		return
	}
	regexPattern := req.RegEx
	log.Println("searchByRegEx pattern:" + regexPattern)

	// Compile the regex pattern
	regex, err := regexp.Compile(regexPattern)
	if err != nil {
		log.Println("searchByRegEx error:", err)
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid regex pattern"})
		return
	}

	// Query the Firestore database for packages
	ctx := context.Background()
	conf := &firebase.Config{ProjectID: "ece461-pj-part2"}
	app, err := firebase.NewApp(ctx, conf)
	if err != nil {
		log.Println("searchByRegEx error:", err)
		c.JSON(http.StatusBadRequest, gin.H{"message": "Unexpected error"})
		return
	}
	client, err := app.Firestore(ctx)
	if err != nil {
		log.Println("searchByRegEx error:", err)
		c.JSON(http.StatusBadRequest, gin.H{"message": "Unexpected error"})
		return
	}
	defer client.Close()

	packages := client.Collection("repositories").Documents(ctx)
	if err != nil {
		log.Println("searchByRegEx error:", err)
		c.JSON(http.StatusBadRequest, gin.H{"message": "Unexpected error"})
		return
	}

	// Filter packages by regex pattern
	matchingPackages := []map[string]string{}
	for {
		doc, err := packages.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Println("searchByRegEx error:", err)
			c.JSON(http.StatusBadRequest, gin.H{"message": "Unexpected error"})
			return
		}
		packageData := doc.Data()
		if regex.MatchString(packageData["name"].(string)) || regex.MatchString(packageData["readme"].(string)) {
			log.Println("searchByRegEx: Version: " + packageData["version"].(string))
			log.Println("searchByRegEx: Name: " + packageData["name"].(string))
			matchingPackages = append(matchingPackages, map[string]string{
				"Version": packageData["version"].(string),
				"Name": packageData["name"].(string),
			})
		}
	}
	if len(matchingPackages) == 0 {
		log.Println("searchByRegEx error: No package found under this regex")
		c.JSON(http.StatusNotFound, gin.H{"message": "No package found under this regex"})
		return
	}
	c.JSON(http.StatusOK, matchingPackages)
}
