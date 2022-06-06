package check

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/usershell"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

type Runner[Args any] struct {
	in         io.Reader
	out        *std.Output
	categories []Category[Args]
}

func NewRunner[Args any](in io.Reader, out *std.Output, categories []Category[Args]) *Runner[Args] {
	return &Runner[Args]{
		in:         in,
		out:        out,
		categories: categories,
	}
}

func (r *Runner[Args]) Check(
	ctx context.Context,
	args Args,
) error {
	ctx, err := usershell.Context(ctx)
	if err != nil {
		return err
	}

	results := r.run(ctx, args)
	if len(results.failed) > 0 {
		return errors.Newf("%d checks failed (%d skipped)", len(results.failed), len(results.skipped))
	}
	r.out.WriteSuccessf("%d checks passed! (%d skipped)", len(results.all), len(results.skipped))
	return nil
}

func (r *Runner[Args]) Fix(
	ctx context.Context,
	getArgs func() Args,
) error {
	ctx, err := usershell.Context(ctx)
	if err != nil {
		return err
	}

	// Keep interactive runner up until all issues are fixed, or all remaining issues are unfixable
	results := runResults{
		failed: []int{1}, // initialize, this gets reset immediately
	}
	var unfixable map[int]bool
	for len(results.failed) != len(unfixable) {
		// Get state
		args := getArgs()
		results = r.run(ctx, args)
		unfixable = make(map[int]bool)

		for _, idx := range results.failed {
			category := r.categories[idx]
			pending := r.out.Pending(output.Styledf(output.StylePending, "%d. %s - Determining status...", idx, category.Name))
			if !category.HasFixable() {
				pending.Complete(output.Styledf(output.StyleFailure, "%d. %s - Cannot be fixed automatically.", idx, category.Name))
				unfixable[idx] = true
				continue
			}

			for _, c := range category.Checks {
				// If category is fixed, we are good to go
				if c.IsMet() {
					continue
				}

				// Otherwise, check if this is fixable at all
				if c.Fix == nil {
					pending.WriteLine(output.Styledf(output.StyleGrey, "%d. %s - %s cannot be fixed automatically.",
						idx, category.Name, c.Name))
					unfixable[idx] = true
					continue
				}

				// TODO send pending
				if err := c.Fix(ctx, IO{
					Input:  r.in,
					Writer: r.out,
				}, args); err != nil {
					pending.Complete(output.Styledf(output.StyleFailure, "%d. %s - Failed to fix %s: %s",
						idx, category.Name, c.Name, err.Error()))
					return err
				}
				pending.WriteLine(output.Styledf(output.StyleSuccess, "%d. %s - Fixed %s!",
					idx, category.Name, c.Name))
			}

			pending.Complete(output.Styledf(output.StyleGrey, "%d. %s - Done", idx, category.Name))
		}
	}

	if len(unfixable) > 0 {
		return errors.Newf("%d categories could not be fixed", len(unfixable))
	}

	return nil
}

func (r *Runner[Args]) Interactive(
	ctx context.Context,
	getArgs func() Args,
) error {
	ctx, err := usershell.Context(ctx)
	if err != nil {
		return err
	}

	// Keep interactive runner up until all issues are fixed.
	results := runResults{
		failed: []int{1}, // initialize, this gets reset immediately
	}
	for len(results.failed) != 0 {
		r.out.Output.ClearScreen()

		// Get args when we run
		results = r.run(ctx, getArgs())

		r.out.WriteWarningf("Some checks failed. Which one do you want to fix?")

		idx, err := r.getNumberOutOf(results.all)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		selectedCategory := r.categories[idx]

		r.out.ClearScreen()

		// GetArgs again here to make sure args are up to date
		err = r.presentFailedCategoryWithOptions(ctx, idx, &selectedCategory, getArgs())
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
	}

	return nil
}

// runResults provides a summary of categories checks results.
type runResults struct {
	all     []int
	failed  []int
	skipped []int
}

// run is the main entrypoint for running the checks in this runner.
func (r *Runner[Args]) run(ctx context.Context, args Args) runResults {
	var results runResults
	for i := range r.categories {
		results.failed = append(results.failed, i)
		results.all = append(results.all, i)
	}

	categoryResults := make(map[string]bool)
	for _, category := range r.categories {
		categoryResults[category.Name] = false
	}

	for i, category := range r.categories {
		idx := i + 1

		if category.Enabled != nil {
			if err := category.Enabled(ctx, args); err != nil {
				r.out.WriteSkippedf("%d. %s %s[SKIPPED: %s]%s",
					idx, category.Name, output.StyleBold, err.Error(), output.StyleReset)
				results.skipped = append(results.skipped, idx)
				results.failed = removeEntry(results.failed, i)
				continue
			}
		}

		pending := r.out.Pending(output.Styledf(output.StylePending, "%d. %s - Determining status...", idx, category.Name))

		// Validate dependents
		var failed bool
		for _, d := range category.DependsOn {
			if !categoryResults[d] {
				failed = true
			}
		}

		// Validate checks
		if !failed {
			for _, c := range category.Checks {
				if err := c.RunCheck(ctx, IO{
					Input:  r.in,
					Writer: pending,
				}, args); err != nil {
					failed = true
				}
			}
		}

		// Report results
		pending.Destroy()
		categoryResults[category.Name] = failed
		if !failed {
			r.out.WriteSuccessf("%d. %s", idx, category.Name)
			results.failed = removeEntry(results.failed, i)
		} else {
			// TODO write failure
			r.out.WriteFailuref("%d. %s", idx, category.Name)
		}
	}

	if len(results.failed) == 0 {
		if len(results.skipped) == 0 {
			r.out.Write("")
			r.out.WriteLine(output.Linef(output.EmojiOk, output.StyleBold, "Everything looks good! Happy hacking!"))
		} else {
			r.out.Write("")
			r.out.WriteWarningf("Some checks were skipped.")
		}
	}

	return results
}

func removeEntry(s []int, val int) (result []int) {
	for _, e := range s {
		if e != val {
			result = append(result, e)
		}
	}
	return result
}

func (r *Runner[Args]) getNumberOutOf(numbers []int) (int, error) {
	var strs []string
	var idx = make(map[int]struct{})
	for _, num := range numbers {
		strs = append(strs, fmt.Sprintf("%d", num+1))
		idx[num+1] = struct{}{}
	}

	for {
		r.out.Writef("[%s]: ", strings.Join(strs, ","))
		var num int
		_, err := fmt.Fscan(r.in, &num)
		if err != nil {
			return 0, err
		}

		if _, ok := idx[num]; ok {
			return num - 1, nil
		}
		r.out.Writef("%d is an invalid choice :( Let's try again?\n", num)
	}
}

func (r *Runner[Args]) presentFailedCategoryWithOptions(ctx context.Context, categoryIdx int, category *Category[Args], args Args) error {
	r.printCategoryHeaderAndDependencies(categoryIdx+1, category)
	fixableCategory := category.HasFixable()

	choices := map[int]string{1: "I want to fix these manually"}
	if fixableCategory {
		choices[2] = "I'm feeling lucky. You try fixing all of it for me."
		choices[3] = "Go back"
	} else {
		choices[2] = "Go back"
	}

	choice, err := r.getChoice(choices)
	if err != nil {
		return err
	}

	switch choice {
	case 1:
		err = r.fixCategoryManually(ctx, categoryIdx, category, args)
	case 2:
		if fixableCategory {
			r.out.ClearScreen()
			err = r.fixCategoryAutomatically(ctx, category, args)
		}
	case 3:
		return nil
	}
	return err
}

func (r *Runner[Args]) printCategoryHeaderAndDependencies(categoryIdx int, category *Category[Args]) {
	r.out.WriteLine(output.Linef(output.EmojiLightbulb, output.CombineStyles(output.StyleSearchQuery, output.StyleBold), "%d. %s", categoryIdx, category.Name))
	r.out.Write("")
	r.out.Write("Checks:")

	for i, dep := range category.Checks {
		idx := i + 1
		if dep.IsMet() {
			r.out.WriteSuccessf("%d. %s", idx, dep.Name)
		} else {
			if dep.checkErr != nil {
				r.out.WriteFailuref("%d. %s: %s", idx, dep.Name, dep.checkErr)
			} else {
				r.out.WriteFailuref("%d. %s: %s", idx, dep.Name, "check failed")
			}
		}
	}
}

func (r *Runner[Args]) getChoice(choices map[int]string) (int, error) {
	for {
		r.out.Write("")
		r.out.WriteNoticef("What do you want to do?")

		for i := 0; i < len(choices); i++ {
			num := i + 1
			desc, ok := choices[num]
			if !ok {
				return 0, errors.Newf("internal error: %d not found in provided choices", i)
			}
			r.out.Writef("%s[%d]%s: %s", output.StyleBold, num, output.StyleReset, desc)
		}

		fmt.Printf("Enter choice: ")

		var s int
		_, err := fmt.Scan(&s)
		if err != nil {
			return 0, err
		}

		if _, ok := choices[s]; ok {
			return s, nil
		}
		r.out.WriteFailuref("Invalid choice")
	}
}

func (r *Runner[Args]) fixCategoryManually(ctx context.Context, categoryIdx int, category *Category[Args], args Args) error {
	for {
		toFix := []int{}

		for i, dep := range category.Checks {
			if dep.IsMet() {
				continue
			}

			toFix = append(toFix, i)
		}

		if len(toFix) == 0 {
			break
		}

		var idx int

		if len(toFix) == 1 {
			idx = toFix[0]
		} else {
			r.out.WriteNoticef("Which one do you want to fix?")
			var err error
			idx, err = r.getNumberOutOf(toFix)
			if err != nil {
				if err == io.EOF {
					return nil
				}
				return err
			}
		}

		dep := category.Checks[idx]

		r.out.WriteLine(output.Linef(output.EmojiFailure, output.CombineStyles(output.StyleWarning, output.StyleBold), "%s", dep.Name))
		r.out.Write("")

		if dep.checkErr != nil {
			r.out.WriteLine(output.Styledf(output.StyleBold, "Encountered the following error:\n\n%s%s\n", output.StyleReset, dep.checkErr))
		}

		r.out.WriteLine(output.Styled(output.StyleBold, "How to fix:"))

		// TODO how to define manual fixes

		pending := r.out.Pending(output.Styled(output.StylePending, "Determining status..."))
		for _, dep := range category.Checks {
			// TODO
			_ = dep.RunCheck(ctx, IO{
				Input:  r.in,
				Writer: pending,
			}, args)
		}
		pending.Destroy()

		r.printCategoryHeaderAndDependencies(categoryIdx, category)
	}

	return nil
}

func (r *Runner[Args]) fixCategoryAutomatically(ctx context.Context, category *Category[Args], args Args) error {
	// now go through the real dependencies
	for _, dep := range category.Checks {
		if dep.checkErr == nil {
			continue
		}

		if err := r.fixDependencyAutomatically(ctx, dep, args); err != nil {
			return err
		}
	}

	return nil
}

func (r *Runner[Args]) fixDependencyAutomatically(ctx context.Context, check *Check[Args], args Args) error {
	r.out.WriteNoticef("Trying my hardest to fix %q automatically...", check.Name)

	if err := check.Fix(ctx, IO{
		Input:  r.in,
		Writer: r.out,
	}, args); err != nil {
		r.out.WriteFailuref("Failed to fix check: %s", err)
		return err
	}

	r.out.WriteSuccessf("Done! %q should be fixed now!", check.Name)

	return nil
}

func waitForReturn() { fmt.Scanln() }
