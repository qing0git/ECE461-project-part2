package main

import (
	"context"
	"net/http"
	"strconv"
	"log"
	firebase "firebase.google.com/go"
	"github.com/gin-gonic/gin"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	"github.com/Masterminds/semver/v3"
)

func searchPackages(c *gin.Context) {
	// Read the "offset" query parameter as a string
	offset := c.DefaultQuery("offset", "1")

	// Parse query
	var queries []PackageQuery
	err := c.ShouldBindJSON(&queries)
	if err != nil {
		log.Println("searchPackages error:", err)
		c.JSON(http.StatusBadRequest, gin.H{"message": "Missing or improperly formed PackageQuery"})
		return
	}
	log.Println("searchPackages: Name[0]" + queries[0].Name)
	log.Println("searchPackages: Version[0]" + queries[0].Version)
	// Check for the "all packages" query
	allPackages := false
	if len(queries) == 1 && queries[0].Name == "*" {
		allPackages = true
		log.Println("searchPackages: Output all packs")
	}

	// Set up Firestore client
	ctx := context.Background()
	sa := option.WithCredentialsFile("./ece461-pj-part2-b75cfa849e87.json")
	app, err := firebase.NewApp(ctx, nil, sa)
	if err != nil {
		log.Println("searchPackages error:", err)
		c.JSON(http.StatusBadRequest, gin.H{"message": "Unexpected error"})
		return
	}
	client, err := app.Firestore(ctx)
	if err != nil {
		log.Println("searchPackages error:", err)
		c.JSON(http.StatusBadRequest, gin.H{"message": "Unexpected error"})
		return
	}
	defer client.Close()

	// Query Firestore
	var results []map[string]interface{}
	if allPackages {
		packages := client.Collection("repositories").Documents(ctx)
		for {
			doc, err := packages.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				log.Println("searchPackages error:", err)
				c.JSON(http.StatusBadRequest, gin.H{"message": "Unexpected error"})
				return
			}
			
			log.Println("searchPackages: append version" + doc.Data()["version"].(string))
			log.Println("searchPackages: append name" + doc.Data()["name"].(string))
			results = append(results, map[string]interface{}{
				"Version": doc.Data()["version"].(string), 
				"Name": doc.Data()["name"].(string), 
				"ID": doc.Ref.ID,
			})
		}
	} else {
		for _, query := range queries {
			packages := client.Collection("repositories").Where("name", "==",query.Name).Documents(ctx)
			for {
				doc, err := packages.Next()
				if err == iterator.Done {
					break
				}
				if err != nil {
					log.Println("searchPackages error:", err)
					c.JSON(http.StatusBadRequest, gin.H{"message": "Unexpected error"})
					return
				}
				version := doc.Data()["version"].(string)
				if query.Version != "" {
					// Compare package version with query version
					if !compareVersion(query.Version, version) {
						continue
					}
				} else {
					continue
				}

				log.Println("searchPackages: append version" + version)
				log.Println("searchPackages: append name" + doc.Data()["name"].(string))
				results = append(results, map[string]interface{}{
					"Version": version, 
					"Name": doc.Data()["name"].(string), 
					"ID": doc.Ref.ID,
				})
			}
		}
	}

	// Check if the result set is too large
	if len(results) > 100 { // Set the maximum limit
		log.Println("searchPackages error: Too many packages returned")
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"message": "Too many packages returned"})
		return
	}

	// Prepare paginated response
	pageSize := 10 // Set your desired page size
	page, _ := strconv.Atoi(offset)
	start := (page - 1) * pageSize
	end := start + pageSize
	if end > len(results) {
		end = len(results)
	}
	pageResults := results[start:end]

	// Send the response
	c.JSON(http.StatusOK, pageResults)
}

func compareVersion(constraint, version string) bool {
	c, err := semver.NewConstraint(constraint)
	if err != nil {
		return false
	}

	v, err := semver.NewVersion(version)
	if err != nil {
		return false
	}

	return c.Check(v)
}