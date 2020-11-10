import * as React from 'react'
import CalculatorIcon from 'mdi-react/CalculatorIcon'
import classNames from 'classnames'
import { pluralize } from '../../../../../shared/src/util/strings'
import { StreamingProgressProps } from './StreamingProgress'

export const StreamingProgressCount: React.FunctionComponent<StreamingProgressProps> = ({ progress }) => (
    <div
        className={classNames('streaming-progress__count p-2 d-flex align-items-center', {
            'streaming-progress__count--in-progress': !progress.done,
        })}
    >
        <CalculatorIcon className="mr-2" />
        {progress.matchCount} {pluralize('result', progress.matchCount)} in {(progress.durationMs / 1000).toFixed(2)}s
        {progress.repositoriesCount && (
            <>
                {' '}
                from {progress.repositoriesCount} {pluralize('repository', progress.repositoriesCount, 'repositories')}
            </>
        )}
    </div>
)
