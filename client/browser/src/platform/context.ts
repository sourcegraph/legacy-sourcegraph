import { combineLatest, merge, Observable, ReplaySubject } from 'rxjs'
import { map, mergeMap, publishReplay, refCount, switchMap, take } from 'rxjs/operators'
import { GraphQLResult } from '../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../shared/src/graphql/schema'
import { PlatformContext } from '../../../../shared/src/platform/context'
import { mutateSettings, updateSettings } from '../../../../shared/src/settings/edit'
import { EMPTY_SETTINGS_CASCADE, gqlToCascade } from '../../../../shared/src/settings/settings'
import { LocalStorageSubject } from '../../../../shared/src/util/LocalStorageSubject'
import { toPrettyBlobURL } from '../../../../shared/src/util/url'
import { ExtensionStorageSubject } from '../browser/ExtensionStorageSubject'
import { observeStorageKey } from '../browser/storage'
import { defaultStorageItems } from '../browser/types'
import { isInPage } from '../context'
import { CodeHost } from '../libs/code_intelligence'
import { getContext } from '../shared/backend/context'
import { requestGraphQL } from '../shared/backend/graphql'
import { sourcegraphUrl } from '../shared/util/context'
import { createExtensionHost } from './extensionHost'
import { editClientSettings, fetchViewerSettings, mergeCascades, storageSettingsCascade } from './settings'
import { createBlobURLForBundle } from './worker'

/**
 * Creates the {@link PlatformContext} for the browser extension.
 */
export function createPlatformContext({ urlToFile }: Pick<CodeHost, 'urlToFile'>): PlatformContext {
    const updatedViewerSettings = new ReplaySubject<Pick<GQL.ISettingsCascade, 'subjects' | 'final'>>(1)

    const context: PlatformContext = {
        /**
         * The active settings cascade.
         *
         * - For unauthenticated users, this is the GraphQL settings plus client settings (which are stored locally
         *   in the browser extension.
         * - For authenticated users, this is just the GraphQL settings (client settings are ignored to simplify
         *   the UX).
         */
        settings: combineLatest(
            merge(
                isInPage
                    ? fetchViewerSettings()
                    : observeStorageKey('sync', 'sourcegraphURL').pipe(switchMap(() => fetchViewerSettings())),
                updatedViewerSettings
            ).pipe(
                publishReplay(1),
                refCount()
            ),
            storageSettingsCascade
        ).pipe(
            map(([gqlCascade, storageCascade]) =>
                mergeCascades(
                    gqlToCascade(gqlCascade),
                    gqlCascade.subjects.some(subject => subject.__typename === 'User')
                        ? EMPTY_SETTINGS_CASCADE
                        : storageCascade
                )
            )
        ),
        updateSettings: async (subject, edit) => {
            if (subject === 'Client') {
                // Support storing settings on the client (in the browser extension) so that unauthenticated
                // Sourcegraph viewers can update settings.
                await updateSettings(context, subject, edit, () => editClientSettings(edit))
                return
            }

            try {
                await updateSettings(context, subject, edit, mutateSettings)
            } catch (error) {
                if ('message' in error && /version mismatch/.test(error.message)) {
                    // The user probably edited the settings in another tab, so
                    // try once more.
                    updatedViewerSettings.next(await fetchViewerSettings().toPromise())
                    await updateSettings(context, subject, edit, mutateSettings)
                }
            }

            updatedViewerSettings.next(await fetchViewerSettings().toPromise())
        },
        queryGraphQL: (request, variables, requestMightContainPrivateInfo) => {
            if (isInPage) {
                return requestGraphQL({
                    ctx: getContext({ repoKey: '', isRepoSpecific: false }),
                    request,
                    variables,
                    url: window.SOURCEGRAPH_URL,
                    requestMightContainPrivateInfo,
                })
            }

            return observeStorageKey('sync', 'sourcegraphURL').pipe(
                take(1),
                mergeMap(
                    (url: string = defaultStorageItems.sourcegraphURL): Observable<GraphQLResult<any>> =>
                        requestGraphQL({
                            ctx: getContext({ repoKey: '', isRepoSpecific: false }),
                            request,
                            variables,
                            url,
                            requestMightContainPrivateInfo,
                        })
                )
            )
        },
        forceUpdateTooltip: () => {
            // TODO(sqs): implement tooltips on the browser extension
        },
        createExtensionHost,
        getScriptURLForExtension: async bundleURL => {
            if (isInPage) {
                return await createBlobURLForBundle(bundleURL)
            }

            // We need to import the extension's JavaScript file (in importScripts in the Web Worker) from a blob:
            // URI, not its original http:/https: URL, because Chrome extensions are not allowed to be published
            // with a CSP that allowlists https://* in script-src (see
            // https://developer.chrome.com/extensions/contentSecurityPolicy#relaxing-remote-script). (Firefox
            // add-ons have an even stricter restriction.)
            const blobURL: string = await browser.runtime.sendMessage({
                type: 'createBlobURL',
                payload: bundleURL,
            })

            return blobURL
        },
        urlToFile: location => {
            if (urlToFile) {
                // Construct URL to file on code host, if possible.
                return urlToFile(location)
            }
            // Otherwise fall back to linking to Sourcegraph (with an absolute URL).
            return `${sourcegraphUrl}${toPrettyBlobURL(location)}`
        },
        sourcegraphURL: sourcegraphUrl,
        clientApplication: 'other',
        sideloadedExtensionURL: isInPage
            ? new LocalStorageSubject<string | null>('sideloadedExtensionURL', null)
            : new ExtensionStorageSubject('sideloadedExtensionURL', null),
    }
    return context
}
