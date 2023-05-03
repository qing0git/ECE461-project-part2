package main

import (
	"net/http"
	"context"
	"log"
	"fmt"
	"regexp"
	firebase "firebase.google.com/go"
	"github.com/gin-gonic/gin"
	"google.golang.org/api/iterator"
	"encoding/base64"
	"io"
	"os"
	"archive/zip"
	"bytes"
	"compress/flate"
	// "math"
	"path/filepath"
	"strings"
	"strconv"
	"os/exec"
	"encoding/json"
	"github.com/tidwall/gjson"
	"github.com/google/go-github/v50/github"
	"golang.org/x/oauth2"
	"github.com/mholt/archiver/v3"
	"github.com/Masterminds/semver/v3"
)

type attribute struct {
	netScore float64
	rampUp float64
	correctness float64
	busFactor float64
	responsiveness float64
	license float64
	goodPinningPractice float64
	pullRequest float64
}

type gitObject struct {
	numCommits string
	numPullRequests int
}

func createPackage(c *gin.Context) {
	// Read the request body and parse the JSON
	var req CreatePackageRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		log.Println("createPackage error:", err)
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request body"})
		return
	}

	ctx := context.Background()
	conf := &firebase.Config{ProjectID: "ece461-pj-part2"}
	app, err := firebase.NewApp(ctx, conf)
	if err != nil {
		log.Println("createPackage error:", err)
		c.JSON(http.StatusBadRequest, gin.H{"message": "Unexpected error"})
		return
	}
	databaseClient, err := app.Firestore(ctx)
	if err != nil {
		log.Println("createPackage error:", err)
		c.JSON(http.StatusBadRequest, gin.H{"message": "Unexpected error"})
		return
	}
	defer databaseClient.Close()

	// Check if both Content and URL are set
	if req.Content != "" && req.URL != "" {
		log.Println("createPackage error: both Content and URL are set")
		c.JSON(http.StatusBadRequest, gin.H{"message": "Content and URL cannot be both set"})
		return
	}

	githubToken := os.Getenv("GITHUB_TOKEN")
	client := github.NewClient(oauth2.NewClient(ctx, oauth2.StaticTokenSource(&oauth2.Token{AccessToken: githubToken})))
	
	var name string
	var version string
	var githubURL string
	var scoreObject attribute
	// Process the request
	// If the request contains a base64 string
	if req.Content != "" {
		log.Println("createPackage: Uploading by base64")
		// Decode the base64 content
		decoded, err := base64.StdEncoding.DecodeString(req.Content)
		if err != nil {
			log.Println("createPackage error:", err)
			c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid base64 content"})
			return
		}

		// Unzip the content
		tempDir, err := os.MkdirTemp("", "repo")
		if err != nil {
			log.Println("createPackage error:", err)
			c.JSON(http.StatusBadRequest, gin.H{"message": "Failed to read the zip content"})
			return
		}
		defer os.RemoveAll(tempDir)

		tempFile := filepath.Join(tempDir, "repo.zip")
		err = os.WriteFile(tempFile, decoded, 0644)
		if err != nil {
			log.Println("createPackage error:", err)
			c.JSON(http.StatusBadRequest, gin.H{"message": "Failed to read the zip content"})
			return
		}

		err = archiver.Unarchive(tempFile, tempDir)
		if err != nil {
			log.Println("createPackage error:", err)
			c.JSON(http.StatusBadRequest, gin.H{"message": "Failed to read the zip content"})
			return
		}

		err = os.Remove(tempFile)
		if err != nil {
			log.Println("createPackage error:", err)
			c.JSON(http.StatusBadRequest, gin.H{"message": "Failed to delete the zip file after unzipped"})
			return
		}

		extractedDirs, err := os.ReadDir(tempDir)
		if err != nil {
			log.Println("createPackage error:", err)
			c.JSON(http.StatusBadRequest, gin.H{"message": "Failed to read the zip content"})
			return
		}

		if len(extractedDirs) == 0 {
			log.Println("createPackage error: Failed to read the zip content")
			c.JSON(http.StatusBadRequest, gin.H{"message": "Failed to read the zip content"})
			return
		}
	
		var packageJSONPath string
		err = filepath.Walk(tempDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() && filepath.Base(path) == "package.json" {
				packageJSONPath = path
				return filepath.SkipDir
			}
			return nil
		})

		if err != nil {
			log.Println("createPackage error:(Finding package.json file)", err)
			c.JSON(http.StatusBadRequest, gin.H{"message": "Fail to find package.json file"})
			return
		}
		//extractedDir := filepath.Join(tempDir, extractedDirs[1].Name())
		//packageJSONPath := filepath.Join(extractedDir, "package.json")
		packageJSON, err := os.ReadFile(packageJSONPath)
	
		if err != nil {
			log.Println("createPackage error:(Missing package.json file)", err)
			c.JSON(http.StatusBadRequest, gin.H{"message": "Missing package.json file"})
			return
		}

		// Get the metadata from the package.json file
		name = gjson.Get(string(packageJSON), "name").String()
		log.Println("createPackage: Uploading by base64: Name: " + name)
		version = gjson.Get(string(packageJSON), "version").String()
		log.Println("createPackage: Uploading by base64: Version: " + version)
		homepage := gjson.Get(string(packageJSON), "homepage").String()
		repositoryURL := gjson.Get(string(packageJSON), "repository.url").String()
		if strings.Contains(homepage, "github.com") {
			githubURL = homepage
		} else if strings.Contains(repositoryURL, "github.com") {
			githubURL = repositoryURL
		} else {
			log.Println("createPackage error: GitHub URL not found")
			c.JSON(http.StatusBadRequest, gin.H{"message": "GitHub URL not found"})
			return
		}
		githubURL = strings.TrimSuffix(githubURL, ".git")
		owner, repo, err := parseGitHubURL(githubURL)
		if err != nil {
			log.Println("createPackage error:", err)
			c.JSON(http.StatusBadRequest, gin.H{"message": "Unexpected error"})
			return
		}
		fullRepo := owner + "/" + repo
		githubURL = "https://github.com/" + fullRepo
		log.Println("createPackage: Uploading by base64: githubURL: " + githubURL)
		scoreObject.goodPinningPractice = calculatePinScore(string(packageJSON))
	}

	// If the request contains a GitHub URL
	if req.URL != "" {
		log.Println("createPackage: Uploading by URL")
		// Download the GitHub repository as a zip file
		req.URL = strings.TrimSuffix(req.URL, ".git")
		owner, repo, err := parseGitHubURL(req.URL)
		if err != nil {
			log.Println("createPackage error:", err)
			c.JSON(http.StatusBadRequest, gin.H{"message": "Unexpected error"})
			return
		}
		fullRepo := owner + "/" + repo
		githubURL = "https://github.com/" + fullRepo
		log.Println("createPackage: Uploading by URL: githubURL: " + githubURL)
		resp, err := http.Get(githubURL + "/archive/refs/heads/master.zip")
		if err != nil || resp.StatusCode != http.StatusOK {
			log.Println("createPackage error:", err)
			c.JSON(http.StatusBadRequest, gin.H{"message": "Failed to download the repository"})
			return
		}
		defer resp.Body.Close()

		// Read the zip content
		zipContent, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Println("createPackage error:", err)
			c.JSON(http.StatusBadRequest, gin.H{"message": "Failed to read the zip content"})
			return
		}

		// Unzip the content
		tempDir, err := os.MkdirTemp("", "repo")
		if err != nil {
			log.Println("createPackage error:", err)
			c.JSON(http.StatusBadRequest, gin.H{"message": "Failed to read the zip content"})
			return
		}
		defer os.RemoveAll(tempDir)

		tempFile := filepath.Join(tempDir, "repo.zip")
		err = os.WriteFile(tempFile, zipContent, 0644)
		if err != nil {
			log.Println("createPackage error:", err)
			c.JSON(http.StatusBadRequest, gin.H{"message": "Failed to read the zip content"})
			return
		}

		err = archiver.Unarchive(tempFile, tempDir)
		if err != nil {
			log.Println("createPackage error:", err)
			c.JSON(http.StatusBadRequest, gin.H{"message": "Failed to read the zip content"})
			return
		}

		err = os.Remove(tempFile)
		if err != nil {
			log.Println("createPackage error:", err)
			c.JSON(http.StatusBadRequest, gin.H{"message": "Failed to delete the zip file after unzipped"})
			return
		}

		extractedDirs, err := os.ReadDir(tempDir)
		if err != nil {
			log.Println("createPackage error:", err)
			c.JSON(http.StatusBadRequest, gin.H{"message": "Failed to read the zip content"})
			return
		}

		if len(extractedDirs) == 0 {
			log.Println("createPackage error: Failed to read the zip content")
			c.JSON(http.StatusBadRequest, gin.H{"message": "Failed to read the zip content"})
			return
		}
	
		var packageJSONPath string
		err = filepath.Walk(tempDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() && filepath.Base(path) == "package.json" {
				packageJSONPath = path
				return filepath.SkipDir
			}
			return nil
		})

		if err != nil {
			log.Println("createPackage error:(Finding package.json file)", err)
			c.JSON(http.StatusBadRequest, gin.H{"message": "Fail to find package.json file"})
			return
		}
		//extractedDir := filepath.Join(tempDir, extractedDirs[1].Name())
		//packageJSONPath := filepath.Join(extractedDir, "package.json")
		packageJSON, err := os.ReadFile(packageJSONPath)
	
		if err != nil {
			log.Println("createPackage error:(Missing package.json file)", err)
			c.JSON(http.StatusBadRequest, gin.H{"message": "Missing package.json file"})
			return
		}

		// Remove the .git directory and create a new zip file
		// Read the zip file into a zip.Reader
		zipReader, err := zip.NewReader(bytes.NewReader(zipContent), int64(len(zipContent)))
		if err != nil {
			log.Println("createPackage error:", err)
			return
		}
		// Create a buffer to store the new zip content
		var newZipBuffer bytes.Buffer
		newZipWriter := zip.NewWriter(&newZipBuffer)

		// Iterate through the files in the original zip, skipping ".git" files and directories
		for _, file := range zipReader.File {
			if strings.Contains(file.Name, "/.git/") || strings.HasSuffix(file.Name, ".git") {
				continue
			}

			newFileHeader, err := zip.FileInfoHeader(file.FileInfo())
			if err != nil {
				log.Println("createPackage error:", err)
				return
			}

			newFileHeader.Name = file.Name
			newFileHeader.Method = zip.Deflate // Set compression method to Deflate

			newFileWriter, err := newZipWriter.CreateHeader(newFileHeader)
			if err != nil {
				log.Println("createPackage error:", err)
				return
			}

			fileReader, err := file.Open()
			if err != nil {
				log.Println("createPackage error:", err)
				return
			}

			// Create a flate.Writer with the best compression level
			flateWriter, _ := flate.NewWriter(newFileWriter, flate.BestCompression)
			_, err = io.Copy(flateWriter, fileReader)
			if err != nil {
				log.Println("createPackage error:", err)
				fileReader.Close()
				return
			}

			fileReader.Close()
			flateWriter.Close() // Close the flate.Writer
		}

		err = newZipWriter.Close()
		if err != nil {
			log.Println("createPackage error:", err)
			return
		}

		// Encode the new zip content as base64
		req.Content = base64.StdEncoding.EncodeToString(newZipBuffer.Bytes())

		// Get the metadata from the package.json file
		name = gjson.Get(string(packageJSON), "name").String()
		log.Println("createPackage: Uploading by URL: Name: " + name)
		version = gjson.Get(string(packageJSON), "version").String()
		log.Println("createPackage: Uploading by URL: Version: " + version)

		scoreObject.goodPinningPractice = calculatePinScore(string(packageJSON))
	}

	iter := databaseClient.Collection("repositories").Where("name", "==", name).Where("version", "==", version).Documents(ctx)
	defer iter.Stop()

	// Initialize counter for matched documents
	count := 0

	// Iterate through documents and increment the counter
	for {
		_, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Printf("Error iterating documents: %v", err)
		}

		count++
	}
	if count != 0 {
		log.Println("createPackage error: Package exists already" + version)
		c.JSON(http.StatusConflict, gin.H{"message": "Package exists already"})
		return
	}

	if strings.Contains(githubURL, "github.com") {
		var gitObj gitObject
		err = githubFunc(githubURL, &scoreObject, &gitObj, 0, ctx, client)
		if err != nil {
			log.Println("createPackage error:", err)
			c.JSON(http.StatusBadRequest, gin.H{"message": "rating error"})
			return
		}
		githubCalcScores(&scoreObject, &gitObj)
	} else {
		log.Println("createPackage error:", githubURL, "is not a valid URL")
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid github url"})
		return
	}

	// scoreObject.netScore = math.Round(scoreObject.netScore * 10) / 10
	log.Printf("createPackage: netscore: %f\n", scoreObject.netScore)
	// scoreObject.rampUp = math.Round(scoreObject.rampUp * 10) / 10
	log.Printf("createPackage: rampup: %f\n", scoreObject.rampUp)
	// scoreObject.correctness = math.Round(scoreObject.correctness * 10) / 10
	log.Printf("createPackage: correctness: %f\n", scoreObject.correctness)
	// scoreObject.busFactor = math.Round(scoreObject.busFactor * 10) / 10
	log.Printf("createPackage: busfactor: %f\n", scoreObject.busFactor)
	// scoreObject.responsiveness = math.Round(scoreObject.responsiveness * 10) / 10
	log.Printf("createPackage: responsivemaintainer: %f\n", scoreObject.responsiveness)
	// scoreObject.license = math.Round(scoreObject.license * 10) / 10
	log.Printf("createPackage: licensescore: %f\n", scoreObject.license)
	// scoreObject.goodPinningPractice = math.Round(scoreObject.goodPinningPractice * 10) / 10
	log.Printf("createPackage: goodpinningpractice: %f\n", scoreObject.goodPinningPractice)
	// scoreObject.pullRequest = math.Round(scoreObject.pullRequest * 10) / 10
	log.Printf("createPackage: pullrequest: %f\n", scoreObject.pullRequest)

	if scoreObject.busFactor < 0.5 && scoreObject.correctness < 0.5 && scoreObject.goodPinningPractice < 0.5 && scoreObject.license < 0.5 && scoreObject.pullRequest < 0.5 && scoreObject.rampUp < 0.5 && scoreObject.responsiveness < 0.5 {
		log.Println("createPackage error: Package is not uploaded due to the disqualified rating")
		c.JSON(http.StatusFailedDependency, gin.H{"message": "Package is not uploaded due to the disqualified rating"})
		return
	}

	repoOwner, repoName, err := parseGitHubURL(githubURL)
	if err != nil {
		log.Println("createPackage error:", err)
		c.JSON(http.StatusBadRequest, gin.H{"message": "Failed to parse GitHub URL"})
		return
	}

	readmeContent, err := getRepoReadme(client, repoOwner, repoName)
	if err != nil {
		log.Println("createPackage error:", err)
		c.JSON(http.StatusBadRequest, gin.H{"message": "Failed to get README"})
		return
	}

	storageClient, err := app.Storage(ctx)
	if err != nil {
		log.Println("createPackage error:", err)
		c.JSON(http.StatusBadRequest, gin.H{"message": "Unexpected error"})
		return
	}

	ref := databaseClient.Collection("repositories").NewDoc()
	_, err = ref.Set(ctx, map[string]interface{}{
		"name": name,
		"version": version,
		"githubURL": githubURL,
		"readme": readmeContent,
		"jsprogram": req.JSProgram,
		"netscore": scoreObject.netScore,
		"rampup": scoreObject.rampUp,
		"correctness": scoreObject.correctness,
		"busfactor": scoreObject.busFactor,
		"responsivemaintainer": scoreObject.responsiveness,
		"licensescore": scoreObject.license,
		"goodpinningpractice": scoreObject.goodPinningPractice,
		"pullrequest": scoreObject.pullRequest,
		})
	if err != nil {
		log.Println("createPackage error:", err)
		c.JSON(http.StatusBadRequest, gin.H{"message": "Unexpected error"})
		return
	}

	bucket, err := storageClient.Bucket("ece461-pj-part2.appspot.com")
	if err != nil {
		log.Println("createPackage error:", err)
		c.JSON(http.StatusBadRequest, gin.H{"message": "Unexpected error"})
		return
	}

	object := bucket.Object(ref.ID)
	writer := object.NewWriter(ctx)
	_, err = writer.Write([]byte(req.Content))
	if err != nil {
		log.Println("createPackage error:Error uploading zip file:", err)
		c.JSON(http.StatusBadRequest, gin.H{"message": "Unexpected error"})
		return
	}

	// Close the writer and ensure the object is uploaded.
	if err := writer.Close(); err != nil {
		log.Println("createPackage error:Error closing writer:", err)
		c.JSON(http.StatusBadRequest, gin.H{"message": "Unexpected error"})
		return
	}

	log.Println("File updated")

	metadata := map[string]string{
		"Name": name,
		"Version": version,
		"ID": ref.ID,
	}

	data := map[string]string{
		"Content": req.Content,
		"JSProgram": req.JSProgram,
	}

	// Return the response
	c.JSON(http.StatusCreated, gin.H{
		"metadata": metadata,
		"data": data,
	})
}

func parseGitHubURL(url string) (string, string, error) {
	re := regexp.MustCompile(`github\.com/([^/]+)/([^/]+)`)
	matches := re.FindStringSubmatch(url)
	if len(matches) < 3 {
		return "", "", fmt.Errorf("failed to parse gitHub url")
	}
	return matches[1], matches[2], nil
}

func getRepoReadme(client *github.Client, owner string, repo string) (string, error) {
	readme, _, err := client.Repositories.GetReadme(context.Background(), owner, repo, nil)
	if err != nil {
		return "", err
	}

	content, err := readme.GetContent()
	if err != nil {
		return "", err
	}

	return content, nil
}

func githubCalcScores(scoreObject *attribute, gitObj *gitObject) {
	//Calculate responsiveness
	if f, err := strconv.ParseFloat(gitObj.numCommits, 32); err == nil {
		scoreObject.responsiveness = f
		//Calculate pullRequest
		scoreObject.pullRequest = float64(gitObj.numPullRequests) / f
		if scoreObject.pullRequest >= 1.0 {
			scoreObject.pullRequest = 1.0
		}
		if scoreObject.pullRequest <= 0.0 {
			scoreObject.pullRequest = 0.0
		}
	}
	maxResValue := 750.0 //max num of commits
	minResValue := 50.0  //min num of commits
	maxResScore := 1.0
	minResScore := 0.0

	if scoreObject.responsiveness <= minResValue {
		scoreObject.responsiveness = minResScore
	} else if scoreObject.responsiveness >= maxResValue {
		scoreObject.responsiveness = maxResScore
	} else {
		normalizedValue := (scoreObject.responsiveness - minResValue) / (maxResValue - minResValue)
		scoreObject.responsiveness = minResScore + normalizedValue*(maxResScore-minResScore)
	}
	//Calculate busFactor
	scoreObject.busFactor = float64(gitObj.numPullRequests)
	maxBusValue := 100.0 //max num of PR
	minBusValue := 10.0  //min num of PR
	maxBusScore := 1.0
	minBusScore := 0.0

	if float64(scoreObject.busFactor) <= minBusValue {
		scoreObject.busFactor = float64(minBusScore)
	} else if float64(scoreObject.busFactor) >= maxBusValue {
		scoreObject.busFactor = float64(maxBusScore)
	} else {
		normalizedValue := (float64(scoreObject.busFactor) - minBusValue) / (maxBusValue - minBusValue)
		scoreObject.busFactor = float64(minBusScore) + float64(normalizedValue)*(float64(maxBusScore)-float64(minBusScore))
	}
	//Calculate correctness
	scoreObject.correctness = float64(gitObj.numPullRequests)
	maxValue := 100.0 //max num of PR
	minValue := 10.0  //min num of PR
	maxScore := 1.0
	minScore := 0.0

	if scoreObject.correctness <= minValue {
		scoreObject.correctness = minScore
	} else if scoreObject.correctness >= maxValue {
		scoreObject.correctness = maxScore
	} else {
		normalizedValue := (scoreObject.correctness - minValue) / (maxValue - minValue)
		scoreObject.correctness = minScore + normalizedValue*(maxScore-minScore)
	}

	//Calculate rampUp
	//rampup = based on branches the less the easier to rampup
	scoreObject.rampUp = float64(gitObj.numPullRequests)
	maxPullValue := 1500.0
	normalizedValue := maxPullValue / scoreObject.rampUp
	if scoreObject.rampUp <= maxPullValue {
		scoreObject.rampUp = 1
	}
	if scoreObject.rampUp > maxPullValue {
		scoreObject.rampUp = normalizedValue
	}

	//avg of all
	scoreObject.netScore = (float64(scoreObject.busFactor) + float64(scoreObject.correctness) + float64(scoreObject.goodPinningPractice) + float64(scoreObject.license) + float64(scoreObject.pullRequest) + float64(scoreObject.rampUp) + float64(scoreObject.responsiveness)) / 7
}

func githubFunc(url string, scoreObject *attribute, gitObj *gitObject, count int, ctx context.Context, client *github.Client) (error){
	owner, repo, err := parseGitHubURL(url)
	if err != nil {
		return err
	}

	src, err := githubSource(url, count)
	if err != nil {
		return err
	}
	cleaned := strings.ReplaceAll(string(src), "\r", "")
	cleaned = strings.ReplaceAll(cleaned, "\n", "")
	gitObj.numCommits = cleaned

	var fullRepo string = owner + "/" + repo
	pullReq, err := githubPullReq(fullRepo)
	if err != nil {
		return err
	}
	gitObj.numPullRequests = pullReq
	licenseBool, err := githubLicense(fullRepo)
	if err != nil {
		return err
	}
	if licenseBool {
		scoreObject.license = 1
	} else {
		scoreObject.license = 0
	}

	//remove recently created directory after info is pulled
	err = os.RemoveAll("clonedir")

	if err != nil {
		return err
	}
	return nil
}

func githubSource(url string, count int) ([]byte, error) {
	//call python script that clones repo and pull number of commits
	command := exec.Command("python3", "cloner.py", url)
	output, err := command.Output()
	if err != nil {
		return nil, err
	}
	return output, nil
}

type PullRequests struct {
	TotalCount int `json:"total_count"`
}

func githubPullReq(repoName string) (int, error) {
	req, _ := http.NewRequest("GET", "https://api.github.com/search/issues?q=is:pr+repo:"+repoName, nil)
	req.Header.Set("Accept", "application/vnd.github+json")
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer res.Body.Close()

	var pullRequests PullRequests
	err = json.NewDecoder(res.Body).Decode(&pullRequests)
	if err != nil {
		return 0, err
	}
	return pullRequests.TotalCount, nil
}

type LicenseType struct {
	LicenseType struct {
		LicenseName string `json:"spdx_id"`
	} `json:"license"`
}

func githubLicense(repoName string) (bool, error) {
	req, _ := http.NewRequest("GET", "https://api.github.com/repos/"+repoName+"/license", nil)
	req.Header.Set("Accept", "application/vnd.github+json")
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return true, err
	}
	defer func() {
		closeErr := res.Body.Close()
		if closeErr != nil && err == nil {
			err = closeErr
		}
	}()

	var license LicenseType
	json.NewDecoder(res.Body).Decode(&license)
	return licenseCompatability(license.LicenseType.LicenseName), nil
}

func licenseCompatability(license string) (compatible bool) {
	licenseArr := [6]string{"MIT", "X11", "Public Domain", "BSD-new", "Apache 2.0", "LGPLv2.1"}

	for _, l := range licenseArr {
		if l == license {
			return true
		}
	}
	return false
}

func isPinned(constraintStr string) bool {
	_, err := semver.NewConstraint(constraintStr)
	return err == nil
}

func calculatePinScore(packageJSON string) float64 {
	devDependencies := gjson.Get(packageJSON, "devDependencies")
	if !devDependencies.IsObject() {
		return 1.0
	}

	pinnedCount := 0
	totalCount := 0

	devDependencies.ForEach(func(key, value gjson.Result) bool {
		totalCount++

		if isPinned(value.String()) {
			pinnedCount++
		}

		return true // continue iterating
	})

	return float64(pinnedCount) / float64(totalCount)
}
