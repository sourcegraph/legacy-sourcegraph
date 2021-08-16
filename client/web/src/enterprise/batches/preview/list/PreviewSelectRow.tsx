import { noop } from 'lodash'
import InfoCircleOutlineIcon from 'mdi-react/InfoCircleOutlineIcon'
import React, { useMemo, useContext } from 'react'

import { pluralize } from '@sourcegraph/shared/src/util/strings'
import { useObservable } from '@sourcegraph/shared/src/util/useObservable'

import { BatchSpecApplyPreviewVariables } from '../../../../graphql-operations'
import { Action, DropdownButton } from '../../DropdownButton'
import { MultiSelectContext } from '../../MultiSelectContext'

import { queryPublishableChangesetSpecIDs as _queryPublishableChangesetSpecIDs } from './backend'

const ACTIONS: Action[] = [
    {
        type: 'unpublish',
        buttonLabel: 'Unpublish',
        dropdownTitle: 'Unpublish',
        dropdownDescription:
            "Remove the selected changesets from this batch change. Unlike archive, this can't be undone.",
        onTrigger: noop,
    },
    {
        type: 'publish',
        buttonLabel: 'Publish',
        dropdownTitle: 'Publish',
        dropdownDescription: 'Re-enqueues the selected changesets for processing, if they failed.',
        onTrigger: noop,
    },
    {
        type: 'publish-draft',
        buttonLabel: 'Publish draft',
        dropdownTitle: 'Publish draft',
        dropdownDescription:
            'Create a comment on all selected changesets. For example, you could ask people for reviews, give an update, or post a cat GIF.',
        onTrigger: noop,
    },
]

export interface PreviewSelectRowProps {
    queryArguments: BatchSpecApplyPreviewVariables
    /** For testing only. */
    dropDownInitiallyOpen?: boolean
    /** For testing only. */
    queryPublishableChangesetSpecIDs?: typeof _queryPublishableChangesetSpecIDs
}

/**
 * Renders the top bar of the PreviewList with the publish status dropdown selector and
 * the X selected label. Provides select ALL functionality.
 */
export const PreviewSelectRow: React.FunctionComponent<PreviewSelectRowProps> = ({
    dropDownInitiallyOpen = false,
    queryPublishableChangesetSpecIDs = _queryPublishableChangesetSpecIDs,
    queryArguments,
}) => {
    const { areAllVisibleSelected, selected, selectAll, totalCount } = useContext(MultiSelectContext)

    const allChangesetSpecIDs: string[] | undefined = useObservable(
        useMemo(() => queryPublishableChangesetSpecIDs(queryArguments), [
            queryArguments,
            queryPublishableChangesetSpecIDs,
        ])
    )

    return (
        <>
            <div className="row align-items-center no-gutters mb-3">
                <div className="ml-2 col d-flex align-items-center">
                    <InfoCircleOutlineIcon className="icon-inline text-muted mr-2" />
                    {selected === 'all' || allChangesetSpecIDs?.length === selected.size ? (
                        <AllSelectedLabel count={allChangesetSpecIDs?.length} />
                    ) : (
                        `${selected.size} ${pluralize('changeset', selected.size)}`
                    )}
                    {selected !== 'all' &&
                        areAllVisibleSelected() &&
                        allChangesetSpecIDs &&
                        allChangesetSpecIDs.length > selected.size && (
                            <button type="button" className="btn btn-link py-0 px-1" onClick={selectAll}>
                                (Select all{totalCount !== undefined && ` ${totalCount}`})
                            </button>
                        )}
                </div>
                <div className="w-100 d-block d-md-none" />
                <div className="m-0 col col-md-auto">
                    <div className="row no-gutters">
                        <div className="col ml-0 ml-sm-2">
                            <DropdownButton
                                actions={ACTIONS}
                                dropdownMenuPosition="right"
                                initiallyOpen={dropDownInitiallyOpen}
                                placeholder="Select action"
                            />
                        </div>
                    </div>
                </div>
            </div>
        </>
    )
}

const AllSelectedLabel: React.FunctionComponent<{ count?: number }> = ({ count }) => {
    if (count === undefined) {
        return <>All changesets selected</>
    }

    return (
        <>
            All {count} {pluralize('changeset', count)} selected
        </>
    )
}
