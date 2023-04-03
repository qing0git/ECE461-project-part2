package main

import (
	"bufio"
	"context"

	//"encoding/binary"
	"encoding/json"
	"fmt"
	"io/ioutil"

	//"math"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	// "github.com/google/go-github/v50/github"
	"github.com/machinebox/graphql"
	// "golang.org/x/oauth2"
)

type attribute struct {
	url            string
	netScore       float64
	rampUp         float64
	correctness    float64
	busFactor      float32
	responsiveness float64
	license        int
}

type gitObject struct {
	// numCommits int
	numCommits string
	//numPullRequests float32
	numPullRequests int

	// graphQL         float32
	license    string
	stargazers int
	issues     int
	releases   int
}

type npmObject struct {
	numCommits     string
	numMaintainers float32
	numBranches    string
	graphQL        float32
	gitRepo        string
	license        string
}

func newURL(url string) *attribute {
	scoreObject := attribute{url: url}
	return &scoreObject
}

/*func newNpmObject(url string) (*npmObject) {
	npmObj := npmObject{}
	return &npmObj
}
func newGitObject(url string) (*gitObject) {
	gitObj := {url: url}
	return &npmObj
}*/

func installDeps() {
	command := exec.Command("go", "mod", "download")
	err := command.Run()
	if err != nil {
		fmt.Println("Error: ", err)
		os.Exit(1)
	} else {
		fmt.Println("Downloaded 31 dependencies...")
		os.Exit(0)
	}
}

func compile() {
	command := exec.Command("go", "build", "run.go")
	err := command.Run()
	if err != nil {
		fmt.Println("Error: ", err)
		os.Exit(1)
	} else {
		fmt.Println("Build successful...")
		os.Exit(0)
	}
}

func test() {
	fmt.Println("test...")
}

func help() {
	fmt.Println("Unknown command\nUsage: ./run [command] [args]\nCommands:\n\tinstall\t\tInstall dependencies\n\tbuild\t\tBuild the project\n\ttest\t\tRun tests\n\tURL FILE\tScore all URLs in the file")
}

func file(filename string) {
	_, err := os.Stat(filename)
	if os.IsNotExist(err) {
		fmt.Println("Error: ", err)
		os.Exit(1)
	}
	file, err := os.Open(filename)
	if err != nil {
		fmt.Println("Error: ", err)
		os.Exit(1)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var urlCount int = 0
	for scanner.Scan() {
		line := scanner.Text()
		//fmt.Println(line)
		scoreObject := newURL(string(line))

		// call sub function to score each line/project
		if strings.Contains(line, "github.com") {
			var gitObj gitObject
			urlCount += 1
			githubFunc(line, scoreObject, &gitObj, urlCount)
			githubCalcScores(scoreObject, &gitObj)
		} else if strings.Contains(line, "npmjs.com") {
			var npmObj npmObject
			urlCount += 1
			npmjs(line, scoreObject, urlCount, &npmObj)
			npmCalcScores(scoreObject, &npmObj)
		} else {
			fmt.Println("Error: ", line, "is not a valid URL")
		}
	}
}

func npmCalcScores(scoreObject *attribute, npmObj *npmObject) {
	//Calculate for correctness
	if f, err := strconv.ParseFloat(strings.TrimSuffix(npmObj.numCommits, "\n"), 64); err == nil {
		scoreObject.correctness = f
	}
	maxValue := 100.0 //max num of commits
	minValue := 10.0  //min num of commits
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

	//Calculate for responsiveness
	if f, err := strconv.ParseFloat(npmObj.numBranches, 32); err == nil {
		scoreObject.responsiveness = f
	}
	maxResValue := 30.0 //max num of branches
	minResValue := 10.0 //min num of branches
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
	//Calculate for busFactor
	scoreObject.busFactor = npmObj.numMaintainers
	maxBusValue := 50.0 //max num of contributors
	minBusValue := 2.0  //min num of contributors
	maxBusScore := 1.0
	minBusScore := 0.0

	if float64(scoreObject.busFactor) <= minBusValue {
		scoreObject.busFactor = float32(minBusScore)
	} else if float64(scoreObject.busFactor) >= maxBusValue {
		scoreObject.busFactor = float32(maxBusScore)
	} else {
		normalizedValue := (float64(scoreObject.busFactor) - minBusValue) / (maxBusValue - minBusValue)
		scoreObject.busFactor = float32(minBusScore) + float32(normalizedValue)*(float32(maxBusScore)-float32(minBusScore))
	}

	//Calculate rampUp
	//rampup = based on branches the less the easier to rampup
	f, err := strconv.ParseFloat(npmObj.numBranches, 32)
	if err != nil {
		// handle error
	} else {
		scoreObject.rampUp = f
	}
	maxBranchValue := 10.0
	normalizedValue := maxBranchValue / scoreObject.rampUp
	if scoreObject.rampUp <= maxBranchValue {
		scoreObject.rampUp = 1
	}
	if scoreObject.rampUp > maxBranchValue {
		scoreObject.rampUp = normalizedValue
	}

	//avg of all
	scoreObject.netScore = (float64(scoreObject.busFactor) + float64(scoreObject.correctness) + float64(scoreObject.correctness) + float64(scoreObject.rampUp)) / 4
	fmt.Printf("{\"URL\":\"%s\", \"NET_SCORE\":%.1f, \"RAMP_UP_SCORE\":%.1f, \"CORRECTNESS_SCORE\":%.1f, \"BUS_FACTOR_SCORE\": %.1f, \"RESPONSIVE_MAINTAINER_SCORE\":%.1f, \"LICENSE_SCORE\":%d}\n", scoreObject.url, scoreObject.netScore, scoreObject.rampUp, scoreObject.correctness, scoreObject.busFactor, scoreObject.responsiveness, scoreObject.license)
}

func githubCalcScores(scoreObject *attribute, gitObj *gitObject) {
	//Calculate responsiveness
	if f, err := strconv.ParseFloat(gitObj.numCommits, 32); err == nil {
		scoreObject.responsiveness = f
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
	scoreObject.busFactor = float32(gitObj.numPullRequests)
	maxBusValue := 100.0 //max num of PR
	minBusValue := 10.0  //min num of PR
	maxBusScore := 1.0
	minBusScore := 0.0

	if float64(scoreObject.busFactor) <= minBusValue {
		scoreObject.busFactor = float32(minBusScore)
	} else if float64(scoreObject.busFactor) >= maxBusValue {
		scoreObject.busFactor = float32(maxBusScore)
	} else {
		normalizedValue := (float64(scoreObject.busFactor) - minBusValue) / (maxBusValue - minBusValue)
		scoreObject.busFactor = float32(minBusScore) + float32(normalizedValue)*(float32(maxBusScore)-float32(minBusScore))
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
	maxPullValue := 20.0
	normalizedValue := maxPullValue / scoreObject.rampUp
	if scoreObject.rampUp <= maxPullValue {
		scoreObject.rampUp = 1
	}
	if scoreObject.rampUp > maxPullValue {
		scoreObject.rampUp = normalizedValue
	}

	//avg of all
	scoreObject.netScore = (float64(scoreObject.busFactor) + float64(scoreObject.correctness) + float64(scoreObject.correctness) + float64(scoreObject.rampUp)) / 4

	fmt.Printf("{\"URL\":\"%s\", \"NET_SCORE\":%.1f, \"RAMP_UP_SCORE\":%.1f, \"CORRECTNESS_SCORE\":%.1f, \"BUS_FACTOR_SCORE\":%.1f, \"RESPONSIVE_MAINTAINER_SCORE\":%.1f, \"LICENSE_SCORE\":%d}\n", scoreObject.url, scoreObject.netScore, scoreObject.rampUp, scoreObject.correctness, scoreObject.busFactor, scoreObject.responsiveness, scoreObject.license)
}

func githubFunc(url string, scoreObject *attribute, gitObj *gitObject, count int) {
	split := strings.Split(url, "/")
	owner := split[len(split)-2]
	repo := split[len(split)-1]
	//print("Owner: ", owner, " Repo: ", repo, "\n")
	gitObj.numCommits = strings.TrimSuffix(string(githubSource(url, count)), "\n")

	var fullRepo string = owner + "/" + repo
	gitObj.numPullRequests = githubPullReq(fullRepo)
	if githubLicense(fullRepo) == true {
		scoreObject.license = 1
	} else {
		scoreObject.license = 0
	}
	//println("git license: ", scoreObject.license)
	//value3 := githubGraphQL
	//intConv, _ := strconv.Atoi(string(value))
	//gitHubGraphQL(repo, owner)
	//fmt.Println("git num commits: ", gitObj.numCommits)
	//fmt.Println("git num PR: ", gitObj.numPullRequests)

	//remove recently created directory after info is pulled
	command2 := exec.Command("rm", "-rf", "clonedir"+strconv.Itoa(count))

	if err := command2.Run(); err != nil {
		fmt.Println(err.Error())
		return
	}
	/*gitObj.issues, gitObj.releases, gitObj.stargazers, gitObj.license = gitHubGraphQL(repo, owner)
	fmt.Println("git issues: ", gitObj.issues)
	fmt.Println("git releases: ", gitObj.releases)
	fmt.Println("git stargazers: ", gitObj.stargazers)
	fmt.Println("git license: ", gitObj.license)*/
	//scoreObject.license = licenseCompatability(gitO)

	/*func githubFunc(url string) {
		split := strings.Split(url, "/")
		owner := split[len(split)-2]
		repo := split[len(split)-1]
		// print("Owner: ", owner, " Repo: ", repo, "\n")
		gitHubGraphQL(repo, owner)
		gitHubRestAPI(repo, owner)


	}*/
}

func npmjs(url string, scoreObject *attribute, count int, npmObj *npmObject) {
	split := strings.Split(url, "/")
	packageName := split[len(split)-1]
	//print("Package: ", packageName, "\n")
	npmRestAPI(packageName, scoreObject, npmObj)
	npmSource(npmObj, count)
	//fmt.Println(npmObj.gitRepo)
	/*fmt.Println("npm commits: ", npmObj.numCommits)
	fmt.Println("npm maintainers: ", npmObj.numMaintainers)*/
	if licenseCompatability(npmObj.license) == true {
		scoreObject.license = 1
	} else {
		scoreObject.license = 0
	}
	localBranchCount(count, npmObj)
	//fmt.Println("numBranches npm ", npmObj.numBranches)
	//calc score/output json
	//remove recently created directory after info is pulled
	command2 := exec.Command("rm", "-rf", "clonedir"+strconv.Itoa(count))

	if err := command2.Run(); err != nil {
		fmt.Println(err.Error())
		return
	}
}

func githubSource(url string, count int) (output []byte) {

	//call python script that clones repo and pull number of commits
	command := exec.Command("python3", "cloner.py", url, strconv.Itoa(count))
	output, err := command.Output()

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	return output

}

func npmSource(npmObj *npmObject, count int) {

	//call python script that clones repo and pull number of commits
	command := exec.Command("python3", "cloner.py", "https://github.com/"+npmObj.gitRepo, strconv.Itoa(count))
	output, err := command.Output()

	if err != nil {
		fmt.Println(err.Error())
		return
	}
	/*b := binary.BigEndian.Uint32(output)
	float := math.Float32frombits(b)*/
	npmObj.numCommits = strings.TrimSuffix(string(output), "\n")

	/*//remove recently created directory after info is pulled
	command2 := exec.Command("rm", "-rf", "clonedir"+strconv.Itoa(count))

	if err := command2.Run(); err != nil {
		fmt.Println(err.Error())
		return
	}*/

}

type PullRequests struct {
	TotalCount int `json:"total_count"`
}

func githubPullReq(repoName string) (value2 int) {
	req, _ := http.NewRequest("GET", "https://api.github.com/search/issues?q=is:pr+repo:"+repoName, nil)
	req.Header.Set("Accept", "application/vnd.github+json")

	client := &http.Client{}
	res, err := client.Do(req)

	if err != nil {
		fmt.Print(err.Error())
		os.Exit(1)
	}

	defer res.Body.Close()

	var pullRequests PullRequests
	json.NewDecoder(res.Body).Decode(&pullRequests)

	return pullRequests.TotalCount
}

type LicenseType struct {
	//License_Type string `json:"license"`
	LicenseType struct {
		LicenseName string `json:"spdx_id"`
	} `json:"license"`
}

func githubLicense(repoName string) bool {
	req, _ := http.NewRequest("GET", "https://api.github.com/repos/"+repoName+"/license", nil)
	req.Header.Set("Accept", "application/vnd.github+json")

	client := &http.Client{}
	res, err := client.Do(req)

	if err != nil {
		fmt.Print(err.Error())
		os.Exit(1)
	}

	defer res.Body.Close()

	var license LicenseType
	json.NewDecoder(res.Body).Decode(&license)

	return licenseCompatability(license.LicenseType.LicenseName)
}

func npmRestAPI(packageName string, scoreObject *attribute, npmObj *npmObject) {

	//append packageName to the api url and send request
	url := "https://registry.npmjs.org/" + packageName
	response, err := http.Get(url)

	if err != nil {
		fmt.Print(err.Error())
		os.Exit(1)
	}

	//read api response into responseData
	responseData, err := ioutil.ReadAll(response.Body)

	if err != nil {
		fmt.Print(err.Error())
		os.Exit(1)
		//log.Fatal(err)
	}

	//creates an object to store json data
	contributors := make(map[string]interface{})

	//Unmarshalls json data into contributors object and returns err
	err = json.Unmarshal(responseData, &contributors)

	if err != nil {
		fmt.Print("failed to decode api response: ", err)
		os.Exit(1)
	}

	//stores list of maintainers into array object
	array := contributors["maintainers"].([]interface{})
	numContributors := len(array) //number of active maintainers for package
	npmObj.numMaintainers = float32(numContributors)
	license := contributors["license"]
	npmObj.license = license.(string)

	//fmt.Print("number of contributors: ", scoreObject.responsiveness)
	//fmt.Print("\nlicense: ", contributors["license"].(string))
	//fmt.Print("\ngithub url: ", contributors["repository"].(map[string]interface{})["url"])
	split := strings.Split(contributors["repository"].(map[string]interface{})["url"].(string), "/")
	owner := split[len(split)-2]
	repo := split[len(split)-1]
	//fmt.Println("Owner: ", owner, " Repo: ", repo)
	//var npmGitString string = owner + "/" + repo
	npmObj.gitRepo = owner + "/" + repo
	//fmt.Print("\ngithub url: ", contributors["repository"].(map[string]interface{})["url"])
	//fmt.Print("\n")
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

/*func gitHubRestAPI(repo string, owner string) {
	apiKey := os.Getenv("GITHUB_API_KEY")
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: apiKey},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)
	_, _, err := client.Repositories.Get(ctx, owner, repo)
	if err != nil {
		fmt.Println(err)
		return
	}

	pullRequests, _, err := client.PullRequests.List(ctx, owner, repo, nil)
	if err != nil {
		fmt.Println("Error fetching pull requests:", err)
		return
	}

	totalPullRequests := len(pullRequests)
	fmt.Printf("Total pull requests: %d\n", totalPullRequests)

	totalIssues, _, err := client.Issues.ListByRepo(ctx, owner, repo, nil)
	fmt.Printf("Total issues: %d\n", len(totalIssues))
}*/

func gitHubGraphQL(repoName string, owner string) (issueCount int, releaseCount int, starCount int, license string) {
	client := graphql.NewClient("https://api.github.com/graphql")
	req := graphql.NewRequest(`
	query ($repoName: String!, $owner: String!) {
		repository(name: $repoName, owner: $owner) {
		  licenseInfo {
			name
		  }
		  pullRequests {
			totalCount
		  }
		  commitComments {
			totalCount
		  }
		  releases {
			totalCount
		  }
		  stargazerCount
		  defaultBranchRef {
			name
			target {
			  ... on Commit {
				id
				history(first: 0) {
				  totalCount
				}
			  }
			}
		  }
		  issues(states: CLOSED) {
			totalCount
		  }
		}
	  }
	`)
	req.Var("repoName", repoName)
	req.Var("owner", owner)
	apiKey := os.Getenv("GITHUB_API_KEY")
	//apiKey := os.Getenv("GITHUB_TOKEN")
	req.Header.Set("Authorization", "Bearer "+apiKey)
	var response map[string]interface{}
	err := client.Run(context.Background(), req, &response)
	if err != nil {
		panic(err)
	}
	repository := response["repository"].(map[string]interface{})
	if repository["licenseInfo"] == nil {
		npmLicense(repoName)
	} else if repository["licenseInfo"].(map[string]interface{})["name"] == "Other" {
		npmLicense(repoName)
	} else {
		licenseInfo := repository["licenseInfo"].(map[string]interface{})["name"].(string)
		fmt.Println("License: ", licenseInfo)
	}
	numIssues := int(repository["issues"].(map[string]interface{})["totalCount"].(float64))
	//  commitCount := int(repository["defaultBranchRef"].(map[string]interface{})["target"].(map[string]interface{})["history"].(map[string]interface{})["totalCount"].(float64))
	// pullRequests := int(repository["pullRequests"].(map[string]interface{})["totalCount"].(float64))
	releases := int(repository["releases"].(map[string]interface{})["totalCount"].(float64))
	stargazerCount := int(repository["stargazerCount"].(float64))
	return numIssues, releases, stargazerCount, license
}

func npmLicense(packageName string) {
	if strings.HasSuffix(packageName, "_npm") {
		packageName = strings.TrimSuffix(packageName, "_npm")
	}
	url := "https://registry.npmjs.org/" + packageName
	response, err := http.Get(url)
	if err != nil {
		fmt.Print(err.Error())
		os.Exit(1)
	}
	responseData, err := ioutil.ReadAll(response.Body)

	if err != nil {
		fmt.Print(err.Error())
		os.Exit(1)
	}
	array := make(map[string]interface{})
	err = json.Unmarshal(responseData, &array)

	if err != nil {
		fmt.Print("failed to decode api response: ", err)
		return
	}
	//fmt.Print("license: ", array["license"], "\n")
}

func localBranchCount(count int, npmObj *npmObject) {
	FolderLoc, err := filepath.Abs("clonedir" + strconv.Itoa(count))
	if err != nil {
		fmt.Println("Filepath for folder not found", err)
		return
	}
	// Git cmd for list of all repos
	out, err := exec.Command("git", "-C", FolderLoc, "branch", "-a").Output()
	if err != nil {
		fmt.Println("Error running git command:", err)
		return
	}

	// Split output onto new lines and return (len - extra versions of origin/head)
	branches := strings.Split(string(out), "\n")
	//fmt.Printf("HERE HERE, %d", len(branches)-3)
	npmObj.numBranches = strconv.Itoa(len(branches) - 3)
}

func main() {
	args := os.Args[1:]
	if args[0] == "install" {
		installDeps()
	} else if args[0] == "build" {
		compile()
	} else if args[0] == "test" {
		test()
	} else if filepath.Ext(args[0]) == ".txt" {
		file(args[0])
	} else {
		help()
		os.Exit(1)
	}
}
