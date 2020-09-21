import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import HelpCircleOutlineIcon from 'mdi-react/HelpCircleOutlineIcon'
import React, { useCallback, useMemo, useState } from 'react'
import { asError } from '../../../shared/src/util/errors'
import { Form } from '../components/Form'
import { eventLogger } from '../tracking/eventLogger'
import { enterpriseTrial, signupTerms } from '../util/features'
import { EmailInput, PasswordInput, UsernameInput } from './SignInSignUpCommon'
import { ErrorAlert } from '../components/alerts'
import classNames from 'classnames'
import * as H from 'history'
import { OrDivider } from './OrDivider'
import GithubIcon from 'mdi-react/GithubIcon'
import { size } from 'lodash'
import { USERNAME_MAX_LENGTH, VALID_USERNAME_REGEXP } from '../user'
import { concat, Observable, of } from 'rxjs'
import { useEventObservable } from '../../../shared/src/util/useObservable'
import { debounceTime, map, share, switchMap, tap } from 'rxjs/operators'
import { typingDebounceTime } from '../search/input/QueryInput'
import CheckIcon from 'mdi-react/CheckIcon'
export interface SignUpArgs {
    email: string
    username: string
    password: string
    requestedTrial: boolean
}

interface SignUpFormProps {
    className?: string

    /** Called to perform the signup on the server. */
    doSignUp: (args: SignUpArgs) => Promise<void>

    buttonLabel?: string
    history: H.History
}
/**
 * TODO: Better naming
 */
interface SignUpFormValidator {
    /**
     * Optional array of synchronous input validators.
     *
     * If there's no problem with the input, void return. Else,
     * return with the reason the input is invalid.
     */
    synchronousValidators?: ((value: string) => string | void)[]

    /**
     * Optional array of asynchronous input validators.
     *
     * If there's no problem with the input, void return. Else,
     * return with the reason the input is invalid.
     */
    asynchronousValidators?: ((value: string) => Promise<string | void>)[]
}

/** Lazily construct this in SignUpForm */
const signUpFormValidators: { [name in 'email' | 'username' | 'password']: SignUpFormValidator } = {
    email: {
        synchronousValidators: [checkEmailFormat, checkEmailPattern],
        asynchronousValidators: [isEmailUnique],
    },
    username: {
        synchronousValidators: [checkUsernameLength, checkUsernamePattern],
        asynchronousValidators: [isUsernameUnique],
    },
    password: {
        synchronousValidators: [checkPasswordLength],
    },
}

type ValidationResult = { kind: 'VALID' } | { kind: 'INVALID'; reason: string }
type ValidationPipeline = (events: Observable<React.ChangeEvent<HTMLInputElement>>) => Observable<ValidationResult>

/**
 * TODO: RxJS integration w/ React component. Create pipeline for
 * useEventObservable? Wrap in useMemo
 *
 * To be consumed by `useEventObservable`
 *
 * @param name
 * @param formValidator
 */
function createValidationPipeline(
    name: string,
    setInputState: (inputState: { value: string; loading: boolean }) => void,
    formValidator: SignUpFormValidator
): ValidationPipeline {
    const { synchronousValidators = [], asynchronousValidators = [] } = formValidator

    /**
     * Validation Pipeline takes an observable<string> and returns an observable
     * of Validation Result
     */
    return function validationPipeline(events): Observable<ValidationResult> {
        const inputValues = events.pipe(
            map(event => event.target.value),
            tap(value => setInputState({ value, loading: true })),
            share()
        )

        // merge synchronous and asynchronous validation. check sync before debounce?

        return inputValues.pipe(
            debounceTime(typingDebounceTime),
            switchMap(value => {
                console.log('vall', value)
                // test synchronous validators first. if reason, return of(reason)
                // looping over validators here because we only need the first reason it's invalid

                for (const validator of synchronousValidators) {
                    const reason = validator(value)
                    if (reason) {
                        return of({ kind: 'INVALID' as const, reason }).pipe(
                            tap(() => setInputState({ value, loading: false }))
                        )
                    }
                }

                // const hi = Promise.all(asynchronousValidators.map(validator => validator(value)))
                // .then(reasons => {
                //     // just need the first reason
                //     for (const reason of reasons) {
                //         if (reason) {
                //             // TODO
                //             break
                //         }
                //     }
                // })
                // .catch(() => {
                //     // noop
                //     // Good behavior. 'Unknown error'
                //     return {value, }
                // })

                // else, kick off async validation. if none of THESE are invalid either, return valid true
                return concat(of({ kind: 'VALID' as const })).pipe(tap(() => setInputState({ value, loading: false })))
            })
        )
    }
}

/**
 *
 */
export const SignUpForm: React.FunctionComponent<SignUpFormProps> = ({ doSignUp, history, buttonLabel, className }) => {
    const [loading, setLoading] = useState(false)
    const [requestedTrial, setRequestedTrial] = useState(false)
    const [error, setError] = useState<Error | null>(null)

    const [emailState, setEmailState] = useState({ value: '', loading: false })
    const [nextEmailFieldChange, emailValidationResult] = useEventObservable<
        React.ChangeEvent<HTMLInputElement>,
        ValidationResult
    >(useMemo(() => createValidationPipeline('email', setEmailState, signUpFormValidators.email), []))

    const [usernameState, setUsernameState] = useState({ value: '', loading: false })
    const [nextUsernameFieldChange, usernameValidationResult] = useEventObservable<
        React.ChangeEvent<HTMLInputElement>,
        ValidationResult
    >(useMemo(() => createValidationPipeline('username', setUsernameState, signUpFormValidators.username), []))

    const [passwordState, setPasswordState] = useState({ value: '', loading: false })
    const [nextPasswordFieldChange, passwordValidationResult] = useEventObservable<
        React.ChangeEvent<HTMLInputElement>,
        ValidationResult
    >(useMemo(() => createValidationPipeline('password', setPasswordState, signUpFormValidators.password), []))

    const canRegister =
        emailValidationResult?.kind === 'VALID' &&
        usernameValidationResult?.kind === 'VALID' &&
        passwordValidationResult?.kind === 'VALID'

    const disabled = loading || !canRegister
    // const disabled = loading

    const handleSubmit = useCallback(
        (event: React.FormEvent<HTMLFormElement>): void => {
            event.preventDefault()
            if (disabled) {
                return
            }

            setLoading(true)
            doSignUp({
                email: emailState.value,
                username: usernameState.value,
                password: passwordState.value,
                requestedTrial,
            }).catch(error => {
                setError(asError(error))
                setLoading(false)
            })
            eventLogger.log('InitiateSignUp')
        },
        [doSignUp, disabled, emailState, usernameState, passwordState, requestedTrial]
    )

    const onRequestTrialFieldChange = useCallback((event: React.ChangeEvent<HTMLInputElement>): void => {
        setRequestedTrial(event.target.checked)
    }, [])

    return (
        <Form
            className={classNames(
                'signin-signup-form',
                'signup-form',
                'test-signup-form',
                'border rounded p-4',
                'text-left',
                className
            )}
            onSubmit={handleSubmit}
            noValidate={true}
        >
            {error && <ErrorAlert className="mb-3" error={error} history={history} />}
            <div className="form-group d-flex flex-column align-content-start">
                <label className="align-self-start">Email</label>
                <div className="signin-signup-form__input-container">
                    <EmailInput
                        className="signin-signup-form__input"
                        onChange={nextEmailFieldChange}
                        required={true}
                        value={emailState.value}
                        disabled={loading}
                        autoFocus={true}
                        placeholder=" "
                    />
                    {emailState.loading ? (
                        <LoadingSpinner className="signin-signup-form__icon" />
                    ) : (
                        emailValidationResult?.kind === 'VALID' && (
                            <CheckIcon className="signin-signup-form__icon" size={20} />
                        )
                    )}
                </div>
                {!emailState.loading && emailValidationResult?.kind === 'INVALID' && (
                    <p>Email bad bc {emailValidationResult.reason}</p>
                )}
            </div>
            <div className="form-group d-flex flex-column align-content-start">
                <label className="align-self-start">Username</label>
                <div className="signin-signup-form__input-container">
                    <UsernameInput
                        className="signin-signup-form__input"
                        onChange={nextUsernameFieldChange}
                        value={usernameState.value}
                        required={true}
                        disabled={loading}
                        placeholder=" "
                    />
                    {usernameState.loading ? (
                        <LoadingSpinner className="signin-signup-form__icon" />
                    ) : (
                        usernameValidationResult?.kind === 'VALID' && (
                            <CheckIcon className="signin-signup-form__icon" size={20} />
                        )
                    )}
                </div>
                {!usernameState.loading && usernameValidationResult?.kind === 'INVALID' && (
                    <p>Email bad bc {usernameValidationResult.reason}</p>
                )}
            </div>
            <div className="form-group d-flex flex-column align-content-start">
                <label className="align-self-start">Password</label>
                <div className="signin-signup-form__input-container">
                    <PasswordInput
                        className="signin-signup-form__input"
                        onChange={nextPasswordFieldChange}
                        value={passwordState.value}
                        required={true}
                        disabled={loading}
                        autoComplete="new-password"
                        placeholder=" "
                    />
                    {passwordState.loading ? (
                        <LoadingSpinner className="signin-signup-form__icon" />
                    ) : (
                        passwordValidationResult?.kind === 'VALID' && (
                            <CheckIcon className="signin-signup-form__icon" size={20} />
                        )
                    )}
                </div>
                {!passwordState.loading && passwordValidationResult?.kind === 'INVALID' ? (
                    <span>Email bad bc {passwordValidationResult.reason}</span>
                ) : (
                    <span>At least 12 characters</span>
                )}
            </div>
            {enterpriseTrial && (
                <div className="form-group">
                    <div className="form-check">
                        <label className="form-check-label">
                            <input className="form-check-input" type="checkbox" onChange={onRequestTrialFieldChange} />
                            Try Sourcegraph Enterprise free for 30 days{' '}
                            {/* eslint-disable-next-line react/jsx-no-target-blank */}
                            <a target="_blank" rel="noopener" href="https://about.sourcegraph.com/pricing">
                                <HelpCircleOutlineIcon className="icon-inline" />
                            </a>
                        </label>
                    </div>
                </div>
            )}
            <div className="form-group mb-0">
                <button className="btn btn-primary btn-block" type="submit" disabled={disabled}>
                    {loading ? <LoadingSpinner className="icon-inline" /> : buttonLabel || 'Sign up'}
                </button>
            </div>
            {window.context.sourcegraphDotComMode && (
                <>
                    {size(window.context.authProviders) > 0 && <OrDivider className="my-4" />}
                    {window.context.authProviders?.map((provider, index) => (
                        // Use index as key because display name may not be unique. This is OK
                        // here because this list will not be updated during this component's lifetime.
                        /* eslint-disable react/no-array-index-key */
                        <div className="mb-2" key={index}>
                            <a href={provider.authenticationURL} className="btn btn-secondary btn-block">
                                {provider.displayName === 'GitHub' && <GithubIcon className="icon-inline" />} Continue
                                with {provider.displayName}
                            </a>
                        </div>
                    ))}
                </>
            )}

            {signupTerms && (
                <p className="mt-3 mb-0">
                    <small className="form-text text-muted">
                        By signing up, you agree to our {/* eslint-disable-next-line react/jsx-no-target-blank */}
                        <a href="https://about.sourcegraph.com/terms" target="_blank" rel="noopener">
                            Terms of Service
                        </a>{' '}
                        and {/* eslint-disable-next-line react/jsx-no-target-blank */}
                        <a href="https://about.sourcegraph.com/privacy" target="_blank" rel="noopener">
                            Privacy Policy
                        </a>
                        .
                    </small>
                </p>
            )}
        </Form>
    )
}

// Synchronous Validators

function checkPasswordLength(password: string): string | void {
    if (password.length < 12) {
        return 'Password should be'
    }
}

function checkUsernameLength(username: string): string | void {
    if (username.length > USERNAME_MAX_LENGTH) {
        return 'Username is'
    }
}

function checkUsernamePattern(username: string): string | void {
    const valid = new RegExp(VALID_USERNAME_REGEXP).test(username)
    if (!valid) {
        return "Username doesn't match the requested format"
    }
}

/**
 * Simple email format validation to catch the most glaring errors
 * and display helpful error messages
 */
function checkEmailFormat(email: string): string | void {
    const parts = email.trim().split('@')
    if (parts.length < 2) {
        return "Please include an '@' in the email address"
    }
    if (parts.length > 2) {
        return "A part following '@' should not contain the symbol '@'"
    }
}

/**
 * Catch-all regex for errors not caught by `checkEmailFormat`.
 * From emailregex.com
 */
function checkEmailPattern(email: string): string | void {
    if (
        // eslint-disable-next-line no-useless-escape
        !/^(([^\s"(),.:;<>@[\\\]]+(\.[^\s"(),.:;<>@[\\\]]+)*)|(".+"))@((\[(?:\d{1,3}\.){3}\d{1,3}])|(([\dA-Za-z\-]+\.)+[A-Za-z]{2,}))$/.test(
            email
        )
    ) {
        return 'Please enter a valid email'
    }
}

// Asynchronous Validators

async function isEmailUnique(email: string): Promise<string | void> {
    try {
        const response = await fetch(`/-/is-email-taken/${email}`)
        switch (response.status) {
            case 200:
                return `The email '${email}' is taken.`
            case 404:
                // Email is unique
                return

            default:
                return 'Unknown error'
        }
    } catch {
        return 'Unknown error'
    }
}

async function isUsernameUnique(username: string): Promise<string | void> {
    try {
        const response = await fetch(`/-/is-username-taken/${username}`)
        switch (response.status) {
            case 200:
                return `The email '${username}' is taken.`
            case 404:
                // Username is unique
                return

            default:
                return 'Unknown error'
        }
    } catch {
        return 'Unknown error'
    }
}
