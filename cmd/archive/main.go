package main

import (
	"log"
	"os"

	"github.com/c_yamada/jira_cloud_bulk_archive/internal/config"
	"github.com/c_yamada/jira_cloud_bulk_archive/internal/jira"
	"github.com/c_yamada/jira_cloud_bulk_archive/pkg/worker"
	"github.com/joho/godotenv"
)

func main() {
	// Configure logger
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	log.Println("Starting JIRA Cloud Bulk Archive Tool")

	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	} else {
		log.Println("Loaded configuration from .env file")
	}

	// Load configuration from environment variables
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	log.Printf("Configuration loaded successfully")
	log.Printf("JIRA Base URL: %s", cfg.JiraBaseURL)
	log.Printf("Project Key: %s", cfg.JiraProjectKey)
	log.Printf("Archive Label: %s", cfg.ArchiveLabel)
	log.Printf("Max Workers: %d", cfg.MaxWorkers)

	// Create JIRA client
	client := jira.NewClient(cfg.JiraBaseURL, cfg.JiraEmail, cfg.JiraAPIToken)

	// Search for issues with the archive label
	log.Printf("Searching for issues with label '%s' in project '%s'...", cfg.ArchiveLabel, cfg.JiraProjectKey)
	issues, err := client.GetAllIssuesByLabel(cfg.JiraProjectKey, cfg.ArchiveLabel)
	if err != nil {
		log.Fatalf("Failed to search for issues: %v", err)
	}

	log.Printf("Found %d issues to archive", len(issues))

	if len(issues) == 0 {
		log.Println("No issues to archive. Exiting.")
		os.Exit(0)
	}

	// Create archiver and process issues concurrently
	archiver := worker.NewArchiver(client, cfg.MaxWorkers)
	results := archiver.ArchiveIssues(issues)

	// Print summary
	worker.PrintSummary(results)

	// Exit with error code if any failures occurred
	hasFailures := false
	for _, result := range results {
		if !result.Success {
			hasFailures = true
			break
		}
	}

	if hasFailures {
		log.Println("Completed with errors")
		os.Exit(1)
	}

	log.Println("All issues archived successfully!")
}
