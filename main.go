package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/drone/drone-template-lib/template"
	"net/http"
	"os"
	"strconv"
	"strings"
)

// Helper function to convert string to number or boolean if possible
func convertToCorrectType(s string) interface{} {
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return f // Return as float if possible
	} else if i, err := strconv.Atoi(s); err == nil {
		return i // Return as int if possible
	} else if b, err := strconv.ParseBool(s); err == nil {
		return b // Return as bool if possible
	}
	return s // Return as string
}

func main() {
	fmt.Println("Starting drone to talk message!")

	// Populate all env variables
	envData := make(map[string]interface{})
	for _, env := range os.Environ() {
		pair := strings.SplitN(env, "=", 2)
		if len(pair) == 2 {
			envData[pair[0]] = convertToCorrectType(pair[1])
		}
	}

	// Read environment variables
	serverURL := os.Getenv("PLUGIN_NEXTCLOUD_SERVER_URL")
	secret := os.Getenv("PLUGIN_BOT_SECRET")
	message := os.Getenv("PLUGIN_MESSAGE")
	roomId := os.Getenv("PLUGIN_ROOM_ID")

	// Check for missing parameters and specify which one
	if serverURL == "" {
		fmt.Println("Error: Missing required parameter NEXTCLOUD_SERVER_URL")
		os.Exit(1)
	}

	if secret == "" {
		fmt.Println("Error: Missing required parameter BOT_SECRET")
		os.Exit(1)
	}

	if roomId == "" {
		fmt.Println("Error: Missing required parameter ROOM_ID")
		os.Exit(1)
	}

	if message == "" {
		// Read Drone environment variables
		buildStatus := os.Getenv("DRONE_BUILD_STATUS")
		branchName := os.Getenv("DRONE_BRANCH")
		branchLink := os.Getenv("DRONE_REPO_LINK")
		commit := strings.TrimSpace(os.Getenv("DRONE_COMMIT_MESSAGE"))
		author := os.Getenv("DRONE_COMMIT_AUTHOR")
		sha := os.Getenv("DRONE_COMMIT_SHA")
		commitLink := os.Getenv("DRONE_COMMIT_LINK")
		link := os.Getenv("DRONE_BUILD_LINK")

		// Set the status icon based on build status
		status := "❌ **Failed**"
		if buildStatus == "success" {
			status = "✅ **Success**"
		}

		// Create the default message
		message = fmt.Sprintf(`
		Status: %s
		Branch: [%s](%s)
		Commit: %s
		Author: %s
		Hash: [%s](%s)
		[View full log here](%s)`, status, branchName, branchLink, commit, author, sha, commitLink, link)

	} else {
		// Render the template with the provided data
		tmpl, err := template.RenderTrim(message, envData)
		if err != nil {
			fmt.Println("Error rendering template:", err)
			os.Exit(1)
		}

		message = tmpl
	}

	// Prepare the request
	url := fmt.Sprintf("%s/ocs/v2.php/apps/spreed/api/v1/bot/%s/message", serverURL, roomId)
	values := map[string]string{"message": message}

	jsonData, err := json.Marshal(values)
	if err != nil {
		fmt.Printf("Error marshalling JSON: %v\n", err)
		os.Exit(1)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		os.Exit(1)
	}

	// Create random 64 characters
	random := hex.EncodeToString(make([]byte, 32))

	// Create a new HMAC
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(random + message))

	// Get signature from the HMAC
	signature := hex.EncodeToString(h.Sum(nil))

	req.Header.Add("OCS-APIRequest", "true")
	req.Header.Add("X-Nextcloud-Talk-Bot-Random", random)
	req.Header.Add("X-Nextcloud-Talk-Bot-Signature", signature)
	req.Header.Add("Content-Type", "application/json")

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error sending request: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	fmt.Printf("Message sent successfully with status code: %d\n", resp.StatusCode)
}
