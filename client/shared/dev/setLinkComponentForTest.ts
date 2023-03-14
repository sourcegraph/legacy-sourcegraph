/* eslint-disable @typescript-eslint/no-require-imports, @typescript-eslint/no-var-requires */

// HACK: when running tests within the client/wildcard package the link component
// state is within a local non-packaged version of Link. Set the AnchorLink in the
// local version instead of the packaged version.
// if (process.env.JS_BINARY__PACKAGE?.startsWith('client/wildcard/')) {
try {
    const { setLinkComponent1, AnchorLink1 } = require('../../wildcard/src/components/Link')
    setLinkComponent1(AnchorLink1)
    const { setLinkComponent, AnchorLink } = require('@sourcegraph/wildcard')
    setLinkComponent(AnchorLink)
} catch (error) {
    // ignore error for now
}
