package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"
)

const useFixture = false

func main() {

	appName, description, err := getAppInfoFromUser()
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Print("\r\n\r\n")

	loaderFrames := [12]string{"_", "_", "_", "-", "`", "`", "'", "´", "-", "_", "_", "_"}
	frameCount := len(loaderFrames)
	frameIdx := 0
	done := make(chan bool)
	ticker := time.NewTicker(70 * time.Millisecond)
	go func() {
		for {
			select {
			case <-done:
				break
			case <-ticker.C:
				fmt.Printf("\rGenerating Terraform code %s", loaderFrames[frameIdx])
				frameIdx++
				if frameIdx == frameCount {
					frameIdx = 0
				}
			}
		}
	}()

	// Generate the Terraform code
	builder := &AppBuilder{
		AppName:     appName,
		Description: description,
		UseFixture:  useFixture,
	}
	terraformCode, err := builder.GetFileContent()

	// Stop the loader animation, wait for the goroutine to finish, and then
	// clear it from the terminal
	ticker.Stop()
	done <- true
	fmt.Printf("\033[1A\033[K")

	if err != nil {
		fmt.Printf("\n\rError: %s\n", err)
		return
	}

	appPath := "./apps/" + appName

	fmt.Printf("\r\nDone! Writing files to %s...\n", appPath)

	// Create the directory for the project
	err = os.MkdirAll(appPath, 0755)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Create the main.tf file
	mainTfFile, err := os.Create(appPath + "/main.tf")
	if err != nil {
		fmt.Println(err)
		return
	}

	// Write the Terraform code to the main.tf file
	_, err = mainTfFile.WriteString(terraformCode)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("Your Terraform code has been written to %s/main.tf\n", appPath)
}

func getAppInfoFromUser() (string, string, error) {
	appNameIsValid := false
	var appName string
	var err error

	for !appNameIsValid {
		appName, err = promptForInput("Please enter the name of your application: ")
		if err != nil {
			fmt.Println(err)
			return "", "", err
		}

		// Replace all spaces, slashes, and periods with hyphens
		re := regexp.MustCompile(`[ /\.]+`)
		appName = re.ReplaceAllString(appName, "-")

		// Check if a directory with that name already exists
		_, err = os.Stat(appName)
		if os.IsNotExist(err) {
			appNameIsValid = true
		} else {
			fmt.Println("A directory with that name already exists. Please choose a different name.")
		}
	}

	description, err := promptForInput("\nPlease enter a description of your application: ")
	if err != nil {
		fmt.Println(err)
		return "", "", err
	}

	return appName, description, nil
}

func promptForInput(question string) (string, error) {
	input := ""
	var err error
	for input == "" {
		fmt.Println(question)
		reader := bufio.NewReader(os.Stdin)
		input, err = reader.ReadString('\n')
		if err != nil {
			return "", err // Return the error to the caller
		}

		// Trim the newline character from the input
		input = strings.TrimSpace(input)
	}
	return input, nil
}

func showLoader(text string, done chan bool, wg *sync.WaitGroup) {
	loaderFrames := [12]string{"_", "_", "_", "-", "`", "`", "'", "´", "-", "_", "_", "_"}
	for {
		defer wg.Done()
		select {
		case <-done:
			return
		default:
			for _, frame := range loaderFrames {
				fmt.Printf("\r%s %s", text, frame)
				time.Sleep(70 * time.Millisecond)
			}
		}
	}
}
