import { flatten } from 'lodash'
import { from, Observable, of, Subscription, Unsubscribable } from 'rxjs'
import {
    map,
    publishReplay,
    refCount,
    startWith,
    switchMap,
    toArray,
    tap,
    first,
    delay,
    debounceTime,
    filter,
} from 'rxjs/operators'
import * as sourcegraph from 'sourcegraph'
import { combineLatestOrDefault } from '../../../../shared/src/util/rxjs/combineLatestOrDefault'
import { isDefined } from '../../../../shared/src/util/types'
import { MAX_RESULTS, OTHER_CODE_ACTIONS, REPO_INCLUDE } from './misc'
import { memoizedFindTextInFiles } from './util'

const TAG_TRAVIS_GO = 'travis-go'

const diagnosticCollection = sourcegraph.languages.createDiagnosticCollection('demo0')

export function registerTravisGo(): Unsubscribable {
    const subscriptions = new Subscription()
    subscriptions.add(startDiagnostics())
    subscriptions.add(registerStatusProvider(diagnostics, diagnosticCollection))
    subscriptions.add(sourcegraph.languages.registerCodeActionProvider(['*'], createCodeActionProvider()))
    return subscriptions
}

function startDiagnostics(): Unsubscribable {
    const subscriptions = new Subscription()
    subscriptions.add(diagnosticCollection)
    subscriptions.add(
        diagnostics.subscribe(entries => {
            diagnosticCollection.set(entries)
        })
    )
    return subscriptions
}

const FIX_GO_VERSION_COMMAND_ID = 'travis-ci.fixGoVersion'

interface FixGoVersionParams {}

function registerStatusProvider(
    diagnostics: Observable<[URL, sourcegraph.Diagnostic[]][]>,
    diagnosticCollection: sourcegraph.DiagnosticCollection
): Unsubscribable {
    const subscriptions = new Subscription()
    subscriptions.add(
        sourcegraph.status.registerStatusProvider('travis-ci', {
            provideStatus: (scope): sourcegraph.Subscribable<sourcegraph.Status | null> => {
                // TODO!(sqs): dont ignore scope
                return diagnostics.pipe(
                    switchMap<[URL, sourcegraph.Diagnostic[]][], Promise<sourcegraph.Status>>(async diagnostics => ({
                        title: 'Travis CI',
                        state: {
                            completion: sourcegraph.CheckResult.Completed,
                            result: sourcegraph.CheckResult.Success,
                            message: 'All builds passing, all repository configuration valid and up-to-date',
                        },
                        sections: {
                            settings: {
                                kind: sourcegraph.MarkupKind.Markdown,
                                value: `Require all projects to configure Travis CI`,
                            },
                            notifications: {
                                kind: sourcegraph.MarkupKind.Markdown,
                                value: `Notify **@sourcegraph/devops** of changes to Travis CI configuration`,
                            },
                        },
                        diagnostics: diagnosticCollection,
                    })),
                    startWith<sourcegraph.Status>({
                        title: 'Travis CI',
                        state: { completion: sourcegraph.CheckResult.InProgress },
                    })
                )
            },
        })
    )
    subscriptions.add(
        sourcegraph.notifications.registerNotificationProvider('travis-ci', {
            provideNotifications: scope =>
                // TODO!(sqs): dont ignore scope
                diagnostics.pipe(
                    switchMap(async diagnostics => [
                        {
                            message: 'Build failures on default branch',
                            type: sourcegraph.NotificationType.Error,
                            actions: [
                                {
                                    command: {
                                        command: 'TODO!(sqs)',
                                        title: 'sourcegraph/codeintellify (31 minutes ago)',
                                    },
                                },
                                {
                                    command: {
                                        command: 'TODO!(sqs)',
                                        title: 'sourcegraph/go-diff (7 hours ago)',
                                    },
                                },
                                {
                                    command: {
                                        command: 'TODO!(sqs)',
                                        title: 'sourcegraph/syntect_server (1 day ago)',
                                    },
                                },
                            ],
                        },
                        ...((diagnostics.length > 0
                            ? [
                                  {
                                      message: 'Outdated Go version specified in Travis CI configuration',
                                      type: sourcegraph.NotificationType.Info,
                                      actions: [
                                          {
                                              plan: {
                                                  title: 'Use latest Go version in all .travis.yml files',
                                                  operations: [
                                                      {
                                                          command: {
                                                              command: FIX_GO_VERSION_COMMAND_ID,
                                                          },
                                                      },
                                                  ],
                                              },
                                          },
                                          {
                                              command: {
                                                  command: 'TODO!(sqs)',
                                                  title: 'Change preferred Go version (1.13)',
                                              },
                                          },
                                      ],
                                  },
                              ]
                            : []) as sourcegraph.Notification[]),
                        {
                            message: 'Missing Travis CI configuration',
                            type: sourcegraph.NotificationType.Warning,
                            actions: [
                                {
                                    plan: {
                                        title: 'Infer .travis.yml (best-effort) for repositories lacking it',
                                        operations: [
                                            {
                                                command: {
                                                    command: 'TODO!(sqs)',
                                                },
                                            },
                                        ],
                                    },
                                },
                            ],
                        },
                    ])
                ),
        })
    )
    subscriptions.add(
        sourcegraph.commands.registerPlanCommand<FixGoVersionParams>(
            FIX_GO_VERSION_COMMAND_ID,
            async (_params, _diagnostics) => {
                // TODO!(sqs): diagnostics should be passed from caller based on what plan asks for
                //
                // TODO!(sqs): pause to give time for diagnostics to be computed, super hacky
                const diags = await diagnostics
                    .pipe(
                        filter(d => d.some(([, diags]) => diags.length > 0)),
                        debounceTime(1500),
                        first()
                    )
                    .toPromise()
                const { edit } = await computeFixAllActionsFromDiagnostics(diags)
                return edit
            }
        )
    )
    return subscriptions
}

const diagnostics: Observable<[URL, sourcegraph.Diagnostic[]][]> = from(sourcegraph.workspace.rootChanges).pipe(
    startWith(void 0),
    map(() => sourcegraph.workspace.roots),
    switchMap(async roots => {
        if (roots.length > 0) {
            return of<[URL, sourcegraph.Diagnostic[]][]>([]) // TODO!(sqs): dont run in comparison mode
        }

        const results = flatten(
            await from(
                memoizedFindTextInFiles(
                    { pattern: '', type: 'regexp' },
                    {
                        repositories: {
                            includes: [REPO_INCLUDE],
                            type: 'regexp',
                        },
                        files: {
                            includes: ['\\.travis\\.yml$'],
                            type: 'regexp',
                        },
                        maxResults: MAX_RESULTS,
                    }
                )
            )
                .pipe(toArray())
                .toPromise()
        )
        return combineLatestOrDefault(
            results.map(async ({ uri }) => {
                const { text } = await sourcegraph.workspace.openTextDocument(new URL(uri))
                const diagnostics: sourcegraph.Diagnostic[] = flatten(
                    findMatchRanges(text, /(^go:)|(^language: go)/g)
                        .slice(0, 1)
                        .map(
                            range =>
                                ({
                                    message: 'Outdated Go version used in Travis CI',
                                    range,
                                    severity: sourcegraph.DiagnosticSeverity.Warning,
                                    tags: [TAG_TRAVIS_GO],
                                } as sourcegraph.Diagnostic)
                        )
                )
                return [new URL(uri), diagnostics] as [URL, sourcegraph.Diagnostic[]]
            })
        ).pipe(map(items => items.filter(isDefined)))
    }),
    switchMap(results => results),
    publishReplay(),
    refCount() //TODO!(sqs): or just share()?
)

function createCodeActionProvider(): sourcegraph.CodeActionProvider {
    return {
        provideCodeActions: async (doc, _rangeOrSelection, context): Promise<sourcegraph.CodeAction[]> => {
            const diag = context.diagnostics.find(isTravisGoDiagnostic)
            if (!diag) {
                return []
            }

            const fixAllAction = await computeFixAllAction()

            return [
                {
                    title: 'Use current Go version',
                    edit: computeFixEdit(diag, doc),
                    diagnostics: [diag],
                },
                fixAllAction.edit && Array.from(fixAllAction.edit.textEdits()).length > 1
                    ? {
                          title: `Fix in all ${Array.from(fixAllAction.edit.textEdits()).length} repositories`,
                          ...fixAllAction,
                      }
                    : null,
                {
                    title: `View Travis CI docs`,
                    command: {
                        title: '',
                        command: 'open',
                        arguments: ['https://docs.travis-ci.com/user/languages/go/'],
                    },
                    diagnostics: [diag],
                },
                ...OTHER_CODE_ACTIONS,
            ].filter(isDefined)
        },
    }
}

function isTravisGoDiagnostic(diag: sourcegraph.Diagnostic): boolean {
    return diag.tags && diag.tags.includes(TAG_TRAVIS_GO)
}

function computeFixEdit(
    diag: sourcegraph.Diagnostic,
    doc: sourcegraph.TextDocument,
    edit = new sourcegraph.WorkspaceEdit()
): sourcegraph.WorkspaceEdit {
    if (!doc.text.includes('1.13.x')) {
        const ranges = findMatchRanges(doc.text, /^go:/g)
        if (ranges.length > 0) {
            for (const range of ranges) {
                edit.insert(new URL(doc.uri), range.end, `\n  - "1.13.x"`)
            }
        } else {
            edit.insert(new URL(doc.uri), doc.positionAt(doc.text.length), `\n\ngo:\n  - "1.13.x"\n`)
        }
    }
    return edit
}

async function computeFixAllAction(): Promise<Pick<sourcegraph.CodeAction, 'edit' | 'diagnostics'>> {
    // TODO!(sqs): Make this listen for new diagnostics and include those too, but that might be
    // super inefficient because it's n^2, so maybe an altogether better/different solution is
    // needed.
    const allTravisGoDiags = sourcegraph.languages
        .getDiagnostics()
        .map(([uri, diagnostics]) => {
            const matchingDiags = diagnostics.filter(isTravisGoDiagnostic)
            return matchingDiags.length > 0
                ? ([uri, matchingDiags] as ReturnType<typeof sourcegraph.languages.getDiagnostics>[0])
                : null
        })
        .filter(isDefined)
    return computeFixAllActionsFromDiagnostics(allTravisGoDiags)
}

async function computeFixAllActionsFromDiagnostics(
    allDiags: [URL, sourcegraph.Diagnostic[]][]
): Promise<Pick<sourcegraph.CodeAction, 'edit' | 'diagnostics'>> {
    const edit = new sourcegraph.WorkspaceEdit()
    for (const [uri, diags] of allDiags) {
        const doc = await sourcegraph.workspace.openTextDocument(uri)
        for (const diag of diags) {
            computeFixEdit(diag, doc, edit)
        }
    }
    return { edit, diagnostics: flatten(allDiags.map(([, diagnostics]) => diagnostics)) }
}

function findMatchRanges(text: string, pattern: RegExp): sourcegraph.Range[] {
    const lines = text.split('\n')
    const ranges: sourcegraph.Range[] = []
    for (const [i, line] of lines.entries()) {
        pattern.lastIndex = 0
        const match = pattern.exec(line)
        if (match) {
            ranges.push(new sourcegraph.Range(i, match.index, i, match.index + match[0].length))
        }
    }
    return ranges
}
