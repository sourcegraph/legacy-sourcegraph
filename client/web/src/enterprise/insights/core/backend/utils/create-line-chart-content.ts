import { LineChartContent } from 'sourcegraph'

import { InsightDataSeries } from '../../../../../graphql-operations'
import { SearchBasedInsightSeries } from '../../types/insight/search-insight'

interface SeriesDataset {
    dateTime: number
    [seriesKey: string]: number
}

export interface InsightData {
    series: {
        label: string
        points: { dateTime: string; value: number }[]
    }[]
}

export function createLineChartContent(
    seriesData: InsightData,
    seriesDefinition: SearchBasedInsightSeries[] = []
): LineChartContent<SeriesDataset, 'dateTime'> {
    // Immutable sort is required to avoid breaking useCallback memoziation in BackendInsight component
    const sortedSeriesSettings = [...seriesDefinition].sort((a, b) => a.query.localeCompare(b.query))
    const dataByXValue = new Map<string, SeriesDataset>()

    for (const [seriesIndex, series] of seriesData.series.entries()) {
        for (const point of series.points) {
            let dataObject = dataByXValue.get(point.dateTime)
            if (!dataObject) {
                dataObject = {
                    dateTime: Date.parse(point.dateTime),
                    // Initialize all series to null (empty chart) value
                    ...Object.fromEntries(seriesData.series.map((series, index) => [`series${index}`, null])),
                }
                dataByXValue.set(point.dateTime, dataObject)
            }
            dataObject[`series${seriesIndex}`] = point.value
        }
    }

    return {
        chart: 'line',
        data: [...dataByXValue.values()],
        series: seriesData.series.map((series, index) => ({
            name: sortedSeriesSettings[index]?.name ?? series.label,
            dataKey: `series${index}`,
            stroke: sortedSeriesSettings[index]?.stroke,
        })),
        xAxis: {
            dataKey: 'dateTime',
            scale: 'time',
            type: 'number',
        },
    }
}

/**
 * Minimal input type model for {@link createLineChartContentFromIndexedSeries} function
 */
export type InsightDataSeriesData = Pick<InsightDataSeries, 'seriesId' | 'label' | 'points'>

/**
 * Generates line chart content for visx chart. Note that this function relies on the fact that
 * all series are indexed. This generator is used only for GQL api, only there we have indexed series
 * for setting-based api see {@link createLineChartContent}
 *
 * @param series - insight series with points data
 * @param seriesDefinition - insight definition with line settings (color, name, query)
 */
export function createLineChartContentFromIndexedSeries(
    series: InsightDataSeriesData[],
    seriesDefinition: SearchBasedInsightSeries[] = []
): LineChartContent<SeriesDataset, 'dateTime'> {
    const dataByXValue = new Map<string, SeriesDataset>()
    const definitionMap = Object.fromEntries<SearchBasedInsightSeries>(
        seriesDefinition.map(definition => [definition.id ?? '', definition])
    )

    for (const line of series) {
        for (const point of line.points) {
            let dataObject = dataByXValue.get(point.dateTime)
            if (!dataObject) {
                dataObject = {
                    dateTime: Date.parse(point.dateTime),
                    // Initialize all series to null (empty chart) value
                    ...Object.fromEntries(series.map(line => [line.seriesId, null])),
                }
                dataByXValue.set(point.dateTime, dataObject)
            }
            dataObject[line.seriesId] = point.value
        }
    }

    return {
        chart: 'line',
        data: [...dataByXValue.values()],
        series: series.map(line => ({
            name: definitionMap[line.seriesId]?.name ?? line.label,
            dataKey: line.seriesId,
            stroke: definitionMap[line.seriesId]?.stroke,
        })),
        xAxis: {
            dataKey: 'dateTime',
            scale: 'time',
            type: 'number',
        },
    }
}
