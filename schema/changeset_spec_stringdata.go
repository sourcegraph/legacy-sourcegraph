// Code generated by stringdata. DO NOT EDIT.

package schema

// ChangesetSpecSchemaJSON is the content of the file "changeset_spec.schema.json".
const ChangesetSpecSchemaJSON = `{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "title": "ChangesetSpec",
  "description": "A changeset specification, which describes a changeset to be created or an existing changeset to be tracked.",
  "type": "object",
  "oneOf": [
    {
      "properties": {
        "baseRepository": {
          "type": "string",
          "description": "The GraphQL ID of the repository that contains the existing changeset on the code host.",
          "examples": ["UmVwb3NpdG9yeTo5Cg=="]
        },
        "externalID": {
          "type": "string",
          "description": "The ID that uniquely identifies the existing changeset on the code host",
          "examples": ["3912", "12"]
        }
      },
      "required": ["baseRepository", "externalID"],
      "additionalProperties": false
    },
    {
      "properties": {
        "baseRepository": {
          "type": "string",
          "description": "The GraphQL ID of the repository that this changeset spec is proposing to change.",
          "examples": ["UmVwb3NpdG9yeTo5Cg=="]
        },
        "baseRef": {
          "type": "string",
          "description": "The full name of the Git ref in the base repository that this changeset is based on (and is proposing to be merged into). This ref must exist on the base repository.",
          "examples": ["refs/heads/master"]
        },
        "baseRev": {
          "type": "string",
          "description": "The base revision this changeset is based on. It is the latest commit in baseRef at the time when the changeset spec was created.",
          "examples": ["4095572721c6234cd72013fd49dff4fb48f0f8a4"]
        },
        "headRepository": {
          "type": "string",
          "description": "The GraphQL ID of the repository that contains the branch with this changeset's changes. Fork repositories and cross-repository changesets are not yet supported. Therefore, headRepository must be equal to baseRepository.",
          "examples": ["UmVwb3NpdG9yeTo5Cg=="]
        },
        "headRef": {
          "type": "string",
          "description": "The full name of the Git ref that holds the changes proposed by this changeset. This ref will be created or updated with the commits.",
          "examples": ["refs/heads/fix-foo"]
        },
        "title": { "type": "string", "description": "The title of the changeset on the code host." },
        "body": { "type": "string", "description": "The body (description) of the changeset on the code host." },
        "commits": {
          "type": "array",
          "description": "The Git commits with the proposed changes. These commits are pushed to the head ref.",
          "minItems": 1,
          "maxItems": 1,
          "items": {
            "title": "GitCommitDescription",
            "type": "object",
            "description": "The Git commit to create with the changes.",
            "additionalProperties": false,
            "required": ["message", "diff"],
            "properties": {
              "message": {
                "type": "string",
                "description": "The Git commit message."
              },
              "diff": {
                "type": "string",
                "description": "The commit diff (in unified diff format)."
              }
            }
          }
        },
        "published": {
          "type": "boolean",
          "description": "Whether to publish the changeset. An unpublished changeset can be previewed on Sourcegraph by any person who can view the campaign, but its commit, branch, and pull request aren't created on the code host. A published changeset results in a commit, branch, and pull request being created on the code host."
        }
      },
      "required": [
        "baseRepository",
        "baseRef",
        "baseRev",
        "headRepository",
        "headRef",
        "title",
        "body",
        "commits",
        "published"
      ],
      "additionalProperties": false
    }
  ]
}
`
