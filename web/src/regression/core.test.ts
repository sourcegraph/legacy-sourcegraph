/**
 * @jest-environment node
 */

import { TestResourceManager } from './util/TestResourceManager'
import { GraphQLClient, createGraphQLClient } from './util/GraphQLClient'
import { Driver } from '../../../shared/src/e2e/driver'
import { getConfig } from '../../../shared/src/e2e/config'
import { getTestTools } from './util/init'
import { ensureLoggedInOrCreateTestUser } from './util/helpers'
import { setUserEmailVerified } from './util/api'
import { ScreenshotVerifier } from './util/ScreenshotVerifier'
import { gql, dataOrThrowErrors } from '../../../shared/src/graphql/graphql'
import { map } from 'rxjs/operators'

describe('Core functionality regression test suite', () => {
    const testUsername = 'test-core'
    const config = getConfig(
        'sudoToken',
        'sudoUsername',
        'gitHubToken',
        'sourcegraphBaseUrl',
        'noCleanup',
        'testUserPassword',
        'logBrowserConsole',
        'slowMo',
        'headless',
        'keepBrowser'
    )

    let driver: Driver
    let gqlClient: GraphQLClient
    let resourceManager: TestResourceManager
    let screenshots: ScreenshotVerifier
    beforeAll(async () => {
        ;({ driver, gqlClient, resourceManager } = await getTestTools(config))
        resourceManager.add(
            'User',
            testUsername,
            await ensureLoggedInOrCreateTestUser(driver, gqlClient, {
                username: testUsername,
                deleteIfExists: true,
                ...config,
            })
        )
        screenshots = new ScreenshotVerifier(driver)
    })

    afterAll(async () => {
        if (!config.noCleanup) {
            await resourceManager.destroyAll()
        }
        if (driver) {
            await driver.close()
        }
        if (screenshots.screenshots.length > 0) {
            console.log(screenshots.verificationInstructions())
        }
    })

    test('User settings are saved and applied', async () => {
        const getSettings = async () => {
            await driver.page.waitForSelector('.view-line')
            return await driver.page.evaluate(() => {
                const editor = document.querySelector('.monaco-editor') as HTMLElement
                return editor ? editor.innerText : null
            })
        }

        await driver.page.goto(config.sourcegraphBaseUrl + `/users/${testUsername}/settings`)
        const previousSettings = await getSettings()
        if (!previousSettings) {
            throw new Error('Previous settings were null')
        }
        const newSettings = '{\xa0/*\xa0These\xa0are\xa0new\xa0settings\xa0*/}'
        await driver.replaceText({
            selector: '.monaco-editor',
            newText: newSettings,
            selectMethod: 'keyboard',
            enterTextMethod: 'paste',
        })
        await driver.page.reload()

        const currentSettings = await getSettings()
        if (currentSettings !== previousSettings) {
            throw new Error(
                `Settings ${JSON.stringify(currentSettings)} did not match (old) saved settings ${JSON.stringify(
                    previousSettings
                )}`
            )
        }

        await driver.replaceText({
            selector: '.monaco-editor',
            newText: newSettings,
            selectMethod: 'keyboard',
            enterTextMethod: 'type',
        })
        await driver.clickElementWithText('Save changes')
        await driver.page.waitForFunction(
            () => document.evaluate("//*[text() = ' Saving...']", document).iterateNext() === null
        )
        await driver.page.reload()

        const currentSettings2 = await getSettings()
        if (JSON.stringify(currentSettings2) !== JSON.stringify(newSettings)) {
            throw new Error(
                `Settings ${JSON.stringify(currentSettings2)} did not match (new) saved settings ${JSON.stringify(
                    newSettings
                )}`
            )
        }

        // Restore old settings
        await driver.replaceText({
            selector: '.monaco-editor',
            newText: previousSettings,
            selectMethod: 'keyboard',
            enterTextMethod: 'paste',
        })
        await driver.clickElementWithText('Save changes')
        await driver.page.waitForFunction(
            () => document.evaluate("//*[text() = ' Saving...']", document).iterateNext() === null
        )
        const previousSettings2 = await getSettings()
        await driver.page.reload()

        const currentSettings3 = await getSettings()
        if (currentSettings3 !== previousSettings2) {
            throw new Error(
                `Settings ${JSON.stringify(currentSettings3)} did not match (old) saved settings ${JSON.stringify(
                    previousSettings2
                )}`
            )
        }
    })

    test('User profile page', async () => {
        const aviURL =
            'https://media2.giphy.com/media/26tPplGWjN0xLybiU/giphy.gif?cid=790b761127d52fa005ed23fdcb09d11a074671ac90146787&rid=giphy.gif'
        const displayName = 'Test Display Name'

        await driver.page.goto(driver.sourcegraphBaseUrl + `/users/${testUsername}/settings/profile`)
        await driver.replaceText({
            selector: '.e2e-user-settings-profile-page__display-name',
            newText: displayName,
        })
        await driver.replaceText({
            selector: '.e2e-user-settings-profile-page__avatar_url',
            newText: aviURL,
            enterTextMethod: 'paste',
        })
        await driver.clickElementWithText('Update profile')
        await driver.page.reload()
        await driver.page.waitForFunction(
            displayName => {
                const el = document.querySelector('.e2e-user-area-header__display-name')
                return el && el.textContent && el.textContent.trim() === displayName
            },
            undefined,
            displayName
        )

        await screenshots.verifySelector(
            'navbar-toggle-is-bart-simpson.png',
            'Navbar toggle avatar is Bart Simpson',
            '.e2e-user-nav-item-toggle'
        )
    })
    test('User emails page', async () => {
        const testEmail = 'sg-test-account@protonmail.com'
        await driver.page.goto(driver.sourcegraphBaseUrl + `/users/${testUsername}/settings/emails`)
        await driver.replaceText({ selector: '.e2e-user-email-add-input', newText: 'sg-test-account@protonmail.com' })
        await driver.clickElementWithText('Add')
        await driver.waitForElementWithText(testEmail)
        await driver.findElementWithText('Verification pending')
        await setUserEmailVerified(gqlClient, testUsername, testEmail, true)
        await driver.page.reload()
        await driver.waitForElementWithText('Verified')
    })

    test('Access tokens work and invalid access tokens return "401 Unauthorized"', async () => {
        await driver.page.goto(config.sourcegraphBaseUrl + `/users/${testUsername}/settings/tokens`)
        await driver.waitForElementWithText('Generate new token', undefined, { timeout: 5000 })
        await driver.clickElementWithText('Generate new token')
        await driver.waitForElementWithText('New access token')
        await driver.replaceText({
            selector: '.e2e-create-access-token-description',
            newText: 'test-regression',
        })
        await driver.waitForElementWithText('Generate token')
        await driver.clickElementWithText('Generate token')
        await driver.waitForElementWithText("Copy the new access token now. You won't be able to see it again.")
        await driver.clickElementWithText('Copy')
        const token = await driver.page.evaluate(() => {
            const tokenEl = document.querySelector('.e2e-access-token')
            if (!tokenEl) {
                return null
            }
            const inputEl = tokenEl.querySelector('input')
            if (!inputEl) {
                return null
            }
            return inputEl.value
        })
        if (!token) {
            throw new Error('Could not obtain access token')
        }
        const gqlClientWithToken = createGraphQLClient({
            baseUrl: config.sourcegraphBaseUrl,
            token,
        })
        await new Promise(resolve => setTimeout(resolve, 2000))
        const currentUsernameQuery = gql`
            query {
                currentUser {
                    username
                }
            }
        `
        const response = await gqlClientWithToken
            .queryGraphQL(currentUsernameQuery)
            .pipe(map(dataOrThrowErrors))
            .toPromise()
        expect(response).toEqual({ currentUser: { username: testUsername } })

        const gqlClientWithInvalidToken = createGraphQLClient({
            baseUrl: config.sourcegraphBaseUrl,
            token: 'aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa',
        })

        let noError = false
        try {
            await gqlClientWithInvalidToken
                .queryGraphQL(currentUsernameQuery)
                .pipe(map(dataOrThrowErrors))
                .toPromise()
            noError = true
        } catch (err) {
            if (!(err as Error).message.includes('401 Unauthorized')) {
                throw new Error(`Unexpected error making GraphQL request with invalid token: ${err}`)
            }
        }
        if (noError) {
            throw new Error('GraphQL request with invalid token completed successfully')
        }
    })

    test('Organizations (admin user)', async () => {
        // TODO(@sourcegraph/web)
    })
    test('Organizations (non-admin user)', async () => {
        // TODO(@sourcegraph/web)
    })
    test('Explore page', async () => {
        // TODO(@sourcegraph/web)
    })
    test('Quicklinks', async () => {
        // TODO(@sourcegraph/web)
    })
})
