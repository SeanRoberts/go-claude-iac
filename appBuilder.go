package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Content struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type Response struct {
	Content []Content `json:"content"`
}

const (
	apiUrl    = "https://api.anthropic.com/v1/messages"
	model     = "claude-3-opus-20240229"
	maxTokens = 4096
)

type AppBuilder struct {
	AppName     string
	Description string
	UseFixture  bool
}

func (a *AppBuilder) GetFileContent() (string, error) {
	var responseText string

	if a.UseFixture {
		body, err := os.ReadFile("fixtures/" + a.AppName + ".txt")
		if err != nil {
			fmt.Println(err)
			return "", err
		}
		return cleanResponse(string(body)), nil
	}

	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
		return "", err
	}

	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		fmt.Println("API_KEY not set in .env file")
		return "", err
	}

	prompt := fmt.Sprintf(`You are a devops wizard who has been asked to help devise Terraform code to deploy a new application to AWS. You should use Terraform best practices and the latest version of Terraform. You should make recommendations based on your
		knowledge of Terraform and the best practices for deploying applications. For example, if the desired application is described as a Rails application you should know to use a combination of ECS, Fargate, and RDS to deploy the application. You are helping application
		developers who are not well versed in devops so it is up to you to be the expert in the room. It is also crucial that you respond only with Terraform code and not with advice or explanations. Your answer will be written directly into files and the project will fail if you reply with anything other than valid
		Terraform code.

		The name of the application is: %s
		The description of the application is as follows:\n\n%s`, a.AppName, a.Description)

	data := map[string]interface{}{
		"model":      model,
		"max_tokens": maxTokens,
		"messages": []map[string]string{
			{
				"role":    "user",
				"content": prompt,
			},
		},
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		fmt.Printf("Error marshalling data: %s\n\ndata was:\n%s", err, data)
		return "", err
	}

	req, err := http.NewRequest("POST", apiUrl, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("Error creating request: %s\n", err)
		return "", err
	}

	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error sending request: %s\n", err)
		return "", err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response body: %s\n", err)
		return "", err
	}

	var response Response
	err = json.Unmarshal(body, &response)
	if err != nil {
		fmt.Printf("Error unmarshalling response: %s\n", err)
		fmt.Printf("Response was:\n%s\n", body)
		return "", err
	}

	responseText = response.Content[0].Text

	return cleanResponse(responseText), nil
}

func cleanResponse(response string) string {
	response = strings.ReplaceAll(response, "```hcl\n", "")
	response = strings.ReplaceAll(response, "```", "")
	response = strings.TrimSpace(response)
	return response
}
