import { useHistory } from 'react-router-dom'

import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { asError } from '@sourcegraph/shared/src/util/errors'

import { Settings } from '../../../../../schema/settings.schema'
import { FORM_ERROR, SubmissionErrors } from '../../../../components/form/hooks/useForm'
import { Insight, isVirtualDashboard } from '../../../../core/types'
import { useDashboard } from '../../../../hooks/use-dashboard'
import { usePersistEditOperations } from '../../../../hooks/use-persist-edit-operations'
import { useQueryParameters } from '../../../../hooks/use-query-parameters'

import { getUpdatedSubjectSettings } from './utils'

export interface UseHandleSubmitProps extends PlatformContextProps<'updateSettings'>, SettingsCascadeProps<Settings> {
    originalInsight: Insight | null
}

export interface useHandleSubmitOutput {
    handleSubmit: (newInsight: Insight) => Promise<SubmissionErrors>
    handleCancel: () => void
}

/**
 * Returns submit and cancel handlers for the edit submit page.
 */
export function usePageHandlers(props: UseHandleSubmitProps): useHandleSubmitOutput {
    const { originalInsight, platformContext, settingsCascade } = props
    const history = useHistory()
    const { dashboardId } = useQueryParameters(['dashboardId'])
    const dashboard = useDashboard({ settingsCascade, dashboardId })

    const { persist } = usePersistEditOperations({ platformContext })

    const handleSubmit = async (newInsight: Insight): Promise<SubmissionErrors> => {
        if (!originalInsight) {
            return
        }

        try {
            const editOperations = getUpdatedSubjectSettings({
                oldInsight: originalInsight,
                newInsight,
                settingsCascade,
            })

            await persist(editOperations)

            if (!dashboard || isVirtualDashboard(dashboard)) {
                // Navigate user to the dashboard page with new created dashboard
                history.push(`/insights/dashboards/${newInsight.visibility}`)

                return
            }

            // If insight's visible area has been changed explicit redirect to new
            // scope dashboard page
            if (dashboard.owner.id !== newInsight.visibility) {
                history.push(`/insights/dashboards/${newInsight.visibility}`)
            } else {
                history.push(`/insights/dashboards/${dashboard.id}`)
            }
        } catch (error) {
            return { [FORM_ERROR]: asError(error) }
        }

        return
    }

    const handleCancel = (): void => {
        history.push(`/insights/dashboards/${dashboard?.id ?? 'all'}`)
    }

    return { handleSubmit, handleCancel }
}
