// The URL to direct users in order to manage their Cody Pro subscription.
export const manageSubscriptionRedirectURL = 'https://accounts.sourcegraph.com/cody/subscription'

/**
 * Note: This is _not_ meant to be used as a React hook, despite the name "use*".
 *
 * useEmbeddedCodyProUi returns if we expect the Cody Pro UI to be served from sourcegraph.com. Meaning
 * we should direct the user to `/cody/manage/subscription` for making changes.
 *
 * If false, we rely on the current behavior. Where users are directed to https://accounts.sourcegraph.com/cody
 * for managing their Cody Pro subscription information.
 */
export function useEmbeddedCodyProUi(): boolean {
    return !!(window.context.frontendCodyProConfig as { stripePublishableKey: string } | undefined)
        ?.stripePublishableKey
}

/**
 * Note that this is a very simplistic approach.
 * "doesThisStringRoughlyResembleAnEmailAddress" would be a more accurate name.
 * And it is definitely not meant to replace the backend validation.
 */
export function isValidEmailAddress(emailAddress: string): boolean {
    return emailRegex.test(emailAddress)
}

/**
 * Regular expression to validate whether a string looks like an email address:
 *  - Contains a single "@" that is not at the beginning or at the end.
 *  - Contains at least one "." after the "@" that is not at the end.
 *
 * NOTE: Keep this in sync with `emailRegex` in the backend
 * (https://sourcegraph.sourcegraph.com/search?q=context:global+r:github.com/sourcegraph/sourcegraph-accounts+f:backend/internal/graph/*+%22var+emailRegex+%3D+regexp.%22&patternType=newStandardRC1&sm=1),
 * and keep in mind that the backend validation has the final say, validation in the web app is only for UX improvement.
 */
const emailRegex = /^[^@]+@[^@]+\.[^@]+$/

/**
 * @param sscUrl The SSC API URL to call. Example: "/checkout/session".
 * @param method E.g. "POST".
 * @param params The body to send to the SSC API. Will be JSON-encoded.
 */
export function fetchThroughSSCProxy(sscUrl: string, method: string, params?: object): Promise<Response> {
    // /.api/ssc/proxy endpoint exchanges the Sourcegraph session credentials for a SAMS access token.
    // And then proxy the request onto the SSC backend, which will actually create the
    // checkout session.
    return fetch(`/.api/ssc/proxy${sscUrl}`, {
        // Pass along the "sgs" session cookie to identify the caller.
        credentials: 'same-origin',
        headers: {
            ...window.context.xhrHeaders,
            'Content-Type': 'application/json',
        },
        method,
        ...(!['GET', 'HEAD'].includes(method) && params ? { body: JSON.stringify(params) } : null),
    })
}
