import { InitData } from '../../../shared/src/api/extension/extensionHost'

/**
 * The information necessary to connect to a Sourcegraph extension.
 */
export interface ExtensionConnectionInfo extends Pick<InitData, 'settingsCascade'> {
    extensionID: string
    jsBundleURL: string
}

/**
 * Executes the callback only on the first message that's received on the port.
 */
export const onFirstMessage = (port: chrome.runtime.Port, callback: (message: any) => void) => {
    const cb = message => {
        port.onMessage.removeListener(cb)
        callback(message)
    }
    port.onMessage.addListener(cb)
}
