package worker

import (
	"fmt"
	"log"
	"strings"

	"github.com/c_yamada/jira_cloud_bulk_archive/internal/jira"
)

// ArchiveResult represents the result of archiving an issue
type ArchiveResult struct {
	IssueKey string
	Success  bool
	Error    error
}

// Archiver handles bulk archiving of JIRA issues
type Archiver struct {
	client    *jira.Client
	batchSize int
}

// NewArchiver creates a new Archiver
func NewArchiver(client *jira.Client, _ int) *Archiver {
	return &Archiver{
		client:    client,
		batchSize: 1000, // Archive up to 1000 issues per batch
	}
}

// ArchiveIssues archives multiple issues using bulk API
func (a *Archiver) ArchiveIssues(issues []jira.Issue) []ArchiveResult {
	totalIssues := len(issues)
	if totalIssues == 0 {
		log.Println("No issues to archive")
		return []ArchiveResult{}
	}

	log.Printf("Starting to archive %d issues using bulk API (batch size: %d)\n", totalIssues, a.batchSize)

	// Split issues into batches
	batches := a.createBatches(issues)
	log.Printf("Created %d batches\n", len(batches))

	// Process each batch sequentially
	var allResults []ArchiveResult
	for batchNum, batch := range batches {
		log.Printf("Processing batch %d/%d (%d issues)\n", batchNum+1, len(batches), len(batch))
		batchResults := a.processBatch(batch)
		allResults = append(allResults, batchResults...)
	}

	return allResults
}

// createBatches splits issues into batches of configured size
func (a *Archiver) createBatches(issues []jira.Issue) [][]jira.Issue {
	var batches [][]jira.Issue
	for i := 0; i < len(issues); i += a.batchSize {
		end := i + a.batchSize
		if end > len(issues) {
			end = len(issues)
		}
		batches = append(batches, issues[i:end])
	}
	return batches
}

// processBatch processes a single batch of issues using the bulk archive API
func (a *Archiver) processBatch(batch []jira.Issue) []ArchiveResult {
	batchSize := len(batch)
	issueKeys := make([]string, batchSize)

	for i, issue := range batch {
		issueKeys[i] = issue.Key
		log.Printf("Batch item %d: Key=%s, ID=%s\n", i, issue.Key, issue.ID)
	}

	log.Printf("Archiving batch of %d issues\n", batchSize)

	// Call bulk archive API
	resp, err := a.client.ArchiveIssues(issueKeys)

	// Process results
	batchResults := make([]ArchiveResult, batchSize)
	for i, issue := range batch {
		if err != nil {
			// Entire batch failed
			batchResults[i] = ArchiveResult{
				IssueKey: issue.Key,
				Success:  false,
				Error:    err,
			}
			log.Printf("Failed to archive %s: %v\n", issue.Key, err)
		} else if resp != nil && resp.Errors != nil && resp.Errors[issue.Key] != "" {
			// Individual issue failed
			batchResults[i] = ArchiveResult{
				IssueKey: issue.Key,
				Success:  false,
				Error:    fmt.Errorf("%s", resp.Errors[issue.Key]),
			}
			log.Printf("Failed to archive %s: %s\n", issue.Key, resp.Errors[issue.Key])
		} else {
			// Success
			batchResults[i] = ArchiveResult{
				IssueKey: issue.Key,
				Success:  true,
				Error:    nil,
			}
			log.Printf("Successfully archived %s\n", issue.Key)
		}
	}

	return batchResults
}

// PrintSummary prints a summary of the archive operation
func PrintSummary(results []ArchiveResult) {
	total := len(results)
	successful := 0
	failed := 0

	fmt.Println("\n" + strings.Repeat("=", 50))
	fmt.Println("Archive Summary")
	fmt.Println(strings.Repeat("=", 50))

	for _, result := range results {
		if result.Success {
			successful++
		} else {
			failed++
			fmt.Printf("Failed: %s - %v\n", result.IssueKey, result.Error)
		}
	}

	fmt.Printf("\nTotal issues: %d\n", total)
	fmt.Printf("Successfully archived: %d\n", successful)
	fmt.Printf("Failed: %d\n", failed)
	fmt.Println(strings.Repeat("=", 50))
}
