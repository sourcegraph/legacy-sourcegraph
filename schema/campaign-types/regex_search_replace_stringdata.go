// Code generated by stringdata. DO NOT EDIT.

package schema

// RegexSearchReplaceCampaignTypeSchemaJSON is the content of the file "campaign-types/regex_search_replace.schema.json".
const RegexSearchReplaceCampaignTypeSchemaJSON = `{
  "$id": "regex-search-replace-spec.json#",
  "$schema": "http://json-schema.org/draft-07/schema#",
  "description": "Schema for regexp search replace",
  "type": "object",
  "properties": {
    "scopeQuery": {
      "type": "string",
      "minLength": 1,
      "description": "Define a scope to narrow down repositories affected by this change. Only GitHub and Bitbucket Server are supported."
    },
    "regexpMatch": {
      "type": "string",
      "minLength": 1,
      "description": "Match this regular expression. RE2 syntax is supported: https://github.com/google/re2/wiki/Syntax"
    },
    "textReplace": {
      "type": "string",
      "description": "Replace the regexpMatch text with this text. You may refer to match groups in regexpMatch using $id or ${id} syntax."
    }
  },
  "required": ["scopeQuery", "regexpMatch", "textReplace"],
  "additionalProperties": false
}
`
