import React, { FunctionComponent, useMemo, useState } from 'react'
import classNames from 'classnames'
import * as H from 'history'

import { AddUserEmailResult, AddUserEmailVariables } from '../../../graphql-operations'
import { gql, dataOrThrowErrors } from '../../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { requestGraphQL } from '../../../backend/graphql'
import { asError, isErrorLike } from '../../../../../shared/src/util/errors'
import { useInputValidation, deriveInputClassName } from '../../../../../shared/src/util/useInputValidation'

import { eventLogger } from '../../../tracking/eventLogger'
import { ErrorAlert } from '../../../components/alerts'
import { LoaderButton } from '../../../components/LoaderButton'
import { LoaderInput } from '../../../../../branded/src/components/LoaderInput'

interface Props {
    user: GQL.ID
    onDidAdd: () => void

    className?: string
    history: H.History
}

interface State {
    loadingOrError?: boolean | Error
}

export const AddUserEmailForm: FunctionComponent<Props> = ({ user, className, onDidAdd, history }) => {
    const [status, setStatus] = useState<State>({})

    const [emailState, nextEmailFieldChange, emailInputReference, overrideEmailState] = useInputValidation(
        useMemo(
            () => ({
                synchronousValidators: [],
                asynchronousValidators: [],
            }),
            []
        )
    )

    const onSubmit: React.FormEventHandler<HTMLFormElement> = async event => {
        event.preventDefault()

        if (emailState.kind === 'VALID') {
            setStatus({ loadingOrError: true })

            try {
                dataOrThrowErrors(
                    await requestGraphQL<AddUserEmailResult, AddUserEmailVariables>(
                        gql`
                            mutation AddUserEmail($user: ID!, $email: String!) {
                                addUserEmail(user: $user, email: $email) {
                                    alwaysNil
                                }
                            }
                        `,
                        { user, email: emailState.value }
                    ).toPromise()
                )
            } catch (error) {
                setStatus({ loadingOrError: asError(error) })
                return
            }

            eventLogger.log('NewUserEmailAddressAdded')
            overrideEmailState()
            setStatus({})
            if (onDidAdd) {
                onDidAdd()
            }
        }
    }

    return (
        <div className={`add-user-email-form ${className || ''}`}>
            <label
                htmlFor="AddUserEmailForm-email"
                className={classNames('align-self-start', {
                    'text-danger font-weight-bold': emailState.kind === 'INVALID',
                })}
            >
                Email address
            </label>
            {/* eslint-disable-next-line react/forbid-elements */}
            <form className="form-inline" onSubmit={onSubmit} noValidate={true}>
                <LoaderInput className={deriveInputClassName(emailState)} loading={emailState.kind === 'LOADING'}>
                    <input
                        id="AddUserEmailForm-email"
                        type="email"
                        name="email"
                        className={classNames(
                            'form-control mr-sm-2 test-user-email-add-input',
                            deriveInputClassName(emailState)
                        )}
                        onChange={nextEmailFieldChange}
                        size={32}
                        value={emailState.value}
                        ref={emailInputReference}
                        required={true}
                        placeholder="Email"
                        autoComplete="email"
                        autoCorrect="off"
                        autoCapitalize="off"
                        spellCheck={false}
                        readOnly={false}
                    />
                </LoaderInput>{' '}
                <LoaderButton
                    loading={typeof status.loadingOrError === 'boolean' ? status.loadingOrError : false}
                    label="Add"
                    type="submit"
                    disabled={typeof status.loadingOrError === 'boolean' ? status.loadingOrError : false}
                    className="btn btn-primary"
                />
                {emailState.kind === 'INVALID' && (
                    <small className="invalid-feedback" role="alert">
                        {emailState.reason}
                    </small>
                )}
            </form>
            {isErrorLike(status.loadingOrError) && (
                <ErrorAlert className="mt-2" error={status.loadingOrError} history={history} />
            )}
        </div>
    )
}
