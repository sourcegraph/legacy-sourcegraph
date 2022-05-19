/**
 * All Sourcegraph Cloud related links
 */
// MAIN
export const VSCE_LINK_DOTCOM = 'https://sourcegraph.com'
export const VSCE_LINK_TOKEN_CALLBACK =
    'https://sourcegraph.com/sign-in?returnTo=user/settings/tokens/new/callback?requestFrom=LOGINVSCE'
export const VSCE_LINK_TOKEN_CALLBACK_TEST =
    'https://sourcegraph.test:3443/sign-in?returnTo=user/settings/tokens/new/callback?requestFrom=LOGINVSCE'

// PARAMS
export const VSCE_SIDEBAR_PARAMS = '?utm_medium=VSCODE&utm_source=sidebar&utm_campaign=vsce-sign-up&utm_content=sign-up'
const VSCE_CALLBACK_CODE = 'LOGINVSCE'
const VSCE_LINK_PARAMS_TOKEN_REDIRECT = {
    returnTo: `user/settings/tokens/new/callback?requestFrom=${VSCE_CALLBACK_CODE}`,
}
const VSCE_LINK_PARAMS_EDITOR = { editor: 'vscode' }
// UTM for Commands
const VSCE_LINK_PARAMS_UTM_COMMANDS = {
    utm_campaign: 'vscode-extension',
    utm_medium: 'direct_traffic',
    utm_source: 'vscode-extension',
    utm_content: 'vsce-commands',
}
// UTM for Sidebar actions
const VSCE_LINK_PARAMS_UTM_SIDEBAR = {
    utm_campaign: 'vsce-sign-up',
    utm_medium: 'VSCODE',
    utm_source: 'sidebar',
    utm_content: 'sign-up',
}
export const VSCE_COMMANDS_PARAMS = new URLSearchParams({ ...VSCE_LINK_PARAMS_UTM_COMMANDS }).toString()

// MISC
export const VSCE_LINK_MARKETPLACE = 'https://marketplace.visualstudio.com/items?itemName=sourcegraph.sourcegraph'
export const VSCE_LINK_USER_DOCS =
    'https://docs.sourcegraph.com/cli/how-tos/creating_an_access_token' + VSCE_SIDEBAR_PARAMS
export const VSCE_LINK_FEEDBACK = 'https://github.com/sourcegraph/sourcegraph/discussions/categories/feedback'
export const VSCE_LINK_ISSUES =
    'https://github.com/sourcegraph/sourcegraph/issues/new?labels=team/integrations,vscode-extension&title=VSCode+Bug+report:+&projects=Integrations%20Project%20Board'
export const VSCE_LINK_TROUBLESHOOT =
    'https://docs.sourcegraph.com/admin/how-to/troubleshoot-sg-extension#vs-code-extension'

// Generate sign-in and sign-up links using the above params
export const VSCE_LINK_SIGNIN = (): string => {
    const uri = new URL(VSCE_LINK_DOTCOM)
    const parameters = new URLSearchParams({
        ...VSCE_LINK_PARAMS_UTM_SIDEBAR,
        ...VSCE_LINK_PARAMS_EDITOR,
        ...VSCE_LINK_PARAMS_TOKEN_REDIRECT,
    }).toString()
    uri.pathname = 'sign-in'
    uri.search = parameters
    return uri.href
}
export const VSCE_LINK_SIGNUP = (): string => {
    const uri = new URL(VSCE_LINK_DOTCOM)
    const parameters = new URLSearchParams({
        ...VSCE_LINK_PARAMS_UTM_SIDEBAR,
        ...VSCE_LINK_PARAMS_EDITOR,
        ...VSCE_LINK_PARAMS_TOKEN_REDIRECT,
    }).toString()
    uri.pathname = 'sign-up'
    uri.search = parameters
    return uri.href
}
