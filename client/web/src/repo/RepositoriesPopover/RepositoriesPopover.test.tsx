import { cleanup, fireEvent, act } from '@testing-library/react'
import React from 'react'

import { MockedTestProvider, waitForNextApolloResponse } from '@sourcegraph/shared/src/testing/apollo'
import { renderWithRouter, RenderWithRouterResult } from '@sourcegraph/shared/src/testing/render-with-router'

import { RepositoriesPopover } from './RepositoriesPopover'
import { MOCK_REQUESTS } from './RepositoriesPopover.mocks'

const repo = {
    id: 'some-repo-id',
    name: 'github.com/sourcegraph/sourcegraph',
}

describe('RevisionsPopover', () => {
    let renderResult: RenderWithRouterResult

    const fetchMoreNodes = async () => {
        fireEvent.click(renderResult.getByText('Show more'))
        await waitForNextApolloResponse()
    }

    const waitForInputDebounce = () => act(() => new Promise(resolve => setTimeout(resolve, 200)))

    beforeEach(async () => {
        renderResult = renderWithRouter(
            <MockedTestProvider mocks={MOCK_REQUESTS}>
                <RepositoriesPopover currentRepo={repo.id} />
            </MockedTestProvider>,
            { route: repo.name }
        )

        await waitForNextApolloResponse()
    })

    afterEach(cleanup)

    it('renders correct number of results', () => {
        expect(renderResult.getAllByRole('link')).toHaveLength(10)
        expect(renderResult.getByText('Show more')).toBeVisible()
    })

    it('renders result nodes correctly', () => {
        const firstNode = renderResult.getByText('some-org/repository-name-0')
        expect(firstNode).toBeVisible()

        const firstLink = firstNode.closest('a')
        expect(firstLink?.getAttribute('href')).toBe('/github.com/some-org/repository-name-0')
    })

    it('fetches remaining results correctly', async () => {
        await fetchMoreNodes()
        expect(renderResult.getAllByRole('link')).toHaveLength(20)
        expect(renderResult.queryByText('Show more')).not.toBeInTheDocument()
    })

    it('searches correctly', async () => {
        const searchInput = renderResult.getByRole('searchbox')
        fireEvent.change(searchInput, { target: { value: 'some query' } })

        await waitForInputDebounce()
        await waitForNextApolloResponse()

        expect(renderResult.getAllByRole('link')).toHaveLength(2)
    })

    it('displays no results correctly', async () => {
        const searchInput = renderResult.getByRole('searchbox')
        fireEvent.change(searchInput, { target: { value: 'some other query' } })

        // Allow input to debounce
        await act(() => new Promise(resolve => setTimeout(resolve, 200)))
        await waitForNextApolloResponse()

        expect(renderResult.queryByRole('link')).not.toBeInTheDocument()
        expect(renderResult.getByTestId('summary')).toHaveTextContent('No repositories matching some other query')
    })
})
