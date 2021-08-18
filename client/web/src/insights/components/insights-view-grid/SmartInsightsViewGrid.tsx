import { isEqual } from 'lodash'
import React, { memo } from 'react'

import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { Settings } from '../../../schema/settings.schema'
import { Insight } from '../../core/types'

import { SmartInsight } from './components/smart-insight/SmartInsight'
import { ViewGrid } from './components/view-grid/ViewGrid'

interface SmartInsightsViewGridProps
    extends TelemetryProps,
        SettingsCascadeProps<Settings>,
        PlatformContextProps<'updateSettings'>,
        ExtensionsControllerProps {
    insights: Insight[]
}

/**
 * Custom props checker for the smart grid component.
 *
 * Ignore settings cascade change and insight body config changes to avoid
 * animations of grid item rerender and grid position items. In some cases (like insight
 * filters updating, we want to ignore insights from settings cascade).
 * But still trigger grid animation rerender if insight ordering or insight count
 * have been changed.
 */
const equalSmartGridProps = (
    previousProps: SmartInsightsViewGridProps,
    nextProps: SmartInsightsViewGridProps
): boolean => {
    const { insights: previousInsights, settingsCascade: previousSettingCascade, ...otherPrepProps } = previousProps
    const { insights: nextInsights, settingsCascade, ...otherNextProps } = nextProps

    if (!isEqual(otherPrepProps, otherNextProps)) {
        return false
    }

    return isEqual(
        previousInsights.map(insight => insight.id),
        nextInsights.map(insight => insight.id)
    )
}

/**
 * Renders grid of smart (stateful) insight card. These cards can independently extract and update
 * the insights settings (settings cascade subjects).
 */
export const SmartInsightsViewGrid: React.FunctionComponent<SmartInsightsViewGridProps> = memo(props => {
    const { telemetryService, insights, platformContext, settingsCascade, extensionsController } = props

    return (
        <ViewGrid viewIds={insights.map(insight => insight.id)} telemetryService={telemetryService}>
            {insights.map(insight => (
                <SmartInsight
                    key={insight.id}
                    insight={insight}
                    telemetryService={telemetryService}
                    platformContext={platformContext}
                    settingsCascade={settingsCascade}
                    extensionsController={extensionsController}
                />
            ))}
        </ViewGrid>
    )
}, equalSmartGridProps)
