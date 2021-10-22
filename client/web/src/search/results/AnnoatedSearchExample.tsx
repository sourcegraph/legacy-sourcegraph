import classNames from 'classnames'
import React from 'react'
import styles from './AnnotatedSearchExample.module.scss'
import CodeBracketsIcon from 'mdi-react/CodeBracketsIcon'
import FormatLetterCaseIcon from 'mdi-react/FormatLetterCaseIcon'
import SearchIcon from 'mdi-react/SearchIcon'
import RegexIcon from 'mdi-react/RegexIcon'

const arrowHeight = 27
const edgesHeight = 8
const topArrowY = 70
const bottomArrowY = 150

function arrow(x: number, width: number, position: 'top' | 'bottom'): React.ReactElement {
    const pointerLine = <line x1={width / 2} x2={width / 2} y1="0" y2={arrowHeight} />
    const centerLine = <line x1="0" x2={width} y1={arrowHeight} y2={arrowHeight} />
    const leftEdge = <line x1="0" x2="0" y1={arrowHeight} y2={arrowHeight + edgesHeight} />
    const rightEdge = <line x1={width} x2={width} y1={arrowHeight} y2={arrowHeight + edgesHeight} />

    let group = (
        <>
            {pointerLine}
            {centerLine}
            {leftEdge}
            {rightEdge}
        </>
    )
    if (position === 'bottom') {
        group = <g transform={`rotate(180, ${width / 2}, ${(arrowHeight + edgesHeight) / 2})`}>{group}</g>
    }
    return (
        <g transform={`translate(${x}, ${position === 'top' ? topArrowY : bottomArrowY})`} className={styles.arrow}>
            {' '}
            {group}
        </g>
    )
}

export const AnnotatedSearchInput: React.FunctionComponent = React.memo(() => (
    <svg
        className={styles.annotatedSearchInput}
        width="800"
        height="270"
        viewBox="0 0 800 270"
        fill="none"
        xmlns="http://www.w3.org/2000/svg"
    >
        <g>
            <path
                d="M56.5 113C56.5 111.619 57.6193 110.5 59 110.5H688.5V143.5H59C57.6193 143.5 56.5 142.381 56.5 141V113Z"
                className={styles.searchBox}
            />
            <text className={styles.code} x="68" y="130.836">
                <tspan className={styles.filter}>context:</tspan>
                <tspan>global</tspan>
            </text>
            <line className={styles.separator} x1="178" y1="118" x2="178" y2="137" />
            <text className={styles.code} x="171.852" y="130.836">
                <tspan className={styles.filter}> repo:</tspan>
                <tspan>sourcegraph/sourcegraph</tspan>
                <tspan className={styles.metaRegexpCharacterSet}>.</tspan>
                <tspan className={styles.metaRegexpRangeQuantifier}>*</tspan>
                <tspan> function auth(){'{'} </tspan>
            </text>
            <FormatLetterCaseIcon className="icon-inline" x="590" y="115" />
            <RegexIcon className="icon-inline" x="620" y="115" />
            <CodeBracketsIcon className="icon-inline" x="650" y="115" />
            <path
                d="M688 110H731C732.105 110 733 110.895 733 112V142C733 143.105 732.105 144 731 144H688V110Z"
                fill="#1475CF"
            />
            <SearchIcon className={classNames(styles.searchIcon, 'icon-inline')} x="698" y="115" />

            {arrow(418, 120, 'top')}
            <text transform="translate(395, 44)">
                <tspan x="0" y="0">
                    By default, search terms are{' '}
                </tspan>
                <tspan x="0" y="16">
                    interpretted literally (without regexp).
                </tspan>
            </text>

            {arrow(188, 30, 'top')}
            <text transform="translate(116, 44)">
                <tspan x="0" y="0">
                    Filters scope your search to repos,{' '}
                </tspan>
                <tspan x="0" y="16">
                    orgs, languages, and more.
                </tspan>
            </text>

            {arrow(68, 108, 'bottom')}
            <text transform="translate(56, 200)">
                <tspan x="0" y="0">
                    By default, Sourcegraph searches the{' '}
                </tspan>
                <tspan x="0" y="16">
                    <tspan className={styles.bold}>global </tspan>
                    context, which is publicly
                </tspan>
                <tspan x="0" y="32">
                    available code on code hosts such as{' '}
                </tspan>
                <tspan x="0" y="48">
                    GitHub and GitLab.
                </tspan>
            </text>

            {arrow(390, 16, 'bottom')}
            <text transform="translate(340, 200)">
                <tspan x="0" y="0">
                    You can use regexp inside{' '}
                </tspan>
                <tspan x="0" y="16">
                    filters, even when not in{' '}
                </tspan>
                <tspan x="0" y="32">
                    regexp mode
                </tspan>
            </text>

            {arrow(590, 82, 'bottom')}
            <text transform="translate(538, 200)">
                <tspan x="0" y="0">
                    Search can be case-sensitive{' '}
                </tspan>
                <tspan x="0" y="16">
                    one of three modes: literal{' '}
                </tspan>
                <tspan x="0" y="32">
                    (default), regexp or structural.
                </tspan>
            </text>
        </g>
    </svg>
))
