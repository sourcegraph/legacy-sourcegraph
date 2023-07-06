package rfc

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"google.golang.org/api/docs/v1"
	"google.golang.org/api/drive/v3"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func Create(ctx context.Context, title string, driveSpec DriveSpec, out *std.Output) error {
	newFile, newRfcId, err := createRFCs(ctx, title, driveSpec, out)
	if err != nil {
		return errors.Wrap(err, "cannot create RFC")
	}

	if driveSpec == PrivateDrive {
		newFile2, err := leaveBreadcrumbForPrivateOnPublic(ctx, newFile, newRfcId, out)
		if err != nil {
			return errors.Wrap(err, "Cannot create breadcrumb file")
		}
		openFile(newFile2, out)
	}
	openFile(newFile, out)
	return nil
}

func findLastRfcIdFor(ctx context.Context, driveSpec DriveSpec, out *std.Output) (int, error) {
	var maxRfcId int = 0
	if err := queryRFCs(ctx, "", driveSpec, func(r *drive.FileList) error {
		if len(r.Files) == 0 {
			return nil
		}
		for _, f := range r.Files {
			matches := rfcIdRegex.FindStringSubmatch(f.Name)
			if len(matches) == 2 {
				if number, err := strconv.Atoi(matches[1]); err == nil {
					if number > maxRfcId {
						maxRfcId = number
					}
				} else {
					return errors.Wrap(err, "Cannot determine RFC ID")
				}
			}
		}
		return nil
	}, out); err != nil {
		return 0, err
	}
	if maxRfcId == 0 {
		return 0, errors.Errorf("Cannot determine next RFC ID")
	}
	return maxRfcId, nil
}

func findNextRfcId(ctx context.Context, out *std.Output) (int, error) {
	out.Write("Checking public RFCs")
	maxPublicRfcId, err := findLastRfcIdFor(ctx, PublicDrive, out)
	if err != nil {
		return 0, err
	}
	out.Write(fmt.Sprintf("Last public RFC = %d", maxPublicRfcId))

	out.Write("Checking private RFCs")
	maxPrivateRfcId, err := findLastRfcIdFor(ctx, PrivateDrive, out)
	if err != nil {
		return 0, err
	}
	out.Write(fmt.Sprintf("Last private RFC = %d", maxPrivateRfcId))

	if maxPublicRfcId > maxPrivateRfcId {
		return maxPublicRfcId + 1, nil
	} else {
		return maxPrivateRfcId + 1, nil
	}
}

func updateRfcContent(ctx context.Context, newFile *drive.File, nextRfcId int, title string,
	driveSpec DriveSpec, out *std.Output) error {
	docService, err := getDocsService(ctx, ScopePermissionsReadWrite, out)
	if err != nil {
		return errors.Wrap(err, "Cannot create docs client")
	}

	doc, err := docService.Documents.Get(newFile.Id).Do()
	if err != nil {
		return errors.Wrap(err, "Cannot access newly created file")
	}

	var change []*docs.Request
	var foundTitle bool = false
	var foundReminder bool = false

	for _, elem := range doc.Body.Content {
		if elem.Paragraph != nil {
			if !foundTitle {
				// First paragraph is the title
				content := elem.Paragraph.Elements[0].TextRun.Content
				matches := rfcDocRegex.FindStringSubmatch(content)
				if len(matches) != 5 {
					return errors.Errorf("Document format mismatch")
				}
				rfcSize := int64(len(matches[1]))
				numberSize := int64(len(matches[2]))
				titleSize := int64(len(matches[4]))

				nextRfcIdStr := strconv.Itoa(nextRfcId)
				change = append(change, []*docs.Request{
					// Replace the title
					{
						DeleteContentRange: &docs.DeleteContentRangeRequest{
							Range: &docs.Range{
								StartIndex: elem.EndIndex - titleSize - 1,
								EndIndex:   elem.EndIndex - 1,
							},
						},
					},
					{
						InsertText: &docs.InsertTextRequest{
							Location: &docs.Location{Index: elem.EndIndex - titleSize - 1},
							Text:     title,
						},
					},
				}...)

				// Replace the number
				change = append(change, []*docs.Request{
					{
						DeleteContentRange: &docs.DeleteContentRangeRequest{
							Range: &docs.Range{
								StartIndex: elem.StartIndex + rfcSize,
								EndIndex:   elem.StartIndex + rfcSize + numberSize,
							},
						},
					},
					{
						InsertText: &docs.InsertTextRequest{
							Location: &docs.Location{Index: elem.StartIndex + 4},
							Text:     nextRfcIdStr,
						},
					},
				}...)

				if driveSpec == PrivateDrive {
					// Add "PRIVATE" to the title
					change = append(change, &docs.Request{
						InsertText: &docs.InsertTextRequest{
							Location: &docs.Location{
								Index: elem.StartIndex + rfcSize + rfcSize,
							},
							Text: "PRIVATE ",
						},
					})
				}

				foundTitle = true
			}
		}

		if elem.Table != nil {
			// First table is the reminder
			if !foundReminder {
				if len(elem.Table.TableRows) != 1 ||
					len(elem.Table.TableRows[0].TableCells) != 1 ||
					len(elem.Table.TableRows[0].TableCells[0].Content) != 1 ||
					len(elem.Table.TableRows[0].TableCells[0].Content[0].Paragraph.Elements) == 0 {
					return errors.Errorf("Reminder table not found")
				}

				content := elem.Table.TableRows[0].TableCells[0].Content[0].
					Paragraph.Elements[0].TextRun.Content
				if strings.Contains(content, "Rename this RFC in this format") {
					// Remove the reminder, as we are doing for the user
					change = append([]*docs.Request{{
						DeleteContentRange: &docs.DeleteContentRangeRequest{
							Range: &docs.Range{
								StartIndex: elem.StartIndex,
								EndIndex:   elem.EndIndex,
							},
						},
					}}, change...)

					foundReminder = true
				}
			}
		}
	}

	if _, err := docService.Documents.BatchUpdate(newFile.Id, &docs.BatchUpdateDocumentRequest{
		Requests: change,
	}).Do(); err != nil {
		return errors.Wrap(err, "Cannot update RFC title")
	}

	return nil
}

func createRFCs(ctx context.Context, title string, driveSpec DriveSpec,
	out *std.Output) (*drive.File, int, error) {
	srv, err := getService(ctx, ScopePermissionsReadWrite, out)
	if err != nil {
		return nil, 0, err
	}

	template, err := srv.Files.Get(ProblemSolutionDriveTemplate).
		Context(ctx).
		SupportsTeamDrives(true).
		SupportsAllDrives(true).
		Do()
	if err != nil {
		return nil, 0, errors.Wrap(err, "failed to get template")
	}
	out.Write(fmt.Sprintf("Using template: %s", template.Name))

	nextRfcId, err := findNextRfcId(ctx, out)
	if err != nil {
		return nil, 0, err
	}
	var privateMark string
	if driveSpec == PrivateDrive {
		privateMark = "PRIVATE "
	}
	rfcFileTitle := fmt.Sprintf("RFC %d %sWIP: %s", nextRfcId, privateMark, title)
	newFileDetails := drive.File{
		Name:    rfcFileTitle,
		Parents: []string{driveSpec.FolderID},
	}

	newFile, err := srv.Files.Copy(template.Id, &newFileDetails).
		SupportsAllDrives(true).
		SupportsTeamDrives(true).
		Do()
	if err != nil {
		return nil, 0, errors.Wrap(err, "failed to create new RFC")
	}
	out.Write(fmt.Sprintf("New RFC created: %s (%s)", newFile.Name, newFile.Id))

	if err := updateRfcContent(ctx, newFile, nextRfcId, title, driveSpec, out); err != nil {
		return nil, 0, errors.Wrap(err, "Cannot update RFC content")
	}

	return newFile, nextRfcId, nil
}

func leaveBreadcrumbForPrivateOnPublic(ctx context.Context, rfcDoc *drive.File, nextRfcId int,
	out *std.Output) (*drive.File, error) {
	srv, err := getService(ctx, ScopePermissionsReadWrite, out)
	if err != nil {
		return nil, err
	}

	docService, err := getDocsService(ctx, ScopePermissionsReadWrite, out)
	if err != nil {
		return nil, errors.Wrap(err, "Cannot create docs client")
	}

	title := fmt.Sprintf("RFC %d is private", nextRfcId)

	newFile, err := srv.Files.Create(&drive.File{
		Name:     title,
		MimeType: "application/vnd.google-apps.document",
		Parents:  []string{PublicDrive.FolderID},
	}).
		SupportsAllDrives(true).
		SupportsTeamDrives(true).
		Do()
	if err != nil {
		return nil, errors.Wrap(err, "Cannot create breadcrumb file")
	}

	_, err = docService.Documents.BatchUpdate(newFile.Id, &docs.BatchUpdateDocumentRequest{
		Requests: []*docs.Request{
			{
				InsertText: &docs.InsertTextRequest{
					Location: &docs.Location{Index: 1},
					Text:     title,
				},
			},
			// Make "private" a link to the private RFC
			{
				UpdateTextStyle: &docs.UpdateTextStyleRequest{
					Range: &docs.Range{
						StartIndex: int64(len(title) - len("private") + 1),
						EndIndex:   int64(len(title) + 1),
					},
					TextStyle: &docs.TextStyle{
						Link: &docs.Link{
							Url: "https://docs.google.com/document/d/" + rfcDoc.Id,
						},
					},
					Fields: "link",
				},
			},
		},
	}).Do()
	if err != nil {
		return nil, errors.Wrap(err, "Cannot update breadcrumb content")
	}

	out.Write(fmt.Sprintf("New RFC public breadcrumb created: %s (%s)", newFile.Name, newFile.Id))
	return newFile, nil
}
