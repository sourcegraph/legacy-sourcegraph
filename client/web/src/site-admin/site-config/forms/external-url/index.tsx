import { FC } from 'react'

import { Alert, Text, Input, useDebounce, Link } from '@sourcegraph/wildcard'

interface ExternalUrlFormProps {
    className?: string
    url?: string
    onChange: (newUrl: string) => void
}
export const ExternalUrlForm: FC<ExternalUrlFormProps> = ({ className, url = '', onChange }) => {
    const debouncedUrl = useDebounce(url, 500)

    return (
        <div className={className}>
            <Text>
                Customize the URL your organization will use to access this Sourcegraph instance. Configuration is
                required in order for Sourcegraph to work correctly. See <Link to="/help/admin/url">documentation</Link>{' '}
                for more information.
            </Text>
            {!debouncedUrl && <Alert variant="danger">You have not yet configured an external URL.</Alert>}

            <Input
                placeholder="https://sourcegraph.example.com"
                onChange={event => onChange(event.target.value)}
                value={url}
                aria-label="External URL"
            />
        </div>
    )
}
