import React from 'react'

import { Meta } from '@storybook/react'

import { WebStory } from '../../../components/WebStory'
import { Series, SeriesBasedChartTypes } from '../../types'

import { SeriesChart } from './SeriesChart'

export default {
    title: 'web/charts/SeriesChart',
    decorators: [story => <WebStory>{() => story()}</WebStory>],
} as Meta

interface StandardDatum {
    a: number | null
    aLink: string
    b: number | null
    bLink: string
    c: number | null
    cLink: string
    x: number | null
}

const DATA: StandardDatum[] = [
    {
        x: 1588965700286 - 4 * 24 * 60 * 60 * 1000,
        a: 4000,
        aLink: 'https://google.com/search',
        b: 15000,
        bLink: 'https://yandex.com/search',
        c: 5000,
        cLink: 'https://twitter.com/search',
    },
    {
        x: 1588965700286 - 3 * 24 * 60 * 60 * 1000,
        a: 4000,
        aLink: 'https://google.com/search',
        b: 26000,
        bLink: 'https://yandex.com/search',
        c: 5000,
        cLink: 'https://twitter.com/search',
    },
    {
        x: 1588965700286 - 2 * 24 * 60 * 60 * 1000,
        a: 5600,
        aLink: 'https://google.com/search',
        b: 20000,
        bLink: 'https://yandex.com/search',
        c: 5000,
        cLink: 'https://twitter.com/search',
    },
    {
        x: 1588965700286 - 1 * 24 * 60 * 60 * 1000,
        a: 9800,
        aLink: 'https://google.com/search',
        b: 19000,
        bLink: 'https://yandex.com/search',
        c: 5000,
        cLink: 'https://twitter.com/search',
    },
    {
        x: 1588965700286,
        a: 6000,
        aLink: 'https://google.com/search',
        b: 17000,
        bLink: 'https://yandex.com/search',
        c: 5000,
        cLink: 'https://twitter.com/search',
    },
]

const SERIES: Series<StandardDatum>[] = [
    {
        dataKey: 'a',
        name: 'A metric',
        color: 'var(--blue)',
        getLinkURL: datum => datum.aLink,
    },
    {
        dataKey: 'b',
        name: 'B metric',
        color: 'var(--warning)',
        getLinkURL: datum => datum.bLink,
    },
    {
        dataKey: 'c',
        name: 'C metric',
        color: 'var(--green)',
        getLinkURL: datum => datum.cLink,
    },
]

export const SeriesLineChart = () => (
    <SeriesChart
        type={SeriesBasedChartTypes.Line}
        data={DATA}
        series={SERIES}
        stacked={false}
        xAxisKey="x"
        width={400}
        height={400}
    />
)
