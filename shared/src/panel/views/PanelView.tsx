import H from 'history'
import React from 'react'
import { concat, Observable, of, Subject, Subscription } from 'rxjs'
import { distinctUntilChanged, map, switchMap } from 'rxjs/operators'
import { PanelViewWithComponent, ViewProviderRegistrationOptions } from '../../api/client/services/view'
import { ActivationProps } from '../../components/activation/Activation'
import { FetchFileCtx } from '../../components/CodeExcerpt'
import { Markdown } from '../../components/Markdown'
import { ExtensionsControllerProps } from '../../extensions/controller'
import { SettingsCascadeProps } from '../../settings/settings'
import { createLinkClickHandler } from '../../util/linkClickHandler'
import { renderMarkdown } from '../../util/markdown'
import { EmptyPanelView } from './EmptyPanelView'
import { HierarchicalLocationsView } from './HierarchicalLocationsView'

interface Props extends ExtensionsControllerProps, SettingsCascadeProps, ActivationProps {
    panelView: PanelViewWithComponent & Pick<ViewProviderRegistrationOptions, 'id'>
    repoName?: string
    history: H.History
    location: H.Location
    isLightTheme: boolean
    fetchHighlightedFileLines: (ctx: FetchFileCtx, force?: boolean) => Observable<string[]>
}

interface State {}

/**
 * A panel view contributed by an extension using {@link sourcegraph.app.createPanelView}.
 */
export class PanelView extends React.PureComponent<Props, State> {
    private subscriptions = new Subscription()
    private componentUpdates = new Subject<Props>()

    public componentWillMount(): void {
        this.subscriptions.add(
            concat([this.props], this.componentUpdates)
                .pipe(
                    map(props => props.panelView.locationProvider),
                    distinctUntilChanged(),
                    switchMap(locationProvider => locationProvider || of(null))
                )
                .subscribe(locations => {
                    if (
                        this.props.activation &&
                        this.props.panelView.id === 'references' &&
                        locations &&
                        locations.length > 0
                    ) {
                        this.props.activation.update({ FoundReferences: true })
                    }
                })
        )
    }

    public componentDidUpdate?(prevProps: Readonly<Props>, prevState: Readonly<State>): void {
        this.componentUpdates.next(this.props)
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        return (
            <div
                onClick={createLinkClickHandler(this.props.history)}
                className="panel__tabs-content panel__tabs-content--scroll"
            >
                {this.props.panelView.content && (
                    <div className="px-2 pt-2">
                        <Markdown dangerousInnerHTML={renderMarkdown(this.props.panelView.content)} />
                    </div>
                )}
                {this.props.panelView.reactElement}
                {this.props.panelView.locationProvider && this.props.repoName && (
                    <HierarchicalLocationsView
                        locations={this.props.panelView.locationProvider}
                        defaultGroup={this.props.repoName}
                        isLightTheme={this.props.isLightTheme}
                        fetchHighlightedFileLines={this.props.fetchHighlightedFileLines}
                        extensionsController={this.props.extensionsController}
                        settingsCascade={this.props.settingsCascade}
                    />
                )}
                {!this.props.panelView.content &&
                    !this.props.panelView.reactElement &&
                    !this.props.panelView.locationProvider && <EmptyPanelView className="mt-3" />}
            </div>
        )
    }
}
