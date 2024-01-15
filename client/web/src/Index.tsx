import type { FC } from 'react'

import { Navigate } from 'react-router-dom'

import { PageRoutes } from './routes.constants'
import { isCodyOnlyLicense } from './util/license'

interface IndexPageProps {}

export const IndexPage: FC<IndexPageProps> = () => {
    if (isCodyOnlyLicense()) {
        return <Navigate replace={true} to={PageRoutes.Cody} />
    }

    // TODO(BolajiOlajide): on a codesearch + cody license, redirect to the last product (between code search and cody)
    // the user visited.
    return <Navigate replace={true} to={PageRoutes.Search} />
}
