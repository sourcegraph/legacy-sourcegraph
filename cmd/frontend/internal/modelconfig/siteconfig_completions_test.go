package modelconfig

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/licensing"
	"github.com/sourcegraph/sourcegraph/internal/modelconfig/types"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestConvertCompletionsConfig(t *testing.T) {
	// Mock out the licensing check confirming Cody is enabled.
	initialMockCheck := licensing.MockCheckFeature
	licensing.MockCheckFeature = func(licensing.Feature) error {
		return nil // Don't fail when checking if Cody is enabled.
	}
	t.Cleanup(func() { licensing.MockCheckFeature = initialMockCheck })
	// Restore our mocked out site config.
	t.Cleanup(func() { conf.Mock(nil) })

	// loadCompletionsConfig sets the supplied completions configuration data
	// into the site config, and then loads it. This extra step ensures that
	// the default values are set as applicable in the returned object.
	// as well as the necessary checks for enabling Cody Pro.
	loadCompletionsConfig := func(userSuppliedCompConfig schema.Completions) *conftypes.CompletionsConfig {
		fauxSiteConfig := schema.SiteConfiguration{
			CodyEnabled:                  pointers.Ptr(true),
			CodyPermissions:              pointers.Ptr(false),
			CodyRestrictUsersFeatureFlag: pointers.Ptr(false),
			LicenseKey:                   "license-key",

			Completions: &userSuppliedCompConfig,
		}
		return conf.GetCompletionsConfig(fauxSiteConfig)
	}
	t.Run("Default", func(t *testing.T) {
		compConfig := loadCompletionsConfig(schema.Completions{
			Provider: "sourcegraph",
		})
		require.NotNil(t, compConfig)

		siteModelConfig, err := convertCompletionsConfig(compConfig)
		require.NoError(t, err)

		assert.Nil(t, siteModelConfig.SourcegraphModelConfig)
		require.NotNil(t, siteModelConfig.ProviderOverrides)
		require.NotNil(t, siteModelConfig.ModelOverrides)

		// ProviderOverrides. Because the default models are from different providers, we stub out
		// three different ProviderOverrides. However, all of these are configured to use the
		// "Sourcegraph API Provider".
		require.Equal(t, 2, len(siteModelConfig.ProviderOverrides))
		assert.EqualValues(t, "anthropic", siteModelConfig.ProviderOverrides[0].ID)
		assert.EqualValues(t, "fireworks", siteModelConfig.ProviderOverrides[1].ID)

		for _, providerOverride := range siteModelConfig.ProviderOverrides {
			// Stock model configuration.
			defModelCfg := providerOverride.DefaultModelConfig
			require.NotNil(t, defModelCfg)
			assert.Equal(t, types.ModelTierEnterprise, defModelCfg.Tier)

			require.Nil(t, providerOverride.ClientSideConfig)

			ssConfig := providerOverride.ServerSideConfig
			require.NotNil(t, ssConfig)
			require.NotNil(t, ssConfig.SourcegraphProvider)

			sgAPIProviderConfig := ssConfig.SourcegraphProvider
			require.NotNil(t, sgAPIProviderConfig)
			assert.Equal(t, "https://cody-gateway.sourcegraph.com", sgAPIProviderConfig.Endpoint)
			assert.NotEmpty(t, sgAPIProviderConfig.AccessToken) // Based on the license key.
		}

		// ModelOverrides
		require.Equal(t, 3, len(siteModelConfig.ModelOverrides))
		assert.EqualValues(t, "anthropic::unknown::claude-3-haiku-20240307", siteModelConfig.ModelOverrides[0].ModelRef)
		assert.EqualValues(t, "anthropic::unknown::claude-3-sonnet-20240229", siteModelConfig.ModelOverrides[1].ModelRef)
		assert.EqualValues(t, "fireworks::unknown::starcoder", siteModelConfig.ModelOverrides[2].ModelRef)

		// DefaultModels
		require.NotNil(t, siteModelConfig.DefaultModels)
		assert.EqualValues(t, "anthropic::unknown::claude-3-haiku-20240307", siteModelConfig.DefaultModels.FastChat)
		assert.EqualValues(t, "anthropic::unknown::claude-3-sonnet-20240229", siteModelConfig.DefaultModels.Chat)
		assert.EqualValues(t, "fireworks::unknown::starcoder", siteModelConfig.DefaultModels.CodeCompletion)
	})

	t.Run("OpenAI", func(t *testing.T) {
		compConfig := loadCompletionsConfig(schema.Completions{
			Provider:        "openai",
			ChatModel:       "gpt-4",
			FastChatModel:   "gpt-3.5-turbo",
			CompletionModel: "gpt-3.5-turbo-instruct",
			AccessToken:     "byok-key",
		})
		require.NotNil(t, compConfig)

		siteModelConfig, err := convertCompletionsConfig(compConfig)
		require.NoError(t, err)

		assert.Nil(t, siteModelConfig.SourcegraphModelConfig)
		require.NotNil(t, siteModelConfig.ProviderOverrides)
		require.NotNil(t, siteModelConfig.ModelOverrides)

		// ProviderOverrides. Default to using "sourcegraph" and Cody Gateway.
		require.Equal(t, 1, len(siteModelConfig.ProviderOverrides))
		providerOverride := siteModelConfig.ProviderOverrides[0]
		assert.EqualValues(t, "openai", providerOverride.ID)
		require.NotNil(t, providerOverride.ServerSideConfig)

		genericProviderConfig := providerOverride.ServerSideConfig.GenericProvider
		require.NotNil(t, genericProviderConfig)
		assert.Equal(t, "https://api.openai.com", genericProviderConfig.Endpoint)
		assert.NotEmpty(t, "byok-key", genericProviderConfig.AccessToken)

		// ModelOverrides
		require.Equal(t, 3, len(siteModelConfig.ModelOverrides))

		// DefaultModels
		require.NotNil(t, siteModelConfig.DefaultModels)
		assert.EqualValues(t, "openai::unknown::gpt-4", siteModelConfig.DefaultModels.Chat)
		assert.EqualValues(t, "openai::unknown::gpt-3.5-turbo", siteModelConfig.DefaultModels.FastChat)
		assert.EqualValues(t, "openai::unknown::gpt-3.5-turbo-instruct", siteModelConfig.DefaultModels.CodeCompletion)
	})

	t.Run("AWSBedrock", func(t *testing.T) {
		t.Run("OnDemandThoughput", func(t *testing.T) {
			compConfig := loadCompletionsConfig(schema.Completions{
				Provider:        "aws-bedrock",
				ChatModel:       "anthropic.claude-3-opus-20240229-v1:0",
				CompletionModel: "anthropic.claude-instant-v1",
				// FastChatModel not set, left to default.
				AccessToken: "", // Leave blank to pick up ambient AWS creds.
				Endpoint:    "us-west-2",
			})
			require.NotNil(t, compConfig)

			siteModelConfig, err := convertCompletionsConfig(compConfig)
			require.NoError(t, err)

			assert.Nil(t, siteModelConfig.SourcegraphModelConfig)
			require.NotNil(t, siteModelConfig.ProviderOverrides)
			require.NotNil(t, siteModelConfig.ModelOverrides)

			// The ID of the ProviderOverride is "anthropic", to match the models referenced.
			// However, the API Provider, i.e. the ProviderOverride's server-side configuration
			// will define how to _use_ this provider, which will be via AWS Bedrock.
			require.Equal(t, 1, len(siteModelConfig.ProviderOverrides))
			providerOverride := siteModelConfig.ProviderOverrides[0]
			assert.EqualValues(t, "anthropic", providerOverride.ID)
			require.NotNil(t, providerOverride.ServerSideConfig)

			awsBedrockConfig := providerOverride.ServerSideConfig.AWSBedrock
			require.NotNil(t, awsBedrockConfig)
			assert.Equal(t, "us-west-2", awsBedrockConfig.Endpoint)
			assert.Equal(t, "", awsBedrockConfig.AccessToken)

			// ModelOverrides
			require.Equal(t, 2, len(siteModelConfig.ModelOverrides))
			{
				m := siteModelConfig.ModelOverrides[1]
				assert.EqualValues(t, "anthropic::unknown::anthropic.claude-instant-v1", m.ModelRef)
				require.Nil(t, m.ServerSideConfig)
			}
			{
				m := siteModelConfig.ModelOverrides[0]
				// Notice how the model ID has been sanitized. (No colon.) But the model name is the same
				// from the site config. (Since that's how the model is identified in its backing API.)
				assert.EqualValues(t, "anthropic::unknown::anthropic.claude-3-opus-20240229-v1_0", m.ModelRef)
				assert.EqualValues(t, "anthropic.claude-3-opus-20240229-v1_0", m.ModelRef.ModelID())
				// Unlike the Model's ID, the Name is unchanged, as this is what AWS expects in the API call.
				assert.EqualValues(t, "anthropic.claude-3-opus-20240229-v1:0", m.ModelName)
				require.Nil(t, m.ServerSideConfig)
			}

			// DefaultModels
			require.NotNil(t, siteModelConfig.DefaultModels)
			assert.EqualValues(t, "anthropic::unknown::anthropic.claude-3-opus-20240229-v1_0", siteModelConfig.DefaultModels.Chat)
			assert.EqualValues(t, "anthropic::unknown::anthropic.claude-instant-v1", siteModelConfig.DefaultModels.FastChat)
			assert.EqualValues(t, "anthropic::unknown::anthropic.claude-instant-v1", siteModelConfig.DefaultModels.CodeCompletion)
		})

		t.Run("ProvisionedThroughput", func(t *testing.T) {
			compConfig := loadCompletionsConfig(schema.Completions{
				Provider:        "aws-bedrock",
				ChatModel:       "anthropic.claude-3-haiku-20240307-v1:0-100k/arn:aws:bedrock:us-west-2:012345678901:provisioned-model/abcdefghijkl",
				CompletionModel: "anthropic.claude-instant-v1",
				// FastChatModel not set, left to default.
				AccessToken: "access-key-id:secret-access-key:session-token",
				Endpoint:    "https://vpce-0000-00000.bedrock-runtime.us-west-2.vpce.amazonaws.com",
			})
			require.NotNil(t, compConfig)

			siteModelConfig, err := convertCompletionsConfig(compConfig)
			require.NoError(t, err)

			assert.Nil(t, siteModelConfig.SourcegraphModelConfig)
			require.NotNil(t, siteModelConfig.ProviderOverrides)
			require.NotNil(t, siteModelConfig.ModelOverrides)

			// ProviderOverrides.
			require.Equal(t, 1, len(siteModelConfig.ProviderOverrides))
			providerOverride := siteModelConfig.ProviderOverrides[0]
			assert.EqualValues(t, "anthropic", providerOverride.ID)
			require.NotNil(t, providerOverride.ServerSideConfig)

			awsBedrockConfig := providerOverride.ServerSideConfig.AWSBedrock
			require.NotNil(t, awsBedrockConfig)
			assert.Equal(t, "access-key-id:secret-access-key:session-token", awsBedrockConfig.AccessToken)
			assert.Equal(t, "https://vpce-0000-00000.bedrock-runtime.us-west-2.vpce.amazonaws.com", awsBedrockConfig.Endpoint)

			// ModelOverrides
			require.Equal(t, 2, len(siteModelConfig.ModelOverrides))

			chatModel := siteModelConfig.ModelOverrides[0]
			assert.EqualValues(t, "anthropic::unknown::anthropic.claude-3-haiku-20240307-v1_0-100k", chatModel.ModelRef)
			assert.EqualValues(t, "anthropic.claude-3-haiku-20240307-v1:0-100k", chatModel.ModelName)
			require.NotNil(t, chatModel.ServerSideConfig)
			require.NotNil(t, chatModel.ServerSideConfig.AWSBedrockProvisionedThroughput)
			assert.Equal(t, "arn:aws:bedrock:us-west-2:012345678901:provisioned-model/abcdefghijkl", chatModel.ServerSideConfig.AWSBedrockProvisionedThroughput.ARN)

			completionModel := siteModelConfig.ModelOverrides[1]
			assert.EqualValues(t, "anthropic::unknown::anthropic.claude-instant-v1", completionModel.ModelRef)
			assert.Nil(t, completionModel.ServerSideConfig)

			// DefaultModels. Note the that model was modified, such as stripping out the ARN.
			require.NotNil(t, siteModelConfig.DefaultModels)
			assert.EqualValues(t, "anthropic::unknown::anthropic.claude-3-haiku-20240307-v1_0-100k", siteModelConfig.DefaultModels.Chat)
			assert.EqualValues(t, "anthropic::unknown::anthropic.claude-instant-v1", siteModelConfig.DefaultModels.FastChat)
			assert.EqualValues(t, "anthropic::unknown::anthropic.claude-instant-v1", siteModelConfig.DefaultModels.CodeCompletion)
		})
	})

	t.Run("AzureOpenAI", func(t *testing.T) {
		t.Run("ExplicitAzureModelNames", func(t *testing.T) {
			compConfig := loadCompletionsConfig(schema.Completions{
				Provider:        "azure-openai",
				ChatModel:       "azure-deployment-id - {7d1f5676-189d-4a97-9dd9-10ae187aba99}",
				CompletionModel: "azure-deployment-id - {a3ffed36-e0da-4005-ab3d-04e37889fd44}",
				FastChatModel:   "azure-deployment-id - {54e1331b-4ccc-44b8-b9ef-59c6b050fdb8}",
				AccessToken:     "azure-portal-api-key",
				Endpoint:        "https://azure-openai.azure.com",

				// Supply model IDs distinct from the Azure deployment information.
				// This was added so we can associate token counts to more useful identifiers,
				// but we use the data to provide better ModelRefs as well.
				AzureChatModel:       "gpt-3.5",
				AzureCompletionModel: "gpt-4o",

				User: "steve holt",
				AzureUseDeprecatedCompletionsAPIForOldModels: true,
			})
			require.NotNil(t, compConfig)

			siteModelConfig, err := convertCompletionsConfig(compConfig)
			require.NoError(t, err)

			assert.Nil(t, siteModelConfig.SourcegraphModelConfig)
			require.NotNil(t, siteModelConfig.ProviderOverrides)
			require.NotNil(t, siteModelConfig.ModelOverrides)

			// The ID of the ProviderOverride is "anthropic", to match the models referenced.
			// However, the API Provider, i.e. the ProviderOverride's server-side configuration
			// will define how to _use_ this provider, which will be via AWS Bedrock.
			require.Equal(t, 1, len(siteModelConfig.ProviderOverrides))
			providerOverride := siteModelConfig.ProviderOverrides[0]
			assert.EqualValues(t, "openai", providerOverride.ID)
			require.NotNil(t, providerOverride.ServerSideConfig)

			azureProviderConfig := providerOverride.ServerSideConfig.AzureOpenAI
			require.NotNil(t, azureProviderConfig)
			assert.Equal(t, "https://azure-openai.azure.com", azureProviderConfig.Endpoint)
			assert.Equal(t, "azure-portal-api-key", azureProviderConfig.AccessToken)
			assert.Equal(t, "steve holt", azureProviderConfig.User)
			assert.True(t, azureProviderConfig.UseDeprecatedCompletionsAPI)

			// ModelOverrides
			// For AzureOpenAI, we verify that the "ModelName" and "ModelID" are set independently, based on
			// the configuration data.
			require.Equal(t, 3, len(siteModelConfig.ModelOverrides))
			{
				// BUG: The "model name alias" configuration doesn't apply to FastChat, hence the awkward ModelID.
				m := siteModelConfig.ModelOverrides[0]
				assert.EqualValues(t, "openai::unknown::azure-deployment-id_-__54e1331b-4ccc-44b8-b9ef-59c6b050fdb8_", m.ModelRef)
				assert.EqualValues(t, "azure-deployment-id - {54e1331b-4ccc-44b8-b9ef-59c6b050fdb8}", m.ModelName)
				require.Nil(t, m.ServerSideConfig)
			}
			{
				m := siteModelConfig.ModelOverrides[1]
				assert.EqualValues(t, "openai::unknown::gpt-3.5", m.ModelRef)
				assert.EqualValues(t, "azure-deployment-id - {7d1f5676-189d-4a97-9dd9-10ae187aba99}", m.ModelName)
				require.Nil(t, m.ServerSideConfig)
			}
			{
				m := siteModelConfig.ModelOverrides[2]
				assert.EqualValues(t, "openai::unknown::gpt-4o", m.ModelRef)
				assert.EqualValues(t, "azure-deployment-id - {a3ffed36-e0da-4005-ab3d-04e37889fd44}", m.ModelName)
				require.Nil(t, m.ServerSideConfig)
			}

			// DefaultModels
			require.NotNil(t, siteModelConfig.DefaultModels)
			assert.EqualValues(t, "openai::unknown::gpt-3.5", siteModelConfig.DefaultModels.Chat)
			assert.EqualValues(t, "openai::unknown::azure-deployment-id_-__54e1331b-4ccc-44b8-b9ef-59c6b050fdb8_", siteModelConfig.DefaultModels.FastChat)
			assert.EqualValues(t, "openai::unknown::gpt-4o", siteModelConfig.DefaultModels.CodeCompletion)
		})
	})

	t.Run("MaxTokens", func(t *testing.T) {
		t.Run("Aliasing", func(t *testing.T) {
			compConfig := loadCompletionsConfig(schema.Completions{
				Provider: "sourcegraph",

				// All 3x kinds of models from in the config point to the same model ID.
				// But but one of them has a different MaxTokens set. So we need to
				// update the ModelRef to disambiguate this case.
				ChatModel:       "model-x",
				CompletionModel: "model-x",
				FastChatModel:   "model-x",

				ChatModelMaxTokens:       10_000,
				CompletionModelMaxTokens: 10_000,
				FastChatModelMaxTokens:   5_000,
			})
			require.NotNil(t, compConfig)

			siteModelConfig, err := convertCompletionsConfig(compConfig)
			require.NoError(t, err)

			// DefaultModels
			require.NotNil(t, siteModelConfig.DefaultModels)
			// Yes, it would make more sense to have "model-x_fast" instead of suffixing the two that shared
			// an alias. But this is dependent on the ordering we check models for deduping.
			assert.EqualValues(t, "sourcegraph::unknown::model-x_chat", siteModelConfig.DefaultModels.Chat)
			assert.EqualValues(t, "sourcegraph::unknown::model-x_chat", siteModelConfig.DefaultModels.CodeCompletion)
			assert.EqualValues(t, "sourcegraph::unknown::model-x", siteModelConfig.DefaultModels.FastChat)

			// ModelOverrides. We only need two. Because Chat and Completions are using the same model,
			// with the same number of max tokens.
			require.Equal(t, 2, len(siteModelConfig.ModelOverrides))
			{
				model := siteModelConfig.ModelOverrides[0]
				assert.EqualValues(t, "sourcegraph::unknown::model-x", model.ModelRef)
				assert.EqualValues(t, 5000, model.ContextWindow.MaxInputTokens)
			}
			{
				model := siteModelConfig.ModelOverrides[1]
				assert.EqualValues(t, "sourcegraph::unknown::model-x_chat", model.ModelRef)
				assert.EqualValues(t, 10_000, model.ContextWindow.MaxInputTokens)
			}
		})

		t.Run("3-Way", func(t *testing.T) {
			compConfig := loadCompletionsConfig(schema.Completions{
				Provider: "sourcegraph",

				ChatModel:       "model-x",
				CompletionModel: "model-x",
				FastChatModel:   "model-x",

				ChatModelMaxTokens:       1_000,
				CompletionModelMaxTokens: 2_000,
				FastChatModelMaxTokens:   3_000,
			})
			require.NotNil(t, compConfig)

			siteModelConfig, err := convertCompletionsConfig(compConfig)
			require.NoError(t, err)

			// DefaultModels
			require.NotNil(t, siteModelConfig.DefaultModels)
			// Yes, it would make more sense to have "model-x_fast" instead of suffixing the two that shared
			// an alias. But this is dependent on the ordering we check models for deduping.
			assert.EqualValues(t, "sourcegraph::unknown::model-x_chat", siteModelConfig.DefaultModels.Chat)
			assert.EqualValues(t, "sourcegraph::unknown::model-x_completion", siteModelConfig.DefaultModels.CodeCompletion)
			assert.EqualValues(t, "sourcegraph::unknown::model-x", siteModelConfig.DefaultModels.FastChat)

			// The ModelOverrides slice is sorted by ModelRef, hence why
			// the ModelRef that wasn't renamed ("model-x") comes first.
			require.Equal(t, 3, len(siteModelConfig.ModelOverrides))
			{
				model := siteModelConfig.ModelOverrides[0]
				assert.EqualValues(t, "sourcegraph::unknown::model-x", model.ModelRef)
				assert.EqualValues(t, 3_000, model.ContextWindow.MaxInputTokens)
			}
			{
				model := siteModelConfig.ModelOverrides[1]
				assert.EqualValues(t, "sourcegraph::unknown::model-x_chat", model.ModelRef)
				assert.EqualValues(t, 1_000, model.ContextWindow.MaxInputTokens)
			}
			{
				model := siteModelConfig.ModelOverrides[2]
				assert.EqualValues(t, "sourcegraph::unknown::model-x_completion", model.ModelRef)
				assert.EqualValues(t, 2_000, model.ContextWindow.MaxInputTokens)
			}
		})
	})

	t.Run("Starcoder", func(t *testing.T) {
		compConfig := loadCompletionsConfig(schema.Completions{
			Provider:        "sourcegraph",
			CompletionModel: "fireworks/starcoder",
		})
		require.NotNil(t, compConfig)

		siteModelConfig, err := convertCompletionsConfig(compConfig)
		require.NoError(t, err)

		// DefaultModels. Claude is used for chat, but the code completions were explicitly set
		// to StarCoder.
		require.NotNil(t, siteModelConfig.DefaultModels)
		assert.EqualValues(t, "anthropic::unknown::claude-3-sonnet-20240229", siteModelConfig.DefaultModels.Chat)
		assert.EqualValues(t, "fireworks::unknown::starcoder", siteModelConfig.DefaultModels.CodeCompletion)
		assert.EqualValues(t, "anthropic::unknown::claude-3-haiku-20240307", siteModelConfig.DefaultModels.FastChat)

		// We set the modle's name to just "starcoder" because that's all the information we have
		// available. But on the Cody Gateway side, it is virtualized. And will actually use a
		// "real" model name like "starcoder2-7b" when resolving the request.
		//
		// See `pickStarCoderModel` in the Cody Gateway code.
		scModel := siteModelConfig.ModelOverrides[2]
		assert.EqualValues(t, "fireworks::unknown::starcoder", scModel.ModelRef)
		assert.EqualValues(t, "starcoder", scModel.ModelName)
	})
}
