import { open } from '@tauri-apps/api/dialog'

import type { GetLocalCodeHostsResult, LocalRepository } from '../../../graphql-operations'

export interface LocalCodeHost {
    id: string
    path: string
    autogenerated: boolean
    isFolder: boolean
    repositories: LocalRepository[]
}

/**
 * Parses gql response and returns paths of all non-autogenerated local services.
 */
export function getLocalServicePaths(data?: GetLocalCodeHostsResult): string[] {
    const localCodeHosts = getLocalServices(data)

    if (!localCodeHosts) {
        return []
    }

    return localCodeHosts.map(item => item.path)
}

/**
 * Returns the local services that have been created manually by user in the setup wizard,
 * it ignores autogenerated on the backend local external services
 */
export function getLocalServices(data?: GetLocalCodeHostsResult, isAutogenerated?: boolean): LocalCodeHost[] {
    if (!data) {
        return []
    }

    return (
        data.localExternalServices
            .filter(
                service => (isAutogenerated ? service.autogenerated : !service.autogenerated)
                // TODO: Determine folder/single repo on the server
            )
            .map(service => ({ ...service, isFolder: service.repositories.length !== 1 })) ?? []
    )
}

/**
 * Generates minimal code host configuration for the local repositories code host.
 */
export function createDefaultLocalServiceConfig(path: string): string {
    return `{ "url":"${window.context.srcServeGitUrl}", "root": "${path}", "repos": ["src-serve-local"] }`
}

type Path = string

interface OpenDialogSettings {
    multiple?: boolean
}

/**
 * Calls native file picker window, returns list of picked files paths.
 * In case if picker was closed/canceled returns null
 */
export async function callFilePicker(settings?: OpenDialogSettings): Promise<Path[] | null> {
    const selected = await open({
        directory: true,
        multiple: settings?.multiple ?? true,
    })

    if (Array.isArray(selected)) {
        return selected
    }

    if (selected !== null) {
        return [selected]
    }

    return null
}
