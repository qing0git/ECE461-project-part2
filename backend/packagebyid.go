package main

import (
	"context"
	"net/http"
	"log"
	"io"
	"os"
	"path/filepath"
	"encoding/base64"
	"strings"
	firebase "firebase.google.com/go"
	"github.com/gin-gonic/gin"
	"cloud.google.com/go/firestore"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/codes"
	"github.com/tidwall/gjson"
	"github.com/google/go-github/v50/github"
	"golang.org/x/oauth2"
	"github.com/mholt/archiver/v3"
)

func getPackageByID(c *gin.Context) {
	id := c.Param("id")
	log.Println("getPackageByID: " + id)

	// Retrieve the package from the Firestore database using the provided ID
	ctx := context.Background()
	conf := &firebase.Config{ProjectID: "ece461-pj-part2"}
	app, err := firebase.NewApp(ctx, conf)
	if err != nil {
		log.Println("getPackageByID error:", err)
		c.JSON(http.StatusBadRequest, gin.H{"message": "Unexpected error"})
		return
	}
	databaseClient, err := app.Firestore(ctx)
	if err != nil {
		log.Println("getPackageByID error:", err)
		c.JSON(http.StatusBadRequest, gin.H{"message": "Unexpected error"})
		return
	}
	defer databaseClient.Close()

	storageClient, err := app.Storage(ctx)
	if err != nil {
		log.Println("getPackageByID error:", err)
		c.JSON(http.StatusBadRequest, gin.H{"message": "Unexpected error"})
		return
	}

	pack, err := databaseClient.Collection("repositories").Doc(id).Get(ctx)
	if err != nil {
		log.Println("getPackageByID error:", err)
		if status.Code(err) == codes.NotFound {
			c.JSON(http.StatusNotFound, gin.H{"message": "Package does not exist"})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Unexpected error"})
		}
		return
	}

	bucket, err := storageClient.Bucket("ece461-pj-part2.appspot.com")
	if err != nil {
		log.Println("getPackageByID error:", err)
		c.JSON(http.StatusBadRequest, gin.H{"message": "Unexpected error"})
		return
	}

	object := bucket.Object(id)
	reader, err := object.NewReader(ctx)
	if err != nil {
		log.Println("getPackageByID error:", err)
		c.JSON(http.StatusBadRequest, gin.H{"message": "Unexpected error"})
		return
	}
	defer reader.Close()

	zipAsbase64, err := io.ReadAll(reader)
	if err != nil {
		log.Println("getPackageByID error:", err)
		c.JSON(http.StatusBadRequest, gin.H{"message": "Unexpected error"})
		return
	}

	log.Println("getPackageByID: Name:" + pack.Data()["name"].(string))
	log.Println("getPackageByID: Version:" + pack.Data()["version"].(string))
	log.Println("getPackageByID: githubURL:" + pack.Data()["githubURL"].(string))
	
	c.JSON(http.StatusOK, map[string]interface{}{
		"metadata": map[string]interface{}{
			"Name": pack.Data()["name"].(string),
			"Version": pack.Data()["version"].(string),
			"ID": id,
		},
		"data": map[string]interface{}{
			"Content": string(zipAsbase64),
			"JSProgram": pack.Data()["jsprogram"].(string),
		},
	})
}

func updatePackageByID(c *gin.Context) {
	id := c.Param("id")
	log.Println("updatePackageByID: " + id)

	// Parse the request body
	var req UpdatePackageRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		log.Println("updatePackageByID error:", err)
		c.JSON(http.StatusBadRequest, gin.H{"message": "Missing or improperly formed field(s) in the request"})
		return
	}

	// Check if both Content and URL are set
	if req.Data.Content != "" && req.Data.URL != "" {
		log.Println("updatePackageByID error: Content and URL cannot be both set", err)
		c.JSON(http.StatusBadRequest, gin.H{"message": "Content and URL cannot be both set"})
		return
	}

	// Check if package exists in the Firestore database
	ctx := context.Background()
	conf := &firebase.Config{ProjectID: "ece461-pj-part2"}
	app, err := firebase.NewApp(ctx, conf)
	if err != nil {
		log.Println("updatePackageByID error:", err)
		c.JSON(http.StatusBadRequest, gin.H{"message": "Unexpected error"})
		return
	}
	databaseClient, err := app.Firestore(ctx)
	if err != nil {
		log.Println("updatePackageByID error:", err)
		c.JSON(http.StatusBadRequest, gin.H{"message": "Unexpected error"})
		return
	}
	defer databaseClient.Close()

	storageClient, err := app.Storage(ctx)
	if err != nil {
		log.Println("updatePackageByID error:", err)
		c.JSON(http.StatusBadRequest, gin.H{"message": "Unexpected error"})
		return
	}

	ref := databaseClient.Collection("repositories").Doc(id)
	pack, err := ref.Get(ctx)
	if err != nil {
		log.Println("updatePackageByID error:", err)
		if status.Code(err) == codes.NotFound {
			c.JSON(http.StatusNotFound, gin.H{"message": "Package does not exist"})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Unexpected error"})
		}
		return
	}
	name := pack.Data()["name"].(string)
	version := pack.Data()["version"].(string)
	log.Println("updatePackageByID: Name:" + name)
	log.Println("updatePackageByID: Version:" + version)

	var newName string
	var newVersion string
	var newGithubURL string
	// Process the request
	// If the request contains a base64 string
	if req.Data.Content != "" {
		log.Println("updatePackageByID: Uploading by base64")
		// Decode the base64 content
		decoded, err := base64.StdEncoding.DecodeString(req.Data.Content)
		if err != nil {
			log.Println("updatePackageByID error:", err)
			c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid base64 content"})
			return
		}

		// Unzip the content
		tempDir, err := os.MkdirTemp("", "repo")
		if err != nil {
			log.Println("updatePackageByID error:", err)
			c.JSON(http.StatusBadRequest, gin.H{"message": "Failed to read the zip content"})
			return
		}
		defer os.RemoveAll(tempDir)

		tempFile := filepath.Join(tempDir, "repo.zip")
		err = os.WriteFile(tempFile, decoded, 0644)
		if err != nil {
			log.Println("updatePackageByID error:", err)
			c.JSON(http.StatusBadRequest, gin.H{"message": "Failed to read the zip content"})
			return
		}

		err = archiver.Unarchive(tempFile, tempDir)
		if err != nil {
			log.Println("updatePackageByID error:", err)
			c.JSON(http.StatusBadRequest, gin.H{"message": "Failed to read the zip content"})
			return
		}

		extractedDirs, err := os.ReadDir(tempDir)
		if err != nil {
			log.Println("updatePackageByID error:", err)
			c.JSON(http.StatusBadRequest, gin.H{"message": "Failed to read the zip content"})
			return
		}

		if len(extractedDirs) == 0 {
			log.Println("updatePackageByID error: Failed to read the zip content")
			c.JSON(http.StatusBadRequest, gin.H{"message": "Failed to read the zip content"})
			return
		}
	
		extractedDir := filepath.Join(tempDir, extractedDirs[0].Name())
		packageJSONPath := filepath.Join(extractedDir, "package.json")
		packageJSON, err := os.ReadFile(packageJSONPath)
	
		if err != nil {
			log.Println("updatePackageByID error:(Missing package.json file)", err)
			c.JSON(http.StatusBadRequest, gin.H{"message": "Missing package.json file"})
			return
		}

		// Get the metadata from the package.json file
		newName = gjson.Get(string(packageJSON), "name").String()
		log.Println("updatePackageByID: Uploading by base64: newName: " + newName)
		newVersion = gjson.Get(string(packageJSON), "version").String()
		log.Println("updatePackageByID: Uploading by base64: newVersion: " + newVersion)
		if newName != name {
			log.Println("updatePackageByID error: Uploading by base64: name does not match")
			c.JSON(http.StatusNotFound, gin.H{"message": "Package does not exist"})
			return
		}
		if newVersion != version {
			log.Println("updatePackageByID error: Uploading by base64: version does not match")
			c.JSON(http.StatusNotFound, gin.H{"message": "Package does not exist"})
			return
		}
		
		homepage := gjson.Get(string(packageJSON), "homepage").String()
		repositoryURL := gjson.Get(string(packageJSON), "repository.url").String()
		newGithubURL = ""
		if strings.Contains(homepage, "github.com") {
			newGithubURL = homepage
		} else if strings.Contains(repositoryURL, "github.com") {
			newGithubURL = repositoryURL
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"message": "GitHub URL not found"})
			return
		}
		newGithubURL = strings.TrimSuffix(newGithubURL, ".git")
		owner, repo, err := parseGitHubURL(newGithubURL)
		if err != nil {
			log.Println("updatePackageByID error:", err)
			c.JSON(http.StatusBadRequest, gin.H{"message": "Unexpected error"})
			return
		}
		fullRepo := owner + "/" + repo
		newGithubURL = "https://github.com/" + fullRepo
		log.Println("updatePackageByID: Uploading by base64: newGithubURL: " + newGithubURL)
	}
	// If the request contains a GitHub URL
	if req.Data.URL != "" {
		log.Println("updatePackageByID: Uploading by URL")
		// Download the GitHub repository as a zip file
		req.Data.URL = strings.TrimSuffix(req.Data.URL, ".git")
		owner, repo, err := parseGitHubURL(req.Data.URL)
		if err != nil {
			log.Println("updatePackageByID error:", err)
			c.JSON(http.StatusBadRequest, gin.H{"message": "Unexpected error"})
			return
		}
		fullRepo := owner + "/" + repo
		newGithubURL = "https://github.com/" + fullRepo
		log.Println("updatePackageByID: Uploading by URL: newGithubURL: " + newGithubURL)
		resp, err := http.Get(req.Data.URL + "/archive/refs/heads/master.zip")
		if err != nil || resp.StatusCode != http.StatusOK {
			log.Println("updatePackageByID error:", err)
			c.JSON(http.StatusBadRequest, gin.H{"message": "Failed to download the repository"})
			return
		}
		defer resp.Body.Close()

		// Read the zip content
		zipContent, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Println("updatePackageByID error:", err)
			c.JSON(http.StatusBadRequest, gin.H{"message": "Failed to read the zip content"})
			return
		}

		// Unzip the content
		tempDir, err := os.MkdirTemp("", "repo")
		if err != nil {
			log.Println("updatePackageByID error:", err)
			c.JSON(http.StatusBadRequest, gin.H{"message": "Failed to read the zip content"})
			return
		}
		defer os.RemoveAll(tempDir)

		tempFile := filepath.Join(tempDir, "repo.zip")
		err = os.WriteFile(tempFile, zipContent, 0644)
		if err != nil {
			log.Println("updatePackageByID error:", err)
			c.JSON(http.StatusBadRequest, gin.H{"message": "Failed to read the zip content"})
			return
		}

		err = archiver.Unarchive(tempFile, tempDir)
		if err != nil {
			log.Println("updatePackageByID error:", err)
			c.JSON(http.StatusBadRequest, gin.H{"message": "Failed to read the zip content"})
			return
		}

		extractedDirs, err := os.ReadDir(tempDir)
		if err != nil {
			log.Println("updatePackageByID error:", err)
			c.JSON(http.StatusBadRequest, gin.H{"message": "Failed to read the zip content"})
			return
		}

		if len(extractedDirs) == 0 {
			log.Println("updatePackageByID error: Failed to read the zip content")
			c.JSON(http.StatusBadRequest, gin.H{"message": "Failed to read the zip content"})
			return
		}
	
		extractedDir := filepath.Join(tempDir, extractedDirs[0].Name())
		packageJSONPath := filepath.Join(extractedDir, "package.json")
		packageJSON, err := os.ReadFile(packageJSONPath)
	
		if err != nil {
			log.Println("updatePackageByID error(Missing package.json file):", err)
			c.JSON(http.StatusBadRequest, gin.H{"message": "Missing package.json file"})
			return
		}

		// Remove the .git directory and create a new zip file
		newZipPath := filepath.Join(tempDir, extractedDirs[0].Name()+"-no-git.zip")
		newZipFile, err := os.Create(newZipPath)
		if err != nil {
			log.Println("updatePackageByID error:", err)
			c.JSON(http.StatusBadRequest, gin.H{"message": "Failed to write the zip content"})
			return
		}
		defer newZipFile.Close()

		newZip := archiver.NewZip()

		err = newZip.Create(newZipFile)
		if err != nil {
			log.Println("updatePackageByID error:", err)
			c.JSON(http.StatusBadRequest, gin.H{"message": "Failed to write the zip content"})
			return
		}
		defer newZip.Close()

		err = filepath.Walk(extractedDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// Skip directories and files inside the ".git" directory
			if strings.Contains(path, "/.git/") || (info.IsDir() && filepath.Base(path) == ".git") {
				return nil
			}

			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()

			relativePath, _ := filepath.Rel(extractedDir, path)
			err = newZip.Write(archiver.File{
				FileInfo: archiver.FileInfo{
					FileInfo:   info,
					CustomName: relativePath,
				},
				ReadCloser: file,
			})

			return err
		})

		if err != nil {
			log.Println("updatePackageByID error:", err)
			c.JSON(http.StatusBadRequest, gin.H{"message": "Failed to write the zip content"})
			return
		}
	
		// Read the new zip file and encode it to base64
		newZipContent, err := os.ReadFile(newZipPath)
		if err != nil {
			log.Println("updatePackageByID error:", err)
			c.JSON(http.StatusBadRequest, gin.H{"message": "Failed to write the zip content"})
			return
		}

		// Encode the new zip content as base64
		req.Data.Content = base64.StdEncoding.EncodeToString(newZipContent)

		// Get the metadata from the package.json file
		newName = gjson.Get(string(packageJSON), "name").String()
		log.Println("updatePackageByID: Uploading by URL: newName: " + newName)
		newVersion = gjson.Get(string(packageJSON), "version").String()
		log.Println("updatePackageByID: Uploading by URL: newVersion: " + newVersion)
		if newName != name {
			log.Println("updatePackageByID error: Uploading by URL: name does not match")
			c.JSON(http.StatusNotFound, gin.H{"message": "Package does not exist"})
			return
		}
		if newVersion != version {
			log.Println("updatePackageByID error: Uploading by URL: version does not match")
			c.JSON(http.StatusNotFound, gin.H{"message": "Package does not exist"})
			return
		}

		newGithubURL = req.Data.URL
	}
	client := github.NewClient(oauth2.NewClient(context.Background(), oauth2.StaticTokenSource(&oauth2.Token{AccessToken: "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"})))
	repoOwner, repoName, err := parseGitHubURL(newGithubURL)
	if err != nil {
		log.Println("updatePackageByID error:", err)
		c.JSON(http.StatusBadRequest, gin.H{"message": "Failed to parse GitHub URL"})
		return
	}

	readmeContent, err := getRepoReadme(client, repoOwner, repoName)
	if err != nil {
		log.Println("updatePackageByID error:", err)
		c.JSON(http.StatusBadRequest, gin.H{"message": "Failed to get README"})
		return
	}

	_, err = ref.Update(ctx, []firestore.Update{
		{Path: "githubURL", Value: newGithubURL},
		{Path: "readme", Value: readmeContent},
		{Path: "jsprogram", Value: req.Data.JSProgram},
	})
	if err != nil {
		log.Println("updatePackageByID error: error updating document:", err)
		c.JSON(http.StatusBadRequest, gin.H{"message": "Failed to parse GitHub URL"})
		return
	}

	bucket, err := storageClient.Bucket("ece461-pj-part2.appspot.com")
	if err != nil {
		log.Println("updatePackageByID error:", err)
		c.JSON(http.StatusBadRequest, gin.H{"message": "Unexpected error"})
		return
	}

	object := bucket.Object(id)
	writer := object.NewWriter(ctx)
	_, err = writer.Write([]byte(req.Data.Content))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Unexpected error"})
		log.Println("updatePackageByID error:Error uploading zip file:", err)
		return
	}

	// Close the writer and ensure the object is uploaded.
	if err := writer.Close(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Unexpected error"})
		log.Println("updatePackageByID error:Error closing writer:", err)
		return
	}

	log.Println("File updated")
	c.JSON(http.StatusOK, gin.H{"message": "Version is updated"})
}

func deletePackageByID(c *gin.Context) {
	id := c.Param("id")
	log.Println("deletePackageByID: " + id)

	// Check if package exists in the Firestore database
	ctx := context.Background()
	conf := &firebase.Config{ProjectID: "ece461-pj-part2"}
	app, err := firebase.NewApp(ctx, conf)
	if err != nil {
		log.Println("deletePackageByID error:", err)
		c.JSON(http.StatusBadRequest, gin.H{"message": "Unexpected error"})
		return
	}
	databaseClient, err := app.Firestore(ctx)
	if err != nil {
		log.Println("deletePackageByID error:", err)
		c.JSON(http.StatusBadRequest, gin.H{"message": "Unexpected error"})
		return
	}
	defer databaseClient.Close()

	storageClient, err := app.Storage(ctx)
	if err != nil {
		log.Println("deletePackageByID error:", err)
		c.JSON(http.StatusBadRequest, gin.H{"message": "Unexpected error"})
		return
	}

	ref := databaseClient.Collection("repositories").Doc(id)
	_, err = ref.Get(ctx)
	if err != nil {
		log.Println("deletePackageByID error:", err)
		if status.Code(err) == codes.NotFound {
			c.JSON(http.StatusNotFound, gin.H{"message": "Package does not exist"})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Unexpected error"})
		}
		return
	}
	
	_, err = ref.Delete(ctx)
	if err != nil {
		log.Println("deletePackageByID error:", err)
		c.JSON(http.StatusBadRequest, gin.H{"message": "Unexpected error"})
		return
	}

	bucket, err := storageClient.Bucket("ece461-pj-part2.appspot.com")
	if err != nil {
		log.Println("deletePackageByID error:", err)
		c.JSON(http.StatusBadRequest, gin.H{"message": "Unexpected error"})
		return
	}

	object := bucket.Object(id)
	if _, err := object.Attrs(ctx); err != nil {
		log.Println("deletePackageByID error: Error retrieving object attributes:", err)
		c.JSON(http.StatusBadRequest, gin.H{"message": "Unexpected error"})
		return
	}

	if err := object.Delete(ctx); err != nil {
		log.Println("deletePackageByID error: Error deleting object:", err)
		c.JSON(http.StatusBadRequest, gin.H{"message": "Unexpected error"})
		return
	}

	log.Println("deletePackageByID: Package is deleted: " + id)
	c.JSON(http.StatusOK, gin.H{"message": "Package is deleted"})
}