import classNames from 'classnames'
import * as React from 'react'

import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { useRedesignToggle } from '@sourcegraph/shared/src/util/useRedesignToggle'

interface ConnectionNodesSummaryProps {
    summary: React.ReactFragment | undefined
    displayShowMoreButton?: boolean
    onShowMore?: () => void
    showMoreClassName?: string
    loading?: boolean
}

const Loading: React.FunctionComponent = () => (
    <span className="filtered-connection__loader test-filtered-connection__loader">
        <LoadingSpinner className="icon-inline" />
    </span>
)

export const ConnectionNodesSummary: React.FunctionComponent<ConnectionNodesSummaryProps> = ({
    summary,
    displayShowMoreButton,
    showMoreClassName,
    onShowMore,
    loading,
}) => {
    const [isRedesignEnabled] = useRedesignToggle()

    const showMoreButton = displayShowMoreButton && (
        <button
            type="button"
            className={classNames(
                'btn btn-sm filtered-connection__show-more',
                isRedesignEnabled ? 'btn-link' : 'btn-secondary',
                showMoreClassName
            )}
            onClick={onShowMore}
        >
            Show more
        </button>
    )

    return (
        <div className="filtered-connection__summary-container">
            {loading ? (
                <Loading />
            ) : (
                <>
                    {summary}
                    {showMoreButton}
                </>
            )}
        </div>
    )
}
