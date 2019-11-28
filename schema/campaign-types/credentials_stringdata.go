// Code generated by stringdata. DO NOT EDIT.

package schema

// CredentialsCampaignTypeSchemaJSON is the content of the file "campaign-types/credentials.schema.json".
const CredentialsCampaignTypeSchemaJSON = `{
  "$id": "credentials-spec.json#",
  "$schema": "http://json-schema.org/draft-07/schema#",
  "description": "Schema for credentials options",
  "type": "object",
  "properties": {
    "matchers": {
      "type": "array",
      "description": "Define matchers to be used",
      "minItems": 1,
      "uniqueItems": true,
      "additionalItems": false,
      "items": {
        "description": "Defines a matcher and it's associated configuration",
        "oneOf": [
          {
            "description": "Checks for .npmrc files with included secrets",
            "type": "object",
            "properties": {
              "type": {
                "type": "string",
                "enum": ["npm"]
              }
            },
            "required": ["type"],
            "additionalProperties": false
          }
        ]
      }
    },
    "scopeQuery": {
      "type": "string",
      "description": "Define a scope to narrow down repositories affected by this change. Only GitHub and Bitbucket Server are supported."
    }
  },
  "required": ["matchers", "scopeQuery"],
  "additionalProperties": false
}
`
