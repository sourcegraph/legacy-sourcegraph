import React, { useCallback, useEffect, useMemo, useState } from 'react'
import { LocalStorageSubject } from '../../../../shared/src/util/LocalStorageSubject'
import { useObservable } from '../../../../shared/src/util/useObservable'
import PuzzleOutlineIcon from 'mdi-react/PuzzleOutlineIcon'
import ChevronDoubleUpIcon from 'mdi-react/ChevronDoubleUpIcon'
import { ButtonLink } from '../../../../shared/src/components/LinkOrButton'
import classNames from 'classnames'
import { ActionsContainer } from '../../../../shared/src/actions/ActionsContainer'
import { ContributableMenu } from '../../../../shared/src/api/protocol'
import { ExtensionsControllerProps } from '../../../../shared/src/extensions/controller'
import * as H from 'history'
import { PlatformContextProps } from '../../../../shared/src/platform/context'
import { TelemetryProps } from '../../../../shared/src/telemetry/telemetryService'
import { ActionItem, ActionItemAction } from '../../../../shared/src/actions/ActionItem'
import PlusIcon from 'mdi-react/PlusIcon'
import { Link } from 'react-router-dom'
import { Key } from 'ts-key-enum'
import { focusable } from 'tabbable'
import { head } from 'lodash'
import { useCarousel } from '../../components/useCarousel'
import MenuUpIcon from 'mdi-react/MenuUpIcon'
import MenuDownIcon from 'mdi-react/MenuDownIcon'
import { haveInitialExtensionsLoaded } from '../../../../shared/src/api/features'
import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'

// Action items bar and toggle are two separate components due to their placement in the DOM tree

const scrollButtonClassName = 'action-items__scroll'

export function useWebActionItems(): Pick<ActionItemsBarProps, 'useActionItemsBar'> &
    Pick<ActionItemsToggleProps, 'useActionItemsToggle'> {
    const toggles = useMemo(() => new LocalStorageSubject('action-items-bar-expanded', true), [])

    const [toggleReference, setToggleReference] = useState<HTMLElement | null>(null)
    const nextToggleReference = useCallback((toggle: HTMLElement) => {
        setToggleReference(toggle)
    }, [])

    const [barReference, setBarReference] = useState<HTMLElement | null>(null)
    const nextBarReference = useCallback((bar: HTMLElement) => {
        setBarReference(bar)
    }, [])

    // Set up keyboard navigation for distant toggle and bar. Remove previous event
    // listeners whenever references change.
    useEffect(() => {
        function onKeyDownToggle(event: KeyboardEvent): void {
            if (event.key === Key.ArrowDown && barReference) {
                const firstBarFocusable = head(focusable(barReference))
                if (firstBarFocusable) {
                    firstBarFocusable.focus()
                    event.preventDefault()
                }
            }
        }

        function onKeyDownBar(event: KeyboardEvent): void {
            if (event.target instanceof HTMLElement && toggleReference && barReference) {
                const focusableChildren = focusable(barReference).filter(
                    elm => !elm.classList.contains('disabled') && !elm.classList.contains(scrollButtonClassName)
                )
                const indexOfTarget = focusableChildren.indexOf(event.target)

                if (event.key === Key.ArrowDown) {
                    // If this is the last focusable element, go back to the toggle
                    if (indexOfTarget === focusableChildren.length - 1) {
                        toggleReference.focus()
                        event.preventDefault()
                        return
                    }

                    const itemToFocus = focusableChildren[indexOfTarget + 1]
                    if (itemToFocus instanceof HTMLElement) {
                        itemToFocus.focus()
                        event.preventDefault()
                        return
                    }
                }

                if (event.key === Key.ArrowUp) {
                    // If this is the first focusable element, go back to the toggle
                    if (indexOfTarget === 0) {
                        toggleReference.focus()
                        event.preventDefault()
                        return
                    }

                    const itemToFocus = focusableChildren[indexOfTarget - 1]
                    if (itemToFocus instanceof HTMLElement) {
                        itemToFocus.focus()
                        event.preventDefault()
                        return
                    }
                }
            }
        }

        toggleReference?.addEventListener('keydown', onKeyDownToggle)
        barReference?.addEventListener('keydown', onKeyDownBar)

        return () => {
            toggleReference?.removeEventListener('keydown', onKeyDownToggle)
            toggleReference?.removeEventListener('keydown', onKeyDownBar)
        }
    }, [toggleReference, barReference])

    const useActionItemsBar = useCallback(() => {
        // `useActionItemsBar` will be used as a hook
        // eslint-disable-next-line react-hooks/rules-of-hooks
        const isOpen = useObservable(toggles)

        return { isOpen, barReference: nextBarReference }
    }, [toggles, nextBarReference])

    const useActionItemsToggle = useCallback(() => {
        // `useActionItemsToggle` will be used as a hook
        // eslint-disable-next-line react-hooks/rules-of-hooks
        const isOpen = useObservable(toggles)

        // eslint-disable-next-line react-hooks/rules-of-hooks
        const toggle = useCallback(() => toggles.next(!isOpen), [isOpen])

        return { isOpen, toggle, toggleReference: nextToggleReference }
    }, [toggles, nextToggleReference])

    return {
        useActionItemsBar,
        useActionItemsToggle,
    }
}

export interface ActionItemsBarProps extends ExtensionsControllerProps, PlatformContextProps, TelemetryProps {
    useActionItemsBar: () => { isOpen: boolean | undefined; barReference: React.RefCallback<HTMLElement> }
    location: H.Location
}

export interface ActionItemsToggleProps extends ExtensionsControllerProps<'extHostAPI'> {
    useActionItemsToggle: () => {
        isOpen: boolean | undefined
        toggle: () => void
        toggleReference: React.RefCallback<HTMLElement>
    }
    className?: string
}

const actionItemClassName = 'action-items__action d-flex justify-content-center align-items-center text-decoration-none'

/**
 *
 */
export const ActionItemsBar = React.memo<ActionItemsBarProps>(props => {
    const { isOpen, barReference } = props.useActionItemsBar()

    const {
        carouselReference,
        canScrollNegative,
        canScrollPositive,
        onNegativeClicked,
        onPositiveClicked,
    } = useCarousel({ direction: 'topToBottom' })

    const haveExtensionsLoaded = useObservable(
        useMemo(() => haveInitialExtensionsLoaded(props.extensionsController.extHostAPI), [props.extensionsController])
    )

    if (!isOpen) {
        return null
    }

    return (
        <div className="action-items__bar p-0 border-left position-relative d-flex flex-column" ref={barReference}>
            <ActionItemsDivider />
            {canScrollNegative && (
                <button
                    type="button"
                    className="btn btn-link action-items__scroll action-items__list-item p-0 border-0"
                    onClick={onNegativeClicked}
                    tabIndex={-1}
                >
                    <MenuUpIcon className="icon-inline" />
                </button>
            )}
            <ActionsContainer
                menu={ContributableMenu.EditorTitle}
                returnInactiveMenuItems={true}
                extensionsController={props.extensionsController}
                empty={null}
                location={props.location}
                platformContext={props.platformContext}
                telemetryService={props.telemetryService}
            >
                {items => (
                    <ul className="action-items__list list-unstyled m-0" ref={carouselReference}>
                        {[
                            ...items,
                            // TODO(tj): Temporary: testing default icons DELETE BEFORE MERGING
                            ...new Array(20).fill(null).map<ActionItemAction>((_value, index) => ({
                                active: false,
                                action: {
                                    category: String(index).slice(-1),
                                    command: 'open',
                                    actionItem: {},
                                    id: `fake-${index}`,
                                },
                            })),
                        ].map((item, index) => (
                            <li key={item.action.id} className="action-items__list-item">
                                <ActionItem
                                    {...props}
                                    {...item}
                                    className={classNames(
                                        actionItemClassName,
                                        !item.action.actionItem?.iconURL &&
                                            `action-items__action--no-icon action-items__icon-${
                                                (index % 5) + 1
                                            } text-sm`
                                    )}
                                    dataContent={
                                        !item.action.actionItem?.iconURL ? item.action.category?.slice(0, 1) : undefined
                                    }
                                    variant="actionItem"
                                    iconClassName="action-items__icon"
                                    pressedClassName="action-items__action--pressed"
                                    inactiveClassName="action-items__action--inactive"
                                    hideLabel={true}
                                    tabIndex={-1}
                                />
                            </li>
                        ))}
                    </ul>
                )}
            </ActionsContainer>
            {canScrollPositive && (
                <button
                    type="button"
                    className="btn btn-link action-items__scroll action-items__list-item p-0 border-0"
                    onClick={onPositiveClicked}
                    tabIndex={-1}
                >
                    <MenuDownIcon className="icon-inline" />
                </button>
            )}
            {haveExtensionsLoaded && <ActionItemsDivider />}
            <ul className="list-unstyled m-0">
                <li className="action-items__list-item">
                    <Link
                        to="/extensions"
                        className={classNames(actionItemClassName, 'action-items__list-item')}
                        data-tooltip="Add extensions"
                    >
                        <PlusIcon className="icon-inline" />
                    </Link>
                </li>
            </ul>
        </div>
    )
})

export const ActionItemsToggle: React.FunctionComponent<ActionItemsToggleProps> = ({
    useActionItemsToggle,
    extensionsController,
    className,
}) => {
    const { isOpen, toggle, toggleReference } = useActionItemsToggle()

    const haveExtensionsLoaded = useObservable(
        useMemo(() => haveInitialExtensionsLoaded(extensionsController.extHostAPI), [extensionsController])
    )

    return (
        <li
            data-tooltip={`${isOpen ? 'Close' : 'Open'} extensions panel`}
            className={classNames(className, 'nav-item border-left')}
        >
            <div
                className={classNames(
                    'action-items__toggle-container',
                    isOpen && 'action-items__toggle-container--open'
                )}
            >
                <ButtonLink
                    className={classNames(actionItemClassName)}
                    onSelect={toggle}
                    buttonLinkRef={toggleReference}
                >
                    {!haveExtensionsLoaded ? (
                        <LoadingSpinner className="icon-inline" />
                    ) : isOpen ? (
                        <ChevronDoubleUpIcon className="icon-inline" />
                    ) : (
                        <PuzzleOutlineIcon className="icon-inline" />
                    )}
                </ButtonLink>
            </div>
        </li>
    )
}

const ActionItemsDivider: React.FunctionComponent<{ className?: string }> = ({ className }) => (
    <li className={classNames(className, 'action-items__divider position-relative rounded-sm d-flex')} />
)
