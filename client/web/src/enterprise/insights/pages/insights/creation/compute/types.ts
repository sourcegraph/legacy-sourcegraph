import { EditableDataSeries } from '../../../../components'

export enum ComputeInsightMap {
    Repositories = 'repositories',
    Path = 'path',
    Author = 'author',
    Date = 'date',
}

export interface CreateComputeInsightFormFields {
    /**
     * Code Insight series setting (name of line, line query, color)
     */
    series: EditableDataSeries[]

    /**
     * Title of code insight
     */
    title: string

    /**
     * Repositories which to be used to get the info for code insights
     */
    repositories: string

    /**
     * The total number of dashboards on which this insight is referenced.
     */
    dashboardReferenceCount: number

    groupBy: ComputeInsightMap
}
