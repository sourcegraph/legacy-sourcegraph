import React, { useEffect, useRef, useState } from 'react'

import CheckCircleIcon from 'mdi-react/CheckCircleIcon'
import ReactDOM from 'react-dom'
import { useLocation } from 'react-router-dom'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Icon } from '@sourcegraph/wildcard'

import { GETTING_STARTED_TOUR_MARKER } from './TourInfo'
import { TourTaskType, TourTaskStepType } from './types'
import { parseURIMarkers } from './utils'

import styles from './Tour.module.scss'

interface TourAgentProps extends TelemetryProps {
    tasks: TourTaskType[]
    onStepComplete: (step: TourTaskStepType) => void
}

export function useTourQueryParameters(): ReturnType<typeof parseURIMarkers> | undefined {
    const location = useLocation()
    const [data, setData] = useState<ReturnType<typeof parseURIMarkers> | undefined>()
    useEffect(() => {
        const { isTour, stepId } = parseURIMarkers(location.search)
        if (!isTour || !stepId) {
            setData(undefined)
        } else {
            setData({ isTour, stepId })
        }
    }, [location])

    return data
}

/**
 * Component to track TourTaskStepType.completeAfterEvents and show info box for steps.
 */
export const TourAgent: React.FunctionComponent<React.PropsWithChildren<TourAgentProps>> = React.memo(
    ({ tasks, telemetryService, onStepComplete }) => {
        // Agent 1: Track completion
        useEffect(() => {
            const filteredSteps = tasks.flatMap(task => task.steps).filter(step => step.completeAfterEvents)
            return telemetryService?.addEventLogListener?.(eventName => {
                const step = filteredSteps.find(step => step.completeAfterEvents?.includes(eventName))
                if (step) {
                    onStepComplete(step)
                }
            })
        }, [telemetryService, tasks, onStepComplete])

        // Agent 2: Track info panel
        const [info, setInfo] = useState<TourTaskStepType['info'] | undefined>()

        const tourQueryParameters = useTourQueryParameters()

        useEffect(() => {
            const info = tasks.flatMap(task => task.steps).find(step => tourQueryParameters?.stepId === step.id)?.info
            setInfo(info)
        }, [tasks, tourQueryParameters?.stepId])

        const domNode = useRef(document.querySelector('#' + GETTING_STARTED_TOUR_MARKER))
        useEffect(() => {
            if (info) {
                domNode.current?.classList.remove('d-none')
            } else {
                domNode.current?.classList.add('d-none')
            }
        }, [info])

        if (!info || !domNode.current) {
            return null
        }

        return ReactDOM.createPortal(
            <div className={styles.infoPanel}>
                <Icon role="img" as={CheckCircleIcon} className={styles.infoIcon} aria-hidden={true} />
                <span>{info}</span>
            </div>,
            domNode.current
        )
    }
)
