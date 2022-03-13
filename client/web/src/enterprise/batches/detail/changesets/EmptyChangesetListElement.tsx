import React from 'react'

import { Link } from '@sourcegraph/wildcard'

import styles from './EmptyChangesetListElement.module.scss'

export const EmptyChangesetListElement: React.FunctionComponent<{}> = () => (
    <div className={styles.emptyChangesetListElementBody}>
        <h3 className="mb-2">No changesets exist</h3>
        <div className={styles.emptyChangesetListElementContent}>
            <span>This batch change is a draft. A batch specification must be executed to create changesets.</span>
            <Link to="">View the most recent specification.</Link>
        </div>
    </div>
)
