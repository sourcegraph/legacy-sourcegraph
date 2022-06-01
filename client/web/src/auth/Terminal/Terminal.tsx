import React from 'react'

import { Code } from '@sourcegraph/wildcard'

import terminalStyles from './Terminal.module.scss'

// 73 '=' characters are the 100% of the progress bar
const CHARACTERS_LENGTH = 73

export const Terminal: React.FunctionComponent<React.PropsWithChildren<unknown>> = ({ children }) => (
    <section className={terminalStyles.terminalWrapper}>
        <ul className={terminalStyles.downloadProgressWrapper}>{children}</ul>
    </section>
)

export const TerminalTitle: React.FunctionComponent<React.PropsWithChildren<unknown>> = ({ children }) => (
    <header className={terminalStyles.terminalTitle}>
        <Code>{children}</Code>
    </header>
)

export const TerminalLine: React.FunctionComponent<React.PropsWithChildren<unknown>> = ({ children }) => (
    <li className={terminalStyles.terminalLine}>{children}</li>
)

export const TerminalDetails: React.FunctionComponent<React.PropsWithChildren<unknown>> = ({ children }) => (
    <div>
        <Code>{children}</Code>
    </div>
)

export const TerminalProgress: React.FunctionComponent<
    React.PropsWithChildren<{ progress: number; character: string }>
> = ({ progress = 0, character = '#' }) => {
    const numberOfChars = Math.ceil((progress / 100) * CHARACTERS_LENGTH)

    return <Code className={terminalStyles.downloadProgress}>{character.repeat(numberOfChars)}</Code>
}
