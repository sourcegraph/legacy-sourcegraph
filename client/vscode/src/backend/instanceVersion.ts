import { gql } from '@sourcegraph/http-client'
import { EventSource } from '@sourcegraph/shared/src/graphql-operations'

import { INSTANCE_VERSION_NUMBER_KEY, LocalStorageService } from '../settings/LocalStorageService'

import { requestGraphQLFromVSCode } from './requestGraphQl'

export function initializeInstantVersionNumber({
    localStorageService,
}: {
    localStorageService: LocalStorageService
}): EventSource {
    requestGraphQLFromVSCode<SiteVersionResult>(siteVersionQuery, {})
        .then(async siteVersionResult => {
            if (siteVersionResult.data) {
                await localStorageService.setValue(
                    INSTANCE_VERSION_NUMBER_KEY,
                    siteVersionResult.data.site.productVersion
                )
            }
        })
        .catch(error => {
            console.error('Failed to get instance version from host:', error)
        })
    /**
     * Regular instance version format: 3.38.2
     * Insider version format: 134683_2022-03-02_5188fes0101
     */
    const versionNumber = localStorageService.getValue(INSTANCE_VERSION_NUMBER_KEY)
    // assume instance version longer than 8 is using insider version
    const flattenVersion = versionNumber.length > 8 ? '999999' : versionNumber.split('.').join()
    // instances below 3.38.0 does not support EventSource.IDEEXTENSION
    return flattenVersion > '3380' ? EventSource.IDEEXTENSION : EventSource.BACKEND
}

const siteVersionQuery = gql`
    query {
        site {
            productVersion
        }
    }
`
interface SiteVersionResult {
    site: {
        productVersion: string
    }
}
