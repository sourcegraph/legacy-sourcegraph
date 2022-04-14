# Deployment Overview

Sourcegraph supports two main deployment types: [Docker Compose](docker-compose/index.md) and [Kubernetes](kubernetes/index.md). Each deployment type will require a different level of investment and technical understanding. What works best depends on the needs and desired outcomes for your business. 

If you aren't currently working with our Customer Engineering team, this [Deployment Overview](deployment_overview.md) will provide a high-level view of what's available and needed depending on the deployment type you choose. 

Specifically, the table provided in the [Options and scenarios](#options-and-scenarios) section will provide some high-level guidance, followed by more detailed descriptions for each type.

Sourcegraph also provides a [resource estimator](#resource-planning) to help predict and plan the required resource for your deployment. This tool ensures you provision appropriate resources to scale your instance.

If you are short on time and looking for a quick way to test Sourcegraph locally, consider running Sourcegraph via our [Docker single container](docker-single-container/index.md). 

Or, if you don't want to bother with setup and configuration [try Sourcegraph Cloud](https://sourcegraph.com) instead.

## Resource planning

Sourcegraph has provided the [Resource Estimator](resource_estimator.md) as a starting point to determine necessary resources based on the size of your deployment. 

We recommend the [Kubernetes](kubernetes/scale) Deployment type if your deployment scenario includes a large codebase and many users.

## Options and scenarios

| Deployment Type                                          | Suggested for                                           | Setup time      | Resource isolation | Auto-healing | Multi-machine |
| -------------------------------------------------------- | ------------------------------------------------------- | --------------- | :----------------: | :----------: | :-----------: |
| [**Docker Compose**](docker-compose/index.md) | **Small & medium** production deployments               | 🟢 5 minutes     |         ✅          |      ✅       |       ❌       |
| [**Kubernetes**](kubernetes/index.md)         | **Medium & large** highly-available cluster deployments | 🟠 30-90 minutes |         ✅          |      ✅       |       ✅       |
| [**Single-container**](docker-single-container/index.md)       | Local testing                                           | 🟢 1 minute      |         ❌          |      ❌       |       ❌       |

Each of the deployment types listed in the table above provides a different level of capability. As mentioned previously, base your deployment type on the needs of your business. However, you should also consider the technical expertise available for your deployment. The sections below provide more detailed recommendations for each deployment type.

### Docker Compose

We recommend this path for initial production deployments. If your requirements change, you can always [migrate to a different deployment type](deployment_overview.md#migrating-to-a-new-deployment-type) later on.

### Kubernetes

We recommend Kubernetes for large enterprises that depend or have an expectation for highly scalable deployments. It is important to note that if you're looking to deploy via the Kubernetes path, you are **expected to have a team familiar with operating Kubernetes clusters**, including but not limited to the use of persistent storage. If there is any doubt about your team's ability to support this, please speak to your Sourcegraph contact about using Docker Compose instead.

### Single Container 

Finally, if you're just starting out, you can [try Sourcegraph Cloud](https://sourcegraph.com) or [run Sourcegraph locally](docker-single-container/index.md).

> NOTE: The Single container option is provided for local proofs-of-concept and is not intended for testing or deployed at a pre-production/production level. Some features, such as Code Insights, are not available when using this deployment type. If you're just starting out and want the absolute quickest setup time, [try Sourcegraph Cloud](https://sourcegraph.com).

## Reference repositories

For Docker Compose and Kubernetes deployments, Sourcegraph provides reference repositories with branches corresponding to the version of Sourcegraph you wish to deploy. The reference repository contains everything you need to spin up and configure your instance depending on your deployment type, which also assists in your upgrade process going forward. 

For more information, follow the install and configuration docs for your specific deployment type: [Docker Compose](https://github.com/sourcegraph/deploy-sourcegraph-docker/) or [Kubernetes](https://github.com/sourcegraph/deploy-sourcegraph/).

## External services

By default, Sourcegraph provides versions of services it needs to operate, including:

- A [PostgreSQL](https://www.postgresql.org/) instance for storing long-term information, such as user data, when using Sourcegraph's built-in authentication provider instead of an external one.
- A second PostgreSQL instance for storing large-volume precise code intelligence data.
- A [Redis](https://redis.io/) instance for storing short-term information such as user sessions.
- A second Redis instance for storing cache data.
- A [MinIO](https://min.io/) instance that serves as a local S3-compatible object storage to hold user uploads before processing. _This data is for temporary storage, and content will be automatically deleted once processed._
- A [Jaeger](https://www.jaegertracing.io/) instance for end-to-end distributed tracing. 

> NOTE: As a best practice, configure your Sourcegraph instance to use an external or managed version of these services. Using a managed version of PostgreSQL can make backups and recovery easier to manage and perform. Using a managed object storage service may decrease hosting costs as persistent volumes are often more expensive than object storage space.

### External services guides
See the following guides to use an external or managed version of each service type.

- See [Using your PostgreSQL server](../external_services/postgres.md) to replace the bundled PostgreSQL instances.
- See [Using your Redis server](../external_services/redis.md) to replace the bundled Redis instances.
- See [Using a managed object storage service (S3 or GCS)](../external_services/object_storage.md) to replace the bundled MinIO instance.
- See [Using an external Jaeger instance](../observability/tracing.md#use-an-external-jaeger-instance) in our [tracing documentation](../observability/tracing.md) to replace the bundled Jaeger instance.Use-an-external-Jaeger-instance

> NOTE: Using Sourcegraph with an external service is a [paid feature](https://about.sourcegraph.com/pricing). [Contact us](https://about.sourcegraph.com/contact/sales) to get a trial license.

### Cloud alternatives

- Amazon Web Services: [AWS RDS for PostgreSQL](https://aws.amazon.com/rds/), [Amazon ElastiCache](https://aws.amazon.com/elasticache/redis/), and [S3](https://aws.amazon.com/s3/) for storing user uploads.
- Google Cloud: [Cloud SQL for PostgreSQL](https://cloud.google.com/sql/docs/postgres/), [Cloud Memorystore](https://cloud.google.com/memorystore/), and [Cloud Storage](https://cloud.google.com/storage) for storing user uploads.
- Digital Ocean: [Digital Ocean Managed Databases](https://www.digitalocean.com/products/managed-databases/) for [Postgres](https://www.digitalocean.com/products/managed-databases-postgresql/), [Redis](https://www.digitalocean.com/products/managed-databases-redis/), and [Spaces](https://www.digitalocean.com/products/spaces/) for storing user uploads.

## Configuration

Configuration at the deployment level focuses on ensuring your Sourcegraph runs optimally, based on the size of your repositories and the number of users. Configuration options will vary based on the type of deployment you choose, so you will want to consult the specific configuration guides for additional information.

## Customization

We refer to configuration at the Administration level as Customization, check out the [customization section TBD](TBD).


## Upgrades

A new version of Sourcegraph is released every month (with patch releases in between as needed). We actively maintain the two most recent monthly releases of Sourcegraph. The [changelog](../../CHANGELOG.md) provides all information related to any changes that are/were in a release.

Depending on your current version and the version you are looking to upgrade rules, there may be specific upgrade instruction and requirements. Checkout the [Upgrade docs](TBD) for additional information and instructions.

## Migration

Sourcegraph uses "migration" to refer to different, yet related activities. This includes normal database migrations which are an automatic part of the upgrade process and the [migrator service](TBD)

It also refers to migration from one deployment type to the other, for example moving from Docker Single Container to Docker Compose.

More details on both cases are provided in our [Migration docs](TBD)