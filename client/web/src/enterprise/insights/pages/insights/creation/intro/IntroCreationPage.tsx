import { ParentSize } from '@visx/responsive'
import classNames from 'classnames'
import React, { useEffect } from 'react'
import { Link } from 'react-router-dom'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { Page } from '../../../../../../components/Page'
import { PageTitle } from '../../../../../../components/PageTitle'
import { LineChart } from '../../../../../../views/components/view/content/chart-view-content/charts/line/LineChart'
import { PieChart } from '../../../../../../views/components/view/content/chart-view-content/charts/pie/PieChart'
import { LinkWithQuery } from '../../../../components/link-with-query'

import { LINE_CHART_DATA, PIE_CHART_DATA } from './charts-mock'
import styles from './IntroCreationPage.module.scss'

interface IntroCreationPageProps extends TelemetryProps {}

/** Displays intro page for insights creation UI. */
export const IntroCreationPage: React.FunctionComponent<IntroCreationPageProps> = props => {
    const { telemetryService } = props

    const logCreateSearchBasedInsightClick = (): void => {
        telemetryService.log('CodeInsightsCreateSearchBasedInsightClick')
    }

    const logCreateCodeStatsInsightClick = (): void => {
        telemetryService.log('CodeInsightsCreateCodeStatsInsightClick')
    }

    const logExploreExtensionsClick = (): void => {
        telemetryService.log('CodeInsightsExploreInsightExtensionsClick')
    }

    useEffect(() => {
        telemetryService.logViewEvent('CodeInsightsCreationPage')
    }, [telemetryService])

    return (
        <Page className="col-8">
            <PageTitle title="Create code insights" />

            <div className="mb-5">
                <h2>Create new insight</h2>

                <p className="text-muted">
                    Code insights analyze your code based on any search query.{' '}
                    <a href="https://docs.sourcegraph.com/code_insights" target="_blank" rel="noopener">
                        Learn more.
                    </a>
                </p>
            </div>

            <div className={classNames(styles.createIntroPageInsights, 'pb-5')}>
                <section
                    className={classNames(styles.createIntroPageInsightCard, 'card card-body p-3')}
                    data-testid="create-search-insights"
                >
                    <h3>Based on your search query</h3>

                    <p>
                        Search-based insights let you create a time series data visualization about your code based on a
                        custom search query.
                    </p>

                    <LinkWithQuery
                        to="/insights/create/search"
                        onClick={logCreateSearchBasedInsightClick}
                        className={classNames(styles.createIntroPageInsightButton, 'btn', 'btn-primary')}
                    >
                        Create search insight
                    </LinkWithQuery>

                    <hr className="ml-n3 mr-n3 mt-4 mb-3" />

                    <p className="text-muted">Example:</p>
                    <div className={styles.createIntroPageChartContainer}>
                        <ParentSize className={styles.createIntroPageChart}>
                            {({ width, height }) => <LineChart width={width} height={height} {...LINE_CHART_DATA} />}
                        </ParentSize>
                    </div>
                </section>

                <section
                    className={classNames(styles.createIntroPageInsightCard, 'card card-body p-3')}
                    data-testid="create-lang-usage-insight"
                >
                    <h3>Language usage</h3>

                    <p>Shows language usage in your repository by lines of code.</p>

                    <LinkWithQuery
                        to="/insights/create/lang-stats"
                        onClick={logCreateCodeStatsInsightClick}
                        className={classNames(styles.createIntroPageInsightButton, 'btn', 'btn-primary')}
                    >
                        Create language usage insight
                    </LinkWithQuery>

                    <hr className="ml-n3 mr-n3 mt-4 mb-3" />

                    <p className="text-muted">Example:</p>
                    <div className={styles.createIntroPageChartContainer}>
                        <ParentSize className={styles.createIntroPageChart}>
                            {({ width, height }) => <PieChart width={width} height={height} {...PIE_CHART_DATA} />}
                        </ParentSize>
                    </div>
                </section>

                <section
                    className={classNames(styles.createIntroPageInsightCard, 'card card-body p-3')}
                    data-testid="explore-extensions"
                >
                    <h3>Based on Sourcegraph extensions</h3>

                    <p>
                        Enable an extension that creates code insights, then follow its README.md to learn how to set up
                        code insights for that extension.
                    </p>

                    <Link
                        to="/extensions?query=category:Insights&experimental=true"
                        onClick={logExploreExtensionsClick}
                        className={classNames(styles.createIntroPageInsightButton, 'btn', 'btn-secondary')}
                    >
                        Explore the extensions
                    </Link>
                </section>
            </div>
        </Page>
    )
}
