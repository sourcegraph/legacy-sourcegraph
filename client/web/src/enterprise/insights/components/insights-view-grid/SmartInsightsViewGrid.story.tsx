import { Meta } from '@storybook/react'
import React from 'react'
import { of } from 'rxjs'

import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { WebStory } from '../../../../components/WebStory'
import {
    LINE_CHART_TESTS_CASES_EXAMPLE,
    LINE_CHART_WITH_HUGE_NUMBER_OF_LINES,
    LINE_CHART_WITH_MANY_LINES,
} from '../../../../views/mocks/charts-content'
import { CodeInsightsBackendContext } from '../../core/backend/code-insights-backend-context'
import { CodeInsightsGqlBackend } from '../../core/backend/gql-api/code-insights-gql-backend'
import { BackendInsight, Insight, InsightExecutionType, InsightType, isCaptureGroupInsight } from '../../core/types'

import { SmartInsightsViewGrid } from './SmartInsightsViewGrid'

export default {
    title: 'web/insights/SmartInsightsViewGridExample',
    decorators: [story => <WebStory>{() => story()}</WebStory>],
    parameters: {
        chromatic: {
            viewports: [576, 1440],
            enableDarkMode: true,
        },
    },
} as Meta

const insightsWithManyLines: Insight[] = [
    {
        id: 'searchInsights.insight.Backend_1',
        type: InsightExecutionType.Backend,
        viewType: InsightType.SearchBased,
        title: 'Backend insight #2',
        series: [{ id: '', query: '', stroke: '', name: '' }],
        visibility: 'personal',
        step: { weeks: 2 },
        filters: { excludeRepoRegexp: '', includeRepoRegexp: '' },
        dashboardReferenceCount: 0,
    },
    {
        id: 'searchInsights.insight.Backend_2',
        type: InsightExecutionType.Backend,
        viewType: InsightType.SearchBased,
        title: 'Backend insight #3',
        series: [],
        visibility: 'personal',
        step: { weeks: 2 },
        filters: { excludeRepoRegexp: '', includeRepoRegexp: '' },
        dashboardReferenceCount: 0,
    },
    {
        id: 'searchInsights.insight.Backend_3',
        type: InsightExecutionType.Backend,
        viewType: InsightType.SearchBased,
        title: 'Backend insight #1',
        series: [
            { id: '', query: '', stroke: '', name: '' },
            { id: '', query: '', stroke: '', name: '' },
            { id: '', query: '', stroke: '', name: '' },
            { id: '', query: '', stroke: '', name: '' },
            { id: '', query: '', stroke: '', name: '' },
            { id: '', query: '', stroke: '', name: '' },
        ],
        visibility: 'personal',
        step: { weeks: 2 },
        filters: { excludeRepoRegexp: '', includeRepoRegexp: '' },
        dashboardReferenceCount: 0,
    },
    {
        id: 'searchInsights.insight.Backend_4',
        type: InsightExecutionType.Backend,
        viewType: InsightType.SearchBased,
        title: 'Backend insight #2',
        series: [{ id: '', query: '', stroke: '', name: '' }],
        visibility: 'personal',
        step: { weeks: 2 },
        filters: { excludeRepoRegexp: '', includeRepoRegexp: '' },
        dashboardReferenceCount: 0,
    },
    {
        id: 'searchInsights.insight.Backend_5',
        type: InsightExecutionType.Backend,
        viewType: InsightType.SearchBased,
        title: 'Backend insight #2',
        series: [
            { id: '', query: '', stroke: '', name: '' },
            { id: '', query: '', stroke: '', name: '' },
            { id: '', query: '', stroke: '', name: '' },
            { id: '', query: '', stroke: '', name: '' },
            { id: '', query: '', stroke: '', name: '' },
            { id: '', query: '', stroke: '', name: '' },
            { id: '', query: '', stroke: '', name: '' },
            { id: '', query: '', stroke: '', name: '' },
            { id: '', query: '', stroke: '', name: '' },
            { id: '', query: '', stroke: '', name: '' },
            { id: '', query: '', stroke: '', name: '' },
            { id: '', query: '', stroke: '', name: '' },
            { id: '', query: '', stroke: '', name: '' },
            { id: '', query: '', stroke: '', name: '' },
            { id: '', query: '', stroke: '', name: '' },
            { id: '', query: '', stroke: '', name: '' },
            { id: '', query: '', stroke: '', name: '' },
            { id: '', query: '', stroke: '', name: '' },
            { id: '', query: '', stroke: '', name: '' },
            { id: '', query: '', stroke: '', name: '' },
        ],
        visibility: 'personal',
        step: { weeks: 2 },
        filters: { excludeRepoRegexp: '', includeRepoRegexp: '' },
        dashboardReferenceCount: 0,
    },
    {
        id: 'searchInsights.insight.Backend_6',
        type: InsightExecutionType.Backend,
        viewType: InsightType.SearchBased,
        title: 'Backend insight #2',
        series: [{ id: '', query: '', stroke: '', name: '' }],
        visibility: 'personal',
        step: { weeks: 2 },
        filters: { excludeRepoRegexp: '', includeRepoRegexp: '' },
        dashboardReferenceCount: 0,
    },
    {
        id: 'searchInsights.insight.Backend_7',
        type: InsightExecutionType.Backend,
        viewType: InsightType.SearchBased,
        title: 'Backend insight #2',
        series: [{ id: '', query: '', stroke: '', name: '' }],
        visibility: 'personal',
        step: { weeks: 2 },
        filters: { excludeRepoRegexp: '', includeRepoRegexp: '' },
        dashboardReferenceCount: 0,
    },
]

class StoryBackendWithManyLinesCharts extends CodeInsightsGqlBackend {
    constructor() {
        super({} as any)
    }

    public getBackendInsightData = (insight: BackendInsight) => {
        if (isCaptureGroupInsight(insight)) {
            throw new Error('This demo does not support capture group insight')
        }

        return of({
            id: insight.id,
            view: {
                title: 'Backend Insight Mock',
                subtitle: 'Backend insight description text',
                content: [
                    insight.series.length >= 6
                        ? insight.series.length >= 15
                            ? LINE_CHART_WITH_HUGE_NUMBER_OF_LINES
                            : LINE_CHART_WITH_MANY_LINES
                        : LINE_CHART_TESTS_CASES_EXAMPLE,
                ],
                isFetchingHistoricalData: false,
            },
        })
    }
}

const codeInsightsApiWithManyLines = new StoryBackendWithManyLinesCharts()

export const SmartInsightsViewGridExample = () => (
    <CodeInsightsBackendContext.Provider value={codeInsightsApiWithManyLines}>
        <SmartInsightsViewGrid insights={insightsWithManyLines} telemetryService={NOOP_TELEMETRY_SERVICE} />
    </CodeInsightsBackendContext.Provider>
)
