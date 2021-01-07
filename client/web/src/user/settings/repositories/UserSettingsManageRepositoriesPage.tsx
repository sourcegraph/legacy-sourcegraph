import React, { FormEvent, useCallback, useEffect, useState } from 'react'
import { TelemetryProps } from '../../../../../shared/src/telemetry/telemetryService'
import { RouteComponentProps } from 'react-router'
import { PageTitle } from '../../../components/PageTitle'
import { CheckboxRepositoryNode } from '../../../components/RepositoryNode'
import { Form } from '../../../../../branded/src/components/Form'
import { Link } from '../../../../../shared/src/components/Link'
import { ExternalServiceKind, ExternalServicesResult, Maybe } from '../../../graphql-operations'
import {
    queryExternalServices,
    setExternalServiceRepos,
    listAffiliatedRepositories,
} from '../../../components/externalServices/backend'
import { ErrorAlert } from '../../../components/alerts'

interface Props extends RouteComponentProps, TelemetryProps {
    userID: string
    routingPrefix: string
}

interface Repo {
    name: string
    codeHost: Maybe<{ kind: ExternalServiceKind; id: string; displayName: string }>
    private: boolean
}

const PER_PAGE = 20

// initial state constants
const emptyRepos: Repo[] = []
const initialRepoState = {
    repos: emptyRepos,
    loading: false,
    loaded: false,
    error: '',
}
const selectionMap = new Map<string, Repo>()
const emptyHosts: ExternalServicesResult['externalServices']['nodes'] = []
const emptyRepoNames: string[] = []
const initialCodeHostState = {
    hosts: emptyHosts,
    loaded: false,
    configuredRepos: emptyRepoNames,
}

/**
 * A page to manage the repositories a user syncs from their connected code hosts.
 */
export const UserSettingsManageRepositoriesPage: React.FunctionComponent<Props> = ({
    history,
    userID,
    routingPrefix,
    telemetryService,
}) => {
    useEffect(() => {
        telemetryService.logViewEvent('UserSettingsRepositories')
    }, [telemetryService])

    // set up state hooks
    const [repoState, setRepoState] = useState(initialRepoState)
    const [selectionState, setSelectionState] = useState({ repos: selectionMap, loaded: false, radio: 'all' })
    const [currentPage, setPage] = useState(1)
    const [query, setQuery] = useState('')
    const [codeHostFilter, setCodeHostFilter] = useState('')
    const [codeHosts, setCodeHosts] = useState(initialCodeHostState)

    useCallback(() => {
        // first we should load code hosts
        if (!codeHosts.loaded) {
            const codeHostSubscription = queryExternalServices({
                first: null,
                after: null,
                namespace: userID,
            }).subscribe(result => {
                const selected: string[] = []
                for (const host of result.nodes) {
                    const cfg = JSON.parse(host.config)
                    switch (host.kind) {
                        case ExternalServiceKind.GITLAB:
                            if (cfg.projects !== undefined) {
                                cfg.projects.map((project: any) => {
                                    selected.push(project.name)
                                })
                            }
                            break
                        case ExternalServiceKind.GITHUB:
                            if (cfg.repos !== undefined) {
                                selected.push(...cfg.repos)
                            }
                            break
                    }
                }
                setCodeHosts({
                    loaded: true,
                    hosts: result.nodes,
                    configuredRepos: selected,
                })
                if (selected.length !== 0) {
                    setSelectionState({
                        repos: selectionState.repos,
                        radio: 'selected',
                        loaded: selectionState.loaded,
                    })
                }
                codeHostSubscription.unsubscribe()
            })
        }
    }, [codeHosts, selectionState.loaded, selectionState.repos, userID])()

    // once we've loaded code hosts and the 'selected' panel is visible, load repos and set the loading state
    useCallback(() => {
        if (selectionState.radio === 'selected' && !repoState.loaded && !repoState.loading) {
            setRepoState({
                repos: emptyRepos,
                loading: true,
                loaded: false,
                error: '',
            })
            const listReposSubscription = listAffiliatedRepositories({
                user: userID,
                codeHost: null,
                query: null,
            }).subscribe(
                result => {
                    setRepoState({
                        repos: result.affiliatedRepositories.nodes,
                        loading: false,
                        loaded: true,
                        error: '',
                    })
                    listReposSubscription.unsubscribe()
                },
                error => {
                    setRepoState({
                        repos: emptyRepos,
                        loading: false,
                        loaded: true,
                        error: String(error),
                    })
                }
            )
        }
    }, [repoState.loaded, repoState.loading, selectionState.radio, userID])()

    // if we've loaded repos we should then populate our selection from
    // code host config.
    if (repoState.loaded && codeHosts.loaded && !selectionState.loaded) {
        const selectedRepos = new Map<string, Repo>()
        codeHosts.configuredRepos.map(repo => {
            selectedRepos.set(repo, repoState.repos.find(fullRepo => fullRepo.name === repo)!)
        })

        let radioState = selectionState.radio
        if (selectionState.radio === 'all' && selectedRepos.size > 0) {
            radioState = 'selected'
        }
        setSelectionState({
            repos: selectedRepos,
            loaded: true,
            radio: radioState,
        })
    }

    // filter our set of repos based on query & code host selection
    const filteredRepos: Repo[] = []
    for (const repo of repoState.repos) {
        if (!repo.name.toLowerCase().includes(query)) {
            continue
        }
        if (codeHostFilter !== '' && repo.codeHost!.id !== codeHostFilter) {
            continue
        }
        filteredRepos.push(repo)
    }

    // create elements for pagination
    const pages: JSX.Element[] = []
    for (let page = 1; page <= Math.ceil(filteredRepos.length / PER_PAGE); page++) {
        pages.push(
            <a className="btn" onClick={() => setPage(page)}>
                <p className={(currentPage === page && 'text-primary') || 'text-muted'}>{page}</p>
            </a>
        )
    }

    // save changes and update code hosts
    const submit = useCallback(
        (event: FormEvent<HTMLFormElement>): void => {
            event.preventDefault()
            for (const host of codeHosts.hosts) {
                const repos: string[] = []
                for (const repo of selectionState.repos.values()) {
                    if (repo.codeHost!.id !== host.id) {
                        continue
                    }
                    repos.push(repo.name)
                }
                setExternalServiceRepos({
                    id: host.id,
                    allRepos: selectionState.radio === 'all',
                    repos: (selectionState.radio === 'selected' && repos) || null,
                }).catch(error => {
                    setRepoState({ ...repoState, error: String(error) })
                })
            }
            history.push(routingPrefix + '/repositories')
        },
        [codeHosts.hosts, history, repoState, routingPrefix, selectionState.radio, selectionState.repos]
    )

    const handleRadioSelect = (changeEvent: React.ChangeEvent<HTMLInputElement>): void => {
        setSelectionState({
            repos: selectionState.repos,
            radio: changeEvent.currentTarget.value,
            loaded: selectionState.loaded,
        })
    }

    const modeSelect: JSX.Element = (
        <Form>
            <div className="d-flex flex-row align-items-baseline">
                <input type="radio" value="all" checked={selectionState.radio === 'all'} onChange={handleRadioSelect} />
                <div className="d-flex flex-column ml-2">
                    <p className="mb-0">Sync all my repositories</p>
                    <p className="text-muted">Will sync all current and future public and private repositories</p>
                </div>
            </div>
            <div className="d-flex flex-row align-items-baseline">
                <input
                    type="radio"
                    value="selected"
                    checked={selectionState.radio === 'selected'}
                    onChange={handleRadioSelect}
                />
                <div className="d-flex flex-column ml-2">
                    <p className="mb-0">Sync selected repositories</p>
                </div>
            </div>
        </Form>
    )

    const filterControls: JSX.Element = (
        <Form className="w-100 d-inline-flex justify-content-between flex-row filtered-connection__form mt-3">
            <div className="d-inline-flex flex-row mr-3 align-items-baseline">
                <p className="text-xl-center text-nowrap mr-2">Code Host:</p>
                <select
                    className="form-control"
                    name="code-host"
                    onChange={event => {
                        setCodeHostFilter(event.target.value)
                    }}
                >
                    <option key="any" value="" label="Any" />
                    {codeHosts.hosts.map(value => (
                        <option key={value.id} value={value.id} label={value.displayName} />
                    ))}
                </select>
            </div>
            <input
                className="form-control filtered-connection__filter"
                type="search"
                placeholder="Search repositories..."
                name="query"
                autoComplete="off"
                autoCorrect="off"
                autoCapitalize="off"
                spellCheck={false}
                onChange={event => {
                    setQuery(event.target.value)
                }}
            />
        </Form>
    )

    const onRepoClicked = useCallback(
        (repo: Repo) => (): void => {
            const newMap = new Map(selectionState.repos)
            if (newMap.has(repo.name)) {
                newMap.delete(repo.name)
            } else {
                newMap.set(repo.name, repo)
            }
            setSelectionState({
                repos: newMap,
                radio: selectionState.radio,
                loaded: selectionState.loaded,
            })
        },
        [selectionState, setSelectionState]
    )

    const rows: JSX.Element = (
        <tbody>
            <tr className="align-items-baseline d-flex" key="header">
                <td className="w-100 d-flex justify-content-flex-start align-items-center">
                    <input
                        className="mr-2"
                        type="checkbox"
                        checked={selectionState.repos.size === filteredRepos.length}
                        onChange={() => {
                            const newMap = new Map<string, Repo>()
                            // if not all repos are selected, we should select all, otherwise empty the selection
                            if (selectionState.repos.size !== filteredRepos.length) {
                                for (const repo of filteredRepos) {
                                    newMap.set(repo.name, repo)
                                }
                            }
                            setSelectionState({
                                repos: newMap,
                                loaded: selectionState.loaded,
                                radio: selectionState.radio,
                            })
                        }}
                    />
                    <span
                        className={
                            ((selectionState.repos.size !== 0 && 'font-weight-bold ') || '') + 'repositories-header'
                        }
                    >
                        {(selectionState.repos.size > 0 && selectionState.repos.size) || 'No'} repositories selected.
                    </span>
                </td>
            </tr>
            {filteredRepos.map((repo, index) => {
                if (index < (currentPage - 1) * PER_PAGE || index >= currentPage * PER_PAGE) {
                    return
                }
                return (
                    <CheckboxRepositoryNode
                        name={repo.name}
                        key={repo.name}
                        onClick={onRepoClicked(repo)}
                        checked={selectionState.repos.has(repo.name)}
                        serviceType={repo.codeHost?.kind || ''}
                        isPrivate={repo.private}
                    />
                )
            })}
        </tbody>
    )

    const loadingAnimation: JSX.Element = (
        <tbody>
            <tr className="mt-2 align-items-baseline d-flex w-80">
                <td className="user-settings-repos__shimmer w-100 h-100 p-3 border-top-0" />
            </tr>
            <tr className="mt-2 align-items-baseline d-flex w-30">
                <td className="user-settings-repos__shimmer w-100 h-100 p-3 border-top-0" />
            </tr>
            <tr className="mt-2 align-items-baseline d-flex w-70">
                <td className="user-settings-repos__shimmer w-100 h-100 p-3 border-top-0" />
            </tr>
        </tbody>
    )
    return (
        <div className="p-2">
            <PageTitle title="Manage Repositories" />
            <h2 className="mb-2">Manage Repositories</h2>
            <p className="text-muted">
                Choose which repositories to sync with Sourcegraph so you can search all your code in one place.
            </p>
            <ul className="list-group">
                <li className="list-group-item" key="body">
                    <div className="p-4" key="description">
                        <h3>Your repositories</h3>
                        <p className="text-muted">
                            Repositories you own or collaborate on from{' '}
                            <a className="text-primary" href={routingPrefix + '/external-services'}>
                                connected code hosts
                            </a>
                        </p>
                        {
                            // display radio button for 'all' or 'selected' repos
                            modeSelect
                        }
                        {
                            // if we're in 'selected' mode, show a list of all the repos on the code hosts to select from
                            selectionState.radio === 'selected' && (
                                <div className="filtered-connection">
                                    {filterControls}
                                    {repoState.error !== '' && <ErrorAlert error={repoState.error} history={history} />}
                                    <table className="filtered-connection test-filtered-connection filtered-connection--noncompact table mt-3">
                                        {
                                            // if we're selecting repos, and the repos are still loading, display the loading animation
                                            selectionState.radio === 'selected' &&
                                                repoState.loading &&
                                                !repoState.loaded &&
                                                loadingAnimation
                                        }
                                        {
                                            // if the repos are loaded display the rows of repos
                                            repoState.loaded && rows
                                        }
                                    </table>
                                    <div className="d-flex flex-direction-row align-items-center w-100 justify-content-center">
                                        {
                                            // pagination control
                                            pages
                                        }
                                    </div>
                                </div>
                            )
                        }
                    </div>
                </li>
            </ul>
            <Form className="mt-4 d-flex" onSubmit={submit}>
                <button type="submit" className="btn btn-primary test-goto-add-external-service-page mr-2">
                    Save changes
                </button>

                <Link
                    className="btn btn-secondary test-goto-add-external-service-page"
                    to={`${routingPrefix}/repositories`}
                >
                    Cancel
                </Link>
            </Form>
        </div>
    )
}
