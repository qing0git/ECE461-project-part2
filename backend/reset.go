package main

import (
	"context"
	"net/http"
	"log"
	firebase "firebase.google.com/go"
	"github.com/gin-gonic/gin"
	"google.golang.org/api/iterator"
	"google.golang.org/api/storage/v1"
)

func resetRegistry(c *gin.Context) {
	// Delete all data from the Firestore database
	ctx := context.Background()
	conf := &firebase.Config{ProjectID: "ece461-pj-part2"}
	app, err := firebase.NewApp(ctx, conf)
	if err != nil {
		log.Println("resetRegistry error:", err)
		c.JSON(http.StatusBadRequest, gin.H{"message": "Unexpected error"})
		return
	}
	databaseClient, err := app.Firestore(ctx)
	if err != nil {
		log.Println("resetRegistry error:", err)
		c.JSON(http.StatusBadRequest, gin.H{"message": "Unexpected error"})
		return
	}
	defer databaseClient.Close()

	// Get all documents in the "packages" collection
	packages := databaseClient.Collection("repositories")
	iter := packages.Documents(ctx)

	// Iterate through the documents and delete them
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Println("resetRegistry error:", err)
			c.JSON(http.StatusBadRequest, gin.H{"message": "Unexpected error while resetting the registry"})
			return
		}

		_, err = doc.Ref.Delete(ctx)
		if err != nil {
			log.Println("resetRegistry error:", err)
			c.JSON(http.StatusBadRequest, gin.H{"message": "Unexpected error while resetting the registry"})
			return
		}
	}

	storageService, err := storage.NewService(ctx)
	if err != nil {
		log.Println("resetRegistry error:", err)
		c.JSON(http.StatusBadRequest, gin.H{"message": "Unexpected error while resetting the registry"})
		return
	}
	objectsService := storage.NewObjectsService(storageService)
	listCall := objectsService.List("ece461-pj-part2.appspot.com")

	var deletedObjects int

	for {
		objects, err := listCall.Do()
		if err != nil {
			log.Fatalf("resetRegistry error: Error listing objects: %v", err)
		}

		for _, object := range objects.Items {
			deleteCall := objectsService.Delete("ece461-pj-part2.appspot.com", object.Name)
			if err := deleteCall.Do(); err != nil {
				log.Printf("resetRegistry error: Error deleting object %s: %v", object.Name, err)
			} else {
				deletedObjects++
				log.Printf("resetRegistry: Deleted object %s", object.Name)
			}
		}

		if objects.NextPageToken == "" {
			break
		}

		listCall.PageToken(objects.NextPageToken)
	}

	log.Printf("resetRegistry: Deleted %d objects\n", deletedObjects)
	c.JSON(http.StatusOK, gin.H{"message": "Registry is reset"})
}
