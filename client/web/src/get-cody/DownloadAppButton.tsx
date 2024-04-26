import classNames from 'classnames'

import { EVENT_LOGGER } from '@sourcegraph/shared/src/telemetry/web/eventLogger'
import { Badge, Icon, Link } from '@sourcegraph/wildcard'

import styles from './DownloadAppButton.module.scss'

interface DownloadAppButtonProps {
    to: string
    icon: string
    buttonText: string
    badgeText: string
    eventName: string
    eventType: string
}

export const DownloadAppButton: React.FunctionComponent<DownloadAppButtonProps> = ({
    to,
    icon,
    buttonText,
    badgeText,
    eventName,
    eventType,
}) => {
    const handleOnClick = (): void => {
        EVENT_LOGGER.log(eventName, { type: eventType })
    }

    return (
        <Link to={to} className={classNames('text-decoration-none', styles.downloadButton)} onClick={handleOnClick}>
            <Icon className={classNames('mr-2', styles.icon)} svgPath={icon} inline={false} aria-hidden={true} />
            {buttonText} <Badge className={classNames('ml-2', styles.badge)}>{badgeText}</Badge>
        </Link>
    )
}
