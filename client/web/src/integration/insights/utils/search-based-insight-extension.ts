
import { View } from 'sourcegraph';

import { ErrorLike } from '@sourcegraph/shared/src/util/errors';

/**
 * Generates simplify version of search insight extension for testing purpose.
 * */
export function generateSearchInsightExtensionBundle(data?: Record<string, View | undefined | ErrorLike>): string {
    const injectedDataString = JSON.stringify(data ?? {})

    return `
        var sourcegraph = require('sourcegraph')
        var insightViewStore = JSON.parse('${injectedDataString}')

        function activate(context) {

            function handleInsights(config) {
                const insights = Object.entries(config).filter(([key]) => key.startsWith('searchInsights.insight.'))

                for (var insight of insights) {
                    const [id, settings] = insight;

                    var provideView = () => insightViewStore[id]

                    context.subscriptions.add(sourcegraph.app.registerViewProvider(id + '.insightsPage', {
                        where: 'insightsPage',
                        provideView,
                    }))
                }
            }

            sourcegraph.configuration.subscribe(() => {
                var config = sourcegraph.configuration.get().value

                handleInsights(config)
            })
        }

        exports.activate = activate
    `
}
