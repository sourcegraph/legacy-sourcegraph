import { MockedProvider } from '@apollo/client/testing'
import { cleanup, within, fireEvent, act } from '@testing-library/react'
import React from 'react'

import { cache } from '@sourcegraph/shared/src/graphql/cache'
import { waitForNextApolloResponse } from '@sourcegraph/shared/src/testing/apollo'
import { renderWithRouter, RenderWithRouterResult } from '@sourcegraph/shared/src/testing/render-with-router'

import { RevisionsPopover, RevisionsPopoverProps } from './RevisionsPopover'
import { MOCK_PROPS, MOCK_REQUESTS } from './RevisionsPopover.mocks'

describe('RevisionsPopover', () => {
    let queries: RenderWithRouterResult

    const fetchMoreNodes = async (currentTab: HTMLElement) => {
        fireEvent.click(within(currentTab).getByText('Show more'))
        await waitForNextApolloResponse()
    }

    const renderPopover = (props?: Partial<RevisionsPopoverProps>): RenderWithRouterResult =>
        renderWithRouter(
            <MockedProvider mocks={MOCK_REQUESTS} cache={cache}>
                <RevisionsPopover {...MOCK_PROPS} {...props} />
            </MockedProvider>,
            { route: `/${MOCK_PROPS.repoName}` }
        )

    afterEach(cleanup)

    describe('Branches', () => {
        let branchesTab: HTMLElement

        beforeEach(async () => {
            queries = renderPopover()

            fireEvent.click(queries.getByText('Branches'))
            await waitForNextApolloResponse()

            branchesTab = queries.getByRole('tabpanel', { name: 'Branches' })
        })

        it('renders correct number of results', () => {
            expect(within(branchesTab).getAllByRole('link')).toHaveLength(50)
            expect(within(branchesTab).getByTestId('summary')).toHaveTextContent(
                '100 branches total (showing first 50)'
            )
            expect(within(branchesTab).getByText('Show more')).toBeVisible()
        })

        it('renders result nodes correctly', () => {
            const firstNode = within(branchesTab).getByText('GIT_BRANCH-0-display-name')
            expect(firstNode).toBeVisible()

            const firstLink = firstNode.closest('a')
            expect(firstLink?.getAttribute('href')).toBe(`/${MOCK_PROPS.repoName}@GIT_BRANCH-0-abbrev-name`)
        })

        it('fetches remaining results correctly', async () => {
            await fetchMoreNodes(branchesTab)
            expect(within(branchesTab).getAllByRole('link')).toHaveLength(100)
            expect(within(branchesTab).getByTestId('summary')).toHaveTextContent('100 branches total')
            expect(within(branchesTab).queryByText('Show more')).not.toBeInTheDocument()
        })

        it('searches correctly', async () => {
            const searchInput = within(branchesTab).getByRole('searchbox')
            fireEvent.change(searchInput, { target: { value: 'some query' } })

            // Allow input to debounce
            await act(() => new Promise(resolve => setTimeout(resolve, 200)))
            await waitForNextApolloResponse()

            expect(within(branchesTab).getAllByRole('link')).toHaveLength(2)
            expect(within(branchesTab).getByTestId('summary')).toHaveTextContent('2 branches matching some query')
        })

        it('displays no results correctly', async () => {
            const searchInput = within(branchesTab).getByRole('searchbox')
            fireEvent.change(searchInput, { target: { value: 'some other query' } })

            // Allow input to debounce
            await act(() => new Promise(resolve => setTimeout(resolve, 200)))
            await waitForNextApolloResponse()

            expect(within(branchesTab).queryByRole('link')).not.toBeInTheDocument()
            expect(within(branchesTab).getByTestId('summary')).toHaveTextContent(
                'No branches matching some other query'
            )
        })

        describe('Speculative search', () => {
            beforeEach(async () => {
                cleanup()
                queries = renderPopover({ allowSpeculativeSearch: true })

                fireEvent.click(queries.getByText('Branches'))
                await waitForNextApolloResponse()

                branchesTab = queries.getByRole('tabpanel', { name: 'Branches' })
            })

            it('displays results correctly by displaying a single speculative result', async () => {
                const searchInput = within(branchesTab).getByRole('searchbox')
                fireEvent.change(searchInput, { target: { value: 'some other query' } })

                // Allow input to debounce
                await act(() => new Promise(resolve => setTimeout(resolve, 200)))
                await waitForNextApolloResponse()

                expect(within(branchesTab).getByRole('link')).toBeInTheDocument()

                const firstNode = within(branchesTab).getByText('some other query')
                expect(firstNode).toBeVisible()

                const firstLink = firstNode.closest('a')
                expect(firstLink?.getAttribute('href')).toBe(`/${MOCK_PROPS.repoName}@some%20other%20query`)
            })
        })
    })

    describe('Tags', () => {
        let tagsTab: HTMLElement

        beforeEach(async () => {
            queries = renderPopover()

            fireEvent.click(queries.getByText('Tags'))
            await waitForNextApolloResponse()

            tagsTab = queries.getByRole('tabpanel', { name: 'Tags' })
        })

        it('renders correct number of results', () => {
            expect(within(tagsTab).getAllByRole('link')).toHaveLength(50)
            expect(within(tagsTab).getByTestId('summary')).toHaveTextContent('100 tags total (showing first 50)')
            expect(within(tagsTab).getByText('Show more')).toBeVisible()
        })

        it('renders result nodes correctly', () => {
            const firstNode = within(tagsTab).getByText('GIT_TAG-0-display-name')
            expect(firstNode).toBeVisible()

            const firstLink = firstNode.closest('a')
            expect(firstLink?.getAttribute('href')).toBe(`/${MOCK_PROPS.repoName}@GIT_TAG-0-abbrev-name`)
        })

        it('fetches remaining results correctly', async () => {
            await fetchMoreNodes(tagsTab)
            expect(within(tagsTab).getAllByRole('link')).toHaveLength(100)
            expect(within(tagsTab).getByTestId('summary')).toHaveTextContent('100 tags total')
            expect(within(tagsTab).queryByText('Show more')).not.toBeInTheDocument()
        })

        it('searches correctly', async () => {
            const searchInput = within(tagsTab).getByRole('searchbox')
            fireEvent.change(searchInput, { target: { value: 'some query' } })

            // Allow input to debounce
            await act(() => new Promise(resolve => setTimeout(resolve, 200)))
            await waitForNextApolloResponse()

            expect(within(tagsTab).getAllByRole('link')).toHaveLength(2)
            expect(within(tagsTab).getByTestId('summary')).toHaveTextContent('2 tags matching some query')
        })
    })

    describe('Commits', () => {
        let commitsTab: HTMLElement

        beforeEach(async () => {
            queries = renderPopover()

            fireEvent.click(queries.getByText('Commits'))
            await waitForNextApolloResponse()

            commitsTab = queries.getByRole('tabpanel', { name: 'Commits' })
        })

        it('renders correct number of results', () => {
            expect(within(commitsTab).getAllByRole('link')).toHaveLength(15)
            expect(within(commitsTab).getByText('Show more')).toBeVisible()
        })

        it('renders result nodes correctly', () => {
            const firstNode = within(commitsTab).getByText('git-commit-oid-0')
            expect(firstNode).toBeVisible()
            expect(within(commitsTab).getByText('Commit 0: Hello world')).toBeVisible()
            expect(firstNode.closest('a')?.getAttribute('href')).toBe(`/${MOCK_PROPS.repoName}@git-commit-oid-0`)
        })

        it('fetches remaining results correctly', async () => {
            await fetchMoreNodes(commitsTab)
            expect(within(commitsTab).getAllByRole('link')).toHaveLength(30)
            expect(within(commitsTab).queryByText('Show more')).not.toBeInTheDocument()
        })

        it('searches correctly', async () => {
            const searchInput = within(commitsTab).getByRole('searchbox')
            fireEvent.change(searchInput, { target: { value: 'some query' } })

            // Allow input to debounce
            await act(() => new Promise(resolve => setTimeout(resolve, 200)))
            await waitForNextApolloResponse()

            expect(within(commitsTab).getAllByRole('link')).toHaveLength(2)
            expect(within(commitsTab).getByTestId('summary')).toHaveTextContent('2 commits matching some query')
        })
    })
})
