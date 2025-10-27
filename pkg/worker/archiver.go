package worker

import (
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/c_yamada/jira_cloud_bulk_archive/internal/jira"
)

// ArchiveResult represents the result of archiving an issue
type ArchiveResult struct {
	IssueKey string
	Success  bool
	Error    error
}

// Archiver handles concurrent archiving of JIRA issues
type Archiver struct {
	client     *jira.Client
	maxWorkers int
}

// NewArchiver creates a new Archiver
func NewArchiver(client *jira.Client, maxWorkers int) *Archiver {
	return &Archiver{
		client:     client,
		maxWorkers: maxWorkers,
	}
}

// ArchiveIssues archives multiple issues concurrently using goroutines
func (a *Archiver) ArchiveIssues(issues []jira.Issue) []ArchiveResult {
	totalIssues := len(issues)
	if totalIssues == 0 {
		log.Println("No issues to archive")
		return []ArchiveResult{}
	}

	log.Printf("Starting to archive %d issues with %d workers\n", totalIssues, a.maxWorkers)

	// Create channels for job distribution and result collection
	jobs := make(chan jira.Issue, totalIssues)
	results := make(chan ArchiveResult, totalIssues)

	// Start worker goroutines
	var wg sync.WaitGroup
	for i := 0; i < a.maxWorkers; i++ {
		wg.Add(1)
		go a.worker(i+1, &wg, jobs, results)
	}

	// Send jobs to workers
	for _, issue := range issues {
		jobs <- issue
	}
	close(jobs)

	// Wait for all workers to finish
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	var archiveResults []ArchiveResult
	for result := range results {
		archiveResults = append(archiveResults, result)
	}

	return archiveResults
}

// worker is a goroutine that processes archive jobs
func (a *Archiver) worker(id int, wg *sync.WaitGroup, jobs <-chan jira.Issue, results chan<- ArchiveResult) {
	defer wg.Done()

	for issue := range jobs {
		log.Printf("[Worker %d] Archiving issue: %s (%s)\n", id, issue.Key, issue.Fields.Summary)

		err := a.client.ArchiveIssue(issue.Key)
		result := ArchiveResult{
			IssueKey: issue.Key,
			Success:  err == nil,
			Error:    err,
		}

		if err != nil {
			log.Printf("[Worker %d] Failed to archive %s: %v\n", id, issue.Key, err)
		} else {
			log.Printf("[Worker %d] Successfully archived %s\n", id, issue.Key)
		}

		results <- result
	}
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
