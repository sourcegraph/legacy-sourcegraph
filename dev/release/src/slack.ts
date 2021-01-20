import got from 'got'
import { readLine, cacheFolder } from './util'

export async function postMessage(message: string, channel: string): Promise<void> {
    const webhookURL = await readLine(
        `Enter the Slack webhook URL corresponding to the #${channel} channel (https://api.slack.com/apps/APULW2LKS/incoming-webhooks?): `,
        `${cacheFolder}/slackWebhookURL-${channel}.txt`
    )
    await got.post(webhookURL, {
        body: JSON.stringify({ text: message, link_names: true }),
    })
}

export function slackURL(text: string, url: string): string {
    return `<${url}|${text}>`
}
