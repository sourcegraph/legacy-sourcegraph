/** @jest-environment setup-polly-jest/jest-environment-node */

import { createDriverForTest, Driver } from '@sourcegraph/shared/src/testing/driver'
import { setupPollyServer } from '@sourcegraph/shared/src/testing/integration/context'
import { settingsID, testUserID } from '@sourcegraph/shared/src/testing/integration/graphQlResults'
import { afterEachSaveScreenshotIfFailedWithJest } from '@sourcegraph/shared/src/testing/screenshotReporter'
import { retry } from '@sourcegraph/shared/src/testing/utils'

import { createWebIntegrationTestContext, WebIntegrationTestContext } from './context'
import { commonWebGraphQlResults } from './graphQlResults'
import { percySnapshotWithVariants } from './utils'

describe('Settings', () => {
    let driver: Driver
    const pollyServer = setupPollyServer(__dirname)

    beforeAll(async () => {
        driver = await createDriverForTest()
    })
    afterAll(() => driver?.close())
    let testContext: WebIntegrationTestContext
    beforeEach(async () => {
        testContext = await createWebIntegrationTestContext({
            driver,
            directory: __dirname,
            pollyServer: pollyServer.polly,
        })
    })
    afterEachSaveScreenshotIfFailedWithJest(() => driver.page)
    afterEach(() => testContext?.dispose())

    describe('User settings page', () => {
        it('updates user settings', async () => {
            testContext.overrideGraphQL({
                ...commonWebGraphQlResults,
                SettingsCascade: () => ({
                    settingsSubject: {
                        settingsCascade: {
                            subjects: [
                                {
                                    latestSettings: {
                                        id: settingsID,
                                        contents: JSON.stringify({}),
                                    },
                                },
                            ],
                        },
                    },
                }),
                OverwriteSettings: () => ({
                    settingsMutation: {
                        overwriteSettings: {
                            empty: {
                                alwaysNil: null,
                            },
                        },
                    },
                }),
                UserAreaUserProfile: () => ({
                    user: {
                        __typename: 'User',
                        id: testUserID,
                        username: 'test',
                        displayName: null,
                        url: '/users/test',
                        settingsURL: '/users/test/settings',
                        avatarURL: null,
                        viewerCanAdminister: true,
                        builtinAuth: true,
                        tags: [],
                    },
                }),
                UserSettingsAreaUserProfile: () => ({
                    node: {
                        __typename: 'User',
                        id: testUserID,
                        username: 'test',
                        displayName: null,
                        url: '/users/test',
                        settingsURL: '/users/test/settings',
                        avatarURL: null,
                        viewerCanAdminister: true,
                        viewerCanChangeUsername: true,
                        siteAdmin: true,
                        builtinAuth: true,
                        createdAt: '2020-03-02T11:52:15Z',
                        emails: [{ email: 'test@sourcegraph.test', verified: true }],
                        organizations: { nodes: [] },
                        permissionsInfo: null,
                        tags: [],
                    },
                }),
            })

            const getSettingsEditorContent = async (): Promise<string | null | undefined> => {
                await driver.page.waitForSelector('.test-settings-file .monaco-editor .view-lines')
                return driver.page.evaluate(
                    () =>
                        document
                            .querySelector<HTMLElement>('.test-settings-file .monaco-editor .view-lines')
                            ?.textContent?.replace(/\u00A0/g, ' ') // Monaco replaces all spaces with &nbsp;
                )
            }

            await driver.page.goto(driver.sourcegraphBaseUrl + '/users/test/settings')

            await driver.page.waitForSelector('.test-settings-file .monaco-editor')
            await driver.page.waitForSelector('.test-save-toolbar-save')

            expect(
                await driver.page.evaluate(
                    () => document.querySelector<HTMLButtonElement>('.test-save-toolbar-save')?.disabled
                )
            ).toBe(true)

            await percySnapshotWithVariants(driver.page, 'Settings page')

            // Replace with new settings
            const newSettings = '{ /* These are new settings */}'
            await driver.replaceText({
                selector: '.test-settings-file .monaco-editor .view-lines',
                newText: newSettings,
                selectMethod: 'keyboard',
                enterTextMethod: 'type',
            })
            await retry(async () => {
                const currentSettings = await getSettingsEditorContent()
                expect(currentSettings).toBe(newSettings)
            })

            expect(
                await driver.page.evaluate(
                    () => document.querySelector<HTMLButtonElement>('.test-save-toolbar-save')?.disabled
                )
            ).toBe(false)

            // Assert mutation is done when save button is clicked
            const overrideSettingsVariables = await testContext.waitForGraphQLRequest(async () => {
                await driver.findElementWithText('Save changes', { action: 'click' })
            }, 'OverwriteSettings')

            expect(overrideSettingsVariables).toEqual({
                contents: newSettings,
                lastID: settingsID,
                subject: testUserID,
            })
        })
    })
})
