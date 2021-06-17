import cookies, { CookieAttributes } from 'js-cookie'
import * as uuid from 'uuid'

import { TelemetryService } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { browserExtensionMessageReceived, handleQueryEvents, pageViewQueryParameters } from './analyticsUtils'
import { serverAdmin } from './services/serverAdminWrapper'
import { getPreviousMonday, redactSensitiveInfoFromURL } from './util'

export const ANONYMOUS_USER_ID_KEY = 'sourcegraphAnonymousUid'
export const COHORT_ID_KEY = 'sourcegraphCohortId'
export const FIRST_SOURCE_URL_KEY = 'sourcegraphSourceUrl'

export class EventLogger implements TelemetryService {
    private hasStrippedQueryParameters = false

    private anonymousUserID?: string
    private cohortID?: string
    private firstSourceURL?: string

    private readonly CookieSettings: CookieAttributes = {
        // 365 days expiry, but renewed on activity.
        expires: 365,
        // Enforce HTTPS
        secure: true,
        // We only read the cookie with JS so we don't need to send it cross-site nor on initial page requests.
        sameSite: 'Strict',
        // Specify the Domain attribute to ensure subdomains (about.sourcegraph.com) can receive this cookie.
        // https://developer.mozilla.org/en-US/docs/Web/HTTP/Cookies#define_where_cookies_are_sent
        domain: location.hostname,
    }

    constructor() {
        // EventLogger is never teared down
        // eslint-disable-next-line rxjs/no-ignored-subscription
        browserExtensionMessageReceived.subscribe(({ platform }) => {
            this.log('BrowserExtensionConnectedToServer', { platform })

            if (localStorage && localStorage.getItem('eventLogDebug') === 'true') {
                console.debug('%cBrowser extension detected, sync completed', 'color: #aaa')
            }
        })
    }

    /**
     * Log a pageview.
     * Page titles should be specific and human-readable in pascal case, e.g. "SearchResults" or "Blob" or "NewOrg"
     */
    public logViewEvent(pageTitle: string, eventProperties?: any, logAsActiveUser = true): void {
        if (window.context?.userAgentIsBot || !pageTitle) {
            return
        }
        pageTitle = `View${pageTitle}`

        const props = pageViewQueryParameters(window.location.href)
        serverAdmin.trackPageView(pageTitle, logAsActiveUser, eventProperties)
        this.logToConsole(pageTitle, props)

        // Use flag to ensure URL query params are only stripped once
        if (!this.hasStrippedQueryParameters) {
            handleQueryEvents(window.location.href)
            this.hasStrippedQueryParameters = true
        }
    }

    /**
     * Log a user action or event.
     * Event labels should be specific and follow a ${noun}${verb} structure in pascal case, e.g. "ButtonClicked" or "SignInInitiated"
     */
    public log(eventLabel: string, eventProperties?: any): void {
        if (window.context?.userAgentIsBot || !eventLabel) {
            return
        }
        serverAdmin.trackAction(eventLabel, eventProperties)
        this.logToConsole(eventLabel, eventProperties)
    }

    private logToConsole(eventLabel: string, object?: any): void {
        if (localStorage && localStorage.getItem('eventLogDebug') === 'true') {
            console.debug('%cEVENT %s', 'color: #aaa', eventLabel, object)
        }
    }

    /**
     * Get the anonymous identifier for this user (used to allow site admins
     * on a Sourcegraph instance to see a count of unique users on a daily,
     * weekly, and monthly basis).
     * If the user doesn't have an anonymours ID yet, it generates one as well
     * as a cohort ID based on the week the user's first visit.
     */
    public getAnonymousUserID(): string {
        let anonymousUserID =
            this.anonymousUserID || cookies.get(ANONYMOUS_USER_ID_KEY) || localStorage.getItem(ANONYMOUS_USER_ID_KEY)
        if (!anonymousUserID) {
            anonymousUserID = uuid.v4()
            this.generateCohortID()
        }
        // Use cookies instead of localStorage so that the ID can be shared with subdomains (about.sourcegraph.com).
        // Always set to renew expiry and migrate from localStorage
        cookies.set(ANONYMOUS_USER_ID_KEY, anonymousUserID, this.CookieSettings)
        localStorage.removeItem(ANONYMOUS_USER_ID_KEY)
        this.anonymousUserID = anonymousUserID
        return anonymousUserID
    }

    /** Generates a cohort ID for the user, which is the Monday of the week they first visited, in YYYY-MM-DD */
    private generateCohortID(): void {
        const cohortID = getPreviousMonday(new Date())
        cookies.set(COHORT_ID_KEY, cohortID, this.CookieSettings)
        this.cohortID = cohortID
    }

    /**
     * The cohort ID is generated when the anonymous user ID is generated.
     * Users that have visited before the introduction of cohort IDs will not have one.
     */
    public getCohortID(): string | undefined {
        const cohortId = this.cohortID || cookies.get(COHORT_ID_KEY)
        this.cohortID = cohortId
        return cohortId
    }

    public getFirstSourceURL(): string {
        const firstSourceURL = this.firstSourceURL || cookies.get(FIRST_SOURCE_URL_KEY) || location.href

        const redactedURL = redactSensitiveInfoFromURL(firstSourceURL)

        // Use cookies instead of localStorage so that the ID can be shared with subdomains (about.sourcegraph.com).
        // Always set to renew expiry and migrate from localStorage
        cookies.set(FIRST_SOURCE_URL_KEY, redactedURL, {
            // 365 days expiry, but renewed on activity.
            expires: 365,
            // Enforce HTTPS
            secure: true,
            // We only read the cookie with JS so we don't need to send it cross-site nor on initial page requests.
            sameSite: 'Strict',
            // Specify the Domain attribute to ensure subdomains (about.sourcegraph.com) can receive this cookie.
            // https://developer.mozilla.org/en-US/docs/Web/HTTP/Cookies#define_where_cookies_are_sent
            domain: location.hostname,
        })

        this.firstSourceURL = firstSourceURL
        return firstSourceURL
    }
}

export const eventLogger = new EventLogger()
