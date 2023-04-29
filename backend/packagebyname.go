package main

import (
	"context"
	"net/http"
	"log"
	firebase "firebase.google.com/go"
	"github.com/gin-gonic/gin"
	//"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

func getPackageByName(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"message": "This system does not support authentication"})
}

func deletePackageByName(c *gin.Context) {
	name := c.Param("name")
	log.Println("deletePackageByName: " + name)

	// Check if package exists in the Firestore database
	ctx := context.Background()
	sa := option.WithCredentialsFile("./ece461-pj-part2-b75cfa849e87.json")
	app, err := firebase.NewApp(ctx, nil, sa)
	if err != nil {
		log.Println("deletePackageByName error: ", err)
		c.JSON(http.StatusBadRequest, gin.H{"message": "Unexpected error"})
		return
	}
	client, err := app.Firestore(ctx)
	if err != nil {
		log.Println("deletePackageByName error: ", err)
		c.JSON(http.StatusBadRequest, gin.H{"message": "Unexpected error"})
		return
	}
	defer client.Close()

	packages := client.Collection("repositories").Where("name", "==", name).Documents(ctx)
	count := 0
	for {
		doc, err := packages.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Println("deletePackageByName error: ", err)
			c.JSON(http.StatusBadRequest, gin.H{"message": "Unexpected error"})
			return
		}
		
		// Delete the package
		_, err = doc.Ref.Delete(ctx)
		if err != nil {
			log.Println("deletePackageByName error: ", err)
			c.JSON(http.StatusBadRequest, gin.H{"message": "Unexpected error"})
			return
		}
		count++
	}
	if count == 0 {
		log.Println("deletePackageByName error: Package does not exist")
		c.JSON(http.StatusNotFound, gin.H{"message": "Package does not exist"})
	} else {
		log.Println("deletePackageByName: deleted " + name)
		c.JSON(http.StatusOK, gin.H{"message": "Package is deleted"})
	}
}