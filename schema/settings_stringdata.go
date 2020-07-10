// Code generated by stringdata. DO NOT EDIT.

package schema

// SettingsSchemaJSON is the content of the file "settings.schema.json".
const SettingsSchemaJSON = `{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "$id": "settings.schema.json#",
  "title": "Settings",
  "description": "Configuration settings for users and organizations on Sourcegraph.",
  "allowComments": true,
  "type": "object",
  "properties": {
    "experimentalFeatures": {
      "title": "SettingsExperimentalFeatures",
      "description": "Experimental features to enable or disable. Features that are now enabled by default are marked as deprecated.",
      "type": "object",
      "additionalProperties": false,
      "properties": {
        "splitSearchModes": {
          "description": "Enables toggling between the current omni search mode, and experimental interactive search mode.",
          "type": "boolean",
          "default": true,
          "!go": { "pointer": true }
        },
        "codeInsights": {
          "description": "Enables code insights on directory pages.",
          "type": "boolean",
          "default": false,
          "!go": { "pointer": true }
        },
        "searchStats": {
          "description": "Enables a new page that shows language statistics about the results for a search query.",
          "type": "boolean",
          "default": false,
          "!go": { "pointer": true }
        },
        "showBadgeAttachments": {
          "description": "Enables the UI indicators for code intelligence precision.",
          "type": "boolean",
          "default": true,
          "!go": { "pointer": true }
        },
        "smartSearchField": {
          "description": "Enables displaying a search field that provides syntax highlighting, hover tooltips and diagnostics for search queries.",
          "type": "boolean",
          "default": true,
          "!go": { "pointer": true }
        },
        "copyQueryButton": {
          "description": "Enables displaying the copy query button in the search bar when hovering over the global navigation bar.",
          "type": "boolean",
          "default": false,
          "!go": { "pointer": true }
        },
        "showRepogroupHomepage": {
          "description": "Enables the repository group homepage ",
          "type": "boolean",
          "default": false,
          "!go": { "pointer": true }
        }
      },
      "group": "Experimental"
    },
    "search.savedQueries": {
      "description": "DEPRECATED: Saved search queries",
      "type": "array",
      "items": {
        "type": "object",
        "properties": {
          "key": {
            "type": "string",
            "description": "Unique key for this query in this file"
          },
          "description": {
            "type": "string",
            "description": "Description of this saved query"
          },
          "query": {
            "type": "string",
            "description": "Query string"
          },
          "showOnHomepage": {
            "type": "boolean",
            "description": "DEPRECATED: saved searches are no longer shown on the homepage. This will be removed in a future release."
          },
          "notify": {
            "type": "boolean",
            "description": "Notify the owner of this configuration file when new results are available"
          },
          "notifySlack": {
            "type": "boolean",
            "description": "Notify Slack via the organization's Slack webhook URL when new results are available"
          }
        },
        "additionalProperties": false,
        "required": ["key", "description", "query"]
      }
    },
    "search.scopes": {
      "description": "Predefined search scopes",
      "type": "array",
      "items": {
        "$ref": "#/definitions/SearchScope"
      }
    },
    "search.repositoryGroups": {
      "description": "Named groups of repositories that can be referenced in a search query using the repogroup: operator.",
      "type": "object",
      "additionalProperties": {
        "type": "array",
        "items": { "type": "string" }
      }
    },
    "search.contextLines": {
      "description": "The default number of lines to show as context below and above search results. Default is 1.",
      "type": "integer",
      "minimum": 0,
      "default": 1
    },
    "search.defaultPatternType": {
      "description": "The default pattern type (literal or regexp) that search queries will be intepreted as.",
      "type": "string",
      "pattern": "literal|regexp"
    },
    "search.includeForks": {
      "description": "Whether searches should include searching forked repositories.",
      "type": "boolean",
      "default": false,
      "!go": { "pointer": true }
    },
    "search.includeArchived": {
      "description": "Whether searches should include searching archived repositories.",
      "type": "boolean",
      "default": false,
      "!go": { "pointer": true }
    },
    "quicklinks": {
      "description": "Links that should be accessible quickly from the home and search pages.",
      "type": "array",
      "items": {
        "$ref": "#/definitions/QuickLink"
      }
    },
    "motd": {
      "description": "DEPRECATED: Use ` + "`" + `notices` + "`" + ` instead.\n\nAn array (often with just one element) of messages to display at the top of all pages, including for unauthenticated users. Users may dismiss a message (and any message with the same string value will remain dismissed for the user).\n\nMarkdown formatting is supported.\n\nUsually this setting is used in global and organization settings. If set in user settings, the message will only be displayed to that user. (This is useful for testing the correctness of the message's Markdown formatting.)\n\nMOTD stands for \"message of the day\" (which is the conventional Unix name for this type of message).",
      "type": "array",
      "items": { "type": "string" }
    },
    "notices": {
      "description": "Custom informational messages to display to users at specific locations in the Sourcegraph user interface.\n\nUsually this setting is used in global and organization settings. If set in user settings, the message will only be displayed to that single user.",
      "type": "array",
      "items": {
        "title": "Notice",
        "type": "object",
        "required": ["message", "location"],
        "properties": {
          "message": {
            "description": "The message to display. Markdown formatting is supported.",
            "type": "string"
          },
          "location": {
            "description": "The location where this notice is shown: \"top\" for the top of every page, \"home\" for the homepage.",
            "type": "string",
            "enum": ["top", "home"]
          },
          "dismissible": {
            "description": "Whether this notice can be dismissed (closed) by the user.",
            "type": "boolean",
            "default": false
          }
        }
      }
    },
    "alerts.showPatchUpdates": {
      "description": "Whether to show alerts for patch version updates. Alerts for major and minor version updates will always be shown.",
      "type": "boolean",
      "default": true
    },
    "alerts.hideObservabilitySiteAlerts": {
      "description": "Disables observability-related site alert banners.",
      "type": "boolean",
      "default": false
    },
    "extensions": {
      "description": "The Sourcegraph extensions to use. Enable an extension by adding a property ` + "`" + `\"my/extension\": true` + "`" + ` (where ` + "`" + `my/extension` + "`" + ` is the extension ID). Override a previously enabled extension and disable it by setting its value to ` + "`" + `false` + "`" + `.",
      "type": "object",
      "propertyNames": {
        "type": "string",
        "description": "A valid extension ID.",
        "pattern": "^([^/]+/)?[^/]+/[^/]+$"
      },
      "additionalProperties": {
        "type": "boolean",
        "description": "` + "`" + `true` + "`" + ` to enable the extension, ` + "`" + `false` + "`" + ` to disable the extension (if it was previously enabled)"
      }
    },
    "codeHost.useNativeTooltips": {
      "description": "Whether to use the code host's native hover tooltips when they exist (GitHub's jump-to-definition tooltips, for example).",
      "type": "boolean",
      "default": false
    },
    "search.uppercase": {
      "description": "When active, any uppercase characters in the pattern will make the entire query case-sensitive.",
      "type": "boolean",
      "default": false,
      "!go": { "pointer": true }
    },
    "search.migrateParser": {
      "description": "If true, uses the new and/or-compatible parser for all search queries. It is a flag to aid transition to the new parser.",
      "type": "boolean",
      "default": false,
      "!go": { "pointer": true }
    }
  },
  "definitions": {
    "SearchScope": {
      "type": "object",
      "additionalProperties": false,
      "required": ["name", "value"],
      "properties": {
        "id": {
          "type": "string",
          "description": "A unique identifier for the search scope.\n\nIf set, a scoped search page is available at https://[sourcegraph-hostname]/search/scope/ID, where ID is this value."
        },
        "name": {
          "type": "string",
          "description": "The human-readable name for this search scope"
        },
        "value": {
          "type": "string",
          "description": "The query string of this search scope"
        },
        "description": {
          "type": "string",
          "description": "A description for this search scope"
        }
      }
    },
    "QuickLink": {
      "type": "object",
      "additionalProperties": false,
      "required": ["name", "url"],
      "properties": {
        "name": {
          "type": "string",
          "description": "The human-readable name for this quick link"
        },
        "url": {
          "type": "string",
          "description": "The URL of this quick link (absolute or relative)"
        },
        "description": {
          "type": "string",
          "description": "A description for this quick link"
        }
      }
    }
  },
  "search.hideSuggestions": {
    "description": "Disable search suggestions below the search bar when constructing queries. Defaults to false.",
    "type": "boolean",
    "default": false,
    "!go": { "pointer": true }
  }
}
`
