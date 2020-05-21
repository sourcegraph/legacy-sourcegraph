package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
)

type Platform string

const (
	PlatformGCP Platform = "gcp"
	PlatformAWS Platform = "aws"
)

type Resource struct {
	Platform   Platform
	Identifier string
	Type       string
	Location   string
	Owner      string
	Meta       map[string]interface{}
}

func (r *Resource) toSlackBlock() (slackBlock, error) {
	meta, err := json.Marshal(r.Meta)
	if err != nil {
		return nil, fmt.Errorf("failed to convert resource to Slack block: %w", err)
	}
	return slackBlock{
		"type": "context",
		"elements": []slackText{
			{
				Type: slackTextMarkdown,
				Text: fmt.Sprintf("*Platform*: %s", r.Platform),
			},
			{
				Type: slackTextMarkdown,
				Text: fmt.Sprintf("*Type*: `%s`", r.Type),
			},
			{
				Type: slackTextMarkdown,
				Text: fmt.Sprintf("*ID*: `%s`", r.Identifier),
			},
			{
				Type: slackTextMarkdown,
				Text: fmt.Sprintf("*Location*: %s", r.Location),
			},
			{
				Type: slackTextMarkdown,
				Text: fmt.Sprintf("*Owner*: %s", r.Owner),
			},
			{
				Type: slackTextMarkdown,
				Text: fmt.Sprintf("*Meta*: `%s`", string(meta)),
			},
		},
	}, nil
}

func hasPrefix(value string, prefixes []string) bool {
	for _, prefix := range awsRegionPrefixes {
		if strings.HasPrefix(value, prefix) {
			return true
		}
	}
	return false
}

func collect(wait *sync.WaitGroup, results chan Resource) []Resource {
	go func() {
		wait.Wait()
		close(results)
	}()
	var resources []Resource
	for {
		r, ok := <-results
		if ok {
			resources = append(resources, r)
		} else {
			return resources
		}
	}
}
