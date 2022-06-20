package main

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"
	"text/template"

	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/open"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const newDiscussionURL = "https://github.com/sourcegraph/sourcegraph/discussions/new"

type stopReadFunc func(lastRead string, err error) bool

func addFeedbackSubcommand(commands []*cli.Command) {
	giveFeedback := false
	feedbackFlag := cli.BoolFlag{
		Name:        "feedback",
		Usage:       "provide feedback about this command by opening up a Github discussion",
		Destination: &giveFeedback,
	}

	for _, command := range commands {
		command.Flags = append(command.Flags, &feedbackFlag)
		action := command.Action
		command.Action = func(ctx *cli.Context) error {
			if giveFeedback {
				return feedbackExec(ctx)
			}

			return action(ctx)
		}

		addFeedbackSubcommand(command.Subcommands)
	}
}

var feedbackCommand = &cli.Command{
	Name:     "feedback",
	Usage:    "opens up a Github discussion page to provide feedback about sg",
	Category: CategoryCompany,
	Action:   feedbackExec,
}

func feedbackExec(ctx *cli.Context) error {
	title, body, err := gatherFeedback(ctx)
	if err != nil {
		return err
	}
	body = addSGInformation(ctx, body)

	if err := sendFeedback(ctx.Context, title, "developer-experience", body); err != nil {
		return err
	}
	return nil
}

func gatherFeedback(ctx *cli.Context) (string, string, error) {
	std.Out.WriteNoticef("Gathering feedback for sg %s", ctx.Command.FullName())

	reader := bufio.NewReader(os.Stdin)
	fmt.Println("What is the title of your feedback ?")
	title, err := reader.ReadString('\n')
	if err != nil {
		return "", "", err
	}

	fmt.Println("Write your feedback below and press <CTRL+D> when you're done.")
	var sb strings.Builder
	for {
		line, err := reader.ReadString('\n')
		sb.WriteString(line)
		sb.WriteByte('\n')
		if err != nil {
			break
		}
	}
	body := sb.String()

	return title, body, nil
}

func addSGInformation(ctx *cli.Context, body string) string {
	tplt := template.Must(template.New("SG").Funcs(template.FuncMap{
		"inline_code": func(s string) string { return fmt.Sprintf("`%s`", s) },
	}).Parse(`{{.Content}}


### {{ inline_code "sg" }} Information

Commit: {{ inline_code .Commit}}
Command: {{ inline_code .Command}}
Flags: {{ inline_code .Flags}}
    `))

	flagPair := []string{}
	for _, f := range ctx.FlagNames() {
		if f == "feedback" {
			continue
		}
		flagPair = append(flagPair, fmt.Sprintf("%s=%v", f, ctx.Value(f)))
	}

	var buf bytes.Buffer
	data := struct {
		Content string
		Tick    string
		Commit  string
		Command string
		Flags   string
	}{
		body,
		"`",
		BuildCommit,
		"sg " + ctx.Command.FullName(),
		strings.Join(flagPair, " "),
	}
	_ = tplt.Execute(&buf, data)

	return buf.String()
}

func sendFeedback(ctx context.Context, title, category, body string) error {
	values := make(url.Values)
	values["category"] = []string{category}
	values["title"] = []string{title}
	values["body"] = []string{body}
	values["labels"] = []string{"sg,team/devx"}

	feedbackURL, err := url.Parse(newDiscussionURL)
	if err != nil {
		return err
	}

	feedbackURL.RawQuery = values.Encode()
	std.Out.WriteNoticef("Launching your browser to complete feedback")

	if err := open.URL(feedbackURL.String()); err != nil {
		return errors.Wrapf(err, "failed to launch browser for url %q", feedbackURL.String())
	}

	return nil
}
