package com.sourcegraph.find;

import com.intellij.openapi.Disposable;
import com.intellij.openapi.project.Project;
import com.intellij.openapi.ui.popup.ActiveIcon;
import com.intellij.openapi.ui.popup.ComponentPopupBuilder;
import com.intellij.openapi.ui.popup.JBPopup;
import com.intellij.openapi.ui.popup.JBPopupFactory;
import com.intellij.openapi.util.Disposer;
import com.sourcegraph.Icons;
import org.cef.browser.CefBrowser;
import org.cef.handler.CefKeyboardHandler;
import org.cef.misc.BoolRef;
import org.jetbrains.annotations.NotNull;

import java.awt.*;
import java.awt.event.KeyEvent;

import static java.awt.event.InputEvent.ALT_DOWN_MASK;

public class SourcegraphWindow implements Disposable {
    private final Project project;
    private final FindPopupPanel mainPanel;
    private JBPopup popup;

    public SourcegraphWindow(@NotNull Project project) {
        this.project = project;

        // Create main panel
        mainPanel = new FindPopupPanel(project);

        Disposer.register(project, this);
    }

    synchronized public void showPopup() {
        if (popup == null || popup.isDisposed()) {
            popup = createPopup();
            popup.showCenteredInCurrentWindow(project);
        }

        popup.setUiVisible(true);

        // If the popup is already shown, hitting alt + a gain should behave the same as the native find in files
        // feature and focus the search field.
        if (mainPanel.getBrowser() != null) {
            mainPanel.getBrowser().focus();
        }
    }

    @NotNull
    private JBPopup createPopup() {
        ComponentPopupBuilder builder = JBPopupFactory.getInstance().createComponentPopupBuilder(mainPanel, mainPanel)
            .setTitle("Sourcegraph")
            .setTitleIcon(new ActiveIcon(Icons.Logo))
            .setCancelOnClickOutside(true)
            .setResizable(true)
            .setModalContext(false)
            .setRequestFocus(true)
            .setFocusable(true)
            .setMovable(true)
            .setBelongsToGlobalPopupStack(true)
            .setCancelOnOtherWindowOpen(true)
            .setCancelKeyEnabled(true)
            .setNormalWindowLevel(true)
            .setCancelCallback(() -> {
                popup.setUiVisible(false);
                // We return false to prevent the default cancellation behavior.
                return false;
            });

        // For some reason, adding a cancelCallback will prevent the cancel event to fire when using the escape key. To
        // work around this, we add a manual listener to both the global key handler (since the editor component seems
        // to work around the default swing event hands long) and the browser panel which seems to handle events in a
        // separate queue.
        registerGlobalKeyListeners();
        registerJBCefClientKeyListeners();

        return builder.createPopup();
    }

    private void registerGlobalKeyListeners() {
        KeyboardFocusManager.getCurrentKeyboardFocusManager()
            .addKeyEventDispatcher(e -> {
                if (e.getID() != KeyEvent.KEY_PRESSED || popup.isDisposed() || !popup.isVisible()) {
                    return true;
                }

                return handleKeyPress(e.getKeyCode(), e.getModifiersEx());
            });
    }

    private void registerJBCefClientKeyListeners() {
        mainPanel.getBrowser().getJBCefClient().addKeyboardHandler(new CefKeyboardHandler() {
            @Override
            public boolean onPreKeyEvent(CefBrowser browser, CefKeyEvent event, BoolRef is_keyboard_shortcut) {
                return false;
            }
            @Override
            public boolean onKeyEvent(CefBrowser browser, CefKeyEvent event) {
                return handleKeyPress(event.windows_key_code, event.modifiers);
            }
        }, mainPanel.getBrowser().getCefBrowser());
    }

    private boolean handleKeyPress(int keyCode, int modifiers) {
        if (keyCode == KeyEvent.VK_ESCAPE) {
            popup.setUiVisible(false);
            return false;
        }

        if (keyCode == KeyEvent.VK_ENTER && (modifiers & ALT_DOWN_MASK) == ALT_DOWN_MASK) {
            mainPanel.getPreviewPanel().openInEditor();
            return false;
        }

        return true;
    }

    @Override
    public void dispose() {
        if (popup != null) {
            popup.dispose();
        }
    }
}
