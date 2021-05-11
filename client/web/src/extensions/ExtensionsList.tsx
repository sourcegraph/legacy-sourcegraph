import * as H from 'history'
import React, { useMemo } from 'react'

import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { isExtensionEnabled } from '@sourcegraph/shared/src/extensions/extension'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { ExtensionCategory, EXTENSION_CATEGORIES } from '@sourcegraph/shared/src/schema/extensionSchema'
import { SettingsCascadeProps, SettingsSubject } from '@sourcegraph/shared/src/settings/settings'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { createRecord } from '@sourcegraph/shared/src/util/createRecord'
import { isErrorLike } from '@sourcegraph/shared/src/util/errors'

import { ErrorAlert } from '../components/alerts'

import { ExtensionCard } from './ExtensionCard'
import { ExtensionListData, ExtensionsEnablement } from './ExtensionRegistry'
import { ExtensionsAreaRouteContext } from './ExtensionsArea'

interface Props
    extends SettingsCascadeProps,
        PlatformContextProps<'settings' | 'updateSettings' | 'requestGraphQL'>,
        Pick<ExtensionsAreaRouteContext, 'authenticatedUser'>,
        ThemeProps {
    subject: Pick<SettingsSubject, 'id' | 'viewerCanAdminister'>
    location: H.Location

    data: ExtensionListData | undefined
    selectedCategory: ExtensionCategory | 'All'
    enablementFilter: ExtensionsEnablement
    query: string
    showMoreExtensions: boolean
    onShowFullCategoryClicked: (category: ExtensionCategory) => void
}

const LOADING = 'loading' as const

/** Categories, but with 'Programming Languages' at the end */
const ORDERED_EXTENSION_CATEGORIES: ExtensionCategory[] = [
    ...EXTENSION_CATEGORIES.filter(category => category !== 'Programming languages'),
    'Programming languages',
]

/**
 * Displays a list of extensions.
 */
export const ExtensionsList: React.FunctionComponent<Props> = ({
    subject,
    settingsCascade,
    platformContext,
    data,
    selectedCategory,
    enablementFilter,
    query,
    showMoreExtensions,
    authenticatedUser,
    onShowFullCategoryClicked,
    ...props
}) => {
    // Settings subjects for extension toggle
    const viewerSubject = useMemo(
        () => settingsCascade.subjects?.find(settingsSubject => settingsSubject.subject.id === subject.id),
        [settingsCascade, subject.id]
    )
    const siteSubject = useMemo(
        () => settingsCascade.subjects?.find(settingsSubject => settingsSubject.subject.__typename === 'Site'),
        [settingsCascade]
    )

    if (!data || data === LOADING) {
        return <LoadingSpinner className="icon-inline" />
    }

    if (isErrorLike(data)) {
        return <ErrorAlert error={data} />
    }

    const { error, extensions, extensionIDsByCategory } = data

    if (Object.keys(extensions).length === 0) {
        return (
            <>
                {error && <ErrorAlert className="mb-2" error={error} />}
                {query ? (
                    <div className="text-muted">
                        No extensions match <strong>{query}</strong>.
                    </div>
                ) : (
                    <span className="text-muted">No extensions found</span>
                )}
            </>
        )
    }

    let categorySections: JSX.Element[]

    if (selectedCategory === 'All') {
        // TODO: fake pagination for now, implement soon.

        const filteredExtensionIDsByCategory = createRecord(ORDERED_EXTENSION_CATEGORIES, category => {
            if (enablementFilter === 'all') {
                return extensionIDsByCategory[category].primaryExtensionIDs
            }

            return extensionIDsByCategory[category].primaryExtensionIDs.filter(
                extensionID =>
                    (enablementFilter === 'enabled') === isExtensionEnabled(settingsCascade.final, extensionID)
            )
        })

        categorySections = ORDERED_EXTENSION_CATEGORIES.filter(
            category =>
                filteredExtensionIDsByCategory[category].length > 0 &&
                // Only show Programming Languages when "show more" was clicked
                (category !== 'Programming languages' || showMoreExtensions)
        ).map(category => {
            const extensionIDsForCategory = filteredExtensionIDsByCategory[category]

            if (extensionIDsForCategory.length > 6) {
                return (
                    <div key={category} className="mt-1">
                        <h3
                            className="extensions-list__category mb-3 font-weight-normal"
                            data-test-extension-category-header={category}
                        >
                            {category}
                        </h3>
                        <div className="extensions-list__cards mt-1">
                            {extensionIDsForCategory.slice(0, 6).map(extensionId => (
                                <ExtensionCard
                                    key={extensionId}
                                    subject={subject}
                                    viewerSubject={viewerSubject?.subject}
                                    siteSubject={siteSubject?.subject}
                                    node={extensions[extensionId]}
                                    settingsCascade={settingsCascade}
                                    platformContext={platformContext}
                                    enabled={isExtensionEnabled(settingsCascade.final, extensionId)}
                                    enabledForAllUsers={
                                        siteSubject ? isExtensionEnabled(siteSubject.settings, extensionId) : false
                                    }
                                    isLightTheme={props.isLightTheme}
                                    settingsURL={authenticatedUser?.settingsURL}
                                    authenticatedUser={authenticatedUser}
                                />
                            ))}
                        </div>
                        <div className="d-flex justify-content-center mt-4">
                            <button
                                type="button"
                                className="btn btn-outline-secondary"
                                onClick={() => onShowFullCategoryClicked(category)}
                            >
                                Show full category
                            </button>
                        </div>
                    </div>
                )
            }

            return (
                <div key={category} className="mt-1">
                    <h3
                        className="extensions-list__category mb-3 font-weight-normal"
                        data-test-extension-category-header={category}
                    >
                        {category}
                    </h3>
                    <div className="extensions-list__cards mt-1">
                        {extensionIDsForCategory.map(extensionId => (
                            <ExtensionCard
                                key={extensionId}
                                subject={subject}
                                viewerSubject={viewerSubject?.subject}
                                siteSubject={siteSubject?.subject}
                                node={extensions[extensionId]}
                                settingsCascade={settingsCascade}
                                platformContext={platformContext}
                                enabled={isExtensionEnabled(settingsCascade.final, extensionId)}
                                enabledForAllUsers={
                                    siteSubject ? isExtensionEnabled(siteSubject.settings, extensionId) : false
                                }
                                isLightTheme={props.isLightTheme}
                                settingsURL={authenticatedUser?.settingsURL}
                                authenticatedUser={authenticatedUser}
                            />
                        ))}
                    </div>
                </div>
            )
        })
    } else {
        // When a category is selected, display all extensions that include this category in their manifest,
        // not just extensions for which this is the primary category.
        const { allExtensionIDs } = extensionIDsByCategory[selectedCategory]

        let extensionIDs = allExtensionIDs

        if (enablementFilter !== 'all') {
            extensionIDs = allExtensionIDs.filter(
                extensionID =>
                    (enablementFilter === 'enabled') === isExtensionEnabled(settingsCascade.final, extensionID)
            )
        }

        categorySections = [
            <div key={selectedCategory} className="mt-1">
                <h3
                    className="extensions-list__category mb-3 font-weight-normal"
                    data-test-extension-category-header={selectedCategory}
                >
                    {selectedCategory}
                </h3>
                <div className="extensions-list__cards mt-1">
                    {extensionIDs.map(extensionId => (
                        <ExtensionCard
                            key={extensionId}
                            subject={subject}
                            viewerSubject={viewerSubject?.subject}
                            siteSubject={siteSubject?.subject}
                            node={extensions[extensionId]}
                            settingsCascade={settingsCascade}
                            platformContext={platformContext}
                            enabled={isExtensionEnabled(settingsCascade.final, extensionId)}
                            enabledForAllUsers={
                                siteSubject ? isExtensionEnabled(siteSubject.settings, extensionId) : false
                            }
                            isLightTheme={props.isLightTheme}
                            settingsURL={authenticatedUser?.settingsURL}
                            authenticatedUser={authenticatedUser}
                        />
                    ))}
                </div>
            </div>,
        ]
    }

    return (
        <>
            {error && <ErrorAlert className="mb-2" error={error} />}
            {categorySections.length > 0 ? (
                categorySections
            ) : (
                <div className="text-muted">
                    No extensions match <strong>{query}</strong> in the selected categories.
                </div>
            )}
        </>
    )
}
