package com.sourcegraph.config;

import com.intellij.openapi.options.Configurable;
import com.intellij.openapi.project.Project;
import com.intellij.openapi.util.Disposer;
import com.intellij.util.messages.MessageBus;
import com.sourcegraph.cody.localapp.LocalAppManager;
import javax.swing.*;
import org.apache.commons.lang3.StringUtils;
import org.jetbrains.annotations.Nls;
import org.jetbrains.annotations.Nullable;

/** Provides controller functionality for application settings. */
public class SettingsConfigurable implements Configurable {

  private static final String MOCK_ACCESS_TOKEN = StringUtils.repeat("x", 40);
  private final Project project;
  private SettingsComponent mySettingsComponent;

  public SettingsConfigurable(Project project) {
    this.project = project;
  }

  @Nls(capitalization = Nls.Capitalization.Title)
  @Override
  public String getDisplayName() {
    return "Cody AI by Sourcegraph";
  }

  @Override
  public JComponent getPreferredFocusedComponent() {
    return mySettingsComponent.getPreferredFocusedComponent();
  }

  @Nullable
  @Override
  public JComponent createComponent() {
    mySettingsComponent = new SettingsComponent();
    Disposer.register(project, mySettingsComponent);
    return mySettingsComponent.getPanel();
  }

  @Override
  public boolean isModified() {
    return !mySettingsComponent.getInstanceType().equals(ConfigUtil.getInstanceType(project))
        || !mySettingsComponent.getEnterpriseUrl().equals(ConfigUtil.getEnterpriseUrl(project))
        || mySettingsComponent.isDotComAccessTokenChanged()
        || mySettingsComponent.isEnterpriseAccessTokenChanged()
        || !mySettingsComponent
            .getCustomRequestHeaders()
            .equals(ConfigUtil.getCustomRequestHeaders(project))
        || !mySettingsComponent
            .getDefaultBranchName()
            .equals(ConfigUtil.getDefaultBranchName(project))
        || !mySettingsComponent
            .getRemoteUrlReplacements()
            .equals(ConfigUtil.getRemoteUrlReplacements(project))
        || mySettingsComponent.isUrlNotificationDismissed()
            != ConfigUtil.isUrlNotificationDismissed()
        || mySettingsComponent.isCodyEnabled() != ConfigUtil.isCodyEnabled()
        || mySettingsComponent.isCodyAutoCompleteEnabled()
            != ConfigUtil.isCodyAutoCompleteEnabled();
  }

  @Override
  public void apply() {
    MessageBus bus = project.getMessageBus();
    PluginSettingChangeActionNotifier publisher =
        bus.syncPublisher(PluginSettingChangeActionNotifier.TOPIC);

    CodyApplicationService aSettings = CodyApplicationService.getInstance();
    CodyProjectService pSettings = CodyProjectService.getInstance(project);

    boolean oldCodyEnabled = ConfigUtil.isCodyEnabled();
    boolean oldCodyAutoCompleteEnabled = ConfigUtil.isCodyAutoCompleteEnabled();
    String oldUrl = ConfigUtil.getSourcegraphUrl(project);
    String newDotComAccessToken = mySettingsComponent.getDotComAccessToken();
    String newEnterpriseAccessToken = mySettingsComponent.getEnterpriseAccessToken();
    String enterpriseUrl = mySettingsComponent.getEnterpriseUrl();
    SettingsComponent.InstanceType newInstanceType = mySettingsComponent.getInstanceType();
    String newUrl;
    if (newInstanceType.equals(SettingsComponent.InstanceType.DOTCOM)) {
      newUrl = ConfigUtil.DOTCOM_URL;
    } else if (newInstanceType.equals(SettingsComponent.InstanceType.ENTERPRISE)) {
      newUrl = enterpriseUrl;
    } else {
      newUrl = LocalAppManager.getLocalAppUrl();
    }
    PluginSettingChangeContext context =
        new PluginSettingChangeContext(
            oldCodyEnabled,
            oldCodyAutoCompleteEnabled,
            oldUrl,
            newUrl,
            mySettingsComponent.isDotComAccessTokenChanged(),
            newDotComAccessToken,
            mySettingsComponent.isEnterpriseAccessTokenChanged(),
            newEnterpriseAccessToken,
            mySettingsComponent.getCustomRequestHeaders(),
            mySettingsComponent.isCodyEnabled(),
            mySettingsComponent.isCodyAutoCompleteEnabled());

    publisher.beforeAction(context);

    if (pSettings.instanceType != null) {
      pSettings.instanceType = newInstanceType.name();
    } else {
      aSettings.instanceType = newInstanceType.name();
    }
    if (pSettings.url != null) {
      pSettings.url = enterpriseUrl;
    } else {
      aSettings.url = enterpriseUrl;
    }
    if (newInstanceType == SettingsComponent.InstanceType.DOTCOM) {
      if (!newDotComAccessToken.equals(MOCK_ACCESS_TOKEN)) {
        if (pSettings.dotComAccessToken != null) {
          pSettings.dotComAccessToken = newDotComAccessToken;
        } else {
          aSettings.setSafeDotComAccessToken(newDotComAccessToken);
          aSettings.isDotComAccessTokenSet = StringUtils.isNotEmpty(newDotComAccessToken);
        }
      }
    }
    if (newInstanceType == SettingsComponent.InstanceType.ENTERPRISE) {
      if (!newEnterpriseAccessToken.equals(MOCK_ACCESS_TOKEN)) {
        if (pSettings.enterpriseAccessToken != null) {
          pSettings.enterpriseAccessToken = newEnterpriseAccessToken;
        } else {
          aSettings.setSafeEnterpriseAccessToken(newEnterpriseAccessToken);
          aSettings.isEnterpriseAccessTokenSet = StringUtils.isNotEmpty(newEnterpriseAccessToken);
        }
      }
    }
    if (pSettings.customRequestHeaders != null) {
      pSettings.customRequestHeaders = mySettingsComponent.getCustomRequestHeaders();
    } else {
      aSettings.customRequestHeaders = mySettingsComponent.getCustomRequestHeaders();
    }
    if (pSettings.defaultBranch != null) {
      pSettings.defaultBranch = mySettingsComponent.getDefaultBranchName();
    } else {
      aSettings.defaultBranch = mySettingsComponent.getDefaultBranchName();
    }
    if (pSettings.remoteUrlReplacements != null) {
      pSettings.remoteUrlReplacements = mySettingsComponent.getRemoteUrlReplacements();
    } else {
      aSettings.remoteUrlReplacements = mySettingsComponent.getRemoteUrlReplacements();
    }
    aSettings.isUrlNotificationDismissed = mySettingsComponent.isUrlNotificationDismissed();
    aSettings.setCodyEnabled(mySettingsComponent.isCodyEnabled());
    aSettings.isCodyAutoCompleteEnabled = mySettingsComponent.isCodyAutoCompleteEnabled();

    publisher.afterAction(context);
  }

  @Override
  public void reset() {
    mySettingsComponent.setInstanceType(ConfigUtil.getInstanceType(project));
    mySettingsComponent.setEnterpriseUrl(ConfigUtil.getEnterpriseUrl(project));
    boolean dotComAccessTokenSet = ConfigUtil.isDotComAccessTokenSet(project);
    mySettingsComponent.setDotComAccessToken(dotComAccessTokenSet ? MOCK_ACCESS_TOKEN : "");
    boolean enterpriseAccessTokenSet = ConfigUtil.isEnterpriseAccessTokenSet(project);
    mySettingsComponent.setEnterpriseAccessToken(enterpriseAccessTokenSet ? MOCK_ACCESS_TOKEN : "");
    mySettingsComponent.setCustomRequestHeaders(ConfigUtil.getCustomRequestHeaders(project));
    String defaultBranchName = ConfigUtil.getDefaultBranchName(project);
    mySettingsComponent.setDefaultBranchName(defaultBranchName);
    String remoteUrlReplacements = ConfigUtil.getRemoteUrlReplacements(project);
    mySettingsComponent.setRemoteUrlReplacements(remoteUrlReplacements);
    mySettingsComponent.setUrlNotificationDismissedEnabled(ConfigUtil.isUrlNotificationDismissed());
    mySettingsComponent.setCodyEnabled(ConfigUtil.isCodyEnabled());
    mySettingsComponent.setCodyAutoCompleteEnabled(ConfigUtil.isCodyAutoCompleteEnabled());
    mySettingsComponent.setDotComAccessTokenNotChanged();
    mySettingsComponent.setEnterpriseAccessTokenNotChanged();
    mySettingsComponent.getPanel().requestFocusInWindow();
  }

  @Override
  public void disposeUIResources() {
    mySettingsComponent = null;
  }
}
