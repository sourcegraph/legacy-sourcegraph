import { InitialParametersSource } from '@sourcegraph/search'
import { SearchPatternType } from '@sourcegraph/shared/src/graphql-operations'

import { setQueryStateFromSettings, useNavbarQueryState } from './navbarSearchQueryState'

describe('navbar query state', () => {
    describe('set state from settings', () => {
        it('sets default search pattern', () => {
            setQueryStateFromSettings({
                subjects: [],
                final: {
                    'search.defaultPatternType': SearchPatternType.regexp,
                },
            })

            expect(useNavbarQueryState.getState()).toHaveProperty('searchPatternType', SearchPatternType.regexp)
        })

        it('sets default case sensitivity', () => {
            setQueryStateFromSettings({
                subjects: [],
                final: {
                    'search.defaultCaseSensitive': true,
                },
            })

            expect(useNavbarQueryState.getState()).toHaveProperty('searchCaseSensitivity', true)
        })
    })

    describe('state initialization precedence', () => {
        it('does not prefer user settings over settings from URL', () => {
            useNavbarQueryState.setState({
                parametersSource: InitialParametersSource.URL,
                searchPatternType: SearchPatternType.regexp,
            })

            setQueryStateFromSettings({
                subjects: [],
                final: {
                    'search.defaultPatternType': SearchPatternType.structural,
                },
            })

            expect(useNavbarQueryState.getState()).toHaveProperty('searchPatternType', SearchPatternType.regexp)
        })
    })
})
