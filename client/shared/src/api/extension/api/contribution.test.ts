import { describe, expect, it, test } from 'vitest'

import { ContributableMenu, type Contributions, type Evaluated } from '@sourcegraph/client-api'
import { type Context, parse, parseTemplate } from '@sourcegraph/template-parser'

import {
    evaluateContributions,
    filterContributions,
    mergeContributions,
    parseContributionExpressions,
} from './contribution'

describe('mergeContributions()', () => {
    const FIXTURE_CONTRIBUTIONS_1: Evaluated<Contributions> = {
        actions: [
            { id: '1.a', command: 'c', title: '1.A', telemetryProps: { feature: '1.a' } },
            { id: '1.b', command: 'c', title: '1.B', telemetryProps: { feature: '1.b' } },
        ],
        menus: {
            [ContributableMenu.GlobalNav]: [{ action: '1.a' }, { action: '1.b' }],
        },
    }

    const FIXTURE_CONTRIBUTIONS_2: Evaluated<Contributions> = {
        actions: [
            { id: '2.a', command: 'c', title: '2.A', telemetryProps: { feature: '2.a' } },
            { id: '2.b', command: 'c', title: '2.B', telemetryProps: { feature: '2.b' } },
        ],
        menus: {
            [ContributableMenu.EditorTitle]: [{ action: '2.a' }, { action: '2.b' }],
        },
    }
    const FIXTURE_CONTRIBUTIONS_MERGED: Evaluated<Contributions> = {
        actions: [
            { id: '1.a', command: 'c', title: '1.A', telemetryProps: { feature: '1.a' } },
            { id: '1.b', command: 'c', title: '1.B', telemetryProps: { feature: '1.b' } },
            { id: '2.a', command: 'c', title: '2.A', telemetryProps: { feature: '2.a' } },
            { id: '2.b', command: 'c', title: '2.B', telemetryProps: { feature: '2.b' } },
        ],
        menus: {
            [ContributableMenu.GlobalNav]: [{ action: '1.a' }, { action: '1.b' }],
            [ContributableMenu.EditorTitle]: [{ action: '2.a' }, { action: '2.b' }],
        },
    }

    test('handles an empty array', () => {
        expect(mergeContributions([])).toEqual({})
    })
    test('handles a single item', () => {
        expect(mergeContributions([FIXTURE_CONTRIBUTIONS_1])).toEqual(FIXTURE_CONTRIBUTIONS_1)
    })
    test('handles multiple items', () => {
        expect(
            mergeContributions([FIXTURE_CONTRIBUTIONS_1, FIXTURE_CONTRIBUTIONS_2, {}, { actions: [] }, { menus: {} }])
        ).toEqual(FIXTURE_CONTRIBUTIONS_MERGED)
    })
})

const FIXTURE_CONTEXT: Context = {
    a: true,
    b: false,
    replaceMe: 'x',
}

describe('filterContributions()', () => {
    it('handles empty contributions', () => {
        expect(filterContributions({})).toEqual({})
    })

    it('handles empty menu contributions', () => {
        const expected: Evaluated<Contributions> = {
            menus: {},
        }
        expect(filterContributions({ menus: {} })).toEqual(expected)
    })

    it('handles non-empty contributions', () => {
        const expected: Evaluated<Contributions> = {
            actions: [
                { id: 'a1', command: 'c', telemetryProps: { feature: 'a1' } },
                { id: 'a2', command: 'c', telemetryProps: { feature: 'a2' } },
                { id: 'a3', command: 'c', telemetryProps: { feature: 'a3' } },
            ],
            menus: {
                [ContributableMenu.GlobalNav]: [{ action: 'a1', when: true }, { action: 'a2' }],
            },
        }
        expect(
            filterContributions({
                actions: [
                    { id: 'a1', command: 'c', telemetryProps: { feature: 'a1' } },
                    { id: 'a2', command: 'c', telemetryProps: { feature: 'a2' } },
                    { id: 'a3', command: 'c', telemetryProps: { feature: 'a3' } },
                ],
                menus: {
                    [ContributableMenu.GlobalNav]: [
                        { action: 'a1', when: true },
                        { action: 'a2' },
                        { action: 'a3', when: false },
                    ],
                },
            })
        ).toEqual(expected)
    })
})

/* eslint-disable no-template-curly-in-string */
describe('evaluateContributions()', () => {
    test('handles empty contributions', () => {
        const expected: Evaluated<Contributions> = {}
        expect(evaluateContributions({}, {})).toEqual(expected)
    })

    test('handles empty array of command contributions', () => {
        const expected: Evaluated<Contributions> = {
            actions: [],
        }
        expect(evaluateContributions({}, { actions: [] })).toEqual(expected)
    })

    test('handles non-empty contributions', () => {
        const input: Contributions = {
            actions: [
                {
                    id: 'a1',
                    command: 'c1',
                    commandArguments: [
                        parseTemplate('${replaceMe}'),
                        parseTemplate('b'),
                        parseTemplate('${replaceMe}'),
                    ],
                    title: parseTemplate('${replaceMe}'),
                    category: parseTemplate('${replaceMe}'),
                    description: parseTemplate('${replaceMe}'),
                    iconURL: parseTemplate('${replaceMe}'),
                    actionItem: {
                        label: parseTemplate('${replaceMe}'),
                        description: parseTemplate('${replaceMe}'),
                        iconDescription: parseTemplate('${replaceMe}'),
                        iconURL: parseTemplate('${replaceMe}'),
                    },
                    telemetryProps: { feature: 'a1' },
                },
                {
                    id: 'a2',
                    command: 'c2',
                    title: parseTemplate('${replaceMe}'),
                    category: parseTemplate('b'),
                    telemetryProps: { feature: 'a2' },
                },
                {
                    id: 'a3',
                    command: 'c3',
                    title: parseTemplate('b'),
                    category: parseTemplate('b'),
                    actionItem: {
                        label: parseTemplate('${replaceMe}'),
                        description: parseTemplate('b'),
                    },
                    telemetryProps: { feature: 'a3' },
                },
            ],
        }
        const origInput = JSON.parse(JSON.stringify(input))
        const expected: Evaluated<Contributions> = {
            actions: [
                {
                    id: 'a1',
                    command: 'c1',
                    commandArguments: ['x', 'b', 'x'],
                    title: 'x',
                    category: 'x',
                    description: 'x',
                    iconURL: 'x',
                    actionItem: {
                        label: 'x',
                        description: 'x',
                        iconDescription: 'x',
                        iconURL: 'x',
                    },
                    telemetryProps: { feature: 'a1' },
                },
                { id: 'a2', command: 'c2', title: 'x', category: 'b', telemetryProps: { feature: 'a2' } },
                {
                    id: 'a3',
                    command: 'c3',
                    title: 'b',
                    category: 'b',
                    actionItem: { label: 'x', description: 'b' },
                    telemetryProps: { feature: 'a3' },
                },
            ],
        }
        expect(evaluateContributions(FIXTURE_CONTEXT, input)).toEqual(expected)
        expect(input).toEqual(origInput)
    })

    test('supports commandArguments with the first element non-evaluated', () => {
        const expected: Evaluated<Contributions> = {
            actions: [
                {
                    id: 'x',
                    command: 'c',
                    commandArguments: ['b', 'x', 'b', 'x'],
                    telemetryProps: { feature: 'x' },
                },
            ],
        }
        expect(
            evaluateContributions(FIXTURE_CONTEXT, {
                actions: [
                    {
                        id: 'x',
                        command: 'c',
                        commandArguments: [
                            parseTemplate('b'),
                            parseTemplate('${replaceMe}'),
                            parseTemplate('b'),
                            parseTemplate('${replaceMe}'),
                        ],
                        telemetryProps: { feature: 'x' },
                    },
                ],
            })
        ).toEqual(expected)
    })

    test('evaluates `actionItem.pressed` if present', () => {
        const input: Contributions = {
            actions: [
                {
                    id: 'a',
                    command: 'c',
                    title: parseTemplate('a'),
                    actionItem: {
                        pressed: parse('a'),
                    },
                    telemetryProps: { feature: 'a' },
                },
            ],
        }
        expect(evaluateContributions(FIXTURE_CONTEXT, input)).toEqual({
            actions: [
                { id: 'a', command: 'c', title: 'a', actionItem: { pressed: true }, telemetryProps: { feature: 'a' } },
            ],
        })
    })
})

describe('parseContributionExpressions()', () => {
    it('should not parse the `id` or `command` values', () => {
        const expected: Contributions = {
            actions: [{ id: '${replaceMe}', command: '${c}', telemetryProps: { feature: 'a' } }],
        }
        expect(
            parseContributionExpressions({
                actions: [{ id: '${replaceMe}', command: '${c}', telemetryProps: { feature: 'a' } }],
            })
        ).toEqual(expected)
    })
})
/* eslint-enable no-template-curly-in-string */
