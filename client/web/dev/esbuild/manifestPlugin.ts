import fs from 'fs'
import path from 'path'

import type * as esbuild from 'esbuild'

import { STATIC_ASSETS_PATH } from '@sourcegraph/build-config'

import { type WebpackManifest, WEBPACK_MANIFEST_PATH } from '../utils'

export const assetPathPrefix = '/.assets'

export const getManifest = (jsEntrypoint?: string, cssEntrypoint?: string): WebpackManifest => ({
    'main.js': path.join(assetPathPrefix, jsEntrypoint ?? 'scripts/main.js'),
    'main.css': path.join(assetPathPrefix, cssEntrypoint ?? 'scripts/main.css'),
    isModule: true,
})

const writeManifest = async (manifest: WebpackManifest): Promise<void> => {
    await fs.promises.writeFile(WEBPACK_MANIFEST_PATH, JSON.stringify(manifest, null, 2))
}

/**
 * An esbuild plugin to write a webpack.manifest.json file (just as Webpack does), for compatibility
 * with our current Webpack build.
 */
export const manifestPlugin: esbuild.Plugin = {
    name: 'manifest',
    setup: build => {
        build.initialOptions.metafile = true

        build.onEnd(async result => {
            console.log(process.cwd())
            const { entryPoints } = build.initialOptions
            const outputs = result?.metafile?.outputs

            if (!entryPoints) {
                console.error('[manifestPlugin] No entrypoints found')
                return
            }
            if (!Array.isArray(entryPoints) || typeof entryPoints[0] !== 'string' || entryPoints.length === 0) {
                console.error('[manifestPlugin] Unexpected entryPoints format')
                return
            }
            const entryPoint = entryPoints[0]
            const relativeEntrypoint = path.relative(process.cwd(), entryPoint)

            if (!outputs) {
                return
            }
            let jsEntrypoint: string | undefined
            let cssEntrypoint: string | undefined

            // Find the entrypoint in the output files
            for (const [asset, output] of Object.entries(outputs)) {
                if (output.entryPoint === relativeEntrypoint) {
                    jsEntrypoint = path.relative(STATIC_ASSETS_PATH, asset)
                    if (output.cssBundle) {
                        cssEntrypoint = path.relative(STATIC_ASSETS_PATH, output.cssBundle)
                    }
                }
            }

            await writeManifest(getManifest(jsEntrypoint, cssEntrypoint))
        })
    },
}
