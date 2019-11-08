import * as constants from '../shared/constants'
import * as path from 'path'
import promClient from 'prom-client'
import { addTags, createTracer, logAndTraceCall, TracingContext } from '../shared/tracing'
import { createCleanFailedJobsProcessor } from './processors/clean-failed-jobs'
import { createCleanOldJobsProcessor } from './processors/clean-old-jobs'
import { createConvertJobProcessor } from './processors/convert'
import { createLogger } from '../shared/logging'
import { createPostgresConnection } from '../shared/database/postgres'
import { createQueue } from '../shared/queue/queue'
import { createUpdateTipsJobProcessor } from './processors/update-tips'
import { ensureDirectory } from '../shared/paths'
import { followsFrom, FORMAT_TEXT_MAP, Span, Tracer } from 'opentracing'
import { instrument } from '../shared/metrics'
import { Job } from 'bull'
import * as metrics from './metrics'
import { Logger } from 'winston'
import * as settings from './settings'
import { startMetricsServer } from './server'
import { waitForConfiguration } from '../shared/config/config'
import { XrepoDatabase } from '../shared/xrepo/xrepo'

/**
 * Wrap a job processor with instrumentation.
 *
 * @param name The job name.
 * @param jobProcessor The job processor.
 * @param logger The logger instance.
 * @param tracer The tracer instance.
 */
const wrapJobProcessor = <T>(
    name: string,
    jobProcessor: (args: T, ctx: TracingContext) => Promise<void>,
    logger: Logger,
    tracer: Tracer | undefined
): ((job: Job) => Promise<void>) => async (job: Job) => {
    logger.debug(`${name} job accepted`, { jobId: job.id })

    // Destructure arguments and injected tracing context
    const { args, tracing } = job.data as { args: T; tracing: object }

    let span: Span | undefined
    if (tracer) {
        // Extract tracing context from job payload
        const publisher = tracer.extract(FORMAT_TEXT_MAP, tracing)
        span = tracer.startSpan(name, publisher ? { references: [followsFrom(publisher)] } : {})
    }

    // Tag tracing context with jobId and arguments
    const ctx = addTags({ logger, span }, { jobId: job.id, ...args })

    await instrument(
        metrics.jobDurationHistogram,
        metrics.jobDurationErrorsCounter,
        (): Promise<void> => logAndTraceCall(ctx, `${name} job`, (ctx: TracingContext) => jobProcessor(args, ctx))
    )
}

/**
 * Runs the worker which accepts LSIF conversion jobs from the work queue.
 *
 * @param logger The logger instance.
 */
async function main(logger: Logger): Promise<void> {
    // Collect process metrics
    promClient.collectDefaultMetrics({ prefix: 'lsif_' })

    // Read configuration from frontend
    const fetchConfiguration = await waitForConfiguration(logger)

    // Configure distributed tracing
    const tracer = createTracer('lsif-worker', fetchConfiguration())

    // Ensure storage roots exist
    await ensureDirectory(settings.STORAGE_ROOT)
    await ensureDirectory(path.join(settings.STORAGE_ROOT, constants.DBS_DIR))
    await ensureDirectory(path.join(settings.STORAGE_ROOT, constants.TEMP_DIR))
    await ensureDirectory(path.join(settings.STORAGE_ROOT, constants.UPLOADS_DIR))

    // Create cross-repo database
    const connection = await createPostgresConnection(fetchConfiguration(), logger)
    const xrepoDatabase = new XrepoDatabase(settings.STORAGE_ROOT, connection)

    // Start metrics server
    startMetricsServer(logger)

    // Create queue to poll for jobs
    const queue = createQueue(settings.REDIS_ENDPOINT, logger)

    const convertJobProcessor = wrapJobProcessor(
        'convert',
        createConvertJobProcessor(xrepoDatabase, fetchConfiguration),
        logger,
        tracer
    )

    const updateTipsJobProcessor = wrapJobProcessor(
        'update-tips',
        createUpdateTipsJobProcessor(xrepoDatabase, fetchConfiguration),
        logger,
        tracer
    )

    const cleanOldJobsProcessor = wrapJobProcessor(
        'clean-old-jobs',
        createCleanOldJobsProcessor(queue, logger),
        logger,
        tracer
    )

    const cleanFailedJobsProcessor = wrapJobProcessor(
        'clean-failed-jobs',
        createCleanFailedJobsProcessor(),
        logger,
        tracer
    )

    // Start processing work
    queue.process('convert', convertJobProcessor).catch(() => {})
    queue.process('update-tips', updateTipsJobProcessor).catch(() => {})
    queue.process('clean-old-jobs', cleanOldJobsProcessor).catch(() => {})
    queue.process('clean-failed-jobs', cleanFailedJobsProcessor).catch(() => {})
}

// Initialize logger
const appLogger = createLogger('lsif-worker')

// Launch!
main(appLogger).catch(error => {
    appLogger.error('failed to start process', { error })
    appLogger.on('finish', () => process.exit(1))
    appLogger.end()
})
