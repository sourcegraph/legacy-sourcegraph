import { useCallback, useContext, useState } from 'react'

import { ErrorLike } from '@sourcegraph/common'

import { eventLogger } from '../../../tracking/eventLogger'
import { CodeInsightsBackendContext, Insight, InsightDashboard } from '../core'
import { getTrackingTypeByInsightType } from '../pings'

interface RemoveInsightInput {
    insight: Pick<Insight, 'id' | 'title' | 'type'>
    dashboard: Pick<InsightDashboard, 'id' | 'title'>
}

const sleep = (milliseconds: number): Promise<void> => new Promise(resolve => setTimeout(resolve, milliseconds))

export interface useRemoveInsightFromDashboardAPI {
    remove: (insight: RemoveInsightInput) => Promise<void>
    loading: boolean
    error: ErrorLike | undefined
}

export function useRemoveInsightFromDashboard(): useRemoveInsightFromDashboardAPI {
    const { removeInsightFromDashboard } = useContext(CodeInsightsBackendContext)

    const [loading, setLoading] = useState<boolean>(false)
    const [error, setError] = useState<ErrorLike | undefined>()

    const handleRemove = useCallback(
        async (input: RemoveInsightInput) => {
            const { insight, dashboard } = input

            // Prevent double call if we already have ongoing request
            if (loading) {
                return
            }

            setLoading(true)
            setError(undefined)

            try {
                await sleep(5000)
                await removeInsightFromDashboard({
                    insightId: insight.id,
                    dashboardId: dashboard.id,
                }).toPromise()

                const insightType = getTrackingTypeByInsightType(insight.type)

                eventLogger.log('InsightRemovalFromDashboard', { insightType }, { insightType })
            } catch (error) {
                // TODO [VK] Improve error UI for removing
                console.error(error)
                setError(error)
            }
        },
        [loading, removeInsightFromDashboard]
    )

    return { remove: handleRemove, loading, error }
}
