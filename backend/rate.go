package main

import (
	"context"
	"net/http"
	"log"
	firebase "firebase.google.com/go"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/codes"
)

func ratePackage(c *gin.Context) {
	id := c.Param("id")
	log.Println("ratePackage: " + id)

	// Retrieve the package from the Firestore database using the provided ID
	ctx := context.Background()
	conf := &firebase.Config{ProjectID: "ece461-pj-part2"}
	app, err := firebase.NewApp(ctx, conf)
	if err != nil {
		log.Println("ratePackage error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Unexpected error"})
		return
	}
	databaseClient, err := app.Firestore(ctx)
	if err != nil {
		log.Println("ratePackage error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Unexpected error"})
		return
	}
	defer databaseClient.Close()

	pack, err := databaseClient.Collection("repositories").Doc(id).Get(ctx)
	if err != nil {
		log.Println("ratePackage error:", err)
		if status.Code(err) == codes.NotFound {
			c.JSON(http.StatusNotFound, gin.H{"message": "Package does not exist"})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Unexpected error"})
		}
		return
	}

	log.Println("ratePackage: Name:" + pack.Data()["name"].(string))
	log.Println("ratePackage: Version:" + pack.Data()["version"].(string))
	log.Println("ratePackage: githubURL:" + pack.Data()["githubURL"].(string))
	log.Printf("ratePackage: BusFactor: %f\n", pack.Data()["busfactor"].(float64))
	log.Printf("ratePackage: Correctness: %f\n", pack.Data()["correctness"].(float64))
	log.Printf("ratePackage: RampUp: %f\n", pack.Data()["rampup"].(float64))
	log.Printf("ratePackage: ResponsiveMaintainer: %f\n", pack.Data()["responsivemaintainer"].(float64))
	log.Printf("ratePackage: LicenseScore: %f\n", pack.Data()["licensescore"].(float64))
	log.Printf("ratePackage: GoodPinningPractice: %f\n", pack.Data()["goodpinningpractice"].(float64))
	log.Printf("ratePackage: PullRequest: %f\n", pack.Data()["pullrequest"].(float64))
	log.Printf("ratePackage: NetScore: %f\n", pack.Data()["netscore"].(float64))

	c.JSON(http.StatusOK, map[string]interface{}{
		"BusFactor": pack.Data()["busfactor"].(float64),
		"Correctness": pack.Data()["correctness"].(float64),
		"RampUp": pack.Data()["rampup"].(float64),
		"ResponsiveMaintainer": pack.Data()["responsivemaintainer"].(float64),
		"LicenseScore": pack.Data()["licensescore"].(float64),
		"GoodPinningPractice": pack.Data()["goodpinningpractice"].(float64),
		"PullRequest": pack.Data()["pullrequest"].(float64),
		"NetScore": pack.Data()["netscore"].(float64),
	})
}
