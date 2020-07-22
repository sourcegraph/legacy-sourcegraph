// Code generated by stringdata. DO NOT EDIT.

package schema

// GitLabSchemaJSON is the content of the file "gitlab.schema.json".
const GitLabSchemaJSON = `{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "$id": "gitlab.schema.json#",
  "title": "GitLabConnection",
  "description": "Configuration for a connection to GitLab (GitLab.com or GitLab self-managed).",
  "allowComments": true,
  "type": "object",
  "additionalProperties": false,
  "required": ["url", "token", "projectQuery"],
  "properties": {
    "url": {
      "description": "URL of a GitLab instance, such as https://gitlab.example.com or (for GitLab.com) https://gitlab.com.",
      "type": "string",
      "pattern": "^https?://",
      "not": {
        "type": "string",
        "pattern": "example\\.com"
      },
      "format": "uri",
      "examples": ["https://gitlab.com", "https://gitlab.example.com"]
    },
    "token": {
      "description": "A GitLab access token with \"api\" scope. If you are enabling permissions with identity provider type \"external\", this token should also have \"sudo\" scope.",
      "type": "string",
      "minLength": 1
    },
    "rateLimit": {
      "description": "Rate limit applied when making background API requests to GitLab.",
      "title": "GitLabRateLimit",
      "type": "object",
      "required": ["enabled", "requestsPerHour"],
      "properties": {
        "enabled": {
          "description": "true if rate limiting is enabled.",
          "type": "boolean",
          "default": true
        },
        "requestsPerHour": {
          "description": "Requests per hour permitted. This is an average, calculated per second.",
          "type": "number",
          "default": 36000,
          "minimum": 0
        }
      },
      "default": {
        "enabled": true,
        "requestsPerHour": 36000
      }
    },
    "gitURLType": {
      "description": "The type of Git URLs to use for cloning and fetching Git repositories on this GitLab instance.\n\nIf \"http\", Sourcegraph will access GitLab repositories using Git URLs of the form http(s)://gitlab.example.com/myteam/myproject.git (using https: if the GitLab instance uses HTTPS).\n\nIf \"ssh\", Sourcegraph will access GitLab repositories using Git URLs of the form git@example.gitlab.com:myteam/myproject.git. See the documentation for how to provide SSH private keys and known_hosts: https://docs.sourcegraph.com/admin/repo/auth#repositories-that-need-http-s-or-ssh-authentication.",
      "type": "string",
      "enum": ["http", "ssh"],
      "default": "http"
    },
    "certificate": {
      "description": "TLS certificate of the GitLab instance. This is only necessary if the certificate is self-signed or signed by an internal CA. To get the certificate run ` + "`" + `openssl s_client -connect HOST:443 -showcerts < /dev/null 2> /dev/null | openssl x509 -outform PEM` + "`" + `. To escape the value into a JSON string, you may want to use a tool like https://json-escape-text.now.sh.",
      "type": "string",
      "pattern": "^-----BEGIN CERTIFICATE-----\n",
      "examples": ["-----BEGIN CERTIFICATE-----\n..."]
    },
    "projects": {
      "description": "A list of projects to mirror from this GitLab instance. Supports including by name ({\"name\": \"group/name\"}) or by ID ({\"id\": 42}).",
      "type": "array",
      "minItems": 1,
      "items": {
        "type": "object",
        "title": "GitLabProject",
        "additionalProperties": false,
        "anyOf": [{ "required": ["name"] }, { "required": ["id"] }],
        "properties": {
          "name": {
            "description": "The name of a GitLab project (\"group/name\") to mirror.",
            "type": "string",
            "pattern": "^[\\w.-]+(/[\\w.-]+)+$"
          },
          "id": {
            "description": "The ID of a GitLab project (as returned by the GitLab instance's API) to mirror.",
            "type": "integer"
          }
        }
      },
      "examples": [
        [{ "name": "group/name" }, { "id": 42 }],
        [{ "name": "gnachman/iterm2" }, { "name": "gitlab-org/gitlab-ce" }]
      ]
    },
    "exclude": {
      "description": "A list of projects to never mirror from this GitLab instance. Takes precedence over \"projects\" and \"projectQuery\" configuration. Supports excluding by name ({\"name\": \"group/name\"}) or by ID ({\"id\": 42}).",
      "type": "array",
      "items": {
        "type": "object",
        "title": "ExcludedGitLabProject",
        "additionalProperties": false,
        "anyOf": [{ "required": ["name"] }, { "required": ["id"] }],
        "properties": {
          "name": {
            "description": "The name of a GitLab project (\"group/name\") to exclude from mirroring.",
            "type": "string",
            "pattern": "^[\\w.-]+(/[\\w.-]+)+$"
          },
          "id": {
            "description": "The ID of a GitLab project (as returned by the GitLab instance's API) to exclude from mirroring.",
            "type": "integer"
          }
        }
      },
      "examples": [
        [{ "name": "group/name" }, { "id": 42 }],
        [{ "name": "gitlab-org/gitlab-ee" }, { "name": "gitlab-com/www-gitlab-com" }]
      ]
    },
    "projectQuery": {
      "description": "An array of strings specifying which GitLab projects to mirror on Sourcegraph. Each string is a URL path and query that targets a GitLab API endpoint returning a list of projects. If the string only contains a query, then \"projects\" is used as the path. Examples: \"?membership=true&search=foo\", \"groups/mygroup/projects\".\n\nThe special string \"none\" can be used as the only element to disable this feature. Projects matched by multiple query strings are only imported once. Here are a few endpoints that return a list of projects: https://docs.gitlab.com/ee/api/projects.html#list-all-projects, https://docs.gitlab.com/ee/api/groups.html#list-a-groups-projects, https://docs.gitlab.com/ee/api/search.html#scope-projects.",
      "type": "array",
      "default": ["none"],
      "items": {
        "type": "string",
        "minLength": 1
      },
      "minItems": 1,
      "examples": [["?membership=true&search=foo", "groups/mygroup/projects"]]
    },
    "repositoryPathPattern": {
      "description": "The pattern used to generate a the corresponding Sourcegraph repository name for a GitLab project. In the pattern, the variable \"{host}\" is replaced with the GitLab URL's host (such as gitlab.example.com), and \"{pathWithNamespace}\" is replaced with the GitLab project's \"namespace/path\" (such as \"myteam/myproject\").\n\nFor example, if your GitLab is https://gitlab.example.com and your Sourcegraph is https://src.example.com, then a repositoryPathPattern of \"{host}/{pathWithNamespace}\" would mean that a GitLab project at https://gitlab.example.com/myteam/myproject is available on Sourcegraph at https://src.example.com/gitlab.example.com/myteam/myproject.\n\nIt is important that the Sourcegraph repository name generated with this pattern be unique to this code host. If different code hosts generate repository names that collide, Sourcegraph's behavior is undefined.",
      "type": "string",
      "default": "{host}/{pathWithNamespace}"
    },
    "nameTransformations": {
      "description": "An array of transformations will apply to the repository name. Currently, only regex replacement is supported. All transformations happen after \"repositoryPathPattern\" is processed.",
      "type": "array",
      "items": {
        "$ref": "#/definitions/NameTransformation"
      },
      "examples": [
        [
          {
            "regex": "\\.d/",
            "replacement": "/"
          },
          {
            "regex": "-git$",
            "replacement": ""
          }
        ]
      ]
    },
    "initialRepositoryEnablement": {
      "description": "Defines whether repositories from this GitLab instance should be enabled and cloned when they are first seen by Sourcegraph. If false, the site admin must explicitly enable GitLab repositories (in the site admin area) to clone them and make them searchable on Sourcegraph. If true, they will be enabled and cloned immediately (subject to rate limiting by GitLab); site admins can still disable them explicitly, and they'll remain disabled.",
      "type": "boolean"
    },
    "authorization": {
      "title": "GitLabAuthorization",
      "description": "If non-null, enforces GitLab repository permissions. This requires that there be an item in the ` + "`" + `auth.providers` + "`" + ` field of type \"gitlab\" with the same ` + "`" + `url` + "`" + ` field as specified in this ` + "`" + `GitLabConnection` + "`" + `.",
      "type": "object",
      "additionalProperties": false,
      "required": ["identityProvider"],
      "properties": {
        "identityProvider": {
          "description": "The source of identity to use when computing permissions. This defines how to compute the GitLab identity to use for a given Sourcegraph user.",
          "type": "object",
          "required": ["type"],
          "properties": {
            "type": {
              "type": "string",
              "enum": ["oauth", "username", "external"]
            }
          },
          "oneOf": [
            { "$ref": "#/definitions/OAuthIdentity" },
            { "$ref": "#/definitions/UsernameIdentity" },
            { "$ref": "#/definitions/ExternalIdentity" }
          ],
          "!go": {
            "taggedUnionType": true
          }
        },
        "ttl": {
          "description": "DEPRECATED: The TTL of how long to cache permissions data. This is 3 hours by default.\n\nDecreasing the TTL will increase the load on the code host API. If you have X private repositories on your instance, it will take ~X/100 API requests to fetch the complete list for 1 user.  If you have Y users, you will incur up to X*Y/100 API requests per cache refresh period (depending on user activity).\n\nIf set to zero, Sourcegraph will sync a user's entire accessible repository list on every request (NOT recommended).\n\nPublic and internal repositories are cached once for all users per cache TTL period.",
          "type": "string",
          "default": "3h"
        }
      }
    }
  },
  "definitions": {
    "OAuthIdentity": {
      "type": "object",
      "additionalProperties": false,
      "required": ["type"],
      "properties": {
        "type": {
          "type": "string",
          "const": "oauth"
        },
        "minBatchingThreshold": {
          "description": "DEPRECATED: The minimum number of GitLab projects to fetch at which to start batching requests to fetch project visibility. Please consult with the Sourcegraph support team before modifying this.",
          "type": "integer",
          "default": 200
        },
        "maxBatchRequests": {
          "description": "DEPRECATED: The maximum number of batch API requests to make for GitLab Project visibility. Please consult with the Sourcegraph support team before modifying this.",
          "type": "integer",
          "default": 300
        }
      }
    },
    "UsernameIdentity": {
      "type": "object",
      "additionalProperties": false,
      "required": ["type"],
      "properties": {
        "type": {
          "type": "string",
          "const": "username"
        }
      }
    },
    "ExternalIdentity": {
      "type": "object",
      "additionalProperties": false,
      "required": ["type", "authProviderID", "authProviderType", "gitlabProvider"],
      "properties": {
        "type": {
          "type": "string",
          "const": "external"
        },
        "authProviderID": {
          "type": "string",
          "description": "The value of the ` + "`" + `configID` + "`" + ` field of the targeted authentication provider."
        },
        "authProviderType": {
          "type": "string",
          "description": "The ` + "`" + `type` + "`" + ` field of the targeted authentication provider."
        },
        "gitlabProvider": {
          "type": "string",
          "description": "The name that identifies the authentication provider to GitLab. This is passed to the ` + "`" + `?provider=` + "`" + ` query parameter in calls to the GitLab Users API. If you're not sure what this value is, you can look at the ` + "`" + `identities` + "`" + ` field of the GitLab Users API result (` + "`" + `curl  -H 'PRIVATE-TOKEN: $YOUR_TOKEN' $GITLAB_URL/api/v4/users` + "`" + `)."
        }
      }
    },
    "NameTransformation": {
      "title": "GitLabNameTransformation",
      "type": "object",
      "additionalProperties": false,
      "anyOf": [{ "required": ["regex", "replacement"] }],
      "properties": {
        "regex": {
          "type": "string",
          "format": "regex",
          "description": "The regex to match for the occurrences of its replacement."
        },
        "replacement": {
          "type": "string",
          "description": "The replacement used to replace all matched occurrences by the regex."
        }
      }
    }
  }
}
`
