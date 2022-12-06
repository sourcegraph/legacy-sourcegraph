import { lazyComponent } from '@sourcegraph/shared/src/util/lazyComponent'
import React from 'react'
import { CodeMirrorQueryInputWrapperProps } from './CodeMirrorQueryInputWrapper'

const CodeMirrorQueryInput = lazyComponent(() => import('./CodeMirrorQueryInputWrapper'), 'CodeMirrorQueryInputWrapper')

export const LazyCodeMirrorQueryInput: React.FunctionComponent<CodeMirrorQueryInputWrapperProps> = props => (
    <CodeMirrorQueryInput {...props} />
)
