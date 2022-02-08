package main

import (
	"context"
	"regexp"
	"strings"

	"github.com/google/go-github/v41/github"
)

type checkResult struct {
	// Reviewed indicates that *any* review has been made on the PR.
	Reviewed bool
	// TestPlan is the content provided after the acceptance checklist checkbox.
	TestPlan string
	// Error indicating any issue that might have occured during the check.
	Error error
}

func (r checkResult) HasTestPlan() bool {
	return r.TestPlan != ""
}

var markdownCommentRegexp = regexp.MustCompile("<!--((.|\n)*?)-->(\n)*")

func checkTestPlan(ctx context.Context, ghc *github.Client, payload *EventPayload) checkResult {
	pr := payload.PullRequest

	// Whether or not this PR was reviewed can be inferred from payload, but an approval
	// might not have any comments so we need to double-check through the GitHub API
	var err error
	reviewed := pr.ReviewComments > 0
	if !reviewed {
		repoParts := strings.Split(payload.Repository.FullName, "/")
		var reviews []*github.PullRequestReview
		// Continue, but return err later
		reviews, _, err = ghc.PullRequests.ListReviews(ctx, repoParts[0], repoParts[1], payload.PullRequest.Number, &github.ListOptions{})
		reviewed = len(reviews) > 0
	}

	// Parse test plan data from body
	sections := strings.Split(pr.Body, "# Test plan")
	if len(sections) < 2 {
		return checkResult{
			Reviewed: reviewed,
			Error:    err,
		}
	}
	testPlanSection := sections[1]
	testPlanRawLines := strings.Split(testPlanSection, "\n")
	var testPlanLines []string
	for _, l := range testPlanRawLines {
		line := strings.TrimSpace(l)
		testPlanLines = append(testPlanLines, line)
	}

	// Merge into single string
	testPlan := strings.Join(testPlanLines, "\n")
	// Remove comments
	testPlan = markdownCommentRegexp.ReplaceAllString(testPlan, "")
	// Remove whitespace
	testPlan = strings.TrimSpace(testPlan)
	return checkResult{
		Reviewed: reviewed,
		TestPlan: testPlan,
		Error:    err,
	}
}
