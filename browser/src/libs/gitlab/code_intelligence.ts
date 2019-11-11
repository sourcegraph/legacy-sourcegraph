import { Omit } from 'utility-types'
import { CodeHost } from '../code_intelligence'
import { CodeView } from '../code_intelligence/code_views'
import { getSelectionsFromHash, observeSelectionsFromHash } from '../code_intelligence/util/selections'
import { ViewResolver } from '../code_intelligence/views'
import { diffDOMFunctions, singleFileDOMFunctions } from './dom_functions'
import { getCommandPaletteMount } from './extensions'
import { resolveCommitFileInfo, resolveDiffFileInfo, resolveFileInfo } from './file_info'
import { getPageInfo, GitLabPageKind } from './scrape'

const toolbarButtonProps = {
    className: 'btn btn-default btn-sm',
}

export function checkIsGitlab(): boolean {
    return !!document.head.querySelector('meta[content="GitLab"]')
}

const adjustOverlayPosition: CodeHost['adjustOverlayPosition'] = ({ top, left }) => {
    const header = document.querySelector('header')
    if (header) {
        top += header.getBoundingClientRect().height
    }
    // When running GitLab from source, we also need to take into account
    // the debug header shown at the top of the page.
    const debugHeader = document.querySelector('#js-peek.development')
    if (debugHeader) {
        top += debugHeader.getBoundingClientRect().height
    }
    return {
        top,
        left,
    }
}

export const getToolbarMount = (codeView: HTMLElement): HTMLElement => {
    const existingMount: HTMLElement | null = codeView.querySelector('.sg-toolbar-mount-gitlab')
    if (existingMount) {
        return existingMount
    }

    const fileActions = codeView.querySelector('.file-actions')
    if (!fileActions) {
        throw new Error('Unable to find mount location')
    }

    const mount = document.createElement('div')
    mount.classList.add('btn-group')
    mount.classList.add('sg-toolbar-mount')
    mount.classList.add('sg-toolbar-mount-gitlab')

    fileActions.insertAdjacentElement('afterbegin', mount)

    return mount
}

const singleFileCodeView: Omit<CodeView, 'element'> = {
    dom: singleFileDOMFunctions,
    getToolbarMount,
    resolveFileInfo,
    toolbarButtonProps,
    getSelections: getSelectionsFromHash,
    observeSelections: observeSelectionsFromHash,
}

const mergeRequestCodeView: Omit<CodeView, 'element'> = {
    dom: diffDOMFunctions,
    getToolbarMount,
    resolveFileInfo: resolveDiffFileInfo,
    toolbarButtonProps,
}

const commitCodeView: Omit<CodeView, 'element'> = {
    dom: diffDOMFunctions,
    getToolbarMount,
    resolveFileInfo: resolveCommitFileInfo,
    toolbarButtonProps,
}

const resolveView: ViewResolver<CodeView>['resolveView'] = (element: HTMLElement): CodeView | null => {
    if (element.classList.contains('discussion-wrapper')) {
        // This is a commented snippet in a merge request discussion timeline
        // (a snippet where somebody added a review comment on a piece of code in the MR),
        // we don't support adding code intelligence on those.
        return null
    }
    const { pageKind } = getPageInfo()

    if (pageKind === GitLabPageKind.Other) {
        return null
    }

    if (pageKind === GitLabPageKind.File) {
        return { element, ...singleFileCodeView }
    }

    if (pageKind === GitLabPageKind.MergeRequest) {
        if (!element.querySelector('.file-actions')) {
            // If the code view has no file actions, we cannot resolve its head commit ID.
            // This can be the case for code views representing added git submodules.
            return null
        }
        return { element, ...mergeRequestCodeView }
    }

    return { element, ...commitCodeView }
}

const codeViewResolver: ViewResolver<CodeView> = {
    selector: '.file-holder',
    resolveView,
}

export const gitlabCodeHost: CodeHost = {
    type: 'gitlab',
    name: 'GitLab',
    check: checkIsGitlab,
    codeViewResolvers: [codeViewResolver],
    adjustOverlayPosition,
    getCommandPaletteMount,
    getContext: () => ({
        ...getPageInfo(),
        privateRepository: window.location.hostname !== 'gitlab.com',
    }),
    commandPaletteClassProps: {
        popoverClassName: 'dropdown-menu command-list-popover--gitlab',
        formClassName: 'dropdown-input',
        inputClassName: 'dropdown-input-field',
        resultsContainerClassName: 'dropdown-content',
        selectedActionItemClassName: 'is-focused',
        noResultsClassName: 'px-3',
    },
    codeViewToolbarClassProps: {
        className: 'code-view-toolbar--gitlab',
        actionItemClass: 'btn btn-sm btn-secondary action-item--gitlab',
        actionItemPressedClass: 'active',
    },
    hoverOverlayClassProps: {
        className: 'card',
        actionItemClassName: 'btn btn-secondary action-item--gitlab',
        actionItemPressedClassName: 'active',
        closeButtonClassName: 'btn',
        infoAlertClassName: 'alert alert-info',
        errorAlertClassName: 'alert alert-danger',
    },
    codeViewsRequireTokenization: true,
}
