package com.sourcegraph.find;

import com.intellij.openapi.actionSystem.AnAction;
import com.intellij.openapi.actionSystem.AnActionEvent;
import com.intellij.openapi.actionSystem.Presentation;
import com.intellij.openapi.actionSystem.impl.ActionButton;
import com.intellij.openapi.options.ShowSettingsUtil;
import com.intellij.openapi.project.DumbAwareAction;
import com.intellij.openapi.project.Project;
import com.intellij.util.ui.JBDimension;
import com.intellij.util.ui.JBEmptyBorder;
import com.intellij.util.ui.JBUI;
import com.intellij.util.ui.components.BorderLayoutPanel;
import com.sourcegraph.Icons;
import com.sourcegraph.config.SettingsConfigurable;
import org.jetbrains.annotations.NotNull;

import javax.swing.*;
import java.awt.*;

public class HeaderPanel extends BorderLayoutPanel {
    private final ActionButton authenticateButton;

    public HeaderPanel(Project project) {
        super();
        setBorder(new JBEmptyBorder(5, 5, 2, 5));

        authenticateButton = createActionButtonThatOpensSettings(project, "Set Up Your Sourcegraph Account", Icons.Account);
        ActionButton settingsButton = createActionButtonThatOpensSettings(project, "Open Plugin Settings", Icons.GearPlain);

        JPanel title = new JPanel(new FlowLayout(FlowLayout.LEFT, 0, 0));
        title.setBorder(new JBEmptyBorder(2, 0, 0, 0));
        title.add(new JLabel("Find with Sourcegraph", Icons.SourcegraphLogo, SwingConstants.LEFT));

        JPanel buttons = new JPanel(new FlowLayout(FlowLayout.RIGHT, 0, 0));
        buttons.add(authenticateButton);
        buttons.add(settingsButton);

        add(title, BorderLayout.WEST);
        add(buttons, BorderLayout.EAST);
    }

    public void setAuthenticated(boolean authenticated) {
        authenticateButton.setVisible(!authenticated);
    }

    @NotNull
    private ActionButton createActionButtonThatOpensSettings(@NotNull Project project, @NotNull String label, @NotNull Icon icon) {
        JBDimension actionButtonSize = JBUI.size(22, 22);

        AnAction action = new DumbAwareAction() {
            @Override
            public void actionPerformed(@NotNull AnActionEvent anActionEvent) {
                ShowSettingsUtil.getInstance().showSettingsDialog(project, SettingsConfigurable.class);
            }
        };
        Presentation presentation = new Presentation(label);
        presentation.setIcon(icon);
        return new ActionButton(action, presentation, "Find with Sourcegraph popup header", actionButtonSize);
    }
}
