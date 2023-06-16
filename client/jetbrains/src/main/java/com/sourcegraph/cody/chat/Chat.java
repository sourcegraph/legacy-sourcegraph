package com.sourcegraph.cody.chat;

import com.intellij.openapi.project.Project;
import com.sourcegraph.cody.UpdatableChat;
import com.sourcegraph.cody.agent.CodyAgent;
import com.sourcegraph.cody.agent.protocol.ExecuteRecipeParams;
import com.sourcegraph.cody.api.ChatUpdaterCallbacks;
import com.sourcegraph.cody.api.CompletionsInput;
import com.sourcegraph.cody.api.CompletionsService;
import com.sourcegraph.cody.api.Message;
import com.sourcegraph.cody.api.Speaker;
import com.sourcegraph.cody.context.ContextFile;
import com.sourcegraph.cody.context.ContextGetter;
import com.sourcegraph.cody.context.ContextMessage;
import com.sourcegraph.cody.prompts.Preamble;
import com.sourcegraph.cody.vscode.CancellationToken;
import java.io.IOException;
import java.util.ArrayList;
import java.util.List;
import java.util.Objects;
import java.util.stream.Collectors;
import java.util.concurrent.ExecutionException;
import java.util.concurrent.atomic.AtomicBoolean;
import java.util.stream.Collectors;
import org.jetbrains.annotations.NotNull;
import org.jetbrains.annotations.Nullable;

public class Chat {
  private final @Nullable String repoName;
  private final @NotNull String instanceUrl;
  private final @NotNull String accessToken;
  private final @NotNull String customRequestHeaders;
  private final @NotNull CompletionsService completionsService;

  /**
   * @param repoName    Like "github.com/sourcegraph/cody"
   * @param instanceUrl Like "https://sourcegraph.com/", with a slash at the end
   */
  public Chat(
      @Nullable String repoName,
      @NotNull String instanceUrl,
      @NotNull String accessToken,
      @NotNull String customRequestHeaders) {
    this.repoName = repoName;
    this.instanceUrl = instanceUrl;
    this.accessToken = accessToken;
    this.customRequestHeaders = customRequestHeaders;
    completionsService = new CompletionsService(instanceUrl, accessToken);
  }

  public void sendMessage(
      @NotNull ChatMessage humanMessage, @Nullable String prefix, @NotNull UpdatableChat chat) {
    List<Message> preamble = Preamble.getPreamble(repoName);

    // TODO: Use the context getting logic from VS Code
    var editorContext = "";
    if (humanMessage.getContextFileContents().size() == 0) {
      editorContext = "I have no file open in the editor right now.";
    } else {
      editorContext = "Here is my current file\n" + humanMessage.getContextFileContents().get(0);
    }

    // Create completions input and add preamble
    var input = new CompletionsInput(new ArrayList<>(), 0.5f, null, 1000, -1, -1);
    input.addMessages(preamble);

    // Get context
    List<ContextMessage> context = null;
    if (repoName != null) {
      try {
        context =
            new ContextGetter(repoName, instanceUrl, accessToken, customRequestHeaders)
                .getContextMessages(humanMessage.getText(), 8, 2, true);
      } catch (IOException e) {
        chat.addMessageToChat(
            ChatMessage.createAssistantMessage(
                "I didn't get a correct response. This is what I encountered while trying to get some context for your ask: \""
                    + e.getMessage()
                    + "\". I'll try to answer without further context."));
      }
    }

    // Use context
    if (context != null) {
      if (context.size() == 0) {
        input.addMessage(
            Speaker.ASSISTANT,
            "I didn't find any context for your ask. I'll try to answer without further context.");
      } else {
        for (ContextMessage message : context) {
          input.addMessage(message.getSpeaker(), message.getText());
        }

        // Collect file names
        List<String> contextFileNames =
            context.stream()
                .map(ContextMessage::getFile)
                .filter(Objects::nonNull)
                .map(ContextFile::getFileName)
                .collect(Collectors.toList());

        // Build and add message
        StringBuilder contextMessageText =
            new StringBuilder(
                "I found some context for your ask. I'll try to answer with the context of these "
                    + contextFileNames.size()
                    + " files:\n");
        contextFileNames.forEach(fileName -> contextMessageText.append(fileName).append("\n"));
        chat.addMessageToChat(ChatMessage.createAssistantMessage(contextMessageText.toString()));
      }
    }

    // Add editor context
    input.addMessage(Speaker.HUMAN, editorContext);
    input.addMessage(Speaker.ASSISTANT, "Ok.");

    // Add human message
    input.addMessage(Speaker.HUMAN, humanMessage.getText());
    input.addMessage(Speaker.ASSISTANT, "");

    completionsService.streamCompletion(
        input,
        new ChatUpdaterCallbacks(chat, prefix),
        CompletionsService.Endpoint.Stream,
        new CancellationToken());
  }

  private void sendMessageViaAgent(
      @NotNull Project project, @NotNull ChatMessage humanMessage, @NotNull UpdatableChat chat)
      throws ExecutionException, InterruptedException {
    final AtomicBoolean isFirstMessage = new AtomicBoolean(false);
    CodyAgent.getClient(project).onChatUpdateMessageInProgress =
        (agentChatMessage) -> {
          if (agentChatMessage.text == null) {
            return;
          }
          ChatMessage chatMessage =
              new ChatMessage(
                  Speaker.ASSISTANT,
                  agentChatMessage.text,
                  agentChatMessage.displayText,
                  agentChatMessage.contextFiles.stream()
                      .map(file -> file.fileName)
                      .collect(Collectors.toList()));
          if (isFirstMessage.compareAndSet(false, true)) {
            chat.addMessageToChat(chatMessage);
          } else {
            chat.updateLastMessage(chatMessage);
          }
        };

    CodyAgent.getInitializedServer(project)
        .thenAcceptAsync(
            server ->
                server.recipesExecute(
                    new ExecuteRecipeParams()
                        .setId("chat-question")
                        .setHumanChatInput(humanMessage.getText())),
            CodyAgent.executorService)
        .get();
    chat.finishMessageProcessing();
  }
}
