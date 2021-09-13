import classNames from 'classnames'
import formatDistanceStrict from 'date-fns/formatDistanceStrict'
import React from 'react'

import { Link } from '@sourcegraph/shared/src/components/Link'

import { DismissibleAlert } from '../components/DismissibleAlert'
import { isProductLicenseExpired, formatRelativeExpirationDate } from '../productSubscription/helpers'

/**
 * A global alert that appears telling the site admin that their license key is about to expire. Even after being dismissed,
 * it reappears every day.
 */
export const LicenseExpirationAlert: React.FunctionComponent<{
    expiresAt: Date
    daysLeft: number
    className?: string
}> = ({ expiresAt, daysLeft, className }) => (
    <DismissibleAlert
        partialStorageKey={`licenseExpiring.${daysLeft}`}
        className={classNames('alert-warning align-items-center', className)}
    >
        Your Sourcegraph license{' '}
        {
            isProductLicenseExpired(expiresAt)
                ? 'expired ' + formatRelativeExpirationDate(expiresAt) // 'Expired two months ago'
                : 'will expire in ' + formatDistanceStrict(expiresAt, Date.now()) // 'Will expire in two months'
        }
        .&nbsp;
        <Link className="site-alert__link" to="/site-admin/license">
            <span className="underline">Renew now</span>
        </Link>
        &nbsp;or&nbsp;
        <a className="site-alert__link" href="https://about.sourcegraph.com/contact">
            <span className="underline">contact Sourcegraph</span>
        </a>
    </DismissibleAlert>
)
