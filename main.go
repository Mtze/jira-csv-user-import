package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"strings"

	jira "github.com/andygrunwald/go-jira"
	"github.com/caarlos0/env"
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
)

var (
	UserPrefix = ""
	authConfig Auth
)

type Auth struct {
	URL      string `env:"JIRA_URL"`
	Username string `env:"JIRA_USERNAME"`
	Password string `env:"JIRA_PASSWORD"`
}

type User struct {
	FirstName string
	LastName  string
	Email     string
	Username  string
	Password  string
}

func main() {
	if is_debug := os.Getenv("DEBUG"); is_debug == "true" {
		log.SetLevel(log.DebugLevel)
		log.Warn("DEBUG MODE ENABLED")
	}

	path, err := getFilenameFromArgs()
	if err != nil {
		log.WithError(err).Fatal("Error getting file")
	}

	LoadConfig(&authConfig)

	users, err := readCSV(path)
	if err != nil {
		log.WithError(err).Fatal("Error reading CSV file")
	}

	log.Debug("Users: ", users)

	log.Debug("Creating Jira client for ", authConfig.URL)
	tp := jira.BasicAuthTransport{
		Username: authConfig.Username,
		Password: authConfig.Password,
	}
	jiraClient, err := jira.NewClient(tp.Client(), authConfig.URL)
	if err != nil {
		log.WithError(err).Fatal("Error creating Jira client")
	}
	log.Info("Jira client created")

	for _, user := range users {
		log.Info("Creating user ", user.Username)
		jiraUser := jira.User{
			Name:         UserPrefix + user.Username,
			Password:     user.Password,
			EmailAddress: user.Email,
			DisplayName:  user.Username,
		}

		jiraClient.User.Create(&jiraUser)
	}
}

func readCSV(filename string) ([]User, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Comma = ';'

	lines, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	var users []User
	for _, line := range lines[1:] {
		users = append(users, User{
			FirstName: line[0],
			LastName:  line[1],
			Email:     line[2],
			Username:  line[3],
			Password:  strings.TrimSpace(line[4]),
		})
	}

	return users, nil
}

func getFilenameFromArgs() (string, error) {
	if len(os.Args) < 2 {
		return "", fmt.Errorf("no filename provided")
	}

	filename := os.Args[1]

	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return "", fmt.Errorf("file %s does not exist", filename)
	}

	return filename, nil
}

func LoadConfig(cfg interface{}) {
	err := godotenv.Load()
	if err != nil {
		log.WithError(err).Warn("Error loading .env file")
	}

	err = env.Parse(cfg)
	if err != nil {
		log.WithError(err).Fatal("Error parsing environment variables")
	}

	log.Debug("Config loaded: ", cfg)
}
