import { gql } from '@sourcegraph/http-client'

export const ReposSelectorSearchQuery = gql`
    query ReposSelectorSearch($query: String!) {
        repositories(query: $query, first: 10) {
            nodes {
                id
                name
                embeddingExists
                externalRepository {
                    id
                    serviceType
                }
                embeddingJobs(first: 1) {
                    nodes {
                        id
                        state
                        failureMessage
                    }
                }
            }
        }
    }
`

export const ReposStatusQuery = gql`
    query ReposStatus($repoNames: [String!]!, $first: Int!) {
        repositories(names: $repoNames, first: $first) {
            nodes {
                id
                name
                embeddingExists
                externalRepository {
                    id
                    serviceType
                }
                embeddingJobs(first: 1) {
                    nodes {
                        id
                        state
                        failureMessage
                    }
                }
            }
        }
    }
`
