package main

import (
	"fmt"
	"github.com/twinj/uuid"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
)

func guuid() string {
	return uuid.Formatter(uuid.NewV4(), uuid.Clean)
}

func command(args ...string) string {
	cmd := exec.Command(args[0], args[1:]...)
	output, _ := cmd.CombinedOutput()

	return strings.TrimSpace(string(output))
}

func info(msg string) {
	fmt.Printf("%s\n", msg)
}

func warning(msg string) {
	fmt.Printf("%s\n", msg)
}

func gitClean() bool {
	return command("git", "status", "--porcelain") == ""
}

func fileContains(file string, str string) bool {
	return fileExists(file) && strings.Contains(readFile(file), str)
}

func fileExists(file string) bool {
	_, err := os.Stat(file)

	return !os.IsNotExist(err)
}

func readFile(fileName string) string {
	content, _ := ioutil.ReadFile(fileName)
	return string(content)
}

func createFile(fileName string, content string) {
	f, _ := os.Create(fileName)
	f.Close()

	appendToFile(fileName, content)
}

func appendToFile(fileName string, content string) {
	f, _ := os.OpenFile(fileName, os.O_RDWR|os.O_APPEND, 0666)

	f.WriteString(content + "\n")
	f.Sync()

	f.Close()
}

func main() {
	if command("heroku", "config:get", "DEVSYNC") == "" {
		info("Setting DEVSYNC configuration key")
		command("heroku", "config:set", "DEVSYNC="+guuid())
	}

	if !gitClean() {
		warning("We will be adding a buildpack to your pipeline, in a commit you can review just after. You have to have a clean `git status` output to continue.")
		return
	}

	if !fileContains("./.buildpacks", "https://github.com/leonkenneth/heroku-buildpack-dev.git") {
		info("Adding buildpack and preparing a commit to push for you")

		if !fileExists("./.buildpacks") {
			// Get current buildpack
			currentBuildpack := command("heroku", "config:get", "BUILDPACK_URL")
			// Create .buildpacks with it
			info(currentBuildpack)
			createFile("./.buildpacks", currentBuildpack)
			// Use multipack buildpack
			command("heroku", "config:set", "BUILDPACK_URL=https://github.com/ddollar/heroku-buildpack-multi.git")
		}

		// Add our buildpack
		appendToFile("./.buildpacks", "https://github.com/leonkenneth/heroku-buildpack-dev.git")

		command("git", "add", ".buildpacks")
		command("git", "commit", "-m", "\"Add herodev\"")
	}

	info("All set.\nIf not done already, you can review our last commit with `git show -p`.\nWhen you're ready, `git push heroku master` to get the magic started.\nThen run `heroku dev:watch` at the beginning of your coding session.")
}
