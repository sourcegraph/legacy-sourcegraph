import { useMemo, FC, PropsWithChildren } from 'react'

import classNames from 'classnames'

import { SearchPatternType } from '@sourcegraph/shared/src/graphql-operations'
import { buildSearchURLQuery } from '@sourcegraph/shared/src/util/url'
import { Button, Link } from '@sourcegraph/wildcard'

import styles from './MonacoPreviewLink.module.scss'

export interface MonacoPreviewLinkProps {
    query: string
    patternType: SearchPatternType
    className?: string
}

export const MonacoPreviewLink: FC<PropsWithChildren<MonacoPreviewLinkProps>> = props => {
    const { query, patternType, className, children } = props
    const queryURL = useMemo(() => `/search?${buildSearchURLQuery(query, patternType, false)}`, [patternType, query])

    return (
        <Button
            className={classNames(styles.previewLink, className)}
            to={queryURL}
            variant="link"
            as={Link}
            target="_blank"
            rel="noopener noreferrer"
        >
            {children}
        </Button>
    )
}
