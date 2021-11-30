import classNames from 'classnames'
import React from 'react'

import styles from './Badge.module.scss'
import { BADGE_SIZES, BADGE_VARIANTS } from './constants'

export interface BadgeProps {
    /**
     * The variant style of the badge. Defaults to `primary`
     */
    variant?: typeof BADGE_VARIANTS[number]
    /**
     * Allows modifying the size of the badge. Supports larger or smaller variants.
     */
    size?: typeof BADGE_SIZES[number]
    /**
     * Render the badge as a rounded pill
     */
    pill?: boolean
    /**
     * Additional text to display on hover
     */
    tooltip?: string
    /**
     * Used to render the badge as a link to a specific URL
     */
    href?: string
    className?: string
}

export const Badge: React.FunctionComponent<BadgeProps> = ({
    children,
    variant = 'primary',
    size,
    pill,
    tooltip,
    className,
    href,
}) => {
    const commonProps = {
        'data-tooltip': tooltip,
        className: classNames(
            'badge',
            `badge-${variant}`,
            size && `badge-${size}`,
            pill && 'badge-pill',
            styles.badge,
            className
        ),
    }

    if (href) {
        return (
            <a href={href} rel="noopener" target="_blank" {...commonProps}>
                {children}
            </a>
        )
    }

    return <span {...commonProps}>{children}</span>
}
