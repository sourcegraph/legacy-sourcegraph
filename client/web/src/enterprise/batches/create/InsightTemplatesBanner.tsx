import React from 'react'

import classNames from 'classnames'

import { Card, CardBody, H4 } from '@sourcegraph/wildcard'

import { CodeInsightsBatchesIcon } from './CodeInsightsBatchesIcon'

import styles from './InsightTemplatesBanner.module.scss'

export const InsightTemplatesBanner: React.FunctionComponent<{ insightTitle: string; type: 'create' | 'edit' }> = ({
    insightTitle,
    type,
}) => {
    const [heading, paragraph]: [React.ReactNode, React.ReactNode] =
        type === 'create'
            ? [
                  'You are creating a batch change from a code insight',
                  <>
                      Let Sourcegraph help you with <strong>{insightTitle}</strong> by preparing a relevant{' '}
                      <strong>batch change</strong>.
                  </>,
              ]
            : [
                  `Start from template for the ${insightTitle}`,
                  `Sourcegraph pre-selected a batch spec for the batch change started from ${insightTitle}.`,
              ]

    return (
        <Card className={classNames('mb-5', styles.banner)}>
            <CardBody>
                <div className="d-flex justify-content-between align-items-center">
                    <CodeInsightsBatchesIcon className="mr-4" />
                    <div className="flex-grow-1">
                        <H4>{heading}</H4>
                        <p className="mb-0">{paragraph}</p>
                    </div>
                </div>
            </CardBody>
        </Card>
    )
}
