package com.sourcegraph.website;

import com.intellij.openapi.actionSystem.AnActionEvent;
import com.intellij.openapi.diagnostic.Logger;
import com.intellij.openapi.editor.Document;
import com.intellij.openapi.editor.Editor;
import com.intellij.openapi.editor.SelectionModel;
import com.intellij.openapi.fileEditor.FileDocumentManager;
import com.intellij.openapi.fileEditor.FileEditorManager;
import com.intellij.openapi.project.DumbAwareAction;
import com.intellij.openapi.project.Project;
import com.intellij.openapi.vfs.VirtualFile;
import com.sourcegraph.browser.URLBuilder;
import com.sourcegraph.git.GitUtil;
import com.sourcegraph.git.RepoInfo;
import org.jetbrains.annotations.NotNull;
import org.jetbrains.annotations.Nullable;

import java.awt.*;
import java.io.IOException;
import java.net.URI;

public abstract class SearchActionBase extends DumbAwareAction {
    public void actionPerformedMode(AnActionEvent e, @NotNull Scope scope) {
        Logger logger = Logger.getInstance(this.getClass());

        // Get project, editor, document, file, and position information.
        final Project project = e.getProject();
        if (project == null) {
            return;
        }
        Editor editor = FileEditorManager.getInstance(project).getSelectedTextEditor();
        if (editor == null) {
            return;
        }
        Document currentDocument = editor.getDocument();
        VirtualFile currentFile = FileDocumentManager.getInstance().getFile(currentDocument);
        if (currentFile == null) {
            return;
        }

        RepoInfo repoInfo = GitUtil.getRepoInfo(currentFile.getPath(), project);

        SelectionModel selection = editor.getSelectionModel();
        String selectedText = selection.getSelectedText();
        if (selectedText == null || selectedText.equals("")) {
            return; // nothing to query
        }

        String remoteUrl = null;
        String branchName = null;
        if (scope == Scope.REPOSITORY) {
            remoteUrl = repoInfo.remoteUrl;
            branchName = repoInfo.branchName;
        }

        String uri = URLBuilder.buildEditorSearchUrl(project, selectedText, remoteUrl, branchName);

        // Open the URL in the browser.
        try {
            Desktop.getDesktop().browse(URI.create(uri));
        } catch (IOException err) {
            logger.debug("failed to open browser");
            err.printStackTrace();
        }
    }

    enum Scope {
        REPOSITORY,
        ANYWHERE
    }

    @Override
    public void update(@NotNull AnActionEvent e) {
        final Project project = e.getProject();
        if (project == null) {
            return;
        }
        String selectedText = getSelectedText(project);
        e.getPresentation().setEnabled(selectedText != null && selectedText.length() > 0);
    }

    @Nullable
    private String getSelectedText(@NotNull Project project) {
        Editor editor = FileEditorManager.getInstance(project).getSelectedTextEditor();
        if (editor == null) {
            return null;
        }
        Document currentDoc = editor.getDocument();
        VirtualFile currentFile = FileDocumentManager.getInstance().getFile(currentDoc);
        if (currentFile == null) {
            return null;
        }
        SelectionModel sel = editor.getSelectionModel();

        return sel.getSelectedText();
    }
}
